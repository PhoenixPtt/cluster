// "cluster.go" file is create by Huxd 2020.07.13
// about cluster operation

package router

import (
	header "clusterHeader"
	"fmt"

	"github.com/gin-gonic/gin"
	myjwt "webserver/router/jwt"
)

// 集群操作相关内容的具体处理函数 /cluster
func initClusterRouter(group *gin.RouterGroup) bool {
	// 使用token验证中间件
	group.Use(myjwt.JWTAuth())

	// Get 相关命令
	group.GET("/resource", getClusterResource)

	// Post 相关命令

	// Options 相关命令
	//group.OPTIONS("/resource", onceToOption)

	return true
}

// 获取集群的资源信息，以及相关的监控信息
func getClusterResource(c *gin.Context) {
	// 解析参数内容，默认是连续获取集群的监控信息
	continueFlag := c.DefaultQuery("continue", "true")

	// 如果连续获取标识为true，则调用连续获取方法
	if continueFlag == "true" {
		// 持续获取集群资源监控数据，并及时返回给前端
		continueToGet(c, fmt.Sprintf("/%v/%v", header.FLAG_CLST, header.FLAG_CLST))
	} else { // 否则仅进行一次数据获取
		// 生成请求信息结构体
		req := requestInf{
			typeFlag: header.FLAG_CLST,
			opertype: header.FLAG_CLST,
		}
		// 获取单次Get信息
		onceToGet(c, req)
	}
}
