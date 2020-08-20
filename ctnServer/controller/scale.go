package controller

import (
	"ctnServer/ctnS"
	"errors"
	"fmt"
	"math"
	"math/rand"
)

/*服务扩容：增加副本的数量*/
func (service *SERVICE) EScale() (agentAddrNumMap map[string]int, err error) {
	var agentAddr []string
	//计算服务需要扩容的副本数量
	dir, scaleNum:=service.getScaleNum()
	if dir<=0{
		err = errors.New("当前副本规模已达到预期规模，无需扩容。")
		return
	}

	//从agent列表中找到在线的agent进行分配
	for node,status:=range service.NodeStatusMap{
		if status == true{
			agentAddr = append(agentAddr, node)
		}
	}

	//随机选择agent
	agentAddrNumMap = make(map[string]int)
	agentAddrNumMap = RandomSelect(agentAddr, scaleNum)
	if len(agentAddrNumMap)==0{
		err = errors.New("没有找到可供容器运行的agent！")
		return
	}

	return
}

/*服务缩容：停止部分或全部副本*/
func (service *SERVICE) DScale() (rplNames []string, err error){
	//计算需要缩容的副本数量
	dir, scaleNum:=service.getScaleNum()
	if dir>=0{
		err = errors.New("服务的运行的副本数量尚未达到目标副本数量，无需缩容。")
		return
	}

	//随机选择运行中的副本
	replicas := service.getReplicas()
	rplNames = make([]string,0,100)
	for _,replica:=range replicas{
		if scaleNum>0{
			pCtnS:=ctnS.GetCtn(replica.CtnName)
			fmt.Println(pCtnS.ID,pCtnS.CtnName)

			if pCtnS!=nil{
				if pCtnS.State=="running"{
					rplNames = append(rplNames,replica.CtnName)
					scaleNum--
				}
			}
		}
	}
	return
}

/*计算需要扩容或者缩容的副本数量*/
func (service *SERVICE) getScaleNum() (int,int){
	var(
		scaleNum int		//需要扩容或缩容的数量
		activeCtnNum int	//正在运行的容器数量
	)

	replicas := service.getReplicas()
	fmt.Println("服务包含的副本数量：",len(replicas))
	for _,rpl:=range replicas{
		pCtn:=ctnS.GetCtn(rpl.CtnName)
		if pCtn==nil{
			fmt.Printf("副本%s对应的容器不存在！序号：%d\n", rpl.RplName, rpl.CtnName)
			continue
		}
		if rpl.Dirty==false{
			if pCtn.State=="running"{
				activeCtnNum++
			}
		}
	}

	switch service.SvcStats {
	case SVC_DEFAULT,SVC_CREATED:
		return 0,0
	case SVC_RUNNING:
		scaleNum = service.SvcScale-activeCtnNum
	case SVC_STOPPED,SVC_REMOVED:
		scaleNum = -activeCtnNum
	}

	fmt.Println("运行的副本数量：",activeCtnNum,"|","需要调整的副本数量：",scaleNum)
	switch  {
	case scaleNum>0:
		return 1,int(math.Abs(float64(scaleNum)))
	case scaleNum<0:
		return -1,int(math.Abs(float64(scaleNum)))
	}
	return 0,0
}

func RandomSelect(addrs []string, selectNum int) map[string]int {
	var(
		selectionMap map[string]int
	)

	aLen := len(addrs)//获得切片长度
	if aLen == 0{
		fmt.Println("没有节点可供分配。")
		return selectionMap
	}

	var selection []string
	selection=make([]string,0,selectNum)
	for i:=0; i<selectNum;i++{
		index:=rand.Intn(aLen)//生成切片序号的随机数
		selection=append(selection, addrs[index])
	}

	selectionMap=make(map[string]int)
	for _,val:=range selection{
		_,ok:=selectionMap[val]
		if !ok{
			selectionMap[val] = 1
		}else{
			selectionMap[val]++
		}
	}

	return selectionMap
}


