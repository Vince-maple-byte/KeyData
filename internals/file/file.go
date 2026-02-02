package file

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"time"

	records "github.com/Vince-maple-byte/KeyData/internals/record"
)

type File struct {
	File  *os.File
	Size  int64
	Index map[string]int64
}

/*This is how we are opening up the file, reading, and writing to the file
We are using os.OpenFile that returns an os.File pointer that allows us to use methods such as
file.Seek(offset int64, whence int) which sets the byte offset in our program to point to where in the file each record is located,
 and whence just specifies if we should start in the beginning of the file, the current offset, or the end of the file

file.Read(b []byte) which would read from the file starting at the current byte offset the os.File is pointing to,
and then adding n bytes where n is the underlining size of the byte slice (remember a slice is just a dynamic array in go);

file.OpenFile() allows us to open the file and return the *os.File, with this we can run the methods mentioned above,
 and pass in some arguments such as os.O_APPEND to specify what we can do in the file;
 the prem argument is only valid if the OpenFile is creating the file for the first time
 and specifies the special bits(setuid, sticky bit), the owner permission, group permissions, and the global permissions in this format
 0664
 Where for the last three digits
 4 = read permission
 2 = write permission
 1 = execute permission
 0 = no permission

*/

// If the file doesn't exist we will just return an empty File object and an network error code since we don't want the
// program to stop running for an invalid file path. This would help us in the future when we accept network calls.
func OpenFile(fileName string) (File, error) {
	//Just created fileEmpty in here so that I can reuse it in other areas where an error can happen
	fileEmpty := File{
		File:  nil,
		Size:  0,
		Index: nil,
	}
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		return fileEmpty, err
	}

	fileInfo, err := file.Stat()

	if err != nil {
		return fileEmpty, err
	}

	size := fileInfo.Size()

	f := File{
		File:  file,
		Size:  size,
		Index: make(map[string]int64),
	}

	if size > 0 {
		fileContents, errf := f.ReadFile(0, int(f.Size))

		if errf != nil {
			return fileEmpty, errf
		}

		f.Index = f.PopulateMap(fileContents)
	}

	return f, err
}

// Were are basically going to go through the entire file record by record decoding the each record
// We are also going to checking the checksum for each record
// And if we encounter an incorrect checksum we will stop from there
// Potential solution: We can make a new file writing the values from the beginning until the records were valid and then just update the file struct to be only pointing to the valid log and deleting the old log
// NOTE: The value part of the map is for the byte offset to locate the record
func (f *File) PopulateMap(fileContents []byte) map[string]int64 {
	m := make(map[string]int64)
	index := 0

	for index < len(fileContents) {

		keySizeSection := index + 13

		if keySizeSection >= len(fileContents) {
			break
		}

		keySize := binary.BigEndian.Uint32(fileContents[keySizeSection : index+17])

		payloadSizeSection := index + 17

		if payloadSizeSection >= len(fileContents) {
			break
		}

		payloadSize := binary.BigEndian.Uint32(fileContents[payloadSizeSection : index+21])

		payloadByte := fileContents[int(keySize)+(index+21) : int(keySize)+(index+21)+int(payloadSize)]

		checksum := binary.BigEndian.Uint32(fileContents[index+8 : index+12])

		var validChecksum bool = records.ChecksumChecker(payloadByte, checksum)

		if validChecksum {
			key := string(fileContents[index+21 : int(keySize)+(index+21)])

			m[key] = int64(index)

			//The length of the record will always be 21 + keySize + payloadSize
			index += int(21 + keySize + payloadSize)
		} else {
			f.trimFile(index)
			break
		}

	}

	return m
}

// This method gets executed when a record in the database is found out to be invalid through checksum.
// When this is the case we just take all of the values before the invalid one and save it to a new file.
func (f *File) trimFile(endPoint int) {
	newFileContents, err := f.ReadFile(0, endPoint)

	if err != nil {
		panic("Something went wrong with trying to read the file")
	}

	fileName := f.File.Name()

	err = os.Remove(fileName)

	if err != nil {
		panic("Unable to delete the file " + fileName)
	}

	f.File, err = os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		panic("Unable to open/create file " + fileName)
	}

	size, err := f.File.Write(newFileContents)

	if err != nil {
		panic("Was not able to write into " + fileName)
	}

	f.Size = int64(size)

}

// If we get -1 that means that the file was not successfully written.
func (f *File) PutContents(key, payload, operation string) (amountAdded int, err error) {
	var recordBytes []byte
	recordBytes, err = records.CreateRecord(key, payload, operation)

	if err != nil {
		return -1, err
	}

	amountAdded, err = f.File.Write(recordBytes)

	if err != nil {
		return -1, err
	}

	f.UpdateMap(recordBytes)

	f.Size += int64(amountAdded)

	return amountAdded, nil
}

// In the case where the key is deleted, we can just return the offset of where the deleted record exists just like a regular read
// and from there we can just read from the tombstone whether it was deleted or not
func (f *File) UpdateMap(contents []byte) {
	var offset int64 = f.Size

	keySize := binary.BigEndian.Uint32(contents[13:17])

	key := string(contents[21 : keySize+21])

	f.Index[key] = offset
}

// We can now read from the file from the starting byte offset to the ending byte
func (f *File) ReadFile(startingPoint, endingPoint int) ([]byte, error) {
	if startingPoint < 0 {
		panic("Incorrect starting point")
	}

	file := f.File

	//We would change the number for seek to be the specific byte offset in the map from the file struct
	_, err := file.Seek(int64(startingPoint), 0)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	b := make([]byte, endingPoint)

	_, error := file.Read(b)

	if error != nil {
		fmt.Println(error)
		return nil, error
	}

	//fmt.Println(string(f));
	return b, nil

}

func (f *File) GetContents(key string) (deleted bool, payload string, timestamp time.Time, err error) {
	keyOffset := f.Index[key]

	content, e := f.ReadFile(int(keyOffset), int(f.Size))

	if e != nil {
		err = e
		return
	}

	timeByte := content[:8]
	tombstone := content[12]
	keySize := binary.BigEndian.Uint32(content[13:17])
	payloadSize := binary.BigEndian.Uint32(content[17:21])
	timestamp = time.Unix(0, int64(binary.BigEndian.Uint64(timeByte)))

	if tombstone == 1 {
		deleted = true
		payload = ""
		return
	} else {
		deleted = false
	}

	if err != nil {
		return
	}

	keySaved := string(content[21 : keySize+21])

	if key != keySaved {
		err = errors.New("Invalid key given")
		return
	}

	payload = string(content[keySize+21 : (keySize+21)+payloadSize])

	return
}
