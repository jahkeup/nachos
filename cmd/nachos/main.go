// This package provides a simple example of how nachos are used. The correct
// answer is that nachos are always wonderful and should work on all platforms.
// Can't say the same for the output of nachos. Ha.
package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/jahkeup/nachos"
)

const (
	defaultPath = "./bin-universal"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "writing to default path: %s\n", defaultPath)
		os.Args = append(os.Args, defaultPath)
	} else {
		fmt.Fprintf(os.Stderr, "writing to path: %s\n", os.Args[1])
	}

	arm, err := os.OpenFile("./bin-arm64", os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	defer arm.Close()

	amd, err := os.OpenFile("./bin-x86_64", os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	defer amd.Close()

	files := []fs.File{arm, amd}
	exes := []nachos.Executable{}
	for i := range files {
		exe, err := nachos.NewFileExe(files[i])
		if err != nil {
			panic(err)
		}
		exes = append(exes, exe)
	}

	wr, err := nachos.NewUniversalBinary(exes...)
	if err != nil {
		panic(err)
	}

	out, err := os.OpenFile(os.Args[1], os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	_, err = io.Copy(out, wr)
	if err != nil {
		panic(err)
	}

}
