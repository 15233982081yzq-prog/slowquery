package main

import (
	"fmt"
)

func chanTest(done chan bool) {
	ch := make(chan int, 1)
	ch <- 10000
	fmt.Println(<-ch)
	ch <- 9923
	go rec(ch, done)

}
func rec(ch chan int, done chan bool) {
	ret := <-ch
	count := 0
	for {

		fmt.Println(ret)
		count++
		if count == 5 {
			done <- true
			return
		}
	}
}
