package main

import (
	"fmt"
)

type Participant struct {
	ID             string
	stableLog      *StableStorage
	protocolTable  *ProtocolTable
	outbox         chan<- Message
	recoveryMode   bool
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
	}
}

func (p *Participant) handlePrepare(msg Message) {
	txID := msg.TxID
	ok := SimulateLocalWork(txID, p.ID)
	vote := ok

	if vote {
		// Force-write PREPARED
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
	}
}

// func (p *Participant) handleDecision(msg Message) {
// 	txID := msg.TxID
// 	decision := msg.Payload.(string)

// 	p.stableLog.Write(LogRecord{TxID: txID, Type: LogRecordType(decision)})
// 	p.protocolTable.Put(txID, &ProtocolEntry{
// 		TxID:     txID,
// 		State:    "DONE",
// 		Decision: decision,
// 	})

// 	Log(p.ID, fmt.Sprintf("Enforced %s for %s", decision, txID))

// 	p.outbox <- Message{
// 		Type:     AckMsg,
// 		TxID:     txID,
// 		Sender:   p.ID,
// 		Receiver: msg.Sender,
// 	}
// }
// 
func (p *Participant) handleDecision(msg Message) {
	txID := msg.TxID
	decision := msg.Payload.(string)

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
	}
}