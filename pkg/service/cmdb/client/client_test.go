package client

import (
	"github.com/stretchr/testify/assert"
	"smart-slowquery/conf"
	"testing"
)

func TestCheck(t *testing.T) {
	cli, err := NewSpaceCMDBClient(initConfig())
	assert.NoError(t, err)
	err = cli.check()
	assert.NoError(t, err)
}

func TestCheckWithError(t *testing.T) {
	_, err := NewSpaceCMDBClient(initNoSpaceHostConfig())
	assert.Error(t, err)
}

func initConfig() *conf.Space {
	return &conf.Space{
		SpaceHost: "127.0.0.1",
		SpaceEnv:  "test",
		User:      "user",
		Pass:      "pass",
	}
}

func initNoSpaceHostConfig() *conf.Space {
	return &conf.Space{
		SpaceEnv: "test",
		User:     "user",
		Pass:     "pass",
	}
}
