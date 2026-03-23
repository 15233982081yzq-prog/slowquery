package scheduler

import (
	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/schedule/task"

	"time"

	"github.com/panjf2000/ants"
	"go.uber.org/atomic"
)

const MinParallel = 1

type Scheduler struct {
	cycle    *time.Ticker
	closer   chan bool
	pool     *ants.Pool
	switcher *atomic.Bool
	tasks    []task.Task
}

func NewScheduler(cfg *conf.SchedulerConfig, tasks []task.Task) (schd *Scheduler, err error) {
	if cfg.Parallel < MinParallel {
		cfg.Parallel = MinParallel
	}
	var pool *ants.Pool

	if pool, err = ants.NewPool(cfg.Parallel, ants.WithPreAlloc(true)); err != nil {
		return schd, err
	}

	log.Infof("NewScheduler ticker cycle:%v", cfg.Cycle)

	return &Scheduler{
		cycle:    time.NewTicker(cfg.Cycle),
		closer:   make(chan bool),
		pool:     pool,
		switcher: atomic.NewBool(true),
		tasks:    tasks,
	}, nil
}

func (s *Scheduler) Start() {
	for {
		select {
		case <-s.closer:
			log.Info("Scheduler Stop!")
			return
		case <-s.cycle.C:
			log.Debugf("schedule cycle getReadyTask ... pool cap:%d ,free:%d ", s.pool.Cap(), s.pool.Free())
			if s.preCheck() {
				log.Warningf("scheduler loop pool.cap:%d ,pool.free:%d ,switcher:%t \n", s.pool.Cap(), s.pool.Free(), s.switcher.Load())
				continue
			}
			s.Action()
		}
	}
}

func (s *Scheduler) Action() {
	for _, task := range s.tasks {
		log.Infof("schedule do action task:%s is finished error:%v", task.Name(), task.DO())
	}
}

func (s *Scheduler) Running() int64 {
	return int64(s.pool.Cap() - s.pool.Free())
}

func (s *Scheduler) Stop() {
	s.switcher.Store(false)
	s.cycle.Stop()
	s.pool.Release()
}

func (s *Scheduler) Pause() {
	s.switcher.Store(false)
}

func (s *Scheduler) Resume() {
	s.switcher.Store(true)
}

func (s *Scheduler) IsOpen() bool {
	return s.switcher.Load()
}

func (s *Scheduler) preCheck() bool {
	return !s.switcher.Load() || s.pool.Free() == 0
}
