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
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"smart-slowquery/thrid-party/soar-dev/ast"
	"smart-slowquery/thrid-party/soar-dev/common"

	"github.com/kr/pretty"
	"github.com/percona/go-mysql/query"
	tidb "github.com/pingcap/parser/ast"
	"vitess.io/vitess/go/vt/sqlparser"
)

// Query4Audit 待评审的SQL结构体，由原SQL和其对应的抽象语法树组成
type Query4Audit struct {
	Query  string              // 原有的查询语句
	Stmt   sqlparser.Statement // 通过Vitess解析出的抽象语法树
	TiStmt []tidb.StmtNode     // 通过TiDB解析出的抽象语法树
}

// NewQuery4Audit return a struct for Query4Audit
// sql：要分析的sql语句 traceID追踪id options：可以选择的参数
func NewQuery4Audit(sql, traceID string, options ...string) (*Query4Audit, error) {
	var err, vErr error
	var charset string
	var collation string

	if len(options) > 0 {
		charset = options[0]
	}

	if len(options) > 1 {
		collation = options[1]
	}
	//options其实是一个切片，这两个if是看数组放的什么东西，这两个是其他函数使用函数的时候必须按照这个顺序传递参数 第一个是charset第二个是collation 不传也没事
	//相对于java的重载 不同的参数列表就需要多几个函数 这个一个函数就可以适配多种情况 很简洁
	q := &Query4Audit{Query: sql}
	// vitess 语法解析不上报，以 tidb parser 为主 可以看到这里的vErr是没有return 只是记录日志打印 但是下面的TiStmt的err是返回了的
	q.Stmt, vErr = sqlparser.Parse(sql) // 真正干活的 解析
	if vErr != nil {
		common.Log.Warn("NewQuery4Audit vitess parse Error: %s, Query: %s, trace_id:%s", vErr.Error(), sql, traceID)
	}

	// TODO: charset, collation
	// tidb parser 语法解析
	q.TiStmt, err = ast.TiParse(sql, charset, collation) // 真正干活的 解析
	return q, err
}

// Rule 评审规则元数据结构
type Rule struct {
	Item     string                  `json:"Item"`     // 规则代号
	Severity string                  `json:"Severity"` // 危险等级：L[0-8], 数字越大表示级别越高
	Summary  string                  `json:"Summary"`  // 规则摘要
	Content  string                  `json:"Content"`  // 规则解释
	Case     string                  `json:"Case"`     // SQL示例
	Position int                     `json:"Position"` // 建议所处SQL字符位置，默认0表示全局建议
	Func     func(*Query4Audit) Rule `json:"-"`        // 函数名
}

/*

## Item单词缩写含义

* ALI   Alias(AS)
* ALT   Alter
* ARG   Argument
* CLA   Classic
* COL   Column
* DIS   Distinct
* ERR   Error, 特指MySQL执行返回的报错信息, ERR.000为vitess语法错误，ERR.001为执行错误，ERR.002为EXPLAIN错误
* EXP   Explain, 由explain模块给
* FUN   Function
* IDX   Optimize, 由index模块给
* JOI   Join
* KEY   Key
* KWR   Keyword
* LCK	Lock
* LIT   Literal
* PRO   Profiling, 由profiling模块给
* RES   Result
* SEC   Security
* STA   Standard
* SUB   Subquery
* TBL   TableName
* TRA   Trace, 由trace模块给

*/

// HeuristicRules 启发式规则列表
var HeuristicRules map[string]Rule

var UnableIndexScenes = map[string]struct{}{
	"FUN.001": {},
	"ARG.011": {},
}

func init() {
	InitHeuristicRules()
}

// InitHeuristicRules ...
func InitHeuristicRules() {
	HeuristicRules = map[string]Rule{
		"OK": {
			Item:     "OK",
			Severity: "L0",
			Summary:  "OK",
			Content:  `OK`,
			Case:     "OK",
			Func:     (*Query4Audit).RuleOK,
		},
		"ALI.001": {
			Item:     "ALI.001",
			Severity: "L0",
			Summary:  "It is recommended to use the AS keyword to explicitly declare an alias",
			Content:  `In column or table aliases (such as "tbl AS alias" ), explicit use of the AS keyword is more understandable than implicit aliases (such as "tbl alias" ). `,
			Case:     "select name from tbl t1 where id < 1000",
			Func:     (*Query4Audit).RuleImplicitAlias,
		},
		"ALI.002": {
			Item:     "ALI.002",
			Severity: "L8",
			Summary:  "It is not recommended to set aliases for the column wildcard character '*'",
			Content:  `Example: "SELECT tbl.* col1, col2" The above SQL sets an alias for the column wildcard. This SQL may have logical errors. You may have intended to query col1, but instead it was the last column of tbl that was renamed. `,
			Case:     "select tbl.* as c1,c2,c3 from tbl where id < 1000",
			Func:     (*Query4Audit).RuleStarAlias,
		},
		"ALI.003": {
			Item:     "ALI.003",
			Severity: "L1",
			Summary:  "Aliases should not be the same as table or column names",
			Content:  `An alias for a table or column that is the same as its real name, making it more difficult for queries to distinguish. `,
			Case:     "select name from tbl as tbl where id < 1000",
			Func:     (*Query4Audit).RuleSameAlias,
		},
		"ALT.001": {
			Item:     "ALT.001",
			Severity: "L4",
			Summary:  "Modifying the default character set of the table will not change the character set of each field in the table",
			Content:  `Many beginners will mistakenly think that ALTER TABLE tbl_name [DEFAULT] CHARACTER SET 'UTF8' will modify the character set of all fields, but in fact it will only affect subsequent newly added fields and will not change the characters of existing fields in the table. set. If you want to modify the character set of all fields in the entire table, it is recommended to use ALTER TABLE tbl_name CONVERT TO CHARACTER SET charset_name ;`,
			Case:     "ALTER TABLE tbl_name CONVERT TO CHARACTER SET charset_name;",
			Func:     (*Query4Audit).RuleAlterCharset,
		},
		// "ALT.002" : {
		// Item: "ALT.002" ,
		// Severity: "L2" ,
		// Summary: "Multiple ALTER requests for the same table are combined into one" ,
		// Content: `Each table structure change will have an impact on online services. Even if it can be adjusted through online tools, please try to reduce the number of operations by merging ALTER requests. `,
		// Case: "ALTER TABLE tbl ADD COLUMN col int, ADD INDEX idx_col (`col`);" ,
		// Func: (*Query4Audit).RuleOK, // This suggestion is given in indexAdvisor
		//},
		"ALT.003": {
			Item:     "ALT.003",
			Severity: "L0",
			Summary:  "Deletion is listed as a high-risk operation. Please check whether the business logic still has dependencies before operation . ",
			Content:  `If the business logic dependencies are not completely eliminated, data may not be written after the column is deleted, or the deleted column data cannot be queried, resulting in program exceptions. In this case, even if you roll back through the backup data, the data written by the user will be lost. `,
			Case:     "ALTER TABLE tbl DROP COLUMN col;",
			Func:     (*Query4Audit).RuleAlterDropColumn,
		},
		"ALT.004": {
			Item:     "ALT.004",
			Severity: "L0",
			Summary:  "Deleting primary keys and foreign keys is a high-risk operation. Please confirm the impact with the DBA before operating",
			Content:  `Primary keys and foreign keys are two important constraints in relational databases. Deleting existing constraints will break the existing business logic. Please confirm the impact with the business development and DBA before operating and think twice. `,
			Case:     "ALTER TABLE tbl DROP PRIMARY KEY;",
			Func:     (*Query4Audit).RuleAlterDropKey,
		},
		"ARG.001": {
			Item:     "ARG.001",
			Severity: "L4",
			Summary:  "It is not recommended to use the preceding wildcard search",
			Content:  `For example "%foo" , if the query parameter has a preceding wildcard character, the existing index cannot be used. `,
			Case:     "select c1,c2,c3 from tbl where name like '%foo'",
			Func:     (*Query4Audit).RulePrefixLike,
		},
		"ARG.002": {
			Item:     "ARG.002",
			Severity: "L1",
			Summary:  "LIKE query without wildcards",
			Content:  `A LIKE query that does not contain wildcards may have a logical error because it is logically the same as an EQUAL query. `,
			Case:     "select c1,c2,c3 from tbl where name like 'foo'",
			Func:     (*Query4Audit).RuleEqualLike,
		},
		"ARG.003": {
			Item:     "ARG.003",
			Severity: "L4",
			Summary:  "Parameter comparison contains implicit conversion and index cannot be used",
			Content:  "Implicit type conversion has the risk of not hitting the index. In the case of high concurrency and large data volume, the consequences of missing the index are very serious.",
			Case:     "SELECT * FROM sakila.film WHERE length >= '60';",
			Func:     (*Query4Audit).RuleOK, // This suggestion is given in IndexAdvisor, RuleImplicitConversion
		},
		"ARG.004": {
			Item:     "ARG.004",
			Severity: "L4",
			Summary:  "IN (NULL)/NOT IN (NULL) is never true",
			Content:  "The correct approach is col IN ('val1', 'val2', 'val3') OR col IS NULL",
			Case:     "SELECT * FROM tb WHERE col IN (NULL);",
			Func:     (*Query4Audit).RuleIn,
		},
		"ARG.005": {
			Item:     "ARG.005",
			Severity: "L1",
			Summary:  "IN should be used with caution, too many elements will cause a full table scan",
			Content:  ` For example: select id from t where num in (1,2,3) For continuous values, use BETWEEN instead of IN: select id from t where num between 1 and 3. When there are too many IN values, MySQL may also enter a full table scan, causing a sharp decline in performance. `,
			Case:     "select id from t where num in(1,2,3)",
			Func:     (*Query4Audit).RuleIn,
		},
		"ARG.006": {
			Item:     "ARG.006",
			Severity: "L1",
			Summary:  "Try to avoid NULL value judgment on fields in the WHERE clause",
			Content:  `Using IS NULL or IS NOT NULL may cause the engine to give up using the index and perform a full table scan, such as: select id from t where num is null ; you can set the default value 0 on num to ensure that there is no NULL in the num column in the table value, and then query like this: select id from t where num=0;`,
			Case:     "select id from t where num is null",
			Func:     (*Query4Audit).RuleIsNullIsNotNull,
		},
		"ARG.007": {
			Item:     "ARG.007",
			Severity: "L3",
			Summary:  "Avoid using pattern matching",
			Content:  `Performance issues are the biggest disadvantage of using pattern matching operators. Another problem with querying using LIKE or regular expression pattern matching is that it may return unexpected results. The best solution is to use special search engine technology to replace SQL, such as Apache Lucene. Another option is to save the results to reduce the overhead of repeated searches. If you must use SQL, consider using a optimize-party extension like the FULLTEXT index in MySQL. But more broadly, you don't necessarily have to use SQL to solve every problem. `,
			Case:     "select c_id,c2,c3 from tbl where c2 like 'test%'",
			Func:     (*Query4Audit).RulePatternMatchingUsage,
		},
		"ARG.008": {
			Item:     "ARG.008",
			Severity: "L1",
			Summary:  "Please try to use the IN predicate when querying index columns with OR",
			Content:  `IN-list predicates can be used for indexed retrievals, and the optimizer can sort the IN-list to match the index's sorting sequence, resulting in more efficient retrievals. Note that the IN-list must contain only constants, or the values of constants, such as external references, that persist during query block execution. `,
			Case:     "SELECT c1,c2,c3 FROM tbl WHERE c1 = 14 OR c1 = 17",
			Func:     (*Query4Audit).RuleORUsage,
		},
		"ARG.009": {
			Item:     "ARG.009",
			Severity: "L1",
			Summary:  "The quoted string contains spaces at the beginning or end",
			Content:  `If there are spaces before and after a VARCHAR column, it may cause logical problems. For example, in MySQL 5.5, 'a' and 'a ' may be considered the same value in the query. `,
			Case:     "SELECT 'abc '",
			Func:     (*Query4Audit).RuleSpaceWithQuote,
		},
		"ARG.010": {
			Item:     "ARG.010",
			Severity: "L1",
			Summary:  "Don't use hints, such as: sql_no_cache, force index, ignore key, straight join, etc.",
			Content:  `hint is used to force SQL to execute according to a certain execution plan, but as the amount of data changes, we cannot guarantee that our original prediction is correct. `,
			Case:     "SELECT * FROM t1 USE INDEX (i1) ORDER BY a;",
			Func:     (*Query4Audit).RuleHint,
		},
		"ARG.011": {
			Item:     "ARG.011",
			Severity: "L3",
			Summary:  "Don't use negative queries, such as: NOT IN/NOT LIKE",
			Content:  `Please try not to use negative queries, which will cause a full table scan and have a greater impact on query performance. `,
			Case:     "select id from t where num not in(1,2,3);",
			Func:     (*Query4Audit).RuleNot,
		},
		"ARG.012": {
			Item:     "ARG.012",
			Severity: "L2",
			Summary:  "Too much data for one-time INSERT/REPLACE",
			Content:  "The performance of batch inserting a large amount of data with a single INSERT/REPLACE statement is poor, and may even cause synchronization delays from the slave database. In order to improve performance and reduce the impact of batch writing data on synchronization delays from the slave database, it is recommended to use the batch insertion method .",
			Case:     "INSERT INTO tb (a) VALUES (1), (2)",
			Func:     (*Query4Audit).RuleInsertValues,
		},
		"ARG.013": {
			Item:     "ARG.013",
			Severity: "L0",
			Summary:  "Chinese full-width quotation marks are used in the DDL statement",
			Content:  `Chinese full-width quotation marks "" or '' are used in the DDL statement. This may be a writing error. Please confirm whether it is as expected.`,
			Case:     `CREATE TABLE tb (a varchar(10) default '""'`,
			Func:     (*Query4Audit).RuleFullWidthQuote,
		},
		"ARG.014": {
			Item:     "ARG.014",
			Severity: "L4",
			Summary:  "The existence of column names in the IN condition may lead to an expansion of the data matching range",
			Content:  `For example: delete from t where id in (1, 2, id) may cause the entire table data to be accidentally deleted. Please double check the correctness of the IN conditions. `,
			Case:     "select id from t where id in(1, 2, id)",
			Func:     (*Query4Audit).RuleIn,
		},
		"CLA.001": {
			Item:     "CLA.001",
			Severity: "L4",
			Summary:  "The outermost SELECT does not specify a WHERE condition",
			Content:  `SELECT statement does not have a WHERE clause and may check more rows than expected (full table scan). For SELECT COUNT(*) type requests, if precision is not required, it is recommended to use SHOW TABLE STATUS or EXPLAIN instead. `,
			Case:     "select id from tbl",
			Func:     (*Query4Audit).RuleNoWhere,
		},
		"CLA.002": {
			Item:     "CLA.002",
			Severity: "L3",
			Summary:  "ORDER BY RAND() is not recommended",
			Content:  `ORDER BY RAND() is a very inefficient way to retrieve random rows from a result set because it sorts the entire result and discards most of its data. `,
			Case:     "select name from tbl where id < 1000 order by rand(number)",
			Func:     (*Query4Audit).RuleOrderByRand,
		},
		"CLA.003": {
			Item:     "CLA.003",
			Severity: "L2",
			Summary:  "It is not recommended to use LIMIT query with OFFSET",
			Content:  `The complexity of paging the result set using LIMIT and OFFSET is O(n^2) and will cause performance issues as the data grows. Use the "bookmark" scanning method to achieve higher paging efficiency. `,
			Case:     "select c1,c2 from tbl where name=xx order by number limit 1 offset 20",
			Func:     (*Query4Audit).RuleOffsetLimit,
		},
		"CLA.004": {
			Item:     "CLA.004",
			Severity: "L2",
			Summary:  "GROUP BY is not recommended for constants",
			Content:  `GROUP BY 1 means GROUP BY based on the first column. If you use numbers in the GROUP BY clause instead of expressions or column names, it can cause problems when the query column order changes. `,
			Case:     "select col1,col2 from tbl group by 1",
			Func:     (*Query4Audit).RuleGroupByConst,
		},
		"CLA.005": {
			Item:     "CLA.005",
			Severity: "L2",
			Summary:  "ORDER BY constant column has no meaning",
			Content:  `There may be errors in the SQL logic ; at most it is a useless operation that does not change the query results. `,
			Case:     "select id from test where id=1 order by id",
			Func:     (*Query4Audit).RuleOrderByConst,
		},
		"CLA.006": {
			Item:     "CLA.006",
			Severity: "L4",
			Summary:  "GROUP BY or ORDER BY in different tables",
			Content:  `This will force the use of temporary tables and filesort, which may have huge performance implications, and may consume a lot of memory and temporary space on disk. `,
			Case:     "select tb1.col, tb2.col from tb1, tb2 where id=1 group by tb1.col, tb2.col",
			Func:     (*Query4Audit).RuleDiffGroupByOrderBy,
		},
		"CLA.008": {
			Item:     "CLA.008",
			Severity: "L2",
			Summary:  "Please add ORDER BY conditions for GROUP BY display",
			Content:  `By default MySQL will sort 'GROUP BY col1, col2, ...' requests in the following order 'ORDER BY col1, col2, ...' . If the GROUP BY statement does not specify the ORDER BY condition, it will cause unnecessary sorting. If sorting is not required, it is recommended to add 'ORDER BY NULL' . `,
			Case:     "select c1,c2,c3 from t1 where c1='foo' group by c2",
			Func:     (*Query4Audit).RuleExplicitOrderBy,
		},
		"CLA.009": {
			Item:     "CLA.009",
			Severity: "L2",
			Summary:  "The condition of ORDER BY is an expression",
			Content:  `Temporary tables are used when the ORDER BY condition is an expression or function. If WHERE is not specified or the result set returned by the WHERE condition is large, the performance will be poor. `,
			Case:     "select description from film where title ='ACADEMY DINOSAUR' order by length-language_id;",
			Func:     (*Query4Audit).RuleOrderByExpr,
		},
		"CLA.010": {
			Item:     "CLA.010",
			Severity: "L2",
			Summary:  "The condition of GROUP BY is an expression",
			Content:  `Temporary tables are used when the GROUP BY condition is an expression or function. If WHERE is not specified or the result set returned by the WHERE condition is large, the performance will be poor. `,
			Case:     "select description from film where title ='ACADEMY DINOSAUR' GROUP BY length-language_id;",
			Func:     (*Query4Audit).RuleGroupByExpr,
		},
		"CLA.011": {
			Item:     "CLA.011",
			Severity: "L1",
			Summary:  "It is recommended to add comments to the table",
			Content:  `Adding comments to the table can make the meaning of the table clearer, thus bringing great convenience to future maintenance. `,
			Case:     "CREATE TABLE `test1` (`ID` bigint(20) NOT NULL AUTO_INCREMENT,`c1` varchar(128) DEFAULT NULL,PRIMARY KEY (`ID`)) ENGINE=InnoDB DEFAULT CHARSET=utf8",
			Func:     (*Query4Audit).RuleTblCommentCheck,
		},
		"CLA.012": {
			Item:     "CLA.012",
			Severity: "L2",
			Summary:  "Broken down a complex footwear query into several simpler queries",
			Content:  `SQL is a very expressive language and you can do a lot in a single SQL query or statement. But that doesn’t mean you have to force yourself to use just one line of code, or that it’s a good idea to do every task with just one line of code. A common consequence of obtaining all results in one query is to obtain a Cartesian product. This happens when there are no conditions between the two tables in the query that limit their relationship. If you directly use two tables to perform a join query without corresponding restrictions, you will get a combination of each row in the first table and each row in the second table. Each such combination becomes a row in the result set, and you end up with a result set with many rows. It's important to consider that these queries are difficult to write, difficult to modify, and difficult to debug. The increasing number of database query requests should be expected. Managers want more complex reports and more fields on the user interface. If your design is complex and a single query, scaling them can be time-consuming and laborious. It's not worth the time spent on these things, either for you or the project. Break down complex spaghetti queries into several simple queries. When you split a complex SQL query, the result may be many similar queries, perhaps differing only in the data type. Writing all of these queries is tedious, so it's best to have a program that automatically generates the code. SQL code generation is a great application. Although SQL supports solving complex problems with a single line of code, don't do anything unrealistic. `,
			Case:     "这是一条很长很长的 SQL，案例略。",
			Func:     (*Query4Audit).RuleSpaghettiQueryAlert,
		},
		/*
		   https://www.datacamp.com/community/tutorials/sql-tutorial-query
		   The HAVING Clause
		   The HAVING clause was originally added to SQL because the WHERE keyword could not be used with aggregate functions. HAVING is typically used with the GROUP BY clause to restrict the groups of returned rows to only those that meet certain conditions. However, if you use this clause in your query, the index is not used, which -as you already know- can result in a query that doesn't really perform all that well.


		   If you’re looking for an alternative, consider using the WHERE clause. Consider the following queries:


		   SELECT state, COUNT(*)
		     FROM Drivers
		    WHERE state IN ('GA', 'TX')
		    GROUP BY state
		    ORDER BY state
		   SELECT state, COUNT(*)
		     FROM Drivers
		    GROUP BY state
		   HAVING state IN ('GA', 'TX')
		    ORDER BY state
		   The first query uses the WHERE clause to restrict the number of rows that need to be summed, whereas the second query sums up all the rows in the table and then uses HAVING to throw away the sums it calculated. In these types of cases, the alternative with the WHERE clause is obviously the better one, as you don’t waste any resources.


		   You see that this is not about limiting the result set, rather about limiting the intermediate number of records within a query.


		   Note that the difference between these two clauses lies in the fact that the WHERE clause introduces a condition on individual rows, while the HAVING clause introduces a condition on aggregations or results of a selection where a single result, such as MIN, MAX, SUM,… has been produced from multiple rows.
		*/
		"CLA.013": {
			Item:     "CLA.013",
			Severity: "L3",
			Summary:  "The HAVING clause is deprecated",
			Content:  `Rewrite the HAVING clause of the query to the query conditions in WHERE, and you can use the index during query processing. `,
			Case:     "SELECT s.c_id,count(s.c_id) FROM s where c = test GROUP BY s.c_id HAVING s.c_id <> ' 1660 ' AND s.c_id <> ' 2 ' order by s.c_id",
			Func:     (*Query4Audit).RuleHavingClause,
		},
		"CLA.014": {
			Item:     "CLA.014",
			Severity: "L2",
			Summary:  "It is recommended to use TRUNCATE instead of DELETE when deleting the entire table",
			Content:  `When deleting the entire table, it is recommended to use TRUNCATE instead of DELETE`,
			Case:     "delete from tbl",
			Func:     (*Query4Audit).RuleNoWhere,
		},
		"CLA.015": {
			Item:     "CLA.015",
			Severity: "L4",
			Summary:  "UPDATE does not specify a WHERE condition",
			Content:  `UPDATE without specifying a WHERE condition is generally fatal, please think twice before proceeding`,
			Case:     "update tbl set col=1",
			Func:     (*Query4Audit).RuleNoWhere,
		},
		"CLA.016": {
			Item:     "CLA.016",
			Severity: "L2",
			Summary:  "Don't UPDATE primary keys",
			Content:  `The primary key is the unique identifier of the record in the data table. It is not recommended to update the primary key column frequently, which will affect metadata statistics and thus affect normal queries. `,
			Case:     "update tbl set col=1",
			Func:     (*Query4Audit).RuleOK, // This suggestion is given to RuleUpdatePrimaryKey in indexAdvisor
		},
		"COL.001": {
			Item:     "COL.001",
			Severity: "L1",
			Summary:  "It is not recommended to use SELECT * type query",
			Content:  `When the table structure changes, using the * wildcard character to select all columns will cause the meaning and behavior of the query to change, possibly causing the query to return more data. `,
			Case:     "select * from tbl where id=1",
			Func:     (*Query4Audit).RuleSelectStar,
		},
		"COL.002": {
			Item:     "COL.002",
			Severity: "L2",
			Summary:  "INSERT/REPLACE does not specify column name",
			Content:  `When the table structure changes, if the INSERT or REPLACE request does not explicitly specify the column name, the request result will be different from expected; it is recommended to use "INSERT INTO tbl(col1, col2)VALUES..." instead. `,
			Case:     "insert into tbl values(1,' name ')",
			Func:     (*Query4Audit).RuleInsertColDef,
		},
		"COL.003": {
			Item:     "COL.003",
			Severity: "L2",
			Summary:  "It is recommended to modify the auto-increment ID to an unsigned type",
			Content:  `It is recommended to modify the auto-increment ID to an unsigned type`,
			Case:     "create table test(`id` int(11) NOT NULL AUTO_INCREMENT)",
			Func:     (*Query4Audit).RuleAutoIncUnsigned,
		},
		"COL.004": {
			Item:     "COL.004",
			Severity: "L1",
			Summary:  "Please add a default value for the column",
			Content:  `Please add a default value for the column. If it is an ALTER operation, please don't forget to write the default value of the original field. The fields have no default values, and the table structure cannot be changed online when the table is large. `,
			Case:     "CREATE TABLE tbl (col int) ENGINE=InnoDB;",
			Func:     (*Query4Audit).RuleAddDefaultValue,
		},
		"COL.005": {
			Item:     "COL.005",
			Severity: "L1",
			Summary:  "The column is not commented",
			Content:  `It is recommended to add comments to each column in the table to clarify the meaning and role of each column in the table. `,
			Case:     "CREATE TABLE tbl (col int) ENGINE=InnoDB;",
			Func:     (*Query4Audit).RuleColCommentCheck,
		},
		"COL.006": {
			Item:     "COL.006",
			Severity: "L3",
			Summary:  "The table contains too many columns",
			Content:  `The table contains too many columns`,
			Case:     "CREATE TABLE tbl ( cols ....);",
			Func:     (*Query4Audit).RuleTooManyFields,
		},
		"COL.007": {
			Item:     "COL.007",
			Severity: "L3",
			Summary:  "Table contains too many text/blob columns",
			Content:  fmt.Sprintf(`The table contains more than %d text/blob columns`, common.Config.MaxTextColsCount),
			Case:     "CREATE TABLE tbl ( cols ....);",
			Func:     (*Query4Audit).RuleTooManyFields,
		},
		"COL.008": {
			Item:     "COL.008",
			Severity: "L1",
			Summary:  "You can use VARCHAR instead of CHAR and VARBINARY instead of BINARY",
			Content:  `First of all, the storage space of variable-length fields is small, which can save storage space. Secondly, for queries, searching in a relatively small field is obviously more efficient. `,
			Case:     "create table t1(id int,name char(20),last_time date)",
			Func:     (*Query4Audit).RuleVarcharVSChar,
		},
		"COL.009": {
			Item:     "COL.009",
			Severity: "L2",
			Summary:  "It is recommended to use precise data types",
			Content:  `In fact, any design that uses the FLOAT, REAL or DOUBLE PRECISION data types is likely to be an anti-pattern. Most applications use floating-point numbers that do not need to meet the maximum/minimum ranges defined by the IEEE 754 standard. The cumulative impact of inexact floating point numbers can be severe when calculating totals. Use the NUMERIC or DECIMAL types in SQL instead of FLOAT and similar data types for fixed-precision decimal storage. These data types store data exactly according to the precision you specify when you define the column. Whenever possible, avoid using floating point numbers. `,
			Case:     "CREATE TABLE tab2 (p_id BIGINT UNSIGNED NOT NULL, a_id BIGINT UNSIGNED NOT NULL, hours float not null, PRIMARY KEY (p_id, a_id))",
			Func:     (*Query4Audit).RuleImpreciseDataType,
		},
		"COL.010": {
			Item:     "COL.010",
			Severity: "L2",
			Summary:  "ENUM/BIT/SET data types are deprecated",
			Content:  `ENUM defines the type of values in the column. When using strings to represent the values in ENUM, the data actually stored in the column is the ordinal number of these values at the time of definition. Therefore, the data for this column is byte aligned, and when you do a sorted query, the results are sorted by the actual stored ordinal values, not by the alphabetical order of the string values. This may not be what you want. There is no syntax for adding or removing a value from an ENUM or check constraint; you can only redefine the column using a new collection. If you're going to scrap an option, you might be bothered with historical data. As a strategy, changes to metadata—that is, changes to table and column definitions—should be uncommon and subject to testing and quality assurance. There is a better solution for constraining optional values in a column: Create a check table with each row containing a candidate value that is allowed to appear in the column; then declare a foreign key constraint on the old table that references the new table. `,
			Case:     "create table tab1(status ENUM(' new ',' in progress ',' fixed '))",
			Func:     (*Query4Audit).RuleValuesInDefinition,
		},
		// This suggestion is migrated from sqlcheck. In actual production environment, every SQL statement for table creation will give this suggestion. If you read too much, you will be unhappy.
		"COL.011": {
			Item:     "COL.011",
			Severity: "L0",
			Summary:  "Use NULL when a unique constraint is required, use NOT NULL only when the column cannot have missing values",
			Content:  `NULL is different from 0, 10 times NULL is still NULL. NULL is not the same as the empty string. The result of combining a string with NULL in standard SQL is still NULL. NULL and FALSE are also different. If the three Boolean operations AND, OR and NOT involve NULL, the results will also confuse many people. When you declare a column as NOT NULL, it means that every value in the column must exist and be meaningful. Use NULL to represent a null value that does not exist of any type. When you declare a column as NOT NULL, it means that every value in the column must exist and be meaningful. `,
			Case:     "select c1,c2,c3 from tbl where c4 is null or c4 <> 1",
			Func:     (*Query4Audit).RuleNullUsage,
		},
		"COL.012": {
			Item:     "COL.012",
			Severity: "L5",
			Summary:  "It is not recommended to set fields of type TEXT, BLOB and JSON to NOT NULL",
			Content:  `TEXT, BLOB and JSON type fields cannot be specified with non-NULL default values. If the NOT NULL restriction is added, writing data without specifying a value for the field may result in writing failure. `,
			Case:     "CREATE TABLE `tb`(`c` longblob NOT NULL);",
			Func:     (*Query4Audit).RuleBLOBNotNull,
		},
		"COL.013": {
			Item:     "COL.013",
			Severity: "L4",
			Summary:  "TIMESTAMP type default value check exception",
			Content:  `TIMESTAMP type recommends setting a default value, and it is not recommended to use 0 or 0000-00-00 00:00:00 as the default value. You can consider using 1970-08-02 01:01:01`,
			Case:     "CREATE TABLE tbl( `id` bigint not null, `create_time` timestamp);",
			Func:     (*Query4Audit).RuleTimestampDefault,
		},
		"COL.014": {
			Item:     "COL.014",
			Severity: "L5",
			Summary:  "Character set specified for column",
			Content:  `It is recommended that the column and table use the same character set. Do not specify the character set of the column separately. `,
			Case:     "CREATE TABLE `tb2` ( `id` int(11) DEFAULT NULL, `col` char(10) CHARACTER SET utf8 DEFAULT NULL)",
			Func:     (*Query4Audit).RuleColumnWithCharset,
		},
		// https://stackoverflow.com/questions/3466872/why-cant-a-text-column-have-a-default-value-in-mysql
		"COL.015": {
			Item:     "COL.015",
			Severity: "L4",
			Summary:  "TEXT, BLOB and JSON type fields cannot specify non-NULL default values",
			Content:  `Non-NULL default values cannot be specified for TEXT, BLOB and JSON type fields in the MySQL database. The maximum length of TEXT is 2^16-1 characters, the maximum length of MEDIUMTEXT is 2^32-1 characters, and the maximum length of LONGTEXT is 2^64-1 characters. `,
			Case:     "CREATE TABLE `tbl` (`c` blob DEFAULT NULL);",
			Func:     (*Query4Audit).RuleBlobDefaultValue,
		},
		"COL.016": {
			Item:     "COL.016",
			Severity: "L1",
			Summary:  "It is recommended to use INT(10) or BIGINT(20) for integer definition",
			Content:  `INT(M) In the integer data type, M represents the maximum display width. In INT(M), the value of M has nothing to do with how much storage space INT(M) occupies. INT(3), INT(4), and INT(8) all occupy 4 bytes of storage space on the disk. Setting integer display width is no longer recommended in higher versions of MySQL. `,
			Case:     "CREATE TABLE tab (a INT(1));",
			Func:     (*Query4Audit).RuleIntPrecision,
		},
		"COL.017": {
			Item:     "COL.017",
			Severity: "L2",
			Summary:  "VARCHAR definition length is too long",
			Content:  fmt.Sprintf(`varchar is a variable-length string, no storage space is allocated in advance, and the length should not exceed %d. If the storage length is too long, MySQL will define the field type as text, create a separate table, and use the primary key to correspond , to avoid affecting the index efficiency of other fields.`, common.Config.MaxVarcharLength),
			Case:     "CREATE TABLE tab (a varchar(3500));",
			Func:     (*Query4Audit).RuleVarcharLength,
		},
		"COL.018": {
			Item:     "COL.018",
			Severity: "L9",
			Summary:  "A deprecated field type is used in the table creation statement",
			Content:  "The following field types are not recommended:" + strings.Join(common.Config.ColumnNotAllowType, ", "),
			Case:     "CREATE TABLE tab (a BOOLEAN);",
			Func:     (*Query4Audit).RuleColumnNotAllowType,
		},
		"COL.019": {
			Item:     "COL.019",
			Severity: "L1",
			Summary:  "It is not recommended to use time data types with precision below the second level",
			Content:  "Using high-precision time data types consumes relatively large storage space; MySQL can only support time data types accurate to microseconds in version 5.6.4 or above. Version compatibility issues need to be considered when using them.",
			Case:     "CREATE TABLE t1 (t TIME(3), dt DATETIME(6));",
			Func:     (*Query4Audit).RuleTimePrecision,
		},
		"DIS.001": {
			Item:     "DIS.001",
			Severity: "L1",
			Summary:  "Eliminate unnecessary DISTINCT conditions",
			Content:  `Too many DISTINCT conditions are a symptom of complex, wraparound queries. Consider breaking complex queries into many simpler queries and reducing the number of DISTINCT conditions. If the primary key column is part of the column's result set, the DISTINCT condition may have no effect. `,
			Case:     "SELECT DISTINCT c.c_id,count(DISTINCT c.c_name),count(DISTINCT c.c_e),count(DISTINCT c.c_n),count(DISTINCT c.c_me),c.c_d FROM (select distinct id, name from B) as e WHERE e.country_id = c.country_id",
			Func:     (*Query4Audit).RuleDistinctUsage,
		},
		"DIS.002": {
			Item:     "DIS.002",
			Severity: "L3",
			Summary:  "COUNT(DISTINCT) results may be different than you expect when using multiple columns",
			Content:  `COUNT(DISTINCT col) Counts the number of unique rows in this column except NULL. Note that COUNT(DISTINCT col, col2) returns 0 if one of the columns is all NULL, even if the other column has a different value. `,
			Case:     "SELECT COUNT(DISTINCT col, col2) FROM tbl;",
			Func:     (*Query4Audit).RuleCountDistinctMultiCol,
		},
		// DIS.003 is inspired by the following link
		// http://www.ijstr.org/final-print/oct2015/Query-Optimization-Techniques-Tips-For-Writing-Efficient-And-Faster-Sql-Queries.pdf
		"DIS.003": {
			Item:     "DIS.003",
			Severity: "L3",
			Summary:  "DISTINCT * has no meaning for tables with primary keys",
			Content:  `When the table already has a primary key, the output result of DISTINCT on all columns is the same as the result of no DISTINCT operation. Please do not add any unnecessary extravagance. `,
			Case:     "SELECT DISTINCT * FROM film;",
			Func:     (*Query4Audit).RuleDistinctStar,
		},
		"FUN.001": {
			Item:     "FUN.001",
			Severity: "L2",
			Summary:  "Avoid using functions or other operators in WHERE conditions",
			Content:  `Although using functions in SQL can simplify many complex queries, queries using functions cannot make use of the indexes that have been established in the table. The query will be a full table scan with poor performance. It is generally recommended to write column names on the left side of the comparison operator and put query filter conditions on the right side of the comparison operator. It is also not recommended to write extra parentheses on both sides of the query comparison conditions, which will cause greater trouble in reading. `,
			Case:     "select id from t where substring(name,1,3)=' abc '",
			Func:     (*Query4Audit).RuleCompareWithFunction,
		},
		"FUN.002": {
			Item:     "FUN.002",
			Severity: "L1",
			Summary:  "Poor performance when using COUNT(*) operations when WHERE conditions are specified or when a non-MyISAM engine is specified",
			Content:  `COUNT(*) is used to count the number of table rows, and COUNT(COL) is used to count the number of rows with non-NULL specified columns. MyISAM tables are specially optimized for COUNT(*) to count the number of rows in the entire table, which is usually very fast. However, for non-MyISAM tables or certain WHERE conditions specified, the COUNT(*) operation needs to scan a large number of rows to obtain accurate results, and the performance is therefore poor. Sometimes some business scenarios do not require a completely accurate COUNT value. In this case, an approximate value can be used instead. The number of rows estimated by the optimizer from EXPLAIN is a good approximation. Executing EXPLAIN does not require actual execution of the query, so the cost is very low. `,
			Case:     "SELECT c3, COUNT(*) AS accounts FROM tab where c2 < 10000 GROUP BY c3 ORDER BY num",
			Func:     (*Query4Audit).RuleCountStar,
		},
		"FUN.003": {
			Item:     "FUN.003",
			Severity: "L3",
			Summary:  "Using string concatenation into nullable columns",
			Content:  `In some query requests, you need to force a certain column or expression to return a non-NULL value to make the query logic simpler, but you do not want to save this value. You can use the COALESCE() function to construct a concatenated expression so that even null-valued columns do not cause the entire expression to become NULL. `,
			Case:     "select c1 || coalesce(' ' || c2 || ' ', ' ') || c3 as c from tbl",
			Func:     (*Query4Audit).RuleStringConcatenation,
		},
		"FUN.004": {
			Item:     "FUN.004",
			Severity: "L4",
			Summary:  "The use of the SYSDATE() function is deprecated",
			Content:  `SYSDATE() function may cause master-slave data inconsistency, please use NOW() function instead of SYSDATE(). `,
			Case:     "SELECT SYSDATE();",
			Func:     (*Query4Audit).RuleSysdate,
		},
		"FUN.005": {
			Item:     "FUN.005",
			Severity: "L1",
			Summary:  "Using COUNT(col) or COUNT(constant) is not recommended",
			Content:  `Do not use COUNT(col) or COUNT(constant) instead of COUNT(*). COUNT(*) is a standard method of counting rows defined by SQL92. It has nothing to do with data, and has nothing to do with NULL and non-NULL. `,
			Case:     "SELECT COUNT(1) FROM tbl;",
			Func:     (*Query4Audit).RuleCountConst,
		},
		"FUN.006": {
			Item:     "FUN.006",
			Severity: "L1",
			Summary:  "Be aware of NPE issues when using SUM(COL)",
			Content:  `When the values of a certain column are all NULL, the return result of COUNT(COL) is 0, but the return result of SUM(COL) is NULL, so you need to pay attention to the NPE problem when using SUM(). You can use the following method to avoid SUM's NPE problem: SELECT IF(ISNULL(SUM(COL)), 0, SUM(COL)) FROM tbl`,
			Case:     "SELECT SUM(COL) FROM tbl;",
			Func:     (*Query4Audit).RuleSumNPE,
		},
		"FUN.007": {
			Item:     "FUN.007",
			Severity: "L1",
			Summary:  "The use of triggers is not recommended",
			Content:  `The execution of triggers has no feedback and logs, hiding the actual execution steps. When a problem occurs in the database, the specific execution status of the trigger cannot be analyzed through slow logs, making it difficult to find problems. In MySQL, triggers cannot be temporarily closed or opened. In scenarios such as data migration or data recovery, triggers need to be temporarily dropped, which may affect the production environment. `,
			Case:     "CREATE TRIGGER t1 AFTER INSERT ON work FOR EACH ROW INSERT INTO time VALUES(NOW());",
			Func:     (*Query4Audit).RuleForbiddenTrigger,
		},
		"FUN.008": {
			Item:     "FUN.008",
			Severity: "L1",
			Summary:  "The use of stored procedures is not recommended",
			Content:  `There is no version control for stored procedures, and it is difficult to achieve business-free upgrade of stored procedures in conjunction with the business. Stored procedures also have problems with expansion and migration. `,
			Case:     "CREATE PROCEDURE simpleproc (OUT param1 INT);",
			Func:     (*Query4Audit).RuleForbiddenProcedure,
		},
		"FUN.009": {
			Item:     "FUN.009",
			Severity: "L1",
			Summary:  "Using custom functions is not recommended",
			Content:  `Using custom functions is not recommended`,
			Case:     "CREATE FUNCTION hello (s CHAR(20));",
			Func:     (*Query4Audit).RuleForbiddenFunction,
		},
		"GRP.001": {
			Item:     "GRP.001",
			Severity: "L2",
			Summary:  "It is not recommended to use GROUP BY for equivalent value query columns",
			Content:  `The columns in GROUP BY use equal value query in the previous WHERE condition. It makes little sense to perform GROUP BY on such columns. `,
			Case:     "select film_id, title from film where release_year=' 2006 ' group by release_year",
			Func:     (*Query4Audit).RuleOK, //RuleGroupByConst

		},
		//RuleGroupByConst，这个在CheckHeuristicRules的时候会直接跳过 反而是在idxAdvisor.HeuristicCheck的时候真正去检查 所以一开始非线上检查和这里结合table info检查不会冲突
		//这里你的release_year已经都是2006了根据他分组没意义 但是这里只是用到了ast 也没有用到table info 所谓的循环也只是检查groupby和where重合了么
		"JOI.001": {
			Item:     "JOI.001",
			Severity: "L2",
			Summary:  "JOIN statement mixes commas and ANSI modes",
			Content:  `Mixing commas and ANSI JOIN when connecting tables is not easy for humans to understand, and the table connection behaviors and priorities of different versions of MySQL are different. When the MySQL version changes, errors may be introduced. `,
			Case:     "select c1,c2,c3 from t1,t2 join t3 on t1.c1=t2.c1,t1.c3=t3,c1 where id>1000",
			Func:     (*Query4Audit).RuleCommaAnsiJoin,
		},
		"JOI.002": {
			Item:     "JOI.002",
			Severity: "L4",
			Summary:  "The same table is joined twice",
			Content:  `The same table appearing at least twice in the FROM clause can be reduced to a single access to the table. `,
			Case:     "select tb1.col from (tb1, tb2) join tb2 on tb1.id=tb.id where tb1.id=1",
			Func:     (*Query4Audit).RuleDupJoin,
		},
		//"JOI.003": {
		// Item: "JOI.003",
		// Severity: "L4",
		// Summary: "OUTER JOIN invalid",
		// Content: `No data is returned from the external table of OUTER JOIN due to WHERE condition error, which will implicitly convert the query to INNER JOIN. For example: select c from L left join R using(c) where La=5 and Rb=10. There may be errors in this SQL logic or the programmer may have a misunderstanding of how OUTER JOIN works, because LEFT/RIGHT JOIN is the abbreviation of LEFT/RIGHT OUTER JOIN. `,
		// Case: "select c1,c2,c3 from t1 left outer join t2 using(c1) where t1.c2=2 and t2.c3=4",
		// Func: (*Query4Audit).RuleOK, // TODO
		//},
		//"JOI.004": {
		// Item: "JOI.004",
		// Severity: "L4",
		// Summary: "Exclusive JOIN is not recommended",
		// Content: `Only the LEFT OUTER JOIN statement with WHERE clause where the right table is NULL, it is possible that the wrong column is used in the WHERE clause, such as: "... FROM l LEFT OUTER JOIN r ON ll = rr WHERE rz IS NULL", the correct logic of this query may be WHERE rr IS NULL. `,
		// Case: "select c1,c2,c3 from t1 left outer join t2 on t1.c1=t2.c1 where t2.c2 is null",
		// Func: (*Query4Audit).RuleOK, // TODO
		//},
		"JOI.005": {
			Item:     "JOI.005",
			Severity: "L2",
			Summary:  "Reduce the number of JOINs",
			Content:  `Too many JOINs are a symptom of complex, wraparound queries. Consider breaking complex queries into many simpler queries and reducing the number of JOINs. `,
			Case:     "select bp1.p_id, b1.d_d as l, b1.b_id from b1 join bp1 on (b1.b_id = bp1.b_id) left outer join (b1 as b2 join bp2 on (b2.b_id = bp2.b_id) ) on (bp1.p_id = bp2.p_id ) join bp21 on (b1.b_id = bp1.b_id) join bp31 on (b1.b_id = bp1.b_id) join bp41 on (b1.b_id = bp1.b_id) where b2. b_id = 0",
			Func:     (*Query4Audit).RuleReduceNumberOfJoin,
		},
		"JOI.006": {
			Item:     "JOI.006",
			Severity: "L4",
			Summary:  "Rewriting nested queries as JOINs often results in more efficient execution and more effective optimization",
			Content:  `In general, non-nested subqueries are always used with correlated subqueries, at most one from a table in the FROM clause, for predicates of ANY, ALL and EXISTS. An unrelated subquery or a subquery from multiple tables in the FROM clause is flattened if it can be determined based on query semantics that the subquery returns at most one row. `,
			Case:     "SELECT s,p,d FROM tbl WHERE p.p_id = (SELECT s.p_id FROM tbl WHERE s.c_id = 100996 AND sq = 1 )",
			Func:     (*Query4Audit).RuleNestedSubQueries,
		},
		"JOI.007": {
			Item:     "JOI.007",
			Severity: "L4",
			Summary:  "It is not recommended to use joint tables to delete or update",
			Content:  `When you need to delete or update multiple tables at the same time, it is recommended to use simple statements. One SQL only deletes or updates one table. Try not to operate multiple tables in the same statement. `,
			Case:     "UPDATE users u LEFT JOIN hobby h ON u.id = h.uid SET u.name = ' pianoboy ' WHERE h.hobby = ' piano ';",
			Func:     (*Query4Audit).RuleMultiDeleteUpdate,
		},
		"JOI.008": {
			Item:     "JOI.008",
			Severity: "L4",
			Summary:  "Don't use cross-database JOIN queries",
			Content:  `Generally speaking, a cross-database JOIN query means that the query statement spans two different subsystems, which may mean that the system coupling is too high or the database table structure design is unreasonable. `,
			Case:     "SELECT s,p,d FROM tbl WHERE p.p_id = (SELECT s.p_id FROM tbl WHERE s.c_id = 100996 AND sq = 1 )",
			Func:     (*Query4Audit).RuleMultiDBJoin,
		},
		// TODO: Check for cross-database transactions. Currently, SOAR does not process transactions.
		"KEY.001": {
			Item:     "KEY.001",
			Severity: "L2",
			Summary:  "It is recommended to use an auto-increment column as the primary key. If you use a joint auto-increment primary key, please use the auto-increment key as the first column",
			Content:  `It is recommended to use an auto-increment column as the primary key. If you use a joint auto-increment primary key, please use the auto-increment key as the first column`,
			Case:     "create table test(`id` int(11) NOT NULL PRIMARY KEY (`id`))",
			Func:     (*Query4Audit).RulePKNotInt,
		},
		"KEY.002": {
			Item:     "KEY.002",
			Severity: "L4",
			Summary:  "There is no primary key or unique key, and the table structure cannot be changed online",
			Content:  `No primary key or unique key, table structure cannot be changed online`,
			Case:     "create table test(col varchar(5000))",
			Func:     (*Query4Audit).RuleNoOSCKey,
		},
		"KEY.003": {
			Item:     "KEY.003",
			Severity: "L4",
			Summary:  "Avoid recursive relationships such as foreign keys",
			Content:  `Data with recursive relationships is common, and data is often organized like a tree or in a hierarchical manner. However, creating a foreign key constraint to enforce a relationship between two columns in the same table can lead to clumsy queries. Each level of the tree corresponds to another connection. You will need to issue a recursive query to get all descendants or all ancestors of a node. The solution is to construct an additional closure table. It records the relationships between all nodes in the tree, not just those with direct parent-child relationships. You can also compare different levels of data design: closure tables, path enumerations, nested sets. Then choose one based on your application's needs. `,
			Case:     "CREATE TABLE tab2 (p_id BIGINT UNSIGNED NOT NULL,a_id BIGINT UNSIGNED NOT NULL,PRIMARY KEY (p_id, a_id),FOREIGN KEY (p_id) REFERENCES tab1(p_id),FOREIGN KEY (a_id) REFERENCES tab3(a_id))",
			Func:     (*Query4Audit).RuleRecursiveDependency,
		},
		// TODO: Add a new composite index. The fields are sorted from large to small according to whether the granularity is large or small. The most differentiated one is on the far left.
		"KEY.004": {
			Item:     "KEY.004",
			Severity: "L0",
			Summary:  "Reminder: Please align index attribute order with query",
			Content:  `If you create a composite index for a column, make sure that the query properties are in the same order as the index properties so that the DBMS uses the index when processing the query. If query and index attribute orders are not aligned, then the DBMS may not be able to use the index during query processing. `,
			Case:     "create index idx1 on tbl (last_name,first_name)",
			Func:     (*Query4Audit).RuleIndexAttributeOrder,
		},
		"KEY.005": {
			Item:     "KEY.005",
			Severity: "L2",
			Summary:  "Too many indexes are created for the table",
			Content:  `Too many indexes are created for the table`,
			Case:     "CREATE TABLE tbl (a int, b int, c int, KEY idx_a (`a`),KEY idx_b(`b`),KEY idx_c(`c`));",
			Func:     (*Query4Audit).RuleTooManyKeys,
		},
		"KEY.006": {
			Item:     "KEY.006",
			Severity: "L4",
			Summary:  "Too many columns in primary key",
			Content:  `Too many columns in primary key`,
			Case:     "CREATE TABLE tbl (a int, b int, c int, PRIMARY KEY(`a`,`b`,`c`));",
			Func:     (*Query4Audit).RuleTooManyKeyParts,
		},
		"KEY.007": {
			Item:     "KEY.007",
			Severity: "L4",
			Summary:  "The primary key is not specified or the primary key is not int or bigint",
			Content:  `The primary key is not specified or the primary key is not int or bigint. It is recommended to set the primary key to int unsigned or bigint unsigned. `,
			Case:     "CREATE TABLE tbl (a int);",
			Func:     (*Query4Audit).RulePKNotInt,
		},
		"KEY.008": {
			Item:     "KEY.008",
			Severity: "L4",
			Summary:  "ORDER BY multiple columns but the sort direction may not be able to use the index",
			Content:  `Before MySQL 8.0, when ORDER BY multiple columns specify different sort directions, the established index cannot be used. `,
			Case:     "SELECT * FROM tbl ORDER BY a DESC, b ASC;",
			Func:     (*Query4Audit).RuleOrderByMultiDirection,
		},
		"KEY.009": {
			Item:     "KEY.009",
			Severity: "L0",
			Summary:  "Please check the uniqueness of the data before adding a unique index",
			Content:  `Please check the data uniqueness of added unique index columns in advance. If the data is not unique, duplicate columns may be automatically deleted when the online table structure is adjusted, which may lead to data loss. `,
			Case:     "CREATE UNIQUE INDEX part_of_name ON customer (name(10));",
			Func:     (*Query4Audit).RuleUniqueKeyDup,
		},
		"KEY.010": {
			Item:     "KEY.010",
			Severity: "L0",
			Summary:  "Full-text indexing is not a silver bullet",
			Content:  `Full-text index is mainly used to solve the performance problem of fuzzy query, but the frequency and concurrency of the query need to be controlled. At the same time, pay attention to adjusting parameters such as ft_min_word_len, ft_max_word_len, ngram_token_size and so on. `,
			Case:     "CREATE TABLE `tb` ( `id` int(10) unsigned NOT NULL AUTO_INCREMENT, `ip` varchar(255) NOT NULL DEFAULT '', PRIMARY KEY (`id`), FULLTEXT KEY `ip` (`ip `) ) ENGINE=InnoDB;",
			Func:     (*Query4Audit).RuleFulltextIndex,
		},
		"KWR.001": {
			Item:     "KWR.001",
			Severity: "L2",
			Summary:  "SQL_CALC_FOUND_ROWS is inefficient",
			Content:  `Because SQL_CALC_FOUND_ROWS does not scale well, it may cause performance problems; it is recommended that the business use other strategies to replace the counting function provided by SQL_CALC_FOUND_ROWS, such as: paging result display, etc. `,
			Case:     "select SQL_CALC_FOUND_ROWS col from tbl where id>1000",
			Func:     (*Query4Audit).RuleSQLCalcFoundRows,
		},
		"KWR.002": {
			Item:     "KWR.002",
			Severity: "L2",
			Summary:  "It is not recommended to use MySQL keywords as column names or table names",
			Content:  `When using keywords as column names or table names, the program needs to escape the column names and table names. If ignored, the request will not be executed. `,
			Case:     "CREATE TABLE tbl (`select` int)",
			Func:     (*Query4Audit).RuleUseKeyWord,
		},
		"KWR.003": {
			Item:     "KWR.003",
			Severity: "L1",
			Summary:  "It is not recommended to use plural numbers for column or table names",
			Content:  `The table name should only represent the entity content in the table, and should not represent the number of entities. The corresponding DO class name is also in singular form, which is consistent with expression habits. `,
			Case:     "CREATE TABLE tbl (`books` int)",
			Func:     (*Query4Audit).RulePluralWord,
		},
		"KWR.004": {
			Item:     "KWR.004",
			Severity: "L1",
			Summary:  "It is not recommended to use multi-byte encoded characters (Chinese) for naming",
			Content:  `It is recommended to use English, numbers, underscores and other characters when naming libraries, tables, columns, and aliases. It is not recommended to use Chinese or other multi-byte encoded characters. `,
			Case:     "select col as column from tb",
			Func:     (*Query4Audit).RuleMultiBytesWord,
		},
		"KWR.005": {
			Item:     "KWR.005",
			Severity: "L1",
			Summary:  "SQL contains unicode special characters",
			Content:  "Some IDEs will automatically insert invisible unicode characters into SQL. For example: non-break space, zero-width space, etc. Under Linux, you can use the `cat -A file.sql` command to view invisible characters.",
			Case:     "update tb set status = 1 where id = 1;",
			Func:     (*Query4Audit).RuleInvisibleUnicode,
		},
		"LCK.001": {
			Item:     "LCK.001",
			Severity: "L3",
			Summary:  "INSERT INTO xx SELECT has a large locking granularity, please be careful",
			Content:  `INSERT INTO xx SELECT Please be careful as the locking granularity is large`,
			Case:     "INSERT INTO tbl SELECT * FROM tbl2;",
			Func:     (*Query4Audit).RuleInsertSelect,
		},
		"LCK.002": {
			Item:     "LCK.002",
			Severity: "L3",
			Summary:  "Please use INSERT ON DUPLICATE KEY UPDATE with caution",
			Content:  `Using INSERT ON DUPLICATE KEY UPDATE when the primary key is an auto-increment key may cause a large number of discontinuous and rapid growth of the primary key, causing the primary key to quickly overflow and unable to continue writing. In extreme cases, it may lead to inconsistency between master and slave data. `,
			Case:     "INSERT INTO t1(a,b,c) VALUES (1,2,3) ON DUPLICATE KEY UPDATE c=c+1;",
			Func:     (*Query4Audit).RuleInsertOnDup,
		},
		"LIT.001": {
			Item:     "LIT.001",
			Severity: "L2",
			Summary:  "Storing IP addresses in character type",
			Content:  `The string literally looks like an IP address, but is not a parameter to INET_ATON(), indicating that the data is stored as characters rather than integers. It is more efficient to store IP addresses as integers. `,
			Case:     "insert into tbl (IP,name) values(' 10.20.306.122 ',' test ')",
			Func:     (*Query4Audit).RuleIPString,
		},
		"LIT.002": {
			Item:     "LIT.002",
			Severity: "L4",
			Summary:  "Date/time not enclosed in quotes",
			Content:  `A query like "WHERE col <2010-02-12" is valid SQL, but may be a bug as it will be interpreted as "WHERE col <1996"; date/time literals should be quoted, And there should be no spaces before and after the quotation marks. `,
			Case:     "select col1,col2 from tbl where time < 2018-01-10",
			Func:     (*Query4Audit).RuleDateNotQuote,
		},
		"LIT.003": {
			Item:     "LIT.003",
			Severity: "L3",
			Summary:  "A collection of related data stored in a column",
			Content:  `Storing IDs as a list as VARCHAR/TEXT columns can cause performance and data integrity issues. Querying such columns requires the use of pattern matching expressions. Using comma-separated lists to do multi-table join queries to locate a row of data is extremely inelegant and time-consuming. This will make it more difficult to verify the ID. Think about it, how much data can the list support at most? Instead of using multi-valued attributes, store the IDs in a separate table so that each individual attribute value can occupy a row. In this way, the crosstab implements a many-to-many relationship between the two tables. This will simplify queries better and also validate IDs more efficiently. `,
			Case:     "select c1,c2,c3,c4 from tab1 where col_id REGEXP ' [[:<:]]12[[:>:]] '",
			Func:     (*Query4Audit).RuleMultiValueAttribute,
		},
		//"LIT.004": {
		// Item: "LIT.004",
		// Severity: "L1",
		// Summary: "Please use a semicolon or a set DELIMITER at the end",
		// Content: `USE database, SHOW DATABASES and other commands also need to end with a semicolon or a set DELIMITER. `,
		// Case: "USE db",
		// Func: (*Query4Audit).RuleOK, // TODO: RuleAddDelimiter
		//},
		"RES.001": {
			Item:     "RES.001",
			Severity: "L4",
			Summary:  "Non-deterministic GROUP BY",
			Content:  `SQL returns columns that are neither in the aggregate function nor in the GROUP BY expression, so the results for these values will be non-deterministic. For example: select a, b, c from tbl where foo="bar" group by a, the result returned by this SQL is uncertain. `,
			Case:     "select c1,c2,c3 from t1 where c2=' foo ' group by c2",
			Func:     (*Query4Audit).RuleNoDeterministicGroupby,
		},
		"RES.002": {
			Item:     "RES.002",
			Severity: "L4",
			Summary:  "LIMIT query without ORDER BY",
			Content:  `LIMIT without ORDER BY can lead to non-deterministic results, depending on the query execution plan. `,
			Case:     "select col1,col2 from tbl where name=xx limit 10",
			Func:     (*Query4Audit).RuleNoDeterministicLimit,
		},
		"RES.003": {
			Item:     "RES.003",
			Severity: "L4",
			Summary:  "UPDATE/DELETE operations use LIMIT conditions",
			Content:  `UPDATE/DELETE operations using LIMIT conditions are as dangerous as not adding WHERE conditions. It may lead to master-slave data inconsistency or slave database synchronization interruption. `,
			Case:     "UPDATE film SET length = 120 WHERE title = ' abc ' LIMIT 1;",
			Func:     (*Query4Audit).RuleUpdateDeleteWithLimit,
		},
		"RES.004": {
			Item:     "RES.004",
			Severity: "L4",
			Summary:  "UPDATE/DELETE operations specify ORDER BY conditions",
			Content:  `Do not specify ORDER BY conditions for UPDATE/DELETE operations. `,
			Case:     "UPDATE film SET length = 120 WHERE title = ' abc ' ORDER BY title",
			Func:     (*Query4Audit).RuleUpdateDeleteWithOrderby,
		},
		"RES.005": {
			Item:     "RES.005",
			Severity: "L4",
			Summary:  "The UPDATE statement may contain logical errors, causing data corruption",
			Content:  "In an UPDATE statement, if you want to update multiple fields, AND cannot be used between the fields, but should be separated by commas.",
			Case:     "update tbl set col = 1 and cl = 2 where col=3;",
			Func:     (*Query4Audit).RuleUpdateSetAnd,
		},
		"RES.006": {
			Item:     "RES.006",
			Severity: "L4",
			Summary:  "Never really compare conditions",
			Content:  "The query condition is never true. If the condition appears in where, it may cause the query to have no matching results.",
			Case:     "select * from tbl where 1 != 1;",
			Func:     (*Query4Audit).RuleImpossibleWhere,
		},
		"RES.007": {
			Item:     "RES.007",
			Severity: "L4",
			Summary:  "A comparison condition that is always true",
			Content:  "The query condition is always true, which may cause the WHERE condition to fail and perform a full table query.",
			Case:     "select * from tbl where 1 = 1;",
			Func:     (*Query4Audit).RuleMeaninglessWhere,
		},
		"RES.008": {
			Item:     "RES.008",
			Severity: "L2",
			Summary:  "LOAD DATA/SELECT ... INTO OUTFILE is not recommended",
			Content:  "SELECT INTO OUTFILE requires granting FILE permission, which may introduce security issues. Although LOAD DATA can increase the speed of data import, it may also cause excessive synchronization delays from the database.",
			Case:     "LOAD DATA INFILE ' data.txt ' INTO TABLE db2.my_table;",
			Func:     (*Query4Audit).RuleLoadFile,
		},
		"RES.009": {
			Item:     "RES.009",
			Severity: "L2",
			Summary:  "It is not recommended to use continuous judgment",
			Content:  "A statement like SELECT * FROM tbl WHERE col = col = ' abc ' may be a writing error. What you may want to express is col = ' abc '. If it is indeed a business requirement, it is recommended to modify it to col = col and col = ' abc '.",
			Case:     "SELECT * FROM tbl WHERE col = col = ' abc '",
			Func:     (*Query4Audit).RuleMultiCompare,
		},
		"RES.010": {
			Item:     "RES.010",
			Severity: "L2",
			Summary:  "Fields defined as ON UPDATE CURRENT_TIMESTAMP in the table creation statement are not recommended to contain business logic",
			Content:  "Fields defined as ON UPDATE CURRENT_TIMESTAMP will be modified together when other fields in the table are updated. If it contains business logic and is visible to users, there will be hidden dangers. If there is subsequent batch modification of data but you do not want to modify the field, it will lead to data errors. ",
			Case:     `CREATE TABLE category (category_id TINYINT UNSIGNED NOT NULL AUTO_INCREMENT, name VARCHAR(25) NOT NULL, last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, PRIMARY KEY (category_id)`,
			Func:     (*Query4Audit).RuleCreateOnUpdate,
		},
		//"RES.011": {
		//	Item:     "RES.011",
		//	Severity: "L2",
		//	Summary:  "更新请求操作的表包含 ON UPDATE CURRENT_TIMESTAMP 字段",
		//	Content:  "定义为 ON UPDATE CURRENT_TIMESTAMP 的字段在该表其他字段更新时会联动修改，请注意检查。如不想修改字段的更新时间可以使用如下方法：UPDATE category SET name='ActioN', last_update=last_update WHERE category_id=1",
		//	Case:     "UPDATE category SET name='ActioN', last_update=last_update WHERE category_id=1",
		//	Func:     (*Query4Audit).RuleOK, // 该建议在indexAdvisor中给 RuleUpdateOnUpdate
		//},
		"SEC.001": {
			Item:     "SEC.001",
			Severity: "L0",
			Summary:  "Please use TRUNCATE operation with caution",
			Content:  `Generally speaking, the fastest way to clear a table is to use the TRUNCATE TABLE tbl_name; statement. However, the TRUNCATE operation is not free. TRUNCATE TABLE cannot return the exact number of deleted rows. If you need to return the number of deleted rows, it is recommended to use the DELETE syntax. The TRUNCATE operation will also reset AUTO_INCREMENT. If you do not want to reset this value, it is recommended to use DELETE FROM tbl_name WHERE 1; instead. The TRUNCATE operation will add a source data lock (MDL) to the data dictionary. When TRUNCATE many tables at one time, it will affect all requests of the entire instance. Therefore, if you want to TRUNCATE multiple tables, it is recommended to use DROP+CREATE to reduce the lock time. `,
			Case:     "TRUNCATE TABLE tbl_name",
			Func:     (*Query4Audit).RuleTruncateTable,
		},
		"SEC.002": {
			Item:     "SEC.002",
			Severity: "L0",
			Summary:  "Do not store passwords in clear text",
			Content:  `It is not safe to store passwords in clear text or to pass passwords over the network in clear text. If an attacker can intercept the SQL statement you use to insert the password, they can read the password directly. In addition, inserting the string entered by the user into a pure SQL statement in clear text will also allow an attacker to discover it. If you can read your password, so can a hacker. The solution is to cryptographically encode the original password using a one-way hash function. Hashing is a function that converts an input string into another new, unrecognizable string. Add random strings to password encryption expressions to defend against "dictionary attacks." Do not enter clear text passwords into SQL queries. Calculate the hash string in the application code and only use the hash string in the SQL query. `,
			Case:     "create table test(id int,name varchar(20) not null,password varchar(200)not null)",
			Func:     (*Query4Audit).RuleReadablePasswords,
		},
		"SEC.003": {
			Item:     "SEC.003",
			Severity: "L0",
			Summary:  "Pay attention to backup when using DELETE/DROP/TRUNCATE and other operations",
			Content:  `It is necessary to back up data before performing high-risk operations. `,
			Case:     "delete from table where col = ' condition '",
			Func:     (*Query4Audit).RuleDataDrop,
		},
		"SEC.004": {
			Item:     "SEC.004",
			Severity: "L0",
			Summary:  "Discover common SQL injection functions",
			Content:  `SLEEP(), BENCHMARK(), GET_LOCK(), RELEASE_LOCK() and other functions usually appear in SQL injection statements, which will seriously affect database performance. `,
			Case:     "SELECT BENCHMARK(10, RAND())",
			Func:     (*Query4Audit).RuleInjection,
		},
		"STA.001": {
			Item:     "STA.001",
			Severity: "L0",
			Summary:  "' != ' operator is non-standard",
			Content:  `"<>" is the inequality operator in standard SQL. `,
			Case:     "select col1,col2 from tbl where type!=0",
			Func:     (*Query4Audit).RuleStandardINEQ,
		},
		"STA.002": {
			Item:     "STA.002",
			Severity: "L1",
			Summary:  "It is recommended not to add spaces after the library name or table name",
			Content:  `When accessing a table or field using the db.table or table.column format, do not add a space after the period, although this is syntactically correct. `,
			Case:     "select col from sakila. film",
			Func:     (*Query4Audit).RuleSpaceAfterDot,
		},
		"STA.003": {
			Item:     "STA.003",
			Severity: "L1",
			Summary:  "The index name is not standardized",
			Content:  `It is recommended that ordinary secondary indexes be prefixed with ` + common.Config.IdxPrefix + ` and unique indexes be prefixed with ` + common.Config.UkPrefix + `. `,
			Case:     "select col from now where type!=0",
			Func:     (*Query4Audit).RuleIdxPrefix,
		},
		"STA.004": {
			Item:     "STA.004",
			Severity: "L1",
			Summary:  "Please do not use characters other than letters, numbers and underscores when naming",
			Content:  `Start with a letter or underscore. Only letters, numbers and underscores are allowed in the name. Please use consistent case and do not use camel case. Do not use consecutive underscores ' __ ' in the name , as this will make it difficult to read. `,
			Case:     "CREATE TABLE ` abc` (a int);",
			Func:     (*Query4Audit).RuleStandardName,
		},
		//"SUB.001": {
		//	Item:     "SUB.001",
		//	Severity: "L4",
		//	Summary:  "MySQL does not optimize subqueries well",
		//	Content:  `MySQL executes a subquery for each row in the outer query as a dependent subquery. This is a common cause of serious performance issues. This may be improved in MySQL version 5.6, but for versions 5.1 and earlier, it is recommended to rewrite this type of query as a JOIN or LEFT OUTER JOIN respectively. `,
		//	Case:     "select col1,col2,col3 from table1 where col2 in(select col from table2)",
		//	Func:     (*Query4Audit).RuleInSubquery,
		//},
		"SUB.002": {
			Item:     "SUB.002",
			Severity: "L2",
			Summary:  "If you don't care about duplication, it is recommended to use UNION ALL instead of UNION",
			Content:  `Unlike UNION which removes duplicates, UNION ALL allows duplicate tuples. If you don't care about duplicate tuples, using UNION ALL would be a faster option. `,
			Case:     "select teacher_id as id,people_name as name from t1,t2 where t1.teacher_id=t2.people_id union select student_id as id,people_name as name from t1,t2 where t1.student_id=t2.people_id",
			Func:     (*Query4Audit).RuleUNIONUsage,
		},
		"SUB.003": {
			Item:     "SUB.003",
			Severity: "L3",
			Summary:  "Consider using EXISTS instead of DISTINCT subquery",
			Content:  `DISTINCT keyword removes duplicates after sorting tuples. Instead, consider using a subquery with the EXISTS keyword, where you can avoid returning the entire table. `,
			Case:     "SELECT DISTINCT c.c_id, c.c_name FROM c,e WHERE e.c_id = c.c_id",
			Func:     (*Query4Audit).RuleDistinctJoinUsage,
		},
		// TODO: 5.6有了semi join 还要把 in 转成 exists 么？
		// Use EXISTS instead of IN to check existence of data.
		// http://www.winwire.com/25-tips-to-improve-sql-query-performance/
		"SUB.004": {
			Item:     "SUB.004",
			Severity: "L3",
			Summary:  "The depth of nested connections in the execution plan is too deep",
			Content:  `MySQL does not optimize subqueries well. MySQL uses each row in the external query as a dependent subquery to execute the subquery. This is a common cause of serious performance issues. `,
			Case:     "SELECT * from tb where id in (select id from (select id from tb))",
			Func:     (*Query4Audit).RuleSubqueryDepth,
		},
		// SUB.005灵感来自 https://blog.csdn.net/zhuocr/article/details/61192418
		"SUB.005": {
			Item:     "SUB.005",
			Severity: "L8",
			Summary:  "LIMIT is not supported for subqueries",
			Content:  "The current MySQL version does not support 'LIMIT & IN/ALL/ANY/SOME' in a subquery.",
			Case:     "SELECT * FROM staff WHERE name IN (SELECT NAME FROM customer ORDER BY name LIMIT 1)",
			Func:     (*Query4Audit).RuleSubQueryLimit,
		},
		"SUB.006": {
			Item:     "SUB.006",
			Severity: "L2",
			Summary:  "Using functions in subqueries is not recommended",
			Content:  `MySQL executes the subquery with each row in the outer query as a dependent subquery. If functions are used in the subquery, even semi-join will make it difficult to perform efficient queries. You can rewrite the subquery as an OUTER JOIN statement and use join conditions to filter the data. `,
			Case:     "SELECT * FROM staff WHERE name IN (SELECT max(NAME) FROM customer)",
			Func:     (*Query4Audit).RuleSubQueryFunctions,
		},
		"SUB.007": {
			Item:     "SUB.007",
			Severity: "L2",
			Summary:  "The outer UNION query with LIMIT output limit, the inner query is recommended to also add LIMIT output limit",
			Content:  `Sometimes MySQL cannot "push down" the restriction conditions from the outer layer to the inner layer, which makes it impossible to apply the conditions that can restrict some of the returned results to the optimization of the inner layer query. For example: (SELECT * FROM tb1 ORDER BY name) UNION ALL (SELECT * FROM tb2 ORDER BY name) LIMIT 20; MySQL will put the results of the two subqueries in a temporary table, and then retrieve 20 results. You can use Add LIMIT 20 to the subquery to reduce the data in the temporary table. (SELECT * FROM tb1 ORDER BY name LIMIT 20) UNION ALL (SELECT * FROM tb2 ORDER BY name LIMIT 20) LIMIT 20;`,
			Case:     "(SELECT * FROM tb1 ORDER BY name LIMIT 20) UNION ALL (SELECT * FROM tb2 ORDER BY name LIMIT 20) LIMIT 20;",
			Func:     (*Query4Audit).RuleUNIONLimit,
		},
		"TBL.001": {
			Item:     "TBL.001",
			Severity: "L4",
			Summary:  "Partitioned tables are not recommended",
			Content:  `Partition table is not recommended`,
			Case:     "CREATE TABLE trb3(id INT, name VARCHAR(50), purchased DATE) PARTITION BY RANGE(YEAR(purchased)) (PARTITION p0 VALUES LESS THAN (1990), PARTITION p1 VALUES LESS THAN (1995), PARTITION p2 VALUES LESS THAN (2000), PARTITION p3 VALUES LESS THAN (2005) );",
			Func:     (*Query4Audit).RulePartitionNotAllowed,
		},
		"TBL.002": {
			Item:     "TBL.002",
			Severity: "L4",
			Summary:  "Please select an appropriate storage engine for the table",
			Content:  `When creating a table or modifying the storage engine of a table, it is recommended to use the recommended storage engine, such as: ` + strings.Join(common.Config.AllowEngines, ","),
			Case:     "create table test(`id` int(11) NOT NULL AUTO_INCREMENT)",
			Func:     (*Query4Audit).RuleAllowEngine,
		},
		"TBL.003": {
			Item:     "TBL.003",
			Severity: "L8",
			Summary:  "Tables named with DUAL have special meaning in the database",
			Content:  `The DUAL table is a virtual table and does not need to be created before it can be used. It is not recommended that the service names the table after DUAL. `,
			Case:     "create table dual(id int, primary key (id));",
			Func:     (*Query4Audit).RuleCreateDualTable,
		},
		"TBL.004": {
			Item:     "TBL.004",
			Severity: "L2",
			Summary:  "The initial AUTO_INCREMENT value of the table is not 0",
			Content:  `AUTO_INCREMENT is not 0, which will cause data holes. `,
			Case:     "CREATE TABLE tbl (a int) AUTO_INCREMENT = 10;",
			Func:     (*Query4Audit).RuleAutoIncrementInitNotZero,
		},
		"TBL.005": {
			Item:     "TBL.005",
			Severity: "L4",
			Summary:  "Please use the recommended character set",
			Content:  `Table character set is only allowed to be set to' ` + strings.Join(common.Config.AllowCharsets, ",") + "'",
			Case:     "CREATE TABLE tbl (a int) DEFAULT CHARSET = latin1;",
			Func:     (*Query4Audit).RuleTableCharsetCheck,
		},
		"TBL.006": {
			Item:     "TBL.006",
			Severity: "L1",
			Summary:  "Views are not recommended",
			Content:  `View is not recommended`,
			Case:     "create view v_today (today) AS SELECT CURRENT_DATE;",
			Func:     (*Query4Audit).RuleForbiddenView,
		},
		"TBL.007": {
			Item:     "TBL.007",
			Severity: "L1",
			Summary:  "The use of temporary tables is not recommended",
			Content:  `The use of temporary tables is not recommended`,
			Case:     "CREATE TEMPORARY TABLE `work` (`time` time DEFAULT NULL) ENGINE=InnoDB;",
			Func:     (*Query4Audit).RuleForbiddenTempTable,
		},
		"TBL.008": {
			Item:     "TBL.008",
			Severity: "L4",
			Summary:  "Please use the recommended COLLATE",
			Content:  `COLLATE is only allowed to be set to'` + strings.Join(common.Config.AllowCollates, ",") + "'",
			Case:     "CREATE TABLE tbl (a int) DEFAULT COLLATE = latin1_bin;",
			Func:     (*Query4Audit).RuleTableCharsetCheck,
		},
	}
}

// 入参：audit要审计的SQL以及解析AST key为规则ID（item）value是规则对象rule内容（rule结构体）
func CheckHeuristicRules(audit *Query4Audit) (rules map[string]Rule) {
	rules = make(map[string]Rule, 0) //
	okFunc := (*Query4Audit).RuleOK  //返回‘OK’的规则

	for item, rule := range HeuristicRules {
		// item不是要忽略的 并且和ok不一样
		if !IsIgnoreRule(item) && fmt.Sprintf("%v", rule.Func) != fmt.Sprintf("%v", okFunc) {
			//for循环 对于每一个sql 循环检查 逐个对比来看sql有没有对应规则的问题 默认都是返回OK的 有问题了才返回自己当前的规则
			r := rule.Func(audit) //这里返回一个规则 如果是ok没违反当前规则 违反了返回自己当前的
			if r.Item == item {   //如果确实违反了 那么就加入到当前sql违反的规则列表之中
				rules[item] = r
			}
		}
	}
	//if len(rules) == 0 {
	//	ruleOK := okFunc(audit)
	//	rules[ruleOK.Item] = ruleOK
	//}
	return rules
}

// IsIgnoreRule 判断是否是过滤规则
// 支持XXX*前缀匹配，OK规则不可设置过滤
func IsIgnoreRule(item string) bool {

	for _, ir := range common.Config.IgnoreRules {
		ir = strings.Trim(ir, "*")
		if strings.HasPrefix(item, ir) && ir != "OK" && ir != "" {
			common.Log.Debug("IsIgnoreRule: %s", item)
			return true
		}
	}
	return false
}

// InBlackList 判断一条请求是否在黑名单列表中
// 如果在返回true，表示不需要评审
// 注意这里没有做指纹判断，是否用指纹在这个函数的外面处理
func InBlackList(sql string) bool {
	in := false
	for _, r := range common.BlackList {
		if sql == r {
			in = true
			break
		}
		re, err := regexp.Compile("(?i)" + r)
		if err == nil {
			if re.FindString(sql) != "" {
				common.Log.Debug("InBlackList: true, regexp: %s, sql: %s", "(?i)"+r, sql)
				in = true
				break
			}
			common.Log.Debug("InBlackList: false, regexp: %s, sql: %s", "(?i)"+r, sql)
		}
	}
	return in
}

// FormatSuggest 格式化输出优化建议
func FormatSuggest(sql string, currentDB string, format string, suggests ...map[string]Rule) (map[string]Rule, string) {
	common.Log.Debug("FormatSuggest, Query: %s", sql)
	var fingerprint, id string
	var buf []string
	var score = 100
	type Result struct {
		ID          string
		Fingerprint string
		Sample      string
		Suggest     map[string]Rule
	}

	// 生成指纹和ID
	if sql != "" {
		fingerprint = query.Fingerprint(sql)
		id = query.Id(fingerprint)
	}

	// 合并重复的建议
	suggest := make(map[string]Rule)
	for _, s := range suggests {
		for item, rule := range s {
			suggest[item] = rule
		}
	}
	suggest = MergeConflictHeuristicRules(suggest)

	// 是否忽略显示OK建议，测试的时候大家都喜欢看OK，线上跑起来的时候OK太多反而容易看花眼
	ignoreOK := false
	for _, r := range common.Config.IgnoreRules {
		if "OK" == r {
			ignoreOK = true
		}
	}

	// 先保证suggest中有元素，然后再根据ignore配置删除不需要的项
	if len(suggest) < 1 {
		suggest = map[string]Rule{"OK": HeuristicRules["OK"]}
	}
	if ignoreOK || len(suggest) > 1 {
		delete(suggest, "OK")
	}
	for k := range suggest {
		if IsIgnoreRule(k) {
			delete(suggest, k)
		}
	}
	common.Log.Debug("FormatSuggest, format: %s", format)
	switch format {
	case "json":
		buf = append(buf, formatJSON(sql, currentDB, suggest))

	case "text":
		for item, rule := range suggest {
			buf = append(buf, fmt.Sprintln("Query: ", sql))
			buf = append(buf, fmt.Sprintln("ID: ", id))
			buf = append(buf, fmt.Sprintln("Item: ", item))
			buf = append(buf, fmt.Sprintln("Severity: ", rule.Severity))
			buf = append(buf, fmt.Sprintln("Summary: ", rule.Summary))
			buf = append(buf, fmt.Sprintln("Content: ", rule.Content))
		}
	case "lint":
		for item, rule := range suggest {
			// lint 中无需关注 OK 和 EXP
			if item != "OK" && !strings.HasPrefix(item, "EXP") {
				buf = append(buf, fmt.Sprintf("%s %s", item, rule.Summary))
			}
		}

	case "markdown", "html", "explain-digest", "duplicate-key-checker":
		if sql != "" && len(suggest) > 0 {
			switch common.Config.ExplainSQLReportType {
			case "fingerprint":
				buf = append(buf, fmt.Sprintf("# Query: %s\n", id))
				buf = append(buf, fmt.Sprintf("```sql\n%s\n```\n", fingerprint))
			case "sample":
				buf = append(buf, fmt.Sprintf("# Query: %s\n", id))
				buf = append(buf, fmt.Sprintf("```sql\n%s\n```\n", sql))
			default:
				buf = append(buf, fmt.Sprintf("# Query: %s\n", id))
				buf = append(buf, fmt.Sprintf("```sql\n%s\n```\n", ast.Pretty(sql, format)))
			}
		}
		// MySQL
		common.Log.Debug("FormatSuggest, start of sortedMySQLSuggest")
		var sortedMySQLSuggest []string
		for item := range suggest {
			if strings.HasPrefix(item, "ERR") {
				if suggest[item].Content == "" {
					delete(suggest, item)
				} else {
					sortedMySQLSuggest = append(sortedMySQLSuggest, item)
				}
			}
		}
		sort.Strings(sortedMySQLSuggest)
		if len(sortedMySQLSuggest) > 0 {
			buf = append(buf, "## MySQL execute failed\n")
		}
		for _, item := range sortedMySQLSuggest {
			buf = append(buf, fmt.Sprintln(suggest[item].Content))
			score = 0
			delete(suggest, item)
		}

		// Explain
		common.Log.Debug("FormatSuggest, start of sortedExplainSuggest")
		if suggest["EXP.000"].Item != "" {
			buf = append(buf, fmt.Sprintln("## ", suggest["EXP.000"].Summary))
			buf = append(buf, fmt.Sprintln(suggest["EXP.000"].Content))
			buf = append(buf, fmt.Sprint(suggest["EXP.000"].Case, "\n"))
			delete(suggest, "EXP.000")
		}
		var sortedExplainSuggest []string
		for item := range suggest {
			if strings.HasPrefix(item, "EXP") {
				sortedExplainSuggest = append(sortedExplainSuggest, item)
			}
		}
		sort.Strings(sortedExplainSuggest)
		for _, item := range sortedExplainSuggest {
			buf = append(buf, fmt.Sprintln("### ", suggest[item].Summary))
			buf = append(buf, fmt.Sprintln(suggest[item].Content))
			buf = append(buf, fmt.Sprint(suggest[item].Case, "\n"))
		}

		// Profiling
		common.Log.Debug("FormatSuggest, start of sortedProfilingSuggest")
		var sortedProfilingSuggest []string
		for item := range suggest {
			if strings.HasPrefix(item, "PRO") {
				sortedProfilingSuggest = append(sortedProfilingSuggest, item)
			}
		}
		sort.Strings(sortedProfilingSuggest)
		if len(sortedProfilingSuggest) > 0 {
			buf = append(buf, "## Profiling信息\n")
		}
		for _, item := range sortedProfilingSuggest {
			buf = append(buf, fmt.Sprintln(suggest[item].Content))
			delete(suggest, item)
		}

		// Trace
		common.Log.Debug("FormatSuggest, start of sortedTraceSuggest")
		var sortedTraceSuggest []string
		for item := range suggest {
			if strings.HasPrefix(item, "TRA") {
				sortedTraceSuggest = append(sortedTraceSuggest, item)
			}
		}
		sort.Strings(sortedTraceSuggest)
		if len(sortedTraceSuggest) > 0 {
			buf = append(buf, "## Trace信息\n")
		}
		for _, item := range sortedTraceSuggest {
			buf = append(buf, fmt.Sprintln(suggest[item].Content))
			delete(suggest, item)
		}

		// Optimize
		common.Log.Debug("FormatSuggest, start of sortedIdxSuggest")
		var sortedIdxSuggest []string
		for item := range suggest {
			if strings.HasPrefix(item, "IDX") {
				sortedIdxSuggest = append(sortedIdxSuggest, item)
			}
		}
		sort.Strings(sortedIdxSuggest)
		for _, item := range sortedIdxSuggest {
			buf = append(buf, fmt.Sprintln("## ", common.MarkdownEscape(suggest[item].Summary)))
			buf = append(buf, fmt.Sprintln("* **Item:** ", item))
			buf = append(buf, fmt.Sprintln("* **Severity:** ", suggest[item].Severity))
			minus, err := strconv.Atoi(strings.Trim(suggest[item].Severity, "L"))
			if err == nil {
				score = score - minus*5
			} else {
				common.Log.Debug("FormatSuggest, sortedIdxSuggest, strconv.Atoi, Error: ", err)
				score = 0
			}
			buf = append(buf, fmt.Sprintln("* **Content:** ", common.MarkdownEscape(suggest[item].Content)))

			if format == "duplicate-key-checker" {
				buf = append(buf, fmt.Sprintf("* **原建表语句:** \n```sql\n%s\n```\n", suggest[item].Case), "\n\n")
			} else {
				buf = append(buf, fmt.Sprint("* **Case:** ", common.MarkdownEscape(suggest[item].Case), "\n\n"))
			}
		}

		// Heuristic
		common.Log.Debug("FormatSuggest, start of sortedHeuristicSuggest")
		var sortedHeuristicSuggest []string
		for item := range suggest {
			if !strings.HasPrefix(item, "EXP") &&
				!strings.HasPrefix(item, "IDX") &&
				!strings.HasPrefix(item, "PRO") {
				sortedHeuristicSuggest = append(sortedHeuristicSuggest, item)
			}
		}
		sort.Strings(sortedHeuristicSuggest)
		for _, item := range sortedHeuristicSuggest {
			buf = append(buf, fmt.Sprintln("##", suggest[item].Summary))
			if item == "OK" {
				continue
			}
			buf = append(buf, fmt.Sprintln("* **Item:** ", item))
			buf = append(buf, fmt.Sprintln("* **Severity:** ", suggest[item].Severity))
			minus, err := strconv.Atoi(strings.Trim(suggest[item].Severity, "L"))
			if err == nil {
				score = score - minus*5
			} else {
				common.Log.Debug("FormatSuggest, sortedHeuristicSuggest, strconv.Atoi, Error: ", err)
				score = 0
			}
			buf = append(buf, fmt.Sprintln("* **Content:** ", common.MarkdownEscape(suggest[item].Content)))
			// buf = append(buf, fmt.Sprint("* **Case:** ", common.MarkdownEscape(suggest[item].Case), "\n\n"))
		}

	default:
		common.Log.Debug("report-type: %s", format)
		buf = append(buf, fmt.Sprintln("Query: ", sql))
		for _, rule := range suggest {
			buf = append(buf, pretty.Sprint(rule))
		}
	}

	// 打分
	var str string
	switch common.Config.ReportType {
	case "markdown", "html":
		if len(buf) > 1 {
			str = buf[0] + "\n" + common.Score(score) + "\n\n" + strings.Join(buf[1:], "\n")
		}
	default:
		str = strings.Join(buf, "\n")
	}

	return suggest, str
}

// JSONSuggest json format suggestion
type JSONSuggest struct {
	ID             string   `json:"ID"`
	Fingerprint    string   `json:"Fingerprint"`
	Score          int      `json:"Score"`
	Sample         string   `json:"Sample"`
	Explain        []Rule   `json:"Explain"`
	HeuristicRules []Rule   `json:"HeuristicRules"`
	IndexRules     []Rule   `json:"IndexRules"`
	Tables         []string `json:"Tables"`
}

func formatJSON(sql string, db string, suggest map[string]Rule) string {
	var id, fingerprint, result string

	fingerprint = query.Fingerprint(sql)
	id = query.Id(fingerprint)

	// Score
	score := 100
	for item := range suggest {
		l, err := strconv.Atoi(strings.TrimLeft(suggest[item].Severity, "L"))
		if err != nil {
			common.Log.Error("formatJSON strconv.Atoi error: %s, item: %s, serverity: %s", err.Error(), item, suggest[item].Severity)
		}
		score = score - l*5
		// ## MySQL execute failed
		if strings.HasPrefix(item, "ERR") && suggest[item].Content != "" {
			score = 0
		}
	}
	if score < 0 {
		score = 0
	}

	sug := JSONSuggest{
		ID:          id,
		Fingerprint: fingerprint,
		Sample:      sql,
		Tables:      ast.SchemaMetaInfo(sql, db),
		Score:       score,
	}

	// Explain info
	var sortItem []string
	for item := range suggest {
		if strings.HasPrefix(item, "EXP") {
			sortItem = append(sortItem, item)
		}
	}
	sort.Strings(sortItem)
	for _, i := range sortItem {
		sug.Explain = append(sug.Explain, suggest[i])
	}
	sortItem = make([]string, 0)

	// Optimize advisor
	for item := range suggest {
		if strings.HasPrefix(item, "IDX") {
			sortItem = append(sortItem, item)
		}
	}
	sort.Strings(sortItem)
	for _, i := range sortItem {
		sug.IndexRules = append(sug.IndexRules, suggest[i])
	}
	sortItem = make([]string, 0)

	// Heuristic rules
	for item := range suggest {
		if !strings.HasPrefix(item, "EXP") && !strings.HasPrefix(item, "IDX") {
			if strings.HasPrefix(item, "ERR") && suggest[item].Content == "" {
				continue
			}
			sortItem = append(sortItem, item)
		}
	}
	sort.Strings(sortItem)
	for _, i := range sortItem {
		sug.HeuristicRules = append(sug.HeuristicRules, suggest[i])
	}
	sortItem = make([]string, 0)

	js, err := json.MarshalIndent(sug, "", "  ")
	if err == nil {
		result = fmt.Sprint(string(js))
	} else {
		common.Log.Error("formatJSON json.Marshal Error: %v", err)
	}
	return result
}

// ListHeuristicRules 打印支持的启发式规则，对应命令行参数-list-heuristic-rules
func ListHeuristicRules(rules ...map[string]Rule) {
	switch common.Config.ReportType {
	case "json":
		js, err := json.MarshalIndent(rules, "", "  ")
		if err == nil {
			fmt.Println(string(js))
		}
	default:
		fmt.Print("# 启发式规则建议\n\n[toc]\n\n")
		for _, r := range rules {
			delete(r, "OK")
			for _, item := range common.SortedKey(r) {
				fmt.Print("## ", common.MarkdownEscape(r[item].Summary),
					"\n\n* **Item**:", r[item].Item,
					"\n* **Severity**:", r[item].Severity,
					"\n* **Content**:", common.MarkdownEscape(r[item].Content),
					"\n* **Case**:\n\n```sql\n", r[item].Case, "\n```\n")
			}
		}
	}
}

// ListTestSQLs 打印测试用的SQL，方便测试，对应命令行参数-list-test-sqls
func ListTestSQLs() {
	for _, sql := range common.TestSQLs {
		fmt.Println(sql)
	}
}
