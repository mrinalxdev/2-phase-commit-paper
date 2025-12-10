package main

import (
	"math/rand"
	"time"
	// "fmt"
)

func SimulateLocalWork(txID string, site string) bool {
	time.Sleep(100 * time.Millisecond)
	

	if site == "P3" {
		return false
	}
	return rand.Intn(100) > 10 
}