package optimize

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"smart-slowquery/pkg/http/request"
	"smart-slowquery/pkg/service/cmdb"
	"smart-slowquery/thrid-party/soar-dev/advisor"
)

var (
	param   = &request.OptimizeRequest{CurrentDB: dbName, User: dbSetting.Config.Username, Pass: dbSetting.Config.Password}
	service = &Service{}
)

func TestGetSuggestSingleTable(t *testing.T) {
	for id, exam := range testExampleSingleTable {
		param.SQL = fmt.Sprintf(exam.sql, testTable1K{}.TableName())
		param.Domains = []*cmdb.Domain{
			{
				Domain: "127.0.0.1",
				Port:   int(port),
				Role:   "master",
			},
		}
		_, indexSuggest, _, err := service.GetSuggests(param, fmt.Sprintf("test_trace_id_%d", id))
		if err != nil {
			assert.Equal(t, exam.noErr, false, err.Error())
			continue
		}
		assert.Equal(t, exam.notWithIndexSuggestCount, len(indexSuggest), param.SQL)
	}
	// 有索引情况
	for id, exam := range testExampleSingleTable {
		param.SQL = fmt.Sprintf(exam.sql, testTable1KWithIndex{}.TableName())
		param.Domains = []*cmdb.Domain{
			{
				Domain: "127.0.0.1",
				Port:   int(port),
				Role:   "master",
			},
		}
		_, indexSuggest, _, err := service.GetSuggests(param, fmt.Sprintf("test_trace_id_%d", id))
		if err != nil {
			assert.Equal(t, exam.noErr, false, err.Error())
			continue
		}
		assert.Equal(t, exam.withIndexSuggestCount, len(indexSuggest), param.SQL)
	}
}

func TestGetSuggestMultipleTables(t *testing.T) {
	// 多表情况
	for id, exam := range testExampleMultipleTables {
		if strings.Count(param.SQL, "%s") > 2 {
			param.SQL = fmt.Sprintf(exam.sql, testTable1K{}.TableName(), testTable1W{}.TableName(), testTable1K{}.TableName(), testTable1W{}.TableName())
		} else {
			param.SQL = fmt.Sprintf(exam.sql, testTable1K{}.TableName(), testTable1W{}.TableName())
		}
		param.Domains = []*cmdb.Domain{
			{
				Domain: "127.0.0.1",
				Port:   int(port),
				Role:   "master",
			},
		}
		_, indexSuggest, _, err := service.GetSuggests(param, fmt.Sprintf("test_trace_id_%d", id))
		if err != nil {
			assert.Equal(t, exam.noErr, false, err.Error())
			continue
		}
		assert.Equal(t, exam.notWithIndexSuggestCount, len(indexSuggest), param.SQL)
	}
}

func TestApi_fetchTableInfo(t *testing.T) {

	tests := []struct {
		db         string
		table      []string
		expected   bool
		user, pass string
	}{
		{
			// 正常执行
			db:       dbName,
			table:    []string{testTable1W{}.TableName()},
			expected: true,
			user:     dbSetting.Config.Username,
			pass:     dbSetting.Config.Password,
		},
		{
			// session 错误
			db:       dbName,
			table:    []string{testTable1W{}.TableName()},
			expected: false,
			user:     "err_user",
			pass:     "err_pass",
		},
		{
			//  find table info err
			db:       "err_db",
			table:    []string{testTable1W{}.TableName()},
			expected: false,
			user:     dbSetting.Config.Username,
			pass:     dbSetting.Config.Password,
		},
	}

	for i, test := range tests {
		indexes, err := fetchTableInfo(test.user, test.pass, test.db, "", test.table, []*cmdb.Domain{
			{
				Domain:     "127.0.0.1",
				Port:       int(port),
				DomainType: "",
				Role:       "master",
			},
		}, "unit_trace_id_"+strconv.Itoa(i))
		assert.Equal(t, indexes != nil, test.expected, i)
		assert.Equal(t, test.expected, err == nil, i)
	}
}

func TestService_formatSuggest(t *testing.T) {
	s, _ := NewService()
	rules := []map[string]advisor.Rule{
		{
			"SUB.001": advisor.Rule{},
		},
		{
			"ARG.005": advisor.Rule{},
		},
		{
			"001": advisor.Rule{},
		},
	}
	s.formatSuggest("", "", rules...)
}

func TestIsNotDMLORDQL(t *testing.T) {
	queryAudit, syntaxErr := advisor.NewQuery4Audit("select * from abc", "")
	assert.Equal(t, nil, syntaxErr)
	isNotDMLORDQL(queryAudit)

	queryAudit, syntaxErr = advisor.NewQuery4Audit("use local", "")
	assert.Equal(t, nil, syntaxErr)
	isNotDMLORDQL(queryAudit)

	queryAudit, syntaxErr = advisor.NewQuery4Audit("show tables", "")
	assert.Equal(t, nil, syntaxErr)
	isNotDMLORDQL(queryAudit)
}
