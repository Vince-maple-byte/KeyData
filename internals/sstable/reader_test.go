package sstable_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/Vince-maple-byte/KeyData/internals/record"
	"github.com/Vince-maple-byte/KeyData/internals/skiplist"
	"github.com/Vince-maple-byte/KeyData/internals/sstable"
)

func TestReadFile(t *testing.T) {
	tests := []struct {
		testName string
		data     [][]byte
		key      string
		expected string
	}{
		{
			testName: "key_size_1",
			data: func() [][]byte {
				d := skiplist.CreateSkiplist()
				for i := range 3200 {
					r, _ := record.CreateRecord(strconv.Itoa(i), strconv.Itoa(i+1), "PUT")
					d.Insert(strconv.Itoa(i), r)
				}

				return d.EntireList()
			}(),
			key:      "8",
			expected: "9",
		},
		{
			testName: "key_size_2",
			data: func() [][]byte {
				d := skiplist.CreateSkiplist()
				for i := range 3200 {

					r, _ := record.CreateRecord(strconv.Itoa(i), strconv.Itoa(i+1), "PUT")
					d.Insert(strconv.Itoa(i), r)
				}
				return d.EntireList()
			}(),
			key:      "31",
			expected: "32",
		},
		{
			testName: "key_size_3",
			data: func() [][]byte {
				d := skiplist.CreateSkiplist()
				for i := range 3200 {

					r, _ := record.CreateRecord(strconv.Itoa(i), strconv.Itoa(i+1), "PUT")
					d.Insert(strconv.Itoa(i), r)
				}
				return d.EntireList()
			}(),
			key:      "432",
			expected: "433",
		},
		{
			testName: "key_size_4",
			data: func() [][]byte {
				d := skiplist.CreateSkiplist()
				for i := range 3200 {

					r, _ := record.CreateRecord(strconv.Itoa(i), strconv.Itoa(i+1), "PUT")
					d.Insert(strconv.Itoa(i), r)
				}
				return d.EntireList()
			}(),
			key:      "3197",
			expected: "3198",
		},
		{
			testName: "key_is_a_direct_match_to_the_index",
			data: func() [][]byte {
				d := skiplist.CreateSkiplist()
				for i := range 3200 {

					r, _ := record.CreateRecord(strconv.Itoa(i), strconv.Itoa(i+1), "PUT")
					d.Insert(strconv.Itoa(i), r)
				}
				return d.EntireList()
			}(),
			key:      "1140",
			expected: "1141",
		},
		{
			testName: "key_if_file_truncated",
			data: func() [][]byte {
				d := skiplist.CreateSkiplist()
				for i := range 240 {

					r, _ := record.CreateRecord(strconv.Itoa(i), strconv.Itoa(i+1), "PUT")
					d.Insert(strconv.Itoa(i), r)
				}
				return d.EntireList()
			}(),
			key:      "104",
			expected: "105",
		},
		{
			testName: "key_if_larger_than_normal",
			data: func() [][]byte {
				d := skiplist.CreateSkiplist()
				for i := range 7000 {

					r, _ := record.CreateRecord(strconv.Itoa(i), strconv.Itoa(i+1), "PUT")
					d.Insert(strconv.Itoa(i), r)
				}
				return d.EntireList()
			}(),
			key:      "4006",
			expected: "4007",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			dir := t.TempDir()
			_, err := sstable.WriteToFile(test.data, dir)

			if err != nil {
				t.Fatalf("Error in writing the file: %s", err.Error())
			}
			fileInfo, err := os.Lstat(filepath.Join(dir, "kd_1.sst"))
			size := fileInfo.Size()
			if err != nil {
				tearDown("../test")
				t.Fatalf("Error in accessing the file stats: %s", err.Error())
			}
			rec, err := sstable.ReadFromFile(filepath.Join(dir, "kd_1.sst"), test.key)

			if rec == nil || err != nil {
				t.Fatalf("Error received: %v\nSize of the file: %d", err.Error(), size)
			}
			if record.GetContents(rec).Payload != test.expected {
				t.Errorf("Test failed for %s\nExpected:%v\nActual:%v",
					test.testName, test.expected, record.GetContents(rec).Payload)
			}

			//tearDown("../test")
		})

	}
}

func TestReadFromAllFiles(t *testing.T) {

	dir := t.TempDir()
	fmt.Println("File Directory:", dir)
	for i := range 10 {
		d := skiplist.CreateSkiplist()
		for i := range 3000 * (i + 1) {

			r, _ := record.CreateRecord(strconv.Itoa(i), strconv.Itoa(i+1), "PUT")
			d.Insert(strconv.Itoa(i), r)
		}
		_, err := sstable.WriteToFile(d.EntireList(), dir)

		if err != nil {
			t.Fatalf("Error in writing the file: %s", err.Error())
		}
	}

	_, err := sstable.ReadFromAllFiles("29994", dir)

	if err != nil {
		t.Errorf("Was not able to find the valid record\n%v", err)
	}

	//tearDown("../test")
}
