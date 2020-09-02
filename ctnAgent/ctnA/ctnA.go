package ctnA

import (
	"bytes"
	"context"
	"ctnCommon/ctn"
	"ctnCommon/headers"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

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

//Agent端容器结构体声明
type CTNA struct {
	ctn.CTN
}

//实现容器接口
var (
	ctx context.Context
	Cli *client.Client

	clis []*client.Client
	err  error
)

func init() {
	ctx = context.Background()

	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err.Error())
	}
}

//创建容器
func (pCtn *CTNA) Create() (string, error) {
	var (
		err     error
		errType string
	)

	//容器为空
	errType, err = check(pCtn, ctn.CREATE)
	if err != nil {
		return errType, err
	}

	if pCtn.isCreated() { //如果已经被创建过，则不允许重复创建
		return ERR_TYPE_CTN_EXIST, nil
	}

	imageSummery, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return ERR_TYPE_IMAGE_GETLIST, err
	}

	index_repo := -1
	index_image := -1
	fmt.Println("镜像仓库列表如下所示：")
	for i, repo := range imageSummery {
		fmt.Printf("镜像仓库序号：%d\t镜像仓库：%#v\n", i, repo)
		for j, v := range repo.RepoTags {
			fmt.Printf("\t镜像序号：%d\t镜像名称：%s\n", j, v)
			if v == pCtn.Image { // 假设需要获取os.Args[k], k = 1
				index_image = j
				break
			}
		}
		if index_image != -1 {
			index_repo = i
			break
		}
	}

	if index_repo == -1 && index_image == -1 {
		fmt.Println("本地仓库中不存在镜像imageTag")
		//本地仓库不存在，从私有仓库下载
		auth, _ := registryAuth(true, "docker", "27MTjlJyZWD0XxLf7C_SxOLlYpaprdzURn-Ec10Ew-U")
		var options types.ImagePullOptions
		options.RegistryAuth = auth
		_, err := cli.ImagePull(ctx, pCtn.Image, options)
		if err != nil {
			return ERR_TYPE_IMAGE_PULL, err
		}
		fmt.Println("从私有仓库中Pull镜像成功")
	} else {
		fmt.Println("镜像在本地仓库已存在！")
		fmt.Printf("序号：[%d,%d]\n", index_repo, index_image)
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: pCtn.Image,
		},
		&container.HostConfig{
			NetworkMode: "host",
		},
		nil,
		"")
	if err != nil {
		return ERR_TYPE_CTN_CREATE, err
	}
	pCtn.ID = resp.ID

	ctnInspect, err := pCtn.Inspect()
	if err != nil {
		return ERR_TYPE_CTN_INFO, err
	}
	pCtn.State = ctnInspect.State.Status
	return "", err
}

//启动容器
func (pCtn *CTNA) Start() (string, error) {
	var (
		err     error
		errType string
	)
	//容器为空
	errType, err = check(pCtn, ctn.START)
	if err != nil {
		return errType, err
	}

	if !pCtn.isCreated() {
		errString := "容器尚未创建，请先创建容器！"
		return ERR_TYPE_CTN_NOTEXIST, errors.New(errString)
	}

	if pCtn.isRunning() { //如果已经在运行，则不允许启动
		return ERR_TYPE_CTN_RUNNING, nil
	}

	if err = cli.ContainerStart(ctx, pCtn.ID, types.ContainerStartOptions{}); err != nil {
		return ERR_TYPE_CTN_START, err
	}

	ctnInspect, err := pCtn.Inspect()
	if err != nil {
		return ERR_TYPE_CTN_INFO, err
	}
	pCtn.State = ctnInspect.State.Status

	return "", err
}

//运行容器
func (pCtn *CTNA) Run() (string, error) {
	var (
		err        error
		errType    string
		ctnInspect ctn.CTN_INSPECT
	)
	//容器为空
	errType, err = check(pCtn, ctn.RUN)
	if err != nil {
		return errType, err
	}

	//如果容器ID不存在，则执行创建容器操作，如果容器ID已存在，则忽略
	if errType, err = pCtn.Create(); err != nil {
		return errType, err
	}

	if errType, err = pCtn.Start(); err != nil {
		//启动失败，则删除Create容器的痕迹，如果删除失败，则返回删除失败
		goto ERROR
	}

	ctnInspect, err = pCtn.Inspect()
	if err != nil {
		return ERR_TYPE_CTN_INFO, err
	}
	pCtn.State = ctnInspect.State.Status
	fmt.Printf("container: %s\t\t%s\n", pCtn.ID[:10], "run")

	return "", err

ERROR:
	var errTypeR string
	var errR error
	if errTypeR, errR = pCtn.Remove(); errR != nil {
		return errTypeR, errR
	}
	return errType, err
}

//停止容器
func (pCtn *CTNA) Stop() (string, error) {
	var (
		err     error
		errType string
	)
	//容器为空
	errType, err = check(pCtn, ctn.STOP)
	if err != nil {
		return errType, err
	}

	//判断ctn是否已经正在运行，如果不是正在运行，则不需要停止，直接返回
	if !pCtn.isRunning() {
		return ERR_TYPE_CTN_NOTRUNNING, nil
	}

	//正常停止容器
	var timeout *time.Duration
	err = cli.ContainerStop(ctx, pCtn.ID, timeout)
	if err != nil {
		return ERR_TYPE_CTN_STOP, err
	}

	ctnInspect, err := pCtn.Inspect()
	if err != nil {
		return ERR_TYPE_CTN_INFO, err
	}
	pCtn.State = ctnInspect.State.Status

	fmt.Printf("container: %s\t\t%s\n", pCtn.ID[:10], "normal stopped")
	return "", err
}

//强制停止容器
func (pCtn *CTNA) Kill() (string, error) {
	var (
		err     error
		errType string
	)
	//容器为空
	errType, err = check(pCtn, ctn.KILL)
	if err != nil {
		return errType, err
	}

	//判断ctn是否已经正在运行，如果不是正在运行，则不需要停止，直接返回
	if !pCtn.isRunning() { //如果已经在运行，则不允许启动
		return ERR_TYPE_CTN_NOTRUNNING, nil
	}

	//正常停止容器
	err = cli.ContainerKill(ctx, pCtn.ID, "")
	if err != nil {
		return ERR_TYPE_CTN_KILL, err
	}

	ctnInspect, err := pCtn.Inspect()
	if err != nil {
		return ERR_TYPE_CTN_INFO, err
	}
	pCtn.State = ctnInspect.State.Status
	fmt.Printf("container:%s\t\t%s\n", pCtn.ID[:10], "force stopped\n")
	return "", err
}

//删除容器
func (pCtn *CTNA) Remove() (string, error) {
	var (
		err     error
		errType string
	)

	errType, err = check(pCtn, ctn.REMOVE)
	if err != nil {
		return errType, err
	}

	errType, err = pCtn.Kill()
	if err != nil {
		return errType, err
	}

	//删除该容器
	err = cli.ContainerRemove(ctx, pCtn.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return ERR_TYPE_CTN_REMOVE, err
	}

	pCtn.State = "not existed"
	fmt.Print("container:", pCtn.ID, "\tnormal remove\n")

	return "", err
}

//获取容器日志
//注意：容器被删除之后无法获取容器日志
func (pCtn *CTNA) GetLog() (string, error) {
	var logs io.ReadCloser
	var errType string
	var err error

	errType, err = check(pCtn, ctn.GETLOG)
	if err != nil {
		return errType, err
	}

	logs, err = cli.ContainerLogs(ctx, pCtn.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return ERR_TYPE_CTN_GETLOG, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(logs)
	logStr := buf.String()

	return logStr, err
}

//查看容器详细信息
func (pCtn *CTNA) Inspect() (ctn.CTN_INSPECT, error) {
	var ctnInspect ctn.CTN_INSPECT
	var ctnJson types.ContainerJSON
	var err error

	_, err = check(pCtn, ctn.INSPECT)
	if err != nil {
		return ctnInspect, err
	}

	ctnJson, err = cli.ContainerInspect(ctx, pCtn.ID)
	if err != nil {
		return ctnInspect, err
	}

	var inspectStream []byte
	inspectStream, err = json.Marshal(ctnJson)
	if err != nil {
		return ctnInspect, err
	}

	err = json.Unmarshal(inspectStream, &ctnInspect)
	if err != nil {
		return ctnInspect, err
	}

	ctnInspect.Created = headers.ToLocalTime(ctnInspect.Created)
	ctnInspect.CreatedString = headers.ToString(ctnInspect.Created, headers.TIME_LAYOUT)
	return ctnInspect, err
}

func registryAuth(isRegisAuth bool, username string, password string) (string, bool) {
	//认证
	var authStr string
	if isRegisAuth {
		authConfig := types.AuthConfig{
			Username: username,
			Password: password,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			panic(err)
			return authStr, false
		}
		authStr = base64.URLEncoding.EncodeToString(encodedJSON)
	}
	return authStr, true
}

//判断容器是否已经被创建
func (ctn *CTNA) isCreated() bool {
	if ctn == nil {
		return false
	}

	if ctn.ID != "" {
		return true
	}

	return false
}

//判断容器是否已经被创建
func (ctn *CTNA) isRunning() bool {
	if !ctn.isCreated() {
		return false
	}

	pCtnInspect, err := ctn.Inspect()
	if err != nil {
		return false
	}

	if pCtnInspect.State.Running {
		return true
	}

	return false
}

func check(pCtn *CTNA, oper string) (string, error) {
	var err error
	if pCtn == nil {
		switch oper {
		case ctn.CREATE:
			err = errors.New("容器指针为空，无法创建该容器")
		case ctn.START:
			err = errors.New("容器指针为空，无法启动该容器")
		case ctn.RUN:
			err = errors.New("容器指针为空，无法运行该容器")
		case ctn.STOP:
			err = errors.New("容器指针为空，无法停止该容器")
		case ctn.KILL:
			err = errors.New("容器指针为空，无法强杀该容器")
		case ctn.REMOVE:
			err = errors.New("容器指针为空，无法删除该容器")
		case ctn.GETLOG:
			err = errors.New("容器指针为空，无法获取该容器日志")
		case ctn.INSPECT:
			err = errors.New("容器指针为空，无法获取该容器详细信息")
		}
		return ERR_TYPE_CTN_NOTEXIST, err
	}
	return "", nil
}
