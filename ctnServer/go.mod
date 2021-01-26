module ctnServer

go 1.14

replace (
	clusterHeader => ../clusterHeader
	ctnCommon => ../ctnCommon
	github.com/docker/docker v1.13.1 => github.com/docker/engine v0.0.0-20200204220554-5f6d6f3f2203
	tcpSocket => ../tcpSocket
//gopkg.in/yaml.v2 => gopkg.in/yaml.v2
)

require (
	clusterHeader v0.0.0-00010101000000-000000000000
	ctnCommon v0.0.0-00010101000000-000000000000
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.2+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	tcpSocket v0.0.0-00010101000000-000000000000
)
