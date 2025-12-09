package main

import (
	"fmt"

	"github.com/Vince-maple-byte/KeyData/internals/file"
)

/*

TODO:

Make the Write method

Make a method to structure the record that the user will type to the binary encoding that we want

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
		b,_ := f.ReadFile();
		for i := 0; i < len(b); i++ {
			fmt.Print(string(b[i]));
		}
	}

	// ch := make(chan string);

	// ch <- "Hello";

	// fmt.Println(<-ch);
	// ch <- "World";
	// fmt.Println(<-ch);
	
}