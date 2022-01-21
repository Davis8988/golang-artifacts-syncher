package main

import (
	"flag"
	"fmt"
)

var (
	file   = flag.String("file", "", "file-name")
	count  = flag.Int("count", 2, "count params")
	repeat = flag.Bool("repeat", false, "Repeat execution")
)

func main() {
	flag.Parse()

	fmt.Println("file name: ", *file)
	fmt.Println("count: ", *count)
	fmt.Println("repeat: ", *repeat)

	fmt.Println("Hello World")
}
