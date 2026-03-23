package mysql

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	meta = &AccessMeta{}
)

func TestMetaTableName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
		msg      string
	}{
		{
			name:     "internal_user_tab",
			expected: true,
			msg:      "same table name",
		},
		{
			name:     "internal_user_tab_clone",
			expected: false,
			msg:      "diff table name",
		},
	}

	for _, test := range tests {
		res := meta.TableName() == test.name
		assert.Equal(t, res, test.expected, test.msg)
	}
}
