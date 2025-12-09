package main

import (
	"fmt"
	"time"
)


func main() {
	fmt.Println("=== Simulating Two-Phase Commit Protocol ===")

	stableLog := NewStableStorage(StableDBFile)
	protocolTable := NewProtocolTable()

	participantIDs := []string{"P1", "P2", "P3"}
	participants := make(map[string]*Participant)
	messages := make(chan Message, 100)

	done := make(chan bool, 1)

	coordinator := NewCoordinator("COORD", participantIDs, stableLog, protocolTable, messages, done)

	for _, id := range participantIDs {
		participants[id] = NewParticipant(id, stableLog, protocolTable, messages)
	}

	txID := generateTxID()
	coordinator.StartTransaction(txID)

	msgCount := 0
	for {
		select {
		case msg := <-messages:
			msgCount++
			if msg.Receiver == "COORD" {
				coordinator.HandleMessage(msg)
			} else if p, ok := participants[msg.Receiver]; ok {
				p.HandleMessage(msg)
			}
			if msgCount > 100 {
				fmt.Println("[SYSTEM] Safety break: too many messages")
				return
			}
		case <-done:
			fmt.Println("\n=== Transaction Completed Successfully ===")
			return
		case <-time.After(5 * time.Second):
			fmt.Println("\n[SYSTEM] Timeout: protocol did not complete in 5s")
			return
		}
	}
}