package memtable 

import (
	"github.com/Vince-maple-byte/KeyData/internals/record"
)

const MAX_SIZE = 3200

type Memtable struct{
	list *Skiplist
	size int
}

func CreateMemtable() *Memtable {
	return &Memtable{
		list: CreateSkiplist(),
		size: 0,
	}
}

func (memtable *Memtable) Write(key, value, operation string) (bool, error) {
	record, err := record.CreateRecord(key, value, operation);

	if err != nil {
		return false, err;
	}

	memtable.list.Insert(key, record);
	memtable.size += 1;

	if(memtable.size == MAX_SIZE) {
		//We are going to call a writeToFile function here
		memtable.list.EmptyList();
		memtable.size = 0;
		
	}
	
	return true, nil;
}

func (memtable *Memtable) Get(key string) ([]byte, error) {
	record, err := memtable.list.Search(key);

	if err != nil {
		return nil,err; 
	}

	return record, nil;
} 
