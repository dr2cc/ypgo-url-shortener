package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("hello")
	os.Exit(0) // want "exit call in main function"
}
