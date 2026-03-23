package request

import (
	"fmt"
	"time"
)

var (
	InitTime = int64(1696651505) //2023-10-07 12:05:05
)

type SlowQueryRank struct {
	OrderBy   string
	DBEnv     string
	Top       int
	StartTime time.Time
	EndTime   time.Time
}

type SlowQueryNewFinger struct {
	OrderBy       string
	Top           int
	DBEnv         string
	StartTime     time.Time
	EndTime       time.Time
	IsNewAppeared bool
}

func BuildSlowQueryRank(order, dbENv string, top int, start, end time.Time) *SlowQueryRank {
	return &SlowQueryRank{
		OrderBy:   order,
		DBEnv:     dbENv,
		Top:       top,
		StartTime: start,
		EndTime:   end,
	}
}

func (req *SlowQueryRank) Valid() error {
	if req.EndTime.Unix() < req.StartTime.Unix() {
		return fmt.Errorf("SlowQueryRank param failed, endTime(%v) > startTime(%v)", req.EndTime, req.StartTime)
	}
	if len(req.DBEnv) == 0 {
		return fmt.Errorf("SlowQueryRank param failed, dbEnv is empty")
	}
	if len(req.OrderBy) == 0 {
		return fmt.Errorf("SlowQueryRank param failed, order is empty")
	}
	if req.Top < 10 || req.Top > 30 {
		return fmt.Errorf("SlowQueryRank param failed, top:%d must in range[10,30]", req.Top)
	}
	return nil
}

func (req *SlowQueryRank) GetTableName() string {
	return "slow_query_log_all_rand"
}

func (req *SlowQueryNewFinger) TableName() string {
	return "slow_query_log_all_rand"
}
