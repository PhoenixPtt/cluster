package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"./agentContainer"
	"./agentDraw"
	"./agentImage"

	"../header"
	// "../tcpSocket"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var imageName string
var isSuccess bool
var ctx context.Context
var cli *client.Client
var err error
var handle string

func main() {

	//init docker client
	ctx = context.Background()
	//创建客户端
	cli, err = client.NewClientWithOpts(client.FromEnv /*, client.WithAPIVersionNegotiation()*/)
	if err != nil {
		panic(err)
		return
	}
	cli.NegotiateAPIVersion(ctx) //切换到于当前能够本机运行的API接口
	// log.Println("version99999999999999999", cli.ClientVersion())

	agentImage.ImagePullAdrr = "myregistry.com"
	agentImage.ImageInit(ctx, cli)

	//test imag
	imagemain()

	//test container and tuli
	// containermain()

	// tcpSocket.ConnectToHost("192.168.43.118", 60000, myReceiveData, myStateChange)
	// for {
	// 	time.Sleep(time.Second)
	// }

}

func myReceiveData(h string, i string, s []byte) {
	log.Println("Agent端监听server端数据输出：", h, i /*, s*/)
	if i == "IMAG" {
		var imageData header.ImageData
		// var sendbyte []byte
		err := header.Decode(s, &imageData)
		if err != nil {
			log.Println("decode data false")
		}

		agentImage.RecieveDataFromServer(handle, imageData)
	}
}

func myStateChange(h string, s uint8) {
	handle = h
	log.Println("Agent端监听server状态改变输出：", h, s)
}

func imagemain() {

	// //1 get image list
	// imageSummery, isSuccess := agentImage.ListImage(false)
	// if !isSuccess {
	// 	//deal error
	// 	return
	// }
	// fmt.Println("序号\tTAG\tIMAGE ID\tCREATED\tSIZE")
	// for index, image := range imageSummery {

	// 	fmt.Printf("%d\t\t\t%s\t\t\t%s\t\t\t%d\t\t\t%d\n", index, image.RepoTags, image.ID[7:7+12], image.Created, image.Size)
	// 	imageID := image.ID[7 : 7+12]
	// 	imageRepoTags := image.RepoTags
	// 	//update map list
	// 	agentImage.ImageIDmap[imageID] = imageRepoTags
	// 	for _, val := range imageRepoTags {
	// 		agentImage.ImageNamemap[val] = imageID
	// 	}
	// }

	// //2 inspect image
	imageName = "ubuntu:16.04.3"
	// // imgid := agentImage.GetImageIDofName(imageName)
	// imageInspect, err := agentImage.InspectImage(imageName)
	// if err != nil {
	// 	//deal error
	// 	log.Println(err)
	// 	return
	// }
	// fmt.Println("imageInspect", imageInspect.ID[7:7+12], imageInspect.RepoTags, imageInspect.Created)

	// //3 history image
	// imageHistory, err := agentImage.HistoryImage(imageName)
	// if err != nil {
	// 	//deal error
	// 	return
	// }
	// for key, val := range imageHistory {
	// 	fmt.Println("key:", key, "comment: ", val.Comment, "Created: ", val.Created,
	// 		"CreatedBy : ", val.CreatedBy, "ID :", val.ID, "Size: ", val.Size, "Tags: ", val.Tags)
	// }

	// //4 tag image and update iamgelist
	// sourceName := "mybuildtest:v1.0"
	// targetName := "mybuild:2.0"
	// err = agentImage.TagImage(sourceName, targetName)
	// if !isSuccess {
	// 	//deal error
	// 	return
	// }

	// // 5 push image to registry
	// isSuccess = agentImage.PushImage(targetName, false)
	// if !isSuccess {
	// 	//deal error
	// 	return
	// }

	// //6 remove image
	// RmimageName := "myregistry.com:5000/library/mybuild:2.0"
	// imageDeleteResponse, isSuccess := agentImage.RemoveImage(RmimageName, true, false)
	// if !isSuccess {
	// 	//deal error
	// 	return
	// }
	// for key, val := range imageDeleteResponse {
	// 	fmt.Println("key :", key, "Deleted: ", val.Deleted, "Untagged : ", val.Untagged)
	// }

	// //7 pull image
	// isSuccess = agentImage.PullImage("mybuild:2.0", false)
	// if !isSuccess {
	// 	//deal error
	// 	return
	// }

	// //8 search image
	// term := "k3s"
	// imageSearch, isSuccess := agentImage.SearchImage(term, 6)
	// if !isSuccess {
	// 	//deal error
	// 	return
	// }

	// for key, val := range imageSearch {
	// 	fmt.Println("key :", key, "StarCount : ", val.StarCount, "IsOfficial  : ", val.IsOfficial, "Name :", val.Name, "IsAutomated :", val.IsAutomated, "Description ", val.Description)
	// }

	// //7 save image
	// imageNames := []string{"rancher/library-traefik:1.7.19", "rancher/metrics-server:v0.3.6"}
	// savePath := "/home/cetc15/桌面/"
	// saveName := "rancher.tar.gz"
	// isSuccess = agentImage.SaveImage(imageNames, savePath, saveName)
	// if !isSuccess {
	// 	//deal error
	// 	return
	// }
	// fmt.Println("保存镜像成功!")

	// //8 load image
	// loadPath := "/home/cetc15/下载/images/nginx1.8.tar"
	// // loadName := "nginx1.8.tar"
	// isSuccess = agentImage.LoadImage(loadPath, true)
	// if !isSuccess {
	// }
	// fmt.Println("加载镜像成功")

	// //9 build image
	// sourcePath := "/home/cetc15/dockfileimage/app"
	// tags := []string{"mypingtest:1.0"}
	// isSuccess = agentImage.BuildImage(sourcePath, tags, true)
	// if !isSuccess {
	// 	//deal error
	// 	return
	// }
	// fmt.Println("build镜像成功")

	// isSuccess = agentImage.UploadImageToRegistry("nginx", "nginx:1.7.9")
	// if !isSuccess {
	// 	//deal error
	// 	return
	// }
}

func containermain() {

	agentContainer.ContainerInit(ctx, cli)

	//获取运行中的容器列表
	_, isSuccess := agentContainer.ListofContainer(true)
	if !isSuccess {
		return
	}

	imageName = "bfirsh/reticulate-splines:latest"
	containerName := "mybfirsh"

	containerID := agentContainer.ContainerNameMap[containerName]
	// log.Println("agentContainer.ContainerNameMap", agentContainer.ContainerNameMap)
	if len(containerID) > 0 {
		//停止所有运行中的容器
		var timeout *time.Duration
		stoperr := cli.ContainerStop(ctx, containerID, timeout)
		if stoperr != nil {
			panic(stoperr)
		}
		fmt.Print("container:", containerID, " stopped\n")
		rmerr := cli.ContainerRemove(ctx, containerID[:10], types.ContainerRemoveOptions{Force: true})
		if rmerr != nil {
			panic(rmerr)
		}
	}

	//创建容器
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		// Cmd:          []string{"echo", "hello world"},
		OpenStdin:    true,
		Tty:          true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	//update container
	agentContainer.UpdateContainerMap(containerName, resp.ID)

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// fmt.Println(resp.ID)

	// //容器日志
	// logstr, isSuccess := agentContainer.LogsofContainer("af23c7ffaad9", true, true, "10m", "", true, "10")
	// if !isSuccess {
	// 	return
	// }

	// log.Println("container log", logstr, "\n")

	// get container stats
	// log.Println("ContainerNameMap\n", agentContainer.ContainerNameMap)

	// 使用time.Ticker
	var ticker *time.Ticker = time.NewTicker(1 * time.Second)
	for t := range ticker.C {
		fmt.Println("Tick at", t)
		DrawAllContainerStats()
	}

	// DrawContainerPLot(containerName)

}

func DrawContainerPLot(containerName string) {

	containerID := agentContainer.ContainerNameMap[containerName]
	// 使用time.Ticker
	var ticker *time.Ticker = time.NewTicker(1 * time.Second)
	mtime := time.Now()
	for t := range ticker.C {
		fmt.Println("Tick at", t)

		stasstr, isSuccess := agentContainer.StatsofContainer(containerID[:10], false)
		if !isSuccess {
			return
		}
		// log.Println("stasstr", containerID, stasstr)
		splitstr := strings.Split(stasstr, " ")
		log.Println("container-name:", splitstr[1], "container-cpu:", splitstr[2], "container-mem:", splitstr[6])
		var curValue [2]float64
		curValue[0] = time.Since(mtime).Seconds()
		cpuval := splitstr[2]
		strTofloat, err := strconv.ParseFloat(cpuval[:len(cpuval)-1], 64)
		if err != nil {
			strTofloat = 0.00
		}
		curValue[1] = strTofloat
		agentDraw.DrawPlot("container-cpu", "time", "cpu%", curValue)
	}
}

func DrawAllContainerStats() {

	//获取运行中的容器列表
	containers, isSuccess := agentContainer.ListofContainer(false)
	if !isSuccess {
		return
	}

	// // 使用time.Ticker
	// var ticker *time.Ticker = time.NewTicker(1 * time.Second)
	// // mtime := time.Now()
	// for t := range ticker.C {

	// fmt.Println("Tick at", t)
	var names []string
	var cpudatas []float64
	var memdatas []float64
	// names := []string{"aaaaaaaaaaa", "bbbbbbbbbbbbb", "ccccccccccccccccc"}
	// cpudatas := []float64{20.0, 50.0, 80.0}
	// memdatas := []float64{30.0, 60.0, 88.0}
	for _, container := range containers {
		stasstr, isSuccess := agentContainer.StatsofContainer(container.ID[:10], false)
		if !isSuccess {
			return
		}
		splitstr := strings.Split(stasstr, " ")
		log.Println("container-name:", splitstr[1], "container-cpu:", splitstr[2], "container-mem:", splitstr[6])
		names = append(names, splitstr[1])

		cpudata, cpuerr := strconv.ParseFloat(strings.Split(splitstr[2], "%")[0], 64)
		memdata, memerr := strconv.ParseFloat(strings.Split(splitstr[6], "%")[0], 64)
		fmt.Println("cpudata", cpudata, "memdata", memdata)
		if cpuerr != nil || memerr != nil {
			log.Printf("%s or %s err", cpuerr, memerr)
			return
		}
		cpudatas = append(cpudatas, cpudata)
		memdatas = append(memdatas, memdata)
	}

	timestr := time.Now().String()
	fmt.Println("timestr", timestr)
	// splitstr := strings.Split(timestr, " ")
	agentDraw.DrawBar( /*splitstr[0]+" "+splitstr[1]*/ "cpuVSmem image", names, cpudatas, memdatas)

	// }

}

// * 整个文件读到内存，适用于文件较小的情况
func readAllIntoMemory(filename string) ([]byte, error) {
	fp, err := os.Open(filename) // 获取文件指针
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	fileInfo, err := fp.Stat()
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, fileInfo.Size())
	_, err = fp.Read(buffer) // 文件内容读取到buffer中
	if err != nil {
		return nil, err
	}
	return buffer, nil
}
