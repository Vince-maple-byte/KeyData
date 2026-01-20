package file_test

import (
	"maps"
	"testing"

	"github.com/Vince-maple-byte/KeyData/internals/file"
	records "github.com/Vince-maple-byte/KeyData/internals/record"
)

func TestReadFile(t *testing.T) {
	// Example test case

}

func TestPutContents(t *testing.T) {

}

func TestOpenFile(t *testing.T) {
	//Testing table
	tests := []struct {
		testName     string
		fileName     string
		expectedSize int64
		expectedErr  bool
	}{
		{
			testName:     "IfFileDoesExists",
			fileName:     "./init.txt",
			expectedSize: 0,
			expectedErr:  false,
		},
		{
			testName:     "IfFileDoesNotExists",
			fileName:     "./internals/file/fake.txt",
			expectedSize: 0,
			expectedErr:  true,
		},
	}

	for _, test := range tests {

		t.Run(test.testName, func(t *testing.T) {
			//When
			f, err := file.OpenFile(test.fileName)

			//Actual
			if test.expectedErr && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !test.expectedErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if f.Size != test.expectedSize {
				t.Errorf(
					"incorrect file size: got %d, want %d",
					f.Size,
					test.expectedSize,
				)
			}

			if !test.expectedErr && f.File == nil {
				t.Errorf("expected file handle to be non-nil")
			}

		})

	}
}

func TestPopulateMap(t *testing.T) {
	record1, _ := records.CreateRecord("foo", "1", "PUT")
	recordF, _ := records.CreateRecord("foo", "4", "PUT")
	record2, _ := records.CreateRecord("bar", "1", "PUT")
	record3, _ := records.CreateRecord("n", "2", "PUT")
	record4, _ := records.CreateRecord("b", "3", "PUT")
	combinedUnique := append(append([]byte{}, record1...), record2...)
	combinedUnique = append(combinedUnique, record3...)
	combinedUnique = append(combinedUnique, record4...)
	combinedSame := append(append([]byte{}, record1...), recordF...)
	tests := []struct {
		testName    string
		ByteSlice   []byte
		expectedMap map[string]int64
	}{
		{
			testName:    "IfByteSliceIsEmpty",
			ByteSlice:   make([]byte, 0),
			expectedMap: make(map[string]int64),
		},
		{
			testName:    "IfByteSliceHasValues",
			ByteSlice:   record1,
			expectedMap: map[string]int64{"foo": 0},
		},
		{
			testName:    "IfByteSliceHasMultipleValues",
			ByteSlice:   combinedUnique,
			expectedMap: map[string]int64{"foo": 0, "bar": 25, "b": 73, "n": 50},
		},
		{
			testName:    "IfByteSliceHasMultiplePutsOftheSameKey",
			ByteSlice:   combinedSame,
			expectedMap: map[string]int64{"foo": 25},
		},
	}

	for _, test := range tests {
		//When
		m := file.PopulateMap(test.ByteSlice)

		if !maps.Equal(test.expectedMap, m) {
			t.Errorf("Map mismatch between\nexpected %v\nand\nactual %v", test.expectedMap, m)
		}
	}
}

func TestUpdateMap(t *testing.T) {
	record1, _ := records.CreateRecord("foo", "1", "PUT")
	record2, _ := records.CreateRecord("foo", "2", "PUT")
	record3, _ := records.CreateRecord("bar", "3", "PUT")
	emptyMap := make(map[string]int64)
	fileEmpty := file.File{
		File:  nil,
		Size:  0,
		Index: emptyMap,
	}

	fileNonEmpty := file.File{
		File:  nil,
		Size:  0,
		Index: map[string]int64{"foo": 0},
	}

	tests := []struct {
		testName           string
		ByteSlice          []byte
		fileStruct         file.File
		expectedFileStruct file.File
	}{
		{
			testName:   "IfMapWasEmpty",
			ByteSlice:  record1,
			fileStruct: fileEmpty,
			expectedFileStruct: file.File{
				File:  nil,
				Size:  0, //Keeping track of the size doesn't matter since the write function will handle that
				Index: map[string]int64{"foo": 0},
			},
		},
		{
			testName:   "IfMapIsNotEmptyAndWeAreAddingAnExistingKey",
			ByteSlice:  record2,
			fileStruct: fileNonEmpty,
			expectedFileStruct: file.File{
				File:  nil,
				Size:  0, //Keeping track of the size doesn't matter since the write function will handle that
				Index: map[string]int64{"foo": 25},
			},
		},
		{
			testName:  "IfMapIsNotEmptyAndWeAreAddingANewKey",
			ByteSlice: record3,
			fileStruct: file.File{
				File:  nil,
				Size:  25,
				Index: map[string]int64{"foo": 0},
			},
			expectedFileStruct: file.File{
				File:  nil,
				Size:  0, //Keeping track of the size doesn't matter since the write function will handle that
				Index: map[string]int64{"foo": 0, "bar": 25},
			},
		},
	}

	for _, test := range tests {
		//When
		test.fileStruct.UpdateMap(test.ByteSlice)

		//Then
		if !maps.Equal(test.fileStruct.Index, test.expectedFileStruct.Index) {
			t.Errorf("Map mismatch between expected %v and actual %v",
				test.fileStruct.Index, test.expectedFileStruct.Index)
		}
	}
}
