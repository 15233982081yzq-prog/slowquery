package task

type Task interface {
	DO() (err error)
	Name() string
}
