package goCurl

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultSSEBuffer = 256
	defaultSSERetry  = 3000
)

var ErrSSEClientNotFound = errors.New("goCurl sse: client not found")

// SSEOptions SSE 服务端配置
type SSEOptions struct {
	BufferSize int           // 每个客户端事件通道缓冲区大小，默认 256
	RetryTime  int64         // 浏览器重连间隔(ms)，默认 3000
	KeepAlive  time.Duration // 心跳间隔，默认 15s；0 关闭心跳

	OnConnect    func(clientID string) // 客户端连接回调
	OnDisconnect func(clientID string) // 客户端断开回调
}

// SSEEvent 表示一条 SSE 事件
type SSEEvent struct {
	ID    string // 事件 ID，断线重连时 Last-Event-ID 的值
	Event string // 事件类型，为空时浏览器触发 onmessage
	Data  string // 事件数据
	Retry int64  // 重连间隔(ms)
}

// SSEClient 表示一个已连接的 SSE 客户端
type SSEClient struct {
	id     string
	events chan *SSEEvent
	done   chan struct{}
}

// ID 返回客户端唯一标识
func (c *SSEClient) ID() string {
	return c.id
}

// SSEServer SSE 服务端
type SSEServer struct {
	mu      sync.RWMutex
	clients map[string]*SSEClient

	opts SSEOptions
}

// mergeSSEDefaultParams 合并默认值与用户配置
func mergeSSEDefaultParams(userOpts SSEOptions) SSEOptions {
	opts := SSEOptions{
		BufferSize: defaultSSEBuffer,
		RetryTime:  defaultSSERetry,
		KeepAlive:  15 * time.Second,
	}
	if userOpts.BufferSize > 0 {
		opts.BufferSize = userOpts.BufferSize
	}
	if userOpts.RetryTime > 0 {
		opts.RetryTime = userOpts.RetryTime
	}
	if userOpts.KeepAlive >= 0 {
		opts.KeepAlive = userOpts.KeepAlive
	}
	opts.OnConnect = userOpts.OnConnect
	opts.OnDisconnect = userOpts.OnDisconnect
	return opts
}

// CreateSSEServer 创建 SSE 服务端，用法与 CreateHttpClient 一致
//
//	server := goCurl.CreateSSEServer()
//	server := goCurl.CreateSSEServer(goCurl.SSEOptions{
//	    KeepAlive: 30 * time.Second,
//	    OnConnect: func(id string) { log.Println("client connected:", id) },
//	})
func CreateSSEServer(opts ...SSEOptions) *SSEServer {
	s := &SSEServer{
		clients: make(map[string]*SSEClient),
	}
	if len(opts) > 0 {
		s.opts = mergeSSEDefaultParams(opts[0])
	} else {
		s.opts = mergeSSEDefaultParams(SSEOptions{})
	}
	return s
}

// ClientCount 返回当前连接的客户端数
func (s *SSEServer) ClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// HasClients 返回是否有客户端连接
func (s *SSEServer) HasClients() bool {
	return s.ClientCount() > 0
}

// ServeHTTP 实现 http.Handler，直接作为路由 handler 使用
//
//	http.Handle("/events", server)
//	mux.Handle("/events", server)
func (s *SSEServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")

	client := &SSEClient{
		id:     newSSEClientID(),
		events: make(chan *SSEEvent, s.opts.BufferSize),
		done:   make(chan struct{}),
	}

	s.mu.Lock()
	s.clients[client.id] = client
	s.mu.Unlock()

	if s.opts.OnConnect != nil {
		s.opts.OnConnect(client.id)
	}

	fmt.Fprintf(w, "retry: %d\n\n", s.opts.RetryTime)
	flusher.Flush()

	var keepAliveTick <-chan time.Time
	if s.opts.KeepAlive > 0 {
		ticker := time.NewTicker(s.opts.KeepAlive)
		defer ticker.Stop()
		keepAliveTick = ticker.C
	}

	ctxDone := r.Context().Done()

	defer func() {
		s.mu.Lock()
		delete(s.clients, client.id)
		s.mu.Unlock()
		close(client.done)
		if s.opts.OnDisconnect != nil {
			s.opts.OnDisconnect(client.id)
		}
	}()

	for {
		select {
		case <-ctxDone:
			return

		case <-keepAliveTick:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()

		case ev, ok := <-client.events:
			if !ok {
				return
			}
			s.writeEvent(w, ev)
			flusher.Flush()
		}
	}
}

func (s *SSEServer) writeEvent(w http.ResponseWriter, ev *SSEEvent) {
	if ev.ID != "" {
		fmt.Fprintf(w, "id: %s\n", ev.ID)
	}
	if ev.Event != "" {
		fmt.Fprintf(w, "event: %s\n", ev.Event)
	}
	if ev.Retry > 0 {
		fmt.Fprintf(w, "retry: %d\n", ev.Retry)
	}
	if ev.Data != "" {
		lines := splitSSELines(ev.Data)
		for _, line := range lines {
			fmt.Fprintf(w, "data: %s\n", line)
		}
	} else {
		fmt.Fprint(w, "data: \n")
	}
	fmt.Fprint(w, "\n")
}

func splitSSELines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		} else if s[i] == '\r' && i+1 < len(s) && s[i+1] == '\n' {
			lines = append(lines, s[start:i])
			i++
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	if len(lines) == 0 && s == "" {
		return []string{}
	}
	return lines
}

// Broadcast 向所有已连接客户端广播事件
func (s *SSEServer) Broadcast(ev *SSEEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.clients {
		select {
		case c.events <- ev:
		default:
		}
	}
}

// BroadcastEvent 便捷方法：向所有客户端广播指定类型的纯数据事件
func (s *SSEServer) BroadcastEvent(eventType, data string) {
	s.Broadcast(&SSEEvent{Event: eventType, Data: data})
}

// Send 向指定客户端发送事件
func (s *SSEServer) Send(clientID string, ev *SSEEvent) error {
	s.mu.RLock()
	c, ok := s.clients[clientID]
	s.mu.RUnlock()
	if !ok {
		return ErrSSEClientNotFound
	}
	select {
	case c.events <- ev:
	default:
	}
	return nil
}

// SendEvent 便捷方法：向指定客户端发送纯数据事件
func (s *SSEServer) SendEvent(clientID, eventType, data string) error {
	return s.Send(clientID, &SSEEvent{Event: eventType, Data: data})
}

// Close 关闭 SSE 服务端，断开所有客户端
func (s *SSEServer) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, c := range s.clients {
		close(c.events)
		delete(s.clients, id)
	}
}

var sseIDCounter int64

func newSSEClientID() string {
	n := atomic.AddInt64(&sseIDCounter, 1)
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%d-%x", n, b)
}
