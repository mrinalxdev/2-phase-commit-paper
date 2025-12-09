package main

import (
	"math/rand"
	"time"
	"fmt"
)

func SimulateLocalWork(txID string, site string) bool {
	Log(site, fmt.Sprintf("Executing transaction %s locally...", txID))
	time.Sleep(100 * time.Millisecond)
	return rand.Intn(100) > 10
}