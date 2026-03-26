package file

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"github.com/Vince-maple-byte/KeyData/internals/memtable"
	"github.com/Vince-maple-byte/KeyData/internals/record"
)

const COMPACTION_SIZE = 4;

//TODO:Finish with the write method for the file i/o; use the diagram that I made as a guide.
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
	

	// file.Write()
	record
} 	


