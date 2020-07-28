module clusterServer

go 1.14

replace (
	github.com/docker/docker v1.13.1 => github.com/docker/engine v0.0.0-20200204220554-5f6d6f3f2203
	clusterHeader => ../clusterHeader
	targz => ../targz
	tcpSocket => ../tcpSocket
)

require (
	clusterHeader v0.0.0-00010101000000-000000000000
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	targz v0.0.0-00010101000000-000000000000
	tcpSocket v0.0.0-00010101000000-000000000000
)
