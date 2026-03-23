package switcher

import "go.uber.org/atomic"

var swt *Switcher

type Switcher struct {
	server *atomic.Bool
}

func InitOpenSwitcher() {
	swt = &Switcher{
		server: atomic.NewBool(true),
	}
}

func CloseServer() {
	swt.server.Store(false)
}

func OpenServer() {
	swt.server.Store(true)
}

func IsServerOpen() bool {
	return swt.server.Load()
}
