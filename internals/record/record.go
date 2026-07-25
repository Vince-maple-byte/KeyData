package record

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"time"

	"github.com/ccoveille/go-safecast/v2"
)

//We are using Big Endian for all the mult byte variables that are going to be saved into the file
//You don't have to worry about the strings since strings are already a sequence of bytes.

// The int value for the checksum is different depending on the data that we provide
// As long as the data is the same between two variables they should always have the same checksum.

type Content struct {
	Timestamp   time.Time
	Checksum    uint32
	Tombstone   uint8
	Keysize     uint32
	Payloadsize uint32
	Key         string
	Payload     string
}

func checkSum(key, payload []byte, timestamp uint64) uint32 {
	buf := make([]byte, 16+len(key)+len(payload))
	keySizeConv, _ := safecast.Convert[uint32](len(key))
	payloadSizeConv, _ := safecast.Convert[uint32](len(payload))
	binary.BigEndian.PutUint64(buf[0:8], timestamp)
	binary.BigEndian.PutUint32(buf[8:12], keySizeConv)
	binary.BigEndian.PutUint32(buf[12:16], payloadSizeConv)

	copy(buf[16:], key)
	copy(buf[16+len(key):], payload)

	return crc32.ChecksumIEEE(buf)
}

func ChecksumChecker(key, payload []byte, timestamp uint64, checksum uint32) bool {
	return checkSum(key, payload, timestamp) == checksum
}

func createTimeStamp() int64 {
	return time.Now().UnixNano()
}

// CreateRecord:
// data == payload
// Timestamp | CRC32 error checksum| Tombstone (It's one byte long; basically 0 and 1 to determine whether this is a deleted key or not) | Key Size | Payload(Value) Size | Key Value |  Payload
func CreateRecord(key, payload, operation string) ([]byte, error) {
	if key == "" {
		return nil, errors.New("invalid Key")
	}

	//Don't need to check for the GET operation since that should automatically go to the ReadFile section.
	if operation == "" || (operation != "PUT" && operation != "DELETE") {
		return nil, fmt.Errorf("invalid operation: %v", operation)
	}

	//Make sure to use time.AppendBinary https://pkg.go.dev/time#example-Time.AppendBinary
	b := make([]byte, 0)

	//Use the binary.BigEndian again to convert the binary back to an int64 number
	time := createTimeStamp()
	timeConv, err := safecast.Convert[uint64](time)

	if err != nil {
		return nil, err
	}

	b = binary.BigEndian.AppendUint64(b, timeConv)
	//fmt.Println(len(b), time);

	//Remember to use BigEndian for reverting it back to a uint32
	b = binary.BigEndian.AppendUint32(b, checkSum([]byte(key), []byte(payload), timeConv))

	//We are adding the tombstone value into the record here
	var Tombstone uint8

	if operation == "DELETE" {
		Tombstone = 1
		payload = ""
	} else {
		Tombstone = 0
	}

	b = append(b, byte(Tombstone))

	//We are adding the key size here remember that key size and value is uint32
	keyBuff := []byte(key)
	keySize := len(keyBuff)
	keySizeConv, err := safecast.Convert[uint32](keySize)

	if err != nil {
		return nil, err
	}

	b = binary.BigEndian.AppendUint32(b, keySizeConv)

	//We are adding the payload size here
	valueBuff := []byte(payload)
	valueSize := len(valueBuff)
	valueSizeConv, err := safecast.Convert[uint32](valueSize)

	if err != nil {
		return nil, err
	}

	b = binary.BigEndian.AppendUint32(b, valueSizeConv)

	//We are adding the key and payload into the record here

	b = append(b, keyBuff...)
	b = append(b, valueBuff...)

	return b, nil
}

func GetContents(content []byte) (result Content) {
	timeByte := content[:8]
	timestamp64, _ := safecast.Convert[int64](binary.BigEndian.Uint64(timeByte))

	result.Timestamp = time.Unix(0, timestamp64)

	result.Checksum = binary.BigEndian.Uint32(content[8:12])

	result.Tombstone = uint8(content[12])

	result.Keysize = binary.BigEndian.Uint32(content[13:17])
	result.Payloadsize = binary.BigEndian.Uint32(content[17:21])

	result.Key = string(content[21 : result.Keysize+21])
	result.Payload = string(content[result.Keysize+21 : (result.Keysize+21)+result.Payloadsize])

	return
}
