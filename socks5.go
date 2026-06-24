package goCurl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

// socks5Auth 定义SOCKS5认证信息
type socks5Auth struct {
	User     string
	Password string
}

// socks5Dialer 实现SOCKS5代理拨号器
type socks5Dialer struct {
	proxyAddr string
	auth      *socks5Auth
}

// newSocks5Dialer 创建SOCKS5拨号器
// proxyAddr 格式: "host:port" 例如 "127.0.0.1:1080"
func newSocks5Dialer(proxyAddr string, auth *socks5Auth) *socks5Dialer {
	return &socks5Dialer{
		proxyAddr: proxyAddr,
		auth:      auth,
	}
}

// DialContext 实现proxy.ContextDialer接口，用于http.Transport
func (d *socks5Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, fmt.Errorf("socks5: unsupported network type: %s", network)
	}

	conn, err := d.dial(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Dial 直接拨号
func (d *socks5Dialer) Dial(network, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}

func (d *socks5Dialer) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	// 1. 连接到SOCKS5代理服务器
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, network, d.proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("socks5: 连接代理服务器失败: %w", err)
	}

	// 2. 认证协商
	if err := d.authNegotiate(conn); err != nil {
		conn.Close()
		return nil, err
	}

	// 3. 发送目标地址请求
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("socks5: 解析目标地址失败: %w", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("socks5: 解析目标端口失败: %w", err)
	}

	if err := d.sendConnectRequest(conn, host, port); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// authNegotiate SOCKS5认证协商
func (d *socks5Dialer) authNegotiate(conn net.Conn) error {
	if d.auth != nil && d.auth.User != "" {
		// 要求用户名/密码认证
		_, err := conn.Write([]byte{0x05, 0x02, 0x00, 0x02}) // 支持无认证(0x00)和用户名密码认证(0x02)
		if err != nil {
			return fmt.Errorf("socks5: 发送认证方法失败: %w", err)
		}

		buf := make([]byte, 2)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return fmt.Errorf("socks5: 读取认证响应失败: %w", err)
		}

		if buf[0] != 0x05 {
			return fmt.Errorf("socks5: 无效的协议版本: %d", buf[0])
		}

		switch buf[1] {
		case 0x02:
			// 服务器选择了用户名/密码认证
			return d.sendAuth(conn)
		case 0x00:
			// 服务器选择了无认证
			return nil
		default:
			return fmt.Errorf("socks5: 不支持的认证方法: %d", buf[1])
		}
	} else {
		// 无认证
		_, err := conn.Write([]byte{0x05, 0x01, 0x00}) // 仅支持无认证
		if err != nil {
			return fmt.Errorf("socks5: 发送认证方法失败: %w", err)
		}

		buf := make([]byte, 2)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return fmt.Errorf("socks5: 读取认证响应失败: %w", err)
		}

		if buf[0] != 0x05 {
			return fmt.Errorf("socks5: 无效的协议版本: %d", buf[0])
		}

		if buf[1] != 0x00 {
			return fmt.Errorf("socks5: 代理服务器要求认证")
		}
	}

	return nil
}

// sendAuth 发送用户名/密码认证
func (d *socks5Dialer) sendAuth(conn net.Conn) error {
	userLen := len(d.auth.User)
	passLen := len(d.auth.Password)
	if userLen > 255 || passLen > 255 {
		return errors.New("socks5: 用户名或密码长度不能超过255")
	}

	msg := make([]byte, 0, 3+userLen+passLen)
	msg = append(msg, 0x01)             // 子协议版本
	msg = append(msg, byte(userLen))    // 用户名长度
	msg = append(msg, d.auth.User...)   // 用户名
	msg = append(msg, byte(passLen))    // 密码长度
	msg = append(msg, d.auth.Password...) // 密码

	if _, err := conn.Write(msg); err != nil {
		return fmt.Errorf("socks5: 发送认证信息失败: %w", err)
	}

	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return fmt.Errorf("socks5: 读取认证结果失败: %w", err)
	}

	if buf[0] != 0x01 {
		return fmt.Errorf("socks5: 无效的认证子协议版本: %d", buf[0])
	}

	if buf[1] != 0x00 {
		return errors.New("socks5: 认证失败，用户名或密码错误")
	}

	return nil
}

// sendConnectRequest 发送连接目标地址请求
func (d *socks5Dialer) sendConnectRequest(conn net.Conn, host string, port int) error {
	// 构建请求: VER(0x05) | CMD(0x01=CONNECT) | RSV(0x00) | ATYP | DST.ADDR | DST.PORT
	msg := []byte{0x05, 0x01, 0x00}

	// 判断地址类型
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			// IPv4
			msg = append(msg, 0x01)           // ATYP: IPv4
			msg = append(msg, ip4...)          // 4 bytes
		} else {
			// IPv6
			msg = append(msg, 0x04)            // ATYP: IPv6
			msg = append(msg, ip.To16()...)     // 16 bytes
		}
	} else {
		// 域名
		if len(host) > 255 {
			return errors.New("socks5: 目标域名长度不能超过255")
		}
		msg = append(msg, 0x03)               // ATYP: DOMAINNAME
		msg = append(msg, byte(len(host)))    // 域名长度
		msg = append(msg, host...)            // 域名字符串
	}

	// 端口 (大端序)
	msg = append(msg, byte(port>>8), byte(port&0xff))

	if _, err := conn.Write(msg); err != nil {
		return fmt.Errorf("socks5: 发送连接请求失败: %w", err)
	}

	// 读取响应
	buf := make([]byte, 10)
	if _, err := io.ReadFull(conn, buf[:4]); err != nil {
		return fmt.Errorf("socks5: 读取响应头失败: %w", err)
	}

	if buf[0] != 0x05 {
		return fmt.Errorf("socks5: 无效的响应版本: %d", buf[0])
	}

	if buf[1] != 0x00 {
		return fmt.Errorf("socks5: 连接请求被拒绝，响应码: %d", buf[1])
	}

	// 根据ATYP跳过剩余的地址字节
	switch buf[3] {
	case 0x01: // IPv4
		if _, err := io.ReadFull(conn, buf[:6]); err != nil { // 4 bytes IP + 2 bytes port - we already read 4
			return fmt.Errorf("socks5: 读取响应地址失败: %w", err)
		}
	case 0x03: // DOMAINNAME
		if _, err := io.ReadFull(conn, buf[:1]); err != nil {
			return fmt.Errorf("socks5: 读取域名长度失败: %w", err)
		}
		domainLen := int(buf[0])
		if _, err := io.ReadFull(conn, make([]byte, domainLen+2)); err != nil {
			return fmt.Errorf("socks5: 读取域名和端口失败: %w", err)
		}
	case 0x04: // IPv6
		if _, err := io.ReadFull(conn, make([]byte, 18)); err != nil { // 16 bytes IP + 2 bytes port
			return fmt.Errorf("socks5: 读取IPv6地址失败: %w", err)
		}
	}

	return nil
}
