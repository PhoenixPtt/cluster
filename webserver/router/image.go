// "image.go" file is create by Huxd 2020.07.27
// it used to init and due image operation

package router

import (
	header "clusterHeader"
	"github.com/gin-gonic/gin"
)

// 镜像操作相关内容的具体处理函数 /image
func initImageRouter(group *gin.RouterGroup) bool {
	// Get 相关命令
	group.GET("/list", getImageList)
	group.GET("/tags", getImageTagList)

	// Post 相关命令
	group.POST("/create", createImage)
	group.POST("/load", loadImage)
	group.POST("/distribute", distributeImage)

	// Delete 相关命令
	group.DELETE("/delete", deleteImage)

	return true
}

// 获取私有镜像仓库中的镜像列表
func getImageList(c *gin.Context)  {
	// 生成请求信息结构体
	req := requestInf{
		typeFlag: header.FLAG_IMAG,
		opertype: header.FLAG_IMAG_LIST,
	}
	// 获取单次Get信息
	onceToGet(c, req)
}

// 获取指定镜像名称的标签列表
func getImageTagList(c *gin.Context)  {
	// 获取指定的镜像名称
	name := c.DefaultQuery("name", "")
	if name == "" {

	}

	// 生成请求信息结构体
	req := requestInf{
		typeFlag: header.FLAG_IMAG,
		opertype: header.FLAG_IMAG_TGLS,
		pars: make([]header.OperPar, 1),
	}
	req.pars[0] = header.OperPar{
		Name: "imagename",
		Value: name,
	}
	// 获取单次Get信息
	onceToGet(c, req)
}

// 通过配置以及文件创建镜像
func createImage(c *gin.Context)  {

}

// 通过文件加载生成镜像
func loadImage(c *gin.Context)  {

}

// 从镜像仓库中分发镜像到agent
func distributeImage(c *gin.Context)  {

}

// 从镜像仓库中删除指定的镜像
func deleteImage(c *gin.Context)  {

}