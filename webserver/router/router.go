package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"webserver/router/errcode"
)

func Init(rout *gin.Engine) bool {
	// 采样频率相关功能
	initFreqRouter(rout.Group("/frequency"))

	// 镜像操作的相关功能
	initImageRouter(rout.Group("/image"))

	// 文件操作相关功能
	initFileRouter(rout.Group("/file"))

	// 集群操作相关功能
	initClusterRouter(rout.Group("/cluster"))

	// 节点操作相关功能
	initNodeRouter(rout.Group("/agent"))

	// 应用服务操作相关功能
	initApplicationServiceRouter(rout.Group("/deployment"))

	// 定义无法找到指定路由的情况下，返回的错误信息
	rout.NoRoute(func(c *gin.Context) {
		serveErrorJSON(c,
			errcode.ErrorCodeNotfound.WithDetail(fmt.Sprintf("URL:%v is not found", c.Request.URL.Path)))
	})

	return true
}

// 采样频率相关内容的具体处理函数 /frequency
func initFreqRouter(group *gin.RouterGroup) bool {
	// 获取采集频率的当前值
	group.GET("/current", func(c *gin.Context) {
		//// 解析请求主体内容
		//var data header.CLST
		//if err := c.ShouldBindJSON(&data); err != nil {
		//	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		//	return
		//}

		//req := requestInf{
		//	typeFlag: "frequency",
		//	opertype: "current",
		//}
		//
		//onceToGet(c, req)
	})
	return true
}

// 添加解决跨域问题的请求头
func addAccessControlAllowOrigin(c *gin.Context) {
	// 注意:在前后端分离过程中，需要注意跨域问题，因此需要设置请求头
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
}

// 返回错误信息
func serveErrorJSON(c *gin.Context, err errcode.Error) {
	// 添加跨域头
	addAccessControlAllowOrigin(c)

	// 返回客户端错误信息
	errcode.ServeJSON(c, err)
}

// 处理Get命令中连续获取指定URL的内容
func continueToGet(c *gin.Context, msgType string) {
	// 添加跨域头
	addAccessControlAllowOrigin(c)

	// 在Gin引入的sse扩展代码中，并未设置Header中Connection属性，所以在这里补充一下
	c.Writer.Header().Set("Connection", "keep-alive")

	// 获取信息管理者对象 [按需生成]
	m, msg := NewManager(msgType)

	// 启用获取节点资源的协程，获取结果后再返回信息
	// 获取链接关闭标志通道对象
	clientGone := c.Writer.CloseNotify()
	// 启动流发送函数，注意**Stream函数是循环发送**，直到CloseNotify标志被写入
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			// 当链接关闭时，删除本链接使用的客户对象
			m.delClient(msg)
			return false
		case message := <-msg:
			//log.Println("send message:", message.content)
			// 当有消息需要发送的时候，使用SSEvent函数来发送数据实体
			if message.content != "" {
				c.SSEvent("message", message.content)
			} else {
				c.SSEvent("message", message.errorMsg)
			}
			return true
		}
	})
}

// 单次Get请求指定URL的方法
func onceToGet(c *gin.Context, reqinfo requestInf) {
	// 添加跨域头
	addAccessControlAllowOrigin(c)

	// 创建通道，等待通道数据返回，将该通道的信息返回给前端
	bOK, msg := getMessage(reqinfo)

	// 如果获取通道数据成功，则返回实际数据，否则返回错误信息
	if bOK {
		c.JSON(200, msg.content)
	} else {
		errcode.ServeJSON(c, msg.errorMsg)
	}
}

// 单次Post请求指定URL的方法
func onceToPost(c *gin.Context, reqinfo requestInf) {
	//// 添加跨域头
	//addAccessControlAllowOrigin(c)
	//
	//// 创建通道，等待通道数据返回，将该通道的信息返回给前端
	//bOK, msg := getMessage(reqinfo)
	//
	//// 如果获取通道数据成功，则返回实际数据，否则返回错误信息
	//if bOK {
	//	c.JSON(200, msg.content)
	//} else {
	//	errcode.ServeJSON(c, msg.errorMsg)
	//}

	// 目前同单次Get请求的操作，暂时使用这样的方法进行操作
	onceToGet(c, reqinfo)
}

// 单次option请求
func onceToOption(c *gin.Context) {
	// 添加跨域头
	addAccessControlAllowOrigin(c)

	// 将预检请求的结果缓存10分钟 86400一天
	// Access-Control-Max-Age方法对完全一样的url的缓存设置生效，多一个参数也视为不同url
	// 也就是说，如果设置了10分钟的缓存，在10分钟内，所有请求第一次会产生options请求，以后就只发送真正的请求了
	c.Writer.Header().Set("Access-Control-Max-Age", "600")

	// 返回200，以及相关数据
	c.Data(200, "", []byte(""))
}

// 获取post发送过来的数据内容，一般作为调试使用，此时读取将导致无法再次读取本次POST的内容
func getPostContent(c *gin.Context)  string {
	bodyByte, _ := ioutil.ReadAll(c.Request.Body)
	body := string(bodyByte)
	return body
}
