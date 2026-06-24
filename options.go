package goCurl

import (
	"bytes"
	"mime/multipart"
	"time"
)

// Options object
type Options struct {
	Headers       map[string]interface{}
	BaseURI       string
	FormParams    map[string]interface{}
	JSON          interface{}
	XML           string
	Timeout       float32
	timeout       time.Duration
	Cookies       interface{}
	Proxy         string
	ProxyType     string // 代理类型: "" 或 "http"(HTTP代理), "socks5"(SOCKS5代理)
	ProxyUser     string // SOCKS5代理用户名(可选)
	ProxyPassword string // SOCKS5代理密码(可选)
	SetResCharset string
}
type FileUpload struct {
	formFileName   string
	srcFilePath    string
	fileUploadBody bytes.Buffer
	multipartWrite *multipart.Writer
}
