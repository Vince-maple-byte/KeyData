package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Vince-maple-byte/KeyData/internals/memtable"
	"github.com/Vince-maple-byte/KeyData/internals/sstable"
	"github.com/ccoveille/go-safecast/v2"
)

type Database struct {
	Memtable *memtable.Memtable
	SSTFiles []*sstable.SSTFile
	Dir      string
	WalPath  string
}

func CreateDatabase(dir, walPath string) (*Database, error) {
	d := &Database{}
	d.Dir = dir
	d.WalPath = walPath

	err := d.Open()

	if err != nil {
		return nil, err
	}

	d.CreateMemtable()

	return d, nil
}

func (db *Database) Open() error {
	db.SSTFiles = make([]*sstable.SSTFile, 0)
	directory, err := os.ReadDir(db.Dir)

	if err != nil {
		return err
	}

	for _, d := range directory {
		sstFile := &sstable.SSTFile{}
		fileName := filepath.Join(db.Dir, d.Name())
		fileInfo, err := d.Info()

		if err != nil {
			return err
		}

		err = sstFile.Open(fileName)

		if err != nil {
			return err
		}

		sstFile.Size = fileInfo.Size()

		err = sstFile.PopulateFooter()

		if err != nil {
			return err
		}

		indexOffsetI64, err := safecast.Convert[int64](sstFile.Footer.IndexOffset)

		if err != nil {
			return err
		}

		indexBlock := make([]byte, (sstFile.Size-24)-indexOffsetI64)
		_, err = sstFile.File.ReadAt(indexBlock, indexOffsetI64)

		if err != nil {
			return err
		}

		sstFile.PopulateIndex(indexBlock)
	}

	return nil
}

func (db *Database) Close() error {
	for _, file := range db.SSTFiles {
		if err := file.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) CreateMemtable() {
	db.Memtable = memtable.CreateMemtable(db.WalPath, db.Dir)
}

func (db *Database) Get(key string) ([]byte, error) {
	//We would check the memtable first
	res, err := db.Memtable.Get(key)

	if err == nil {
		return res, nil
	}
	// Then all of the SSTFiles,
	// once we make the bloom filters, this will be faster since
	// we will get a good answer as to whether the key is in the file
	res, err = sstable.ReadFromAllFiles(key, db.SSTFiles)

	if err == nil {
		return res, nil
	}

	return nil, fmt.Errorf("not able to find key/value pair for %s", key)
}

func (db *Database) Put(key, val string) (bool, error) {
	ok, sstfile, err := db.Memtable.Write(key, val, "PUT")

	if !ok {
		return ok, err
	}

	if sstfile != nil {
		db.SSTFiles = append(db.SSTFiles, sstfile)
	}

	if err == nil {
		return true, err
	}

	compactErr := fmt.Errorf("need to compact the files")

	if err.Error() == compactErr.Error() {
		newFiles, errF := sstable.Compact(db.SSTFiles, db.Dir)

		if errF != nil {
			return false, errF
		}

		db.SSTFiles = newFiles
		return true, errF
	}

	return ok, err
}

func (db *Database) Delete(key string) (bool, error) {
	ok, sstfile, err := db.Memtable.Write(key, "", "DELETE")

	if !ok {
		return ok, err
	}

	if sstfile != nil {
		db.SSTFiles = append(db.SSTFiles, sstfile)
	}

	if err == nil {
		return true, err
	}

	compactErr := fmt.Errorf("need to compact the files")

	if errors.Is(err, compactErr) {
		newFiles, err := sstable.Compact(db.SSTFiles, db.Dir)

		if err != nil {
			return false, err;
		}

		db.SSTFiles = newFiles;

		return true, nil
	}

	return ok, err
}
