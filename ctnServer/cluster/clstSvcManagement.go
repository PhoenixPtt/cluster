package cluster

//服务管理接口
type SVC_MANAGEMENT interface {
	CreateSvc(ymlFileName string, fileType string) (errType string,err error)	//创建服务
	StartSvc(sName string) (errType string,err error)							//启动服务
	ScaleSvc(sName string) (errType string,err error)							//更改服务的规模
	StopSvc(sName string) (errType string,err error)							//停止服务
	RemoveSvc(sName string) (errType string,err error)							//删除服务
	GetSvcNames() []string														//获取所有服务名称
	GetSvc(svcName string) *SERVICE												//获取指定名称的服务
}

//创建服务
func (cluster *CLUSTER) CreateSvc(fileName string, fileType string) (err error) {
	cluster.Mutex.Lock()
	defer cluster.Mutex.Unlock()

	//执行创建操作
	pSvc := NewServiceFromFile(fileName, fileType)//创建服务对象
	go pSvc.WatchRpl()
	for nodeName, status:=range cluster.NodeStatusMap{//服务更新节点信息
		pSvc.SetNodeStatus(nodeName, status)
	}

	err = cluster.check(pSvc.SvcName, SCREATE)//检查服务名的合法性
	if err!=nil{
		return err
	}
	cluster.ServiceMap[pSvc.SvcName] = pSvc//添加服务至集群

	err = pSvc.Create()
	if err!=nil{
		return err
	}

	return
}

//启动服务
func (cluster *CLUSTER) StartSvc(svcName string) (err error)  {
	err = cluster.check(svcName, SSTART)//检查服务名的合法性
	if err!=nil{
		return err
	}

	pSvc:=cluster.GetSvc(svcName)

	err = pSvc.Start()
	if err!=nil{
		return err
	}

	return
}

//调整服务规模
func (cluster *CLUSTER) ScaleSvc(svcName string, scalNum int) (err error)  {
	err = cluster.check(svcName, SSCALE)//检查服务名的合法性
	if err!=nil{
		return err
	}

	pSvc:=cluster.GetSvc(svcName)
	err = pSvc.Scale(scalNum)
	if err!=nil{
		return err
	}
	return
}

//停止服务
func (cluster *CLUSTER) StopSvc(svcName string) (err error)  {
	err = cluster.check(svcName, SSTOP)//检查服务名的合法性
	if err!=nil{
		return err
	}

	pSvc:=cluster.GetSvc(svcName)
	err = pSvc.Stop()
	if err!=nil{
		return err
	}
	return
}

//删除服务
func (cluster *CLUSTER) RemoveSvc(svcName string) (err error)  {
	err = cluster.check(svcName, SREMOVE)//检查服务名的合法性
	if err!=nil{
		return err
	}

	pSvc:=cluster.GetSvc(svcName)
	err = pSvc.Remove()
	if err!=nil{
		return err
	}

	//删除服务
	//delete(cluster.ServiceMap, svcName)//不是在此处删除
	return
}

//获取集群中所有的服务名称
func (cluster *CLUSTER) GetSvcNames() []string{
	var svcNames []string
	for key,_:=range cluster.ServiceMap{
		svcNames=append(svcNames,key)
	}
	return svcNames
}

//从集群获取指定的服务
func (cluster *CLUSTER) GetSvc(svcName string) *SERVICE {
	_, ok := cluster.ServiceMap[svcName]
	if ok {
		return cluster.ServiceMap[svcName]
	}
	return nil
}

//判断服务是否在集群中已经存在
func (cluster *CLUSTER) Contains(svcName string) bool {
	_, ok := cluster.ServiceMap[svcName]
	if ok {
		return true
	}
	return false
}

//检查服务对象
func(cluster *CLUSTER) check (svcName string, operName string) (err error) {
	//判断服务是否已经存在
	if !cluster.Contains(svcName) {
		switch operName {
		case SCREATE:
		case SSTART:
			err = errString(svcName, "不存在，无法执行启动操作。")
		case SSCALE:
			err = errString(svcName, "不存在，无法执行调整规模操作。")
		case SSTOP:
			err = errString(svcName, "不存在，无法执行停止操作。")
		case SREMOVE:
			err = errString(svcName, "不存在，无法执行删除操作。")
		default:
		}
	}

	if cluster.Contains(svcName) {
		switch operName {
		case SCREATE:
			err = errString(svcName, "已存在，无法执行创建操作。")
		case SSTART:
		case SSCALE:
		case SSTOP:
		case SREMOVE:
		default:
		}
	}
	return
}


