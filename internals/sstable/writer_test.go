package sstable_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

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
		name := fmt.Sprintf("%s_%d.txt", "testfile", i)
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

}
