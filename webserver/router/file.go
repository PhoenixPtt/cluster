// "file.go" file is create by Huxd 2020.07.27
// it used to transfer file operation.
// 前端使用webuploader进行文件的传输，所以在文件接收时目前仅针对webuploader的信息格式进行处理

package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"time"
	"webserver/router/errcode"
	mymd5 "webserver/router/md5"
)

// 文件信息结构体，表明接收或发送的文件基本信息
type FileInformation struct {
	UUID     string // 文件的唯一标识
	FileType string // 文件类型
	FilePath string // 文件路径
	FileName string // 文件名称
	FileSize string // 文件大小
	SumBlock uint   // 文件分割的总块数
	MD5      string // 文件的md5值
}

// 接收文件对象，用于在接收文件时对文件数据的管理
type ReceiveFileObject struct {
	FileInformation
	RecBlockStatus []bool // 接收块的状态
}

// 文件块信息结构体
type FileBlock struct {
	BlockIndex int    // 文件块的序号，基于0
	BlockSize  string // 文件块的大小
	MD5        string // 文件块的MD5值
}

// 文件操作相关内容的具体处理函数 /file
func initFileRouter(group *gin.RouterGroup) bool {
	// Get 相关命令
	//group.GET("/list", getImageList)

	// Post 相关命令
	group.POST("/upload", receiveFile)

	// Delete 相关命令
	//group.DELETE("/delete", deleteImage)

	// 处理非Get时，可能进行的OPTION请求
	group.OPTIONS("/upload", onceToOption)

	return true
}

// 接收文件的方法
func receiveFile(c *gin.Context) {
	// 读取multipart/form内容
	form, err := c.MultipartForm()
	// 如果没有错误，则执行解析并处理
	if err == nil {
		// 解析包含的关系信息，这里需要注意单文件和多文件的问题
		fmt.Println(form.Value)
		// 根据webuploader的相关协议，暂时不需要遍历File对象
		files := form.File["file"]
		// 目前在FileHeader切片中，一般只会有一个FileHeader对象
		for _, file := range files {
			fmt.Println(file.Filename)
			fmt.Println(file.Header)
			fmt.Println(file.Size)

			//f,_ := file.Open()
			//buf,_:= ioutil.ReadAll(f)
		}
	} else {
		serveErrorJSON(c,
			errcode.ErrorCodeUnknown.WithMessage(fmt.Sprintf("获取multipart/form内容失败，错误信息:%v", err)))
		return
	}
	// 读取文件的信息内容
	//info := getFileInformation(c)

	// 读取文件块的信息内容
	//block := getFileBlackInformation(c)

	// 下面的步骤是一次发送就可以上传文件到服务端并存储的过程
	//file, err := c.FormFile("file")
	//if err == nil {
	//			fmt.Println(file.Filename)
	//			fmt.Println(file.Header)
	//			fmt.Println(file.Size)
	//
	//			c.SaveUploadedFile(file, fmt.Sprintf("/home/cetc/%v", file.Filename))
	//}
	fmt.Println("=====================================================")

	c.JSON(200, gin.H{"status": "ok"})
	return

	//// 根据文件和文件块的信息内容，处理接收到的数据内容
	//if info.SumBlock == 1 {
	//	// 此时文件就在一个块中，则直接获取内容
	//	//c.FormFile()
	//
	//} else if info.SumBlock > 1 {
	//	// 此时文件在多个块中，需要进行多块写入
	//
	//} else {
	//	// 此时文件块个数错误，应直接返回错误信息
	//	serveErrorJSON(c,
	//		errcode.ErrorCodeUnsupported.WithMessage(fmt.Sprintf("传输文件时，文件块的个数应大于0，当前值为:%v", info.SumBlock)))
	//}
}

// 判断src路径是否存在，可以判断的对象包括：文件或文件夹
func IsExist(src string) bool {
	// 根据路径指向对象的状态，确定是否存在
	if _, err := os.Stat(src); os.IsNotExist(err) {
		//fmt.Println("file does not exist")
		return false
	}

	return true
}

// 生成文件的UUID
func createFileUUID(filePath string) string {
	// 打开文件，如果打开文件失败，则返回空的UUID，表示失败
	f, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		fmt.Println("open file error： ", err)
		return ""
	}

	// 记录当前时间，用于计算文件操作的耗时
	st := time.Now()
	defer func() {
		f.Close()
		fmt.Println("读取文件内容并生成UUID值的耗时：", time.Now().Sub(st).Seconds(), "s")
	}()

	// 返回获取文件的MD5校验码
	return mymd5.GetFileMD5HashCode(f)
}

// 获取文件信息
func getFileInformation(c *gin.Context) (info FileInformation) {
	return info
}

// 获取文件块信息
func getFileBlackInformation(c *gin.Context) (info FileBlock) {
	return info
}
