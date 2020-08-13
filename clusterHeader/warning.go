package header

const (
	WARNING_LEVEL_COUNT 	uint8		= 4		// 告警级别
	WARNING_MAX_COUNT 		uint32    	= 20	// 最多记录的告警数量
)

type WarningInfo struct {
	Count			uint32			// 总告警数量
	CountPerLevel 	[]WarningCountOfType  		// 每个级别的告警数量
	CountPerNode    []WarningCountOfType		// 每个节点,每个类型的告警数量
	Warning 		[]WarningItem  	// 所有的告警条目
}

type WarningItem struct {
	Level			uint8	 		// 警告级别 0-4 0-无警告
	Time 			string			// 告警时间
	Msg				string			// 告警详细信息
	Type			string			// 告警类型：暂定 CPU 内存 硬盘 应用服务 其它
	NodeId			string			// 告警所属节点的ID
}

// 节点的告警数量
type WarningCountOfType struct {
	All				uint32			// 当前节点的总告警数量
	Cpu				uint32			// 当前节点的隶属于CPU的告警数量
	Mem				uint32			// 当前节点的隶属于内存的告警数量
	FileSystem		uint32			// 当前节点的隶属于文件系统（硬盘）的告警数量
	AppService		uint32			// 当前节点的隶属于应用服务的告警数量
	Other			uint32			// 当前节点的不属于上述4种告警的其它告警数量
}

