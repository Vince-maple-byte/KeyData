package sstable

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/Vince-maple-byte/KeyData/internals/record"
	"github.com/Vince-maple-byte/KeyData/internals/skiplist"
	"github.com/ccoveille/go-safecast/v2"
)

//Change these methods to be SSTFile complient
type file_buckets float64

const (
	COMPACTION_SIZE              = 4
	INDEX_BLOCK                  = 20
	SMALL           file_buckets = 0.5
	MEDIUM          file_buckets = 1.0
	LARGE           file_buckets = 1.5
	OVERSIZE        file_buckets = 2.0
)
// We are going to be doing size based compaction for compacting these files
// The amount of files that need to be a similar size
// TODO: Need to make the bucket map into a persistent map that is used throughout the entire state of the
// program.w
func buckets(files []*SSTFile) (map[file_buckets][]*SSTFile, error) {

	var average_size int64
	var total_size int64

	buckets := make(map[file_buckets][]*SSTFile)

	buckets[SMALL] = make([]*SSTFile, 0)
	buckets[MEDIUM] = make([]*SSTFile, 0)
	buckets[LARGE] = make([]*SSTFile, 0)
	buckets[OVERSIZE] = make([]*SSTFile, 0)

	for _, file := range files {

		total_size += file.Size
	}

	average_size = total_size / int64(len(files))

	for _, file := range files {
		size := file.Size

		switch {
		case float64(size) >= float64(average_size)*float64(OVERSIZE):
			buckets[OVERSIZE] = append(buckets[OVERSIZE], file)
		case float64(size) >= float64(average_size)*float64(LARGE):
			buckets[LARGE] = append(buckets[LARGE], file)
		case float64(size) >= float64(average_size)*float64(MEDIUM):
			buckets[MEDIUM] = append(buckets[MEDIUM], file)
		default:
			buckets[SMALL] = append(buckets[SMALL], file)
		}
	}

	return buckets, nil
}

// When we are doing the concurrency option. Compaction and the buckets will be done in a separate class,
// and in it's own separate thread that will be run periodically every time interval that we decide
func Compact(files []*SSTFile, dir string) ([]*SSTFile, error) {
	bucketMap, err := buckets(files)

	if err != nil {
		return files, err
	}

	minThreshold := 4
	maxThreshold := 32

	//Slight potential optimization: We can have this continue compacting each bucket of similar sizes,
	// but we have to recalculate the average file size and bucket arrangement for Each of the files
	for _, val := range bucketMap {
		bucketSize := len(val)

		if bucketSize >= minThreshold && bucketSize <= maxThreshold {
			skiplist := skiplist.CreateSkiplist()
			for _, file := range val {
				// #nosec G304 -- Reading SSTables discovered via os.ReadDir from the internal storage directory.
				footer := make([]byte, 24)
				_, err := file.File.ReadAt(footer, file.Size-24)

				if err != nil {
					return files, err
				}
				//We do this so that we only take into account the file block, and not the index or footer
				if binary.BigEndian.Uint64(footer[16:]) != 0xDEADBEEFDEADBEEF {
					if err := file.Close(); err != nil {
						return files, fmt.Errorf("warning: not able to close the file %s: %v",
						file.FileName, err)	
					}
					if err := os.Remove(file.FileName); err != nil {
						return files, fmt.Errorf("warning: failed to remove corrupt SSTable %s: %v",
							file.FileName, err)
					}
					return files, fmt.Errorf("invalid footer magic")
				}
				fileBlockEnds := binary.BigEndian.Uint64(footer[8:16])

				//We are going through each file and from there we will save each key/value pair record into a skiplist
				fileData := make([]byte, file.Footer.IndexOffset)
				_, err = file.File.ReadAt(fileData, 0)

				if err != nil {
					return files, err
				}

				for i := uint64(0); i < fileBlockEnds; {
					time := fileData[i : i+8]

					checksum := binary.BigEndian.Uint32(fileData[i+8 : i+12])

					keySize := int(binary.BigEndian.Uint32(fileData[i+13 : i+17]))
					payloadSize := int(binary.BigEndian.Uint32(fileData[i+17 : i+21]))
					keySizeUi64, err := safecast.Convert[uint64](keySize)

					if err != nil {
						return files, err
					}

					payloadSizeUi64, err := safecast.Convert[uint64](payloadSize)

					if err != nil {
						return files, err
					}

					key := fileData[i+21 : i+keySizeUi64+21]
					payload := fileData[i+keySizeUi64+21 : i+(keySizeUi64+21)+payloadSizeUi64]

					//If the checksum is invalid, we ignore the rest of the file
					if !record.ChecksumChecker(key, payload, binary.BigEndian.Uint64(time), checksum) {
						break
					}

					skiplist.Insert(string(key), fileData[i:i+(keySizeUi64+21)+payloadSizeUi64])
					i = i + (keySizeUi64 + 21) + payloadSizeUi64

				}
			}

			//We take the entire skiplist and write it into a new file
			fileContents := skiplist.EntireList()

			//Technically speaking we can just recall the write to file again to make the new file
			// Since the Entire write operation is there.
			// So I'm planning on calling the Compact function after the write to file function goes through in the Memtable class
			newSSTable, err := WriteToFile(fileContents, dir)

			if err != nil {
				return files, err
			}

			newSlice := make([]*SSTFile, 0, len(files)-len(val))
			// Delete all of the old files once the new file is committed
			for _, oldFile := range val {
				err := oldFile.Delete()
				if err != nil {
					return files, err
				}
			}

			toDelete := make(map[*SSTFile]struct{})

			for _, sst := range val {
				toDelete[sst] = struct{}{}
			}

			for _, sst := range files {
				if _, exists := toDelete[sst]; exists {
					continue
				}

				newSlice = append(newSlice, sst)
			}

			newSlice = append(newSlice, newSSTable)
			return newSlice, nil
		}

	}

	return files, nil
}

func ExportBuckets(files []*SSTFile) (map[file_buckets][]*SSTFile, error) {
	result, err := buckets(files)

	return result, err
}
