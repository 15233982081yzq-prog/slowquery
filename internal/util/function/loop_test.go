package function

import (
	"time"

	"fmt"
	"testing"
)

func TestLoopNoError(t *testing.T) {
	go Loop("TestLoopNoError", func() error {
		fmt.Printf("test retry no error\n")
		return nil
	}, 100*time.Millisecond)
	time.Sleep(2 * time.Second)
}

func TestLoopError(t *testing.T) {
	go Loop("TestLoopError", func() error {
		fmt.Printf("test retry error\n")
		return fmt.Errorf("return error")
	}, 100*time.Millisecond)
	time.Sleep(2 * time.Second)
}
