package string

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	superset = []string{"abc", "bcd", "cde", "def"}
	subset   = []string{"abc", "bcd"}
	diffset  = []string{"", ""}
)

func TestContainInSlice(t *testing.T) {
	var (
		array = []string{"slice", "in", "contain"}
		has   = "slice"
		noHas = "sli"
	)
	assert.True(t, ContainInSlice(array, has))
	assert.False(t, ContainInSlice(array, noHas), "should no has")
}

func TestSplit(t *testing.T) {
	arr := Split("xxx,yyy", ",")
	assert.NotEmpty(t, arr)
	assert.Equal(t, arr, []string{"xxx", "yyy"})
}

func TestIsSubsetOK(t *testing.T) {
	bl, _ := IsSubset(subset, superset)
	assert.True(t, bl)
}

func TestIsSubsetFail(t *testing.T) {
	bl, _ := IsSubset(diffset, superset)
	assert.False(t, bl)
}

func TestIncrementString(t *testing.T) {
	bl, _ := IncrementString("1", 0)
	assert.Equal(t, bl, "1")

	_, err := IncrementString("a", 0)
	assert.NotEqual(t, nil, err)
}

func TestIsSubset(t *testing.T) {
	r, _ := IsSubset(subset, superset)
	assert.True(t, r)
	r, _ = IsSubset(subset, diffset)
	assert.False(t, r)
}
