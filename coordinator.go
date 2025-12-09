package main

import (
	"fmt"
)

type Coordinator struct {
	ID            string
	stableLog     *StableStorage
	protocolTable *ProtocolTable
	participants  []string
	outbox        chan<- Message
	done          chan<- bool 
}

func NewCoordinator(id string, participants []string, stableLog *StableStorage, pt *ProtocolTable, outbox chan<- Message, done chan<- bool) *Coordinator {
	return &Coordinator{
		ID:            id,
		stableLog:     stableLog,
		protocolTable: pt,
		participants:  participants,
		outbox:        outbox,
		done:          done,
	}
}

func (c *Coordinator) StartTransaction(txID string) {
	Log(c.ID, fmt.Sprintf("Starting 2PC for transaction %s", txID))

	entry := &ProtocolEntry{
		TxID:         txID,
		State:        "INIT",
		Participants: c.participants,
		Acks:         make(map[string]bool),
		Votes:        make(map[string]bool), // track votes
	}
	c.protocolTable.Put(txID, entry)

	// Phase 1: Send PREPARE to all
	for _, p := range c.participants {
		c.outbox <- Message{
			Type:     PrepareMsg,
			TxID:     txID,
			Sender:   c.ID,
			Receiver: p,
		}
		Log(c.ID, fmt.Sprintf("Sent PREPARE to %s", p))
	}
}

func (c *Coordinator) HandleMessage(msg Message) {
	entry := c.protocolTable.Get(msg.TxID)
	if entry == nil {
		return
	}

	switch msg.Type {
	case VoteMsg:
		c.handleVote(msg, entry)
	case AckMsg:
		c.handleAck(msg, entry)
	}
}

func (c *Coordinator) handleVote(msg Message, entry *ProtocolEntry) {
	if entry.State != "INIT" {
		return // already decided
	}

	sender := msg.Sender
	vote := msg.Payload.(bool)

	Log(c.ID, fmt.Sprintf("Received vote %v from %s", vote, sender))


	entry.Votes[sender] = vote
	if !vote {
		Log(c.ID, "Aborting immediately due to NO vote")
		c.makeDecision(entry.TxID, "abort")
		return
	}


	if len(entry.Votes) == len(c.participants) {
		c.makeDecision(entry.TxID, "commit")
	}
}

// func (c *Coordinator) makeDecision(txID, decision string) {
// 	entry := c.protocolTable.Get(txID)
// 	if entry == nil || entry.State != "INIT" {
// 		return
// 	}

// 	entry.State = "DECIDED"
// 	entry.Decision = decision
// 	c.protocolTable.Put(txID, entry)

// 	c.stableLog.Write(LogRecord{TxID: txID, Type: LogRecordType(decision)})
// 	Log(c.ID, fmt.Sprintf("Made decision: %s for %s", decision, txID))

// 	for _, p := range c.participants {
// 		if decision == "commit" {
// 			// Send to all
// 			c.outbox <- Message{
// 				Type:     DecisionMsg,
// 				TxID:     txID,
// 				Sender:   c.ID,
// 				Receiver: p,
// 				Payload:  decision,
// 			}
// 			Log(c.ID, fmt.Sprintf("Sent %s to %s", decision, p))
// 		} else if decision == "abort" {
// 			// Only send to those who voted YES (prepared)
// 			if vote, ok := entry.Votes[p]; ok && vote {
// 				c.outbox <- Message{
// 					Type:     DecisionMsg,
// 					TxID:     txID,
// 					Sender:   c.ID,
// 					Receiver: p,
// 					Payload:  decision,
// 				}
// 				Log(c.ID, fmt.Sprintf("Sent %s to %s (was prepared)", decision, p))
// 			} else {
// 				// Implicit abort: they already aborted locally
// 				entry.Acks[p] = true // mark as "handled"
// 			}
// 		}
// 	}

// 	// Check if all ACKs are already implicit
// 	if len(entry.Acks) == len(c.participants) {
// 		c.finalize(txID)
// 	}
// }
// 
// 
func (c *Coordinator) makeDecision(txID, decision string) {
	entry := c.protocolTable.Get(txID)
	if entry == nil || entry.State != "INIT" {
		return
	}

	entry.State = "DECIDED"
	entry.Decision = decision
	c.protocolTable.Put(txID, entry)

	c.stableLog.Write(LogRecord{TxID: txID, Type: LogRecordType(decision)})
	Log(c.ID, fmt.Sprintf("Made decision: %s for %s", decision, txID))

	if decision == "commit" {
		for _, p := range c.participants {
			c.outbox <- Message{Type: DecisionMsg, TxID: txID, Sender: c.ID, Receiver: p, Payload: "commit"}
		}
	} else { // abort
		for _, p := range c.participants {
			if entry.Votes[p] { // voted YES → must receive abort
				c.outbox <- Message{Type: DecisionMsg, TxID: txID, Sender: c.ID, Receiver: p, Payload: "abort"}
				Log(c.ID, fmt.Sprintf("Sent abort to %s (was prepared)", p))
			} else {
				entry.Acks[p] = true
				Log(c.ID, fmt.Sprintf("Implicit ACK from %s (voted NO)", p))
			}
		}
	}

	if len(entry.Acks) == len(c.participants) {
		c.finalize(txID)
	}
}

// func (c *Coordinator) handleAck(msg Message, entry *ProtocolEntry) {
// 	if entry.State != "DECIDED" {
// 		return
// 	}

// 	sender := msg.Sender
// 	entry.Acks[sender] = true

// 	if len(entry.Acks) == len(c.participants) {
// 		c.finalize(entry.TxID)
// 	}
// }
// 

func (c *Coordinator) handleAck(msg Message, entry *ProtocolEntry) {
	if entry.State != "DECIDED" {
		return
	}
	entry.Acks[msg.Sender] = true
	if len(entry.Acks) == len(c.participants) {
		c.finalize(entry.TxID)
	}
}

// func (c *Coordinator) finalize(txID string) {
// 	c.stableLog.Write(LogRecord{TxID: txID, Type: EndRecord})
// 	c.protocolTable.Delete(txID)
// 	Log(c.ID, fmt.Sprintf("Finalized and forgot transaction %s", txID))
// }
// 
func (c *Coordinator) finalize(txID string) {
	c.stableLog.Write(LogRecord{TxID: txID, Type: EndRecord})
	c.protocolTable.Delete(txID)
	Log(c.ID, fmt.Sprintf("Finalized and forgot transaction %s", txID))
	if c.done != nil {
		c.done <- true
	}
}