package header

import (
	"time"
)

type CTN_STATS struct {
	Read         			time.Time
	Preread      			time.Time
	PidsStats    			PidsStats
	BlkioStats   			BlkioStats
	NumProcs     			int
	StorageStats 			StorageStats
	CpuStats     			CPUStats
	PrecpuStats  			CPUStats
	MemoryStats  			MemoryStats
	Name         			string
	ID           			string
	//根据需要自己添加的内容
	CPUUsageCalc    		float64
	PercpuUsageCalc 		[]float64
}

type PidsStats struct {
	Current 				int
}

type IoServiceBytesRecursive struct {
	Major 					int
	Minor 					int
	Op    					string
	Value 					int
}

type IoServicedRecursive struct {
	Major 					int
	Minor 					int
	Op    					string
	Value 					int
}

type BlkioStats struct {
	IoServiceBytesRecursive []IoServiceBytesRecursive
	IoServicedRecursive     []IoServicedRecursive
	IoQueueRecursive        []interface{}
	IoServiceTimeRecursive  []interface{}
	IoWaitTimeRecursive     []interface{}
	IoMergedRecursive       []interface{}
	IoTimeRecursive         []interface{}
	SectorsRecursive        []interface{}
}

type StorageStats struct {
}

type CPUUsage struct {
	TotalUsage        		float64
	PercpuUsage       		[]float64
	UsageInKernelmode 		int
	UsageInUsermode   		int
}

type ThrottlingData struct {
	Periods          		int
	ThrottledPeriods 		int
	ThrottledTime    		int
}

type CPUStats struct {
	CPUUsage       			CPUUsage
	SystemCPUUsage 			float64
	OnlineCpus     			int
	ThrottlingData 			ThrottlingData
}

type Stats struct {
	ActiveAnon              float64
	ActiveFile              int
	Cache                   int
	Dirty                   int
	HierarchicalMemoryLimit int64
	HierarchicalMemswLimit  int64
	InactiveAnon            int
	InactiveFile            int
	MappedFile              int
	Pgfault                 int
	Pgmajfault              int
	Pgpgin                  int
	Pgpgout                 int
	Rss                     int
	RssHuge                 int
	TotalActiveAnon         int
	TotalActiveFile         int
	TotalCache              int
	TotalDirty              int
	TotalInactiveAnon       int
	TotalInactiveFile       int
	TotalMappedFile         int
	TotalPgfault            int
	TotalPgmajfault         int
	TotalPgpgin             int
	TotalPgpgout            int
	TotalRss                int
	TotalRssHuge            int
	TotalUnevictable        int
	TotalWriteback          int
	Unevictable             int
	Writeback               int
}

type MemoryStats struct {
	Usage    				int
	MaxUsage 				int
	Stats    				Stats
	Limit    				float64
}
