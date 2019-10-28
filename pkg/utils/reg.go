package utils

import "regexp"

func RegexpMatchStr(regx string, data string) [][]string {
	reg := regexp.MustCompile(regx)
	pm := reg.FindAllStringSubmatch(data, -1)
	return pm
}
