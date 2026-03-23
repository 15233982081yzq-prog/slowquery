package intersect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	l1, l2, l3 []string
)

func TestSimple(t *testing.T) {
	l := Simple(l1, l2)
	assert.Equal(t, len(l), 0)
	l = Simple(l1, l3)
	assert.Equal(t, len(l), 1)
}

func TestSorted(t *testing.T) {
	l := Sorted(l1, l3)
	assert.Equal(t, len(l), 0)
}

func TestHash(t *testing.T) {
	l := Hash(l1, l2)
	assert.Equal(t, len(l), 0)
	l = Hash(l1, l3)
	assert.Equal(t, l[0], "finger")
}

func TestContainsGeneric(t *testing.T) {
	r := containsGeneric(l1, "hello")
	assert.True(t, r)
	r = containsGeneric(l1, "xxx")
	assert.False(t, r)
}

func init() {
	l1 = []string{"hello", "finger", "foot", "k8s"}
	l2 = []string{"service_mesh", "vintage", "pending"}
	l3 = []string{"finger", "crd", "csi"}
}
