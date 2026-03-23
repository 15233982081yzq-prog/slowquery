package clickhouse

import (
	"github.com/stretchr/testify/assert"
	"smart-slowquery/conf"
	"smart-slowquery/pkg/store/request"
	"testing"
)

var flusher *Flusher

func TestNewFlusher(t *testing.T) {
	flusher = NewFlusher(ckClient, &conf.CKFlusher{
		Batch: 50,
	})
	assert.NotNil(t, flusher)
	// 防止影响， 先删除全部，再进行flush
	flusher.ckCli.db.Where("1=1").Delete(&request.SlowQueryLog{})
}

func TestFlusher_Append(t *testing.T) {
	tem := fakeData()
	for _, v := range tem {
		_, _ = flusher.Append(v)
	}
	err := flusher.FlushAll()
	assert.Nil(t, err)
}
