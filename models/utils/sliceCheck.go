package utils

import (
	"strings"
)

func InStrings(str, strs, sep string) bool {
	strList := strings.Split(strs, sep)
	return InSliceStr(str, strList)
}

func InSliceStr(str string, slice []string) bool {
	for _, s := range slice {
		if str == s {
			return true
		}
	}
	return false
}

func InSliceInt(i int, slice []int) bool {
	for _, s := range slice {
		if i == s {
			return true
		}
	}
	return false
}

func InSliceInt64(i int64, slice []int64) bool {
	for _, s := range slice {
		if i == s {
			return true
		}
	}
	return false
}

func InSliceUint(i uint, slice []uint) bool {
	for _, s := range slice {
		if i == s {
			return true
		}
	}
	return false
}

func InSliceUint64(i uint64, slice []uint64) bool {
	for _, s := range slice {
		if i == s {
			return true
		}
	}
	return false
}

func RemoveSliceStringDuplicate(slice []string) []string {
	resSlice := []string{}
	tmpMap := make(map[string]bool)
	for _, val := range slice {
		if _, ok := tmpMap[val]; !ok {
			resSlice = append(resSlice, val)
			tmpMap[val] = true
		}
	}
	return resSlice
}
