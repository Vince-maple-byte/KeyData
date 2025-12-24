package main

import (
	"fmt"
	//"time"

	// "github.com/Vince-maple-byte/KeyData/internals/file"

	//"github.com/Vince-maple-byte/KeyData/internals/record"

	// "unsafe"

	"github.com/Vince-maple-byte/KeyData/tests"
)

/*

TODO:

Make the Write method

How the structure of the timestamp would look like
Timestamp | CRC32 error checksum| Tombstone (It's one byte long; basically 0 and 1 to determine whether this is a deleted key or not) | Key Size | Payload(Value) Size | Key Value |  Payload

How many bytes each element in the record is:
time stamp: 8 bytes
CRC32: 8 bytes
Tombstone: 1 byte
key size: 8 bytes
payload size: 8 bytes
key: (key size) bytes
payload: (payload size) bytes

Have it so that in the Write method after we are done adding the record into the file, make sure to add flush(), we update the index map with the key and byte offset

Make a method where we would read the file and update all of the index map key/value pairs(key and byte offset) when we open the file for the first time when the program runs

Make some unit tests to test the methods

From there we can move on to SSTable, compaction, LSM Tables, and finally bloom tables


*/

func main() {
	fmt.Println("Hello World");
	fmt.Print("Nice goals\n");
	

	// f, err := file.OpenFile("./internals/file/init.txt");

	// if(err != nil) {
	// 	fmt.Println(err);
	// }
	// } else {

	// 	//Just checking if the CRC32 is working properly
	// 	b,_ := f.ReadFile();
	// 	t,_ := f.ReadFile();
	// 	fmt.Println(string(b));
	// 	fmt.Println(records.CheckSum(string(b)));

	// 	var str string = "This is a text file.";
	// 	s := "This is a text file.";

	// 	fmt.Println(records.CheckSum(str));
	// 	fmt.Println(records.CheckSum(s));

	// 	if records.CheckSum(str) == records.CheckSum(s) {
	// 		fmt.Println("These checksums are equal: \n" + str + "\n" + s + "\n");
	// 	}

	// 	if records.CheckSum(string(t)) == records.CheckSum(string(b)) {
	// 		fmt.Println("These checksums are equal: \n" + string(b) + "\n" + string(t) + "\n");
	// 	}

	// 	fmt.Println(len((b)))
	// }


	//How we are going to make the timestamps
	

	// fmt.Println(buffer);
	// fmt.Println(len(buffer));

	// //fmt.Println("size of this in bytes",unsafe.Sizeof([]byte("This")));
	// var parseTime time.Time;
	// err = parseTime.UnmarshalBinary(buffer);
	// if err != nil {
	// 	panic(err);
	// }

	// fmt.Printf("t: %v\n", t);
	// fmt.Printf("parseTime: %v\n", parseTime);
	// fmt.Printf("equal: %v\n", parseTime.Equal(t));


	// record,err := records.CreateRecord("a", "15348", "PUT");

	// if err != nil {
	// 	panic(err);
	// }

	// var parseTime time.Time;
	// err = parseTime.UnmarshalBinary(record[:15]);
	// fmt.Println(record);

	// if err != nil {
	// 	panic(err);
	// }

	// fmt.Printf("Time: %v\n", parseTime);

	
	recordTests.TestRecordTimeStampsIsShownAndAddedCorrectly();

	recordTests.TestRecordCheckSumIfAddedCorrectly();

	recordTests.TestRecordTombstoneIfAddedCorrectly();

	recordTests.TestRecordIfKeySizeIsSavedCorrectly();
	
	recordTests.TestRecordIfPayloadSizeIsSavedCorrectly();
	
	recordTests.TestRecordIfKeyIsSavedCorrectly();

	recordTests.TestRecordIfPayloadIsSavedCorrectly();
	
}