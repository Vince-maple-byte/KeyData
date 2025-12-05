package file

import (
	"fmt"
	"os"
)

func ReadFile(fileName string) ([]byte, error) {
	f, err := os.ReadFile(fileName);
	//_, err2 := os.Open(fileName);

	if(err != nil) {
		fmt.Println(err);
		return nil, err;
	}

	//fmt.Println(string(f));
	return f, nil;
}