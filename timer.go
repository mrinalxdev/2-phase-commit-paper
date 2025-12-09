package main

import (
	"time"
)


func StartTimeout(d time.Duration) <-chan bool {
	c := make(chan bool, 1)
	go func() {
		time.Sleep(d)
		c <- true
	}()
	return c
}