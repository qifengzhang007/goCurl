## goCurl

- 基于goz改造，感谢 `原作者（github.com/idoubi/goz.git）`提供了大量、优质的基础代码.  
- 相比原版变化：
>   1.增加了文件下载功能  
>   2.GetBody() 返回io.ReaderCloser ,而不是原版本中的文本格式数据,今后该函数将专门负责处理流式数据，因此代码逻辑处理完毕，必须使用io.ReaderCloser 接口提供的函数关闭。  
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
	cli := goCurl.CreateCurlClient()

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
	}
	txt, err := resp.GetContents()

```
## 使用详情测试用例  
[测试用例文档](./test/request_test.go)  
