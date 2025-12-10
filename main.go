package main

import (
	"fmt"
	"time"

	"github.com/Vince-maple-byte/KeyData/internals/file"

	records "github.com/Vince-maple-byte/KeyData/internals/record"
)

/*

TODO:

Make the Write method

Make a method to structure the record that the user will type to the binary encoding that we want

How the structure of the timestamp would look like
Timestamp | CRC32 error checksum| Tombstone (It's one byte long; basically 0 and 1 to determine whether this is a deleted key or not) | Key Size | Payload(Value) Size | Key Value |  Payload

Have it so that in the Write method after we are done adding the record into the file, make sure to add flush(), we update the index map with the key and byte offset

Make a method where we would read the file and update all of the index map key/value pairs(key and byte offset) when we open the file for the first time when the program runs

Make some unit tests to test the methods

From there we can move on to SSTable, compaction, LSM Tables, and finally bloom tables


*/

func main() {
	fmt.Println("Hello World");
	fmt.Print("Nice goals\n");
	

	f, err := file.OpenFile("./internals/file/init.txt");

	if(err != nil) {
		fmt.Println(err);
	} else {

		//Just checking if the CRC32 is working properly
		b,_ := f.ReadFile();
		t,_ := f.ReadFile();
		fmt.Println(string(b));
		fmt.Println(records.CheckSum(string(b)));

		var str string = "This is a text file.";
		s := "This is a text file.";

		fmt.Println(records.CheckSum(str));
		fmt.Println(records.CheckSum(s));

		if records.CheckSum(str) == records.CheckSum(s) {
			fmt.Println("These checksums are equal: \n" + str + "\n" + s + "\n");
		}

		if records.CheckSum(string(t)) == records.CheckSum(string(b)) {
			fmt.Println("These checksums are equal: \n" + string(b) + "\n" + string(t) + "\n");
		}

		fmt.Println(len((b)))
	}


	//How we are going to make the timestamps
	t := time.Now();
	var buffer []byte;
	buffer, err = t.AppendBinary(buffer);
	if err != nil {
		panic(err)
	}


	fmt.Println(buffer);

	var parseTime time.Time
	err = parseTime.UnmarshalBinary(buffer[:])
	if err != nil {
		panic(err)
	}

	fmt.Printf("t: %v\n", t)
	fmt.Printf("parseTime: %v\n", parseTime)
	fmt.Printf("equal: %v\n", parseTime.Equal(t))
}