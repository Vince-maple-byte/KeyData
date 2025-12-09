package main

import (
	"fmt"
)

func main() {
	// fmt.Println("Hello World");
	// fmt.Print("Nice goals\n");
	

	// f, err := file.OpenFile("./internals/file/init.txt");

	// if(err != nil) {
	// 	fmt.Println(err);
	// } else {
	// 	b,_ := f.ReadFile();
	// 	for i := 0; i < len(b); i++ {
	// 		fmt.Print(string(b[i]));
	// 	}
	// }

	ch := make(chan string);

	ch <- "Hello";

	fmt.Println(<-ch);
	// ch <- "World";
	// fmt.Println(<-ch);
	
}