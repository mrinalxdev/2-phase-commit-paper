package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== Simulating Two-Phase Commit Protocol ===")

	// Coordinator storage and table
	coordStableLog := NewStableStorage("coord_" + StableDBFile)
	coordProtocolTable := NewProtocolTable()

	participantIDs := []string{"P1", "P2", "P3"}
	participants := make(map[string]*Participant)
	messages := make(chan Message, 100)
	done := make(chan bool, 1)

	for _, id := range participantIDs {
		participantStable := NewStableStorage(id + "_" + StableDBFile)
		participantTable := NewProtocolTable()
		participants[id] = NewParticipant(id, participantStable, participantTable, messages)
	}

	coordinator := NewCoordinator("COORD", participantIDs, coordStableLog, coordProtocolTable, messages, done)

	txID := generateTxID()
	coordinator.StartTransaction(txID)

	timeout := time.After(10 * time.Second)
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
				Log("SYSTEM", "Safety break: too many messages")
				close(done)
				return
			}
		case <-done:
			fmt.Println("\n=== Transaction Completed Successfully ===")
			return
		case <-timeout:
			Log("SYSTEM", "Timeout: protocol did not complete in 10s")
			close(done)
			return
		}
	}
}