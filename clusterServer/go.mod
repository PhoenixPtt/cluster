module clusterServer

go 1.14

replace (
	clusterHeader => ../clusterHeader
	tcpSocket => ../tcpSocket
)

require (
	clusterHeader v0.0.0-00010101000000-000000000000
	github.com/shirou/gopsutil v2.20.6+incompatible
	tcpSocket v0.0.0-00010101000000-000000000000
)
