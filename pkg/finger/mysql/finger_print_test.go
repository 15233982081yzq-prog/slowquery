package mysql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	fp     *FingerPrint
	fpHint *FingerPrint

	tests = []struct {
		input  string
		expect string
	}{
		{"SELECT `id`, `col_name1`, `col_name2`, `col_name3`, `col_name4`, `col_name5`, `col_name6` FROM `dml_randgen_table_3` LIMIT 0,1000;", "select `id` , `col_name1` , `col_name2` , `col_name3` , `col_name4` , `col_name5` , `col_name6` from `dml_randgen_table_3` limit ... ;"},
		{"select _utf8mb4'123'", "select (_charset) ?"},
		{"SELECT 1", "select ?"},
		{"select null", "select ?"},
		{"select \\N", "select ?"},
		{"SELECT `null`", "select `null`"},
		{"select * from user where name = 'lisi'", "select * from user where `name` = ?"},
		{"select * from b where id = 1", "select * from `b` where `id` = ?"},
		{"select 1 from b where id in (1, 3, '3', 1, 2, 3, 4)", "select ? from `b` where `id` in ( ... )"},
		{"select 1 from b where id in (1, a, 4)", "select ? from `b` where `id` in ( ? , `a` , ? )"},
		{"select 1 from b order by 2", "select ? from `b` order by 2"},
		{"select /*+ a hint */ 1", "select ?"},
		{"select /* a hint */ 1", "select ?"},
		{"select truncate(1, 2)", "select truncate ( ... )"},
		{"select from a ", "select from `a`"},
		{"select -1 + - 2 + b - c + 0.2 + (-2) from c where d in (1, -2, +3)", "select ? + ? + `b` - `c` + ? + ( ? ) from `c` where `d` in ( ... )"},
		{"select * from t where a <= -1 and b < -2 and c = -3 and c > -4 and c >= -5 and e is 1", "select * from `t` where `a` <= ? and `b` < ? and `c` = ? and `c` > ? and `c` >= ? and `e` is ?"},
		{"select count(a), b from t group by 2", "select count ( `a` ) , `b` from `t` group by 2"},
		{"select count(a), b, c from t group by 2, 3", "select count ( `a` ) , `b` , `c` from `t` group by 2 , 3"},
		{"select count(a), b, c from t group by (2, 3)", "select count ( `a` ) , `b` , `c` from `t` group by ( 2 , 3 )"},
		{"select a, b from t order by 1, 2", "select `a` , `b` from `t` order by 1 , 2"},
		{"select count(*) from t", "select count ( ? ) from `t`"},
		{"select * from t Force Index(kk)", "select * from `t`"},
		{"select * from t USE Index(kk)", "select * from `t`"},
		{"select * from t Ignore Index(kk)", "select * from `t`"},
		{"select * from t1 straight_join t2 on t1.id=t2.id", "select * from `t1` join `t2` on `t1` . `id` = `t2` . `id`"},
		{"select * from `table`", "select * from `table`"},
		{"select * from `30`", "select * from `30`"},
		{"select * from `select`", "select * from `select`"},
		{"select * from              `select`", "select * from `select`"},
		{"select *                                   from              `select`", "select * from `select`"},
		{"select 				*                                   from              `select`", "select * from `select`"},
		{"select * from 🥳", "select * from `🥳`"},
		// test syntax error, it will be checked by parser, but it should not make normalize dead loop.
		{"select * from t ignore index(", "select * from `t` ignore index"},
		{"select * from User where name = '#' and age = '#'", "select * from user where `name` = ? and `age` = ?"},
		{"select * from User where name = # and age = #", "select * from user where `name` ="},                                                                              //TODO 对#号支持有问题
		{"select *\nfrom User\nwhere 1 = 1\n{? and name = #name# ;nullable:true}\n{? and age = #age# }", "select * from user where ? = ? { ? and `name` = { ? and `age` ="}, //TODO 有问题
		{"select /*+ ", "select "},
		{"select 1 / 2", "select ? / ?"},
		{"select * from t where a = 40 limit ?, ?", "select * from `t` where `a` = ? limit ..."},
		{"select * from t where a > ?", "select * from `t` where `a` > ?"},
		{"select * from t where a between 10 and 30", "select * from `t` where `a` between ? and ?"},
		{"seLect * fRom t wHerE a bEtWeen 10 And 30", "select * from `t` where `a` between ? and ?"},
		{"select @a=b from t", "select @a = `b` from `t`"},
		{"select * from `table", "select * from"},
		{"Select * from t where (i, j) in ((1,1), (2,2))", "select * from `t` where ( `i` , `j` ) in ( ( ... ) )"},
		{"insert into t values (1,1), (2,2)", "insert into `t` values ( ... )"},
		{"insert into t values (1), (2)", "insert into `t` values ( ... )"},
		{"insert into t values (1)", "insert into `t` values ( ? )"},
		{"insert into t (column) values(1)", "insert into `t` ( column ) values ( ? )"},
		{"INSERT INTO t ValuEs (1,1), (2,2)", "insert into `t` values ( ... )"},
		{"INSERT INTO t VALUES (1), (2)", "insert into `t` values ( ... )"},
		{"INSERT into t VALUES (1)", "insert into `t` values ( ? )"},
		{"insert INTO t (column) values(1)", "insert into `t` ( column ) values ( ? )"},
		{"insert 			INTO 			t (column) 				values(1)", "insert into `t` ( column ) values ( ? )"},
		{"delete 				from 				t 						where id = 1", "delete from `t` where `id` = ?"},
		{"delete from t where id = 1", "delete from `t` where `id` = ?"},
		{"delete from t where id > 10", "delete from `t` where `id` > ?"},
		{"DELETE frOM t where id between 10 and 100", "delete from `t` where `id` between ? and ?"},
		{"delete from sbtest2 where sbtest2.id not in (select id from sbtest1);", "delete from `sbtest2` where `sbtest2` . `id` not in ( select `id` from `sbtest1` ) ;"},
		{"UpdaTe 				t sEt j = 2 				wHerE id 		between ? and ?", "update `t` set `j` = ? where `id` between ? and ?"},
		{"UpdaTe t sEt j = 2 wHerE id between ? and ?", "update `t` set `j` = ? where `id` between ? and ?"},
		{"update t set j = 2 where id = 1", "update `t` set `j` = ? where `id` = ?"},
		{"update t set j = 2 where id > 1", "update `t` set `j` = ? where `id` > ?"},
		{"UPDATE sbtest1 set k=103 , cnt=2 , col_1 = 'col'  where id >0;", "update `sbtest1` set `k` = ? , `cnt` = ? , `col_1` = ? where `id` > ? ;"},
		{"update sbtest1,(select id,c from sbtest2 where sbtest2.id>2) as sbtest2New set sbtest1.c = sbtest2New.c where sbtest1.id = sbtest2New.id;", "update `sbtest1` , ( select `id` , `c` from `sbtest2` where `sbtest2` . `id` > ? ) as `sbtest2new` set `sbtest1` . `c` = `sbtest2new` . `c` where `sbtest1` . `id` = `sbtest2new` . `id` ;"},
		{"UPDATE automation_test_samll_table SET col_name2 = 11 where col_name3 != col_name2*col_name1;", "update `automation_test_samll_table` set `col_name2` = ? where `col_name3` != `col_name2` * `col_name1` ;"},
		{"update sbtest1 set c=101 where id >0 limit 1;", "update `sbtest1` set `c` = ? where `id` > ? limit ? ;"},
		{"update SBTEST1  \n    set c=101              where id > 0 limit 10;", "update `sbtest1` set `c` = ? where `id` > ? limit ? ;"},
		{"update sbtest1 /* a hint */ ;", "update `sbtest1` ;"},
		{"update sbtest1 -- a hint * 9ejsyb;", "update `sbtest1`"},
	}

	testsKeepHint = []struct {
		input  string
		expect string
	}{
		{"update sbtest1 /* a hint */ ;", "update `sbtest1` ;"},
		{"update sbtest1 -- a hint * 9ejsyb;", "update `sbtest1`"},
		{"select /* sasdyebapsb */;", "select ;"},
		{"select -- aksieb *shsa;", "select"},
		{"SELect /*+ ", "select "},
		{"select /*+ trace dea06b7bffddeed73be864f674eb7600:0000006bdc2b36ca:0000000000000000 */ 1", "select /*+ trace dea06b7bffddeed73be864f674eb7600:0000006bdc2b36ca:0000000000000000 */ ?"},
		{"insert /*+ trace dea06b7bffddeed73be864f674eb7600:0000006bdc2b36ca:0000000000000000 */ into t values (1)", "insert /*+ trace dea06b7bffddeed73be864f674eb7600:0000006bdc2b36ca:0000000000000000 */ into `t` values ( ? )"},
		{"update /*+ trace dea06b7bffddeed73be864f674eb7600:0000006bdc2b36ca:0000000000000000 */ t set j = 2 where id = 1", "update /*+ trace dea06b7bffddeed73be864f674eb7600:0000006bdc2b36ca:0000000000000000 */ `t` set `j` = ? where `id` = ?"},
		{"delete /*+ trace dea06b7bffddeed73be864f674eb7600:0000006bdc2b36ca:0000000000000000 */ from t where id = 1", "delete /*+ trace dea06b7bffddeed73be864f674eb7600:0000006bdc2b36ca:0000000000000000 */ from `t` where `id` = ?"},
	}
)

func TestNewFingerPrint(t *testing.T) {
	assert.NotNil(t, fp)
	assert.NotNil(t, fpHint)
}

func TestFingerPrint_Normalize(t *testing.T) {
	for _, test := range tests {
		assert.Equal(t, fp.Normalize(test.input), test.expect)
	}
}

func TestNewFingerPrint_Normalize_ByKeepHint(t *testing.T) {
	for _, test := range testsKeepHint {
		assert.Equal(t, fpHint.Normalize(test.input), test.expect)
		fmt.Printf(" originsql:%s \n normalize:%s \n expect:   %s \n", test.input, fpHint.Normalize(test.input), test.expect)
	}
}

func TestFingerPrint_Digest(t *testing.T) {
	for _, test := range tests {
		assert.Equal(t, fp.Digest(test.input).String(), fp.DigestNormalize(test.expect).String())
		fmt.Printf("input:%s,expect:%s \n", fp.Digest(test.input).String(), fp.DigestNormalize(test.expect).String())
		fmt.Printf(" originsql:%s \n normalize:%s \n expect:   %s \n", test.input, fp.Normalize(test.input), test.expect)
	}
}

func TestFingerPrint_NormalizeDigest(t *testing.T) {
	for _, test := range tests {
		normalize, digest := fp.NormalizeDigest(test.input)
		assert.Equal(t, fp.Normalize(test.input), normalize)
		assert.Equal(t, normalize, test.expect)

		assert.Equal(t, digest.String(), fp.DigestNormalize(normalize).String())
		assert.Equal(t, digest.String(), fp.Digest(test.input).String())
		assert.Equal(t, fp.Digest(test.input).String(), fp.DigestNormalize(test.expect).String())
	}
}

func init() {
	fp = NewFingerPrint(false)
	fpHint = NewFingerPrint(true)
}
