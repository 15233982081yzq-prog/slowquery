package sys

import (
	"os"
	"syscall"
)

const (
	ShopeeSpaceEmail = "EMAIL"
	TimeNormalFormat = "2006-01-02 15:04:05"
)

var Signs = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT,
}
