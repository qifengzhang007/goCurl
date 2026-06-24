package test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	goCurl "github.com/qifengzhang007/goCurl"
)

// TestCreateSSEServer 测试创建 SSE 服务端
func TestCreateSSEServer(t *testing.T) {
	s := goCurl.CreateSSEServer()
	if s == nil {
		t.Fatal("CreateSSEServer() returned nil")
	}
	if s.ClientCount() != 0 {
		t.Fatalf("expected 0 clients, got %d", s.ClientCount())
	}
	if s.HasClients() {
		t.Fatal("expected HasClients() to be false")
	}
}

// TestCreateSSEServerWithOptions 测试带配置创建 SSE 服务端
func TestCreateSSEServerWithOptions(t *testing.T) {
	s := goCurl.CreateSSEServer(goCurl.SSEOptions{
		BufferSize: 512,
		RetryTime:  5000,
		KeepAlive:  0,
	})
	if s == nil {
		t.Fatal("CreateSSEServer() returned nil")
	}
}

// TestSSEClientConnect 测试客户端连接
func TestSSEClientConnect(t *testing.T) {
	var connectedID string
	var wg sync.WaitGroup
	wg.Add(1)

	s := goCurl.CreateSSEServer(goCurl.SSEOptions{
		OnConnect: func(id string) {
			connectedID = id
			wg.Done()
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()

	go func() {
		s.ServeHTTP(w, req)
	}()

	wg.Wait()

	if connectedID == "" {
		t.Fatal("expected OnConnect to be called with client ID")
	}
	if s.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", s.ClientCount())
	}
	if !s.HasClients() {
		t.Fatal("expected HasClients() to be true")
	}

	s.Close()
	time.Sleep(50 * time.Millisecond)
	if s.ClientCount() != 0 {
		t.Fatalf("expected 0 clients after close, got %d", s.ClientCount())
	}
}

// TestSSEClientDisconnect 测试客户端断开
func TestSSEClientDisconnect(t *testing.T) {
	var disconnectedID string
	var wg sync.WaitGroup
	wg.Add(1)

	s := goCurl.CreateSSEServer(goCurl.SSEOptions{
		OnDisconnect: func(id string) {
			disconnectedID = id
			wg.Done()
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()

	go func() {
		s.ServeHTTP(w, req)
	}()

	time.Sleep(50 * time.Millisecond)
	s.Close()
	wg.Wait()

	if disconnectedID == "" {
		t.Fatal("expected OnDisconnect to be called")
	}
}

// TestSSEServeHTTPHeaders 测试 SSE 响应头
func TestSSEServeHTTPHeaders(t *testing.T) {
	s := goCurl.CreateSSEServer()

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		s.ServeHTTP(w, req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("expected text/event-stream, got %s", ct)
	}

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Fatalf("expected no-cache, got %s", cacheControl)
	}

	connection := w.Header().Get("Connection")
	if connection != "keep-alive" {
		t.Fatalf("expected keep-alive, got %s", connection)
	}

	s.Close()
	<-done
}

// TestSSEBroadcast 测试广播
func TestSSEBroadcast(t *testing.T) {
	s := goCurl.CreateSSEServer(goCurl.SSEOptions{KeepAlive: 0})

	req1 := httptest.NewRequest(http.MethodGet, "/events", nil)
	w1 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/events", nil)
	w2 := httptest.NewRecorder()

	go s.ServeHTTP(w1, req1)
	go s.ServeHTTP(w2, req2)

	time.Sleep(50 * time.Millisecond)

	if s.ClientCount() != 2 {
		t.Fatalf("expected 2 clients, got %d", s.ClientCount())
	}

	s.Broadcast(&goCurl.SSEEvent{
		ID:    "1",
		Event: "message",
		Data:  "hello world",
	})

	time.Sleep(50 * time.Millisecond)

	body1 := w1.Body.String()
	body2 := w2.Body.String()

	if !strings.Contains(body1, "hello world") {
		t.Fatalf("client 1 did not receive broadcast, body: %s", body1)
	}
	if !strings.Contains(body2, "hello world") {
		t.Fatalf("client 2 did not receive broadcast, body: %s", body2)
	}

	s.Close()
}

// TestSSESendEvent 测试向指定客户端发送
func TestSSESendEvent(t *testing.T) {
	var clientID string
	var mu sync.Mutex

	s := goCurl.CreateSSEServer(goCurl.SSEOptions{
		KeepAlive: 0,
		OnConnect: func(id string) {
			mu.Lock()
			clientID = id
			mu.Unlock()
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()

	go s.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	id := clientID
	mu.Unlock()

	err := s.SendEvent(id, "custom", `{"msg":"test"}`)
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	body := w.Body.String()
	if !strings.Contains(body, "event: custom") {
		t.Fatalf("expected event: custom, body: %s", body)
	}
	if !strings.Contains(body, `{"msg":"test"}`) {
		t.Fatalf("expected data, body: %s", body)
	}

	s.Close()
}

// TestSSESendEventClientNotFound 测试向不存在客户端发送
func TestSSESendEventClientNotFound(t *testing.T) {
	s := goCurl.CreateSSEServer()
	err := s.SendEvent("nonexistent", "msg", "data")
	if err != goCurl.ErrSSEClientNotFound {
		t.Fatalf("expected ErrSSEClientNotFound, got %v", err)
	}
}

// TestSSEWriteEvent 测试事件格式化（通过真实 SSE 连接验证）
func TestSSEWriteEvent(t *testing.T) {
	s := goCurl.CreateSSEServer(goCurl.SSEOptions{KeepAlive: 0})

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		s.ServeHTTP(w, req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	s.Broadcast(&goCurl.SSEEvent{
		ID:    "evt-1",
		Event: "update",
		Data:  `{"key":"value"}`,
		Retry: 5000,
	})

	time.Sleep(50 * time.Millisecond)
	s.Close()
	<-done

	body := w.Body.String()
	if !strings.Contains(body, "id: evt-1") {
		t.Error("missing id field")
	}
	if !strings.Contains(body, "event: update") {
		t.Error("missing event field")
	}
	if !strings.Contains(body, `data: {"key":"value"}`) {
		t.Errorf("missing data field, body: %s", body)
	}
	if !strings.Contains(body, "retry: 5000") {
		t.Error("missing retry field")
	}
}

// TestSSEWriteEventMultilineData 测试多行 data 字段
func TestSSEWriteEventMultilineData(t *testing.T) {
	s := goCurl.CreateSSEServer(goCurl.SSEOptions{KeepAlive: 0})

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		s.ServeHTTP(w, req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	s.Broadcast(&goCurl.SSEEvent{Data: "line1\nline2\nline3"})

	time.Sleep(50 * time.Millisecond)
	s.Close()
	<-done

	body := w.Body.String()
	if !strings.Contains(body, "data: line1\n") {
		t.Errorf("missing line1, body: %s", body)
	}
	if !strings.Contains(body, "data: line2\n") {
		t.Errorf("missing line2, body: %s", body)
	}
	if !strings.Contains(body, "data: line3\n") {
		t.Errorf("missing line3, body: %s", body)
	}
}

// TestSSEClose 测试关闭
func TestSSEClose(t *testing.T) {
	s := goCurl.CreateSSEServer(goCurl.SSEOptions{KeepAlive: 0})

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		s.ServeHTTP(w, req)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	if s.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", s.ClientCount())
	}

	s.Close()
	<-done

	if s.ClientCount() != 0 {
		t.Fatalf("expected 0 clients after close, got %d", s.ClientCount())
	}
}
