package hint

import (
	"regexp"
	"strings"
)

const pattern = `\btrace\s+([^*/]+)`

const comment = `/\*.*?\*/`

func GetSqlTraceHint(msg string) (hint string) {
	// 编译正则表达式
	reg := regexp.MustCompile(pattern)

	// 查找注释内容
	matches := reg.FindStringSubmatch(msg)

	if len(matches) > 1 {
		// 获取第一个匹配项，即 "trace" 信息
		hint = strings.TrimSpace(matches[1])
	}
	return hint
}

func RemoveHint(sql string) string {
	// 编译正则表达式
	reg := regexp.MustCompile(comment)
	return reg.ReplaceAllString(sql, "")
}
