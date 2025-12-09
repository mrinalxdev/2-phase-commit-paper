package main

import "sync"

type ProtocolEntry struct {
	TxID         string
	State        string 
	Decision     string 
	Participants []string
	Acks         map[string]bool
	Votes        map[string]bool 
}

type ProtocolTable struct {
	mu      sync.Mutex
	entries map[string]*ProtocolEntry
}

func NewProtocolTable() *ProtocolTable {
	return &ProtocolTable{
		entries: make(map[string]*ProtocolEntry),
	}
}

func (pt *ProtocolTable) Get(txID string) *ProtocolEntry {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	return pt.entries[txID]
}

func (pt *ProtocolTable) Put(txID string, entry *ProtocolEntry) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.entries[txID] = entry
}

func (pt *ProtocolTable) Delete(txID string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	delete(pt.entries, txID)
}