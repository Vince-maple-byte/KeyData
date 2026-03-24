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

func (memtable *Memtable) Write(key, value string) (bool, error) {
	record, err := record.CreateRecord(key, value, "PUT");

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

//TODO: Fix this so that it is added into the skiplist as a record with an empty value and a tombstone of 1
func (memtable *Memtable) Delete(key string) (bool, error) {
	_, err := memtable.list.Delete(key);

	if err != nil {
		return false, err;
	} 

	return true, nil;
}

func (memtable *Memtable) Get(key string) ([] byte, error) {
	record, err := memtable.list.Search(key);

	if err != nil {
		return nil,err; 
	}

	return record, nil;
} 
