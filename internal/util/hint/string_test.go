package hint

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

var (
	messageHint = "SELECT /*+ trace dea06b7bffddeed73be864f674eb7600:0000006bdc2b36ca:0000000000000000 */ * FROM user_test"
	message     = "SELECT * FROM user_test"
	target      = "dea06b7bffddeed73be864f674eb7600:0000006bdc2b36ca:0000000000000000"
)

func TestGetPatternInfo(t *testing.T) {
	hint := GetSqlTraceHint(message)
	assert.Equal(t, "", hint)
}

func TestGetPatternInfo_Hint(t *testing.T) {
	hint := GetSqlTraceHint(messageHint)
	assert.Equal(t, target, hint)
}

func TestRemoveHint(t *testing.T) {
	sql := RemoveHint("select * from abc")
	assert.NotEqual(t, "", sql)
}
