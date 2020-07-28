module webservermain

go 1.14

replace (
    clusterHeader => ../clusterHeader
    targz => ../targz
    tcpSocket => ../tcpSocket
	clusterServer => ../clusterServer
	webserver => ../webserver
)

require (
    clusterServer v0.0.0-00010101000000-000000000000
    webserver v0.0.0-00010101000000-000000000000
)
