package agentContainer

import (
	"context"
	"fmt"

	"io"
	// "encoding/json"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"

	"strings"
	"time"

	// "github.com/docker/docker/pkg/stdcopy"
	// "os"
	"bufio"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var Ctx context.Context
var Cli *client.Client
var isSuccess bool
var ContainerNameMap map[string]string //key ContainerName
var containerName string

type StatsResult struct {
	read          string       `json:"read"`
	preread       string       `json:"preread"`
	pids_stats    PidStats     `json:"pids_stats"`
	name          string       `json:"name"`
	id            string       `json:"id"`
	blkio_stats   BlkioStats   `json:"blkio_stats"`
	num_procs     int          `json:"num_procs"`
	storage_stats StorageStats `json:"storage_stats"`
	cpu_stats     CpuStats     `json:"cpu_stats"`
	precpu_stats  CpuStats     `json:"precpu_stats"`
	memory_stats  MemStats     `json:"memory_stats"`
	networks      NetWorks     `json:"networks"`
}

type PidStats struct {
	current int `json:"current"`
}

type BlkioStats struct {
	io_service_bytes_recursive []string `json:"io_service_bytes_recursive"`
	io_serviced_recursive      []string `json:"io_serviced_recursive"`
	io_queue_recursive         []string `json:"io_queue_recursive"`
	io_service_time_recursive  []string `json:"io_service_time_recursive"`
	io_wait_time_recursive     []string `json:"io_wait_time_recursive"`
	io_merged_recursive        []string `json:"io_merged_recursive"`
	io_time_recursive          []string `json:"io_time_recursive"`
	sectors_recursive          []string `json:"sectors_recursive"`
}

type StorageStats struct {
}
type CpuStats struct {
	total_usage         int64          `json:"total_usage"`
	percpu_usage        [8]int64       `json:"percpu_usage"`
	usage_in_kernelmode int64          `json:"usage_in_kernelmode"`
	usage_in_usermode   int64          `json:"usage_in_usermode"`
	system_cpu_usage    int64          `json:"system_cpu_usage"`
	online_cpus         int            `json:"online_cpus"`
	throttling_data     ThrottlingData `json:"throttling_data"`
}

type ThrottlingData struct {
	periods           int `json:"periods"`
	throttled_periods int `json:"throttled_periods"`
	throttled_time    int `json:"throttled_time"`
}

type MemStats struct {
	usage     int64   `json:"usage"`
	max_usage int64   `json:"max_usage"`
	limit     int64   `json:"limit"`
	stats     Memstat `json:"stats"`
}

type Memstat struct {
	active_anon               int64 `json:"active_anon"`
	active_file               int64 `json:"active_file"`
	cache                     int64 `json:"cache"`
	dirty                     int64 `json:"dirty"`
	hierarchical_memory_limit int64 `json:"hierarchical_memory_limit"`
	hierarchical_memsw_limit  int64 `json:"hierarchical_memsw_limit"`
	inactive_anon             int64 `json:"inactive_anon"`
	inactive_file             int64 `json:"inactive_file"`
	mapped_file               int64 `json:"mapped_file"`
	pgfault                   int64 `json:"pgfault"`
	pgmajfault                int64 `json:"pgmajfault"`
	pgpgin                    int64 `json:"pgpgin"`
	pgpgout                   int64 `json:"pgpgout"`
	rss                       int64 `json:"rss"`
	rss_huge                  int64 `json:"rss_huge"`
	total_active_anon         int64 `json:"total_active_anon"`
	total_active_file         int64 `json:"total_active_file"`
	total_cache               int64 `json:"total_cache"`
	total_dirty               int64 `json:"total_dirty"`
	total_inactive_anon       int64 `json:"total_inactive_anon"`
	total_inactive_file       int64 `json:"total_inactive_file"`
	total_mapped_file         int64 `json:"total_mapped_file"`
	total_pgfault             int64 `json:"total_pgfault"`
	total_pgmajfault          int64 `json:"total_pgmajfault"`
	total_pgpgin              int64 `json:"total_pgpgin"`
	total_pgpgout             int64 `json:"total_pgpgout"`
	total_rss                 int64 `json:"total_rss"`
	total_rss_huge            int64 `json:"total_rss_huge"`
	total_unevictable         int64 `json:"total_unevictable"`
	total_writeback           int64 `json:"total_writeback"`
	unevictable               int64 `json:"unevictable"`
	writeback                 int64 `json:"writeback"`
}

type NetWorks struct {
	eth0 Eth0 `json:"eth0"`
}

type Eth0 struct {
	rx_bytes   int64 `json:"rx_bytes"`
	rx_packets int64 `json:"rx_packets"`
	rx_dropped int64 `json:"rx_dropped"`
	tx_bytes   int64 `json:"tx_bytes"`
	tx_packets int64 `json:"tx_packets"`
	tx_errors  int64 `json:"tx_errors"`
	tx_dropped int64 `json:"tx_dropped"`
}

func ContainerInit(ctx context.Context, cli *client.Client) {
	ContainerNameMap = make(map[string]string) //让ContainerNameMap可编辑
	Ctx = ctx
	Cli = cli
}

func ListofContainer(all bool) ([]types.Container, bool) {

	mtime := time.Now()
	//获取运行中的容器列表
	containers, err := Cli.ContainerList(Ctx, types.ContainerListOptions{All: all})
	if err != nil {
		panic(err)
		isSuccess = false
	}
	log.Printf("get container logs time is :%s\n", time.Since(mtime))

	for key, val := range containers {
		for _, name := range val.Names {
			ContainerNameMap[name[1:]] = val.ID
		}
		fmt.Println("key:", key, val.ID, val.Names)
	}
	isSuccess = true
	return containers, isSuccess

}
func LogsofContainer(container string, showStdout bool, showStderr bool, since string, until string, timestamps bool, tail string) (string, bool) {
	mtime := time.Now()
	//容器日志
	out, err := Cli.ContainerLogs(Ctx, container, types.ContainerLogsOptions{
		ShowStdout: showStdout,
		ShowStderr: showStderr,
		Since:      since,
		Until:      until,
		Timestamps: timestamps,
		Follow:     false,
		Tail:       tail,
		Details:    false,
	})

	if err != nil {
		panic(err)
		isSuccess = false
	}
	log.Printf("get container logs time is :%s\n", time.Since(mtime))

	// //输出容器日志
	// stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	body, err := ioutil.ReadAll(out)
	if err != nil {
		// handle error
		fmt.Println("ioutil.ReadAll err=", err)
		isSuccess = false
	}

	logstr := string(body)
	isSuccess = true
	return logstr, isSuccess
}

func StatsofContainer(container string, stream bool) (string, bool) {

	mtime := time.Now()

	body, isSuccess := ExecCmd("docker", "stats", "--no-stream", container /*, "grep", container*/)
	cmdstr := delete_extra_space(body)
	// out, err := Cli.ContainerStats(Ctx, container, stream)

	// if err != nil {
	// 	panic(err)
	// 	isSuccess = false
	// }
	log.Printf("get container stats time is :%s\n", time.Since(mtime))

	// body, err := ioutil.ReadAll(out.Body)
	// if err != nil {
	// 	// handle error
	// 	fmt.Println("ioutil.ReadAll err=", err)
	// 	isSuccess = false
	// }

	// bosystr := string(body)

	// log.Println("insert to transfer json", bosystr)
	// // stdcopy.StdCopy(os.Stdout, os.Stderr, outa.Body)

	// var jsondata StatsResult
	// errjson := json.Unmarshal([]byte(bosystr), &jsondata)
	// if errjson != nil {
	// 	// handle error
	// 	fmt.Println("errjson err=", errjson)
	// 	isSuccess = false
	// }
	// fmt.Println("jsondata", jsondata)
	// // cpustats := jsondata.cpu_stats
	// // syscpu := jsondata.system_cpu_usage
	// // memorystats := jsondata.memory_stats

	// // // 打印索引和其他数据
	// // fmt.Println(`cpu total_usage`, cpustats.total_usage, "\n")
	// // fmt.Println(`system_cpu_usage`, syscpu, "\n")
	// // fmt.Println(`mem usage`, memorystats.usage, "\n")
	// // fmt.Println(`mem max_usage`, memorystats.max_usage, "\n")
	// isSuccess = true
	return cmdstr, isSuccess
}

func UpdateContainerMap(containerName string, containerid string) {
	ContainerNameMap[containerName] = containerid
}

func ExecCmd(cmdName string, arg ...string) (string, bool) {

	var cmdString string
	cmd := exec.Command(cmdName, arg...)
	//创建获取命令输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Error:can not obtain stdout pipe for command:%s\n", err)
		return "", false
	}
	//执行命令
	if err := cmd.Start(); err != nil {
		fmt.Println("Error:The command is err,", err)
		return "", false
	}

	cmdID := arg[2]
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			// break
			return "", false
		} else if strings.Contains(line, cmdID) {
			cmdString = line
			break
		}
	}

	// //读取所有输出
	// bytes, err := ioutil.ReadAll(stdout)
	// if err != nil {
	// 	fmt.Println("ReadAll Stdout:", err.Error())
	// 	return false
	// }
	if err := cmd.Wait(); err != nil {
		fmt.Println("wait:", err.Error())
		return "", false
	}
	// cmdString = string(bytes)
	// readarr := strings.Split(cmdString, "\n")
	// cmdString = readarr[1]

	fmt.Printf("stdout:\n %s", cmdString)
	return cmdString, true

}

func delete_extra_space(s string) string {
	//删除字符串中的多余空格，有多个空格时，仅保留一个空格
	s1 := strings.Replace(s, "	", " ", -1)       //替换tab为空格
	regstr := "\\s{2,}"                          //两个及两个以上空格的正则表达式
	reg, _ := regexp.Compile(regstr)             //编译正则表达式
	s2 := make([]byte, len(s1))                  //定义字符数组切片
	copy(s2, s1)                                 //将字符串复制到切片
	spc_index := reg.FindStringIndex(string(s2)) //在字符串中搜索
	for len(spc_index) > 0 {                     //找到适配项
		s2 = append(s2[:spc_index[0]+1], s2[spc_index[1]:]...) //删除多余空格
		spc_index = reg.FindStringIndex(string(s2))            //继续在字符串中搜索
	}
	return string(s2)
}
