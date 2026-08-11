package sstable

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/Vince-maple-byte/KeyData/internals/record"
	"github.com/ccoveille/go-safecast/v2"
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

// TODO: Change WriteToFile to be SSTFile struct complient
func WriteToFile(list [][]byte, filePath string) (*SSTFile, error) {
	filename := ""
	files, err := os.ReadDir(filePath)

	if err != nil {
		return nil, err
	}

	maxIndex := 0
	if len(files) < 1 {
		filename = "kd_1.sst"
	} else {
		for _, f := range files {
			str := strings.Split(f.Name(), "_")[1]
			str = strings.Split(str, ".")[0]
			idx, err := strconv.Atoi(str)
			if err == nil && idx > maxIndex {
				maxIndex = idx
			}
		}
		filename = "kd_" + strconv.Itoa(maxIndex+1) + ".sst"
	}

	sstFile := &SSTFile{}
	// #nosec G304 -- Creating SSTables enumerated from the internal storage directory.
	err = sstFile.Open(filepath.Join(filePath, filename))

	if err != nil {
		return nil, err
	}

	offset := fileOffset(list)
	index := createIndexBlock(list, offset)
	footer, err := createFooter(list)

	if err != nil {
		return nil, err
	}

	list = append(list, index)
	list = append(list, footer)

	content := slices.Concat(list...)

	sstFile.PopulateIndex(index)
	sstFile.Size, err = safecast.Convert[int64](len(content))
	if err != nil {
		return nil, err
	}
	indexOffset, err := safecast.Convert[uint64](sstFile.Size - int64(len(index)+len(footer)))
	sstFile.Footer = &Footer{
		Magic:       uint64(0xDEADBEEFDEADBEEF),
		IndexOffset: indexOffset,
	}

	_, err = sstFile.File.Write(content)

	if err != nil {
		return nil, err
	}

	if err := sstFile.File.Sync(); err != nil {
		return nil, err
	}

	return sstFile, nil
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
	index := make([]byte, 0, INDEX_BLOCK*160)

	//I made this specific change for the index block since in the
	if len(contentList) < INDEX_BLOCK*160 {
		contents := record.GetContents(contentList[0])

		index = binary.BigEndian.AppendUint32(index, contents.Keysize)

		index = append(index, contents.Key...)

		index = binary.BigEndian.AppendUint64(index, offset[0])
	} else {
		for i := 0; i < len(contentList); i = i + (len(contentList) / INDEX_BLOCK) {
			contents := record.GetContents(contentList[i])

			index = binary.BigEndian.AppendUint32(index, contents.Keysize)

			index = append(index, contents.Key...)

			index = binary.BigEndian.AppendUint64(index, offset[i])
		}
	}

	//var size uint64 = uint64(len(index))
	//index = append(binary.BigEndian.AppendUint64([]byte{}, size), index...)

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

func createFooter(list [][]byte) ([]byte, error) {
	if len(list) <= 0 {
		return nil, errors.New("not able to create the footer because the record list is too small")
	}
	footer := make([]byte, 0, 24)
	footer = binary.BigEndian.AppendUint64(footer, 0)
	var size uint64

	for _, rec := range list {
		size += uint64(len(rec))
	}
	footer = binary.BigEndian.AppendUint64(footer, size)
	footer = binary.BigEndian.AppendUint64(footer, uint64(0xDEADBEEFDEADBEEF))

	return footer, nil
}

func ExportFooter(list [][]byte) ([]byte, error) {
	return createFooter(list)
}

func ExportFileOffset(contentList [][]byte) []uint64 {
	return fileOffset(contentList)
}
