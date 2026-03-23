package mysql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDailyReportLogTableName(t *testing.T) {
	tests := []struct {
		data     *DailyReportLog
		name     string
		expected bool
		msg      string
	}{
		{
			data:     &DailyReportLog{},
			name:     "slow_query_daily_rank_email_log",
			expected: true,
			msg:      "same table name",
		},
		{
			data:     &DailyReportLog{},
			name:     "slow_query_daily_rank_email_log_clone",
			expected: false,
			msg:      "diff table name",
		},
	}

	for _, test := range tests {
		res := test.data.TableName() == test.name
		assert.Equal(t, res, test.expected, test.msg)
	}
}

func TestDailyReportLog_Save(t *testing.T) {
	tests := []struct {
		data     *DailyReportLog
		expected bool
		msg      string
	}{
		{
			data: &DailyReportLog{
				TaskUUID:    "",
				TaskName:    "",
				ProductLine: "",
				DBEnv:       "",
				Owners:      "",
				Leaders:     "",
				ReportDay:   time.Now().AddDate(0, 0, -1),
				CreateTime:  time.Now().AddDate(0, 0, -1),
			},
			expected: true,
			msg:      "save err",
		},
		{
			data: &DailyReportLog{
				TaskUUID:    "",
				TaskName:    "",
				ProductLine: "",
				DBEnv:       "",
				Owners:      "",
				Leaders:     "",
				ReportDay:   time.Now().AddDate(0, 0, -2),
				CreateTime:  time.Now().AddDate(0, 0, -2),
			},
			expected: false,
			msg:      "save success",
		},
	}

	for _, test := range tests {
		err := test.data.save(dbClient)
		assert.Nil(t, err, test.msg)
		err = SaveDailyReportLog(dbClient, test.data)
		assert.Error(t, err, test.msg)

		_, err = FindDailyReportLog(dbClient, "", "")
		assert.Equal(t, nil, err)
	}
}
