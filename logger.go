package main

import (
	"fmt"
	"sync"
)

var logMutex sync.Mutex

func Log(site, msg string) {
	logMutex.Lock()
	fmt.Printf("[%s] %s\n", site, msg)
	logMutex.Unlock()
}