package network

// This is the memtable
type InternalMemtable interface {
	Write(key, value, operation string) (bool, error)
	Get(key string) ([]byte, error)
}

var NetworkMemtable func() InternalMemtable
