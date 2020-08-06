// "application.go" file is create by Huxd 2020.07.27
// it used to init and due service operation.

package router

import (
	header "clusterHeader"
	"github.com/gin-gonic/gin"
)

// 应用服务操作使用的结构体
type deploymentData struct {
	name     string // 应用服务名称
	replicas uint8  // 应用服务的副本数量
}

// 应用服务操作相关内容的具体处理函数 /deployment(暂定)
func initApplicationServiceRouter(group *gin.RouterGroup) bool {
	// Get 相关命令
	group.GET("/resource", getDeploymentResource)
	group.GET("/start", startDeployment)
	group.GET("/stop", stopDeployment)
	group.GET("/restart", restartDeployment)

	// Post 相关命令
	group.POST("/create", createDeployment)
	group.POST("/replicas", configDeploymentReplicas)

	// Delete 相关命令
	group.DELETE("/delete", deleteDeployment)

	// 处理非Get时，可能进行的OPTION请求
	group.OPTIONS("/create", onceToOption)
	group.OPTIONS("/replicas", onceToOption)
	group.OPTIONS("/delete", onceToOption)

	return true
}

// 获取应用服务的资源信息
func getDeploymentResource(c *gin.Context) {
	// 解析参数内容，确定需要获取的应用服务是指定的还是全部的
	name := c.DefaultQuery("name", "")

	// 生成请求信息结构体
	req := requestInf{
		typeFlag: header.FLAG_SERV,
		opertype: header.FLAG_SERV,
	}

	// 如果指定的应用服务名称为空，则获取全部应用服务的资源信息
	if name == "" {

	} else { // 否则设定指定应用服务的名称

	}

	// 获取单次Get信息
	onceToGet(c, req)
}

// 启动应用服务，可指定应用服务名称和副本数量
func startDeployment(c *gin.Context) {

}

// 停止应用服务，可指定应用服务名称
func stopDeployment(c *gin.Context) {

}

// 重启应用服务，可指定应用服务名称和副本数量
func restartDeployment(c *gin.Context) {

}

// 创建应用服务，包含配置信息
func createDeployment(c *gin.Context) {

}

// 设置应用服务的副本数量
func configDeploymentReplicas(c *gin.Context) {

}

// 删除应用服务，需要指定应用服务的名称
func deleteDeployment(c *gin.Context) {

}
