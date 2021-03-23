package test

import (
	"github.com/qifengzhang007/goCurl"
	"io/ioutil"
	"log"
	"testing"
)

//  get 网站编码为 gbk
// 主要测试 get 请求以及自动转换被采集网站的编码，保证返回的数据是正常的
func TestRequest_Get(t *testing.T) {
	cli := goCurl.CreateHttpClient()
	resp, err := cli.Get("http://hq.sinajs.cn/list=sh601006")
	if err != nil && resp == nil {
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}
	if err != nil {
		t.Errorf("请求出错：%s\n", err.Error())
	} else {
		txt, err := resp.GetContents()
		if err == nil {
			t.Logf("请求结果：%s\n", txt)
		} else {
			t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
		}
	}
}

//  https 以及 表单参数
//  get请求参数如果不是特别长，建议和地址拼接在一起请求,例如： https://www.oschina.net/search?scope=project&q=golang
//  如果参数比较长，您也可以按照表单参数方式提交
func TestRequest_Get_withQuery_arr(t *testing.T) {
	cli := goCurl.CreateHttpClient()
	//  cli.Get 切换成 cli.Post 就是 post 方式提交表单参数
	//resp, err := cli.Post("http://127.0.0.1:8091/postWithFormParams", goCurl.Options{
	resp, err := cli.Get("https://www.oschina.net/search", goCurl.Options{
		FormParams: map[string]interface{}{
			"random": 12345,
			"scope":  "project",
			"q":      "golang",
		},
	})
	if err != nil {
		t.Errorf("osChina请求出错：%s\n", err.Error())
	} else {
		txt, err := resp.GetContents()
		if err == nil {
			t.Logf("请求结果：%s\n", txt)
		} else {
			t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
		}
	}

}

//  post提交 json 数据
//  注意：这里的 header 头字段 Content-Type 必须设置为 application/json 格式
func TestRequest_Post_withJSON(t *testing.T) {
	cli := goCurl.CreateHttpClient()

	resp, err := cli.Post("http://127.0.0.1:8091/post-with-json", goCurl.Options{
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
		t.Errorf("请求出错：%s\n", err.Error())
	} else {
		txt, err := resp.GetContents()
		if err == nil {
			t.Logf("请求结果：%s\n", txt)
		} else {
			t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
		}
	}
}

// 设置代理ip访问目标站点
// 测试期间我们使用了 http://http.taiyangruanjian.com/ 代理站点提供的每天免费试用ip
// 但是试用之前需要注册注册 [用户名] ,然后将 [用户名]以及您的外网ip添加至白名单才可以试用它们的代理，添加白名单地址：http://120.55.162.147/addlongip?username=用户名&white=需要添加的ip

func TestRequest_Get_withProxy(t *testing.T) {
	cli := goCurl.CreateHttpClient()

	resp, err := cli.Get("http://myip.top/", goCurl.Options{
		Timeout: 5.0,
		Proxy:   "http://39.96.11.196:3211", // 该ip需要自己去申请每日免费试用
	})
	if err != nil {
		t.Errorf("请求出错：%s\n", err.Error())
	} else {
		txt, err := resp.GetContents()
		if err == nil {
			// Proxy 参数设置 和 取消 您在这里将会看见不同的返回ip
			t.Logf("请求结果：%s\n", txt)
		} else {
			t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
		}
	}
}

// 文件下载
// 参数一 > 要下载的资源地址
// 参数二 > 指定下载路径（服务器最好指定绝对路径）
// 参数三 > 文件名，如果不设置，那么自动使用被下载的原始文件名
func TestRequest_Down(t *testing.T) {
	cli := goCurl.CreateHttpClient()
	_, err := cli.Down("http://139.196.101.31:2080/GinSkeleton.jpg", "./", "ginskeleton.jpg", goCurl.Options{
		Timeout: 60.0,
	})
	if err == nil {
		t.Log("下载完成，请检查指定的下载目录")
	} else {
		t.Errorf("单元测试失败,文件下载失败，相关错误：%s", err.Error())
	}
}

// 获取 cookie
func TestRequest_GetCookies(t *testing.T) {
	cli := goCurl.CreateHttpClient()
	resp, err := cli.Get(`http://www.iwencai.com/diag/block-detail?pid=10751&codes=600422&codeType=stock&info={"view":{"nolazy":1}}`)

	if err != nil {
		t.Errorf("采集同花顺站点发生错误：%s\n", err.Error())
	} else {
		// 全量获取cookie
		for index, value := range resp.GetCookies() {
			t.Logf("序号：%d, %s\n", index, value.String())
		}
		// 根据键获取指定的 cookie
		t.Logf("PHPSESSID对应的cookie值：%s\n", resp.GetCookie("PHPSESSID"))
	}

}

// 提交cookie
func TestRequest_Post_withCookies_str(t *testing.T) {
	cli := goCurl.CreateHttpClient()

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

// 提交cookie（二） , 并从 body 体读取返回值（）
func TestRequest_Post_withCookies_map(t *testing.T) {
	cli := goCurl.CreateHttpClient()

	resp, err := cli.Post("http://127.0.0.1:8091/post-with-cookies", goCurl.Options{
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
	// 如果请求的返回结果是从body体读取的二进制数据，必须使用 body.Close()  函数关闭
	// 此外必须注意的是，该函数是直接从缓冲区获取的二进制，对方的编码类型如果有中文（gbk系列）就会是乱码,需要自己转换，转换代码参见 getContents（） 函数
	if bytes, err := ioutil.ReadAll(body); err == nil {
		t.Logf("%s", bytes)
	} else {
		t.Errorf("单元测试失败，错误明细：%s\n", err.Error())
	}
}

//  Put 方式提交数据
func TestRequest_Put(t *testing.T) {
	cli := goCurl.CreateHttpClient()

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
	cli := goCurl.CreateHttpClient()

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
