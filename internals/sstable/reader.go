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
	indexBlockByte := make([]byte, 8)
	file.ReadAt(indexBlockByte, int64(binary.BigEndian.Uint64(footer[8:16])))
	indexBlockLoc := binary.BigEndian.Uint64(indexBlockByte)
	lowKeyOffset := indexBlockLoc
	highKeyOffset := indexBlockLoc

	for i := indexBlockLoc; i < uint64(fileInfo.Size()-24); {
		lowKeyOffset = i
		keySize := make([]byte, 4)
		file.ReadAt(keySize, int64(i))
		offsetKey := make([]byte, int64(binary.BigEndian.Uint32(keySize)))
		file.ReadAt(offsetKey, int64(i)+4)
		offset := make([]byte, 8)
		file.ReadAt(offset, int64(i+4+uint64(binary.BigEndian.Uint32(keySize))))
		//keyOffset := binary.BigEndian.Uint64(offset)

		i = i + 4 + uint64(binary.BigEndian.Uint32(keySize)) + 8
		if key <= string(offsetKey) {
			break
		} else {
			lowKeyOffset = highKeyOffset
			highKeyOffset = i
		}
	}

	if lowKeyOffset == highKeyOffset {
		return nil, errors.New("Key is not inside of the file")
	}

	//TODO: Make the range to
	indexBlockRange := make([]byte, highKeyOffset-lowKeyOffset)
	file.ReadAt(indexBlockRange, int64(lowKeyOffset))

	for i := lowKeyOffset; i < highKeyOffset; {
		keySize := binary.BigEndian.Uint32(indexBlockRange[i+13 : i+17])
		payloadSize := binary.BigEndian.Uint32(indexBlockRange[i+17 : i+21])

		currRecord := record.GetContents(indexBlockRange[i : i+21+uint64(keySize)+uint64(payloadSize)])

		if !record.ChecksumChecker(
			[]byte(currRecord.Key),
			[]byte(currRecord.Payload),
			uint64(currRecord.Timestamp.UnixNano()),
			currRecord.Checksum) {
			return nil, errors.New("This section of the file is corrupted, can not retrieve the key/value pairs from here")
		}

		if currRecord.Key == key {
			return indexBlockRange[i : i+21+uint64(keySize)+uint64(payloadSize)], nil
		}

		i += 21 + uint64(keySize) + uint64(payloadSize)
	}

	return nil, errors.New("Key could not be found")
}
