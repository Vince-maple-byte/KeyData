package network

// This is the memtable
type InternalMemtable interface {
	Write(key, value, operation string) (bool, error)
	Get(key string) ([]byte, error)
}

var NetworkMemtable func() InternalMemtable

// This is the filepath struct
type Dir struct {
	Path string
}

var FileDir *Dir
