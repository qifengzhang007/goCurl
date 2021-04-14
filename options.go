package goCurl

import "time"

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
