package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func joinStr(strs []string) string {
	return "[" + strings.Join(strs, ", ") + "]"
}

func generateTxID() string {
	return fmt.Sprintf("TX%d", rand.Intn(9000)+1000)
}