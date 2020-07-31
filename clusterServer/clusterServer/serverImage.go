package clusterServer

import (
	"clusterHeader"
	"encoding/json"
	"tcpSocket"
)

func ReceiveDataFromAgent(handle string, pkgId uint16, data []byte) {

	// 解析数据
	var imageData header.ImageData
	json.Unmarshal(data, &imageData)
	// 返回结果
	AnswerRequest(pkgId, imageData)

	// 输出日志
	//dealType := imagedata.DealType
	//imageName := imageData.ImageName
	//tags := imageData.Tags
	//body := []byte(imageData.ImageBody)
	//result := imageData.Result
	//tiperr := imageData.TipError
	//grade := tcpSocket.TCP_TPYE_CONTROLLER
	//
	//if dealType == header.FLAG_IMAG_REMO {
	//	grade = tcpSocket.TCP_TYPE_FILE
	//}


	//log.Println("server端接收agent镜像操作的返回信息", imageName, tags, len(body), result , tiperr)
	// SERVER端将处理结果写入日志系统

}

func ProcessImageFlagDataFromClient(pkgId uint16, imageData *header.ImageData) {

	// 镜像操作需要最优的节点进行处理，如果未找到有效最优节点，则直接返回失败
	h := GetBestAgentForImageOper()
	if h == "" {
		AnswerRequestOfImage(pkgId, "当前无在线计算节点，无法执行此操作", "FALSE", nil, imageData)
		return
	}

	// 按照子类型/操作进行相关操作
	dealType := imageData.DealType

	switch dealType {
	case header.FLAG_IMAG_LIST:
		GetImageList(pkgId, imageData)
	case header.FLAG_IMAG_TGLS:
		GetImageTagList(pkgId, imageData)
	case header.FLAG_IMAG_REMO:
		RemoveImageFromRepository(pkgId, imageData)
	case header.FLAG_IMAG_UPDT: // 私有仓库操作
		UpdateImageRepository(h, pkgId, imageData)
	case header.FLAG_IMAG_DIST,
		header.FLAG_IMAG_PUSH,
		header.FLAG_IMAG_SAVE: // 镜像分发、推送、保存
		writeAgentData(h, tcpSocket.TCP_TPYE_CONTROLLER, pkgId, header.FLAG_IMAG, header.JsonByteArray(imageData))
	case header.FLAG_IMAG_BUID,
		header.FLAG_IMAG_LOAD: // 构建镜像、下载镜像
		writeAgentData(h, tcpSocket.TCP_TYPE_FILE, pkgId, header.FLAG_IMAG, header.JsonByteArray(imageData))
	default: // 子标识/操作错误
		AnswerRequestOfImage(pkgId, "子标识/操作错误!", "FALSE", nil, imageData)
	}
}

// 为镜像操作寻找最优的Agnet执行
func GetBestAgentForImageOper() string {
	if nodes.Count() == 0 {	// 如果当前集群内的计算节点数量为0，则直接返回空
		return ""
	} else {	// 遍历所有的计算节点，找到最优的节点返回。目前是返回第一个在线的节点，后续更新
		for _,h := range nodes.GetNodeIds() {
			if nodes.GetState(h) == true {
				return h
			}
		}

		// 如果所有计算节点都不在线，则返回空
		return ""
	}
}
