package sstable

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Vince-maple-byte/KeyData/internals/record"
	"github.com/ccoveille/go-safecast/v2"
)

type SSTFile struct {
	Generation int
	FileName   string
}

// How to read from the file.
// We first checking the magic number inside of the footer to see if the file is valid
// We get the byte offset of the index block, and proceed to do:
// We check if the key inside of the byte offset is greater than or equal to our key
// If it is greater: we go to the previous byte offset inside of the index and search between the previous byte offset
// and the byte offset were it is greater than the key.
// if it is equal than we can just return that key/value pair inside of the file at that byte offset
// If the key is not inside of the range in which we stated before, than we can just return nil and an error message
// stating that the key can't be found
func ReadFromFile(filePath, key string) ([]byte, error) {
	file, err := os.Open(filePath)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	footer := make([]byte, 24)

	_, err = file.ReadAt(footer, fileInfo.Size()-24)
	
	if err != nil {
		return nil, err;
	}
	//Checking the magic number first
	if binary.BigEndian.Uint64(footer[16:]) != uint64(0xDEADBEEFDEADBEEF) {
		err = errors.New("incorrect magic number inside of the file. The file is invalid")
		return nil, err
	}

	//Getting the index block offset
	indexBlockLoc := binary.BigEndian.Uint64(footer[8:16])
	lowKeyOffset := uint64(0)
	highKeyOffset := uint64(0)

	for i := indexBlockLoc; i <= uint64(fileInfo.Size()-24); {
		keySize := make([]byte, 4)
		_,err := file.ReadAt(keySize, int64(i))
		
		if err != nil {
			return nil, err;
		}

		offsetKey := make([]byte, int64(binary.BigEndian.Uint32(keySize)))
		_, err = file.ReadAt(offsetKey, int64(i)+4)
		if err != nil {
			return nil, err;
		}

		//This gives us the location of the offset where it is saved in the data portion of the file
		offsetLoc := make([]byte, 8)
		_,err = file.ReadAt(offsetLoc, int64(i+4+uint64(binary.BigEndian.Uint32(keySize))))

		if err != nil {
			return nil, err;
		}

		keyOffset := binary.BigEndian.Uint64(offsetLoc)

		//Go to the next location in the index block
		i = i + 4 + uint64(binary.BigEndian.Uint32(keySize)) + 8

		if key == string(offsetKey) {
			curr := make([]byte, 21)
			_, err := file.ReadAt(curr, int64(keyOffset))

			if err != nil {
				return nil, err;
			}

			payloadSize := binary.BigEndian.Uint32(curr[17:21])

			entireRecord := make([]byte, 21+binary.BigEndian.Uint32(keySize)+payloadSize)
			keyOffsetConv, err := safecast.Convert[int64](keyOffset)

			if err != nil {
				return nil, err;
			}

			_,err = file.ReadAt(entireRecord, keyOffsetConv)
			
			if err != nil {
				return nil, err;
			}

			return entireRecord, nil
		}

		lowKeyOffset = highKeyOffset
		highKeyOffset = keyOffset

		if key < string(offsetKey) {
			break
		}

	}

	if lowKeyOffset == highKeyOffset {
		return nil, errors.New("key is not inside of the file: Key is larger than any key in the file")
	}

	//TODO: Make the range to

	for i := lowKeyOffset; i < highKeyOffset; {
		curr := make([]byte, 21)
		_,err := file.ReadAt(curr, int64(i))

		if err != nil {
			return nil, err;
		}
		
		keySize := binary.BigEndian.Uint32(curr[13:17])
		payloadSize := binary.BigEndian.Uint32(curr[17:21])

		entireRecord := make([]byte, 21+keySize+payloadSize)
		_,err = file.ReadAt(entireRecord, int64(i))

		if err != nil {
			return nil, err;
		}

		currRecord := record.GetContents(entireRecord)

		if !record.ChecksumChecker(
			[]byte(currRecord.Key),
			[]byte(currRecord.Payload),
			uint64(currRecord.Timestamp.UnixNano()),
			currRecord.Checksum) {
			return nil, errors.New("this section of the file is corrupted, can not retrieve the key/value pairs from here")
		}

		if currRecord.Key == key {
			return entireRecord, nil
		}

		if currRecord.Key > key {
			break
		}

		i += 21 + uint64(keySize) + uint64(payloadSize)
	}

	return nil, errors.New("key could not be found")
}

func ReadFromAllFiles(key string, dir string) ([]byte, error) {
	fileDir, err := os.ReadDir(dir)

	var files []SSTFile

	if err != nil {
		return nil, err
	}

	for _, file := range fileDir {
		gen, err := parseGeneration(filepath.Join(dir, file.Name()))

		if err != nil {
			continue
		}

		files = append(files, SSTFile{
			Generation: gen,
			FileName:   file.Name(),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Generation > files[j].Generation
	})

	var currContent record.Content
	var result []byte
	for _, file := range files {
		fmt.Println(file.FileName)
		res, err := ReadFromFile(filepath.Join(dir, file.FileName), key)

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

func parseGeneration(fileName string) (int, error) {
	str := strings.Split(fileName, "_")[1]
	str = strings.Split(str, ".")[0]
	idx, err := strconv.Atoi(str)

	return idx, err
}
