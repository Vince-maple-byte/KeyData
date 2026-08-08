package sstable

import (
	"encoding/binary"
	"errors"

	"github.com/Vince-maple-byte/KeyData/internals/record"
	"github.com/ccoveille/go-safecast/v2"
)

// How to read from the file.
// We first checking the magic number inside of the footer to see if the file is valid
// We get the byte offset of the index block, and proceed to do:
// We check if the key inside of the byte offset is greater than or equal to our key
// If it is greater: we go to the previous byte offset inside of the index and search between the previous byte offset
// and the byte offset were it is greater than the key.
// if it is equal than we can just return that key/value pair inside of the file at that byte offset
// If the key is not inside of the range in which we stated before, than we can just return nil and an error message
// stating that the key can't be found
func ReadFromFile(key string, file *SSTFile) ([]byte, error) {

	footer := make([]byte, 24)

	_, err := file.File.ReadAt(footer, file.Size-24)

	if err != nil {
		return nil, err
	}
	//Checking the magic number first
	if binary.BigEndian.Uint64(footer[16:]) != uint64(0xDEADBEEFDEADBEEF) {
		err = errors.New("incorrect magic number inside of the file. The file is invalid")
		return nil, err
	}

	//Getting the index block offset
	startOffset := uint64(0)
	endOffset := uint64(0)

	for i := 0; i < len(file.Index); i++ {

		indexKey := file.Index[i].Key
		keyOffset := file.Index[i].Offset

		if i < len(file.Index)-1 {
			endOffset = file.Index[i+1].Offset
		} else {
			endOffset = file.Footer.IndexOffset
		}

		if key <= string(indexKey) {
			break
		}

		startOffset = keyOffset
	}

	for i := startOffset; i <= endOffset; {
		curr := make([]byte, 21)
		convertI, err := safecast.Convert[int64](i)

		if err != nil {
			return nil, err
		}

		_, err = file.File.ReadAt(curr, convertI)

		if err != nil {
			return nil, err
		}

		keySize := binary.BigEndian.Uint32(curr[13:17])
		payloadSize := binary.BigEndian.Uint32(curr[17:21])

		entireRecord := make([]byte, 21+keySize+payloadSize)

		_, err = file.File.ReadAt(entireRecord, convertI)

		if err != nil {
			return nil, err
		}

		currRecord := record.GetContents(entireRecord)
		timestampUi64, err := safecast.Convert[uint64](currRecord.Timestamp.UnixNano())

		if err != nil {
			return nil, err
		}

		if !record.ChecksumChecker(
			[]byte(currRecord.Key),
			[]byte(currRecord.Payload),
			timestampUi64,
			currRecord.Checksum) {
			return nil, errors.New("this section of the file is corrupted, can not retrieve the key/value pairs from here")
		}

		if currRecord.Key == key {
			return entireRecord, nil
		}

		// if currRecord.Key > key {
		// 	break
		// }

		i += 21 + uint64(keySize) + uint64(payloadSize)
	}

	return nil, errors.New("key could not be found")
}

// We have to fix the ReadFromAllFiles to use the []SSTFile struct instead dir string
func ReadFromAllFiles(key string, files []*SSTFile) ([]byte, error) {

	//By default we are going to assume that the SSTFiles are going to be sorted since the new files
	// always get written in the back
	// sort.Slice(files, func(i, j int) bool {
	// 	return files[i].Generation > files[j].Generation
	// })

	var currContent record.Content
	var result []byte
	for _, file := range files {
		res, err := ReadFromFile(key, file)

		if err == nil {
			contents := record.GetContents(res)

			if contents.Tombstone != 0 {
				return nil, errors.New("key/value pair doesn't exist")
			}

			if currContent == (record.Content{}) {
				currContent = contents
				result = res
			} else if contents.Timestamp.After(currContent.Timestamp) {
				currContent = contents
			}
		}
	}

	if currContent != (record.Content{}) {
		return result, nil
	} else {
		return nil, errors.New("the key does not exist in any of the files")
	}
}
