/*
 * Copyright 2018 Xiaomi, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package advisor

import (
	"fmt"
	"sort"
	"strings"

	"smart-slowquery/thrid-party/soar-dev/ast"
	"smart-slowquery/thrid-party/soar-dev/common"
	"smart-slowquery/thrid-party/soar-dev/database"

	"github.com/dchest/uniuri"
	"vitess.io/vitess/go/vt/sqlparser"
)

const (
	// IndexNameMaxLength Ref. https://dev.mysql.com/doc/refman/8.0/en/identifiers.html
	IndexNameMaxLength = 64
)

// IndexAdvisor 索引建议需要使用到的所有信息
type IndexAdvisor struct {
	db        string                                         // 数据库名
	Ast       sqlparser.Statement                            // Vitess Parser生成的抽象语法树
	where     []*common.Column                               // 所有where条件中用到的列
	whereEQ   []*common.Column                               // where条件中可以加索引的等值条件列
	whereINEQ []*common.Column                               // where条件中可以加索引的非等值条件列
	groupBy   []*common.Column                               // group by可以加索引列
	orderBy   []*common.Column                               // order by可以加索引列
	joinCond  [][]*common.Column                             // 由于join condition跨层级间索引不可共用，需要多一个维度用来维护层级关系
	IndexMeta map[string]map[string]*database.TableIndexInfo //db-table 每一个表有一个table信息 里面有表的列和索引信息
	TableInfo map[string]*database.TableInfo
	TraceID   string
}

// IndexInfo 创建一条索引需要的信息
type IndexInfo struct {
	Name          string           `json:"name"`           // 索引名称
	Database      string           `json:"database"`       // 数据库名
	Table         string           `json:"table"`          // 表名
	DDL           string           `json:"ddl"`            // ALTER, CREATE 等类型的 DDL 语句//TODO为啥还要这个？因为这就是建立索引的句子通过这个去重
	ColumnDetails []*common.Column `json:"column_details"` // 列详情
}

// IndexAdvises IndexAdvises列表
type IndexAdvises []IndexInfo

// mergeAdvices 合并索引建议，去重复索引建议
// dst是目标索引列表存储合并之后的 scr建议的索引还没有建立呢
func mergeAdvices(dst []IndexInfo, src ...IndexInfo) IndexAdvises {
	if len(src) == 0 {
		return dst
	}

	for _, newIdx := range src {
		has := false
		for _, idx := range dst {
			if newIdx.DDL == idx.DDL {
				common.Log.Debug("merge index %s and %s", idx.Name, newIdx.Name)
				has = true
			}
		}

		if !has {
			dst = append(dst, newIdx)
		}
	}

	return dst
}

// NewAdvisor 构造一个 IndexAdvisor 的时候就会对其本身结构初始化
// 获取 condition 中的等值条件、非等值条件，以及group by 、 order by信息
func NewAdvisor(db string, tableInfo map[string]*database.TableInfo, q Query4Audit, traceID string) (*IndexAdvisor, error) {
	common.Log.Debug("Enter: NewAdvisor(), Caller: %s, trace_id:%s", common.Caller(), traceID)

	advisor := &IndexAdvisor{
		db:  db,
		Ast: q.Stmt,
		// 所有的FindXXXXCols尽最大可能先排除不需要加索引的列，但由于元数据在此阶段尚未补齐，给出的列有可能也无法添加索引
		// 后续需要通过CompleteColumnsInfo + calcCardinality补全后再进一步判断
		joinCond:  ast.FindJoinCols(q.Stmt),
		whereEQ:   ast.FindWhereEQ(q.Stmt),
		whereINEQ: ast.FindWhereINEQ(q.Stmt),
		groupBy:   ast.FindGroupByCols(q.Stmt),
		orderBy:   ast.FindOrderByCols(q.Stmt),
		where:     ast.FindAllCols(q.Stmt, ast.WhereExpression),
		IndexMeta: make(map[string]map[string]*database.TableIndexInfo),
		TableInfo: tableInfo,
		TraceID:   traceID,
	}

	for table, info := range tableInfo {
		if advisor.IndexMeta[db] == nil {
			advisor.IndexMeta[db] = make(map[string]*database.TableIndexInfo)
		}
		advisor.IndexMeta[db][table] = info.IndexInfo
	}
	return advisor, nil
}

/*

关于如何添加索引：
在《Relational Database Index Design and the Optimizers》一书中，作者提出著名的的三星索引理论（Three-Star Index）

To Qualify for the First Star:
Pick the columns from all equal predicates (WHERE COL = . . .).
Make these the first columns of the index—in any order. For CURSOR41, the three-star index will begin with
columns LNAME, CITY or CITY, LNAME. In both cases the index slice that must be scanned will be as thin as possible.

To Qualify for the Second Star:
Add the ORDER BY columns. Do not change the order of these columns, but ignore columns that were already
picked in step 1. For example, if CURSOR41 had redundant columns in the ORDER BY, say ORDER BY LNAME,
FNAME or ORDER BY FNAME, CITY, only FNAME would be added in this step. When FNAME is the third index column,
the result table will be in the right order without sorting. The first FETCH call will return the row with
the smallest FNAME value.

To Qualify for the Third Star:
Add all the remaining columns from the SELECT statement. The order of the columns added in this step
has no impact on the performance of the SELECT, but the cost of updates should be reduced by placing volatile
columns at the end. Now the index contains all the columns required for an index-only access path.

索引添加算法正是以这个理想化索策略添为基础，尽可能的给予"三星"索引建议。

但又如《High Performance MySQL》一书中所说，索引并不总是最好的工具。只有当索引帮助存储引擎快速查找到记录带来的好处大于其
带来的额外工作时，索引才是有效的。

因此，在三星索引理论的基础上引入启发式索引算法，在第二颗星的实现上做了部分改进，对于非等值条件只会添加散粒度最高的一列到索引中，
并基于总体列的使用情况作出判断，按需对order by、group by添加索引，由此来想`增强索引建议的通用性。

*/

// IndexAdvise 索引优化建议算法入口主函数
// TODO 索引顺序该如何确定
func (idxAdv *IndexAdvisor) IndexAdvise() IndexAdvises {

	// 检查否是否含有子查询
	subQueries := ast.FindSubquery(0, idxAdv.Ast) //TODO AST到底长啥样 可以根据AST找到他的子查询？
	var subQueryAdvises []IndexInfo
	// 含有子查询对子查询进行单独评审，子查询评审建议报错忽略
	if len(subQueries) > 0 {
		for _, subSQL := range subQueries { //这里忽略了数组的索引 _代指索引
			stmt, err := sqlparser.Parse(subSQL) //根据sql生成语法树
			if err != nil {
				continue
			}
			q := Query4Audit{
				Query: subSQL,
				Stmt:  stmt,
			}
			subIdxAdv, _ := NewAdvisor(idxAdv.db, idxAdv.TableInfo, q, idxAdv.TraceID) //生成索引建议结构体
			subQueryAdvises = append(subQueryAdvises, subIdxAdv.IndexAdvise()...)      //把子查询的索引建议放进去
		}
	}

	// 变量初始化，用于存放索引信息，是一个二维的Map，db-table-column
	// 对应java的 Map<String, Map<String, List<Column>>> indexList = new HashMap<>();
	indexList := make(map[string]map[string][]*common.Column)

	// 为用到的每一列填充库名，表名等信息 补全信息，比如u其实是user表 把信息补齐
	var newJoinCond [][]*common.Column
	for _, oldJoinCols := range idxAdv.joinCond {
		newJoinCond = append(newJoinCond, CompleteColumnsInfo(idxAdv, oldJoinCols))
	}
	idxAdv.joinCond = newJoinCond
	//TODO 怎么做到补全列信息的
	idxAdv.where = CompleteColumnsInfo(idxAdv, idxAdv.where)
	idxAdv.whereEQ = CompleteColumnsInfo(idxAdv, idxAdv.whereEQ)
	idxAdv.whereINEQ = CompleteColumnsInfo(idxAdv, idxAdv.whereINEQ)
	idxAdv.groupBy = CompleteColumnsInfo(idxAdv, idxAdv.groupBy)
	idxAdv.orderBy = CompleteColumnsInfo(idxAdv, idxAdv.orderBy)

	// 是否指定Where条件，打标签
	hasWhere := false
	err := sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch where := node.(type) {
		case *sqlparser.Subquery:
			return false, nil
		case *sqlparser.Where:
			if where != nil {
				hasWhere = true
			}
		}
		return true, nil
	}, idxAdv.Ast)
	common.LogIfError(err, "trace_id:%s", idxAdv.TraceID)

	// 获取哪些列被忽略
	// 共计有三种 where（where出现的列）whereEQ（等值查询）、whereINEQ（不是等值查询）
	// 但是where 上面是暴力提取的 在where使用函数有计算这些都扔在了where中 这些应该放到ignore之中
	var ignore []*common.Column
	usedCols := append(idxAdv.whereINEQ, idxAdv.whereEQ...)

	for _, whereCol := range idxAdv.where {
		isUsed := false
		for _, used := range usedCols {
			if whereCol.Equal(used) {
				isUsed = true
			}
		}

		if !isUsed {
			common.Log.Debug("column %s in `%s`.`%s` will ignore when adding index, trace_id:%s", whereCol.DB, whereCol.Table, whereCol.Name, idxAdv.TraceID)
			ignore = append(ignore, whereCol) //添加到ignore之中
		}

	}

	// 索引优化算法入口，从这里开始放大招 这个SQL得有where才能优化
	if hasWhere {
		// 有Where条件的先分析 等值条件
		for _, index := range idxAdv.whereEQ {
			// 对应列在前面已经按散粒度由大到小排序好了
			idxAdv.mergeIndex(indexList, index) //把等值查询的列 一次循环一个列 放到了建议建立索引列表
		}
		// 若存在非等值查询条件，可以给第一个非等值条件添加索引
		if len(idxAdv.whereINEQ) > 0 {
			idxAdv.mergeIndex(indexList, idxAdv.whereINEQ[0]) //只给第一个范围查询的字段放到建议建立索引列表
		}
		// 有WHERE条件，但 WHERE 条件未能给出索引建议就不能再加 GROUP BY 和 ORDER BY 建议了
		if len(ignore) == 0 {
			// if不为0，会有在where中计算/使用函数情况，so，触发全表扫描没必要再为group by和order by加索引了
			// if为0 为了加快group by和order by 还可以给它们加索引
			for _, index := range idxAdv.groupBy {
				idxAdv.mergeIndex(indexList, index)
			}

			// OrderBy//TODO：这里为什么不可以给它们建立一个联合索引->因为你小丑了 groupby之后 会把同类型的合并 一般groupby会有一个聚合函数，比如sum（）聚合之后orderby这个 有groupby肯定不能给orderby加索引的
			// 没有 GroupBy 时可以为 OrderBy 加索引，有group by order by自己就是无序的了
			if len(idxAdv.groupBy) == 0 {
				for _, index := range idxAdv.orderBy {
					idxAdv.mergeIndex(indexList, index)
				}
			}
		}
	} else {
		// 未指定 Where 条件的，只需要 GroupBy 和 OrderBy 的索引建议
		for _, index := range idxAdv.groupBy {
			idxAdv.mergeIndex(indexList, index)
		}

		// OrderBy
		// 没有GroupBy 时可以为 OrderBy 加索引
		// 没有 where 条件时 OrderBy 的索引仅能够在索引覆盖的情况下被使用//TODO：业务考虑->我每个表都有主键索引，那么我回表差主键索引不也很快吗？
		// 实际不可以，比如Index(age)PrimaryKey（id）第一个age=1，id=19，第二个age=1，id=100，磁盘指针移动时间很多 效率不如直接全表扫描
		// 如果是覆盖索引，不需要回表，只有一个order by没有where哪些不回表 可以建立索引

		// if len(idxAdv.groupBy) == 0 {
		// 	for _, index := range idxAdv.orderBy {
		// 		mergeIndex(indexList, index)
		// 	}
		// }
	}

	// 开始整合索引信息，添加索引
	var indexes []IndexInfo

	// 根据join table的信息给予优化建议
	joinTableMeta := ast.FindJoinTable(idxAdv.Ast, nil).SetDefault(idxAdv.db)
	indexes = mergeAdvices(indexes, idxAdv.buildJoinIndex(joinTableMeta)...) //JOIN的索引建议

	indexes = mergeAdvices(indexes, idxAdv.buildIndex(indexList)...) //indexList的索引建议

	indexes = mergeAdvices(indexes, subQueryAdvises...) //子查询的索引建议

	// 在开启 env 的情况下，检查数据库版本，字段类型，索引总长度
	indexes = idxAdv.idxColsTypeCheck(indexes)

	// 在开启 env 的情况下，会对索引进行检查，对全索引进行过滤
	// DDL 语句在前面步骤已生成，这里做最终的去重、重名处理和线上对比，其次了
	return idxAdv.mergeIndexes(indexes)
}

// idxColsTypeCheck 对超长的字段添加前缀索引，剔除无法添索引字段的列
// TODO: 暂不支持 fulltext 索引，
func (idxAdv *IndexAdvisor) idxColsTypeCheck(idxList []IndexInfo) []IndexInfo {

	var indexes []IndexInfo

	for _, idx := range idxList {
		var newCols []*common.Column
		var newColInfo []string
		// 索引总长度
		idxBytesTotal := 0
		isOverFlow := false
		for _, col := range idx.ColumnDetails {
			// 获取字段 bytes
			bytes := col.GetDataBytes(idxAdv.TableInfo[col.Table].Version)
			tmpCol := col.Name
			overFlow := 0
			// 加上该列后是否索引长度过长
			if bytes < 0 {
				// bytes < 0 说明字段的长度是无法计算的
				common.Log.Warning("%s.%s data type not support %s, can't add index, trace_id:%s",
					col.Table, col.Name, col.DataType, idxAdv.TraceID)
				continue
			}

			// idx bytes overflow
			if total := idxBytesTotal + bytes; total > common.Config.MaxIdxBytes {

				common.Log.Debug("bytes: %d, idxBytesTotal: %d, total: %d, common.Config.MaxIdxBytes: %d, trace_id:%s",
					bytes, idxBytesTotal, total, common.Config.MaxIdxBytes, idxAdv.TraceID)

				overFlow = total - common.Config.MaxIdxBytes
				isOverFlow = true

			} else {
				idxBytesTotal = total
			}

			// common.Config.MaxIdxColBytes 默认大小 767
			if bytes > common.Config.MaxIdxBytesPerColumn || isOverFlow {
				// In 5.6, you may not include a column that equates to
				// bigger than 767 bytes: VARCHAR(255) CHARACTER SET utf8 or VARCHAR(191) CHARACTER SET utf8mb4.
				// In 5.7  you may not include a column that equates to
				// bigger than 3072 bytes.

				// v : 在 col.Character 字符集下每个字符占用 v bytes
				v, ok := common.CharSets[strings.ToLower(col.Character)]
				if !ok {
					// 找不到对应字符集，不添加索引
					// 如果出现不认识的字符集，认为每个字符占用4个字节
					common.Log.Warning("%s.%s(%s) charset not support yet %s, use default 4 bytes length, trace_id:%s",
						col.Table, col.Name, col.DataType, col.Character, idxAdv.TraceID)
					v = 4
				}

				// 保留两个字节的安全余量
				length := (common.Config.MaxIdxBytesPerColumn - 2) / v
				if isOverFlow {
					// 在索引中添加该列会导致索引长度过长，建议根据需求转换为合理的前缀索引
					// _OPR_SPLIT_ 是自定的用于后续处理的特殊分隔符
					common.Log.Warning("adding index '%s(%s)' to table '%s' causes the index to be too long, overflow is %d, We would not recommend indexing，trace_id:%s",
						col.Name, col.DataType, col.Table, overFlow, idxAdv.TraceID)
					tmpCol += fmt.Sprintf("_OPR_SPLIT_(N)")
				} else {
					// 索引没有过长，可以加一个最长的前缀索引
					common.Log.Warning("index column too large: %s.%s --> %s.%s(%d), data type: %s, trace_id:%s",
						col.Table, col.Name, col.Table, tmpCol, length, col.DataType, idxAdv.TraceID)
					tmpCol += fmt.Sprintf("_OPR_SPLIT_(%d)", length)
				}

			}

			newCols = append(newCols, col)
			newColInfo = append(newColInfo, tmpCol)
		}

		// 为新索引重建索引语句
		idxName := "idx_"
		idxCols := ""
		for i, newCol := range newColInfo {
			// 对名称和可能存在的长度进行拼接
			// 用等号进行分割
			tmp := strings.Split(newCol, "_OPR_SPLIT_")
			idxName += tmp[0]
			if len(tmp) > 1 {
				idxCols += tmp[0] + "`" + tmp[1]
			} else {
				idxCols += tmp[0] + "`"
			}

			if i+1 < len(newColInfo) {
				idxName += "_"
				idxCols += ",`"
			}
		}

		// 索引名称最大长度64
		if len(idxName) > IndexNameMaxLength {
			common.Log.Warn("index '%s' name large than IndexNameMaxLength, trace_id:%s", idxName, idxAdv.TraceID)
			idxName = strings.TrimRight(idxName[:IndexNameMaxLength], "_")
		}

		// 新的alter语句
		newDDL := fmt.Sprintf("alter table `%s`.`%s` add index `%s` (`%s)", idx.Database,
			idx.Table, idxName, idxCols)

		// 将筛选改造后的索引信息信息加入到新的索引列表中
		idx.ColumnDetails = newCols
		idx.DDL = newDDL
		indexes = append(indexes, idx)
	}

	return indexes
}

// mergeIndexes 与线上环境对比，将给出的索引建议进行去重
func (idxAdv *IndexAdvisor) mergeIndexes(idxList []IndexInfo) []IndexInfo {
	// TODO 暂不支持前缀索引去重
	if common.Config.TestDSN.Disable {
		return rmSelfDupIndex(idxList)
	}

	var indexes []IndexInfo
	for _, idx := range idxList {

		// 检测是否存在重复索引
		indexMeta := idxAdv.IndexMeta[idx.Database][idx.Table]
		isExisted := false

		// 检测无索引列的情况
		if len(idx.ColumnDetails) < 1 {
			continue
		}

		if existedIndexes := indexMeta.FindIndex(database.IndexColumnName, idx.ColumnDetails[0].Name); len(existedIndexes) > 0 {
			for _, existedIdx := range existedIndexes {
				// flag: 用于标记已存在的索引是否是约束条件
				isConstraint := false

				var cols []string
				var colsDetail []*common.Column

				// 把已经存在的 key 摘出来遍历一遍对比是否是包含关系
				for _, col := range indexMeta.FindIndex(database.IndexKeyName, existedIdx.KeyName) {
					cols = append(cols, col.ColumnName)
					colsDetail = append(colsDetail, &common.Column{
						Name:  col.ColumnName,
						Table: idx.Table,
						DB:    idx.ColumnDetails[0].DB,
					})
				}

				// 判断已存在的索引是否属于约束条件(唯一索引、主键)
				// 这里可以忽略是否含有外键的情况，因为索引已经重复了，添加了新索引后原先重复的索引是可以删除的。
				if existedIdx.NonUnique == 0 {
					common.Log.Debug("%s.%s表%s为约束条件, trace_id:%s", idx.Database, idx.Table, existedIdx.KeyName, idxAdv.TraceID)
					isConstraint = true
				}

				// 如果已存在的索引与索引建议存在重叠，则说明无需添加新索引或可能需要给出删除索引的建议
				if common.IsColsPart(colsDetail, idx.ColumnDetails) {
					idxName := existedIdx.KeyName
					// 如果已经存在的索引包含需要添加的索引，则无需添加
					if len(colsDetail) >= len(idx.ColumnDetails) {
						common.Log.Info(" `%s`.`%s` %s already had a index `%s`",
							idx.Database, idx.Table, strings.Join(cols, ","), idxName)
						isExisted = true
						continue
					}

					// 库、表、列名需要用反撇转义
					// TODO: 关于外键索引去重的优雅解决方案
					if !isConstraint {
						if common.Config.AllowDropIndex {
							alterSQL := fmt.Sprintf("alter table `%s`.`%s` drop index `%s`", idx.Database, idx.Table, idxName)
							indexes = append(indexes, IndexInfo{
								Name:          idxName,
								Database:      idx.Database,
								Table:         idx.Table,
								DDL:           alterSQL,
								ColumnDetails: colsDetail,
							})
						} else {
							common.Log.Warning("In table `%s`, the new index of column `%s` contains index `%s`,"+
								" maybe you could drop one of them, trace_id:%s", existedIdx.Table,
								strings.Join(cols, ","), idxName, idxAdv.TraceID)
						}
					}
				}
			}
		}

		if !isExisted {
			// 检测索引名称是否重复?
			if existedIndexes := indexMeta.FindIndex(database.IndexKeyName, idx.Name); len(existedIndexes) > 0 {
				var newName string
				idxSuffix := getRandomIndexSuffix()
				if len(idx.Name) < IndexNameMaxLength-len(idxSuffix) {
					newName = idx.Name + idxSuffix
				} else {
					newName = idx.Name[:IndexNameMaxLength-len(idxSuffix)] + idxSuffix
				}

				common.Log.Warning("duplicate index name '%s', new name is '%s', trace_id:%s", idx.Name, newName, idxAdv.TraceID)
				idx.DDL = strings.Replace(idx.DDL, idx.Name, newName, -1)
				idx.Name = newName
			}

			// 添加合并
			indexes = mergeAdvices(indexes, idx)
		}

	}

	// 对索引进行去重
	return rmSelfDupIndex(indexes)
}

// getRandomIndexSuffix format: _xxxx, length: 5
func getRandomIndexSuffix() string {
	return fmt.Sprintf("_%s", uniuri.New()[:4])
}

// rmSelfDupIndex 去重传入的[]IndexInfo中重复的索引
func rmSelfDupIndex(indexes []IndexInfo) []IndexInfo {
	var resultIndex []IndexInfo
	tmpIndexList := indexes
	for _, a := range indexes {
		tmp := a
		for i, b := range tmpIndexList {
			if common.IsColsPart(tmp.ColumnDetails, b.ColumnDetails) && tmp.Name != b.Name {
				if len(b.ColumnDetails) > len(tmp.ColumnDetails) {
					common.Log.Debug("remove duplicate index: %s", tmp.Name)
					tmp = b
				}

				if i < len(tmpIndexList) {
					tmpIndexList = append(tmpIndexList[:i], tmpIndexList[i+1:]...)
				} else {
					tmpIndexList = tmpIndexList[:i]
				}

			}
		}
		resultIndex = mergeAdvices(resultIndex, tmp)
	}

	return resultIndex
}

// buildJoinIndex 检查Join中使用的库表是否需要添加索引并给予索引建议
func (idxAdv *IndexAdvisor) buildJoinIndex(meta common.Meta) []IndexInfo {
	var indexes []IndexInfo
	for _, IndexCols := range idxAdv.joinCond {
		// 如果该列的库表为join condition中需要添加索引的库表
		indexColsList := make(map[string]map[string][]*common.Column)
		for _, col := range IndexCols {
			idxAdv.mergeIndex(indexColsList, col)
		}

		indexes = mergeAdvices(indexes, idxAdv.buildIndex(indexColsList)...)
	}
	return indexes
}

// buildIndex 尽可能的将 map[string]map[string][]*common.Column 转换成 []IndexInfo
// 此处不判断索引是否重复
func (idxAdv *IndexAdvisor) buildIndex(idxList map[string]map[string][]*common.Column) []IndexInfo {
	var indexes []IndexInfo
	for db, tbs := range idxList {
		for tb, cols := range tbs {

			// 单个索引中含有的列收 config 中参数限制
			if len(cols) > common.Config.MaxIdxColsCount {
				cols = cols[:common.Config.MaxIdxColsCount]
			}

			var colNames []string
			for _, col := range cols {
				//很多防御性编程了，原来从where加入到list的现在反过来去判断是不是列的表名和库名缺失，有可能这两个是null
				if col.DB == "" || col.Table == "" {
					common.Log.Warn("can not get the meta info of column '%s'", col.Name)
					continue
				}
				colNames = append(colNames, col.Name)
			}

			if len(colNames) == 0 {
				continue
			}
			//生成索引名字：
			idxName := common.Config.IdxPrefix + strings.Join(colNames, "_")

			// 索引名称最大长度64
			if len(idxName) > IndexNameMaxLength {
				common.Log.Warn("index '%s' name large than IndexNameMaxLength, trace_id:%s", idxName, idxAdv.TraceID)
				idxName = strings.TrimRight(idxName[:IndexNameMaxLength], "_") //太长就截尾
			}

			alterSQL := fmt.Sprintf("alter table `%s`.`%s` add index `%s` (`%s`)", db, tb,
				idxName, strings.Join(colNames, "`,`"))

			indexes = append(indexes, IndexInfo{
				Name:          idxName,
				Database:      idxAdv.db,
				Table:         tb,
				DDL:           alterSQL,
				ColumnDetails: cols,
			})
		}
	}
	return indexes
}

// buildIndexWithNoEnv 忽略原数据，给予最基础的索引
//func (idxAdv *IndexAdvisor) buildIndexWithNoEnv(indexList map[string]map[string][]*common.Column) []IndexInfo {
//	// 如果不获取数据库原信息，则不去判断索引是否重复，且只给单列加索引
//	var indexes []IndexInfo
//	for _, tableIndex := range indexList {
//		for _, indexCols := range tableIndex {
//			for _, col := range indexCols {
//				if col.Table == "" {
//					common.Log.Warn("can not get the meta info of column '%s'", col.Name)
//					continue
//				}
//				idxName := common.Config.IdxPrefix + col.Name
//				// 库、表、列名需要用反撇转义
//				alterSQL := fmt.Sprintf("alter table `%s`.`%s` add index `%s` (`%s`)", idxAdv.vEnv.RealDB(col.DB), col.Table, idxName, col.Name)
//				if col.DB == "" {
//					alterSQL = fmt.Sprintf("alter table `%s` add index `%s` (`%s`)", col.Table, idxName, col.Name)
//				}
//
//				indexes = append(indexes, IndexInfo{
//					Name:          idxName,
//					Database:      idxAdv.vEnv.RealDB(col.DB),
//					Table:         col.Table,
//					DDL:           alterSQL,
//					ColumnDetails: []*common.Column{col},
//				})
//			}
//
//		}
//	}
//	return indexes
//}

// mergeIndex 将索引用到的列去重后合并到一起
func (idxAdv *IndexAdvisor) mergeIndex(idxList map[string]map[string][]*common.Column, column *common.Column) {

	db := column.DB
	tb := column.Table
	if idxList[db] == nil {
		idxList[db] = make(map[string][]*common.Column)
	}
	if idxList[db][tb] == nil {
		idxList[db][tb] = make([]*common.Column, 0)
	} //添加列对应的数据库和表名

	// 去除重复列Append
	exist := false
	for _, cl := range idxList[db][tb] {
		if cl.Name == column.Name {
			exist = true
		}
	}

	// 将 DB 替换成 vEnv 中的数据库名称
	dbInVEnv := db

	indexMeta := idxAdv.IndexMeta[dbInVEnv][tb] //获取索引信息
	// 主键列不需要追加
	pr := indexMeta.FindIndex(database.IndexKeyName, "PRIMARY")
	for _, c := range pr {
		if c.ColumnName == column.Name {
			exist = true
		}
	}

	if !exist {
		idxList[db][tb] = append(idxList[db][tb], column)
	}
}

// CompleteColumnsInfo(idxAdv.Ast, joinCols, idxAdv.db))

// CompleteColumnsInfo 补全索引可能会用到列的所属库名、表名等信息
func CompleteColumnsInfo(indexAdvisor *IndexAdvisor, cols []*common.Column) []*common.Column {
	// 如果传过来的列是空的，没必要跑逻辑
	if len(cols) == 0 {
		return cols
	}

	// 从 Ast 中拿到 DBStructure，包含所有表的相关信息
	dbs := ast.GetMeta(indexAdvisor.Ast, nil)

	// 此处生成的 meta 信息中不应该含有""db的信息，若 DB 为空则认为是已传入的 db 为默认 db 并进行信息补全
	// BUG Fix:
	// 修补 dbs 中空 DB 的导致后续补全列信息时无法获取正确 table 名称的问题
	if _, ok := dbs[""]; ok {
		dbs[indexAdvisor.db] = dbs[""]
		delete(dbs, "")
	}

	// 判断是单表还是多表
	tableCount := 0
	for db := range dbs {
		for tb := range dbs[db].Table {
			if tb != "" {
				tableCount++
			}
		}
	}

	for _, col := range cols {
		for db := range dbs {
			// 处理有别名的列,真表名替换别名
			for _, tb := range dbs[db].Table {
				for _, tbAlias := range tb.TableAliases {
					if col.Table != "" && col.Table == tbAlias {
						common.Log.Debug("column '%s' prefix change: %s --> %s", col.Name, col.Table, tb.TableName)
						col.Table = tb.TableName
						col.DB = db
						break
					}
				}
			}
			// 判断表的数量， 单表情况，ast解析不出表名,需要手动补充
			if tableCount == 1 {
				for _, tb := range dbs[db].Table {
					col.Table = tb.TableName
				}
			}

			for _, info := range indexAdvisor.TableInfo {
				for _, c := range info.ColInfo {
					if col.Name == c.Name && col.Table == c.Table {
						col.DataType = c.DataType
						col.Table = c.Table
						col.DB = db
						col.Character = c.Character
						col.Collation = c.Collation
					}
				}
			}
		}
	}

	return cols
}

// calcCardinality 计算每一列的散粒度
// 这个函数需要在补全列的库表信息之后再调用，否则无法确定要计算列的归属
//func (idxAdv *IndexAdvisor) calcCardinality(cols []*common.Column) []*common.Column {
//	common.Log.Debug("Enter: calcCardinality(), Caller: %s", common.Caller())
//	tmpDB := *idxAdv.vEnv
//	for _, col := range cols {
//		// 补全对应列的库->表->索引信息到IndexMeta
//		// 这将在后面用于判断某一列是否为主键或单列唯一索引，快速返回散粒度
//		if col.DB == "" {
//			col.DB = idxAdv.vEnv.Database
//		}
//		realDB := idxAdv.vEnv.DBHash(col.DB)
//		if idxAdv.IndexMeta[realDB] == nil {
//			idxAdv.IndexMeta[realDB] = make(map[string]*database.TableIndexInfo)
//		}
//
//		if idxAdv.IndexMeta[realDB][col.Table] == nil {
//			tmpDB.Database = realDB
//			indexInfo, err := tmpDB.ShowIndex(col.Table)
//			if err != nil {
//				// 如果是不存在的表就会报错，报错的可能性有三个：
//				// 1.数据库错误  2.表不存在  3.临时表
//				// 而这三种错误都是不需要在这一层关注的，直接跳过
//				common.Log.Warn("calcCardinality error: %v", err)
//				continue
//			}
//
//			// 将获取的索引信息以db.tb 维度组织到 IndexMeta 中
//			idxAdv.IndexMeta[realDB][col.Table] = indexInfo
//		}
//
//		// 检查对应列是否为主键或单列唯一索引，如果满足直接返回1，不再重复计算，提高效率
//		// 多列复合唯一索引不能跳过计算，单列普通索引不能跳过计算
//		for _, index := range idxAdv.IndexMeta[realDB][col.Table].Rows {
//			// 根据索引的名称判断该索引包含的列数，列数大于1即为复合索引
//			columnCount := len(idxAdv.IndexMeta[realDB][col.Table].FindIndex(database.IndexKeyName, index.KeyName))
//			if col.Name == index.ColumnName {
//				// 主键、唯一键 无需计算散粒度
//				if (index.KeyName == "PRIMARY" || index.NonUnique == 0) && columnCount == 1 {
//					common.Log.Debug("column '%s' is PK or UK, no need to calculate cardinality.", col.Name)
//					col.Cardinality = 1
//					break
//				}
//			}
//
//		}
//
//		// 给非 PRIMARY、UNIQUE 的列计算散粒度
//		if col.Cardinality != 1 {
//			col.Cardinality = idxAdv.vEnv.ColumnCardinality(col.Table, col.Name)
//		}
//	}
//
//	return cols
//}

// Format 用于格式化输出索引建议
func (idxAdvs IndexAdvises) Format() map[string]Rule {
	rulesMap := make(map[string]Rule)
	number := 1
	rules := make(map[string]*Rule)
	sqls := make(map[string][]string)

	for _, advise := range idxAdvs {
		advKey := advise.Database + advise.Table

		if _, ok := sqls[advKey]; !ok {
			sqls[advKey] = make([]string, 0)
		}

		sqls[advKey] = append(sqls[advKey], advise.DDL)

		if _, ok := rules[advKey]; !ok {
			summary := fmt.Sprintf("Add an index to the %s table in the %s database", advise.Database, advise.Table)
			if advise.Database == "" {
				summary = fmt.Sprintf("Add index to %s table", advise.Table)
			}

			rules[advKey] = &Rule{
				Summary:  summary,
				Content:  "",
				Severity: "L2",
			}
		}

		for _, col := range advise.ColumnDetails {
			// 为了更好地显示效果
			if common.Config.Sampling {
				cardinal := fmt.Sprintf("%0.2f", col.Cardinality*100)
				if cardinal != "0.00" {
					rules[advKey].Content += fmt.Sprintf("Add an index to column %s with granularity of: %s%%; ",
						col.Name, cardinal)
				}
			} else {
				rules[advKey].Content += fmt.Sprintf("Add index to column %s;", col.Name)
			}
		}
		// 清理多余的标点
		rules[advKey].Content = strings.Trim(rules[advKey].Content, common.Config.Delimiter)
	}

	var sortAdvs []string
	for adv := range rules {
		sortAdvs = append(sortAdvs, adv)
	}
	sort.Strings(sortAdvs)

	for _, adv := range sortAdvs {
		key := fmt.Sprintf("IDX.%03d", number)
		ddl := ast.MergeAlterTables(sqls[adv]...)
		// 由于传入合并的SQL都是一张表的，所以一定只会输出一条ddl语句
		for _, v := range ddl {
			rules[adv].Case = v
		}

		if rules[adv].Case == "" {
			rules[adv].Summary = "The length of the index exceeds 3072. Index optimization is not recommended for the time being."
			rules[adv].Content = "The length of the index exceeds 3072. Index optimization is not recommended for the time being."
		}
		// set item
		rules[adv].Item = key

		rulesMap[key] = *rules[adv]

		number++
	}

	return rulesMap
}

// HeuristicCheck 依赖数据字典的启发式检查
// IndexAdvisor会基于线上表结构等数据，所以放在这里实现
func (idxAdv *IndexAdvisor) HeuristicCheck() map[string]Rule {
	var rule Rule
	heuristicSuggest := make(map[string]Rule)

	ruleFuncs := []func(*IndexAdvisor) Rule{
		(*IndexAdvisor).RuleImplicitConversion, // ARG.003
		(*IndexAdvisor).RuleGroupByConst,       // CLA.004
		(*IndexAdvisor).RuleOrderByConst,       // CLA.005
		(*IndexAdvisor).RuleUpdatePrimaryKey,   // CLA.016
	}

	for _, f := range ruleFuncs {
		rule = f(idxAdv)
		if rule.Item != "OK" {
			heuristicSuggest[rule.Item] = rule
		}
	}
	return heuristicSuggest
}

// DuplicateKeyChecker 对所有用到的库表检查是否存在重复索引
func DuplicateKeyChecker(conn *database.Connector, databases ...string) map[string]Rule {
	common.Log.Debug("Enter:  DuplicateKeyChecker, Caller: %s", common.Caller())
	// 复制一份online connector,防止环境切换影响其他功能的使用
	tmpOnline := *conn
	ruleMap := make(map[string]Rule)
	number := 1

	// 错误处理，用于汇总所有的错误
	funcErrCheck := func(err error) {
		if err != nil {
			if sug, ok := ruleMap["ERR.003"]; ok {
				sug.Content += fmt.Sprintf("; %s", err.Error())
			} else {
				ruleMap["ERR.003"] = Rule{
					Item:     "ERR.003",
					Severity: "L8",
					Content:  err.Error(),
				}
			}
		}
	}

	// 不指定 DB 的时候检查 online dsn 中的 DB
	if len(databases) == 0 {
		databases = append(databases, tmpOnline.Database)
	}

	for _, db := range databases {
		// 获取所有的表
		tmpOnline.Database = db
		tables, err := tmpOnline.ShowTables()

		if err != nil {
			funcErrCheck(err)
			if !common.Config.DryRun {
				return ruleMap
			}
		}

		for _, tb := range tables {
			// 获取表中所有的索引
			idxMap := make(map[string][]*common.Column)
			idxInfo, err := tmpOnline.ShowIndex(tb)
			if err != nil {
				funcErrCheck(err)
				if !common.Config.DryRun {
					return ruleMap
				}
			}

			// 枚举所有的索引信息，提取用到的列
			for _, idx := range idxInfo.Rows {
				if _, ok := idxMap[idx.KeyName]; !ok {
					idxMap[idx.KeyName] = make([]*common.Column, 0)
					for _, col := range idxInfo.FindIndex(database.IndexKeyName, idx.KeyName) {
						idxMap[idx.KeyName] = append(idxMap[idx.KeyName], &common.Column{
							Name:  col.ColumnName,
							Table: tb,
							DB:    db,
						})
					}
				}
			}

			// 对索引进行重复检查
			hasDup := false
			content := ""

			for k1, cl1 := range idxMap {
				for k2, cl2 := range idxMap {
					if k1 != k2 && common.IsColsPart(cl1, cl2) {
						// by pass primary key
						if k1 == "PRIMARY" || k2 == "PRIMARY" {
							continue
						}
						hasDup = true
						col1Str := common.JoinColumnsName(cl1, ", ")
						col2Str := common.JoinColumnsName(cl2, ", ")
						content += fmt.Sprintf("索引%s(%s)与%s(%s)重复;", k1, col1Str, k2, col2Str)
						common.Log.Debug(" %s.%s has duplicate index %s(%s) <--> %s(%s)", db, tb, k1, col1Str, k2, col2Str)
					}
				}
				delete(idxMap, k1)
			}

			// TODO 重复索引检查添加对约束及索引的判断，提供重复索引的删除功能
			if hasDup {
				tmpOnline.Database = db
				ddl, _ := tmpOnline.ShowCreateTable(tb)
				key := fmt.Sprintf("IDX.%03d", number)
				ruleMap[key] = Rule{
					Item:     key,
					Severity: "L2",
					Summary:  fmt.Sprintf("%s.%s存在重复的索引", db, tb),
					Content:  content,
					Case:     ddl,
				}
				number++
			}
		}
	}

	return ruleMap
}
