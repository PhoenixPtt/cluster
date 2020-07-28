package header

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return os.IsExist(err)
}

func ReadByteTofile(loadpath string, filename string, bodybyte []byte) error {
	f, err := os.Create(loadpath + filename)
	if err != nil {
		fmt.Println(err)
		return err
	}
	n2, err := f.Write(bodybyte)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return err
	}
	fmt.Println(n2, "bytes written successfully")
	err = f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func TarFile(tarName string) error {
	cmd := exec.Command("tar", "-cvf", tarName+".tar", tarName)

	//创建获取命令输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error:can not obtain stdout pipe for command:%s\n", err)
		return err
	}
	//执行命令
	if err := cmd.Start(); err != nil {
		fmt.Println("Error:The command is err,", err)
		return err
	}
	//读取所有输出
	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		fmt.Println("ReadAll Stdout:", err.Error())
		return err
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println("wait:", err.Error())
		return err
	}
	fmt.Printf("stdout:\n\n %s", bytes)
	return nil

}
