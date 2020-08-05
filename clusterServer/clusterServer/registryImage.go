package clusterServer

import (
	"clusterHeader"
	"clusterServer/registry"
	"errors"
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
		AnswerRequestOfImage(pkgId, "", "FALSE", "镜像仓库不在线!", imageData)
		return
	}

	// 从镜像仓库获取镜像列表
	imageList, _, err := registry.GetRepositoryList()
	if err != nil {
		AnswerRequestOfImage(pkgId, "", "FALSE", "获取镜像列表失败!"+err.Error(), imageData)
	} else {
		AnswerRequestOfImage(pkgId, header.JsonString(imageList), "SUCCESS", "获取镜像列表成功!", imageData)
	}
}

// 获取镜像标签列表
func GetImageTagList(pkgId uint16, imageData *header.ImageData) {
	// 如果镜像仓库不在线，直接返回错误
	if !registry.IsOnline() {
		AnswerRequestOfImage(pkgId, "", "FALSE", "镜像仓库不在线!", imageData)
		return
	}

	// 获取镜像标签列表
	tagList, _, err := registry.GetTagList(imageData.ImageName)
		if err != nil {
			AnswerRequestOfImage(pkgId, "", "FALSE", "获取镜像"+imageData.ImageName+"的标签列表失败!"+err.Error(), imageData)
		} else {
			AnswerRequestOfImage(pkgId, header.JsonString(tagList), "SUCCESS", "获取镜像"+imageData.ImageName+"的标签列表成功!", imageData)
		}

}

// 从镜像仓库删除镜像
func RemoveImageFromRepository(pkgId uint16, imageData *header.ImageData) {
	// 如果镜像仓库不在线，直接返回错误
	if !registry.IsOnline() {
		AnswerRequestOfImage(pkgId, "", "FALSE", "镜像仓库不在线!", imageData)
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

		err := errors.New(strings.Join(errorList, "\n"))

		if errorCount > 0 {
			AnswerRequestOfImage(pkgId, "", "FALSE", err.Error(), imageData)
		} else {
			AnswerRequestOfImage(pkgId, "", "SUCCESS", err.Error(), imageData)
		}

}

func UpdateImageRepository(h string, pkgId uint16, imageData *header.ImageData) {
	// 如果镜像仓库不在线，直接返回错误
	if !registry.IsOnline() {
		AnswerRequestOfImage(pkgId, "", "FALSE", "镜像仓库不在线!", imageData)
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
				registry.DeleteRepsitory(imageData.ImageName, tag)
			}

				//use tcp to agent
		newdata := header.ImageData{}.From(header.FLAG_IMAG_PUSH, imageData.ImageName, tags, imageData.ImageBody, "", "")
		writeAgentData(h, tcpSocket.TCP_TYPE_FILE, pkgId, header.FLAG_IMAG, header.JsonByteArray(newdata))

		}

}