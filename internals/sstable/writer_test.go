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

func startUp(data ...string) error {
	for i := range 10 {
		targetDir := "../test"
		name := fmt.Sprintf("%s_%d.txt", "testfile", i)
		fullPath := filepath.Join(targetDir, name)
		file, err := os.Create(fullPath)

		if err != nil {
			return err
		}

		//This is 31 bytes long for the even method and the increasing method will be +1
		w, _ := record.CreateRecord(data[i], data[i], "PUT")

		file.Write(w)
		file.Close()
	}

	return nil
}

func TestBucketsForEvenlyDistributedFiles(t *testing.T) {
	err := startUp()
	filePath := "../test"
	if err != nil {
		t.Fatalf("Could not start up the test")
	}

	filePath = "../test"
	files, err := os.ReadDir(filePath)

	if err != nil {
		t.Fatalf("Could not make the file path")
	}

	buckets := sstable.ExportCompact(files)
	totalSize := 310
	average := totalSize / 10

	for key, bucket := range buckets {

		if len(bucket) == 0 {
			//t.Logf("This bucket: %v is empty", key)
		} else {
			t.Logf("This bucket %v is %d length\n", key, len(bucket))

			for _, item := range bucket {
				if float64(item.Size()) < float64(average)*float64(key) {
					t.Errorf("This file should not be in this bucket:%v\nFile name:%s\tFile size:%d",
						key, item.Name(), item.Size())
				}
				//t.Logf("File %s is %d large\n", item.Name(), item.Size())
			}
		}
	}

}
