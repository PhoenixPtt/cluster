package clusterServer

import "sync"

type PkgId struct {
	id    uint16
	mutex sync.Mutex
}

var packageId PkgId

func NewPkgId() uint16 {
	packageId.mutex.Lock()
	defer packageId.mutex.Unlock()
	packageId.id++
	if packageId.id < 100 {
		packageId.id = 100
	}
	return packageId.id
}

func HandleFromPkgId(pkgId uint16) (string,bool) {
	packageId.mutex.Lock()
	defer packageId.mutex.Unlock()

	h,ok := pkgIdMap[pkgId]
	return h,ok
}



var pkgIdMap map[uint16]string = make(map[uint16]string)
