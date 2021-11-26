package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	log.Printf("I'm built for %q %q, neat!", runtime.GOOS, runtime.GOARCH)
	log.Printf("GOOS:\t%s", runtime.GOOS)
	log.Printf("GOARCH:\t%s", runtime.GOARCH)

	file, err := exec.LookPath("file")
	if err != nil {
		log.Println("I was going to show you what 'file' says, but its not on path.")
	}

	self, err := os.Executable()
	if err != nil {
		log.Fatal("hmm, can't find the executable you ran\n", err)
	}

	cmd := exec.Command(file, self)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	log.Printf("Hi, I'm %q and this is what you host thinks of me:", self)
	if err := cmd.Run(); err != nil {
		log.Fatal("error:", err)
	}
}
