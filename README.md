###  goCurl

> 基于goz改造，感谢 `原作者（github.com/idoubi/goz.git）`提供了大量、优质的基础代码.
> 相比原版变化：
-  1.增加了文件下载功能
-   2.GetBody() 返回io.ReaderCloser ,而不是原版本中的文本格式数据,今后该函数将专门负责处理流式数据，因此代码逻辑处理完毕，必须使用io.ReaderCloser 接口提供的函数（Close）关闭。
-   3.原版本的GetBody()被现有版本GetContents()代替，由于文本数据,一次性返回，因此不需要手动关闭，程序会自动释放相关io资源。
-   4.删除、简化了原版本中为处理数据类型转换而定义的ResponseBody,本版本中使用系统系统默认的数据类型转换即可，简单快捷。
-   5.增强原版本中表单参数只能传递string、[]string的问题，该版本支持数字、文本、[]string等。
-   6.增加请求时浏览器自带的默认参数，完全模拟浏览器发送数据。
-   7.增加被请求的网站数据编码自动转换功能（采集网站时不需要考虑对方是站点的编码类型，gbk系列、utf8全程自动转换）。
-   8.增加获取服务端设置的 `cookie` 功能。
-   9.升级 `GET` 请求,提交表单参数时相关语法规范为与 `POST` 一致.
-   10.从实用主义出发，我们重新编写了使用文档，复制文档中的代码即可快速解决业务问题.
-   11.增加简体中文与utf-8编码互转函数,不管是发送还是接受都随意对字符进行编码转换.
-   12.增加 `XML` 格式数据提交，方便对接java类语言开发的 `webservice` 接口.
-   13.创建 `httpClient` 对象时使用 `sync.pool` 临时对象池,使客户端的创建更加高效,服务器资源占用更低,满足开发者频繁创建客户端采集数据.
-   14.增加 `sse` 客户端, 用于支持、处理 `h5` 推出的 `sse` 数据推送技术(例如：ChatGpt 接口的 stream 请求方式) .

### 安装 goCurl 包  
```code 

# 请自行在tag标签检查最新版本，本次使用 v1.4.0
go  get github.com/qifengzhang007/goCurl@v1.4.0

```

###  快速入门  
```code
    // step 1 :  创建 httpClient 客户端
	cli := goCurl.CreateHttpClient()

    // step 2 ： 请求并获取结果
	resp, err := cli.Get("https://www.baidu.com")
	fmt.Printf("请求参数：%v\n", resp.GetRequest())

	if err != nil {
		fmt.Printf("单元测试失败,错误明细：%s\n", err.Error())
	}else{
		txt, err := resp.GetContents()
       fmt.Println("响应结果：",txt)
    }

    // 其他可选参数选项,请结合语法篇理解与使用即可
    // step 3 ： 设置请求参数选项(非必选参数)
    type Options struct {
        Headers    map[string]interface{}
        BaseURI    string
        FormParams map[string]interface{}
        JSON       interface{}
        Timeout    float32   // 超时时间，单位：秒， 如果不设置或者设置为 0 表示程序一直等待不自动中断
        Cookies    interface{}
        Proxy      string
        // 如果请求的站点响应头  Header["Content-Type"]  中没有明确的 charset=utf-8 、charset=gb2312 等
        // 则程序无法自动转换，会给出错误提示，需要创建客户端时手动设置对方站点编码，例如：设置 SetResCharset 为 utf-8 、GB18030（gbk 和 gb2312 的超集） 等
        SetResCharset string   
        XML           string    // xml 数据，最终按照文本格式发送出去
    }

```
### [点击查看基本语法详情](./test/request_test.go)     
> 基本语法涵盖了 get 、post 、xml、json等市面上绝大部分常用场景的接口数据提交测试用例.





### 爬虫核心-高级实战部分
>   1.我们基于 `goCurl` 编写一段互联网数据采集代码.  
>   2.要求必须是并发采集,支持控制并发量.本段代码的并发量是持续并发量.        
>   3.数据采集使用我们提供的如下代码，你会发现一切是如此简洁与高效.            
     
```code 

// 本次采集数据的股票交易主要数据地址如下( 以60和00开头，其他的暂时不涉及,例如 300、688等）
//http://hq.sinajs.cn/list=sh601006
//http://hq.sinajs.cn/list=sz002414

func TestSpiderStock(t *testing.T){
	var StockUri="/list="
	var stockList=[]string{"601606","600522","002414","600151","000725","601238","002812","601360"}

	var wg sync.WaitGroup
	var conCurrentChan =make(chan int,4)   //  通道缓冲容量设置4，用于控制并发始终保持在4个

	wg.Add(len(stockList))
	// 采集数据逻辑开始
	for index,value :=range stockList{
		//使用通道控制并发，由于通道并发设置的是4，所以启动的协程最多只能有4个，超过4个就会阻塞，一直等到前面的协程执行完毕，才会继续追加新的任务
		conCurrentChan<-index
		// 每一个协程执行完毕，记得释放相关资源
		go func(stockCode string) {
			defer func() {
				wg.Done()
				<-conCurrentChan
			}()
		// 这里开始写采集数据的逻辑，我们本次是从新浪的股票接口抓取股票当日的主要成交信息
		var tmpUri string
		if strings.HasPrefix(stockCode,"60"){
			tmpUri=StockUri+"sh"+stockCode
		}else{
			tmpUri=StockUri+"sz"+stockCode
		}
		httpClient:=goCurl.CreateHttpClient(goCurl.Options{
				Headers: map[string]interface{}{
				"Referer": "http://vip.stock.finance.sina.com.cn",
				},
				SetResCharset: "GB18030",
				BaseURI:       "http://hq.sinajs.cn",
				})
			if response,err:=httpClient.Get(tmpUri);err==nil{
				//将结果的处理拆分为独立的函数，保持采集逻辑简洁
				if content,err:=response.GetContents();err==nil{
					HandleResponse(value,content)
				}else{
					fmt.Printf("获取请求结果出错：%s\n",err.Error())
				}
			}else{
				fmt.Printf("请求出错：%s\n",err.Error())
			}
		}(value)
	}

	wg.Wait()
	close(conCurrentChan)

}
// 处理采集结果, 根据传递的参数和对应结果进行业务逻辑处理  
func HandleResponse(code,content string){
	fmt.Printf("%s, %s\n",code,content)
}

```  

>结果如下：
```code 
601606, var hq_str_sh601606="长城军工,16.660,16.220,17.020,17.540,16.43  ...   ...
                                                               
002414, var hq_str_sz002414="高德红外,38.380,38.370,40.400,40.840,38.3 ...   ...
                                                               
00522, var hq_str_sh600522="中天科技,11.230,11.300,11.630,11.670,11.2 ...   ...
                                                               
600151, var hq_str_sh600151="航天机电,8.810,8.560,8.540,9.120,8.480,8. ...   ...
                                                               
000725, var hq_str_sz000725="京东方Ａ,5.530,5.500,5.570,5.600,5.490,5. ...   ...
                                                               
601238, var hq_str_sh601238="广汽集团,10.320,10.380,10.690,10.760,10.3 ...   ...
                                                               
601360, var hq_str_sh601360="三六零,18.790,18.680,18.650,18.950,18.560 ...   ...
                                                               
002812, var hq_str_sz002812="恩捷股份,78.660,78.210,83.350,84.190,78.2 ...   ...

```


### 避坑指南 
- 1.关于 `goCurl` 包自动解析被采集的网站编码为 `utf-8` 、`GB18030（GBK 和 GB2312 的超集）`  说明
- 2.我们以新浪的某个页面地址 (http://hq.sinajs.cn/list=sh601006) 为例，F12查看基本的响应格式：
```code   
Cache-Control: no-cache
Connection: Keep-Alive
Content-Encoding: gzip
Content-Length: 169
Content-Type: application/javascript; charset=GB18030
```
- 3.本包自动解析对方站点编码类型主要是根据以上响应头中的键：`Content-Type: text/html;charset=utf-8` ,自动查找 `charset` 对应的值,如果对方站点响应不完整，则会提示相关错误，需要在采集数据前人工确认对方站点编码类型，手动设置 `options` 参数 .
- 4.手动设置对方站点编码类型的示例语法
```code   

	cli := goCurl.CreateHttpClient()

	resp, err := cli.Post("http://www.webxml.com.cn/WebServices/ChinaZipSearchWebService.asmx/getSupportCity", goCurl.Options{
		Headers: map[string]interface{}{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		FormParams: map[string]interface{}{
			"byProvinceName": "重庆", // 参数选项：上海、北京、天津、重庆 等。这个接口在postman测试有时候也是很稳定，可以更换参数多次测试
		},
		# 这里手动设置对方站点编码类型为 utf-8，其他可选项：GB18030 （GBK 和 GB2312 请使用 GB18030 代替）
		SetResCharset: "utf-8",
		Timeout:       10,
	})

```
