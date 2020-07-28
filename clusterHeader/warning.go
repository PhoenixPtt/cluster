package header

const (
	WARNING_LEVEL_COUNT 	uint8		= 4		// 告警级别
	WARNING_MAX_COUNT 		uint32    	= 20	// 最多记录的告警数量
)

type WarningInfo struct {
	Count			uint32			// 总告警数量
	CountPerLevel 	[]uint32  		// 每个级别的告警数量
	Warning 		[]WarningItem  	// 所有的告警条目
}

type WarningItem struct {
	Level			uint8	 		// 警告级别 0-4 0-无警告
	Time 			string
	Msg				string
	Type			string
}

