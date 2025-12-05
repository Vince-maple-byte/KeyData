package main

import (
	"fmt"

	"github.com/Vince-maple-byte/KeyData/internals/file"
)

func main() {
	fmt.Println("Hello World");
	fmt.Print("Nice goals\n");
	

	f, err := file.ReadFile("./internals/file/init.txt");

	if(err != nil) {
		fmt.Println(err);
	} else {
		for i := 0; i < len(f); i++ {
			fmt.Println(string(f[i]));
		}
	}
}