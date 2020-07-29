module webserver

go 1.14

require (
	clusterHeader v0.0.0-00010101000000-000000000000
	clusterServer v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.6.3
)

replace (
	clusterHeader => ../clusterHeader
	clusterServer => ../clusterServer
	targz => ../targz
	tcpSocket => ../tcpSocket
)