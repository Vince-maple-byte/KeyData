package recordTests;

import (
	"fmt"
	"github.com/Vince-maple-byte/KeyData/internals/record"
	"time"
	"encoding/binary"
	"strconv"
)

func TestRecordTimeStampsIsShownAndAddedCorrectly() {
	record,err := records.CreateRecord("a", "15348", "PUT");

	if err != nil {
		panic(err);
	}

	//var parseTime time.Time;
	parseTime := time.Unix(0, int64(binary.BigEndian.Uint64(record[:8])));

	fmt.Printf("Time: %v\n", parseTime);
}

func TestRecordCheckSumIfAddedCorrectly() {
	record,err := records.CreateRecord("a", "15348", "PUT");

	if err != nil {
		panic(err);
	}

	checkSum := binary.BigEndian.Uint32(record[8:12]);

	fmt.Println("Check sum:",checkSum);
}

func TestRecordTombstoneIfAddedCorrectly(){
	record, err := records.CreateRecord("a", "15348", "PUT");

	if err != nil {
		panic(err);
	}

	tombstoneActual := uint8(record[12]);
	tombstoneExpected := uint8(0);

	if tombstoneActual != tombstoneExpected {
		panic("Unit test TestRecordTombstoneIfAddedCorrectly failed "+ strconv.Itoa(int(tombstoneActual))+" does not equal to "+strconv.Itoa(int(tombstoneExpected)));
	} else {
		fmt.Println("Test TestRecordTombstoneIfAddedCorrectly Passed");
	}
}

func TestRecordIfKeySizeIsSavedCorrectly() {
	var key string = "a";
	record, err := records.CreateRecord(key, "15348", "PUT");

	if err != nil {
		panic(err);
	}

	keySizeActual := binary.BigEndian.Uint32(record[13:17]);
	keySizeExpected := uint32(len(key));

	if keySizeActual != keySizeExpected {
		panic("Unit test TestRecordIfKeySizeIsSavedCorrectly failed "+ strconv.Itoa(int(keySizeActual))+" does not equal to "+strconv.Itoa(int(keySizeExpected)));
	} else {
		fmt.Println("Test TestRecordIfKeySizeIsSavedCorrectly Passed");
	}
}

func TestRecordIfPayloadSizeIsSavedCorrectly() {
	var payload string = "15348";
	record, err := records.CreateRecord("key", payload, "PUT");

	if err != nil {
		panic(err);
	}

	payloadSizeActual := binary.BigEndian.Uint32(record[17:21]);
	payloadSizeExpected := uint32(len(payload));

	if payloadSizeExpected != payloadSizeActual {
		panic("Unit test TestRecordIfPayloadSizeIsSavedCorrectly failed "+ strconv.Itoa(int(payloadSizeActual))+" does not equal to "+strconv.Itoa(int(payloadSizeExpected)));
	} else {
		fmt.Println("Test TestRecordIfPayloadSizeIsSavedCorrectly Passed");
	}
}

func TestRecordIfKeyIsSavedCorrectly() {
	var key string = "a";
	record, err := records.CreateRecord(key, "15348", "PUT");

	if err != nil {
		panic(err);
	}

	keySize := binary.BigEndian.Uint32(record[13:17]);
	keyActual := string(record[21:keySize+21]);

	if keyActual != key {
		panic("Unit test TestRecordIfKeyIsSavedCorrectly failed "+ keyActual+" does not equal to "+key);
	} else {
		fmt.Println("Test TestRecordIfKeyIsSavedCorrectly Passed");
	}
}

func TestRecordIfPayloadIsSavedCorrectly() {
	var key string = "a";
	var payload string = "15348";
	record, err := records.CreateRecord(key, payload, "PUT");

	if err != nil {
		panic(err);
	}

	payloadSize := binary.BigEndian.Uint32(record[17:21]);
	keySize := binary.BigEndian.Uint32(record[13:17]);
	payloadActual := string(record[keySize+21:(keySize+21)+payloadSize]);

	if payloadActual != payload {
		panic("Unit test TestRecordIfPayloadIsSavedCorrectly failed " + payloadActual+" does not equal to "+payload);
	} else {
		fmt.Println("Test TestRecordIfPayloadIsSavedCorrectly Passed");
	}
	
}