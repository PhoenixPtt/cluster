module webservermain

go 1.14

replace (
	clusterHeader => ../clusterHeader
	clusterServer => ../clusterServer
	targz => ../targz
	tcpSocket => ../tcpSocket
	webserver => ../webserver
)

require (
	clusterHeader v0.0.0-00010101000000-000000000000
	clusterServer v0.0.0-00010101000000-000000000000
	github.com/shirou/gopsutil v2.20.6+incompatible
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae
	webserver v0.0.0-00010101000000-000000000000
)
