package main

import (
	"fmt"
	"bdev/aof"
)

func main() {
	f, _ := aof.Create("test.dat")
	fmt.Printf("f = %v\n", f)
}

