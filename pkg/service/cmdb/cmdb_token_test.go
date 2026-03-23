package cmdb

import (
	envUtil "smart-slowquery/internal/util/env"

	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetToken(t *testing.T) {
	token, err := GetToken(envUtil.ServerLiveEnv)
	assert.NoError(t, err)
	fmt.Printf("env:%s ,token:%v \n", envUtil.ServerLiveEnv, token)
	token, err = GetToken(envUtil.ServerNonLiveEnv)
	assert.NoError(t, err)
	fmt.Printf("env:%s ,token:%v \n", envUtil.ServerNonLiveEnv, token)
}

func TestNewSpaceUserPassTokenFetcher(t *testing.T) {
	tmp := NewSpaceUserPassTokenFetcher("db_tools_archive", "A5RvMQUU7KDP", "test")
	tmp.FetchToken()

}
