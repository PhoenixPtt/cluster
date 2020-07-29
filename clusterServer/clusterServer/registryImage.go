package clusterServer

import (
	"clusterHeader"
	"clusterServer/registry"
	"strings"
	"tcpSocket"
)

func init() {
	registry.InitialConnect()
}

// 获取镜像列表
func GetImageList(pkgId uint16, imageData *header.ImageData) {
	// 如果镜像仓库不在线，直接返回错误
	if !registry.IsOnline() {
		AnswerRequestOfImage(pkgId, "镜像仓库不在线!", "FALSE", nil, imageData)
		return
	}

	// 从镜像仓库获取镜像列表
	imageList, _, err := registry.GetRepositoryList()
	if err != nil {
		AnswerRequestOfImage(pkgId, "获取镜像列表失败!", "FALSE", err, imageData)
	} else {
		AnswerRequestOfImage(pkgId, header.JsonString(imageList), "TRUE", nil, imageData)
	}
}

// 获取镜像标签列表
func GetImageTagList(pkgId uint16, imageData *header.ImageData) {
	// 如果镜像仓库不在线，直接返回错误
	if !registry.IsOnline() {
		AnswerRequestOfImage(pkgId, "镜像仓库不在线!", "FALSE", nil, imageData)
		return
	}

	// 获取镜像标签列表
	tagList, _, err := registry.GetTagList(imageData.ImageName)
		if err != nil {
			AnswerRequestOfImage(pkgId, "获取镜像标签列表失败!", "FALSE", err, imageData)
		} else {
			AnswerRequestOfImage(pkgId, header.JsonString(tagList), "TRUE", nil, imageData)
		}

}

// 从镜像仓库删除镜像
func RemoveImageFromRepository(pkgId uint16, imageData *header.ImageData) {
	// 如果镜像仓库不在线，直接返回错误
	if !registry.IsOnline() {
		AnswerRequestOfImage(pkgId, "镜像仓库不在线!", "FALSE", nil, imageData)
		return
	}

	// 处理不带标签的情况
	tags := imageData.Tags
	if len(tags) <= 0 {
		tags = append(tags, "latest")
	}

	var errorList []string // 删除失败和成功信息
	var errorCount int = 0 // 删除失败的总数

	// 依次删除所有镜像
		for _, tag := range tags {
			// 删除一个镜像
			isDeleteSuccess := registry.DeleteRepsitory(imageData.ImageName, tag)
			if !isDeleteSuccess {
				errorList = append(errorList, "从仓库删除镜像 " + imageData.ImageName + ":" + tag + " 失败")
				errorCount++
			} else {
				errorList = append(errorList, "从仓库删除镜像 " + imageData.ImageName + ":" + tag + " 成功")
			}
		}

		if errorCount > 0 {
			AnswerRequestOfImage(pkgId, strings.Join(errorList, "\n"), "FALSE", nil, imageData)
		} else {
			AnswerRequestOfImage(pkgId, strings.Join(errorList, "\n"), "TRUE", nil, imageData)
		}

}

func UpdateImageRepository(h string, pkgId uint16, imageData *header.ImageData) {
	// 如果镜像仓库不在线，直接返回错误
	if !registry.IsOnline() {
		AnswerRequestOfImage(pkgId, "镜像仓库不在线!", "FALSE", nil, imageData)
		return
	}

	// 处理不带标签的情况
	tags := imageData.Tags
	if len(tags) <= 0 {
		tags = append(tags, "latest")
	}

		// 更新操作
		for _, tag := range tags {
			isExist := registry.IsExistRepositoryTag(imageData.ImageName, tag)
			//judge if the image is exist
			if isExist {
				//delete image in registry
				registry.DeleteRepsitory("library/"+imageData.ImageName, tag)
			}

				//use tcp to agent
		newdata := header.ImageData{}.From(header.FLAG_IMAG_PUSH, imageData.ImageName, tags, imageData.ImageBody, "", nil)
		writeAgentData(h, tcpSocket.TCP_TYPE_FILE, pkgId, header.FLAG_IMAG, header.JsonByteArray(newdata))

		}

}