package time

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	timestamp = int64(1690293199)
	target    = "2023-07-25 21:53:19"
)

func TestUnixTimeFormat(t *testing.T) {
	ts := UnixTimeFormat(timestamp, SecondFormat)
	assert.Equal(t, target, ts)
}

func TestYesterdayTime(t *testing.T) {
	yd := YesterdayTime()
	assert.Equal(t, UnixTimeFormat(yd.Unix(), DayFormat), UnixTimeFormat(time.Now().AddDate(0, 0, -1).Unix(), DayFormat))
}
