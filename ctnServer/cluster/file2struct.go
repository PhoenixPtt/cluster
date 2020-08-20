package cluster

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
)

func YmlFile2Struct(ymlFileName string) (svcCfg *SVC_CFG) {
	yamlFile, err := ioutil.ReadFile(ymlFileName)
	if err != nil {
		return nil
	}

	svcCfg = &SVC_CFG{}
	err = yaml.UnmarshalStrict(yamlFile, svcCfg)
	fmt.Printf("%#v", svcCfg)

	if err != nil {
		return nil
	}

	return svcCfg
}

func JsonFile2Struct(ymlFileName string)  (svcCfg *SVC_CFG)  {
	err := yaml.Unmarshal([]byte(ymlFileName), svcCfg)
	if err!=nil{
		return nil
	}
	return svcCfg
}
