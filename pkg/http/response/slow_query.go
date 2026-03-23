package response

import (
	storeResp "smart-slowquery/pkg/store/response"
)

type SlowQueryListVo struct {
	TotalNum  int        `json:"total_num"`
	TotalPage int        `json:"total_page"`
	PageSize  int        `json:"page_size"`
	Querys    []*QueryVo `json:"querys"`
}

type QueryVo struct {
	FingerID    string  `json:"finger_id"`
	FingerSql   string  `json:"finger_sql"`
	ClusterUUID string  `json:"cluster_uuid"`
	DBName      string  `json:"db_name"`
	TotalTime   float64 `json:"total_time"`
	AvgTime     float64 `json:"avg_time"`
	Count       int     `json:"count"`
	AppearType  string  `json:"appear_type"`
}

func BuildSlowQueryListVo(querys *[]storeResp.SlowQuery, total, pageSize int, mmap map[string]bool) *SlowQueryListVo {
	resp := &SlowQueryListVo{
		TotalNum:  total,
		TotalPage: getTotalPage(total, pageSize),
		PageSize:  pageSize,
	}

	for _, query := range *querys {
		if mmap != nil {
			if result, ok := mmap[query.FingerID]; ok && !result {
				query.NewAppearFlag = 1
			}
		}
		queryVo := &QueryVo{
			FingerID:    query.FingerID,
			FingerSql:   query.FingerSql,
			ClusterUUID: query.ClusterUUID,
			DBName:      query.DBName,
			TotalTime:   query.TotalTime,
			AvgTime:     query.AvgTime,
			Count:       query.Count,
			AppearType:  query.GetAppearType(),
		}
		resp.Querys = append(resp.Querys, queryVo)
	}
	return resp
}

func getTotalPage(total, pageSize int) (totalPage int) {
	var delta int
	if total%pageSize > 0 {
		delta = 1
	}
	return total/pageSize + delta
}

type SlowQueryDetailVo struct {
	Info      *BaseInfoVo  `json:"slow_query_detail"`
	Statement *StatementVo `json:"sql_statement"`
}

func BuildExplainVo(result *storeResp.ExplainResult) *ExplainVo {
	var (
		explainError   string
		explainInfosVo []*ExplainInfoVo
	)
	if result != nil {
		if result.GetExplainError() != nil {
			explainError = result.GetExplainError().Error()
		} else {
			for _, explainInfo := range result.GetExplainInfos() {
				vo := &ExplainInfoVo{
					ID:           explainInfo.GetID(),
					SelectType:   explainInfo.GetSelectType(),
					Table:        explainInfo.GetTable(),
					Partitions:   explainInfo.GetPartitions(),
					Typ:          explainInfo.GetType(),
					PossibleKeys: explainInfo.GetPossibleKeys(),
					Key:          explainInfo.GetKey(),
					KeyLen:       explainInfo.GetKeyLen(),
					Ref:          explainInfo.GetRef(),
					Rows:         explainInfo.Rows.Int64,
					Filtered:     explainInfo.Filtered.Float64,
					Extra:        explainInfo.GetExtra(),
				}
				explainInfosVo = append(explainInfosVo, vo)
			}
		}
	}
	return &ExplainVo{
		ExplainInfos: explainInfosVo,
		ExplainError: explainError,
	}
}

func BuildSlowQueryDetailVo(stmt *storeResp.QueryStatement, dbName string, user string) *SlowQueryDetailVo {
	vo := &SlowQueryDetailVo{
		Info: &BaseInfoVo{
			Database:   dbName,
			ClientUser: user,
		},
		Statement: &StatementVo{
			Statement:     stmt.Statement,
			StatementType: stmt.GetStatementType(),
			NewAppeared:   stmt.NewAppearFlag > 0,
		},
	}
	if vo.Statement.NewAppeared {
		vo.Statement.NewAppearedTime = stmt.LogTime.Unix()
	}
	return vo
}

type BaseInfoVo struct {
	Database   string `json:"db_name"`
	ClientUser string `json:"client_user"`
}

type StatementVo struct {
	Statement       string  `json:"statement"`
	InstanceHost    string  `json:"instance_host,omitempty"`
	InstancePort    int32   `json:"instance_port,omitempty"`
	QueryTime       float64 `json:"query_time,omitempty"`
	LockTime        float64 `json:"lock_time,omitempty"`
	StatementType   string  `json:"statement_type"`
	NewAppeared     bool    `json:"new_appeared,omitempty"`
	NewAppearedTime int64   `json:"new_appeared_time,omitempty"`
}

type ExplainInfoVo struct {
	ID           string  `json:"id"`
	SelectType   string  `json:"select_type"`
	Table        string  `json:"table"`
	Partitions   string  `json:"partitions"`
	Typ          string  `json:"type"`
	PossibleKeys string  `json:"possible_keys"`
	Key          string  `json:"key"`
	KeyLen       string  `json:"key_len"`
	Ref          string  `json:"ref"`
	Rows         int64   `json:"rows"`
	Filtered     float64 `json:"filtered"`
	Extra        string  `json:"extra"`
}

type Rule struct {
	Item    string `json:"item"`           // 规则代号
	Summary string `json:"summary"`        // 规则摘要
	Content string `json:"content"`        // 规则解释
	Case    string `json:"case,omitempty"` // SQL示例
}

type ExplainVo struct {
	ExplainInfos []*ExplainInfoVo `json:"explain_infos"`
	ExplainError string           `json:"explain_error"`
}

type FingerClientTraceabilityVo struct {
	FingerID string               `json:"finger_id"`
	Stats    []*ClientHostStatsVo `json:"client_rank"`
}

type ClientHostStatsVo struct {
	ClientHost string `json:"client_host"`
	Count      int    `json:"count"`
}

func BuildFingerClientTraceabilityVo(fingerID string, stats *[]storeResp.ClientHostStats) *FingerClientTraceabilityVo {
	var statsVo []*ClientHostStatsVo

	for _, stat := range *stats {
		statVo := &ClientHostStatsVo{
			ClientHost: stat.ClientHost,
			Count:      stat.Count,
		}
		statsVo = append(statsVo, statVo)
	}

	return &FingerClientTraceabilityVo{
		FingerID: fingerID,
		Stats:    statsVo,
	}
}

type FingerStatementsVo struct {
	FingerID   string         `json:"finger_id"`
	TotalNum   int            `json:"total_num"`
	TotalPage  int            `json:"total_page"`
	PageSize   int            `json:"page_size"`
	Statements []*StatementVo `json:"statements"`
}

func BuildFingerStatementsVo(fingerID string, total, pageSize int, stmts *[]storeResp.QueryStatement) *FingerStatementsVo {
	var stmtsVo []*StatementVo

	for _, stmt := range *stmts {
		stmtVo := &StatementVo{
			Statement:     stmt.Statement,
			InstanceHost:  stmt.DbHost,
			InstancePort:  stmt.DbPort,
			QueryTime:     stmt.QueryTime,
			LockTime:      stmt.LockTime,
			StatementType: stmt.GetStatementType(),
		}
		stmtsVo = append(stmtsVo, stmtVo)
	}
	return &FingerStatementsVo{
		FingerID:   fingerID,
		TotalNum:   total,
		TotalPage:  getTotalPage(total, pageSize),
		PageSize:   pageSize,
		Statements: stmtsVo,
	}
}
