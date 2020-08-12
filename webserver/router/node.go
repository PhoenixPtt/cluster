// "node.go" file is create by Huxd 2020.07.13
// it used to init and due node operation

package router

import (
	header "clusterHeader"
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
	// 解析参数内容，获取指定的节点名称或IP或标志
	nodeID := c.DefaultQuery("name", "")

	// 生成请求信息结构体
	req := requestInf{
		typeFlag: header.FLAG_NODE,
		opertype: header.FLAG_NODE,
	}

	// 如果为空，则获取全部节点的资源
	if nodeID == "" {

	} else { // 否则获取指定节点的资源

	}

	// 获取单次Get信息
	onceToGet(c, req)
}
