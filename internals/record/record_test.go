package record_test

import (
	"encoding/binary"
	"hash/crc32"
	"testing"
	"time"

	"github.com/Vince-maple-byte/KeyData/internals/record"
)

func TestRecordTimeStamp(t *testing.T) {
	tests := []struct {
		testName  string
		key       string
		payload   string
		operation string
		expected  time.Time
	}{
		{
			testName:  "RecordTimeStampsIsShownAndAddedCorrectly",
			key:       "a",
			payload:   "1234",
			operation: "PUT",
		},
		{
			testName:  "RecordTimeStampsIsShownAndAddedCorrectly",
			key:       "abcdefghijklmnopqrstuvwxyz",
			payload:   "1",
			operation: "PUT",
		},
		{
			testName:  "RecordTimeStampsIsShownAndAddedCorrectly",
			key:       "a",
			payload:   "",
			operation: "DELETE",
		},
		{
			testName:  "RecordTimeStampsIsShownAndAddedCorrectly",
			key:       "j",
			payload:   "1234",
			operation: "PUT",
		},
	}

	for _, test := range tests {
		record, err := record.CreateRecord(test.key, test.payload, test.operation)
		now := time.Now()

		if err != nil {
			t.Errorf("Unable to create the record")
		}

		//var parseTime time.Time;
		parseTime := time.Unix(0, int64(binary.BigEndian.Uint64(record[:8])))

		if parseTime.IsZero() {
			t.Errorf("%v is zero", parseTime)
		}

		if parseTime.After(now) || now.Sub(parseTime) > time.Second {
			t.Errorf("timestamp %v is not within the expected time range.", parseTime)
		}
	}
}

func TestRecordCheckSum(t *testing.T) {
	tests := []struct {
		testName  string
		key       string
		payload   string
		operation string
		expected  uint32
	}{
		{
			testName:  "Test 1",
			key:       "a",
			payload:   "abc",
			operation: "PUT",
			expected: func() uint32 {
				data, _ := record.CreateRecord("a", "abc", "PUT")
				f := make([]byte, 20)
				copy(f[0:8], data[:8])
				binary.BigEndian.PutUint32(f[8:12], uint32(len("a")))
				binary.BigEndian.PutUint32(f[12:16], uint32(len("abc")))
				copy(f[16:17], []byte("a"))
				copy(f[17:], []byte("abc"))
				return crc32.ChecksumIEEE(f)
			}(),
		},
		{
			testName:  "Test 2",
			key:       "a",
			payload:   "",
			operation: "DELETE",
			expected: func() uint32 {
				data, _ := record.CreateRecord("a", "", "PUT")
				f := make([]byte, 17)
				copy(f[0:8], data[:8])
				binary.BigEndian.PutUint32(f[8:12], uint32(len("a")))
				binary.BigEndian.PutUint32(f[12:16], uint32(len("")))
				copy(f[16:17], []byte("a"))
				copy(f[17:], []byte(""))
				return crc32.ChecksumIEEE(f)
			}(),
		},
		{
			testName:  "Test 3",
			key:       "a",
			payload:   "a",
			operation: "PUT",
			expected: func() uint32 {
				data, _ := record.CreateRecord("a", "a", "PUT")
				f := make([]byte, 18)
				copy(f[0:8], data[:8])
				binary.BigEndian.PutUint32(f[8:12], uint32(len("a")))
				binary.BigEndian.PutUint32(f[12:16], uint32(len("a")))
				copy(f[16:17], []byte("a"))
				copy(f[17:], []byte("a"))
				return crc32.ChecksumIEEE(f)
			}(),
		},
	}

	for _, test := range tests {
		record, err := record.CreateRecord(test.key, test.payload, test.operation)

		if err != nil {
			t.Errorf("Unable to create the record")
		}

		//var parseTime time.Time;
		checksum := binary.BigEndian.Uint32(record[8:12])

		if checksum != test.expected {
			t.Errorf("incorrect checksum for %v:\nExpected checksum %v\nActual checksum %v", test.testName, test.expected, checksum)
		}
	}
}

func TestRecordTombstone(t *testing.T) {
	tests := []struct {
		testName  string
		key       string
		payload   string
		operation string
		expected  uint8
	}{
		{
			testName:  "Test Put",
			key:       "a",
			payload:   "abc",
			operation: "PUT",
			expected:  0,
		},
		{
			testName:  "Test Delete",
			key:       "a",
			payload:   "",
			operation: "DELETE",
			expected:  1,
		},
	}

	for _, test := range tests {
		record, err := record.CreateRecord(test.key, test.payload, test.operation)

		if err != nil {
			t.Errorf("Unable to create the record")
		}

		//var parseTime time.Time;
		tombstone := uint8(record[12])

		if tombstone != test.expected {
			t.Errorf("incorrect checksum for %v", test.testName)
		}
	}
}

func TestRecordKeySize(t *testing.T) {
	tests := []struct {
		testName  string
		key       string
		payload   string
		operation string
		expected  uint32
	}{
		{
			testName:  "Test Key Size For Put",
			key:       "a",
			payload:   "abc",
			operation: "PUT",
			expected:  1,
		},
		{
			testName:  "Test Key Size for Delete",
			key:       "abcdefghijklmnopqrstuvwxyz",
			payload:   "a",
			operation: "DELETE",
			expected:  26,
		},
	}

	for _, test := range tests {
		record, err := record.CreateRecord(test.key, test.payload, test.operation)

		if err != nil {
			t.Errorf("Unable to create the record")
		}

		//var parseTime time.Time;
		keySize := binary.BigEndian.Uint32(record[13:17])

		if keySize != test.expected {
			t.Errorf("incorrect checksum for %v", test.testName)
		}
	}
}

func TestRecordPayloadSize(t *testing.T) {
	tests := []struct {
		testName  string
		key       string
		payload   string
		operation string
		expected  uint32
	}{
		{
			testName:  "Test Payload Size For Put",
			key:       "a",
			payload:   "abc",
			operation: "PUT",
			expected:  3,
		},
		{
			testName:  "Test Payload Size for Delete",
			key:       "abcdefghijklmnopqrstuvwxyz",
			payload:   "abcdefghijklmnopqrstuvwxyz",
			operation: "DELETE",
			expected:  0,
		},
	}

	for _, test := range tests {
		record, err := record.CreateRecord(test.key, test.payload, test.operation)

		if err != nil {
			t.Errorf("Unable to create the record")
		}

		//var parseTime time.Time;
		payloadSize := binary.BigEndian.Uint32(record[17:21])

		if payloadSize != test.expected {
			t.Errorf("incorrect checksum for %v", test.testName)
		}
	}
}

func TestRecordKey(t *testing.T) {
	tests := []struct {
		testName  string
		key       string
		payload   string
		operation string
		expected  string
	}{
		{
			testName:  "Test Key Size For Put",
			key:       "a",
			payload:   "abc",
			operation: "PUT",
			expected:  "a",
		},
		{
			testName:  "Test Key Size for Delete",
			key:       "abcdefghijklmnopqrstuvwxyz",
			payload:   "",
			operation: "DELETE",
			expected:  "abcdefghijklmnopqrstuvwxyz",
		},
	}

	for _, test := range tests {
		record, err := record.CreateRecord(test.key, test.payload, test.operation)

		if err != nil {
			t.Errorf("Unable to create the record")
		}

		//var parseTime time.Time;
		keySize := binary.BigEndian.Uint32(record[13:17])
		key := string(record[21 : keySize+21])

		if key != test.expected {
			t.Errorf("incorrect checksum for %v", test.testName)
		}
	}
}

func TestRecordPayload(t *testing.T) {
	tests := []struct {
		testName  string
		key       string
		payload   string
		operation string
		expected  string
	}{
		{
			testName:  "Test Payload Size For Put",
			key:       "a",
			payload:   "abcdefghijklmnopqrstuvwxyz",
			operation: "PUT",
			expected:  "abcdefghijklmnopqrstuvwxyz",
		},
		{
			testName:  "Test Payload Size for Delete",
			key:       "abcdefghijklmnopqrstuvwxyz",
			payload:   "abcdefghijklmnopqrstuvwxyz",
			operation: "DELETE",
			expected:  "",
		},
	}

	for _, test := range tests {
		records, err := record.CreateRecord(test.key, test.payload, test.operation)

		if err != nil {
			t.Errorf("Unable to create the record")
		}

		//var parseTime time.Time;
		keySize := binary.BigEndian.Uint32(records[13:17])
		//payloadSize := binary.BigEndian.Uint32(record[17:21])
		payload := string(records[keySize+21:])

		if payload != test.expected {
			t.Errorf("incorrect checksum for %v", test.testName)
		}
	}
}

func TestGetContents(t *testing.T) {
	key := "helllpp"
	payload := "hhhh"
	records, err := record.CreateRecord(key, payload, "PUT")

	if err != nil {
		t.Errorf("Unable to create the record")
	}

	getContents := record.GetContents(records)

	if time.Unix(0, int64(binary.BigEndian.Uint64(records[:8]))) != getContents.Timestamp {
		t.Errorf("Was not able to retrieve the proper timestamp\nExpected:%v\nActual:%v",
			time.Unix(0, int64(binary.BigEndian.Uint64(records[:8]))),
			getContents.Timestamp)
	}

	if getContents.Checksum != binary.BigEndian.Uint32(records[8:12]) {
		t.Errorf("Was not able to retrieve the proper checksum\nExpected:%v\nActual:%v",
			binary.BigEndian.Uint32(records[8:12]), getContents.Checksum)
	}

	if getContents.Tombstone != uint8(records[12]) {
		t.Errorf("Was not able to retrieve the proper tombstone\nExpected:%v\nActual:%v",
			uint8(records[12]), getContents.Tombstone)
	}

	if getContents.Keysize != binary.BigEndian.Uint32(records[13:17]) {
		t.Errorf("Was not able to retrieve the proper key size\nExpected:%v\nActual:%v",
			binary.BigEndian.Uint32(records[13:17]), getContents.Keysize)
	}

	if getContents.Payloadsize != binary.BigEndian.Uint32(records[17:21]) {
		t.Errorf("Was not able to retrieve the proper payload size\nExpected:%v\nActual:%v",
			binary.BigEndian.Uint32(records[17:21]), getContents.Payloadsize)
	}

	if getContents.Key != string(records[21:getContents.Keysize+21]) {
		t.Errorf("Was not able to retrieve the proper key\nExpected:%v\nActual:%v",
			binary.BigEndian.Uint32(records[21:getContents.Keysize+21]), getContents.Key)
	}

	if getContents.Payload != string(records[getContents.Keysize+21:]) {
		t.Errorf("Was not able to retrieve the proper payload\nExpected:%v\nActual:%v",
			string(records[getContents.Keysize+21:]), getContents.Payload)
	}
}
