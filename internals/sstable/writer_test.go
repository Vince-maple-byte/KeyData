package sstable_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Vince-maple-byte/KeyData/internals/memtable"
	"github.com/Vince-maple-byte/KeyData/internals/record"
	"github.com/Vince-maple-byte/KeyData/internals/sstable"
)

func TestCheck(t *testing.T) {
	if 2+2 == 5 {
		t.Error()
	}
}

func startUp(data ...string) (int, error) {
	size := 0
	for i := range 10 {
		targetDir := "../test"
		name := fmt.Sprintf("%s_%d.sst", "testfile", i)
		fullPath := filepath.Join(targetDir, name)
		file, err := os.Create(fullPath)

		if err != nil {
			return 0, err
		}

		//This is 31 bytes long for the even method and the increasing method will be +1
		w, _ := record.CreateRecord(data[i], data[i], "PUT")
		size += len(w)

		file.Write(w)
		file.Close()
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
		if err != nil {
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

func TestCompactFilesForFiles(t *testing.T) {
	data := []string{"data", "data", "data", "data", "data", "data", "data", "data", "data", "data"}

	startUp(data...)

	sstable.MergeList = func() sstable.ListMerger {
		return memtable.CreateSkiplist()
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

	for range 10 {
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
}
