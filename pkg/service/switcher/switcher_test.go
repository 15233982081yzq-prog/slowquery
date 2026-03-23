package switcher

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCloseServer(t *testing.T) {
	CloseServer()
	assert.False(t, IsServerOpen())
}

func TestOpenServer(t *testing.T) {
	OpenServer()
	assert.True(t, IsServerOpen())
}

func init() {
	InitOpenSwitcher()
}
