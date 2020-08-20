module clusterServer

go 1.14

replace (
	ctnCommon => ../ctnCommon
	//	github.com/docker/docker v1.13.1 => github.com/docker/engine v0.0.0
	github.com/docker/docker v1.13.1 => github.com/docker/engine v0.0.0-20200204220554-5f6d6f3f2203
	targz => ../targz
	tcpSocket => ../tcpSocket

)

require (
	ctnCommon v0.0.0-00010101000000-000000000000
	github.com/containerd/containerd v1.3.7 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381 // indirect
	google.golang.org/grpc v1.31.0 // indirect
	tcpSocket v0.0.0-00010101000000-000000000000
)
