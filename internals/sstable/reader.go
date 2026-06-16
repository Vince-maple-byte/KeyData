package sstable

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"

	"github.com/Vince-maple-byte/KeyData/internals/record"
)

// How to read from the file
// We first checking the magic number inside of the footer to see if the file is valid
// We get the byte offset of the index block, and proceed to do:
// We check if the key inside of the byte offset is greater than or equal to our key
// If it is greater: we go to the previous byte offset inside of the index and search between the previous byte offset
// and the byte offset were it is greater than the key.
// if it is equal than we can just return that key/value pair inside of the file at that byte offset
// If the key is not inside of the range in which we stated before, than we can just return nil and an error message
// stating that the key can't be found
func ReadFromFile(filePath, key string) ([]byte, error) {
	file, err := os.Open(filepath.Join("../test", filePath))

	defer file.Close()
	if err != nil {
		return nil, err
	}
	fileInfo, _ := file.Stat()
	footer := make([]byte, 24)

	file.ReadAt(footer, fileInfo.Size()-24)

	//Checking the magic number first
	if binary.BigEndian.Uint64(footer[16:]) != uint64(0xDEADBEEFDEADBEEF) {
		err = errors.New("Incorrect magic number inside of the file. The file is invalid")
		return nil, err
	}

	//Getting the index block offset
	indexBlockLoc := binary.BigEndian.Uint64(footer[8:16])
	lowKeyOffset := uint64(0)
	highKeyOffset := uint64(0)

	for i := indexBlockLoc; i <= uint64(fileInfo.Size()-24); {
		//highKeyOffset = i
		keySize := make([]byte, 4)
		file.ReadAt(keySize, int64(i))
		offsetKey := make([]byte, int64(binary.BigEndian.Uint32(keySize)))
		file.ReadAt(offsetKey, int64(i)+4)
		//This gives us the location of the offset where it is saved in the data portion of the file
		offsetLoc := make([]byte, 8)
		file.ReadAt(offsetLoc, int64(i+4+uint64(binary.BigEndian.Uint32(keySize))))
		keyOffset := binary.BigEndian.Uint64(offsetLoc)

		//Go to the next location in the index block
		i = i + 4 + uint64(binary.BigEndian.Uint32(keySize)) + 8

		if key == string(offsetKey) {
			curr := make([]byte, 21)
			file.ReadAt(curr, int64(keyOffset))
			payloadSize := binary.BigEndian.Uint32(curr[17:21])

			entireRecord := make([]byte, 21+binary.BigEndian.Uint32(keySize)+payloadSize)
			file.ReadAt(entireRecord, int64(keyOffset))
			return entireRecord, nil
		}

		lowKeyOffset = highKeyOffset
		highKeyOffset = keyOffset

		if key < string(offsetKey) {
			break
		}
	}

	if lowKeyOffset == highKeyOffset {
		return nil, errors.New("Key is not inside of the file: Key is larger than any key in the file")
	}

	//TODO: Make the range to

	for i := lowKeyOffset; i < highKeyOffset; {
		curr := make([]byte, 21)
		file.ReadAt(curr, int64(i))
		keySize := binary.BigEndian.Uint32(curr[13:17])
		payloadSize := binary.BigEndian.Uint32(curr[17:21])

		entireRecord := make([]byte, 21+keySize+payloadSize)
		file.ReadAt(entireRecord, int64(i))

		currRecord := record.GetContents(entireRecord)

		if !record.ChecksumChecker(
			[]byte(currRecord.Key),
			[]byte(currRecord.Payload),
			uint64(currRecord.Timestamp.UnixNano()),
			currRecord.Checksum) {
			return nil, errors.New("This section of the file is corrupted, can not retrieve the key/value pairs from here")
		}

		if currRecord.Key == key {
			return entireRecord, nil
		}

		if currRecord.Key > key {
			break
		}

		i += 21 + uint64(keySize) + uint64(payloadSize)
	}

	return nil, errors.New("Key could not be found")
}
