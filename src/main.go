package main

import (
	"flag"
	"os"
	"fmt"
)

var (
	file   = flag.String("file", "", "file-name")
	count  = flag.Int("count", 2, "count params")
	repeat = flag.Bool("repeat", false, "Repeat execution")
)

func main() {
	flag.Parse()
	var envVal = os.Getenv("david")
	fmt.Println("file name: ", *file)
	fmt.Println("count: ", *count)
	fmt.Println("repeat: ", *repeat)
	fmt.Println("envVal: '{}'", envVal)

	fmt.Println("Hello World")
}
