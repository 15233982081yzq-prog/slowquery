package daily

import (
	"smart-slowquery/pkg/cron/daily/task"
	"smart-slowquery/pkg/log"
)

type Runner struct {
	tasks []task.Task
}

func NewRunner(tasks []task.Task) (runner *Runner, err error) {
	return &Runner{
		tasks: tasks,
	}, nil
}

func (r *Runner) Action() {
	for _, task := range r.tasks {
		log.Infof("cron task action:%s is finished error:%v", task.Name(), task.DO())
	}
}
