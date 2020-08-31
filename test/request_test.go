package test

import (
	"fmt"
	"github.com/qifengzhang007/goCurl"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
)

//  get 网站编码为 gbk
// 主要测试 get 请求以及自动转换被采集网站的编码
func TestRequest_Get(t *testing.T) {
	cli := goCurl.CreateCurlClient()
	resp, err := cli.Get("http://hq.sinajs.cn/list=sh601006")
	if err != nil && resp == nil {
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}

	txt, err := resp.GetContents()
	if err == nil {
		t.Logf("请求结果：%s\n", txt)
	} else {
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}
}

// 获取 cookie
func TestRequest_GetCookies(t *testing.T) {
	cli := goCurl.CreateCurlClient()
	resp, err := cli.Get(`http://www.iwencai.com/diag/block-detail?pid=10751&codes=600422&codeType=stock&info={"view":{"nolazy":1}}`)

	if err != nil {
		t.Errorf("采集同花顺站点发生错误：%s\n", err.Error())
	}
	// 全量获取cookie
	for index, value := range resp.GetCookies() {
		t.Logf("序号：%d, %s\n", index, value.String())
	}
	// 根据键获取指定的 cookie
	t.Logf("PHPSESSID对应的cookie值：%s\n", resp.GetCookie("PHPSESSID"))
}

// 文件下载
func TestRequest_Down(t *testing.T) {
	cli := goCurl.CreateCurlClient()
	_, err := cli.Down("http://139.196.101.31:2080/GinSkeleton.jpg", "./", "ginskeleton.jpg", goCurl.Options{
		Timeout: 60.0,
	})
	if err == nil {
		t.Log("下载完成，请检查指定的下载目录")
	} else {
		t.Errorf("单元测试失败,文件下载失败，相关错误：%s", err.Error())
	}
}

//  https 以及 表单参数
//  get请求参数如果不是特别长，建议和地址拼接在一起请求,例如： https://www.oschina.net/search?scope=project&q=golang
//  如果参数比较长，您也可以按照表单参数提交
func TestRequest_Get_withQuery_arr(t *testing.T) {
	cli := goCurl.CreateCurlClient()
	//  cli.Get 切换成 cli.Post 就是 post 方式提交表单参数
	resp, err := cli.Get("https://www.oschina.net/search", goCurl.Options{
		FormParams: map[string]interface{}{
			"random": 12345,
			"scope":  "project",
			"q":      "golang",
		},
	})
	if err != nil {
		t.Errorf("osChina请求出错：%s\n", err.Error())
	}
	txt, err := resp.GetContents()
	if err == nil {
		t.Logf("请求结果：%s\n", txt)
	} else {
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}
}

//  post提交 json 数据
//  注意：这里的 header 头字段 Content-Type 必须设置为 json 格式
func TestRequest_Post_withJSON(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Post("http://rap2.taobao.org:38080/app/mock/243176/post_demo/json", goCurl.Options{
		Headers: map[string]interface{}{
			"Content-Type": "application/json",
		},
		JSON: struct {
			Code int      `json:"code"`
			Msg  string   `json:"msg"`
			Data []string `json:"data"`
		}{200, "OK", []string{"hello", "world"}},
	})
	if err != nil {
		log.Fatalln(err)
	}

	txt, err := resp.GetContents()
	if err == nil {
		t.Logf("请求结果：%s\n", txt)
	} else {
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}
}

// 设置代理ip访问目标站点

func TestRequest_Get_withProxy(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Get("http://139.196.101.31:20203/", goCurl.Options{
		Timeout: 5.0,
		Proxy:   "http://39.96.11.196:3111",
	})
	if err != nil {
		log.Fatalln(err)
	}
	txt, err := resp.GetContents()
	if err == nil {
		t.Logf("请求结果：%s\n", txt)
	} else {
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}
}
func TestSocksproxy(t *testing.T) {
	urli := url.URL{}
	//设置一个http代理服务器格式
	urlproxy, _ := urli.Parse("http://39.96.11.196:3211")
	//设置一个http客户端
	client := &http.Client{
		Transport: &http.Transport{ //设置代理服务器
			Proxy: http.ProxyURL(urlproxy),
		},
	}
	//访问地址http://myip.top
	rqt, err := http.NewRequest("GET", "http://myip.top", nil)
	if err != nil {
		println("接口获取IP失败!")
		return
	}
	//添加一个识别信息
	rqt.Header.Add("User-Agent", "Lingjiang")
	//处理返回结果
	response, _ := client.Do(rqt)
	defer response.Body.Close()
	//读取内容
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	//显示获取到的IP地址
	fmt.Println("socks5:", string(body))
	return

}

// 提交cookie
func TestRequest_Post_withCookies_str(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Post("http://127.0.0.1:8091/post-with-cookies", goCurl.Options{
		Cookies: "cookie1=value1;cookie2=value2",
	})
	if err != nil {
		log.Fatalln(err)
	}

	txt, err := resp.GetContents()
	if err == nil {
		t.Logf("请求结果：%s\n", txt)
	} else {
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}
}

// 提交cookie , 并从 body 体读取返回值
func TestRequest_Post_withCookies_map(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Post("http://101.132.69.236/api/v2/test_network", goCurl.Options{
		Cookies: map[string]string{
			"cookie1": "value1",
			"cookie2": "value2",
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	body := resp.GetBody()
	defer func() {
		_ = body.Close()
	}()
	if bytes, err := ioutil.ReadAll(body); err == nil {
		t.Logf("%s", bytes)
	} else {
		t.Errorf("单元测试失败，错误明细：%s\n", err.Error())
	}
}

//  Put 方式提交数据
func TestRequest_Put(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Put("http://127.0.0.1:8091/put")
	if err != nil {
		log.Fatalln(err)
	}

	txt, err := resp.GetContents()
	if err == nil {
		t.Logf("请求结果：%s\n", txt)
	} else {
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}
}

//  Delete方式提交数据
func TestRequest_Delete(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Delete("http://127.0.0.1:8091/delete")
	if err != nil {
		log.Fatalln(err)
	}
	txt, err := resp.GetContents()
	if err == nil {
		t.Logf("请求结果：%s\n", txt)
	} else {
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}
}
