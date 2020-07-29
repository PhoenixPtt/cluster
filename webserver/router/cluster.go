// "cluster.go" file is create by Huxd 2020.07.13
// about cluster operation

package router

import (
	"fmt"

	header "clusterHeader"

	"github.com/gin-gonic/gin"
)

// 集群操作相关内容的具体处理函数 /cluster
func initClusterRouter(group *gin.RouterGroup) bool {
	// 获取集群资源信息以及相关的监控信息
	group.GET("/resource", func(c *gin.Context) {
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
	})

	return true
}
