package chapterfilter

import (
	"fmt"
	"github.com/yangyouwei/xiaoshuonew/conflib"
	"regexp"
)


func Makemap() map[string]int{
	var RulesMap = make(map[string]int)
	rules1 := *conflib.Chapterrules1.Rules
	for _,v := range rules1 {
		RulesMap[v] = 0
	}
	rules2 := *conflib.Chapterrules2.Rules
	for _, v := range rules2 {
		RulesMap[v] = 0
	}
	return RulesMap
}

func IfMatch(r string,s []byte)bool {
	isok , err := regexp.Match(r,s)
	if err != nil {
		fmt.Println(err)
	}
	if isok {
		return true
	}
		return false
}

//排序 返回次数最多的 rule
func RulesSort(m map[string]int) string {
	max := 0
	r := ""
	for k, v := range m {
		if v == 0 {
			continue
		}

		if v > max {
			max = v
			r = k
		}
	}
	return r
}