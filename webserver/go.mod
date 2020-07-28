module webserver

go 1.14

require (
    clusterServer v0.0.0-00010101000000-000000000000
    github.com/gin-gonic/gin v1.6.3
)

replace (
    clusterHeader => ../clusterHeader
    targz => ../targz
    tcpSocket => ../tcpSocket
	clusterServer => ../clusterServer
)
