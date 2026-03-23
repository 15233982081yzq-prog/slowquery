package optimize

type example struct {
	sql                      string
	disc                     string
	withIndexSuggestCount    int
	notWithIndexSuggestCount int
	noErr                    bool
}

/*
table:   table_1k  table_1w
1. 表结构信息
CREATE TABLE `table_xxx` (
	`id` bigint(20) NOT NULL AUTO_INCREMENT,
	`user_name` longtext,
	`db_name` longtext,
	`db_env` longtext,
	`age` bigint(20) DEFAULT NULL,
	`info` text,
	`create_time` datetime(3) DEFAULT NULL,
	`update_time` datetime(3) DEFAULT NULL,
	PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1
2. 测试 索引场景下 添加的索引
alter table table_1k add index `idx_test`(`user_name`,`age`);
*/

var testExampleSingleTable = []example{
	{
		sql:                      `select *  from %s where user_name="slow" and age = 12`,
		disc:                     "where等值组合查询",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s where db_env='live' and user_name="slow" and age = 12`,
		disc:                     "where等值组合查询",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s where age > 12 and create_time > '2023-12-12 12:12:12'`,
		disc:                     "where范围组合查询",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s where id = 1001 and age > 12`,
		disc:                     "where等值+范围组合查询",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s where user_name not in ('aa', 'bb', 'cc')`,
		disc:                     "where中存在not in 条件",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s where user_name in ('aa', 'bb', 'cc')`,
		disc:                     "where中存在in 条件",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s where user_name in ('aa', 'bb', 'cc') and age = 18`,
		disc:                     "where中存在in条件 + 其他范围条件",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select create_time, count(*)  from %s where age > 18 group by create_time`,
		disc:                     "带有count(*)的查询",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select env, count(*)  from %s group by db_env having db_env != 'aaa'`,
		disc:                     "group+having条件出现",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select env, count(*)  from %s where age > 12 group by db_env having db_env != 'aaa'`,
		disc:                     "group+having条件出现 + 其他where条件",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select * from user %s where db_env !=null`,
		disc:                     "where条件中存在列=NULL 获!=NULL",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s group by create_time`,
		disc:                     "没有where条件的group by查询",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s order by create_time desc`,
		disc:                     "没有where条件的order by查询",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s group by db_env order by create_time desc`,
		disc:                     "有where条件的order by+ group by查询",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s where age+1<18`,
		disc:                     "存在计算条件的查询", // 存在计算或者函数，不能进行索引推荐
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s where month(create_time)=7 limit 1`,
		disc:                     "存在函数计算的查询",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s where info regexp '^[a-z]+$'`,
		disc:                     "使用正则表达式",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select *  from %s where db_env=info`,
		disc:                     "判断两列相等情况",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select db_env,(case when age=18 then 22 else 0 end) as remark  from %s`,
		disc:                     "使用case查询",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
}

var testExampleMultipleTables = []example{
	{
		sql:                      `select a.id from %s as a inner join %s as b on a.db_env=b.db_env`,
		disc:                     "双表：单个连接条件inner联合查询",
		notWithIndexSuggestCount: 2,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select a.id from %s as a left join %s as b on a.db_env=b.db_env`,
		disc:                     "双表：单个连接条件left联合查询",
		notWithIndexSuggestCount: 2,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select a.id from %s as a right join %s as b on a.db_env=b.db_env`,
		disc:                     "双表：单个连接条件right联合查询",
		notWithIndexSuggestCount: 2,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select  * from %s a inner join %s b on a.db_env=b.db_env and a.age=b.age`,
		disc:                     "双表：多个连接条件inner联合查询",
		notWithIndexSuggestCount: 2,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select  * from %s a left join %s b on a.db_env=b.db_env and a.age=b.age`,
		disc:                     "双表：多个连接条件left联合查询",
		notWithIndexSuggestCount: 2,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select  * from %s a right join %s b on a.db_env=b.db_env and a.age=b.age`,
		disc:                     "双表：多个连接条件right联合查询",
		notWithIndexSuggestCount: 2,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select  * from %s a inner join %s b on a.db_env=b.db_env and a.age=b.age where a.create_time > '2021-12-30 12:12:12'`,
		disc:                     "双表：多个inner连接条件+where条件",
		notWithIndexSuggestCount: 2,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select  * from %s a left join %s b on a.db_env=b.db_env and a.age=b.age where a.create_time > '2021-12-30 12:12:12'`,
		disc:                     "双表：多个left连接条件+where条件",
		notWithIndexSuggestCount: 2,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select  * from %s a right join %s b on a.db_env=b.db_env and a.age=b.age where a.create_time > '2021-12-30 12:12:12'`,
		disc:                     "双表：多个inner连接条件+where条件",
		notWithIndexSuggestCount: 2,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select * from %s a inner join %s b on a.db_env=b.db_env WHERE b.age in (11, 21,31)`,
		disc:                     "双表：存在判断in的where条件",
		notWithIndexSuggestCount: 2,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select * from %s a inner join %s b on a.db_env=b.db_env WHERE b.age not in (11, 21,31)`,
		disc:                     "双表：存在判断not in的where条件",
		notWithIndexSuggestCount: 1, // 索引优化被过滤了，因为存在not in
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select * from %s a left join %s b on a.db_env=b.db_env union select * from %s a right join %s b on a.db_env=b.db_env`,
		disc:                     "多表：存在union条件",
		notWithIndexSuggestCount: 0,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `SELECT a. db_env, a. create_time FROM %s a STRAIGHT_JOIN %s b ON a.db_env= b.db_env`,
		disc:                     "双表：STRAIGHT_JOIN",
		notWithIndexSuggestCount: 2,
		withIndexSuggestCount:    1,
	},
	{
		sql:                      `select * from %s where age in (select age from %s where db_env > 'live')`,
		disc:                     "双表：嵌套子查询",
		notWithIndexSuggestCount: 1,
		withIndexSuggestCount:    1,
	},
}
