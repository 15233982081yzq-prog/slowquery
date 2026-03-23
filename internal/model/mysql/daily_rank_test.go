package mysql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDBSlowQueryDailyRankTableName(t *testing.T) {
	tests := []struct {
		data     *DBSlowQueryDailyRank
		name     string
		expected bool
		msg      string
	}{
		{
			data:     &DBSlowQueryDailyRank{},
			name:     "slow_query_db_daily_rank",
			expected: true,
			msg:      "same table name",
		},
		{
			data:     &DBSlowQueryDailyRank{},
			name:     "slow_query_db_daily_rank_clone",
			expected: false,
			msg:      "diff table name",
		},
	}

	for _, test := range tests {
		res := test.data.TableName() == test.name
		assert.Equal(t, res, test.expected, test.msg)
	}
}

func TestFingerSlowQueryDailyRankTableName(t *testing.T) {
	tests := []struct {
		data     *FingerSlowQueryDailyRank
		name     string
		expected bool
		msg      string
	}{
		{
			data:     &FingerSlowQueryDailyRank{},
			name:     "slow_query_finger_daily_rank",
			expected: true,
			msg:      "same table name",
		},
		{
			data:     &FingerSlowQueryDailyRank{},
			name:     "slow_query_finger_daily_rank_clone",
			expected: false,
			msg:      "diff table name",
		},
	}

	for _, test := range tests {
		res := test.data.TableName() == test.name
		assert.Equal(t, res, test.expected, test.msg)
	}
}

func TestDBSlowQueryDailyRank_Save(t *testing.T) {
	tests := []struct {
		data     *DBSlowQueryDailyRank
		expected bool
		msg      string
	}{
		{
			data: &DBSlowQueryDailyRank{
				SerialNo:    0,
				ClusterUUID: "uuid_0",
				DBName:      "uuid_0",
				DBEnv:       "env_0",
				RankOrder:   "",
				RankScore:   0,
				SqlCount:    0,
				WeekOnWeek:  "",
				RankDay:     time.Now().AddDate(0, 0, -1),
				CreateTime:  time.Now().AddDate(0, 0, -1),
			},
			expected: true,
			msg:      "save err",
		},
		{
			data: &DBSlowQueryDailyRank{
				SerialNo:    1,
				ClusterUUID: "",
				DBName:      "",
				DBEnv:       "",
				RankOrder:   "",
				RankScore:   0,
				SqlCount:    0,
				WeekOnWeek:  "",
				RankDay:     time.Now().AddDate(0, 0, -2),
				CreateTime:  time.Now().AddDate(0, 0, -2),
			},
			expected: false,
			msg:      "save success",
		},
	}

	for _, test := range tests {
		err := test.data.Save(dbClient)
		assert.Nil(t, err, test.msg)
		res := SaveDBDailyRank(dbClient, []*DBSlowQueryDailyRank{test.data}) == nil
		assert.Equal(t, res, test.expected, test.msg)

		_, err = FindDBDailyRank(dbClient, "query_time", time.Now().String(), "live", 10)
		assert.Equal(t, nil, err)

		FindFingerDailyRank(dbClient, "query_time", time.Now().String(), "live", 10)
		assert.Equal(t, nil, err)
	}

}

func TestFingerSlowQueryDailyRank_Save(t *testing.T) {
	tests := []struct {
		data     *FingerSlowQueryDailyRank
		expected bool
		msg      string
	}{
		{
			data: &FingerSlowQueryDailyRank{
				SerialNo:    0,
				FingerID:    "",
				FingerSql:   "",
				ClusterUUID: "",
				DBName:      "",
				DBEnv:       "",
				RankOrder:   "",
				RankScore:   0,
				SqlCount:    0,
				WeekOnWeek:  "",
				RankDay:     time.Now().AddDate(0, 0, -1),
				CreateTime:  time.Now().AddDate(0, 0, -1),
			},
			expected: true,
			msg:      "save err",
		},
		{
			data: &FingerSlowQueryDailyRank{
				SerialNo:    0,
				FingerID:    "",
				FingerSql:   "",
				ClusterUUID: "",
				DBName:      "",
				DBEnv:       "",
				RankOrder:   "",
				RankScore:   0,
				SqlCount:    0,
				WeekOnWeek:  "",
				RankDay:     time.Now().AddDate(0, 0, -2),
				CreateTime:  time.Now().AddDate(0, 0, -2),
			},
			expected: false,
			msg:      "save success",
		},
	}

	for _, test := range tests {
		err := test.data.Save(dbClient)
		assert.Nil(t, err, test.msg)
		err = SaveFingerDailyRank(dbClient, []*FingerSlowQueryDailyRank{test.data})
		assert.Error(t, err, test.msg)
	}

}
