// registry.go
package registry

import (
	"fmt"
	// "io/ioutil"
	"encoding/json"
	"log"
	"registry/request"
)

// 定义镜像列表结构体
type Imagelist struct {
	Repositories []string
}

// 定义指定镜像Tag列表结构体
type Imagetaglist struct {
	Name string
	Tags []string
}

// 定义镜像对象结构体
type repositoryObject struct {
	dockerContentDigest string
	schemaVersion       int
	mediaType           string
	config              baseDigest
	layers              []baseDigest
}

// 定义基础digest数据结构体
type baseDigest struct {
	mediaType string
	size      int
	digest    string
}

// 私有镜像仓库的地址
var registry_endpoint string
var registry_API_Address string

var DEBUG bool

// 初始化函数，在调用registry包时将优先执行
func init() {
	// 测试使用，查看是否调用registry包
	fmt.Printf("<=== initial registry package ===> \r\n")

	// 暂时使用完整的路径，后续可根据域名以及端口号，生成实际的地址字符串
	registry_endpoint = "myregistry.com:5000"
	registry_API_Address = fmt.Sprintf("https://%v/v2/", registry_endpoint)
}

// 初始化私有镜像仓库的连接
func InitialConnect( /*crtfile string, usr string, pw string, registry_endpoint string*/) bool {
	var bOK bool = true // IsExistRepository

	// 暂时不处理变量修改，因为未知的设定操作

	// 主动创建http.Client对象
	if !request.CreateHttpClient() {
		bOK = false
	}

	// 根据bOK的状态，进行log信息的显示
	var show string
	if bOK {
		show = "Initial registry connect object success"
	} else {
		show = "Initial registry connect object fail"
	}

	if DEBUG {
		log.Println(show)
	}

	return bOK
}

// 验证私有镜像仓库是否启动  IsExistRepository
func IsOnline() bool {
	req, err := request.Get(registry_API_Address, request.Param{})
	if req == nil || err != nil {
		if DEBUG {
			log.Println(fmt.Sprintf("%v is not online", registry_endpoint))
		}
		return false
	} else {
		if DEBUG {
			log.Println(fmt.Sprintf("%v is online", registry_endpoint))
		}
		return true
	}
}

// 获取镜像仓库中的镜像列表
func GetRepositoryList() (imagelist Imagelist, iCount int, err error) {
	// 初始化部分返回值变量，默认镜像仓库中的镜像数量为-1，表示获取失败
	iCount = -1

	// 生成url命令
	url := fmt.Sprintf("%v_catalog", registry_API_Address)

	// 执行获取镜像列表的命令
	req, err := request.Get(url, request.Param{})
	if req == nil || err != nil {
		if DEBUG {
			log.Println("GetRepositoryList() exec http GET fail, ", err)
		}
		return
	}

	// 转换获取的json数据为map[string]interface{}类型
	// 不能直接使用ImageList结构体，是因为返回的repositories字段不是可导出的，如需设置为可导出，需要讲首字母大写
	bUnmarshalOK, data := getJSONData(req.Body)
	if !bUnmarshalOK {
		err = fmt.Errorf("GetRepositoryList() exec getJSONData fail")
		return
	}

	// 获取repositories对应的属性
	value, ok := data["repositories"]
	if ok {
		// 根据value类型进行处理，如果是接口切片类型，则进行处理，否则输出不是期望值的文本内容
		switch activeValue := value.(type) {
		case []interface{}:
			// 获取切片长度，如果大于0，则进行赋值操作
			if iCount = len(activeValue); iCount > 0 {
				// 创建字符串切片，并指定切片大小
				imagelist.Repositories = make([]string, iCount)
				// 遍历接口切片，并将镜像名称赋值给Imagelist对象
				for index, imagename := range activeValue {
					// 从万用接口类型转换为字符串，需要使用如下操作，而不是常规的转换方法
					imagelist.Repositories[index] = imagename.(string)
				}
			}
		default:
			err = fmt.Errorf("repositories value type is not expect type")
			if DEBUG {
				fmt.Println(err)
			}
		}
	}

	return imagelist, iCount, err

	// 下面是一种安全的解码未知结构的json数据
	// var r interface{}
	// var bodybyte []byte = []byte(req.Body)
	// err = json.Unmarshal(bodybyte, &r)
	// if err != nil {
	// 	fmt.Println("request.Body is unmarshal fail, error code: ", err)
	// 	return
	// } else {
	// 	// req.Dump()
	// 	// fmt.Println("---------------------------------------")
	// 	// fmt.Printf(req.Body)

	// 	// 解析返回的json数据
	// 	temImagelist, ok := r.(map[string]interface{})
	// 	if !ok {
	// 		fmt.Println("request.Body is not expect")
	// 		err = nil
	// 		return
	// 	}

	// 	for key, value := range temImagelist {
	// 		switch valueType := value.(type) {
	// 		case string:
	// 			fmt.Println(key, "is string", valueType, value)
	// 		case []interface{}:
	// 			fmt.Println(key, "is an array: ")
	// 			for i, iv := range valueType {
	// 				fmt.Println(i, iv)
	// 			}

	// 		}
	// 	}
	// }
}

// 获取指定镜像名对象的Tag列表
func GetTagList(imagename string) (taglist Imagetaglist, iCount int, err error) {
	// 初始化部分返回值变量，默认获取的tag数量为-1，表示获取失败
	iCount = -1

	// 如果镜像名称为空，则直接返回
	if len(imagename) <= 0 {
		return
	}

	// 生成url命令
	url := fmt.Sprintf("%v%v/tags/list", registry_API_Address, imagename)

	// 执行获取指定镜像的标签列表的命令
	req, err := request.Get(url, request.Param{})
	if req == nil || err != nil {
		if DEBUG {
			log.Println("GetTagList() exec http GET fail, ", err)
		}
		return
	}

	// 如果返回的状态码不是200，则需要提示并退出
	if req.StatusCode != 200 {
		err = fmt.Errorf("Get Tag List is fail, error information: %v", req.Body)
		if DEBUG {
			log.Println(url, err)
		}
		return
	}

	// 转换获取的json数据为map[string]interface{}类型
	bUnmarshalOK, data := getJSONData(req.Body)
	if !bUnmarshalOK {
		err = fmt.Errorf("GetTagList() exec getJSONData fail")
		return
	}

	// 获取tag列表
	value, ok := data["tags"]
	if ok {
		// 根据value类型进行处理，如果是接口切片类型，则进行处理，否则输出不是期望值的文本内容
		switch activeValue := value.(type) {
		case []interface{}:
			// 获取切片长度，如果大于0，则进行赋值操作
			if iCount = len(activeValue); iCount > 0 {
				// 创建字符串切片，并指定切片大小
				taglist.Tags = make([]string, iCount)
				// 遍历接口切片，并将镜像名称赋值给Imagetaglist对象
				for index, tag := range activeValue {
					// 从万用接口类型转换为字符串，需要使用如下操作，而不是常规的转换方法
					taglist.Tags[index] = tag.(string)
				}
			}
		default:
			err = fmt.Errorf("tags value type is not expect type")
			if DEBUG {
				fmt.Println(err)
			}
		}
	}

	// 获取body中的name字段内容，如果正常获取，且同形参imagename相同，则赋值给taglist对象
	// 此处需要注意，并不会因为不等而返回错误信息！！！！！！！！！
	value, ok = data["name"]
	if ok {
		strName := value.(string)
		if strName == imagename {
			taglist.Name = strName
		}
	}

	return taglist, iCount, err
}

// 查找是否存在指定的镜像名称和标签的镜像
func IsExistRepositoryTag(imagename string, tag string) (bExist bool) {
	// 初始化返回值
	bExist = false

	// 验证镜像名称是否为空，如果为空直接返回
	if len(imagename) <= 0 {
		return false
	}

	// 如果标签名称为空，则查找是否存在该镜像名称的镜像
	if len(tag) <= 0 {
		imagelist, _, _ := GetRepositoryList()
		for _, image := range imagelist.Repositories {
			if image == imagename {
				return true
			}
		}
		return false
	}

	// 此时镜像名称和标签名称都不为空，则直接使用是否存在指定镜像名称和标签的表单函数
	bExist, _ = existManifest(imagename, tag)

	if bExist && DEBUG {
		log.Println("Repository ", imagename, tag, "is exist!")
	}

	return bExist
}

// 删除指定镜像名称和标签的镜像
func DeleteRepsitory(imagename string, tag string) bool {

	// 验证镜像名称和标签是否为空，如果为空直接返回
	if len(imagename) <= 0 || len(tag) <= 0 {
		if DEBUG {
			log.Println("Delete Repository fail, imagename or tag is empty")
		}
		return false
	}

	// 获取指定镜像名称和标签的表单
	bGet, reposObject := getManifest(imagename, tag)
	if !bGet {
		if DEBUG {
			log.Println("get manifest fail, imagename:", imagename, tag)
		}
		return false
	}

	// 执行删除对应镜像表单的操作
	url := fmt.Sprintf("%v%v/manifests/%v", registry_API_Address, imagename, reposObject.dockerContentDigest)

	req, err := request.Delete(url, request.Param{})
	if req == nil || err != nil {
		if DEBUG {
			log.Println("DeleteRepsitory exec http DELETE is fail, ", err)
		}
		return false
	}

	if req.StatusCode == 202 {
		if DEBUG {
			log.Println(fmt.Sprintf("DeleteRepsitory %v:%v success!", imagename, tag))
		}
		return true
	}

	return false
}

// 判断是否存在指定镜像名称和标签的表单
func existManifest(repository string, reference string) (bExist bool, strDigest string) {
	// """ check to see it manifest exist """

	// 生成url
	url := fmt.Sprintf("%v%v/manifests/%v", registry_API_Address, repository, reference)

	// 执行Head命令
	req, err := request.Head(url, request.Param{})
	if req == nil || err != nil {
		if DEBUG {
			log.Println("existManifest() exec http HEAD is fail, ", err)
		}
		return false, ""
	}

	// 判断http返回的状态值，如果不是200则说明失败
	if req.StatusCode == 200 {
		respHeader := req.Response.Header["Docker-Content-Digest"]
		return true, respHeader[0]
	} else {
		if DEBUG {
			log.Println("existManifest(): response Status is ", req.Status)
		}
		return false, ""
	}
}

// 获取指定镜像名称和标签的表单信息
func getManifest(repository string, reference string) (bSuccess bool, object repositoryObject) {
	// """ get manifest for tag or digest """

	// 生成url
	url := fmt.Sprintf("%v%v/manifests/%v", registry_API_Address, repository, reference)

	// 执行获取指定镜像名称和标签的表单的命令
	req, err := request.Get(url, request.Param{})
	if req == nil || err != nil {
		if DEBUG {
			log.Println("getManifest() exec http GET is fail, ", err)
		}
		return false, object
	}

	// 判断http返回的状态值，如果不是200则说明失败
	if req.StatusCode == 200 {
		// 解析获取的req.Body内容
		object = analysisManifestData(req.Body)

		// 获取Docker-Content-Digest信息，并赋值到object中
		respHeader := req.Response.Header["Docker-Content-Digest"]
		object.dockerContentDigest = respHeader[0]

		return true, object
	} else {
		if DEBUG {
			log.Println("getManifest(): response Status is ", req.Status)
		}
		return false, object
	}

	return false, object
}

// 开始垃圾回收工作

// 判断是否存在指定镜像名称和digest的blob
func ExistBlobs(repository string, digest string) bool {
	// """ check to see it blob exist """

	// 生成url
	url := fmt.Sprintf("%v%v/blobs/%v", registry_API_Address, repository, digest)

	// 执行HEAD命令
	req, err := request.Head(url, request.Param{})
	if req == nil || err != nil {
		if DEBUG {
			log.Println("existBlobs() exec http HEAD is fail, ", err)
		}
		return false
	}

	// 判断http返回的状态值，如果不是200则说明失败
	if req.StatusCode == 200 {
		return true
	} else {
		if DEBUG {
			log.Println("existBlobs(): response Status is ", req.Status)
		}
		return false
	}
}

// 删除指定的镜像名称和digest的blob
func DeleteBlobs(repository string, digest string) bool {
	// 生成url
	url := fmt.Sprintf("%v%v/blobs/%v", registry_API_Address, repository, digest)

	// 执行HEAD命令
	req, err := request.Delete(url, request.Param{})
	if req == nil || err != nil {
		if DEBUG {
			log.Println("deleteBlobs() exec http DELETE is fail, ", err)
		}
		return false
	}

	// 判断http返回的状态值，如果不是202则说明失败
	if req.StatusCode == 202 {
		return true
	} else {
		if DEBUG {
			log.Println("deleteBlobs(): response Status is ", req.Status)
		}
		return false
	}
}

// 解析获取的表单内容，返回repositoryObject
func analysisManifestData(body string) (object repositoryObject) {
	// 转换获取的json数据为map[string]interface{}类型
	bOK, data := getJSONData(body)

	if !bOK {
		return object
	}

	// 解析data中的数据内容
	for key, value := range data {
		//fmt.Println(key, reflect.TypeOf(value), value)
		//fmt.Println(key, reflect.TypeOf(value))
		switch activeValue := value.(type) {
		case float64:
			if key == "schemaVersion" {
				object.schemaVersion = int(activeValue)
			}
		case string:
			if key == "mediaType" {
				object.mediaType = activeValue
			}
		case map[string]interface{}:
			if key == "config" {
				object.config = getBaseDigestObject(activeValue)
			}
		case []interface{}:
			if key == "layers" {
				for _, layer := range activeValue {
					layerData := layer.(map[string]interface{})
					object.layers = append(object.layers, getBaseDigestObject(layerData))
				}
			}
		default:
		}
	}

	return object
}

// 解析并获取baseDigest对象
func getBaseDigestObject(data map[string]interface{}) (dObject baseDigest) {
	for key, value := range data {
		// fmt.Println(key, reflect.TypeOf(value))
		switch activeValue := value.(type) {
		case float64:
			if key == "size" {
				dObject.size = int(activeValue)
			}
		case string:
			if key == "mediaType" {
				dObject.mediaType = activeValue
			} else if key == "digest" {
				dObject.digest = activeValue
			}
		default:
		}
	}

	return
}

// 获取转换后的json数据
func getJSONData(body string) (bool, map[string]interface{}) {
	// 转换获取的json数据为map[string]interface{}类型
	data := make(map[string]interface{})
	var bodybyte []byte = []byte(body)

	// 执行json格式信息的解码操作
	err := json.Unmarshal(bodybyte, &data)
	if err != nil {
		if DEBUG {
			fmt.Println("JSON unmarshal fail, error information: ", err)
			fmt.Println("Body information: ", body)
		}
		return false, data
	} else {
		return true, data
	}
}
