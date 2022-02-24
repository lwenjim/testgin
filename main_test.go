package main

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	ch := make(chan int)
	go func() {
		for {
			ch <- 1
			//time.Sleep(time.Second)
		}
	}()
	for i := range ch {
		fmt.Println(i)
	}
}
