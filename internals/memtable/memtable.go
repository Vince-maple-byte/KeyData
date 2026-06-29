package memtable

import (
	"encoding/binary"
	"errors"
	"os"

	"github.com/Vince-maple-byte/KeyData/internals/record"
	"github.com/Vince-maple-byte/KeyData/internals/sstable"
)

const MAX_SIZE = 3200

type Memtable struct {
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
	record, err := record.CreateRecord(key, value, operation)

	if err != nil {
		return false, err
	}

	ok, err := WriteToWal(record, "")

	if !ok {
		return false, err
	}

	memtable.list.Insert(key, record)
	memtable.size += 1

	if memtable.size >= MAX_SIZE {
		content := memtable.list.EntireList()
		_, errF := sstable.WriteToFile(content, "./internals/")

		if errF != nil {
			return false, errF
		}

		//For now, when running test we just replace the folder path as the test folder.
		err = sstable.Compact("../data")

		if err != nil {
			return false, err
		}

		memtable.list.EmptyList()
		memtable.size = 0

	}

	return true, nil
}

func (memtable *Memtable) Get(key string) ([]byte, error) {
	records, err := memtable.list.Search(key)

	if err != nil {
		return nil, err
	}

	content := record.GetContents(records)

	if content.Tombstone != 0 {
		return nil, errors.New("Key/Value pair doesn't exist")
	}

	return records, nil
}

func WriteToWal(record []byte, filePath string) (bool, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	defer file.Close()

	if err != nil {
		return false, errors.New("Not able to create/open the wal file")
	}

	_, err = file.Write(record)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *Memtable) MemtableStartUp(filePath string) (bool, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	defer file.Close()

	if err != nil {
		return false, err
	}

	fileInfo, err := file.Stat()

	if err != nil {
		return false, err
	}

	for i := int64(0); i < fileInfo.Size(); {
		header := make([]byte, 21)
		file.ReadAt(header, i)

		keySize := binary.BigEndian.Uint32(header[13:17])
		payloadSize := binary.BigEndian.Uint32(header[17:21])

		keyValuePair := make([]byte, keySize+payloadSize)
		file.ReadAt(keyValuePair, i+21)

		ok := record.ChecksumChecker(
			keyValuePair[:keySize],
			keyValuePair[keySize:],
			binary.BigEndian.Uint64(header[0:8]),
			binary.BigEndian.Uint32(header[8:12]),
		)

		if !ok {
			break
		}

		rec := append(header, keyValuePair...)
		m.list.Insert(string(keyValuePair[:keySize]), rec)
	}

	return true, nil
}
