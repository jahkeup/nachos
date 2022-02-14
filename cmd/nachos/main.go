package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

var errUsage = errors.New("usage error")
var errRun = errors.New("run error")

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	err := _main()
	switch {
	case err == nil:
		os.Exit(0)
	case errors.Is(err, errUsage):
		os.Exit(2)
	case errors.Is(err, errRun):
		os.Exit(1)
	default:
		os.Exit(3)
	}
}

func _main() error {
	cli, fs := NewCLI()
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Println("invalid arguments")
		fs.Usage()
		return errUsage
	}
	cli.Inputs = fs.Args()

	fsys := os.DirFS(".")

	if err := assertValidInputs(fsys, cli.Inputs); err != nil {
		log.Printf("invalid inputs: %s", err)
		fs.Usage()
		return fmt.Errorf("assertValidInputs: %w", errRun)
	}
	if err := assertValidOutput(fsys, cli); err != nil {
		log.Printf("invalid output: %s", err)
		fs.Usage()
		return errUsage
	}

	log.Printf("building universal binary from %q", cli.Inputs)
	rdc, err := run(fsys, cli)
	if err != nil {
		log.Printf("failed to build: %s", err)
		return errRun
	}
	defer rdc.Close()

	outputName := cli.GetOutputName()
	log.Printf("writing output to %q", outputName)
	output, err := os.OpenFile(cli.GetOutputName(), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o0755)
	if err != nil {
		log.Printf("failed to write output file: %s", err)
		return errRun
	}
	defer output.Close()
	io.Copy(output, rdc)

	return nil
}
