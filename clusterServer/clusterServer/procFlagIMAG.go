package clusterServer

import header "clusterHeader"

// web 响应网络数据,非阻塞
// token 	身份认证标识、含用户、主机等信息
// data		接收到的数据包，类型根据flag指定
// respChan	回复当前消息的数据通道，填充数据后不要关闭，且只能填充一次，不可读取
// 所有的镜像类操作都需要向某个最优的节点获取，必须等待数据
func procFlagIMAG(token interface{}, data interface{}, respChan chan<- interface{}) (err error) {
	// 数据结构为 header.ImageData
	imageData := data.(header.ImageData)

	// 生成一个请求
	h := NewRequest(respChan)

	// 处理镜像类操作
	ProcessImageFlagDataFromClient(h, &imageData)

	return nil
}

// 对请求进行回复
func AnswerRequestOfImage(pkgId uint16, imageBody string, result string, err string, imageData *header.ImageData) {
	imageData.ImageBody = imageBody
	imageData.Result = result
	imageData.TipError = err
	//if err == nil {
	//	imageData.TipError = ""
	//} else {
	//	imageData.TipError = err.Error()
	//}
	AnswerRequest(pkgId, *imageData)
}