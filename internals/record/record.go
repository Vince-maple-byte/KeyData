package records

import (
	"errors"
	"hash/crc32"
)

//The int value for the checksum is different depending on the data that we provide
//As long as the data is the same between two variables they should always have the same checksum.
func CheckSum(data string) uint32 {	
	return crc32.ChecksumIEEE([]byte(data));
}

//data == payload
//FIXME: Make sure to finish with this function using this format 
//Timestamp | CRC32 error checksum| Tombstone (It's one byte long; basically 0 and 1 to determine whether this is a deleted key or not) | Key Size | Payload(Value) Size | Key Value |  Payload
func CreateRecord(key, payload, operation string) ([]byte, error) {
	if key == "" {
		return nil, errors.New("Invalid Key");
	}

	//Don't need to check for the GET operation since that should automatically go to the ReadFile section.
	if operation == "" || (operation != "PUT" && operation != "DELETE") {
		return nil, errors.New("Invalid Operation");
	}

	//Make sure to use time.AppendBinary https://pkg.go.dev/time#example-Time.AppendBinary
	b := make([]byte, 10);

	return b, nil;
}