package main

type MessageType string

const (
	PrepareMsg MessageType = "PREPARE"
	VoteMsg    MessageType = "VOTE"
	DecisionMsg MessageType = "DECISION"
	AckMsg     MessageType = "ACK"
	InquiryMsg MessageType = "INQUIRY"
)

type Message struct {
	Type     MessageType
	TxID     string
	Sender   string
	Receiver string
	Payload  any // bool for VoteMsg ("yes"/"no"), string for Decision ("commit"/"abort")
}