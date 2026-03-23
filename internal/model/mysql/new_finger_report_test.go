package mysql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDailyNewFingerReportLog(t *testing.T) {
	tests := []struct {
		data     *DailyNewFingerReportLog
		name     string
		expected bool
		msg      string
	}{
		{
			data:     &DailyNewFingerReportLog{},
			name:     "slow_query_daily_new_finger_email_log",
			expected: true,
			msg:      "same table name",
		},
		{
			data:     &DailyNewFingerReportLog{},
			name:     "slow_query_daily_new_finger_email_log_c",
			expected: false,
			msg:      "diff table name",
		},
	}

	for _, test := range tests {
		res := test.data.TableName() == test.name
		assert.Equal(t, res, test.expected, test.msg)
	}
}

func TestSaveDailyNewFingerReportLog_Save(t *testing.T) {
	tests := []struct {
		data     *DailyNewFingerReportLog
		expected bool
		msg      string
	}{
		{
			data: &DailyNewFingerReportLog{
				TaskUUID:    "",
				TaskName:    "",
				ProductLine: "",
				DBEnv:       "",
				Owners:      "",
				Leaders:     "",
				NewFinger:   0,
				NewSqlQuery: 0,
				ReportDay:   time.Now(),
				CreateTime:  time.Now(),
			},
			expected: true,
			msg:      "save err",
		},
	}
	for _, test := range tests {
		err := test.data.save(dbClient)
		assert.Nil(t, err, test.msg)
		err = SaveDailyNewFingerReportLog(dbClient, test.data)
		assert.Error(t, err, test.msg)

		_, err = FindDailyNewFingerReportLog(dbClient, "", "")
		assert.Equal(t, nil, err)
	}
}
