package mysql

import (
	"fmt"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	_ "github.com/pingcap/tidb/types/parser_driver"
)

var (
	SelectStatement  = "select"
	InsertStatement  = "insert"
	UpdateStatement  = "update"
	DeleteStatement  = "delete"
	UnknownStatement = "unknown"
)

// 判断sql类型的函数 tidb parser解析成ast 比较复杂 这里调包 可以有时间额外学习
// ParseSqlStatement 函数作为smart-slowquery项目的基础组件，承担了SQL语句类型识别的重要职责。
// 它采用TiDB parser作为底层解析引擎，通过简洁高效的设计，为后续的慢查询分析、索引优化等模块提供了必要的基础信息。
// 该函数的设计体现了单一职责原则和性能优先的设计思路，非常适合在大规模慢查询日志处理场景中使用。
// 虽然它对DDL/DCL语句的支持有限，但这正是其设计的优势所在——通过专注于核心功能，保证了处理效率和可靠性。这些也不需要优化建议
func ParseSqlStatement(sql string) (string, error) {
	p := parser.New()

	stmtNodes, _, err := p.Parse(sql, "", "") //把sql字符串解析成结构化的语法书节点AST
	if err != nil {
		return "", fmt.Errorf("parsing sql error:%s", err.Error())
	} //解析失败
	//stmtNode, err := p.ParseOneStmt(sql, "", "")
	//if err != nil {
	//	return "", fmt.Errorf("parsing sql error:%s", err.Error())
	//}
	//判断类型
	for _, stmtNode := range stmtNodes {
		switch stmtNode.(type) {
		case *ast.SelectStmt:
			return SelectStatement, nil //主要优化类型
		case *ast.InsertStmt:
			return InsertStatement, nil //insert性能可能是磁盘io锁竞争也不优化
		case *ast.UpdateStmt:
			return UpdateStatement, nil
		case *ast.DeleteStmt:
			return DeleteStatement, nil
		default:
			return UnknownStatement, nil
		}
	}
	return UnknownStatement, nil
	//只返回日志第一条sql的类型，sql内部的这里不研究-ddl和dcl都归类为unknownStatement，dml是性能杀手
	//慢日志就是一条sql-复杂的sql后续有模块专门分析-这些不需要优化
}
