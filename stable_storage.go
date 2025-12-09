package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type StableStorage struct {
	db *sql.DB
}

func NewStableStorage(filename string) *StableStorage {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS log (
		tx_id TEXT,
		record_type TEXT
	);`)
	if err != nil {
		panic(err)
	}
	return &StableStorage{db: db}
}

func (s *StableStorage) Write(record LogRecord) {
	_, err := s.db.Exec("INSERT INTO log(tx_id, record_type) VALUES(?, ?)",
		record.TxID, string(record.Type))
	if err != nil {
		panic(err)
	}
}

func (s *StableStorage) Read(txID string) []LogRecord {
	rows, err := s.db.Query("SELECT record_type FROM log WHERE tx_id = ?", txID)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var records []LogRecord
	for rows.Next() {
		var typ string
		rows.Scan(&typ)
		records = append(records, LogRecord{TxID: txID, Type: LogRecordType(typ)})
	}
	return records
}

func (s *StableStorage) Delete(txID string) {
	s.db.Exec("DELETE FROM log WHERE tx_id = ?", txID)
}