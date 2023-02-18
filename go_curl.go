package goCurl

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"sync"
)

var curSiteCookiesJar, _ = cookiejar.New(nil)
var httpCli = sync.Pool{
	New: func() interface{} {
		return &http.Client{
			Jar: curSiteCookiesJar,
		}
	},
}

// 创建一个 HttpClient 客户端用于发送请求
func CreateHttpClient(opts ...Options) *Request {
	var hClient = httpCli.Get().(*http.Client)
	defer httpCli.Put(hClient)
	req := &Request{
		cli: hClient,
	}
	if len(opts) > 0 {
		req.opts = mergeDefaultParams(defaultHeader(), opts[0])
	} else {
		req.opts = defaultHeader()
	}
	req.cookiesJar = curSiteCookiesJar
	return req
}

// 合并用户提供的header头字段信息，用户提供的header头优先于默认头字段信息
// @defaultHeaders 默认 header 参数
// @options[0]  用户方法 GET 、POST等提交的参数
// @options[1]  CreateHttpClient 时初始化的参数
func mergeDefaultParams(defaultHeaders Options, options ...Options) Options {
	if len(options) == 0 {
		return defaultHeader()
	} else {

		if len(options) == 2 {
			if options[0].Headers == nil {
				options[0].Headers = make(map[string]interface{}, 1)
			}
			for key, value := range options[1].Headers {
				if _, exists := options[0].Headers[key]; !exists {
					options[0].Headers[key] = fmt.Sprintf("%v", value)
				}
			}
			// header 头参数参数合并完成后，继续合并以下几个参数:BaseURI 、Timeout等
			if options[0].BaseURI == "" && options[1].BaseURI != "" {
				options[0].BaseURI = options[1].BaseURI
			}

			if options[0].Timeout <= 0 && options[1].Timeout >= 0 {
				options[0].Timeout = options[1].Timeout
			}
			if options[0].Proxy == "" && options[1].Proxy != "" {
				options[0].Proxy = options[1].Proxy
			}
			if options[0].Cookies == nil && options[1].Cookies != nil {
				options[0].Cookies = options[1].Cookies
			}
			if options[0].SetResCharset == "" && options[1].SetResCharset != "" {
				options[0].SetResCharset = options[1].SetResCharset
			}
		}

		for key, value := range defaultHeaders.Headers {
			if options[0].Headers != nil {
				if _, exists := options[0].Headers[key]; !exists {
					options[0].Headers[key] = fmt.Sprintf("%v", value)
				}
			} else {
				options[0].Headers = make(map[string]interface{}, 1)
				options[0].Headers[key] = fmt.Sprintf("%v", value)
			}
		}

		return options[0]
	}
}

// 默认设置headers头信息，尽可能伪装成为真实的浏览器
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
