package clusterServer

import (
	"log"

	"clusterHeader"
	"clusterServer/registry"
	"tcpSocket"
)

func registryReceiveDataFromClient(handle string, pkgId uint16, sendbyte []byte) {

	log.Println("registry处理client数据", handle, len(sendbyte))

	var imagedata header.ImageData
	err := header.Decode(sendbyte, &imagedata)
	if err != nil {
		log.Println("decode data false")
		return
	}

	dealType := imagedata.DealType
	imagename := imagedata.ImageName
	tags := imagedata.Tags
	if len(tags) <= 0 {
		tags = append(tags, "latest")
	}
	imagebody := []byte(imagedata.ImageBody)
	var grade = tcpSocket.TCP_TPYE_CONTROLLER
	var sendData []byte

	// 初始化私有镜像仓库连接
	if registry.InitialConnect() {
		// 获取私有镜像仓库是否在线
		bOnline := registry.IsOnline()
		// 如果镜像仓库在线，才可执行具体操作
		if bOnline {

			switch dealType {
			case header.FLAG_IMAG_LIST:
				{
					imagelist, _, err := registry.GetRepositoryList()
					if err != nil {
						sendData = []byte("获取镜像列表失败!")
						updateImageData(pkgId, sendData, "FALSE", err, imagedata)
						return
					}
					imageListByte, err := header.Encode(imagelist)
					if err != nil {
						//结构体解析错误
						sendData = []byte("获取镜像列表成功，返回结果过程中结构体解析错误!")
						updateImageData(pkgId, sendData, "FALSE", err, imagedata)
						return
					}
					sendData = imageListByte
				}
			case header.FLAG_IMAG_TGLS:
				{
					taglist, _, err := registry.GetTagList(imagename)
					if err != nil {
						sendData = []byte("获取镜像标签列表失败!")
						updateImageData(pkgId, sendData, "FALSE", err, imagedata)
						return
					}
					taglistByte, err := header.Encode(taglist)
					if err != nil {
						//结构体解析错误
						sendData = []byte("获取镜像标签列表成功，返回结果过程中结构体解析错误!")
						updateImageData(pkgId, sendData, "FALSE", err, imagedata)
						return
					}
					sendData = taglistByte
				}
			case header.FLAG_IMAG_REMO:
				{
					for _, tag := range tags {
						isDeleteSuccess := registry.DeleteRepsitory(imagename, tag)
						if !isDeleteSuccess {
							//结构体解析错误
							sendData = []byte("删除私有仓库镜像失败!")
							updateImageData(pkgId, sendData, "FALSE", nil, imagedata)
							return
						}
					}
					sendData = []byte("删除私有仓库镜像成功!")
				}
			case header.FLAG_IMAG_UPDT:
				{
					//更新操作
					for _, tag := range tags {
						isExist := registry.IsExistRepositoryTag(imagename, tag)
						//judge if the image is exist
						if isExist {
							//delete image in registry
							isDeleteSuccess := registry.DeleteRepsitory("library/"+imagename, tag)
							if !isDeleteSuccess {
								sendData = []byte("更新操作过程中，删除私有仓库镜像失败!")
								updateImageData(pkgId, sendData, "FALSE", nil, imagedata)
								return
							}
						}
						//use tcp to agent
						dealType = header.FLAG_IMAG_PUSH
						newdata := header.ImageData{}.From(dealType, imagename, tags, imagebody, "", nil)
						sendbyte, err := header.Encode(newdata)
						if err != nil {
							//结构体解析错误
							sendData = []byte("更新镜像过程中，删除操作成功，push前结构体解析错误!")
							updateImageData(pkgId, sendData, "FALSE", err, imagedata)
							return
						}
						grade = tcpSocket.TCP_TYPE_FILE
						writeAgentData("", grade, pkgId, header.FLAG_IMAG, sendbyte)
					}
				}
			default:
				{
					sendData = []byte("没有匹配的操作类型")
					updateImageData(pkgId, sendData, "FALSE", nil, imagedata)
				}

			}
			updateImageData(pkgId, sendData, "SUCCESS", nil, imagedata)

		} else {
			sendData = []byte("镜像仓库不在线!")
			updateImageData(pkgId, sendData, "FALSE", nil, imagedata)
			return
		}
	} else {
		sendData = []byte("私有镜像仓库连接失败!")
		updateImageData(pkgId, sendData, "FALSE", nil, imagedata)
		return
	}

}

func updateImageData(pkgId uint16, imagebody []byte, result string, err error, imagedata header.ImageData) {

	imagedata.ImageBody = string(imagebody)
	imagedata.Result = result
	imagedata.TipError = err.Error()
	returnResultToClient(pkgId, imagedata)
}
