package records

import (
	"errors"
	"hash/crc32"
	"time"
	"encoding/binary"
	//"fmt"
)

//We are using Big Endian for all the mult byte variables that are going to be saved into the file
//You don't have to worry about the strings since strings are already a sequence of bytes.

//The int value for the checksum is different depending on the data that we provide
//As long as the data is the same between two variables they should always have the same checksum.
func checkSum(data string) uint32 {	
	return crc32.ChecksumIEEE([]byte(data));
}

func createTimeStamp() int64  {
	return time.Now().UnixNano();
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
	b := make([]byte, 0);

	//Use the binary.BigEndian again to convert the binary back to an int64 number
	time := createTimeStamp();
	
	b = binary.BigEndian.AppendUint64(b, uint64(time));
	//fmt.Println(len(b), time);

	//Remember to use BigEndian for reverting it back to a uint32
	b = binary.BigEndian.AppendUint32(b, checkSum(payload));
	
	//We are adding the tombstone value into the record here
	var Tombstone uint8;

	if operation == "DELETE" {
		Tombstone = 1;
	} else {
		Tombstone = 0;
	}

	b = append(b, byte(Tombstone));

	//We are adding the key size here remember that key size and value is uint32
	keyBuff := []byte(key);
	keySize := uint32(len(keyBuff));

	b = binary.BigEndian.AppendUint32(b, keySize);


	//We are adding the payload size here
	valueBuff := []byte(payload);
	valueSize := uint32(len(valueBuff));

	b = binary.BigEndian.AppendUint32(b, valueSize);


	//We are adding the key and payload into the record here
	for _,v := range(keyBuff) {
		b = append(b, v);
	}

	for _,v := range(valueBuff) {
		b = append(b, v);
	}


	return b, nil;
}