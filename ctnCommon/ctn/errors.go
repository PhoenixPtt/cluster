package ctn

const (
	ERR_TYPE_IMAGE_GETLIST  = "镜像：获取镜像列表失败"
	ERR_TYPE_IMAGE_PULL     = "镜像：拉去镜像失败"
	ERR_TYPE_CTN_EXIST      = "容器：容器已存在"
	ERR_TYPE_CTN_NOTEXIST   = "容器：容器不存在"
	ERR_TYPE_CTN_RUNNING    = "容器：容器正在运行"
	ERR_TYPE_CTN_NOTRUNNING = "容器：容器未运行"
	ERR_TYPE_CTN_CREATE     = "容器：创建容器失败"
	ERR_TYPE_CTN_INFO       = "容器：获取容器信息失败"
	ERR_TYPE_CTN_START      = "容器：启动容器失败"
	ERR_TYPE_CTN_STOP       = "容器：停止容器失败"
	ERR_TYPE_CTN_KILL       = "容器：强杀容器失败"
	ERR_TYPE_CTN_REMOVE     = "容器：删除容器失败"
	ERR_TYPE_CTN_GETLOG     = "容器：获取容器日志失败"
)
