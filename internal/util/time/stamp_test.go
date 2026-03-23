package time

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStartOfTheDayStamp(t *testing.T) {
	ts := StartOfTheDayStamp(YesterdayTime())
	assert.NotEmpty(t, target, ts)
}

func TestEndOfTheDayStamp(t *testing.T) {
	ts := EndOfTheDayStamp(YesterdayTime())
	assert.NotEmpty(t, target, ts)
}

func TestTimeTsGap(t *testing.T) {
	diff := EndOfTheDayStamp(YesterdayTime()).Unix() - StartOfTheDayStamp(YesterdayTime()).Unix()
	assert.NotEmpty(t, diff)
	assert.Equal(t, diff, int64(86399))
}

func TestGetUnixTimeStamp(t *testing.T) {
	ts := GetUnixTimeStamp(-7)
	fmt.Printf("ts:%d", ts)
}
