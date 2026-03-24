package main

import (
	"fmt"

	"github.com/Vince-maple-byte/KeyData/internals/file"
	//"time"
	// "github.com/Vince-maple-byte/KeyData/internals/file"
	//"github.com/Vince-maple-byte/KeyData/internals/record"
	// "unsafe"
)

/*

TODO:
Refactor the file folder to handle file operations for both in memory data store (skiplist) and the file i/o in case we have a file miss. So get rid of the map, and the everything else.

How the structure of the timestamp would look like
Timestamp | CRC32 error checksum| Tombstone (It's one byte long; basically 0 and 1 to determine whether this is a deleted key or not) | Key Size | Payload(Value) Size | Key Value |  Payload

How many bytes each element in the record is:
time stamp: 8 bytes
CRC32: 4 bytes
Tombstone: 1 byte
key size: 4 bytes
payload size: 4 bytes
key: (key size) bytes
payload: (payload size) bytes

Have it so that in the Write method after we are done adding the record into the file, make sure to add flush(), we update the index map with the key and byte offset

Make a method where we would read the file and update all of the index map key/value pairs(key and byte offset) when we open the file for the first time when the program runs

Make some unit tests to test the methods

From there we can move on to SSTable, compaction, LSM Tables, and finally bloom tables


*/

func main() {
	var database string
	fmt.Println("Welcome to KeyData")
	fmt.Print("Enter your database name to get started\n")

	_, err := fmt.Scan(&database)

	for err != nil {
		fmt.Println("Please enter a valid database name")
		_, err = fmt.Scan(&database)
	}

	database += ".log"

	f, errf := file.OpenFile(database)
	defer f.File.Close()

	if errf != nil {
		panic(errf)
	}

	fmt.Println("Database", database, "opened successfully")
	willContinue := true

	for willContinue {
		fmt.Println("Enter an operation: PUT, DELETE, GET")

		var operation string
		_, err := fmt.Scan(&operation)

		if err != nil {
			break
		}

		if operation == "PUT" {
			fmt.Println("Enter a key")
			var key string
			fmt.Scan(&key)
			fmt.Println("Enter the data that you want to save")
			var payload string
			fmt.Scan(&payload)

			numAdded, _ := f.PutContents(key, payload, "PUT")

			if numAdded > -1 {
				fmt.Println("Was able to successfully add the key/value pair into the database")
			} else {
				fmt.Println("Was not able to successfully add the key/value pair into the database")
			}

		}
		if operation == "DELETE" {
			fmt.Println("Enter a key")
			var key string
			fmt.Scan(&key)

			numAdded, _ := f.PutContents(key, "", "DELETE")

			if numAdded > -1 {
				fmt.Println("Was able to successfully delete the key/value pair into the database")
			} else {
				fmt.Println("Was not able to successfully delete the key/value pair into the database")
			}

		}
		if operation == "GET" {
			fmt.Println("Enter a key")
			var key string
			fmt.Scan(&key)

			deleted, payload, timestamp, err := f.GetContents(key)

			if err != nil {
				fmt.Println("Was able not able to retrieve the file contents for", key)
			} else {
				fmt.Printf("For key %v:\nThe payload: %v\nDeleted: %v\nThe timestamp: %v\n", key, payload, deleted, timestamp)
			}

		}
		operation = ""

		fmt.Println("Do you want to continue?(Y/N)")
		var choice string
		fmt.Scan(&choice)

		if choice == "N" || choice == "n" {
			break
		}

	}

}
