package clickhouse

import (
	"smart-slowquery/conf"
	"smart-slowquery/internal/util/function"
	"smart-slowquery/pkg/store/request"

	"fmt"
	"sync"
	"time"
)

type window struct {
	cycle time.Duration
	batch int
}

type buffer struct {
	query []*request.SlowQueryLog
	start time.Time
	window
}

func (bf *buffer) append(msg *request.SlowQueryLog) error {
	if msg == nil {
		return fmt.Errorf("slowQuery is empty")
	}

	bf.query = append(bf.query, msg)
	if len(bf.query) == 1 {
		bf.start = time.Now()
	}
	return nil
}

func (bf *buffer) trigger() bool {
	return len(bf.query) >= bf.window.batch || time.Since(bf.start) >= bf.window.cycle && len(bf.query) > 0
}

func (bf *buffer) flush() []*request.SlowQueryLog {
	return bf.query
}

func (bf *buffer) clean() {
	bf.query = bf.query[:0]
	bf.start = time.Time{}
}

type Flusher struct {
	rw    *sync.RWMutex
	bf    *buffer
	conf  *conf.CKFlusher
	ckCli *Client
}

func NewFlusher(ckCli *Client, config *conf.CKFlusher) *Flusher {
	return &Flusher{
		rw:    new(sync.RWMutex),
		ckCli: ckCli,
		conf:  config,
		bf: &buffer{
			window: window{
				cycle: config.Cycle.Duration,
				batch: config.Batch,
			},
		},
	}
}

func (f *Flusher) Append(slowQuery *request.SlowQueryLog) (err error, flushed bool) {
	f.rw.Lock()
	defer f.rw.Unlock()

	if err := f.bf.append(slowQuery); err != nil {
		return err, false
	}

	if f.bf.trigger() {
		return f.flush(3), true
	}
	return nil, false
}

func (f *Flusher) FlushAll() (err error) {
	f.rw.Lock()
	defer f.rw.Unlock()
	return f.flush(3)
}

func (f *Flusher) flush(retry int) (err error) {
	return function.Retry("clickhouse flush", func() error {
		if err := f.ckCli.batchPut(f.bf.flush()); err != nil {
			return err
		}
		f.bf.clean()
		return nil
	}, retry)
}
