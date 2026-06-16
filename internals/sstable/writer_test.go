package sstable_test

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Vince-maple-byte/KeyData/internals/memtable"
	"github.com/Vince-maple-byte/KeyData/internals/record"
	"github.com/Vince-maple-byte/KeyData/internals/sstable"
)

func startUp(data ...string) (int, error) {

	size := 0
	for range 10 {
		file := make([][]byte, 0, 10)
		for i := range 10 {
			w, _ := record.CreateRecord(data[i], data[i], "PUT")
			file = append(file, w)
		}
		_, err := sstable.WriteToFile(file)

		if err != nil {
			return 0, err
		}
		size++

	}

	return size, nil
}

func tearDown(filePath string) {
	files, _ := os.ReadDir(filePath)

	for _, file := range files {
		os.Remove(filepath.Join(filePath, file.Name()))

	}
}

func TestBucketsForFiles(t *testing.T) {
	tests := []struct {
		testName string
		data     []string
	}{
		{
			testName: "Same size for each file",
			data:     []string{"data", "data", "data", "data", "data", "data", "data", "data", "data", "data"},
		},
		{
			testName: "Size increasing for each file",
			data: []string{"data1", "data11", "data111", "data1111", "data11111",
				"data111111", "data1111111", "data11111111", "data111111111", "data1111111111"},
		},
		{
			testName: "Size decreasing for each file",
			data: []string{"data1111111111", "data111111111", "data11111111", "data1111111", "data111111",
				"data11111", "data1111", "data111", "data11", "data1"},
		},
		{
			testName: "Random order of size for each file",
			data: []string{"data111", "data1111", "data1", "data1111111", "data111111111",
				"data11111", "data1111", "data1111111111", "data11", "data11111111"},
		},
	}

	for _, test := range tests {
		totalSize, err := startUp(test.data...)
		filePath := "../test"
		if err != nil || totalSize == 0 {
			t.Fatalf("Could not start up the test")
		}

		buckets := sstable.ExportBuckets(filePath)
		average := totalSize / 10

		for key, bucket := range buckets {

			if len(bucket) != 0 {
				t.Logf("This bucket %v is %d length\n", key, len(bucket))

				for _, item := range bucket {
					if float64(item.Size()) < float64(average)*float64(key) {
						t.Errorf("For test:%s\nThis file should not be in this bucket:%v\nFile name:%s\tFile size:%d",
							test.testName,
							key, item.Name(), item.Size())
					}
				}
			}
		}
	}

	tearDown("../test")
}

func TestCompactFiles(t *testing.T) {
	data := []string{"data", "data", "data", "data", "data", "data", "data", "data", "data", "data"}

	//startUp(data...)

	sstable.MergeList = func() sstable.ListMerger {
		return memtable.CreateSkiplist()
	}

	//size := 0
	for range 10 {
		file := make([][]byte, 0, 10)
		for i := range 10 {
			w, _ := record.CreateRecord(data[i], data[i], "PUT")
			file = append(file, w)
		}
		_, err := sstable.WriteToFile(file)

		t.Logf("Size: %d", len(file))

		if err != nil || len(file) == 0 {
			t.Fatalf("Not to able properly make the file: size=%d", len(file))
		}

	}

	err := sstable.Compact("../test")

	if err != nil {
		t.Errorf("Not able to complete the compaction\n Recieved this error code:\n%v", err)
	}

	files, err := os.ReadDir("../test")

	if err != nil {
		t.Fatalf("Not able to access the directory for testing: ../test")
	}

	if len(files) != 1 {
		t.Errorf("Improper amount of files inside of the test directory:\nExpected: %d; Actual:%d", 1, len(files))
	}

	tearDown("../test")
}

func TestWriteToFile(t *testing.T) {
	fileContents := make([][]byte, 0, 10)

	for range 1 {
		r, _ := record.CreateRecord("a", "1", "PUT")
		fileContents = append(fileContents, r)
	}

	ok, err := sstable.WriteToFile(fileContents)

	if err != nil {
		t.Errorf("Error encountered: %v\n", err)
	}

	if !ok {
		t.Errorf("Not able to create the file")
	}

	file, _ := os.Open("../test/kd_1.sst")
	info, _ := file.Stat()

	if info.Size() == 0 {
		t.Error("Did not populate file")
	}

	file.Close()
	tearDown("../test")
}

func TestFooter(t *testing.T) {
	fileContents := make([][]byte, 0)

	for i := range 3200 {
		c, _ := record.CreateRecord(fmt.Sprintf("d%v", i), fmt.Sprintf("d%v", i), "PUT")
		fileContents = append(fileContents, c)
	}

	_, err := sstable.WriteToFile(fileContents)

	if err != nil {
		t.Fatal(err.Error())
	}

	file, err := os.Open("../test/kd_1.sst")
	//defer file.Close()

	if err != nil {
		t.Fatal(err.Error())
	}

	footer := make([]byte, 24)
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	_, err = file.ReadAt(footer, int64(size-24))

	if err != nil {
		t.Fatal(err.Error())
	}

	if binary.BigEndian.Uint64(footer[:8]) != uint64(0) {
		t.Errorf("Footer saved incorrect offset for the start of the file:%d", binary.BigEndian.Uint64(footer[:8]))
	}

	indexBlockLoc := make([]byte, 8)
	file.ReadAt(indexBlockLoc, int64(binary.BigEndian.Uint64(footer[8:16])))
	keySize := make([]byte, 4)

	file.ReadAt(keySize, int64(binary.BigEndian.Uint64(footer[8:16])))
	key := make([]byte, binary.BigEndian.Uint32(keySize))
	file.ReadAt(key, int64(binary.BigEndian.Uint64(footer[8:16]))+4)
	offset := make([]byte, 8)
	file.ReadAt(offset, int64(binary.BigEndian.Uint64(footer[8:16])+4+uint64(binary.BigEndian.Uint32(keySize))))
	//keyOffset := binary.BigEndian.Uint64(offset)

	//keySize := binary.BigEndian.Uint64();
	fmt.Printf("Footer bytes: %x\n", footer)
	t.Logf("key size=%d", binary.BigEndian.Uint32(keySize))
	t.Logf("Index Block Loc=%v", binary.BigEndian.Uint64(footer[8:16]))
	if binary.BigEndian.Uint64(offset) != 0 {
		t.Errorf("Footer saved incorrect offset for the start of the index block:%d\nkey=%v",
			binary.BigEndian.Uint64(offset), string(key))
	}
	if binary.BigEndian.Uint64(footer[16:24]) != uint64(0xDEADBEEFDEADBEEF) {
		t.Errorf("Footer saved incorrect magic number: %v", binary.BigEndian.Uint64(footer[16:24]))
	}

	file.Close()
	tearDown("../test")
}
