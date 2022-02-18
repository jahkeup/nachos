package nachos

import (
	"context"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jahkeup/nachos/internal/testing/bins"
)

func TestExecuteUniversalBinary(t *testing.T) {
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

	outdir := t.TempDir()
	out, err := os.OpenFile(filepath.Join(outdir, "universal-bin"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o0755)
	if err != nil {
		t.Fatalf("could not open output file to run with: %s", err)
	}
	_, err = io.Copy(out, rdc)
	out.Close()
	if err != nil {
		t.Fatalf("could not write binary: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, filepath.Join(outdir, "universal-bin"))

	procText, err := cmd.CombinedOutput()
	t.Logf("command output: %s", string(procText))
	if err != nil {
		t.Fatalf("command errored: %s", err)
	}
}
