package main

import (
	"fmt"
)

type Participant struct {
	ID            string
	stableLog     *StableStorage
	protocolTable *ProtocolTable
	outbox        chan<- Message
	recoveryMode  bool
}

func NewParticipant(id string, stableLog *StableStorage, pt *ProtocolTable, outbox chan<- Message) *Participant {
	return &Participant{
		ID:            id,
		stableLog:     stableLog,
		protocolTable: pt,
		outbox:        outbox,
	}
}

func (p *Participant) HandleMessage(msg Message) {
	switch msg.Type {
	case PrepareMsg:
		p.handlePrepare(msg)
	case DecisionMsg:
		p.handleDecision(msg)
	case InquiryMsg:
		p.handleInquiry(msg)
	default:
		Log(p.ID, fmt.Sprintf("Ignoring unknown message type: %s", msg.Type))
	}
}

// func (p *Participant) handlePrepare(msg Message) {
// 	txID := msg.TxID
// 	// Log(p.ID, fmt.Sprintf("Executing transaction %s locally...", txID))
// 	ok := SimulateLocalWork(txID, p.ID)
// 	vote := ok

// 	if vote {
// 		// Force-write PREPARED
// 		p.stableLog.Write(LogRecord{TxID: txID, Type: PrepareRecord})
// 		p.protocolTable.Put(txID, &ProtocolEntry{
// 			TxID:     txID,
// 			State:    "PREPARED",
// 			Decision: "",
// 		})
// 		p.outbox <- Message{
// 			Type:     VoteMsg,
// 			TxID:     txID,
// 			Sender:   p.ID,
// 			Receiver: msg.Sender,
// 			Payload:  true,
// 		}
// 		Log(p.ID, fmt.Sprintf("Voted YES for %s", txID))
// 	} else {
// 		p.outbox <- Message{
// 			Type:     VoteMsg,
// 			TxID:     txID,
// 			Sender:   p.ID,
// 			Receiver: msg.Sender,
// 			Payload:  false,
// 		}
// 		Log(p.ID, fmt.Sprintf("Voted NO for %s", txID))
// 		// No need to keep state since we're aborting
// 	}
// }

func (p *Participant) handlePrepare(msg Message) {
	txID := msg.TxID
	
	if entry := p.protocolTable.Get(txID); entry != nil && entry.State != "" {
		Log(p.ID, fmt.Sprintf("Ignoring duplicate PREPARE for %s (already in state %s)", txID, entry.State))
		return
	}

	ok := SimulateLocalWork(txID, p.ID)
	vote := ok

	if vote {
		p.stableLog.Write(LogRecord{TxID: txID, Type: PrepareRecord})
		p.protocolTable.Put(txID, &ProtocolEntry{
			TxID:     txID,
			State:    "PREPARED",
			Decision: "",
		})
		p.outbox <- Message{
			Type:     VoteMsg,
			TxID:     txID,
			Sender:   p.ID,
			Receiver: msg.Sender,
			Payload:  true,
		}
		Log(p.ID, fmt.Sprintf("Voted YES for %s", txID))
	} else {
		p.outbox <- Message{
			Type:     VoteMsg,
			TxID:     txID,
			Sender:   p.ID,
			Receiver: msg.Sender,
			Payload:  false,
		}
		Log(p.ID, fmt.Sprintf("Voted NO for %s", txID))
		// Clean up since we're aborting
		p.protocolTable.Delete(txID)
	}
}


func (p *Participant) handleDecision(msg Message) {
	txID := msg.TxID
	decision, ok := msg.Payload.(string)
	if !ok {
		Log(p.ID, fmt.Sprintf("Invalid decision payload for %s", txID))
		return
	}

	p.stableLog.Write(LogRecord{TxID: txID, Type: LogRecordType(decision)})
	p.protocolTable.Put(txID, &ProtocolEntry{
		TxID:     txID,
		State:    "DONE",
		Decision: decision,
	})

	Log(p.ID, fmt.Sprintf("Enforced %s for %s", decision, txID))


	p.outbox <- Message{
		Type:     AckMsg,
		TxID:     txID,
		Sender:   p.ID,
		Receiver: msg.Sender,
	}
	Log(p.ID, fmt.Sprintf("Sent ACK to COORD"))
}

func (p *Participant) handleInquiry(msg Message) {
	txID := msg.TxID
	records := p.stableLog.Read(txID)
	if len(records) == 0 {
		p.outbox <- Message{
			Type:     DecisionMsg,
			TxID:     txID,
			Sender:   p.ID,
			Receiver: msg.Sender,
			Payload:  "abort",
		}
		Log(p.ID, fmt.Sprintf("Presumed ABORT for forgotten transaction %s", txID))
	} else {
		last := records[len(records)-1]
		decision := "abort"
		if last.Type == CommitRecord {
			decision = "commit"
		}
		p.outbox <- Message{
			Type:     DecisionMsg,
			TxID:     txID,
			Sender:   p.ID,
			Receiver: msg.Sender,
			Payload:  decision,
		}
		Log(p.ID, fmt.Sprintf("Responded to inquiry with %s for %s", decision, txID))
	}
}