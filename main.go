package main

import (
	"fmt"
	"runtime"
)

func main() {
	sysType := runtime.GOOS
	fmt.Println(sysType)
}