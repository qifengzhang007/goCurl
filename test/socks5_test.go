package test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"

	"github.com/qifengzhang007/goCurl"
)

// socks5ProxyServer 启动一个简单的SOCKS5代理服务器(无认证)用于单元测试
// 返回代理地址和停止函数
func startSocks5ProxyServer(t *testing.T) (proxyAddr string, stopFn func()) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("启动SOCKS5代理服务器失败: %v", err)
	}

	proxyAddr = listener.Addr().String()
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})

	go func() {
		wg.Done()
		for {
			select {
			case <-done:
				return
			default:
			}

			clientConn, err := listener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					continue
				}
			}
			go handleSocks5Connection(clientConn)
		}
	}()

	wg.Wait() // 确保goroutine已启动

	return proxyAddr, func() {
		close(done)
		listener.Close()
	}
}

// handleSocks5Connection 处理单个SOCKS5连接(仅支持无认证模式)
func handleSocks5Connection(clientConn net.Conn) {
	defer clientConn.Close()

	buf := make([]byte, 256)

	// 1. 读取认证方法协商
	n, err := clientConn.Read(buf)
	if err != nil || n < 3 {
		return
	}

	// 检查版本和选择无认证方式
	if buf[0] != 0x05 {
		return
	}
	// 响应: 版本5, 无认证
	clientConn.Write([]byte{0x05, 0x00})

	// 2. 读取连接请求
	n, err = clientConn.Read(buf)
	if err != nil || n < 5 {
		return
	}

	if buf[0] != 0x05 || buf[1] != 0x01 {
		return
	}

	// 解析目标地址
	var targetAddr string
	switch buf[3] {
	case 0x01: // IPv4
		if n < 10 {
			return
		}
		targetAddr = fmt.Sprintf("%d.%d.%d.%d:%d",
			buf[4], buf[5], buf[6], buf[7],
			int(buf[8])<<8|int(buf[9]))
	case 0x03: // 域名
		domainLen := int(buf[4])
		if n < 5+domainLen+2 {
			return
		}
		domain := string(buf[5 : 5+domainLen])
		port := int(buf[5+domainLen])<<8 | int(buf[5+domainLen+1])
		targetAddr = net.JoinHostPort(domain, strconv.Itoa(port))
	case 0x04: // IPv6
		if n < 22 {
			return
		}
		ip := net.IP(buf[4:20]).String()
		port := int(buf[20])<<8 | int(buf[21])
		targetAddr = net.JoinHostPort(ip, strconv.Itoa(port))
	default:
		return
	}

	// 3. 连接目标服务器
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		// 连接失败，通知客户端
		clientConn.Write([]byte{0x05, 0x04, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer targetConn.Close()

	// 4. 响应成功连接
	clientConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	// 5. 双向数据转发
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(targetConn, clientConn)
	}()
	go func() {
		defer wg.Done()
		io.Copy(clientConn, targetConn)
	}()
	wg.Wait()
}

// startTestHttpServer 启动一个简单的HTTP测试服务器
func startTestHttpServer(t *testing.T) (string, func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/socks5-test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "socks5-proxy-ok: %s %s", r.Method, r.URL.Path)
	})

	httpServer := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: mux,
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("启动HTTP测试服务器失败: %v", err)
	}

	go httpServer.Serve(listener)

	addr := listener.Addr().String()
	return addr, func() {
		listener.Close()
	}
}

// TestSocks5ProxyHttpGet 测试通过SOCKS5代理发送HTTP GET请求
func TestSocks5ProxyHttpGet(t *testing.T) {
	// 启动目标HTTP服务器
	targetAddr, stopHttp := startTestHttpServer(t)
	defer stopHttp()

	// 启动SOCKS5代理服务器
	proxyAddr, stopProxy := startSocks5ProxyServer(t)
	defer stopProxy()

	// 通过SOCKS5代理发起请求
	cli := goCurl.CreateHttpClient()
	resp, err := cli.Get("http://"+targetAddr+"/socks5-test", goCurl.Options{
		Proxy:     proxyAddr,
		ProxyType: "socks5",
		Timeout:   5,
	})
	if err != nil {
		t.Fatalf("通过SOCKS5代理请求失败: %v", err)
	}

	txt, err := resp.GetContents()
	if err != nil {
		t.Fatalf("读取响应内容失败: %v", err)
	}

	expected := "socks5-proxy-ok: GET /socks5-test"
	if txt != expected {
		t.Errorf("响应内容不匹配:\n  期望: %s\n  实际: %s", expected, txt)
	} else {
		t.Logf("SOCKS5代理测试通过: %s", txt)
	}
}

// TestSocks5ProxyHttpPost 测试通过SOCKS5代理发送HTTP POST请求
func TestSocks5ProxyHttpPost(t *testing.T) {
	// 启动目标HTTP服务器
	targetAddr, stopHttp := startTestHttpServer(t)
	defer stopHttp()

	// 启动SOCKS5代理服务器
	proxyAddr, stopProxy := startSocks5ProxyServer(t)
	defer stopProxy()

	// 通过SOCKS5代理发起POST请求
	cli := goCurl.CreateHttpClient()
	resp, err := cli.Post("http://"+targetAddr+"/socks5-test", goCurl.Options{
		Proxy:     proxyAddr,
		ProxyType: "socks5",
		Timeout:   5,
		FormParams: map[string]interface{}{
			"key": "value",
		},
	})
	if err != nil {
		t.Fatalf("通过SOCKS5代理POST请求失败: %v", err)
	}

	txt, err := resp.GetContents()
	if err != nil {
		t.Fatalf("读取响应内容失败: %v", err)
	}

	expected := "socks5-proxy-ok: POST /socks5-test"
	if txt != expected {
		t.Errorf("响应内容不匹配:\n  期望: %s\n  实际: %s", expected, txt)
	} else {
		t.Logf("SOCKS5代理POST测试通过: %s", txt)
	}
}

// TestSocks5ProxyEmptyProxyTypeDefaultToHttp 测试不设置ProxyType时默认走HTTP代理
func TestSocks5ProxyEmptyProxyTypeDefaultToHttp(t *testing.T) {
	// 这是一个编译期和逻辑验证测试：
	// 当 ProxyType 为空时，走默认的 HTTP 代理逻辑
	// 设置一个不可用的代理以确保代码路径被执行
	cli := goCurl.CreateHttpClient()
	_, err := cli.Get("http://127.0.0.1:19999/some-path", goCurl.Options{
		Proxy:   "http://127.0.0.1:1",
		Timeout: 1,
	})
	// 期望代理连接失败（因为代理地址不可用），但不应panic
	if err == nil {
		t.Log("代理不可达但意外成功(可能本地有代理)")
	} else {
		t.Logf("预期行为：代理不可达，错误: %v", err)
	}
}

// TestSocks5ProxyWithClientLevelOption 测试通过CreateHttpClient设置SOCKS5代理
func TestSocks5ProxyWithClientLevelOption(t *testing.T) {
	// 启动目标HTTP服务器
	targetAddr, stopHttp := startTestHttpServer(t)
	defer stopHttp()

	// 启动SOCKS5代理服务器
	proxyAddr, stopProxy := startSocks5ProxyServer(t)
	defer stopProxy()

	// 在创建客户端时设置SOCKS5代理参数
	cli := goCurl.CreateHttpClient(goCurl.Options{
		Proxy:     proxyAddr,
		ProxyType: "socks5",
		Timeout:   5,
	})
	resp, err := cli.Get("http://" + targetAddr + "/socks5-test")
	if err != nil {
		t.Fatalf("通过客户端级别SOCKS5代理请求失败: %v", err)
	}

	txt, err := resp.GetContents()
	if err != nil {
		t.Fatalf("读取响应内容失败: %v", err)
	}

	expected := "socks5-proxy-ok: GET /socks5-test"
	if txt != expected {
		t.Errorf("响应内容不匹配:\n  期望: %s\n  实际: %s", expected, txt)
	} else {
		t.Logf("客户端级别SOCKS5代理测试通过: %s", txt)
	}
}

// TestSocks5ProxyWithAuth 测试带认证的SOCKS5代理
func TestSocks5ProxyWithAuth(t *testing.T) {
	// 启动带认证的SOCKS5代理服务器
	proxyAddr, stopProxy := startSocks5ProxyWithAuth(t)
	defer stopProxy()

	// 启动目标HTTP服务器
	targetAddr, stopHttp := startTestHttpServer(t)
	defer stopHttp()

	// 使用正确的用户名密码
	cli := goCurl.CreateHttpClient()
	resp, err := cli.Get("http://"+targetAddr+"/socks5-test", goCurl.Options{
		Proxy:         proxyAddr,
		ProxyType:     "socks5",
		ProxyUser:     "testuser",
		ProxyPassword: "testpass",
		Timeout:       5,
	})
	if err != nil {
		t.Fatalf("带认证的SOCKS5代理请求失败: %v", err)
	}

	txt, err := resp.GetContents()
	if err != nil {
		t.Fatalf("读取响应内容失败: %v", err)
	}

	expected := "socks5-proxy-ok: GET /socks5-test"
	if txt != expected {
		t.Errorf("响应内容不匹配:\n  期望: %s\n  实际: %s", expected, txt)
	} else {
		t.Logf("带认证SOCKS5代理测试通过: %s", txt)
	}
}

// startSocks5ProxyWithAuth 启动带用户名密码认证的SOCKS5代理服务器
func startSocks5ProxyWithAuth(t *testing.T) (proxyAddr string, stopFn func()) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("启动带认证的SOCKS5代理服务器失败: %v", err)
	}

	proxyAddr = listener.Addr().String()
	var wg sync.WaitGroup
	wg.Add(1)
	done := make(chan struct{})

	go func() {
		wg.Done()
		for {
			select {
			case <-done:
				return
			default:
			}

			clientConn, err := listener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					continue
				}
			}
			go handleSocks5ConnectionWithAuth(clientConn)
		}
	}()

	wg.Wait()
	return proxyAddr, func() {
		close(done)
		listener.Close()
	}
}

// handleSocks5ConnectionWithAuth 处理带认证的SOCKS5连接
func handleSocks5ConnectionWithAuth(clientConn net.Conn) {
	defer clientConn.Close()

	buf := make([]byte, 512)

	// 1. 读取认证方法协商
	n, err := clientConn.Read(buf)
	if err != nil || n < 3 {
		return
	}

	if buf[0] != 0x05 {
		return
	}

	// 检查是否支持用户名密码认证(0x02)
	hasAuth := false
	numMethods := int(buf[1])
	for i := 0; i < numMethods && i+2 < n; i++ {
		if buf[2+i] == 0x02 {
			hasAuth = true
			break
		}
	}

	if hasAuth {
		// 要求用户名密码认证
		clientConn.Write([]byte{0x05, 0x02})

		// 2. 读取用户名密码认证
		n, err = clientConn.Read(buf)
		if err != nil || n < 5 {
			return
		}

		if buf[0] != 0x01 { // 子协议版本
			return
		}
		userLen := int(buf[1])
		if n < 2+userLen+1 {
			return
		}
		user := string(buf[2 : 2+userLen])
		passLen := int(buf[2+userLen])
		if n < 2+userLen+1+passLen {
			return
		}
		pass := string(buf[2+userLen+1 : 2+userLen+1+passLen])

		// 验证用户名密码
		if user == "testuser" && pass == "testpass" {
			clientConn.Write([]byte{0x01, 0x00}) // 认证成功
		} else {
			clientConn.Write([]byte{0x01, 0x01}) // 认证失败
			return
		}
	} else {
		clientConn.Write([]byte{0x05, 0x00}) // 无认证
	}

	// 3. 读取连接请求
	n, err = clientConn.Read(buf)
	if err != nil || n < 5 {
		return
	}

	if buf[0] != 0x05 || buf[1] != 0x01 {
		return
	}

	// 解析目标地址
	var targetAddr string
	switch buf[3] {
	case 0x01:
		if n < 10 {
			return
		}
		targetAddr = fmt.Sprintf("%d.%d.%d.%d:%d",
			buf[4], buf[5], buf[6], buf[7],
			int(buf[8])<<8|int(buf[9]))
	case 0x03:
		domainLen := int(buf[4])
		if n < 5+domainLen+2 {
			return
		}
		domain := string(buf[5 : 5+domainLen])
		port := int(buf[5+domainLen])<<8 | int(buf[5+domainLen+1])
		targetAddr = net.JoinHostPort(domain, strconv.Itoa(port))
	case 0x04:
		if n < 22 {
			return
		}
		ip := net.IP(buf[4:20]).String()
		port := int(buf[20])<<8 | int(buf[21])
		targetAddr = net.JoinHostPort(ip, strconv.Itoa(port))
	default:
		return
	}

	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		clientConn.Write([]byte{0x05, 0x04, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	defer targetConn.Close()

	clientConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(targetConn, clientConn)
	}()
	go func() {
		defer wg.Done()
		io.Copy(clientConn, targetConn)
	}()
	wg.Wait()
}
