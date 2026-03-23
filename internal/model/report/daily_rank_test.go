package report

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDBSlowQueryRankDaily_GetRankDBNames(t *testing.T) {
	tests := []struct {
		data     *DBSlowQueryRankDaily
		expected []string
		msg      string
	}{
		{
			data: &DBSlowQueryRankDaily{
				Rank: []*DBQueryTime{
					{
						DBName: "DB_ONE",
					},
					{
						DBName: "DB_TWO",
					},
				},
			},
			expected: []string{"DB_ONE", "DB_TWO"},
			msg:      "match same result",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.data.GetRankDBNames(), test.expected)
	}

}

func TestFingerSlowQueryRankDaily_GetRankDBNames(t *testing.T) {
	tests := []struct {
		data     *FingerSlowQueryRankDaily
		expected []string
		msg      string
	}{
		{
			data: &FingerSlowQueryRankDaily{
				Rank: []*FingerQueryTime{
					{
						DBName: "DB_ONE",
					},
					{
						DBName: "DB_TWO",
					},
				},
			},
			expected: []string{"DB_ONE", "DB_TWO"},
			msg:      "match same result",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.data.GetRankDBNames(), test.expected)
	}
}

func TestListDBName(t *testing.T) {
	var tmp = NewFingerDailyReport{
		NewFingerInfos: []*NewFingerInfo{
			{
				DBName:      "1",
				NewFinger:   0,
				NewSqlQuery: 0,
			},
		},
	}
	assert.Equal(t, 1, len(tmp.GetRankDBNames()))
	assert.Equal(t, 0, tmp.NewFingerCount())
	assert.Equal(t, 0, tmp.NewSqlQueryCount())

	tmpInfo := NewFingerInfos{
		{
			DBName:      "1",
			NewFinger:   0,
			NewSqlQuery: 0,
		},
	}
	assert.Equal(t, 0, tmpInfo.NewFingerCount())
	assert.Equal(t, 0, tmpInfo.NewSqlQueryCount())
}
