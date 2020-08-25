module clusterServer

go 1.14

replace (
	clusterHeader => ../clusterHeader
	ctnCommon => ../ctnCommon
	ctnServer => ../ctnServer
	tcpSocket => ../tcpSocket
)

require (
	clusterHeader v0.0.0-00010101000000-000000000000
	ctnCommon v0.0.0-00010101000000-000000000000
	ctnServer v0.0.0-00010101000000-000000000000
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/shirou/gopsutil v2.20.6+incompatible
	tcpSocket v0.0.0-00010101000000-000000000000
)
