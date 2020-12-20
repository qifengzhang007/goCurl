package goCurl

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

// Request object
type Request struct {
	opts                 Options
	cli                  *http.Client
	req                  *http.Request
	body                 io.Reader
	subGetFormDataParmas string
	cookiesJar           *cookiejar.Jar
}

// Get send get request
func (r *Request) Get(uri string, opts ...Options) (*Response, error) {
	return r.Request("GET", uri, opts...)
}

// Get method  download files
func (r *Request) Down(resourceUrl string, savePath, saveName string, opts ...Options) (bool, error) {
	var vError error
	var vResponse *Response
	uri, err := url.ParseRequestURI(resourceUrl)
	if err != nil {
		return false, err
	}
	if vResponse, vError = r.Request("GET", resourceUrl, opts...); vError == nil {
		filename := path.Base(uri.Path)
		if len(saveName) > 0 {
			filename = saveName
		}
		if vResponse.GetStatusCode() == 200 || vResponse.GetContentLength() > 0 {
			body := vResponse.GetBody()
			return r.saveFile(body, savePath+filename)
		} else {
			return false, errors.New("被下载的文件内容为空")
		}
	}
	return false, vError
}

func (r *Request) saveFile(body io.ReadCloser, fileName string) (bool, error) {
	var isOccurError bool
	var OccurError error
	defer func() {
		_ = body.Close()
	}()
	reader := bufio.NewReaderSize(body, 1024*50) //相当于一个临时缓冲区(设置为可以单次存储5M的文件)，每次读取以后就把原始数据重新加载一份，等待下一次读取
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return false, err
	}
	writer := bufio.NewWriter(file)
	buff := make([]byte, 50*1024)

	for {
		currReadSize, readerErr := reader.Read(buff)
		if currReadSize > 0 {
			_, OccurError = writer.Write(buff[0:currReadSize])
			if OccurError != nil {
				isOccurError = true
				break
			}
		}
		// 读取结束
		if readerErr == io.EOF {
			_ = writer.Flush()
			break
		}
	}
	// 如果没有发生错误，就返回 true
	if isOccurError == false {
		return true, nil
	} else {
		return false, OccurError
	}

}

// Post send post request
func (r *Request) Post(uri string, opts ...Options) (*Response, error) {
	return r.Request("POST", uri, opts...)
}

// Put send put request
func (r *Request) Put(uri string, opts ...Options) (*Response, error) {
	return r.Request("PUT", uri, opts...)
}

// Patch send patch request
func (r *Request) Patch(uri string, opts ...Options) (*Response, error) {
	return r.Request("PATCH", uri, opts...)
}

// Delete send delete request
func (r *Request) Delete(uri string, opts ...Options) (*Response, error) {
	return r.Request("DELETE", uri, opts...)
}

// Options send options request
func (r *Request) Options(uri string, opts ...Options) (*Response, error) {
	return r.Request("OPTIONS", uri, opts...)
}

// Request send request
func (r *Request) Request(method, uri string, opts ...Options) (*Response, error) {

	r.opts = mergeHeaders(defaultHeader(), opts...)
	switch method {
	case http.MethodGet, http.MethodDelete:
		uri = uri + r.parseGetFormData()
		req, err := http.NewRequest(method, uri, nil)
		if err != nil {
			return nil, err
		}

		r.req = req
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodOptions:
		// parse body
		r.parseBody()

		req, err := http.NewRequest(method, uri, r.body)
		if err != nil {
			return nil, err
		}

		r.req = req
	default:
		return nil, errors.New("invalid request method")
	}
	r.opts.Headers["Host"] = fmt.Sprintf("%v", r.req.Host)
	// parseOptions
	r.parseOptions()

	// parseClient
	r.parseClient()

	// parse headers
	r.parseHeaders()

	// parse cookies
	r.parseCookies()

	_resp, err := r.cli.Do(r.req)
	resp := &Response{
		resp:       _resp,
		req:        r.req,
		cookiesJar: r.cookiesJar,
		err:        err,
	}

	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (r *Request) parseOptions() {
	// default timeout 30s
	if r.opts.Timeout == 0 {
		r.opts.Timeout = 30
	}
	r.opts.timeout = time.Duration(r.opts.Timeout*1000) * time.Millisecond
}

func (r *Request) parseClient() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	if r.opts.Proxy != "" {
		proxy, err := url.Parse(r.opts.Proxy)
		if err == nil {
			tr.Proxy = http.ProxyURL(proxy)
		} else {
			fmt.Println(r.opts.Proxy+"代理设置错误：", err.Error())
		}
	}

	r.cli = &http.Client{
		Timeout:   r.opts.timeout,
		Transport: tr,
		Jar:       r.cookiesJar,
	}
}

func (r *Request) parseCookies() {
	switch r.opts.Cookies.(type) {
	case string:
		cookies := r.opts.Cookies.(string)
		r.req.Header.Add("Cookie", cookies)
	case map[string]string:
		cookies := r.opts.Cookies.(map[string]string)
		for k, v := range cookies {
			r.req.AddCookie(&http.Cookie{
				Name:  k,
				Value: v,
			})
		}
	case []*http.Cookie:
		cookies := r.opts.Cookies.([]*http.Cookie)
		for _, cookie := range cookies {
			r.req.AddCookie(cookie)
		}
	}
}

func (r *Request) parseHeaders() {
	if r.opts.Headers != nil {
		for k, v := range r.opts.Headers {
			if vv, ok := v.([]string); ok {
				for _, vvv := range vv {
					r.req.Header.Add(k, vvv)
				}
				continue
			}
			vv := fmt.Sprintf("%v", v)
			r.req.Header.Set(k, vv)
		}
	}
}

func (r *Request) parseBody() {
	// application/x-www-form-urlencoded
	if r.opts.FormParams != nil {
		values := url.Values{}
		for k, v := range r.opts.FormParams {
			if vv, ok := v.([]string); ok {
				for _, vvv := range vv {
					values.Add(k, vvv)
				}
				continue
			}
			vv := fmt.Sprintf("%v", v)
			values.Set(k, vv)
		}
		r.body = strings.NewReader(values.Encode())

		return
	}

	// application/json
	if r.opts.JSON != nil {
		b, err := json.Marshal(r.opts.JSON)
		if err == nil {
			r.body = bytes.NewReader(b)

			return
		}
	}

	return
}

// 解析 get 方式传递的 formData(application/x-www-form-urlencoded)
func (r *Request) parseGetFormData() string {
	if r.opts.FormParams != nil {
		values := url.Values{}
		for k, v := range r.opts.FormParams {
			if vv, ok := v.([]string); ok {
				for _, vvv := range vv {
					values.Add(k, vvv)
				}
				continue
			}
			vv := fmt.Sprintf("%v", v)
			values.Set(k, vv)
		}
		r.subGetFormDataParmas = values.Encode()
		return "?" + r.subGetFormDataParmas
	} else {
		return ""
	}
}
