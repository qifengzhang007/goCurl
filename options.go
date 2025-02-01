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
	SetResCharset string
}
type FileUpload struct {
	formFileName   string
	srcFilePath    string
	fileUploadBody bytes.Buffer
	multipartWrite *multipart.Writer
}
