package time

import (
	"fmt"
	"testing"
	"time"
)

func TestName(t *testing.T) {

	fmt.Println(HourMinuteSecond("18:39:44"))
	fmt.Println(ConvertToBaseDate(time.Now()))
}
