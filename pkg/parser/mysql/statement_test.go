package mysql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	sel     = "SELECT * FROM table_name WHERE column_name = 42"
	ins     = "INSERT INTO tableName SELECT * FROM t WHERE T1='a'"
	upd     = "UPDATE table_name SET column1 = value1, column2 = value2 WHERE condition;"
	del     = "DELETE FROM table_name WHERE condition;"
	unknown = "alter table user drop name;"
	errSql  = "abc eusw suuwh;"
)

func TestFilterSelectStatement(t *testing.T) {
	stmt, err := ParseSqlStatement(sel)
	assert.NoError(t, err)
	assert.Equal(t, "select", stmt)
}

func TestFilterDeleteStatement(t *testing.T) {
	stmt, err := ParseSqlStatement(del)
	assert.NoError(t, err)
	assert.Equal(t, "delete", stmt)
}

func TestFilterInsertStatement(t *testing.T) {
	stmt, err := ParseSqlStatement(ins)
	assert.NoError(t, err)
	assert.Equal(t, "insert", stmt)
}

func TestFilterUpdateStatement(t *testing.T) {
	stmt, err := ParseSqlStatement(upd)
	assert.NoError(t, err)
	assert.Equal(t, "update", stmt)
}

func TestUnknownStatement(t *testing.T) {
	stmt, err := ParseSqlStatement(unknown)
	assert.NoError(t, err)
	assert.Equal(t, "unknown", stmt)
}

func TestErrorSqlStatement(t *testing.T) {
	_, err := ParseSqlStatement(errSql)
	assert.Error(t, err)
}
