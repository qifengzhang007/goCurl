package goCurl

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/axgle/mahonia"
	"io"
	"mime/multipart"
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
	opts           Options
	cli            *http.Client
	req            *http.Request
	body           io.Reader
	formDataParams string
	cookiesJar     *cookiejar.Jar
	fileUpload     FileUpload
}

// Get send get request
func (r *Request) Get(uri string, opts ...Options) (*Response, error) {
	return r.Request("GET", uri, opts...)
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

// Down method  download files
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
			return false, errors.New(downloadFileIsEmpty)
		}
	}
	return false, vError
}

func (r *Request) saveFile(body io.ReadCloser, fileName string) (bool, error) {
	var isOccurError bool
	var OccurError error
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	defer func() {
		_ = body.Close()
		_ = file.Close()
	}()
	reader := bufio.NewReader(body)
	if err != nil {
		return false, err
	}
	writer := bufio.NewWriter(file)
	buff := make([]byte, 4096)

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

// UploadFile
// @url 文件上传地址
// @formFileName 服务端接受文件表单的键名，例如：file
// @filePath 待上传的文件完整路径
func (r *Request) UploadFile(uri, formFileName, filePath string, opts ...Options) (resp *Response, err error) {
	var file *os.File
	var formFileIoWrite io.Writer

	// 打开要上传的文件
	if file, err = os.Open(filePath); err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建一个缓冲区来存储multipart表单数据
	r.fileUpload.multipartWrite = multipart.NewWriter(&r.fileUpload.fileUploadBody)
	// 创建一个表单文件字段
	if formFileIoWrite, err = r.fileUpload.multipartWrite.CreateFormFile(formFileName, filePath); err != nil {
		return nil, err
	}

	// 将文件内容复制到表单文件字段中
	if _, err = io.Copy(formFileIoWrite, file); err != nil {
		return nil, err
	}

	return r.Request(postWithFileUpload, uri, opts...)
}

// Sse  客户端请求，持续获取服务端推送的数据流
func (r *Request) Sse(method, uri string, fn func(msgType, content string) bool, options ...Options) (err error) {
	var tmpOptions = defaultHeader()
	if len(options) > 0 {
		tmpOptions = mergeDefaultParams(tmpOptions, options[0])
	}
	resp := &Response{}
	if strings.ToUpper(method) == http.MethodGet {
		resp, err = r.Get(uri, tmpOptions)
	} else if strings.ToUpper(method) == http.MethodPost {
		resp, err = r.Post(uri, tmpOptions)
	}

	if err == nil {
		body := resp.GetBody()
		defer func() {
			_ = body.Close()
		}()
		ioReader := bufio.NewReader(body)
		for {
			if bys, err := ioReader.ReadBytes('\n'); err == nil && len(bys) > 4 {
				delim := []byte{':', ' '}
				byteSliceSlice := bytes.Split(bys, delim)
				if len(byteSliceSlice) == 2 {
					if !fn(string(byteSliceSlice[0]), string(byteSliceSlice[1])) {
						return nil
					}
				}
			} else {
				// 如果ioreader关联的缓冲区没有内容，休眠5毫秒（避免死循环导致cpu占用率过高）
				// 相对网络请求的耗时, 5ms 时间几乎不构成任何影响
				time.Sleep(time.Millisecond * 5)
			}
		}
	} else {
		return errors.New(err.Error())
	}
}

// Options send options request
func (r *Request) Options(uri string, opts ...Options) (*Response, error) {
	return r.Request("OPTIONS", uri, opts...)
}

// Request send request
func (r *Request) Request(method, uri string, opts ...Options) (*Response, error) {
	if len(opts) > 0 {
		r.opts = mergeDefaultParams(defaultHeader(), opts[0], r.opts)
	}
	switch method {
	case http.MethodGet, http.MethodDelete:
		uri = r.opts.BaseURI + r.parseFormData(method, uri)
		req, err := http.NewRequest(method, uri, nil)
		if err != nil {
			return nil, err
		}
		r.req = req
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodOptions:
		// parse body
		r.parseBody()
		uri = r.opts.BaseURI + r.parseFormData(method, uri)
		req, err := http.NewRequest(method, uri, r.body)
		if err != nil {
			return nil, err
		}
		r.req = req
	case postWithFileUpload:
		r.parseBodyWithFileUpload()
		uri = r.opts.BaseURI + r.parseFormData(method, uri)
		req, err := http.NewRequest(http.MethodPost, uri, r.body)
		if err != nil {
			return nil, err
		}
		r.req = req
	default:
		return nil, errors.New(invalidMethod)
	}
	r.opts.Headers["Host"] = fmt.Sprintf("%v", r.req.Host)

	// parseTimeout
	r.parseTimeout()

	// parseClient
	r.parseClient()

	// parse headers
	r.parseHeaders()

	// parse cookies
	r.parseCookies()
	_resp, err := r.cli.Do(r.req)
	resp := &Response{
		resp:          _resp,
		req:           r.req,
		cookiesJar:    r.cookiesJar,
		err:           err,
		setResCharset: r.opts.SetResCharset,
	}

	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (r *Request) parseTimeout() {
	if r.opts.Timeout > 0 {
		r.opts.timeout = time.Duration(r.opts.Timeout*1000) * time.Millisecond
	} else {
		r.opts.Timeout = 0
	}
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
			fmt.Println(r.opts.Proxy+proxyError, err.Error())
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
			if strings.ReplaceAll(v, " ", "") != "" {
				r.req.AddCookie(&http.Cookie{
					Name:  k,
					Value: v,
				})
			}
		}
	case []*http.Cookie:
		cookies := r.opts.Cookies.([]*http.Cookie)
		for _, cookie := range cookies {
			if cookie != nil {
				r.req.AddCookie(cookie)
			}
		}
	}
}

func (r *Request) parseHeaders() {
	if r.opts.Headers != nil {
		for k, v := range r.opts.Headers {
			if vv, ok := v.([]string); ok {
				for _, vvv := range vv {
					if strings.ReplaceAll(vvv, " ", "") != "" {
						r.req.Header.Add(k, vvv)
					}
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
					if strings.ReplaceAll(vvv, " ", "") != "" {
						values.Add(k, vvv)
					}
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

	// text/xml
	if r.opts.XML != "" {
		r.body = strings.NewReader(r.opts.XML)
		return
	}

	return
}

func (r *Request) parseBodyWithFileUpload() {
	if r.opts.FormParams != nil {
		values := url.Values{}
		for k, v := range r.opts.FormParams {
			if vv, ok := v.([]string); ok {
				for _, vvv := range vv {
					if strings.ReplaceAll(vvv, " ", "") != "" {
						values.Add(k, vvv)
						_ = r.fileUpload.multipartWrite.WriteField(k, vvv)
					}
				}
			}
			vv := fmt.Sprintf("%v", v)
			_ = r.fileUpload.multipartWrite.WriteField(k, vv)
		}
	}
	// 关闭multipart writer以完成表单数据的写入
	_ = r.fileUpload.multipartWrite.Close()

	r.opts.Headers["Content-Type"] = r.fileUpload.multipartWrite.FormDataContentType()
	r.body = &r.fileUpload.fileUploadBody

}

// 解析 get 方式传递的 formData(application/x-www-form-urlencoded)
func (r *Request) parseFormData(method, uri string) string {
	uriPre := ""
	uriParse, err := url.Parse(uri)
	values := url.Values{}
	if err != nil {
		return ""
	}

	uriPre = uriParse.Scheme + "://" + uriParse.Host + uriParse.Path
	for _, item := range strings.Split(uriParse.RawQuery, "&") {
		if key, val, find := strings.Cut(item, "="); find {
			val = fmt.Sprintf("%v", val)
			values.Set(key, val)
		}
	}

	if r.opts.FormParams != nil && (method == http.MethodGet || method == http.MethodDelete) {
		for k, v := range r.opts.FormParams {
			if vv, ok := v.([]string); ok {
				for _, vvv := range vv {
					if strings.ReplaceAll(vvv, " ", "") != "" {
						values.Add(k, vvv)
					}
				}
			}
			vv := fmt.Sprintf("%v", v)
			values.Set(k, vv)
		}
	}
	queryParams := ""
	if len(values.Encode()) > 0 {
		queryParams = "?" + values.Encode()
	}
	return uriPre + queryParams
}

// （接受到的）简体中文 转换为 utf-8
func (r *Request) SimpleChineseToUtf8(vBytes []byte) string {
	return mahonia.NewDecoder("GB18030").ConvertString(string(vBytes))
}

// （一般是go 语言发送的数据）utf-8 转换为  简体中文发出去
func (r *Request) Utf8ToSimpleChinese(vBytes []byte, charset ...string) string {
	if len(charset) == 0 {
		return mahonia.NewEncoder("GB18030").ConvertString(string(vBytes))
	} else {
		return mahonia.NewEncoder(charset[0]).ConvertString(string(vBytes))
	}
}
