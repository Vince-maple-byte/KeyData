package main

//"time"
// "github.com/Vince-maple-byte/KeyData/internals/file"
//"github.com/Vince-maple-byte/KeyData/internals/record"
// "unsafe"

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
	


}
