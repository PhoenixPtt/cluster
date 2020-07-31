// "node.go" file is create by Huxd 2020.07.13
// it used to init and due node operation

package router

import (
	header "clusterHeader"
	"fmt"
	"github.com/gin-gonic/gin"
)

// 节点（也就是Agent端）操作相关内容的具体处理函数 /agent
func initNodeRouter(group *gin.RouterGroup) bool {
	// Get 相关命令
	group.GET("/resource", getNodeResource)

	// Post 相关命令


	// Options 相关命令


	return true
}

// 获取节点资源的信息
func getNodeResource(c *gin.Context) {
	// 解析参数内容，确定是否时连续获取操作
	continueFlag := c.DefaultQuery("continue", "true")

	// 如果连续获取标识为true，则调用连续获取方法
	if continueFlag == "true" {
		// 持续获取集群资源监控数据，并及时返回给前端
		continueToGet(c, fmt.Sprintf("/%v/%v", header.FLAG_NODE, header.FLAG_NODE))
	} else { // 否则仅进行一次数据获取
		// 生成请求信息结构体
		req := requestInf{
			typeFlag: header.FLAG_NODE,
			opertype: header.FLAG_NODE,
		}
		// 获取单次Get信息
		onceToGet(c, req)
	}
}