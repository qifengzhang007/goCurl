package goCurl

// 可能的错误常量值
const (
	// 程序没有识别到对方站点编码类型，无法自动转码，出错，需要手动设置正在请求的站点编码值
	charsetDecoderError = `程序没有自动检测到对方站点响应 Header["Content-Type"] 编码类型,请创建客户端时手动指定 options 的参数 SetResCharset 值，可选值(GB18030、GBK、utf-8等)`
	// 代理设置错误
	proxyError = "代理设置错误，一般是设置的代理ip地址不可用 "
	// 下载的目标文件内容为空
	downloadFileIsEmpty = "被下载的文件内容为空"
)
