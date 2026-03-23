package clickhouse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"smart-slowquery/pkg/store/request"
)

// Known: 100 random pieces of data are inserted

func TestCKReader_GetQueryStatementsCount(t *testing.T) {
	tests := []struct {
		parameter  *request.SlowQueryStatementWithOrderBy
		expectedNo int
		expected   bool
		msg        string
	}{
		{
			parameter: &request.SlowQueryStatementWithOrderBy{
				SlowQueryStatement: request.SlowQueryStatement{
					FingerID:  "finger_id_0",
					DbName:    "db_0",
					DbEnv:     "env_0",
					Instances: []string{"host_0"},
				},
				OrderBy: "create_time",
			},
			expectedNo: 10,
			expected:   true,
			msg:        "normal select",
		},

		{
			parameter: &request.SlowQueryStatementWithOrderBy{
				SlowQueryStatement: request.SlowQueryStatement{
					FingerID:  "finger_id_1",
					DbName:    "db_0",
					DbEnv:     "env_1",
					Instances: []string{"host_0"},
				},
				OrderBy: "create_time",
			},
			expectedNo: 0,
			expected:   true,
			msg:        "normal select, zero data",
		},
		{
			parameter: &request.SlowQueryStatementWithOrderBy{
				SlowQueryStatement: request.SlowQueryStatement{
					FingerID:  "finger_id_1",
					DbName:    "db_0",
					DbEnv:     "",
					Instances: []string{"host_0"},
				},
				OrderBy: "create_time",
			},
			expectedNo: -1,
			expected:   false,
			msg:        "error select, DbEnv is enpty",
		},
	}

	for _, test := range tests {
		count, err := ckReadClient.GetQueryStatementsCount(test.parameter)
		assert.Equal(t, test.expectedNo, count)
		assert.Equal(t, test.expected, err == nil)
	}
}

func TestCKReader_GetNewFingerSlowQueryReportRecord(t *testing.T) {
	tests := []struct {
		parameter  *request.SlowQueryNewFinger
		expectedNo int
		expected   bool
		msg        string
	}{
		{
			parameter: &request.SlowQueryNewFinger{
				OrderBy:       "query_time",
				DBEnv:         "live",
				StartTime:     time.Now(),
				EndTime:       time.Now(),
				IsNewAppeared: false,
			},
			expectedNo: -1,
			expected:   false,
			msg:        "error select, DbEnv is enpty",
		},
	}

	for _, test := range tests {
		_, err := ckReadClient.GetNewFingerSlowQueryReportRecord(test.parameter)
		if err != nil {
			t.Errorf("ck client GetNewFingerSlowQueryReportRecord error:%s", err.Error())
		}
	}
}

func TestCKReader_GetQueryStatements(t *testing.T) {
	tests := []struct {
		parameter *request.SlowQueryStatementWithOrderBy
		expected  bool
		msg       string
	}{
		{
			parameter: &request.SlowQueryStatementWithOrderBy{
				SlowQueryStatement: request.SlowQueryStatement{
					FingerID:  "finger_id_0",
					DbName:    "db_0",
					DbEnv:     "env_0",
					Instances: []string{"host_0"},
				},
				OrderBy: "create_time",
				Limit:   1,
			},
			expected: true,
			msg:      "normal select",
		},

		{
			parameter: &request.SlowQueryStatementWithOrderBy{
				SlowQueryStatement: request.SlowQueryStatement{
					FingerID:  "finger_id_1",
					DbName:    "db_0",
					DbEnv:     "env_1",
					Instances: []string{"host_0"},
				},
				OrderBy: "create_time",
				Limit:   1,
			},
			expected: false,
			msg:      "normal select, zero data",
		},
		{
			parameter: &request.SlowQueryStatementWithOrderBy{
				SlowQueryStatement: request.SlowQueryStatement{
					FingerID:  "finger_id_1",
					DbName:    "db_0",
					DbEnv:     "",
					Instances: []string{"host_0"},
				},
				OrderBy: "create_time",
				Limit:   1,
			},
			expected: false,
			msg:      "error select, DbEnv is enpty",
		},
	}

	for _, test := range tests {
		stmts, err := ckReadClient.GetQueryStatements(test.parameter)
		assert.Equal(t, err == nil, test.expected, err == nil)
		assert.Equal(t, test.expected, stmts != nil)
	}
}

func TestCKReader_GetQueryStatementOne(t *testing.T) {
	tests := []struct {
		parameter *request.SlowQueryStatement
		expected  bool
		msg       string
	}{
		{
			parameter: &request.SlowQueryStatement{
				FingerID:   "finger_id_0",
				DbName:     "db_0",
				DbEnv:      "env_0",
				Instances:  []string{"host_0"},
				AppearType: request.BuildAppearType("all", 0),
			},
			expected: true,
			msg:      "normal select",
		},

		{
			parameter: &request.SlowQueryStatement{
				FingerID:   "finger_id_1",
				DbName:     "db_0",
				DbEnv:      "env_1",
				Instances:  []string{"host_0"},
				AppearType: request.BuildAppearType("all", 0),
			},
			expected: false,
			msg:      "normal select, zero data",
		},
		{
			parameter: &request.SlowQueryStatement{
				FingerID:   "finger_id_1",
				DbName:     "db_0",
				DbEnv:      "",
				Instances:  []string{"host_0"},
				AppearType: request.BuildAppearType("all", 0),
			},
			expected: false,
			msg:      "error select, DbEnv is enpty",
		},
	}
	for _, test := range tests {
		stmts, err := ckReadClient.GetQueryStatementOne(test.parameter)
		assert.Equal(t, test.expected, err == nil)
		assert.Equal(t, test.expected, stmts != nil)
	}
}

func TestCKReader_GetClientUsers(t *testing.T) {
	tests := []struct {
		parameter  *request.SlowQueryClientUsers
		expected   bool
		expectedNo int
		msg        string
	}{
		{
			parameter: &request.SlowQueryClientUsers{
				FingerID:  "finger_id_0",
				DbName:    "db_0",
				DbEnv:     "env_0",
				Instances: []string{"host_0"},
			},
			expected:   true,
			expectedNo: 1,
			msg:        "normal select",
		},

		{
			parameter: &request.SlowQueryClientUsers{
				FingerID:  "finger_id_1",
				DbName:    "db_0",
				DbEnv:     "env_1",
				Instances: []string{"host_0"},
			},
			expectedNo: 0,
			expected:   false,
			msg:        "normal select, zero data",
		},
		{
			parameter: &request.SlowQueryClientUsers{
				FingerID:  "finger_id_1",
				DbName:    "db_0",
				DbEnv:     "",
				Instances: []string{"host_0"},
			},
			expectedNo: 0,
			expected:   false,
			msg:        "error select, DbEnv is enpty",
		},
	}
	for _, test := range tests {
		users, err := ckReadClient.GetClientUsers(test.parameter)
		assert.Equal(t, test.expected, err == nil)
		assert.Equal(t, test.expectedNo, len(users))
	}
}

func TestCKReader_GetClientHostsStats(t *testing.T) {
	tests := []struct {
		parameter *request.SlowQueryClientHostsStats
		expected  bool
		msg       string
	}{
		{
			parameter: &request.SlowQueryClientHostsStats{
				FingerID:  "finger_id_0",
				DbName:    "db_0",
				DbEnv:     "env_0",
				Instances: []string{"host_0"},
			},
			expected: true,
			msg:      "normal select",
		},

		{
			parameter: &request.SlowQueryClientHostsStats{
				FingerID:  "finger_id_1",
				DbName:    "db_0",
				DbEnv:     "env_1",
				Instances: []string{"host_0"},
			},
			expected: true,
			msg:      "normal select, zero data, err nil and data nil",
		},
		{
			parameter: &request.SlowQueryClientHostsStats{
				FingerID:  "finger_id_1",
				DbName:    "db_0",
				DbEnv:     "",
				Instances: []string{"host_0"},
			},
			expected: false,
			msg:      "error select, DbEnv is enpty",
		},
	}
	for _, test := range tests {
		hostsStats, err := ckReadClient.GetClientHostsStats(test.parameter)
		assert.Equal(t, test.expected, err == nil)
		assert.Equal(t, test.expected, hostsStats != nil)
	}
}

func TestCKReader_GetSlowQueryList(t *testing.T) {
	tests := []struct {
		parameter *request.SlowQueryList
		expected  bool
		msg       string
	}{
		{
			parameter: &request.SlowQueryList{
				DbNames:      []string{"db_0"},
				DbEnv:        "env_0",
				Instances:    []string{"host_0"},
				OrderBy:      "total_time",
				ClusterUUIDs: []string{"uuid_0"},
				AppearType:   request.BuildAppearType("all", 0),
			},
			expected: true,
			msg:      "normal select",
		},

		{
			parameter: &request.SlowQueryList{
				DbNames:      []string{"db_0"},
				DbEnv:        "env_1",
				Instances:    []string{"host_0"},
				OrderBy:      "total_time",
				ClusterUUIDs: []string{"uuid_0"},
				AppearType:   request.BuildAppearType("all", 0),
			},
			expected: true,
			msg:      "normal select, zero data, err nil and data nil",
		},
		{
			parameter: &request.SlowQueryList{
				DbNames:      []string{"db_0"},
				DbEnv:        "",
				Instances:    []string{"host_0"},
				OrderBy:      "total_time",
				ClusterUUIDs: []string{"uuid_0"},
				AppearType:   request.BuildAppearType("all", 0),
			},
			expected: false,
			msg:      "error select, DbEnv is enpty",
		},
	}
	for _, test := range tests {
		slowQueryList, err := ckReadClient.GetSlowQueryList(test.parameter)
		assert.Equal(t, test.expected, err == nil)
		assert.Equal(t, test.expected, slowQueryList != nil)
	}
}

func TestCKReader_GetSlowQueryCount(t *testing.T) {
	tests := []struct {
		parameter  *request.SlowQueryCount
		expected   bool
		expectedNo int
		msg        string
	}{
		{
			parameter: &request.SlowQueryCount{
				DbNames:      []string{"db_0"},
				DbEnv:        "env_0",
				Instances:    []string{"host_0"},
				ClusterUUIDs: []string{"uuid_0"},
				AppearType:   request.BuildAppearType("all", 0),
			},
			expected:   true,
			expectedNo: 1,
			msg:        "normal select",
		},

		{
			parameter: &request.SlowQueryCount{
				DbNames:      []string{"db_0"},
				DbEnv:        "env_1",
				Instances:    []string{"host_0"},
				ClusterUUIDs: []string{"uuid_0"},
				AppearType:   request.BuildAppearType("all", 0),
			},
			expected:   true,
			expectedNo: 0,
			msg:        "normal select, zero data, err nil and data nil",
		},
		{
			parameter: &request.SlowQueryCount{
				DbNames:      []string{"db_0"},
				DbEnv:        "",
				Instances:    []string{"host_0"},
				ClusterUUIDs: []string{"uuid_0"},
				AppearType:   request.BuildAppearType("all", 0),
			},
			expected:   false,
			expectedNo: -1,
			msg:        "error select, DbEnv is enpty",
		},
	}
	for _, test := range tests {
		total, err := ckReadClient.GetSlowQueryCount(test.parameter)
		assert.Equal(t, test.expected, err == nil)
		assert.Equal(t, test.expectedNo, total)
	}
}

func TestCKReader_GetSlowQueryDBStatistics(t *testing.T) {
	tests := []struct {
		parameter  *request.SlowQueryDBStatistic
		expected   bool
		expectedNo int
		msg        string
	}{
		{
			parameter: &request.SlowQueryDBStatistic{
				DbNames:      []string{"db_0"},
				DbEnv:        "env_0",
				ClusterUUids: []string{"uuid_0"},
				StartTime:    time.Now().Add(-24 * time.Hour).Unix(),
				EndTime:      time.Now().Unix(),
			},
			expected:   true,
			expectedNo: 1,
			msg:        "normal select",
		},
		{
			parameter: &request.SlowQueryDBStatistic{
				DbNames:      []string{"db_0", "db_1", "db_2"},
				DbEnv:        "env_0",
				ClusterUUids: []string{"uuid_0", "uuid_1", "uuid_2"},
				StartTime:    time.Now().Add(-24 * time.Hour).Unix(),
				EndTime:      time.Now().Unix(),
			},
			expected:   true,
			expectedNo: 1,
			msg:        "normal select",
		},
		{
			parameter: &request.SlowQueryDBStatistic{
				DbNames:      []string{},
				DbEnv:        "env_0",
				ClusterUUids: []string{"uuid_0", "uuid_1", "uuid_2"},
				StartTime:    time.Now().Add(-24 * time.Hour).Unix(),
				EndTime:      time.Now().Unix(),
			},
			expected:   false,
			expectedNo: 0,
			msg:        "normal select",
		},
	}
	for _, test := range tests {
		list, err := ckReadClient.GetSlowQueryDBStatistics(test.parameter)
		assert.Equal(t, test.expected, err == nil)
		if err == nil && list != nil {
			assert.Equal(t, test.expectedNo, len(*list))
		}
	}
}

// ck 中 alert 相关单元测试
func TestCKReader_AlertRead(t *testing.T) {
	tests := []struct {
		parameter *request.AlertMessageSearch
		expected  bool
		msg       string
	}{
		{
			parameter: &request.AlertMessageSearch{
				BaseRequest:  request.BaseRequest{},
				CMDBs:        []string{"fakeA"},
				DataBaseName: "fakeDB",
				Env:          "",
				Severity:     "Severity",
				Status:       "Warn",
				TemplateName: "TemplateName",
				StartTime:    time.Now(),
				EndTime:      time.Now(),
				IsMute:       false,
				Limit:        10,
				Offset:       0,
			},
			expected: false,
			msg:      "normal select",
		},
		{
			parameter: &request.AlertMessageSearch{
				BaseRequest:  request.BaseRequest{},
				CMDBs:        []string{"fakeA"},
				DataBaseName: "fakeDB",
				Env:          "fake",
				Severity:     "Severity",
				Status:       "Warn",
				TemplateName: "TemplateName",
				StartTime:    time.Now(),
				EndTime:      time.Now(),
				IsMute:       false,
				Limit:        10,
				Offset:       0,
			},
			expected: false,
			msg:      "normal select",
		},
		{
			parameter: &request.AlertMessageSearch{
				BaseRequest:  request.BaseRequest{},
				CMDBs:        []string{"fakeA"},
				DataBaseName: "fakeDB",
				Env:          "fake",
				Severity:     "Severity",
				Status:       "Warn",
				TemplateName: "TemplateName",
				StartTime:    time.Now(),
				EndTime:      time.Now(),
				IsMute:       true,
				Limit:        10,
				Offset:       0,
			},
			expected: false,
			msg:      "normal select",
		},
	}

	for _, test := range tests {
		ckReadClient.GetAlertMessage(test.parameter)
		ckReadClient.GetAlertMessageCount(test.parameter)

		ckReadClient.GetUnMutedAndTTLStatusMessageList()

		ckReadClient.DeleteMute("")

		ckReadClient.GetAlertMessageCountBySpecCond(&request.AlertMessageCountSearch{
			BaseRequest: request.BaseRequest{},
			CMDBs:       test.parameter.CMDBs,
			Env:         test.parameter.Env,
			StartTime:   test.parameter.StartTime,
			EndTime:     test.parameter.EndTime,
			IsMute:      false,
		})

		ckReadClient.UpdateAlertMessageStatus("", "")

		ckReadClient.CreateAlertMute(&request.AlertMute{})

		ckReadClient.GetAlertMute("")

		ckReadClient.UpdateMuteStatus("", "")

		ckReadClient.BatchUpdateMuteStatusToTTL("", "", 0)

	}
}

// ck 中 operator 相关单元测试
func TestCKReader_Operator(t *testing.T) {
	tests := []struct {
		parameter *request.AlertOperatorLog
		expected  bool
		msg       string
	}{
		{
			parameter: &request.AlertOperatorLog{
				BaseRequest: request.BaseRequest{},
				Operator:    "faker",
				ActionID:    "ActionID",
				ActionType:  "ActionType",
				ActionName:  "ActionName",
				Env:         "Env",
			},
			expected: true,
			msg:      "normal select",
		},
		{
			parameter: &request.AlertOperatorLog{
				BaseRequest: request.BaseRequest{},
				Operator:    "faker",
				ActionID:    "",
				ActionType:  "ActionType",
				ActionName:  "ActionName",
				Env:         "Env",
			},
			expected: true,
			msg:      "normal select",
		},
	}

	for _, test := range tests {
		err := ckReadClient.CreateAlertOperatorLog(test.parameter)
		assert.NotEqual(t, nil, err)
	}
}

// ck 中 alert message 相关单元测试

func TestCKReader_Message(t *testing.T) {
	tests := []struct {
		parameter *request.AlertMessage
		expected  bool
		msg       string
	}{
		{
			parameter: &request.AlertMessage{
				BaseRequest:    request.BaseRequest{},
				MonitorAlertID: "",
				MonitorRuleID:  "MonitorRuleID",
				AlertRuleUUID:  "AlertRuleUUID",
				AlertStrategy:  "AlertStrategy",
				AlertRuleName:  "AlertRuleName",
				ChannelUUID:    "ChannelUUID",
				CMDB:           "CMDB",
				DataBaseName:   "DataBaseName",
				Env:            "Env",
				Status:         "Status",
				Severity:       "Severity",
				Message:        "Message",
				ACKBy:          "ACKBy",
				LabelInfo:      "LabelInfo",
				TemplateName:   "TemplateName",
				LastAlertTime:  time.Now(),
				CreateTime:     time.Now(),
				UpdateTime:     time.Now(),
			},
			expected: true,
			msg:      "normal select",
		},
		{
			parameter: &request.AlertMessage{
				BaseRequest:    request.BaseRequest{},
				MonitorAlertID: "MonitorAlertID",
				MonitorRuleID:  "MonitorRuleID",
				AlertRuleUUID:  "AlertRuleUUID",
				AlertStrategy:  "AlertStrategy",
				AlertRuleName:  "AlertRuleName",
				ChannelUUID:    "ChannelUUID",
				CMDB:           "CMDB",
				DataBaseName:   "DataBaseName",
				Env:            "Env",
				Status:         "Status",
				Severity:       "Severity",
				Message:        "Message",
				ACKBy:          "ACKBy",
				LabelInfo:      "LabelInfo",
				TemplateName:   "TemplateName",
				StartTime:      uint64(time.Now().Unix()),
				LastAlertTime:  time.Now(),
				CreateTime:     time.Now(),
				UpdateTime:     time.Now(),
			},
			expected: false,
			msg:      "normal select",
		},
	}

	for _, test := range tests {
		err := ckReadClient.CreateAlertMessage(test.parameter)
		assert.Equal(t, test.expected, err != nil)

		err = ckReadClient.UpdateAlertMessage(test.parameter)
		assert.NotEqual(t, nil, err)

		_, err = ckReadClient.GetAlertMessageByAlertID(1001)
		assert.NotEqual(t, nil, err)

		_, err = ckReadClient.GetAlertMessageByMonitorRuleIdAndStatus(test.parameter.MonitorRuleID, test.parameter.Status)
		assert.Equal(t, nil, err)

		_, err = ckReadClient.GetAlertMessageByAlertIDs([]string{test.parameter.MonitorRuleID})
		assert.Equal(t, nil, err)

	}
}

func TestCKReader_Mute(t *testing.T) {
	tests := []struct {
		parameter *request.AlertMute
		expected  bool
		msg       string
	}{
		{
			parameter: &request.AlertMute{
				BaseRequest:    request.BaseRequest{},
				MonitorMuteID:  "MonitorMuteID",
				Env:            "",
				MuteTitle:      "MuteTitle",
				RuleUUID:       "RuleUUID",
				MonitorAlertID: "MonitorAlertID",
				Status:         "Status",
				MuteFilter:     "MuteFilter",
				StartTime:      1,
				EndTime:        2,
				Creator:        "faker",
				CreateTime:     time.Now(),
				UpdateTime:     time.Now(),
			},
			expected: true,
			msg:      "normal select",
		},
		{
			parameter: &request.AlertMute{
				BaseRequest:    request.BaseRequest{},
				MonitorMuteID:  "MonitorMuteID",
				Env:            "Env",
				MuteTitle:      "MuteTitle",
				RuleUUID:       "RuleUUID",
				MonitorAlertID: "MonitorAlertID",
				Status:         "Status",
				MuteFilter:     "MuteFilter",
				StartTime:      1,
				EndTime:        2,
				Creator:        "faker",
				CreateTime:     time.Now(),
				UpdateTime:     time.Now(),
			},
			expected: true,
			msg:      "normal select",
		},
	}

	for _, test := range tests {
		ckReadClient.CreateAlertMute(test.parameter)
	}
}

func TestCKReader_Rank(t *testing.T) {
	tests := []struct {
		parameter *request.SlowQueryRank
		expected  bool
		msg       string
	}{
		{
			parameter: &request.SlowQueryRank{
				OrderBy:   "env",
				DBEnv:     "env",
				Top:       9,
				StartTime: time.Now(),
				EndTime:   time.Now(),
			},
			expected: true,
			msg:      "normal select",
		},
		{
			parameter: &request.SlowQueryRank{
				OrderBy:   "env",
				DBEnv:     "env",
				Top:       10,
				StartTime: time.Now(),
				EndTime:   time.Now(),
			},
			expected: true,
			msg:      "normal select",
		},
	}

	for _, test := range tests {
		_, err := ckReadClient.GetDBSlowQueryRank(test.parameter)
		assert.Equal(t, test.expected, err != nil)

		_, err = ckReadClient.GetFingerSlowQueryRank(test.parameter)
		assert.Equal(t, test.expected, err != nil)
	}
}
func TestCKReader_GetInstanceHosts(t *testing.T) {
	tests := []struct {
		parameter *request.SlowQueryInstanceHosts
		expected  bool
		msg       string
	}{
		{
			parameter: &request.SlowQueryInstanceHosts{
				DbName:    "DbName",
				DbEnv:     "",
				StartTime: 1,
				EndTime:   10,
			},
			expected: true,
			msg:      "normal select",
		},
		{
			parameter: &request.SlowQueryInstanceHosts{
				DbName:    "DbName",
				DbEnv:     "DbEnv",
				StartTime: 1,
				EndTime:   10,
			},
			expected: false,
			msg:      "normal select",
		},
	}

	for _, test := range tests {
		_, err := ckReadClient.GetInstanceHosts(test.parameter)
		assert.Equal(t, test.expected, err != nil)
	}
}
