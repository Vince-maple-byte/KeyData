package file

import (
	"encoding/binary"
	"fmt"
	"os"

	records "github.com/Vince-maple-byte/KeyData/internals/record"
)

type File struct {
	File *os.File
	Size int64
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

//If the file doesn't exist we will just return an empty File object and an network error code since we don't want the
//program to stop running for an invalid file path. This would help us in the future when we accept network calls.
func OpenFile(fileName string) (File,error) {
	//Just created fileEmpty in here so that I can reuse it in other areas where an error can happen
	fileEmpty := File{
			File: nil,
			Size: 0,
			Index: nil,
		}
	file,err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644);

	if err != nil {
		return fileEmpty, err;		
	}

	fileInfo,err := file.Stat();

	if err != nil {
		return fileEmpty, err;
	}

	size := fileInfo.Size();

	f := File{
		File: file,
		Size: size,
		Index: make(map[string]int64),
	}

	fileContents,errf := f.ReadFile(0);

	if errf != nil {
		return fileEmpty, errf;
	}

	f.Index = PopulateMap(fileContents);

	return f, err;
}

//Were are basically going to go through the entire file record by record decoding the each record
//We are also going to checking the checksum for each record
//And if we encounter an incorrect checksum we will stop from there 
// TODO:(We need to figure out a mechanism to get rid of the all values starting from the invalid record to the end of the file. )
//Potential solution: We can make a new file writing the values from the beginning until the records were valid and then just update the file struct to be only pointing to the valid log and deleting the old log
//NOTE: The value part of the map is for the byte offset to locate the record
//TODO: Check if it is the correct checksum
func PopulateMap(fileContents []byte) map[string]int64 {
	m := make(map[string]int64);
	index := 0;

	for index < len(fileContents) {
		keySizeSection := index + 13;

		if keySizeSection >= len(fileContents) {
			break;
		}

		keySize := binary.BigEndian.Uint32(fileContents[keySizeSection:17+index]);

		payloadSizeSection := index + 17;

		if payloadSizeSection >= len(fileContents) {
			break;
		}

		payloadSize := binary.BigEndian.Uint32(fileContents[payloadSizeSection:21+index]);

		key := string(fileContents[index+21:int(keySize)+(index+21)]);

		m[key] = int64(index);

		//The length of the record will always be 21 + keySize + payloadSize
		index += int(21 + keySize + payloadSize);
	}


	return m;
}

//If we get -1 that means that the file was not successfully written.
//FIXME: TODO: Have the WriteFile method update the map with the key and current byte offset 
func (f *File) WriteFile(key, payload, operation string) (amountAdded int, err error){
	var recordBytes []byte;
	recordBytes,err = records.CreateRecord(key,payload,operation);

	if err != nil {
		return -1, err;
	}

	amountAdded,err = f.File.Write(recordBytes);

	if err != nil {
		return -1, err;
	}

	return amountAdded, nil;
}

//TODO: Make the Update Map method so that we can update the byte offset to find the key in the database
func (f *File) UpdateMap(contents []byte) {
	
} 

func (f *File) ReadFile(startingPoint int) ([]byte, error) {
	if startingPoint < 0 {
		panic("Incorrect starting point");
	}
	//f, err := os.OpenFile(fileName,os.O_APPEND|os.O_CREATE,os.ModeAppend);
	file := f.File;

	//We would change the number for seek to be the specific byte offset in the map from the file struct
	_,err := file.Seek(int64(startingPoint), 0);
	if err != nil {
		fmt.Println(err);
		return nil, err;
	}
	

	b := make([]byte, f.Size);

	_,error := file.Read(b);
	
	if error != nil {
		fmt.Println(error);
		return nil, error;
	}

	//fmt.Println(string(f));
	return b, nil;

}
