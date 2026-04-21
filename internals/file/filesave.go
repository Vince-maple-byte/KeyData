package file

import (
	"encoding/binary"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/Vince-maple-byte/KeyData/internals/memtable"
	"github.com/Vince-maple-byte/KeyData/internals/record"
)

const COMPACTION_SIZE = 4;

func ReadFromStorage() ([]byte, error) {
	filepath := "./internals/data";

	files, err := os.ReadDir(filepath);

	for i := len(files) - 1; i >= 0; i-- {
		var record []byte;
		record, err := readfile(files[i].Name());

		if err == nil {
			return record, nil;
		}
	}

	return nil, errors.New("Key/Value pair could not be found")
}

func readfile(filename string) ([]byte, error) {
	file, err := os.Open(filename);

	if err != nil {
		return nil, err;
	}

	fileInfo, err := file.Stat();

	if err != nil {
		return nil, err;
	}

	size := fileInfo.Size();

	_, err = file.Seek(size - int64(128), 0);

	if err != nil {
		return nil, err;
	}

	footer := make([]byte, 128);

	file.Read(footer);

	fileOffset := binary.BigEndian.Uint64(footer[:64]);

	indexOffset := binary.BigEndian.Uint64(footer[64:]);

	// TODO: finish with this method. Start with the index block and do a binary search for each key in the block
	// until you find a key that is equal to greater than that block while being less than the next key. If the
	// key is not equal than we just do the binary search again but this time in the block of the key that is represented 
	// in the key in our file block. If it is still not there then the key/value pair doesn't exist.
}


//TODO:Make unit tests for write method for the file i/o; use the diagram that I made as a guide.
//How the file is going to look like
//File block, index block, footer
//The file block is going to contain all of the contents of the file aka the records
//The index block is going to contain the byte offset of the location of certain records in the file block
//The index block is going to be sorted and when doing file reading we will first start out in the end of the index block
//until we find a key that's less than or equal to the key saved
//Footer block tells us where the file block starts and the index block starts. This will always be a set
//byte size of 128 bytes long for the info of file block is 64 bytes and the info for index block is 64 bytes long 
func WriteToFile(list memtable.Skiplist) (bool, error) {
	filepath := "./internals/data";
	filename := ""
	files, err := os.ReadDir(filepath);

	if err != nil {
		return false, err;
	}

	if len(files) < 1 {
		filename = "kd_1.sst"
	} else {
		index,err := strconv.Atoi(strings.Split(files[len(files)-1].Name(), "_")[1]);

		if err != nil {
			return false, err;
		}
		filename = "kd_"+strconv.Itoa(index+1)+".sst"	
	}

	file, err := os.Create(filename);

	if err != nil {
		return false, err;
	}
	
	contents := list.EntireList();

	//Create file block
	fileblock := fileblock(contents);

	//Create the indexing block for the file 
	indexblock := indexblock(contents);
	
	//Create the footer block for the file 
	footerblock := footerblock(fileblock);

	write := append(fileblock, indexblock...)
	write = append(write, footerblock...)

	_,err = file.Write(write)
	
	if err != nil {
		return false, err;
	}

	err = file.Chmod(0444);

	if err != nil {
		return false, err;
	}
	
	err = file.Close();

	if err != nil {
		return false, err;
	}

	
	return true, nil
} 	

func fileblock(contents [][]byte) []byte {
	block := make([]byte, 0);

	for _,v := range(contents) {
		block = append(block, v...);
	}

	return block;
}

func indexblock(contents [][]byte) []byte {
	block := make([]byte, 0);
	currByteoffset := uint64(0);
	
	for i := 0; i < len(contents); i++ {
		_,_,_,_,_,key,_ := record.GetContents(contents[i]);	 	
		if i % COMPACTION_SIZE == 0 {
			block = append(block, []byte(key)...)
			block = append(block, binary.BigEndian.AppendUint64(make([]byte, 0), currByteoffset)...)
		}
		currByteoffset += uint64(i);
	}
	return block;
}

func footerblock(fileblock []byte) []byte {
	block := make([]byte, 0);

	block = append(block, binary.BigEndian.AppendUint64(make([]byte, 0), uint64(0))...)
	block = append(block, binary.BigEndian.AppendUint64(make([]byte, 0), uint64(len(fileblock)))...)

	return block
}

