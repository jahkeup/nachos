package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Printf("GOOS=%s\n", runtime.GOOS)
	fmt.Printf("GOARCH=%s\n", runtime.GOARCH)
}
