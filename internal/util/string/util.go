package string

import (
	"strconv"
	"strings"
)

func IsSubset(subset, superset []string) (bl bool, diffs []string) {
	bl = true
	checkSet := make(map[string]bool, len(superset))
	for _, element := range superset {
		checkSet[element] = true
	}
	for _, value := range subset {
		if ok := checkSet[value]; ok {
			delete(checkSet, value)
		} else {
			bl = false
			diffs = append(diffs, value)
		}
	}
	return
}

func ContainInSlice(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func Split(str string, sep string) (res []string) {
	if len(str) == 0 {
		return res
	}
	return strings.Split(str, sep)
}

func IncrementString(input string, delta int) (string, error) {
	// 将字符串转换为整数
	var (
		number int
		err    error
	)
	if number, err = strconv.Atoi(input); err != nil {
		return "", err
	}

	// 对整数进行加1操作
	newNumber := number + delta

	// 将结果整数转换回字符串
	newNumberStr := strconv.Itoa(newNumber)

	return newNumberStr, nil
}
