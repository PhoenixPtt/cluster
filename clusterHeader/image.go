package header

const (
	FLAG_IMAG       = "IMAG"	// 镜像操作
	FLAG_IMAG_LIST  = "IMLIST" //0 列表镜像
	FLAG_IMAG_TGLS  = "IMTGLS" //1 镜像标签列表
	FLAG_IMAG_REMO  = "IMREMO" //2 删除镜像
	FLAG_IMAG_UPDT  = "IMUPDT" //3 更新镜像
	FLAG_IMAG_BUID  = "IMBUID" //4 构建镜像
	FLAG_IMAG_LOAD  = "IMLOAD" //5 导入镜像
	FLAG_IMAG_PUSH  = "IMPUSH" //6 推送镜像
	FLAG_IMAG_SAVE  = "IMSAVE" //7 保存镜像
	FLAG_IMAG_DIST  = "IMDIST" //8 分发镜像
)

type ImageData struct {
	DealType  		string
	ImageName 		string
	Tags      		[]string
	ImageBody 		string
	Result    		string
	TipError  		string
}

func (i ImageData) From(dealType string, imageName string, tags []string, imagebody []byte, result string, err error) *ImageData {
	ImageData := &ImageData{}
	ImageData.DealType = dealType
	ImageData.ImageName = imageName
	ImageData.Tags = tags
	ImageData.ImageBody = string(imagebody)
	ImageData.Result = result
	ImageData.TipError = err.Error()
	return ImageData
}

//type ImageInfoData struct {
//	ImageName  string
//	Tag        string
//	UploadTime string
//	ImageSize  string
//	UploadUser string
//}
