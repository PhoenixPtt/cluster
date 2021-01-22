package ctnA

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

//判断镜像是否存在
func IsImageExisted(cli *client.Client, ctx context.Context, imgName string) (bExisted bool) {
	var (
		imageSummery []types.ImageSummary
		err          error
		imgRepo      types.ImageSummary
		img          string
	)

	//初始化变量
	bExisted = false

	//获取本地镜像列表
	if imageSummery, err = cli.ImageList(ctx, types.ImageListOptions{}); err != nil {
		//一般情况下，都能获取成功，因此不做错误判断。
	}

	//遍历本地镜像仓库列表，判断镜像是否存在
	for _, imgRepo = range imageSummery {
		for _, img = range imgRepo.RepoTags {
			if img == imgName {
				bExisted = true
				break
			}
		}
	}
	return
}

//从私有仓库拉去镜像
func ImagePull(cli *client.Client, ctx context.Context, imgName string) (err error) {
	var (
		auth    string
		options types.ImagePullOptions
	)

	//登录镜像仓库
	auth, _ = registryAuth(true, "docker", "27MTjlJyZWD0XxLf7C_SxOLlYpaprdzURn-Ec10Ew-U")

	//拉取镜像
	options.RegistryAuth = auth
	_, err = cli.ImagePull(ctx, imgName, options)

	return
}

func registryAuth(isRegisAuth bool, username string, password string) (authStr string, bSuccess bool) {
	var (
		encodedJSON []byte
		authConfig  types.AuthConfig
		err         error
	)

	//初始化变量
	authConfig = types.AuthConfig{
		Username: username,
		Password: password,
	}
	bSuccess = true

	if isRegisAuth {
		if encodedJSON, err = json.Marshal(authConfig); err != nil {
			bSuccess = false
			return
		}
		authStr = base64.URLEncoding.EncodeToString(encodedJSON)
	}
	return
}
