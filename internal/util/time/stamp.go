package time

import (
	"time"
)

func StartOfTheDayStamp(day time.Time) time.Time {
	// 给定日志的00:00:00,unix时间戳
	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Local().Location())
}

func EndOfTheDayStamp(day time.Time) time.Time {
	// 给定日志的23:59:59,unix时间戳
	return time.Date(day.Year(), day.Month(), day.Day(), 23, 59, 59, 0, day.Local().Location())
}

func GetUnixTimeStamp(diffDay int) int64 {
	// 获取当前时间
	now := time.Now()
	sevenDaysAgo := now.AddDate(0, 0, diffDay)
	return sevenDaysAgo.Unix()
}
