package notify

const template = "- Logical DB:%v \n\n" +
	"- Environment:%v \n\n" +
	"- ExplainQuery: \n\n" +
	"	- sql:%v \n\n" +
	"	- connection_id:%v \n\n" +
	"	- timeout:%v \n\n" +
	"	- kill connection:%v \n\n"

var DevNotifyMarkDownTemplate = "__%v slow query explain occur timeout __\n\n " + template

type DBANotifyParam struct {
	IsOk    bool
	LogicDB string
	Env     string
	Query   *ExecQuery
}

type ExecQuery struct {
	Sql          string
	ConnectionID int
	Timeout      bool
	KillHung     bool
}

func BuildDBANotifyParam(killed bool, logicDB, env, sql string, connID int) *DBANotifyParam {
	return &DBANotifyParam{
		IsOk:    false,
		LogicDB: logicDB,
		Env:     env,
		Query: &ExecQuery{
			Sql:          sql,
			ConnectionID: connID,
			Timeout:      true,
			KillHung:     killed,
		},
	}
}
