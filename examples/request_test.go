package goCurl

import (
	"fmt"
	"github.com/qifengzhang007/goCurl"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

//  get 网站编码为 gbk
// 主要测试 get 请求以及自动转换被采集网站的编码
func TestRequest_Get(t *testing.T) {
	cli := goCurl.CreateCurlClient()
	resp, err := cli.Get("http://hq.sinajs.cn/list=sh601006")
	if err != nil {
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
//  get 请求参数如果不是特别长，建议和地址拼接在一起请求,例如： https://www.oschina.net/search?scope=project&q=golang
//  如果参数比较长，您也可以按照表单参数提交
func TestRequest_Get_withQuery_arr(t *testing.T) {
	cli := goCurl.CreateCurlClient()
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

func TestRequest_Get_withProxy(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Get("https://www.fbisb.com/ip.php", goCurl.Options{
		Timeout: 5.0,
		Proxy:   "http://127.0.0.1:1087",
	})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(resp.GetStatusCode())
	// Output: 200
	fmt.Println(resp.GetContents())
	// Output: 116.153.43.128
}

func TestRequest_Post(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Post("http://127.0.0.1:8091/post")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%T", resp)
	// Output: *goCurl.Response
}

func TestRequest_Post_withHeaders(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Post("http://127.0.0.1:8091/post-with-headers", goCurl.Options{
		Headers: map[string]interface{}{
			"User-Agent": "testing/1.0",
			"Accept":     "application/json",
			"X-Foo":      []string{"Bar", "Baz"},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	headers := resp.GetRequest().Header["X-Foo"]
	fmt.Println(headers)
	// Output: [Bar Baz]
}

func TestRequest_Post_withCookies_str(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Post("http://127.0.0.1:8091/post-with-cookies", goCurl.Options{
		Cookies: "cookie1=value1;cookie2=value2",
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%d", resp.GetContentLength())
	//Output: 385
}

func TestRequest_Post_withCookies_map(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	//resp, err := cli.Post("http://127.0.0.1:8091/post-with-cookies", goCurl.Options{
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
	bytes, _ := ioutil.ReadAll(body)
	fmt.Printf("%s", bytes)
	// Output: {"code":200,"msg":"OK","data":""}
}

func TestRequest_Post_withCookies_obj(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	cookies := make([]*http.Cookie, 0, 2)
	cookies = append(cookies, &http.Cookie{
		Name:     "cookie133",
		Value:    "value1",
		Domain:   "httpbin.org",
		Path:     "/cookies",
		HttpOnly: true,
	})
	cookies = append(cookies, &http.Cookie{
		Name:   "cookie2",
		Value:  "value2",
		Domain: "httpbin.org",
		Path:   "/cookies",
	})

	resp, err := cli.Post("http://127.0.0.1:8091/post-with-cookies", goCurl.Options{
		Cookies: cookies,
	})
	if err != nil {
		log.Fatalln(err)
	}

	body := resp.GetBody()
	fmt.Printf("%T", body)
	//Output: *http.cancelTimerBody
}
func TestRequest_SimplePost(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Post("http://101.132.69.236/api/v2/test_network", goCurl.Options{
		Headers: map[string]interface{}{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		FormParams: map[string]interface{}{
			"key1": "value1",
			"key2": []string{"value21", "value22"},
			"key3": "333",
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	contents, err := resp.GetContents()
	if err != nil {
		t.Errorf("TestRequest_SimplePost,单元测试失败,相关错误：%s", err.Error())
	} else {
		t.Log(contents)
	}
}

func TestRequest_Get_withFormParams(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Get("http://127.0.0.1:20191/api/v1/home/news", goCurl.Options{
		Headers: map[string]interface{}{
			"Content-Type": "application/x-www-form-urlencoded;charset=utf-8",
		},
		FormParams: map[string]interface{}{
			"newsType": "portal3",
			"page":     "2",
			"limit":    "52",
		},
	})
	if err != nil {
		t.Errorf("TestRequest_Post_withFormParams 单元测试失败，相关错误：%s\n", err.Error())
	}

	text, err := resp.GetContents()
	if err == nil {
		t.Logf("%v", text)
	} else {
		t.Errorf("TestRequest_Post_withFormParams 获取请求结果失败，相关错误：%s\n", err.Error())
	}
}

func TestRequest_Post_withFormParams(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Post("http://127.0.0.1:8091/post-with-form-params", goCurl.Options{
		Headers: map[string]interface{}{
			"Content-Type": "application/x-www-form-urlencoded;charset=utf-8",
		},
		FormParams: map[string]interface{}{
			"key1": 2020,
			"key2": []string{"value21", "value22"},
			"key3": "abcd张",
		},
	})
	if err != nil {
		t.Errorf("TestRequest_Post_withFormParams 单元测试失败，相关错误：%s\n", err.Error())
	}

	text, err := resp.GetContents()
	if err == nil {
		t.Logf("%v", text)
	} else {
		t.Errorf("TestRequest_Post_withFormParams 获取请求结果失败，相关错误：%s\n", err.Error())
	}

}

func TestRequest_Post_withJSON(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Post("http://127.0.0.1:8091/post-with-json", goCurl.Options{
		Headers: map[string]interface{}{
			"Content-Type": "application/json",
		},
		JSON: struct {
			Key1 string   `json:"key1"`
			Key2 []string `json:"key2"`
			Key3 int      `json:"key3"`
		}{"value1", []string{"value21", "value22"}, 333},
	})
	if err != nil {
		log.Fatalln(err)
	}

	body := resp.GetBody()
	defer func() {
		_ = body.Close()
	}()
	fmt.Printf("%T", body)
	// Output:  *http.cancelTimerBody
}

func TestRequest_Put(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Put("http://127.0.0.1:8091/put")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%T", resp)
	// Output: *goCurl.Response
}

func TestRequest_Patch(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Patch("http://127.0.0.1:8091/patch")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%T", resp)
	// Output: *goCurl.Response
}

func TestRequest_Delete(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Delete("http://127.0.0.1:8091/delete")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%T", resp)
	// Output: *goCurl.Response
}

func TestRequest_Options(t *testing.T) {
	cli := goCurl.CreateCurlClient()

	resp, err := cli.Options("http://127.0.0.1:8091/options")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%T", resp)
	// Output: *goCurl.Response
}
