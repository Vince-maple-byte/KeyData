package recordTests;

import (
	"fmt"
	"github.com/Vince-maple-byte/KeyData/internals/record"
	"time"
	"encoding/binary"
)

func TestRecordTimeStampsIsShownAndAddedCorrectly() {
	record,err := records.CreateRecord("a", "15348", "PUT");

	if err != nil {
		panic(err);
	}


	
	//var parseTime time.Time;
	parseTime := time.Unix(0, int64(binary.BigEndian.Uint64(record[:8])));
	// err = parseTime.UnmarshalBinary(record[0:8]);
	// fmt.Println(record);

	// if err != nil {
	// 	panic(err);
	// }

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