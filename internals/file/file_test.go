package file_test

import (
	"testing"

	"github.com/Vince-maple-byte/KeyData/internals/file"
	records "github.com/Vince-maple-byte/KeyData/internals/record"
)

func TestReadFile(t *testing.T) {
	// Example test case
	t.Run("SampleTest", func(t *testing.T) {
		expected := 42
		actual := 42 // Replace with actual function call
		if expected != actual {
			t.Errorf("expected %d, got %d", expected, actual)
		}
	})
}

func TestWriteFile(t *testing.T){

}

func TestOpenFile(t *testing.T){
	//Testing table
	tests := []struct{
		testName string
		fileName string
		expectedSize int64
		expectedErr bool
	}{
		{
			testName: "IfFileDoesExists",
			fileName: "./init.txt",
			expectedSize: 46,
			expectedErr: false,
		},
		{
			testName: "IfFileDoesNotExists",
			fileName: "./internals/file/fake.txt",
			expectedSize: 0,
			expectedErr: true,
		},
	}

	for _, test := range tests {

		t.Run(test.testName, func(t *testing.T) {
			//When
			f,err := file.OpenFile(test.fileName);

			//Actual
			if test.expectedErr && err == nil {
				t.Errorf("Expected error, got nil");
			}

			if !test.expectedErr && err != nil{
				t.Errorf("Unexpected error: %v", err);
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

func TestPopulateMap(t *testing.T){
	record,_ := records.CreateRecord("foo","1","PUT");
	tests := []struct{
		testName string
		ByteSlice []byte
		expectedMap map[string]int64
	}{
		{
			testName: "IfByteSliceIsEmpty",
			ByteSlice: make([]byte,0),
			expectedMap: make(map[string]int64),
		},
		{
			testName: "IfByteSliceHasValues",
			ByteSlice: record,
			expectedMap: map[string]int64{"foo":1,"bar":2,"k1":3},
		},
	}

	for _, test := range tests {
		//When
		m := file.PopulateMap(make([]byte,0));
	}
}

func TestUpdateMap(t *testing.T){

}

