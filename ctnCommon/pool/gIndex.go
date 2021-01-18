package pool

var(
	g_index int
)

func init()  {
	g_index = 0
}

func AddIndex()  {
	g_index++
}

func GetIndex() int {
	return g_index
}

