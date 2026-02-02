package records_test

import (
	"encoding/binary"
	"hash/crc32"
	"testing"
	"time"

	records "github.com/Vince-maple-byte/KeyData/internals/record"
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
		record, err := records.CreateRecord(test.key, test.payload, test.operation)
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
			expected:  crc32.ChecksumIEEE([]byte("abc")),
		},
		{
			testName:  "Test 2",
			key:       "a",
			payload:   "",
			operation: "DELETE",
			expected:  crc32.ChecksumIEEE([]byte("")),
		},
		{
			testName:  "Test 3",
			key:       "a",
			payload:   "a",
			operation: "PUT",
			expected:  crc32.ChecksumIEEE([]byte("a")),
		},
	}

	for _, test := range tests {
		record, err := records.CreateRecord(test.key, test.payload, test.operation)

		if err != nil {
			t.Errorf("Unable to create the record")
		}

		//var parseTime time.Time;
		checksum := binary.BigEndian.Uint32(record[8:12])

		if checksum != test.expected {
			t.Errorf("incorrect checksum for %v", test.testName)
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
		record, err := records.CreateRecord(test.key, test.payload, test.operation)

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
		record, err := records.CreateRecord(test.key, test.payload, test.operation)

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
		record, err := records.CreateRecord(test.key, test.payload, test.operation)

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
		record, err := records.CreateRecord(test.key, test.payload, test.operation)

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
		record, err := records.CreateRecord(test.key, test.payload, test.operation)

		if err != nil {
			t.Errorf("Unable to create the record")
		}

		//var parseTime time.Time;
		keySize := binary.BigEndian.Uint32(record[13:17])
		//payloadSize := binary.BigEndian.Uint32(record[17:21])
		payload := string(record[keySize+21:])

		if payload != test.expected {
			t.Errorf("incorrect checksum for %v", test.testName)
		}
	}
}
