package memtable

import (
	"encoding/binary"
	"errors"
	"os"

	"fmt"
	"github.com/Vince-maple-byte/KeyData/internals/record"
	"github.com/Vince-maple-byte/KeyData/internals/sstable"
)

const MAX_SIZE = 3200

type Memtable struct {
	list    *Skiplist
	size    int
	WalFilePath string
	DataDir     string
}

func CreateMemtable(wal, dir string) *Memtable {
	return &Memtable{
		list:    CreateSkiplist(),
		size:    0,
		WalFilePath: wal,
		DataDir:     dir,
	}
}

func (m *Memtable) Write(key, value, operation string) (bool, error) {
	record, err := record.CreateRecord(key, value, operation)

	if err != nil {
		return false, err
	}

	ok, err := m.writeToWal(record)

	if !ok {
		return false, err
	}

	m.list.Insert(key, record)
	m.size += 1

	if m.size >= MAX_SIZE {
		content := m.list.EntireList()
		_, errF := sstable.WriteToFile(content, m.DataDir)

		if errF != nil {
			return false, errF
		}

		//For now, when running test we just 
		err = sstable.Compact(m.DataDir)

		if err != nil {
			return false, err
		}

		m.list.EmptyList()
		m.size = 0
		os.Remove(m.WalFilePath)

	}

	return true, nil
}

func (m *Memtable) Get(key string) ([]byte, error) {
	records, err := m.list.Search(key)

	if err != nil {
		return nil, err
	}

	content := record.GetContents(records)

	if content.Tombstone != 0 {
		return nil, errors.New("Key/Value pair doesn't exist")
	}

	return records, nil
}

func (m Memtable) writeToWal(record []byte) (bool, error) {
	file, err := os.OpenFile(m.WalFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	
	if err != nil {
		errMessage := fmt.Sprintf("Not able to create/open the wal file %s", m.WalFilePath)
		return false, errors.New(errMessage);	
	}

	defer file.Close()

	_, err = file.Write(record)

	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *Memtable) MemtableStartUp() (bool, error) {
	file, err := os.OpenFile(m.WalFilePath, os.O_APPEND|os.O_CREATE|os.O_RDONLY, 0644)
	
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
		m.list.Insert(string(rec[21: keySize + 21]), rec)

		m.size += 1
		i += int64(len(rec))
	}

	return true, nil
}
