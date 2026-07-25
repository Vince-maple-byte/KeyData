package sstable

type ListMerger interface {
	Insert(key string, value []byte)
	EntireList() [][]byte
}

var MergeList func() ListMerger
