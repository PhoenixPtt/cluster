package cluster

import (
	"errors"
	"fmt"
	"sort"
)

func errString (str1 string, str2 string) (err error) {
	str := fmt.Sprintf("%s%s", str1, str2)
	err = errors.New(str)
	return
}

func infoString(str1 string, str2 string) (str string) {
	str = fmt.Sprintf("%s%s", str1, str2)
	return
}

func SortMap(mp map[int]string) (newMp []int){
	newMp = make([]int, 0)
	for k, _ := range mp {
		newMp = append(newMp, k)
	}
	sort.Ints(newMp)
	return
}
