package goCurl

import (
	"net/http"
	"net/http/cookiejar"
)

var curSiteCookiesJar, _ = cookiejar.New(nil)

func CreateHttpClient(opts ...Options) *Request {
	req := &Request{
		cli: &http.Client{
			Jar: curSiteCookiesJar,
		},
	}
	if len(opts) > 0 {
		req.opts = mergeDefaultParams(defaultHeader(), opts[0], Options{})
	} else {
		req.opts = defaultHeader()
	}
	req.cookiesJar = curSiteCookiesJar
	return req
}

// mergeDefaultParams merges default, client-level, and method-level options with clear precedence:
// Priority (high to low): method opts > client opts > built-in defaults
// @defaultOpts: Built-in default options
// @clientOpts: Options provided when creating HTTP client via CreateHttpClient
// @methodOpts: Options provided when calling Get/Post/Put/Down/UploadFile methods
func mergeDefaultParams(defaultOpts, clientOpts, methodOpts Options) Options {
	result := defaultOpts

	if clientOpts.BaseURI != "" {
		result.BaseURI = clientOpts.BaseURI
	}
	if methodOpts.BaseURI != "" {
		result.BaseURI = methodOpts.BaseURI
	}

	if clientOpts.Timeout > 0 {
		result.Timeout = clientOpts.Timeout
	}
	if methodOpts.Timeout > 0 {
		result.Timeout = methodOpts.Timeout
	}

	if clientOpts.Proxy != "" {
		result.Proxy = clientOpts.Proxy
	}
	if methodOpts.Proxy != "" {
		result.Proxy = methodOpts.Proxy
	}

	if clientOpts.Cookies != nil {
		result.Cookies = clientOpts.Cookies
	}
	if methodOpts.Cookies != nil {
		result.Cookies = methodOpts.Cookies
	}

	if clientOpts.SetResCharset != "" {
		result.SetResCharset = clientOpts.SetResCharset
	}
	if methodOpts.SetResCharset != "" {
		result.SetResCharset = methodOpts.SetResCharset
	}

	if clientOpts.FormParams != nil {
		result.FormParams = clientOpts.FormParams
	}
	if methodOpts.FormParams != nil {
		result.FormParams = methodOpts.FormParams
	}

	if clientOpts.JSON != nil {
		result.JSON = clientOpts.JSON
	}
	if methodOpts.JSON != nil {
		result.JSON = methodOpts.JSON
	}

	if clientOpts.XML != "" {
		result.XML = clientOpts.XML
	}
	if methodOpts.XML != "" {
		result.XML = methodOpts.XML
	}

	result.Headers = mergeHeaders(defaultOpts.Headers, clientOpts.Headers, methodOpts.Headers)

	return result
}

func mergeHeaders(defaults, client, method map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range defaults {
		result[k] = v
	}
	for k, v := range client {
		result[k] = v
	}

	for k, v := range method {
		result[k] = v
	}

	return result
}

// 默认设置headers头信息，模拟浏览器默认参数
func defaultHeader() Options {
	headers := Options{
		Headers: map[string]interface{}{
			"User-Agent":   "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.81 Safari/537.36 SE 2.X MetaSr 1.0",
			"Accept":       "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
			"Content-Type": "application/x-www-form-urlencoded;charset=utf-8",
			//	特别提醒：真实的浏览器该值为 Accept-Encoding: gzip, deflate，表示浏览器接受压缩后的二进制，浏览器端再解析为html展示，
			//	但是HttpClient解析就麻烦了，所以必须为空或者不设置该值，接受原始数据。否则很容易出现乱码
			"Accept-Encoding":           "",
			"Accept-Language":           "zh-CN,zh;q=0.9",
			"Upgrade-Insecure-Requests": "1",
			"Connection":                "close",
			"Cache-Control":             "max-age=0",
			"Host":                      "",
		},
	}
	return headers
}
