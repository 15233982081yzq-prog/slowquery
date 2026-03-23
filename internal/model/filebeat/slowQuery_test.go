package filebeat

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	message = "{\"@timestamp\":\"2023-06-27T10:57:18.368Z\",\"@metadata\":{\"beat\":\"filebeat\",\"type\":\"_doc\",\"version\":\"8.5.3\"},\"fields\":{\"env\":\"stage\",\"cluster_id\":\"cluster_id_xxx\",\"instance_id\":\"instance_id_xxx\",\"role\":\"master\"},\"message\":\"# User@Host: root[root] @  [127.0.0.1]  Id:     9\\n# Query_time: 0.004320  Lock_time: 0.000002 Rows_sent: 0  Rows_examined: 0\\nSET timestamp=1687863438;\\ninsert user_test values(0,\\\"ccc\\\");\"}"
)

func Test(t *testing.T) {
	var slowQuery *SlowQuery
	assert.Nil(t, slowQuery)

	err := json.Unmarshal([]byte(message), &slowQuery)
	assert.NoError(t, err)
	assert.NotNil(t, slowQuery)
	assert.Equal(t, slowQuery.TimeStamp, "2023-06-27T10:57:18.368Z")
}

func TestValid(t *testing.T) {
	tests := []struct {
		data     *SlowQuery
		expected bool
		msg      string
	}{
		{
			data: &SlowQuery{
				SlowLog: nil,
				Fields: &Field{
					ClusterUUID:  "fakeClusterUUID",
					ClusterType:  "fakeClusterType",
					InstanceHost: "fakeInstanceHost",
					InstancePort: 99,
					Env:          "fakeEnv",
				},
			},
			expected: false,
			msg:      "when SlowQuery.SlowLog is nil, valid fail",
		},
		{
			data: &SlowQuery{
				SlowLog: &MysqlSlowLog{},
				Fields: &Field{
					ClusterUUID:  "fakeClusterUUID",
					ClusterType:  "fakeClusterType",
					InstanceHost: "fakeInstanceHost",
					InstancePort: 99,
					Env:          "fakeEnv",
				},
			},
			expected: false,
			msg:      "when SlowQuery.SlowLog.TimeStamp is 0, valid fail",
		},
		{
			data: &SlowQuery{
				SlowLog: &MysqlSlowLog{
					CurrentDB: "fakeDB",
					TimeStamp: time.Now().Unix(),
					LockTime:  0,
					QueryTime: 99,
					Query:     "fake sql",
					FingerID:  "fake fingerID",
					FingerSql: "fake FingerSql",
				},
				Fields: &Field{
					ClusterUUID:  "fakeClusterUUID",
					ClusterType:  "fakeClusterType",
					InstanceHost: "fakeInstanceHost",
					InstancePort: 99,
					Env:          "fakeEnv",
				},
			},
			expected: true,
			msg:      "normal struct, valid succ",
		},
		{
			data: &SlowQuery{
				SlowLog: &MysqlSlowLog{
					CurrentDB: "fakeDB",
					TimeStamp: time.Now().Unix(),
					LockTime:  0,
					QueryTime: 99,
					Query:     "fake sql",
					FingerID:  "fake fingerID",
					FingerSql: "fake FingerSql",
				},
				Fields: &Field{},
			},
			expected: false,
			msg:      "when Field is empty, valid fail",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.data.Valid(), test.expected)
	}
}
