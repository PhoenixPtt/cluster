// "image.go" file is create by Huxd 2020.07.27
// it used to init and due image operation

package router

import (
	header "clusterHeader"
	"github.com/gin-gonic/gin"
	"webserver/router/errcode"
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

	// 处理非Get时，可能进行的OPTION请求
	group.OPTIONS("/create", optionImage)
	group.OPTIONS("/load", optionImage)
	group.OPTIONS("/distribute", optionImage)
	group.OPTIONS("/delete", optionImage)

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
		serveErrorJSON(c, errcode.ErrorCodeUnsupported.WithMessage("获取指定镜像的标签列表时镜像名称不可为空"))
		return
	}

	// 生成请求信息结构体
	req := requestInf{
		typeFlag: header.FLAG_IMAG,
		opertype: header.FLAG_IMAG_TGLS,
		pars: make([]header.OperPar, 1),
	}
	// 添加一个参数，用于指定镜像名称
	req.pars[0] = header.OperPar{
		Name: "imagename",
		Value: name,
	}
	// 获取单次Get信息
	onceToGet(c, req)
}

// 通过配置以及文件创建镜像
func createImage(c *gin.Context)  {
	// 生成请求信息结构体
	req := requestInf{
		typeFlag: header.FLAG_IMAG,
		opertype: header.FLAG_IMAG_BUID,
	}

	// 读取Post中Body的内容，目前确定时JSON格式的
	readImageData(c, &req)

	// 单次Post内容
	onceToPost(c, req)
}

// 通过文件加载生成镜像
func loadImage(c *gin.Context)  {
	// 生成请求信息结构体
	req := requestInf{
		typeFlag: header.FLAG_IMAG,
		opertype: header.FLAG_IMAG_LOAD,
	}

	// 读取Post中Body的内容，目前确定时JSON格式的
	readImageData(c, &req)

	// 执行单次Post操作
	onceToPost(c, req)
}

// 从镜像仓库中分发镜像到agent
func distributeImage(c *gin.Context)  {
	// 生成请求信息结构体
	req := requestInf{
		typeFlag: header.FLAG_IMAG,
		opertype: header.FLAG_IMAG_DIST,
	}

	// 读取Post中Body的内容，目前确定时JSON格式的
	readImageData(c, &req)

	// 执行单次Post操作
	onceToPost(c, req)
}

// 从镜像仓库中删除指定的镜像
func deleteImage(c *gin.Context)  {
	// 生成请求信息结构体
	req := requestInf{
		typeFlag: header.FLAG_IMAG,
		opertype: header.FLAG_IMAG_REMO,
	}

	// 读取Post中Body的内容，目前确定是JSON格式的
	readImageData(c, &req)

	// 执行单次Post操作
	onceToGet(c, req)
}

// 处理Image中的option请求
func optionImage(c *gin.Context) {
	addAccessControlAllowOrigin(c)

	c.Data(200, "", []byte(""))
}

// 读取并绑定指定的json结构体
func readImageData(c *gin.Context, req *requestInf) {
	// 获取body中的内容，在本方法中是header.ImageData结构体类型的JSON数据
	var jData header.ImageData
	if err := c.ShouldBindJSON(&jData); err != nil {
		req.body = ""
	} else {
		req.body = jData
	}
}