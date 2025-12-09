package main

type LogRecordType string

const (
	PrepareRecord LogRecordType = "PREPARED"
	CommitRecord  LogRecordType = "COMMIT"
	AbortRecord   LogRecordType = "ABORT"
	EndRecord     LogRecordType = "END"
)

type LogRecord struct {
	TxID string
	Type LogRecordType
}