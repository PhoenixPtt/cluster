package header

import (
	"fmt"
	"io/ioutil"
	"os/exec"
)

func ExecCmd(cmdName string, arg ...string) (string, error) {

	cmd := exec.Command(cmdName, arg...)
	//创建获取命令输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error:can not obtain stdout pipe for command:%s\n", err)
		return "", err
	}
	//执行命令
	if err := cmd.Start(); err != nil {
		fmt.Println("Error:The command is err,", err)
		return "", err
	}
	//读取所有输出
	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Println("ReadAll Stdout:", err.Error())
		return "", err
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println("wait:", err.Error())
		return "", err
	}
	// fmt.Printf("stdout:\n\n %s", bytes)
	return string(bytes), nil
}
