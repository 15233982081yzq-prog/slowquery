package clickhouse

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"smart-slowquery/pkg/store/request"
)

func TestClient_GetFingerIdWithLastSeenTimeCount(t *testing.T) {
	tests := []struct {
		parameter  *request.SlowQueryDistinctFinger
		expectedNo int64
		expected   bool
		msg        string
	}{
		{
			parameter: &request.SlowQueryDistinctFinger{
				Offset: 0,
				Limit:  10,
			},
			expectedNo: 10,
			expected:   false,
			msg:        "yes",
		},
	}

	for _, test := range tests {
		count, err := ckClient.GetFingerIdCountByTimeRange(test.parameter)
		if err != nil {
			t.Errorf("ck client GetFingerIdCountByTimeRange error:%s", err.Error())
		} else {
			assert.Equal(t, test.expectedNo, count, test.msg)
		}
	}

}

func TestClient_GetFingerIdList(t *testing.T) {
	tests := []struct {
		parameter  *request.SlowQueryDistinctFinger
		expectedNo int
		expected   bool
		msg        string
	}{
		{
			parameter: &request.SlowQueryDistinctFinger{
				Offset: 0,
				Limit:  10,
			},
			expectedNo: 0,
			expected:   false,
			msg:        "yes",
		},
	}

	for _, test := range tests {
		list, err := ckClient.GetFingerIdList(test.parameter)
		if err != nil {
			t.Errorf("ck client GetFingerIdCount error:%s", err.Error())
		} else {
			assert.Equal(t, test.expectedNo, len(list), test.msg)
		}
	}

}
