package header

const (
	FLAG_MESG   = "MESG"	// 节点设置
)

type MESG struct {
	Oper
	SubFlag string
	Text	string
}
