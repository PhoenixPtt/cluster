package ctn

import (
	"time"
)

type CTN_STATS struct {
	Read         time.Time    `json:"read"`
	Preread      time.Time    `json:"preread"`
	PidsStats    PidsStats    `json:"pids_stats"`
	BlkioStats   BlkioStats   `json:"blkio_stats"`
	NumProcs     int          `json:"num_procs"`
	StorageStats StorageStats `json:"storage_stats"`
	CpuStats     CPUStats     `json:"cpu_stats"`
	PrecpuStats  CPUStats     `json:"precpu_stats"`
	MemoryStats  MemoryStats  `json:"memory_stats"`
	Name         string       `json:"name"`
	ID           string       `json:"id"`
	//根据需要自己添加的内容
	CPUUsageCalc    float64
	PercpuUsageCalc []float64
}
type PidsStats struct {
	Current int `json:"current"`
}
type IoServiceBytesRecursive struct {
	Major int    `json:"major"`
	Minor int    `json:"minor"`
	Op    string `json:"op"`
	Value int    `json:"value"`
}
type IoServicedRecursive struct {
	Major int    `json:"major"`
	Minor int    `json:"minor"`
	Op    string `json:"op"`
	Value int    `json:"value"`
}
type BlkioStats struct {
	IoServiceBytesRecursive []IoServiceBytesRecursive `json:"io_service_bytes_recursive"`
	IoServicedRecursive     []IoServicedRecursive     `json:"io_serviced_recursive"`
	IoQueueRecursive        []interface{}             `json:"io_queue_recursive"`
	IoServiceTimeRecursive  []interface{}             `json:"io_service_time_recursive"`
	IoWaitTimeRecursive     []interface{}             `json:"io_wait_time_recursive"`
	IoMergedRecursive       []interface{}             `json:"io_merged_recursive"`
	IoTimeRecursive         []interface{}             `json:"io_time_recursive"`
	SectorsRecursive        []interface{}             `json:"sectors_recursive"`
}
type StorageStats struct {
}
type CPUUsage struct {
	TotalUsage        float64   `json:"total_usage"`
	PercpuUsage       []float64 `json:"percpu_usage"`
	UsageInKernelmode int       `json:"usage_in_kernelmode"`
	UsageInUsermode   int       `json:"usage_in_usermode"`
}
type ThrottlingData struct {
	Periods          int `json:"periods"`
	ThrottledPeriods int `json:"throttled_periods"`
	ThrottledTime    int `json:"throttled_time"`
}
type CPUStats struct {
	CPUUsage       CPUUsage       `json:"cpu_usage"`
	SystemCPUUsage float64        `json:"system_cpu_usage"`
	OnlineCpus     int            `json:"online_cpus"`
	ThrottlingData ThrottlingData `json:"throttling_data"`
}
type Stats struct {
	ActiveAnon              float64 `json:"active_anon"`
	ActiveFile              int     `json:"active_file"`
	Cache                   int     `json:"cache"`
	Dirty                   int     `json:"dirty"`
	HierarchicalMemoryLimit int64   `json:"hierarchical_memory_limit"`
	HierarchicalMemswLimit  int64   `json:"hierarchical_memsw_limit"`
	InactiveAnon            int     `json:"inactive_anon"`
	InactiveFile            int     `json:"inactive_file"`
	MappedFile              int     `json:"mapped_file"`
	Pgfault                 int     `json:"pgfault"`
	Pgmajfault              int     `json:"pgmajfault"`
	Pgpgin                  int     `json:"pgpgin"`
	Pgpgout                 int     `json:"pgpgout"`
	Rss                     int     `json:"rss"`
	RssHuge                 int     `json:"rss_huge"`
	TotalActiveAnon         int     `json:"total_active_anon"`
	TotalActiveFile         int     `json:"total_active_file"`
	TotalCache              int     `json:"total_cache"`
	TotalDirty              int     `json:"total_dirty"`
	TotalInactiveAnon       int     `json:"total_inactive_anon"`
	TotalInactiveFile       int     `json:"total_inactive_file"`
	TotalMappedFile         int     `json:"total_mapped_file"`
	TotalPgfault            int     `json:"total_pgfault"`
	TotalPgmajfault         int     `json:"total_pgmajfault"`
	TotalPgpgin             int     `json:"total_pgpgin"`
	TotalPgpgout            int     `json:"total_pgpgout"`
	TotalRss                int     `json:"total_rss"`
	TotalRssHuge            int     `json:"total_rss_huge"`
	TotalUnevictable        int     `json:"total_unevictable"`
	TotalWriteback          int     `json:"total_writeback"`
	Unevictable             int     `json:"unevictable"`
	Writeback               int     `json:"writeback"`
}
type MemoryStats struct {
	Usage    int     `json:"usage"`
	MaxUsage int     `json:"max_usage"`
	Stats    Stats   `json:"stats"`
	Limit    float64 `json:"limit"`
}
