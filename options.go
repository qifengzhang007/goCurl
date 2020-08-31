package goCurl

import "time"

// Options object
type Options struct {
	Headers    map[string]interface{}
	BaseURI    string
	FormParams map[string]interface{}
	JSON       interface{}
	Timeout    float32
	timeout    time.Duration
	Cookies    interface{}
	Proxy      string
}
