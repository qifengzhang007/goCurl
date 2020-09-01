## goCurl

- 基于goz改造，感谢 `原作者（github.com/idoubi/goz.git）`提供了大量、优质的基础代码.  
- 相比原版变化：
>   1.增加了文件下载功能  
>   2.GetBody() 返回io.ReaderCloser ,而不是原版本中的文本格式数据,今后该函数将专门负责处理流式数据，因此代码逻辑处理完毕，必须使用io.ReaderCloser 接口提供的函数（Close）关闭。  
>   3.原版本的GetBody()被现有版本GetContents()代替，由于文本数据,一次性返回，因此不需要手动关闭，程序会自动释放相关io资源。  
>   4.删除、简化了原版本中为处理数据类型转换而定义的ResponseBody,本版本中使用系统系统默认的数据类型转换即可，简单快捷。  
>   5.增强原版本中表单参数只能传递string、[]string的问题，该版本支持数字、文本、[]string等。  
>   6.增加请求时浏览器自带的默认参数，完全模拟浏览器发送数据。  
>   7.增加被请求的网站数据编码自动转换功能（采集网站时不需要考虑对方是站点的编码类型，gbk系列、utf8全程自动转换）。  
>   8.增加获取服务端设置的cookie功能。    
>   9.升级get请求,提交表单参数时相关语法规范为与post一致.   
>   10.从实用主义出发，我们重新编写了使用文档，复制文档中的代码即可快速用于业务中去.    

## 快速入门  
```code
    // step 1 :  创建 curl 客户端
	cli := goCurl.CreateHttpClient()

    // step 2 ： 设置请求参数选项(非必选参数)
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

    // step 3 ： 请求并获取结果
	resp, err := cli.Get("http://hq.sinajs.cn/list=sh601006")
	if err != nil  && resp==nil{
		t.Errorf("单元测试失败,错误明细：%s\n", err.Error())
	}else{
	txt, err := resp.GetContents()
    }

```
## 基本语法篇      
[进入详情](./test/request_test.go)

## 高级部分
>   1.我们基于 `goCurl` 编写一段互联网数据采集代码.  
>   2.要求必须是并发采集,支持控制并发量.本段代码的并发量是持续并发量.        
>   3.其实数据采集不需要各种爬虫框架，使用我们提供的如下代码，你会发现一切是如此简洁与高效.            
     
```code 

// 本次采集数据的股票交易主要数据地址如下( 以60和00开头，其他的暂时不涉及,例如 300、688等）
//http://hq.sinajs.cn/list=sh601006
//http://hq.sinajs.cn/list=sz002414

func TestSpiderStock(t *testing.T){
	var StockUri="http://hq.sinajs.cn/list="
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
		httpClient:=goCurl.CreateHttpClient()
			if response,err:=httpClient.Get(tmpUri);err==nil{
				//将结果的处理拆分为独立的函数，保持采集逻辑简洁
				if content,err:=response.GetContents();err==nil{
					HandleResponse(content)
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
// 处理采集结果
func HandleResponse(content string){
	//这里处理采集到的的结果文本数据
	fmt.Printf("%s\n",content)
}

```  

>结果如下：
```code 
var hq_str_sh601606="长城军工,16.660,16.220,17.020,17.540,16.43  ...   ...
                                                               
var hq_str_sz002414="高德红外,38.380,38.370,40.400,40.840,38.3 ...   ...
                                                               
var hq_str_sh600522="中天科技,11.230,11.300,11.630,11.670,11.2 ...   ...
                                                               
var hq_str_sh600151="航天机电,8.810,8.560,8.540,9.120,8.480,8. ...   ...
                                                               
var hq_str_sz000725="京东方Ａ,5.530,5.500,5.570,5.600,5.490,5. ...   ...
                                                               
var hq_str_sh601238="广汽集团,10.320,10.380,10.690,10.760,10.3 ...   ...
                                                               
var hq_str_sh601360="三六零,18.790,18.680,18.650,18.950,18.560 ...   ...
                                                               
var hq_str_sz002812="恩捷股份,78.660,78.210,83.350,84.190,78.2 ...   ...

```
