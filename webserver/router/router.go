package router

import (
	"fmt"
	_ "fmt"
	"github.com/gin-gonic/gin"
	"io"
	"webserver/router/errcode"
)

func Init(rout *gin.Engine) bool {
	// 采样频率相关功能
	if !initFreqRouter(rout.Group("/frequency")) {
		return false
	}

	// 镜像操作的相关功能
	if !initImageRouter(rout.Group("/image")) {
		return false
	}

	// 文件操作相关功能
	if !initFileRouter(rout.Group("/file")) {
		return false
	}

	// 集群操作相关功能
	if !initClusterRouter(rout.Group("/cluster")) {
		return false
	}

	// 节点操作相关功能
	if !initNodeRouter(rout.Group("/agent")) {
		return false
	}

	// 应用服务操作相关功能
	if !initApplicationServiceRouter(rout.Group("/deployment")) {
		return false
	}

	// 定义无法找到指定路由的情况下，返回的错误信息
	rout.NoRoute(func(c *gin.Context) {
		errcode.ServeJSON(c, errcode.ErrorCodeNotfound.WithDetail(fmt.Sprintf("URL:%v is not found", c.Request.URL.Path)))
	})

	return true
}

// 采样频率相关内容的具体处理函数 /frequency
func initFreqRouter(group *gin.RouterGroup) bool {
	// 获取采集频率的当前值
	group.GET("/current", func(c *gin.Context) {
		simpleToGet(c, group.BasePath()+"/current")
	})
	return true
}

// 镜像操作相关内容的具体处理函数 /image
func initImageRouter(group *gin.RouterGroup) bool {
	return true
}

// 文件操作相关内容的具体处理函数 /file
func initFileRouter(group *gin.RouterGroup) bool {
	return true
}

// 应用服务操作相关内容的具体处理函数 /deployment(暂定)
func initApplicationServiceRouter(group *gin.RouterGroup) bool {
	return true
}

// 添加解决跨域问题的请求头
func addAccessControlAllowOrigin(c *gin.Context) {
	// 注意:在前后端分离过程中，需要注意跨域问题，因此需要设置请求头
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
}

// 处理Get命令中连续获取指定URL的内容
func continueToGet(c *gin.Context, msgType string){
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
			c.SSEvent("message", message.content)
			return true
		}
	})
}

// 单次Get请求指定URL的内容
func simpleToGet(c *gin.Context, msgType string) {
	// 添加跨域头
	addAccessControlAllowOrigin(c)

	// 创建通道，等待通道数据返回，将该通道的信息返回给前端
	bOK, msg := getMessage(msgType)

	if bOK {
		c.JSON(200, msg.content)
	} else {
		errcode.ServeJSON(c, msg.errorMsg)
	}
}
