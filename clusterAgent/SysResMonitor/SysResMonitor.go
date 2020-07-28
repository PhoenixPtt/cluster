// resMon project resMon.go
package SysResMonitor

import (
	header "clusterHeader"
	"fmt"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
)

// public
var cpuStatus header.CpuStatus
var memStatus header.StorageStatus
var diskStatus header.StorageStatus

// private
var curPid int
var curCpuUsedPercent float64
var coreCount int
var preIndex int = 0
var curIndex int = 1
var curPidPreIndex int = 0
var curPidCurIndex int = 1
var idle, total [2][]uint64
var curIdle, curTotal [2]uint64
var coreUsedPercent []float64

func init() {
	curPid = unix.Getpid()
	coreCount = runtime.NumCPU()
	coreUsedPercent = make([]float64, coreCount)

	// 获得初始资源情况
	idle[0], total[0] = getCpuUsed()
	curIdle[0], curTotal[0] = getCurCpuUsed()
}

// host info
func GetHostInfo() host.InfoStat {
	hInfo, _ := host.Info()
	return *hInfo
}

func getCurPid() int {
	return curPid
}

func GetCpuStatus() header.CpuStatus {
	// 获取CPU使用率
	curIndex = 1 - preIndex

	idle[curIndex], total[curIndex] = getCpuUsed()

	cw := len(idle[0])
	for i := 1; i < cw; i++ {
		idleTicks := float64(idle[curIndex][i] - idle[preIndex][i])
		totalTicks := float64(total[curIndex][i] - total[preIndex][i])
		coreUsedPercent[i-1] = 100 * (totalTicks - idleTicks) / totalTicks
	}
	cpuStatus.SetCoreUsedPercent(coreUsedPercent)

	for i := 0; i < cw; i++ {
		idleTicks := float64(idle[curIndex][i] - idle[preIndex][i])
		totalTicks := float64(total[curIndex][i] - total[preIndex][i])
		cpuStatus.SetUsedPercentData(100 * (totalTicks - idleTicks) / totalTicks)
		break
	}

	preIndex = curIndex

	return cpuStatus
}

func GetMemStatus() header.StorageStatus {
	memInfo, _ := mem.VirtualMemory()
	memStatus.SetUsedData(memInfo.Total-memInfo.Available, memInfo.Total)
	return memStatus
}

func GetDiskStatus() header.StorageStatus {
	parts, _ := disk.Partitions(false)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	diskStatus.SetUsedData(diskInfo.Total-diskInfo.Free, diskInfo.Total)
	return diskStatus
}

func GetCurProcessMemUsedPercent() float64 {
	contents, err := ioutil.ReadFile("/proc/" + strconv.Itoa(curPid) + "/statm")
	if err != nil {
		panic(err)
		return 0
	}

	fields := strings.Fields(string(contents))
	if len(fields) > 2 {
		val, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			fmt.Println("Error: ", fields[1], err)
		}
		memInfo, _ := mem.VirtualMemory()
		return 409600 * val / float64(memInfo.Total) // tally up all the numbers to get total ticks
	}

	return 0
}

func GetCurProcessCpuUsedPercent() float64 {
	// 获取CPU使用率
	curPidCurIndex = 1 - curPidPreIndex

	curIdle[curPidCurIndex], curTotal[curPidCurIndex] = getCurCpuUsed()

	idleTicks := float64(curIdle[curPidCurIndex] - curIdle[curPidPreIndex])
	totalTicks := float64(curTotal[curPidCurIndex] - curTotal[curPidPreIndex])
	curCpuUsedPercent = 100 * idleTicks / totalTicks

	curPidPreIndex = curPidCurIndex

	return curCpuUsedPercent
}

func getCpuUsed() (idle, total []uint64) {
	// 获得cpu和每核心的资源
	idle = make([]uint64, coreCount+1)
	total = make([]uint64, coreCount+1)
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	cw := int(coreCount + 1)
	for index := 0; index < cw; index++ {
		fields := strings.Fields(lines[index])
		numFields := len(fields)
		for i := 1; i < numFields; i++ {
			val, err := strconv.ParseUint(fields[i], 10, 64)
			if err != nil {
				fmt.Println("Error: ", i, fields[i], err)
			}
			total[index] += val // tally up all the numbers to get total ticks
			if i == 4 {         // idle is the 5th field in the cpu line
				idle[index] = val
			}
		}
	}
	return
}

func getCurCpuUsed() (idle, total uint64) {

	// 获得cpu和每核心的资源
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	fields := strings.Fields(lines[0])
	numFields := len(fields)
	for i := 1; i < numFields; i++ {
		val, _ := strconv.ParseUint(fields[i], 10, 64)
		total += val // tally up all the numbers to get total ticks
	}

	// 获得当前进程的cpu
	contents, _ = ioutil.ReadFile("/proc/" + strconv.Itoa(curPid) + "/stat")
	fields = strings.Fields(string(contents))
	if len(fields) > 16 {
		for i := 13; i <= 16; i++ {
			val, err := strconv.ParseUint(fields[i], 10, 64)
			if err != nil {
				fmt.Println("Error: ", i, fields[i], err)
			} else {
				idle += val // tally up all the numbers to get total ticks
			}
		}
	} else {
		idle = 0
	}

	return
}
