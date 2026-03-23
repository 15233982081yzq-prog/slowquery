package time

import "time"

var (
	SecondFormat = "2006-01-02 15:04:05"
	DayFormat    = "2006-01-02"
)

func UnixTimeFormat(timestamp int64, format string) string {
	return time.Unix(timestamp, 0).Format(format)
}

func YesterdayTime() time.Time {
	return time.Now().AddDate(0, 0, -1)
}

func TodayTime() time.Time {
	return time.Now()
}
