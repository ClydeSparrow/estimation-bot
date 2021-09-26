package util

import "strings"

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func GetStringInBetweenTwoString(str string, startS string, endS string) string {
	var result string = ""

	s := strings.Index(str, startS)
	if s == -1 {
		return result
	}
	newS := str[s+len(startS):]
	e := strings.LastIndex(newS, endS)
	if e == -1 {
		return result
	}
	result = newS[:e]
	return result
}

func CopyMap(m map[string]string) map[string]string {
	cp := make(map[string]string)
	for k, v := range m {
		cp[k] = v
	}

	return cp
}

func Pluralize(keyword string, count int) string {
	if count > 1 {
		if keyword == "person" {
			return "people"
		}
		return keyword + "s"
	}
	return keyword
}
