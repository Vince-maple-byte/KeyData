package sstable

import (
	"encoding/binary"
	"io/fs"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/Vince-maple-byte/KeyData/internals/record"
)

type file_buckets float64

const (
	COMPACTION_SIZE              = 4
	INDEX_BLOCK                  = 20
	SMALL           file_buckets = 0.5
	MEDIUM          file_buckets = 1.0
	LARGE           file_buckets = 1.5
	OVERSIZE        file_buckets = 2.0
)

// TODO:Finish with the write method for the file i/o; use the diagram that I made as a guide.
func WriteToFile(list [][]byte) (bool, error) {
	filepath := "../data"
	filename := ""
	files, err := os.ReadDir(filepath)

	if err != nil {
		return false, err
	}

	if len(files) < 1 {
		filename = "kd_1.sst"
	} else {
		index, err := strconv.Atoi(strings.Split(files[len(files)-1].Name(), "_")[1])

		if err != nil {
			return false, err
		}
		filename = "kd_" + strconv.Itoa(index+1) + ".sst"
	}

	file, err := os.Create(filename)

	if err != nil {
		return false, err
	}

	contents := list
	offset := fileOffset(contents)
	index := createIndexBlock(contents, offset)

	contents = append(contents, index)

	content := slices.Concat(contents...)

	_, err = file.Write(content)

	if err != nil {
		return false, err
	}

	file.Sync()

	buckets(files)

	return true, nil
}

//Create the indexing block for the file
/*
	The indexing block is responsible for allowing us to perform reads much more efficiently.

	How is it done:
	The skiplist would be called to write into the file after 3200 entries are filled.
	When this is done, we can categorize the file into 20 sections of 160 elements.

	In the indexing block, each entry would represent a specific block.
	The entry will contain the key and the byte offset of the lowest key value in the block.
	So it would look like this for example
	keyA=0,keyB=202,keyC=403, etc.

	Since all of the records in the file are organized in sorted order, we can use the block to traverse through the file much quicker

	For example, if we have a key called N, we would start by comparing the key of the 20th block,
	if the key is greater than or equal to the key N we would start from the byte offset of that block,
	else we keep on going backwards by 1 block until we find a block where the key is greater than or equal to the block.

	Only issue with this method. In the case where the key is not located in the file, we would still need to do this traversal,
	Solution using bloom filters.

	Format:
	Size of index block: uint64
	Each entry will follow like this:
	KeySize: uint32 bytes long;
	Key: n bytes long (string) (n for keysize)
	Byte offset: uint64
*/

func createIndexBlock(contentList [][]byte, offset []uint64) []byte {
	index := make([]byte, 0, INDEX_BLOCK*16)

	for i := 0; i < len(contentList); i = i + (len(contentList) / INDEX_BLOCK) {
		contents := record.GetContents(contentList[i])

		index = binary.BigEndian.AppendUint32(index, contents.Keysize)

		index = append(index, contents.Key...)

		index = binary.BigEndian.AppendUint64(index, offset[i])
	}

	var size uint64 = uint64(len(index))
	index = append(binary.BigEndian.AppendUint64([]byte{}, size), index...)

	return index
}

func fileOffset(contentList [][]byte) []uint64 {
	var offsetTracker uint64 = 0
	offset := make([]uint64, len(contentList))

	for i, v := range contentList {
		recordsize := len(v)

		offset[i] = offsetTracker

		offsetTracker += uint64(recordsize)
	}

	return offset
}

// We are going to be doing size based compaction for compacting these files
// The amount of files that need to be a similar size
func buckets(files []os.DirEntry) map[file_buckets][]fs.FileInfo {
	// min_threshold := 4
	// max_threshold := 32
	var average_size int64
	var total_size int64
	//var largest_size int64 = 0
	//var min

	buckets := make(map[file_buckets][]fs.FileInfo)
	//
	//

	buckets[SMALL] = make([]fs.FileInfo, 0)
	buckets[MEDIUM] = make([]fs.FileInfo, 0)
	buckets[LARGE] = make([]fs.FileInfo, 0)
	buckets[OVERSIZE] = make([]fs.FileInfo, 0)

	for _, file := range files {

		fileInfo, err := file.Info()

		if err != nil {
			continue
		}

		total_size += fileInfo.Size()
	}

	average_size = total_size / int64(len(files))

	for _, file := range files {
		fileInfo, _ := file.Info()
		size := fileInfo.Size()

		switch {
		case float64(size) >= float64(average_size)*float64(OVERSIZE):
			buckets[OVERSIZE] = append(buckets[OVERSIZE], fileInfo)
		case float64(size) >= float64(average_size)*float64(LARGE):
			buckets[LARGE] = append(buckets[LARGE], fileInfo)
		case float64(size) >= float64(average_size)*float64(MEDIUM):
			buckets[MEDIUM] = append(buckets[MEDIUM], fileInfo)
		default:
			buckets[SMALL] = append(buckets[SMALL], fileInfo)
		}
	}

	return buckets
}

func ExportBuckets(files []os.DirEntry) map[file_buckets][]fs.FileInfo {
	return buckets(files)
}
