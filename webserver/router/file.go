// "file.go" file is create by Huxd 2020.07.27
// it used to transfer file operation.
// 前端使用webuploader进行文件的传输，所以在文件接收时目前仅针对webuploader的信息格式进行处理

package router

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"mime/multipart"
	"os"
	"os/user"
	"strconv"
	"time"
	"webserver/router/errcode"
	mymd5 "webserver/router/md5"
)

// 文件信息结构体，表明接收或发送的文件基本信息
type FileInformation struct {
	ID               string // 文件的标识
	FileType         string // 文件类型
	FilePath         string // 文件路径
	FileName         string // 文件名称
	FileSize         string // 文件大小
	SumBlock         int    // 文件分割的总块数，当总块数为1时表示不分块
	LastModifiedDate string // 文件最后修改的时间
	MD5              string // 文件的md5值
}

// 接收文件对象，用于在接收文件时对文件数据的管理
type ReceiveFileObject struct {
	FInfo          FileInformation // 文件信息结构体
	BlockArray     []FileBlock     // 文件块数组
	RecBlockStatus []bool          // 接收块的状态
	WriteIndex     int             // 待写入块号序号
}

// 文件块信息结构体
type FileBlock struct {
	BlockIndex int    // 文件块的序号，基于0
	BlockSize  int64  // 文件块的大小
	Content    []byte // 文件块的内容
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
	fmt.Println("<=====================================================")
	// 读取multipart/form内容
	form, err := c.MultipartForm()
	// 如果没有错误，则执行解析并处理
	if err == nil {
		// 获取文件信息，同时获取块序号（如果未分块，则块序号为默认值0）
		info, blockIndex := getFileInformation(form.Value)

		// 声明block对象，以及获取块对象是否成功的标识
		var block FileBlock
		var bSuccess bool
		// 根据WebUploader的相关协议，暂时不需要遍历Form的File对象，直接取值的key值由WebUploader中的fileVal值指定
		files := form.File["file"]
		// 目前在FileHeader切片中，一般只会有一个FileHeader对象，如果存在多个FileHeader对象使用最后一个该对象的数据内容
		for _, file := range files {
			// 获取文件块的信息内容
			block, bSuccess = getFileBlackInformation(file)
			// 如果获取块信息出现故障，则返回服务器内部错误
			if !bSuccess {
				serveErrorJSON(c,
					errcode.ErrorCodeUnknown.WithMessage(fmt.Sprintf("获取文件块信息时出现问题")))
				return
			}
			// 补充设定文件块序号，因为块序号是在之前解析出来的
			block.BlockIndex = blockIndex
		}

		// 根据文件的总块数确定是单块文件还是多块文件
		if info.SumBlock == 0 {
			// 此时为单块文件，直接进行文件保存操作，此时文件不会超过前端限制的文件大小
			bSuccess, err := singleWriteFile(info, block.Content)
			if !bSuccess {
				serveErrorJSON(c,
					errcode.ErrorCodeUnknown.WithMessage(fmt.Sprintf("写入文件出现问题：", err)))
				return
			}
		} else if info.SumBlock > 0 {
			// 此时为多块文件，需要创建多块文件处理对象，进行多块文件的处理
		}
	} else {
		serveErrorJSON(c,
			errcode.ErrorCodeUnknown.WithMessage(fmt.Sprintf("获取multipart/form内容失败，错误信息:%v", err)))
		return
	}

	// 下面的步骤是一次发送就可以上传文件到服务端并存储的过程
	//file, err := c.FormFile("file")
	//if err == nil {
	//			fmt.Println(file.Filename)
	//			fmt.Println(file.Header)
	//			fmt.Println(file.Size)
	//
	//			c.SaveUploadedFile(file, fmt.Sprintf("/home/cetc/%v", file.Filename))
	//}
	fmt.Println("=====================================================>")

	c.JSON(200, gin.H{"status": "ok"})
	return
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

// 单次写入文件的全部内容
func singleWriteFile(fileInfo FileInformation, content []byte) (bool, error) {
	// 声明MD5计算对象，并计算出MD5字符串
	hash := md5.New()
	hash.Write(content)
	strMD5 := hex.EncodeToString(hash.Sum(nil))
	// 如果前端计算的MD5值不为空，则进行文件内容的MD5值校验；如果前端计算的MD5值为空，则不进行校验
	if fileInfo.MD5 != "" && strMD5 != fileInfo.MD5 {
		return false, fmt.Errorf("文件的MD5值校验失败！")
	}

	// 配置文件的路径
	disFilePath := fileInfo.FilePath + fileInfo.FileName
	// 如果目标文件存在，则进行删除操作
	if IsExist(disFilePath) {
		if err := os.Remove(disFilePath); err != nil {
			return false, err
		}
	}

	// 目前是由当前程序创建的文件才能使用
	file, err := os.OpenFile(disFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if err != nil {
		fmt.Println("<singleWriteFile> open file error： ", err)
		return false, err
	}
	// 创建写缓冲区，默认的缓冲区是4096个字节
	writeBuf := bufio.NewWriterSize(file, 1024*1024)
	// 启动计时工具，统计耗时
	st := time.Now()
	defer func() {
		// 将缓冲区中的内容写入到文件中
		writeBuf.Flush()
		// 关闭文件
		file.Close()
		// 如果是调试模式，则输出写文件的时间值
		if DebugMode {
			fmt.Println("singleWriteFile方法 写文件的耗时：", time.Now().Sub(st).Seconds(), "s")
		}
	}()

	// 将数据内容写入文件
	size, _ := strconv.Atoi(fileInfo.FileSize)
	n, err := writeBuf.Write(content)
	if n != size || err != nil {
		fmt.Println("写入文件字节数错误 或 文件写入时发生错误： ", err)
		return false, err
	}

	return true, nil
}

//// 读取后写入文件，并校验MD5值
//func writeFile() bool {
//	// 配置文件的路径
//	disFilePath := "/home/cetc/wt资料/测试写文件.rar"
//
//	// 如果目标文件存在，则进行删除操作
//	if IsExist(disFilePath) {
//		err := os.Remove(disFilePath)
//		return false
//	}
//	// 目前是由当前程序创建的文件才能使用
//	file, err := os.OpenFile(disFilePath, os.O_WRONLY | os.O_CREATE | os.O_EXCL, os.ModePerm)
//	if err != nil {
//		fmt.Println("open dis file error： ", err)
//		return false
//	}
//	// 创建写缓冲区，默认的缓冲区是4096个字节
//	writeBuf := bufio.NewWriterSize(file, 1024*1024)
//
//	// 启动计时工具，统计耗时
//	st := time.Now()
//	defer func() {
//		writeBuf.Flush()
//		file.Close()
//
//		fmt.Println("writeFile写文件的耗时：", time.Now().Sub(st).Seconds(), "s")
//	}()
//
//	//info, _ := f1.Stat()
//	//fmt.Println(info)
//
//	hash := md5.New()
//	var sum int = 0
//	var bytes []byte = make([]byte, 1024*1024)
//	for {
//		n, err := readBuf.Read(bytes)
//		if err == io.EOF || n == 0 {
//			break
//		}
//		sum = sum + n
//		fmt.Println("本次读取字节数：", n, " 总共读取字节数：", sum, " 本次数据的MD5值为：", mymd5.GetMD5HashCode(bytes))
//
//		hash.Write(bytes[:n])
//
//		// 写入缓冲区
//		n2, err := writeBuf.Write(bytes[:n])
//		if n2 != n || err != nil {
//			fmt.Println("文件写入错误： ", err)
//			break
//		}
//	}
//
//	strMD5 := hex.EncodeToString(hash.Sum(nil))
//	fmt.Println("文件的MD5值是：", strMD5)
//
//	return true
//}

// 创建指定路径文件的UUID （暂时使用文件内容的MD5值）
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

// 从multipart/form内容中获取文件信息
func getFileInformation(origin map[string][]string) (FileInformation, int) {
	// 调试模式输出信息
	if DebugMode {
		fmt.Println(origin)
	}

	// 创建FileInformation对象
	info := FileInformation{
		SumBlock: 0,
	}

	// 定义块序号变量
	var chunkIndex int = 0
	// 在WebUploader中确定有ID、文件最后修改时间、name、size、type、chunk、chunks、md5信息
	for key, value := range origin {
		switch key {
		case "id": // id，由WebUploader前端生成
			info.ID = value[0]
		case "lastModifiedDate": // 最后修改的日期时间
			info.LastModifiedDate = value[0]
		case "name": // 文件名称
			info.FileName = value[0]
		case "size": // 文件大小
			info.FileSize = value[0]
		case "type": // 文件类型
			info.FileType = value[0]
		case "chunk": // 块序号，默认为0，当存在时使用实际值
			// 如果是块号，则进行转换并判断是否为大于0的数值，大于0则进行赋值
			if chunk, err := strconv.Atoi(value[0]); err == nil && chunk > 0 {
				chunkIndex = chunk
			}
		case "chunks": // 块总数，默认为0，当WebUploader分块时，此处总是大于0
			// 如果是分块数，则进行转换并判断是否为大于0的数值，大于0则进行赋值
			if chunks, err := strconv.Atoi(value[0]); err == nil && chunks > 0 {
				info.SumBlock = chunks
			}
		case "md5": // 文件的MD5哈希值，并不是每个消息中都有这个内容
			info.MD5 = value[0]
		}
	}

	// 存储文件的路径目前有全局变量设置，所以在这里暂定为以下路径
	u, err := user.Current()
	if nil == err {
		info.FilePath = u.HomeDir + "/test/"
	}
	//info.FilePath = "/home/cetc/test/"

	// 返回文件信息对象和块序号
	return info, chunkIndex
}

// 获取文件块信息
func getFileBlackInformation(file *multipart.FileHeader) (block FileBlock, bSuccess bool) {
	// 输出调试信息
	if DebugMode {
		fmt.Println(file.Filename)
		fmt.Println(file.Header)
		fmt.Println(file.Size)
	}

	// 获取块内容的大小
	block.BlockSize = file.Size

	// 获取块内容的数据
	if f, err := file.Open(); err != nil {
		return block, false
	} else {
		if buf, err := ioutil.ReadAll(f); err != nil {
			return block, false
		} else {
			block.Content = buf
		}
	}

	return block, true
}
