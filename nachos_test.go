//go:build !darwin
// +build !darwin

package nachos

import (
	"io"
	"io/fs"
	"testing"

	"github.com/jahkeup/nachos/internal/testing/bins"
)

func TestNewUniversalBinary(t *testing.T) {
	type input struct {
		Name string
		File fs.File
		Size int64
	}
	var inputs []*input

	amd64, err := bins.FS.Open(bins.AMD64BinaryName)
	if err != nil {
		t.Fatalf("required bin stub not present: %s", err)
	}
	if stat, err := amd64.Stat(); err != nil {
		t.Fatalf("required bin stub stat errox: %s", err)
	} else {
		inputs = append(inputs, &input{
			Name: stat.Name(),
			File: amd64,
			Size: stat.Size(),
		})
	}

	arm64, err := bins.FS.Open(bins.ARM64BinaryName)
	if err != nil {
		t.Fatalf("required bin stub not present: %s", err)
	}
	if stat, err := arm64.Stat(); err != nil {
		t.Fatalf("required bin stub stat errox: %s", err)
	} else {
		inputs = append(inputs, &input{
			Name: stat.Name(),
			File: amd64,
			Size: stat.Size(),
		})
	}

	var execs []Executable
	var combinedSize int64

	for i := range inputs {
		input := inputs[i]
		ef, err := NewFileExe(input.File)
		if err != nil {
			t.Errorf("cannot use stub %q as file exe: %s", input.Name, err)
		}
		execs = append(execs, ef)
	}

	if len(execs) != 2 {
		t.Fatal("cannot use stubs in test")
	}

	rdc, err := NewUniversalBinary(execs...)
	if err != nil {
		t.Errorf("cannot build universal binary")
		return
	}
	defer rdc.Close()
	n, err := io.Copy(io.Discard, rdc)
	if err != nil {
		t.Fatal(err)
	}
	if n <= combinedSize {
		t.Fatal("produced binary did not wrap executables")
	}
}
