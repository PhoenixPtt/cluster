package agentImage

import (
	// "bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"clusterHeader"
	"targz"
	"tcpSocket"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
)

var Ctx context.Context
var Cli *client.Client
var ImageIDmap map[string][]string //key imageID
var ImageNamemap map[string]string //key imageName
var ImagePullAdrr string
var UserName string
var PassWord string
var DockerfilePath = "/home/cetc15/dockfileimage/"
var ImageSavePath = "/home/cetc15/桌面/"
//var ImageLoadPath = "/home/cetc15/下载/images/"
var ImageLoadPath = "/home/cetc15/test/"

const (
	TagImages = iota //0
	RmiImages        //1
	PullImages
	BuildImages
)

//初始化agent
func ImageInit() {
	ImageIDmap = make(map[string][]string) //让imageidMaplist可编辑
	ImageNamemap = make(map[string]string) //让imageidMaplist可编辑
	Ctx = context.Background()
	Cli, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	ImagePullAdrr = "myregistry.com"
	UserName = "docker"
	PassWord = "27MTjlJyZWD0XxLf7C_SxOLlYpaprdzURn-Ec10Ew-U"

	//获取镜像列表，对镜像map参数进行初始化
	imageSummery, err := ListImage(false)
	if err != nil {
		//deal error
		fmt.Println("获取镜像列表失败！")
		return
	}

	if len(imageSummery) <= 0 {
		return
	}
	fmt.Println("序号\tTAG\tIMAGE ID\tCREATED\tSIZE")
	for index, image := range imageSummery {

		fmt.Printf("%d\t\t\t%s\t\t\t%s\t\t\t%d\t\t\t%d\n", index, image.RepoTags, image.ID[7:7+12], image.Created, image.Size)
		imageID := image.ID[7 : 7+12]
		imageRepoTags := image.RepoTags
		//update map list
		ImageIDmap[imageID] = imageRepoTags
		for _, val := range imageRepoTags {
			ImageNamemap[val] = imageID
		}
	}

}

//获取镜像列表
func ListImage(isShowChild bool) ([]types.ImageSummary, error) {

	mtime := time.Now()
	imageSummery, err := Cli.ImageList(Ctx, types.ImageListOptions{All: isShowChild})
	if err != nil {
		log.Printf(err.Error())
		return imageSummery, err
	}
	log.Printf("get imagelist time is :%s\n imagelist number is %d\n", time.Since(mtime), len(imageSummery))
	return imageSummery, nil
}

//构建镜像
func BuildImage(sourcePath string, imageName string, tags []string, remove bool) error {

	var err error

	Cli.SetCustomHTTPHeaders(map[string]string{"Content-type": "application/x-tar"})
	//tar file
	splitstr := strings.Split(sourcePath, "/")
	num := len(splitstr)
	var dirstr []string
	for _, val := range splitstr[0 : num-1] {
		dirstr = append(dirstr, val)
	}
	tarName := splitstr[len(splitstr)-1]
	baseFile := strings.Join(dirstr, "/")
	fmt.Println(tarName, baseFile)

	tarfilename := sourcePath + ".tar"
	err = targz.TarGz(sourcePath, tarfilename, true)
	if err != nil {
		//use exec to tar file
		os.Chdir(baseFile)
		mytarerr := header.TarFile(tarName)
		if mytarerr != nil {
			return mytarerr
		}
	}

	dockerBuildContext, err := os.Open(tarfilename)
	if err != nil {
		log.Printf(err.Error())
		return err
	}
	defer dockerBuildContext.Close()

	dockerfile := tarName + "/Dockerfile"
	var imageNameList []string
	for _, tag := range tags {
		imageNameList = append(imageNameList, imageName+":"+tag)
	}
	options := types.ImageBuildOptions{
		Tags:       imageNameList,
		Remove:     remove,
		Dockerfile: dockerfile,
	}

	log.Println(dockerfile)
	mtime := time.Now()
	imagebuildresp, err := Cli.ImageBuild(Ctx, dockerBuildContext, options)
	if err != nil {
		log.Printf(err.Error())
		return err
	}
	image, err := ioutil.ReadAll(imagebuildresp.Body)
	if err != nil {
		log.Printf(err.Error())
		return err

	}

	log.Printf("build image time is : %s\n body is %s", time.Since(mtime), string(image))
	return nil
}

//通过客户端的字节流（应用程序）构建镜像
//func BuildImageOfBinaryProcess(imageName string, tags []string, bodybyte []byte, remove bool) error {
func BuildImageOfBinaryProcess(imageName string, tags []string, filename string, remove bool) error {
	var err error
	err = os.Chdir(ImageLoadPath)
	if err != nil {
		return err
	}
	//判断可执行文件是否存在
	execfile := "exe-to-app.sh"
	_, filestat := os.Stat(execfile)

	if filestat != nil {
		dstFile, fileerr := os.Create(execfile)
		if fileerr != nil {
			return fileerr
		}
		content := "#!/bin/sh\n# 可执行程序名\nappname=$1\n# 目标文件夹\ndst=\"./app\"\n# 利用 ldd 提取依赖库的具体路径\n" +
			"liblist=$(ldd $appname | awk '{ if (match($3,\"/\")){ printf(\"%s \"), $3 } }')\n# 目标文件夹的检测\n" +
			"if [ ! -d $dst ];then\nmkdir $dst\nfi\n# 拷贝库文件和可执行程序到目标文件夹\ncp $liblist $dst\nmv $appname $dst\ncp $appname $dst\n\n" +
			"chmod -R 777 $dst\ncd $dst\n\n#写dockerfile文件\necho FROM ubuntu:16.04.3 > Dockerfile\necho MAINTAINER Docker CETC15 >> Dockerfile\n" +
			"echo ADD ./app/* /lib/ >> Dockerfile\necho ADD ./app/$appname /a.out >> Dockerfile\necho WORKDIR / >> Dockerfile\n" +
			"#echo CMD [ \"mv\",\"/lib/\"$appname,\"/\" ] >> Dockerfile\necho CMD [ '\"./a.out\"' ] >> Dockerfile\n\n" +
			"chmod -R 777 Dockerfile\n#mv Dockerfile $dst"
		_,writeerr := dstFile.WriteString(content + "\n")
		dstFile.Close()
		if(writeerr != nil){
			return  writeerr
		}
	}

	_, execerr := header.ExecCmd("chmod", "777", execfile)
	if execerr != nil {
		return execerr
	}

	Cli.SetCustomHTTPHeaders(map[string]string{"Content-type": "application/x-tar"})
	tarName := "app"

	//readerr := header.ReadByteTofile(DockerfilePath, imageName, bodybyte)
	//if readerr != nil {
	//	return readerr
	//}

	//exec image file create the binary process cli
	_, err = header.ExecCmd("./exe-to-app.sh", filename)
	if err != nil {
		return err
	}

	//tar file
	sourcePath := ImageLoadPath + tarName
	tarfilename := sourcePath + ".tar"
	err = targz.TarGz(sourcePath, tarfilename, true)
	if err != nil {
		//use exec to tar file
		err = header.TarFile(tarName)
		if err != nil {
			return err
		}
	}
	dockerBuildContext, err := os.Open(tarfilename)
	if err != nil {
		log.Printf(err.Error())
		return err
	}
	defer dockerBuildContext.Close()

	dockerfile := tarName + "/Dockerfile"
	var imageNameList []string
	for _, tag := range tags {
		imageNameList = append(imageNameList, imageName+":"+tag)
	}

	options := types.ImageBuildOptions{
		Tags:       imageNameList,
		Remove:     remove,
		Dockerfile: dockerfile,
	}

	mtime := time.Now()
	imagebuildresp, err := Cli.ImageBuild(Ctx, dockerBuildContext, options)
	if err != nil {
		log.Printf(err.Error())
		return err
	}

	image, err := ioutil.ReadAll(imagebuildresp.Body)
	if err != nil {
		log.Printf(err.Error())
		return err

	}
	log.Printf("成功构建镜像的时间是 : %s\n body is %s", time.Since(mtime), string(image))

	for _, image := range imageNameList {
		err := UpdateImage(BuildImages, image, "")
		if err != nil {
			log.Println("更新镜像列表失败")
		}
	}

	return nil
}
func InspectImage(imageName string) (types.ImageInspect, error) {

	mtime := time.Now()
	imageInspect, bodyStr, err := Cli.ImageInspectWithRaw(Ctx, imageName)
	if err != nil {
		log.Printf(err.Error())
		return imageInspect, err

	}

	log.Printf("discribe image time is :%s\n %s\n", time.Since(mtime), bodyStr)
	return imageInspect, nil

}

func HistoryImage(imageID string) ([]image.HistoryResponseItem, error) {

	mtime := time.Now()
	imageHistory, err := Cli.ImageHistory(Ctx, imageID)
	if err != nil {
		log.Printf(err.Error())
		return imageHistory, err
	}

	log.Printf("discribe image time is :%s\n bodyStr is %s\n ", time.Since(mtime))
	return imageHistory, nil

}

func TagImage(source string, target string, isUpload bool) error {

	var targetname string
	if isUpload { //是否上传至私有仓库
		targetname = ImagePullAdrr + ":5000/" + target
	} else {
		targetname = target
	}
	mtime := time.Now()
	err := Cli.ImageTag(Ctx, source, targetname)
	if err != nil {
		log.Printf(err.Error())
		return err
	}

	log.Printf("tag image time is :%s\n", time.Since(mtime))
	updateerr := UpdateImage(TagImages, source, targetname)
	if updateerr != nil {
		log.Println("更新镜像列表失败")
	}
	return nil

}

func RemoveImage(imageName string, force bool, pruneChildren bool) ([]types.ImageDeleteResponseItem, error) {

	mtime := time.Now()
	imageDeleteResponse, err := Cli.ImageRemove(Ctx, imageName, types.ImageRemoveOptions{Force: force, PruneChildren: pruneChildren})
	if err != nil {
		log.Printf(err.Error())
		return imageDeleteResponse, err
	}

	log.Printf("remove image time is :%s\n", time.Since(mtime))
	rmerr := UpdateImage(RmiImages, imageName, "")
	if rmerr != nil {
		log.Println("更新镜像列表失败")
	}
	return imageDeleteResponse, nil

}

func SearchImage(term string, limit int) ([]registry.SearchResult, error) {

	mtime := time.Now()
	imageSearch, err := Cli.ImageSearch(Ctx, term, types.ImageSearchOptions{Limit: limit})
	if err != nil {
		log.Printf(err.Error())
		return imageSearch, err
	}

	log.Printf("search image time is :%s\n", time.Since(mtime))
	return imageSearch, nil

}

func SaveImage(imageNames []string, savePath string, saveName string) error {

	var err error
	mtime := time.Now()

	ioreadcloser, err := Cli.ImageSave(Ctx, imageNames)
	if err != nil {
		log.Printf(err.Error())
		return err
	}
	defer ioreadcloser.Close()

	image, err := ioutil.ReadAll(ioreadcloser)
	if err != nil {
		log.Printf(err.Error())
		return err
	}

	err = ioutil.WriteFile(savePath+saveName, image, 0644)
	if err != nil {
		log.Printf(err.Error())
		return err
	}

	log.Printf("save image time is :%s\n", time.Since(mtime))
	return nil

}

func SaveImageToAgent(imageNames []string, savePath string, saveName string) ([]byte, error) {

	mtime := time.Now()

	ioreadcloser, err := Cli.ImageSave(Ctx, imageNames)
	if err != nil {
		log.Printf(err.Error())
		return []byte(""), err
	}
	defer ioreadcloser.Close()

	image, err := ioutil.ReadAll(ioreadcloser)
	if err != nil {
		log.Printf(err.Error())
		return []byte(""), err
	}

	log.Printf("save image time is :%s\n", time.Since(mtime))
	return image, nil

}

func LoadImage(loadPath string, quiet bool) error {

	mtime := time.Now()
	bufReader, err := ioutil.ReadFile(loadPath)
	if err != nil {
		log.Printf(err.Error())
		return err
	}

	imageLoadResponse, err := Cli.ImageLoad(Ctx, bytes.NewReader(bufReader), quiet)
	if err != nil {
		log.Printf(err.Error())
		return err
	}

	defer imageLoadResponse.Body.Close()
	log.Printf("load image time is :%s\n", time.Since(mtime))
	return nil

}

func GetImageIDofName(imageName string) string {

	imageID := ImageNamemap[imageName]

	return imageID

}

func UpdateImage(updateType int, sourceName string, targetName string) error {

	imageSummery, err := ListImage(false)
	if err != nil {
		//deal error
		return err
	}
	// get srcImageID according sourceName
	srcImageID := ImageNamemap[sourceName]
	log.Println("UpdateImage", PullImages, srcImageID)
	for _, image := range imageSummery {

		imageID := image.ID[7 : 7+12]
		tags := image.RepoTags

		//if type is PullImages,need to give value to srcImageID
		if srcImageID == "" {
			for _, tag := range tags {
				if strings.Contains(tag, sourceName) {
					srcImageID = imageID
				}
			}
		}

		//update current idmap
		if imageID == srcImageID {
			delete(ImageIDmap, imageID)
			ImageIDmap[imageID] = tags
			break
		}
	}

	switch updateType {
	case TagImages:
		ImageNamemap[targetName] = srcImageID
		fmt.Println("TagImages", sourceName, srcImageID, targetName, ImageNamemap[targetName], ImageIDmap[srcImageID], len(ImageIDmap))
	case RmiImages:
		delete(ImageNamemap, sourceName)
		fmt.Println("RmiImages", sourceName, srcImageID, ImageIDmap[srcImageID], len(ImageIDmap))
	case PullImages:
		ImageNamemap[sourceName] = srcImageID
		fmt.Println("PullImages", sourceName, srcImageID, ImageNamemap[sourceName], ImageIDmap[srcImageID], len(ImageIDmap))
	case BuildImages:
		ImageNamemap[sourceName] = srcImageID
		fmt.Println("BuildImages", sourceName, srcImageID, ImageNamemap[sourceName], ImageIDmap[srcImageID], len(ImageIDmap))
	default:
		log.Panicln("nothing match!")
	}

	fmt.Println("更新镜像列表成功！\n")

	return nil

}

func GetImageListNum() int {

	// imageSummery, err := ImageList(ctx, cli, false)
	// if err != nil {
	// 	//deal error
	// 	return 0
	// }

	return len(ImageIDmap)
}

func RegistryAuth(isRegisAuth bool, username string, password string) (string, error) {
	//认证
	var authStr string
	if isRegisAuth {
		authConfig := types.AuthConfig{
			Username: username,
			Password: password,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			log.Printf(err.Error())
			return authStr, err
		}
		authStr = base64.URLEncoding.EncodeToString(encodedJSON)
	}
	return authStr, nil

}
func PullImage(imageName string, all bool /*, isRegisAuth bool, username string, password string*/) error {

	if !strings.Contains(imageName, ":") && !all {
		imageName = imageName + ":latest"
	}

	authStr, err := RegistryAuth(true, UserName, PassWord)
	if err != nil {
		return err
	}
	adrr := ImagePullAdrr + ":5000/" + imageName
	//pull镜像
	reader, err := Cli.ImagePull(Ctx, adrr, types.ImagePullOptions{All: all, RegistryAuth: authStr})
	if err != nil {
		log.Printf(err.Error())
		return err
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)
	updateerr := UpdateImage(PullImages, adrr, "")
	if updateerr != nil {
		log.Println("更新镜像列表失败")
	}
	return nil

}

func PushImage(imageName string, all bool) error {

	var adrr string
	if !strings.Contains(imageName, ImagePullAdrr) {
		adrr = ImagePullAdrr + ":5000/" + imageName
	} else {
		adrr = imageName
	}

	authStr, err := RegistryAuth(true, UserName, PassWord)
	if err != nil {
		return err
	}
	reader, err := Cli.ImagePush(Ctx, adrr, types.ImagePushOptions{All: all, RegistryAuth: authStr})
	if err != nil {
		log.Printf(err.Error())
		return err
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)
	// isSuccess := header.ExecCmd("docker", "push", adrr)
	// if !isSuccess {
	// 	return false
	// }
	return nil
}

//func UploadImageToRegistry(handle string, pkgId uint16, imageName string, tags []string, imagebody []byte) error {
func UploadImageToRegistry(handle string, pkgId uint16, imageName string, tags []string, filename string) error {
	var err error
	dealType := header.FLAG_IMAG_LOAD
	////add image to docker image list
	//loadPath := ImageLoadPath + imageName + ".tar"
	//ioreader := bytes.NewBuffer(imagebody)
	//err = targz.Tar(ioreader, loadPath)
	//if err != nil {
	//	returnResultToServer(handle, pkgId, dealType, imageName, tags, "", "FALSE", "LOAD操作中，压缩字节流的过程失败！"+err.Error())
	//	return err
	//}
	loadPath := ImageLoadPath + filename
	//load image
	err = LoadImage(loadPath, true)
	if err != nil {
		//返回客户端结果
		returnResultToServer(handle, pkgId, dealType, imageName, tags, "", "FALSE", "LOAD操作中，导入镜像的过程失败！"+err.Error())
		return err
	}

	log.Println("load镜像成功！")
	//tag image
	for _, tag := range tags {
		tagName := imageName + ":" + tag
		err = TagImage(tagName, tagName, true)
		if err != nil {
			//返回客户端结果
			errstr := "LOAD操作中，load成功后，tag" + tagName + "镜像的过程失败！"
			returnResultToServer(handle, pkgId, dealType, imageName, tags, "", "FALSE", errstr+err.Error())
			return err
		}
		//push image to registry
		err = PushImage(tagName, false)
		if err != nil {
			//返回客户端结果
			_, rmerr := RemoveImage(tagName, true, false)
			if rmerr != nil {
				//deal error
				rmerrstr := "LOAD操作中，load成功后，tag" + tagName + "成功," + "push失败后，未能删除被tag的镜像！"

				returnResultToServer(handle, pkgId, dealType, imageName, tags, "", "FALSE", rmerrstr+rmerr.Error())
				return rmerr
			}
			pusherr := "LOAD操作中，load成功后，tag" + tagName + "成功," + "push失败后，已经删除被tag的镜像！"
			returnResultToServer(handle, pkgId, dealType, imageName, tags, "", "FALSE", pusherr+err.Error())
			return err
		}
	}

	log.Println("uploadImage success")

	return nil

}

func RecieveDataFromServer(handle string, pkgId uint16, imagedata header.ImageData) {

	dealType := imagedata.DealType
	imagename := imagedata.ImageName
	tags := imagedata.Tags

	fmt.Println("imagedataimagedataimagedata6666666666666666", imagedata)
	if (dealType == "") || (imagename == "") {
		returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", "镜像信息不全，请重新操作！")
		return
	}
	if len(tags) <= 0 {
		tags = append(tags, "latest")
	}
	imagebody := imagedata.ImageBody

	var sendData string
	switch dealType {
	case header.FLAG_IMAG_BUID:
		// build image 先将接收到的base64编码的二进制文件进行解码
		//basebd, _ := base64.StdEncoding.DecodeString(imagebody)
		if strings.Contains(imagebody, ".") {
			returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", "BUILD操作中，导入的文件非二进制文件！")
			return
		}
		err := BuildImageOfBinaryProcess(imagename, tags, imagebody, true)
		if err != nil {
			//返回客户端结果
			returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", "BUILD操作中，构建镜像过程失败！"+err.Error())
			return
		}

		for _, tag := range tags {
			tagName := imagename + ":" + tag
			err := TagImage(tagName, tagName, true)
			if err != nil {
				//返回客户端结果
				returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", "BUILD操作中，tag镜像过程失败！"+err.Error())
				return
			}

			//push
			pushname := ImagePullAdrr + ":5000/" + tagName
			pusherr := PushImage(pushname, false)
			if pusherr != nil {
				//deal error remove image
				_, rmerr := RemoveImage(pushname, true, false)
				if rmerr != nil {
					//返回客户端结果
					returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", "BUILD操作中，删除tag的镜像过程失败！"+rmerr.Error())
					return
				}
			}
			log.Println("uploadImage success")
			//delete local image
			_, rmerr := RemoveImage(tagName, true, true)
			if rmerr != nil {
				//返回客户端结果
				returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", "BUILD操作中，删除被tag的镜像过程失败！"+rmerr.Error())
				return
			}
		}
		//删除app两个文件
		//_, execerr := header.ExecCmd("rm", "-rf", DockerfilePath+"app", DockerfilePath+"app.tar")
		//if execerr != nil {
		//	log.Println("删除app两个文件失败")
		//}
		log.Println("文件删除成功")
		sendData = "agent端构建镜像" + imagename + ":[" + strings.Join(tags, ",") + "] 成功"
	case header.FLAG_IMAG_LOAD:
		//basebd,_ := base64.StdEncoding.DecodeString(imagebody)

		if !strings.Contains(imagebody, "tar") {
			returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", "LOAD操作中，导入的文件格式错误！")
			return
		}
		err := UploadImageToRegistry(handle, pkgId, imagename, tags, imagebody) //[]byte(imagebody)
		if err == nil {
			log.Println("agent端加载镜像成功")
			sendData = "agent端加载镜像" + imagename + ":[" + strings.Join(tags, ",") + "] 成功"
		} else {
			returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", "LOAD操作失败！"+err.Error())
			return
		}
	case header.FLAG_IMAG_PUSH:
		//update image In registry imageName string, tags []string
		for _, tag := range tags {
			//push image to registry
			tagName := imagename + ":" + tag
			pusherr := PushImage(tagName, false)
			if pusherr != nil {
				//返回客户端结果
				returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", "PUSH操作中，推送镜像的过程失败！"+pusherr.Error())
				return
			}
		}
		log.Println("agent端推送镜像成功")
		sendData = "agent端推送镜像" + imagename + ":[" + strings.Join(tags, ",") + "] 成功"
	case header.FLAG_IMAG_SAVE:
		saveName := imagename + ".tar"
		bodybyte, err := SaveImageToAgent(tags, ImageSavePath, saveName)
		if err != nil {
			//返回客户端结果
			returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", "SAVE操作中，保存镜像的过程失败！"+err.Error())
			return
		}
		log.Println("agent端保存镜像成功")
		imagebody = string(bodybyte)
		sendData = "agent端保存镜像" + imagename + ":[" + strings.Join(tags, ",") + "] 成功"
		//sendData = string(bodybyte)
	// case header.DELETE:
	// 	for _, tag := range tags {
	// 		tagName := imagename + ":" + tag
	// 		// RmimageName := ImagePullAdrr + ":5000/" + tagName
	// 		imageDeleteResponse, err := RemoveImage(tagName, true, true)
	// 		if err != nil {
	// 			//返回客户端结果
	// 			returnResultToServer(handle, dealType, imagename, tags, []byte("DELETE操作中，删除"+RmimageName+"镜像的过程失败！"), "FALSE", err)
	// 			return
	// 		}
	// 		for key, val := range imageDeleteResponse {
	// 			fmt.Println("key :", key, "Deleted: ", val.Deleted, "Untagged : ", val.Untagged)
	// 		}
	// 	}
	// 	log.Println("agent端删除镜像成功")
	// 	sendData = "agent端删除镜像成功"
	case header.FLAG_IMAG_DIST:
		for _, tag := range tags {
			tagName := imagename + ":" + tag
			// distractName := ImagePullAdrr + ":5000/" + tagName
			//判断镜像是否存在
			//不存在，将镜像tag后再push
			err := PullImage(tagName, false)
			if err != nil {
				//返回客户端结果
				errstr := "DISTRACT操作中，分发" + tagName + "镜像的过程失败！"
				returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", errstr+err.Error())
				return
			}
		}
		log.Println("agent端分发镜像成功")
		sendData = "agent端分发镜像" + imagename + ":[" + strings.Join(tags, ",") + "] 成功"
	// case header.TAG:
	// 	for _, tag := range tags {
	// 		tagName := imagename + ":" + tag
	// 		err := TagImage(tagName, tagName, true)
	// 		if err != nil {
	// 			//返回客户端结果
	// 			returnResultToServer(handle, pkgId, dealType, imagename, tags, "TAG操作中，tag"+tagName+"镜像的过程失败！", "FALSE", err)
	// 			return
	// 		}
	// 	}
	// 	log.Println("agent端标签镜像成功")
	// 	sendData = "agent端标签镜像成功"
	default:
		returnResultToServer(handle, pkgId, dealType, imagename, tags, "", "FALSE", "没有匹配的操作类型")
		return
	}

	if dealType != header.FLAG_IMAG_SAVE {
		imagebody = ""
	}
	returnResultToServer(handle, pkgId, dealType, imagename, tags, imagebody, "SUCCESS", sendData)

}

func returnResultToServer(handle string, pkgId uint16, dealType string, imagename string, tags []string, imagebody string, result string, err string) {
	if dealType == header.FLAG_IMAG_BUID {
		//删除生成的文件
		fileerr := deleteDockerfile()
		if fileerr != "" {
			err = err + "fileerr: " + fileerr
		}
	}
	newdata := header.ImageData{}.From(dealType, imagename, tags, imagebody, result, err)
	sendbyte := header.JsonByteArray(newdata)

	log.Println("agent端返回给server端的数据", dealType, imagename, tags, result, err)
	var grade = tcpSocket.TCP_TPYE_CONTROLLER
	if dealType == header.FLAG_IMAG_SAVE {
		grade = tcpSocket.TCP_TYPE_FILE
	}

	tcpSocket.WriteData(handle, grade, pkgId, "IMAG", sendbyte)
}


func FileChecker(filename string) bool {
	//file_path := BasePath + filename
	_, err := os.Stat(filename)
	if err == nil {
		return true
	} else {
		return false
	}
}
func deleteDockerfile() string {
	fileNameList := []string{ImageLoadPath + "app", ImageLoadPath + "app.tar"}
	errstr := []string{}
	for _, value := range fileNameList {
		if(FileChecker(value)){
			_, execerr := header.ExecCmd("rm", "-rf", value)
			if execerr != nil {
				errstr = append(errstr, "删除"+value+"失败: "+execerr.Error())
			}
		}
	}
	if len(errstr) <= 0 {
		return strings.Join(errstr, ",")
	}
	log.Println("文件删除成功")
	return ""
}
