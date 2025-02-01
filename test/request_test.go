package test

import (
	"fmt"
	"github.com/qifengzhang007/goCurl"
	"io"
	"log"
	"testing"
)

//	get 网站编码为 gbk
//
// 主要测试 get 请求以及自动转换被采集网站的编码，保证返回的数据是正常的
func TestRequestGet(t *testing.T) {

	// 创建 http 客户端的时候可以直接填充一些公共参数，后续请求会复用
	cli := goCurl.CreateHttpClient(goCurl.Options{
		Headers: map[string]interface{}{
			"Referer": "http://vip.stock.finance.sina.com.cn",
		},
		SetResCharset: "GB18030",
		BaseURI:       "",
	})
	resp, err := cli.Get("http://hq.sinajs.cn/list=sz002594")
	//t.Logf("请求参数：%v\n", resp.GetRequest())
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

func TestRequestGet2(t *testing.T) {

	// 创建 http 客户端的时候可以直接填充一些公共参数，后续请求会复用
	cli := goCurl.CreateHttpClient(goCurl.Options{
		Headers: map[string]interface{}{
			"Connection": "keep-alive",
		},
	})
	for i := 1; i < 5; i++ {
		resp, err := cli.Get("http://49.232.145.118:20171/api/v1/portal/news?newsType=10&page=1&limit=50")
		//t.Logf("请求参数：%v\n", resp.GetRequest())
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

}

// https 以及 表单参数
// get请求参数如果不是特别长，建议和地址拼接在一起请求,例如： https://www.oschina.net/search?scope=project&q=golang&random=123215
func TestRequestGetWithQuery(t *testing.T) {
	cli := goCurl.CreateHttpClient()
	//  cli.Get 切换成 cli.Post 就是 post 方式提交表单参数
	//resp, err := cli.Post("http://127.0.0.1:8091/postWithFormParams", goCurl.Options{
	resp, err := cli.Get("https://www.oschina.net/search?scope=project&q=golang&random=123215", goCurl.Options{
		SetResCharset: "UTF-8",
		Headers: map[string]interface{}{
			"Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
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

// https 以及 表单参数
// get请求参数如果不是特别长，建议和地址拼接在一起请求,例如： https://www.oschina.net/search?scope=project&q=golang&random=123215
func TestRequestGetWithQuery2(t *testing.T) {
	targetUrl := `https://datacenter-web.eastmoney.com/api/data/v1/get`
	cli := goCurl.CreateHttpClient()
	resp, err := cli.Get(targetUrl, goCurl.Options{
		FormParams: map[string]interface{}{
			"sortColumns": "BILLBOARD_NET_AMT,TRADE_DATE,SECURITY_CODE",
			"sortTypes":   "-1,-1,1",
			"pageSize":    50,
			"pageNumber":  1,
			"reportName":  "RPT_DAILYBILLBOARD_DETAILSNEW",
			"columns":     "SECURITY_CODE,SECUCODE,SECURITY_NAME_ABBR,TRADE_DATE,EXPLAIN,CLOSE_PRICE,CHANGE_RATE,BILLBOARD_NET_AMT,BILLBOARD_BUY_AMT,BILLBOARD_SELL_AMT,BILLBOARD_DEAL_AMT,ACCUM_AMOUNT,DEAL_NET_RATIO,DEAL_AMOUNT_RATIO,TURNOVERRATE,FREE_MARKET_CAP,EXPLANATION,D1_CLOSE_ADJCHRATE,D2_CLOSE_ADJCHRATE,D5_CLOSE_ADJCHRATE,D10_CLOSE_ADJCHRATE,SECURITY_TYPE_CODE",
			"source":      "WEB",
			"client":      "WEB",
			"filter":      "(TRADE_DATE<='2024-11-06')(TRADE_DATE>='2024-11-06')",
		},
		SetResCharset: "utf-8",
		Headers: map[string]interface{}{
			"Content-Type": "application/x-www-form-urlencoded;charset=utf-8",
		},
	})
	if err != nil {
		t.Errorf("eastmoney 请求出错：%s\n", err.Error())
	} else {
		txt, err := resp.GetContents()
		if err == nil {
			t.Logf("请求结果：%s\n", txt)
		} else {
			t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
		}
	}

}

// GO 语言 UTF8 环境发送 简体中文数据
func TestRequestSendChinese(t *testing.T) {
	cli := goCurl.CreateHttpClient()
	resp, err := cli.Get("http://139.196.101.31:2080/test_json.php", goCurl.Options{
		FormParams: map[string]interface{}{
			//"user_name":"你好，该字段发送出去的数据为简体中文编码",  // 对方站点只接受 简体中文，这种不编码直接发出去就会报错
			"user_name": cli.Utf8ToSimpleChinese([]byte("该字段发送出去的数据为简体中文编码")), // 第二个参数：默认编码为 GB18030，（GBK 、GB18030 都是简体中文，go编码器中没有 gb2312）
		},
		//Headers: map[string]interface{}{
		//	"Content-Type": "application/x-www-form-urlencoded;charset=gb2312",
		//},
	})
	if err != nil {
		t.Errorf("发送简体中文测试出错：%s\n", err.Error())
	} else {
		txt, err := resp.GetContents()
		if err == nil {
			t.Logf("请求结果：%s\n", txt)
		} else {
			t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
		}
	}

}

// post提交 json 数据
// 注意：这里的 header 头字段 Content-Type 必须设置为 application/json 格式
func TestRequestPostWithJSON(t *testing.T) {
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

//	post向 webservice接口提交 xml 数据(以表单参数形式提交x-www-form-urlencoded)
//	webservice测试地址以及接口说明：http://www.webxml.com.cn/WebServices/ChinaZipSearchWebService.asmx/getSupportCity
//
// 浏览器打开以上地址，F12 可以查看webservice 接口以表单形式是如何发送数据的
func TestRequestPostFormDataWithXml(t *testing.T) {
	cli := goCurl.CreateHttpClient()

	resp, err := cli.Post("http://www.webxml.com.cn/WebServices/ChinaZipSearchWebService.asmx/getSupportCity", goCurl.Options{
		Headers: map[string]interface{}{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		FormParams: map[string]interface{}{
			"byProvinceName": "重庆", // 参数选项：上海、北京、天津、重庆 等。这个接口在postman测试有时候也是很稳定，可以更换参数多次测试
		},
		SetResCharset: "utf-8",
		Timeout:       10,
	})
	if err != nil {
		t.Errorf("请求出错：%s\n", err.Error())
	} else {
		txt, err := resp.GetContents()
		if err == nil {
			t.Logf("请求结果：\n%s\n", txt)
		} else {
			t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
		}
	}
}

// post向 webservice接口提交 xml 数据（以raw方式提交）
// webservice测试地址以及接口说明：http://www.webxml.com.cn/WebServices/ChinaZipSearchWebService.asmx
func TestRequestPostRawWithXml(t *testing.T) {
	cli := goCurl.CreateHttpClient(goCurl.Options{
		SetResCharset: "utf-8",
	})

	// 需要提交的 xml 数据格式，发送前请转换为以下文本格式
	// 正式业务我们的参数是动态的
	// 那么就事先需要定义好go语言的结构体，最终将绑定好参数的结构体转为xml格式数据
	// 关于结构体转 xml 格式代码参见：https://blog.csdn.net/f363641380/article/details/87651427
	xml := `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getSupportCity xmlns="http://WebXml.com.cn/">
      <byProvinceName>上海</byProvinceName>
    </getSupportCity>
  </soap:Body>
</soap:Envelope>
`

	resp, err := cli.Post("http://www.webxml.com.cn/WebServices/ChinaZipSearchWebService.asmx", goCurl.Options{
		Headers: map[string]interface{}{
			"Content-Type": "text/xml; charset=utf-8",
			"SOAPAction":   "http://WebXml.com.cn/getSupportCity", //  该参数按照业务方的具体要求传递
		},
		XML:     xml,
		Timeout: 20,
	})
	fmt.Printf("请求参数：%#+v\n", resp.GetRequest())
	if err != nil {
		t.Errorf("请求出错：%s\n", err.Error())
	} else {
		txt, err2 := resp.GetContents()
		if err2 == nil {
			t.Logf("请求结果:\n%v\n", txt)
		} else {
			t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
		}
	}
}

// 设置代理ip访问目标站点
// 测试期间我们使用了 http://http.taiyangruanjian.com/ 代理站点提供的每天免费试用ip
// 但是试用之前需要注册注册 [用户名] ,然后将 [用户名]以及您的外网ip添加至白名单才可以试用它们的代理，添加白名单地址：http://120.55.162.147/addlongip?username=用户名&white=需要添加的ip

func TestRequestGetWithProxy(t *testing.T) {
	cli := goCurl.CreateHttpClient()

	resp, err := cli.Get("http://myip.top/", goCurl.Options{
		Timeout: 60,
		Proxy:   "http://113.241.137.248:4330", // 该ip需要自己去申请每日免费试用
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
func TestRequestDown(t *testing.T) {
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
func TestRequestGetCookies(t *testing.T) {
	cli := goCurl.CreateHttpClient()
	resp, err := cli.Get(`https://www.baidu.com`)

	if err != nil {
		t.Errorf("采集百度首页cookie发生错误：%s\n", err.Error())
	} else {
		// 全量获取cookie
		for index, value := range resp.GetCookies() {
			t.Logf("序号：%d, %s\n", index, value.String())
		}
		// 根据键获取指定的 cookie
		t.Logf("BAIDUID 对应的cookie值：%s\n", resp.GetCookie("BAIDUID"))
	}

}

// 提交cookie
func TestRequestPostWithCookiesStr(t *testing.T) {
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
func TestRequestPostWithCookiesMap(t *testing.T) {
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
	if bytes, err := io.ReadAll(body); err == nil {
		t.Logf("%s", bytes)
	} else {
		t.Errorf("单元测试失败，错误明细：%s\n", err.Error())
	}
}

// Put 方式提交数据
func TestRequestPut(t *testing.T) {
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

// Delete方式提交数据
func TestRequestDelete(t *testing.T) {
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

// SseGet 通过sse客户端的get请求获取服务端持续推送的数据流
func TestRequestSseGet(t *testing.T) {
	sseServerUrl := "https://92.push2.eastmoney.com/api/qt/stock/details/sse?fields1=f1,f2,f3,f4&fields2=f51,f52,f53,f54,f55&mpi=2000&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&pos=-0&secid=1.600460&wbp2u=|0|0|0|web"
	cli := goCurl.CreateHttpClient()

	var options = goCurl.Options{
		Headers: map[string]interface{}{
			"Content-Type":  "text/event-stream",
			"Cache-Control": "no-cache",
			"Connection":    "keep-alive",
		},
		Timeout: -1,
	}
	// Sse 方法会阻塞目前的代码，如果需要异步接收处理sseClient收到的消息，请使用go协程启动该方法
	err := cli.Sse("get", sseServerUrl, func(msgType, content string) bool {

		switch msgType {
		case "event":
			// 事件类型的消息格式
			t.Logf("(event)事件类型的消息：\n%+v\n", content)
		case "data":
			// 数据类型的消息格式
			t.Logf("服务端推送的业务数据(data)：\n%+v\n", content)
		}

		// 这里是回调函数的返回值：
		// true  表示持续接受服务端的推送数据，
		// false 表示接受一次服务端的推送数据后，主动关闭客户端不在接受后续数据
		return true
	}, options)
	if err != nil {
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}

}

// Sse 通过sse客户端的post方式请求 chatgpt 服务器接口
func TestRequestSse(t *testing.T) {

	// 定义一个 chatgpt 聊天发送的结构体
	type chaGpt struct {
		Model    string `json:"model"`
		Stream   bool   `json:"stream"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	// chatGpt 聊天接口
	openaiChatUrl := "https://api.openai.com/v1/chat/completions"
	apiKey := "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" // 请填写自己的openai apikeuy，本次测试的很快就会失效
	cli := goCurl.CreateHttpClient()

	var chat = goCurl.Options{
		Headers: map[string]interface{}{
			"Authorization": "Bearer " + apiKey,
			"Content-Type":  "application/json",
			"Cache-Control": "no-cache",
			"Connection":    "keep-alive",
		},
		Timeout: -1,
		JSON: chaGpt{
			Model:  "gpt-3.5-turbo",
			Stream: true,
			Messages: []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				{Role: "user", Content: "请列出一份go语言面试常见的30个核心知识点"},
			},
		},
	}
	// Sse 方法会阻塞目前的代码，如果需要异步接收处理sseClient收到的消息，请使用go协程启动该方法
	err := cli.Sse("post", openaiChatUrl, func(msgType, content string) bool {
		//fmt.Printf("收到chatgpt原始事件：%+v\n", msgType)
		switch msgType {
		case "event":
			// 事件类型的消息格式
			t.Logf("(event)事件类型的消息：\n%+v\n", content)
		case "data":
			// 数据类型的消息格式,
			t.Logf("服务端推送的业务数据(data)：\n%+v\n", content)
		}

		// 这里是回调函数的返回值：
		// true  表示持续接受服务端的推送数据，
		// false 表示接受一次服务端的推送数据后，主动关闭客户端不在接受后续数据
		// 例如：在 chatGpt 聊天场景中，最后一行返回了 [DONE] ，如果返回的字符串为 [DONE]  直接return false 结束整个请求的生命周期
		return true
	}, chat)
	if err != nil {
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}

}
