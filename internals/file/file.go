package file

import (
	"fmt"
	"os"

	records "github.com/Vince-maple-byte/KeyData/internals/record"
)

type File struct {
	file *os.File
	size int64
	index map[string]int64
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

//FIXME: We need to read from the entire file and update the map indexes
func OpenFile(fileName string) (File,error) {
	file,err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644);

	if err != nil {
		panic(err.Error());		
	}

	fileInfo,_ := file.Stat();

	size := fileInfo.Size();

	f := File{
		file: file,
		size: size,
		index: make(map[string]int64),
	}

	return f, err;
}


//If we get -1 that means that the file was not successfully written.
func (f *File) WriteFile(key, payload, operation string) (amountAdded int, err error){
	var recordBytes []byte;
	recordBytes,err = records.CreateRecord(key,payload,operation);

	if err != nil {
		return -1, err;
	}

	amountAdded,err = f.file.Write(recordBytes);

	if err != nil {
		return -1, err;
	}

	return amountAdded, nil;
}

func (f *File) ReadFile() ([]byte, error) {
	//f, err := os.OpenFile(fileName,os.O_APPEND|os.O_CREATE,os.ModeAppend);
	file := f.file;

	//We would change the number for seek to be the specific byte offset in the map from the file struct
	_,err := file.Seek(0, 0);
	if err != nil {
		fmt.Println(err);
		return nil, err;
	}
	

	b := make([]byte, f.size);

	_,error := file.Read(b);
	
	if error != nil {
		fmt.Println(error);
		return nil, error;
	}

	//fmt.Println(string(f));
	return b, nil;

}