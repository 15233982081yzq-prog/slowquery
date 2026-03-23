package time

import "time"

func HourMinuteSecond(s string) (t time.Time, err error) {
	check, err := time.Parse("15:04:05", s)
	if err != nil {
		return time.Time{}, err
	}
	return check, nil
}

func ConvertToBaseDate(t time.Time) time.Time {
	baseDate := time.Date(0, time.January, 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	return baseDate
}

func GetToday20AMTimestamp() int64 {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, now.Location()).Unix()
}
