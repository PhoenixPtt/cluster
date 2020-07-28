module clusterHeader

go 1.14

replace (
		github.com/docker/docker v1.13.1 => github.com/docker/engine v0.0.0-20200204220554-5f6d6f3f2203

)
require (
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/shirou/gopsutil v2.20.6+incompatible
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae // indirect
)
