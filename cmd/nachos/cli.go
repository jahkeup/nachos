package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/jahkeup/nachos"
)

type CLI struct {
	Output         string
	Inputs         []string
	ArchFlagValues inputSlice
}

type inputSlice []string

// Return the recorded values' slice as a string.
func (i *inputSlice) String() string {
	return fmt.Sprintf("%q", []string(*i))
}

// Set records the argument in a string slice.
func (i *inputSlice) Set(s string) string {
	*i = append(*i, s)
	return ""
}

// GetOutputName provides the effective output path name. If an output was not
// given as an argument then a guessed name based on the inputs is derived, if
// possible.
func (cli *CLI) GetOutputName() string {
	if cli.Output != "" {
		return cli.Output
	}

	if len(cli.Inputs) != 2 {
		return ""
	}

	return guessOverlapName(cli.Inputs[0], cli.Inputs[1])
}

func NewCLI() (*CLI, *flag.FlagSet) {
	cli := &CLI{}

	fs := flag.NewFlagSet("nachos", flag.ContinueOnError)
	fs.StringVar(&cli.Output, "output", "", "Output executable name (uses a guessed name if not provided)")
	// this let's common CLI invocations *look* like lipo without much effort,
	// the arch argument is simply ignored.
	fs.String("arch", "", "for compatibility with lipo CLI")

	return cli, fs
}

func guessOverlapName(s1, s2 string) string {
	name1 := cleanComparisonNames(s1)
	name2 := cleanComparisonNames(s2)

	var (
		commonLen int
	)
	if len(name2) < len(name1) {
		commonLen = len(name2)
	} else {
		commonLen = len(name1)
	}

	var (
		fromFront string
		frontDone bool
	)
	var (
		fromEnd string
		endDone bool
	)
	for i := 0; i < commonLen; i++ {
		if name1[i] == name2[i] {
			if !frontDone {
				fromFront += string(name1[i])
			}
		} else {
			frontDone = true
		}

		if name1[len(name1)-1-i] == name2[len(name2)-1-i] {
			if !endDone {
				fromEnd = string(name1[len(name2)-1-i]) + fromEnd
			}
		} else {
			endDone = true
		}
	}

	var ret string
	if fromFront != "" && fromFront != s1 && fromFront != s2 {
		ret = fromFront
	} else if fromEnd != "" && fromEnd != s1 && fromEnd != s2 {
		ret = fromEnd
	} else {
		return ""
	}

	// use only the basename of a path if that's what we have.
	ret = filepath.Base(ret)
	// trim any cruft on the end
	ret = strings.Trim(ret, "-_")

	return ret
}

func cleanComparisonNames(s string) string {
	if s == "" {
		return ""
	}

	suffices := []string{
		"amd64", "x86_64",
		"arm64", "aarch64",
		"darwin",
	}
	prefixes := append(suffices, []string{
		"darwin", // leading element?
	}...)

	name := filepath.Base(s)
	for _, sfx := range suffices {
		name = strings.TrimSuffix(name, sfx)
		name = strings.TrimRight(name, "-_")
	}

	for _, pfx := range prefixes {
		name = strings.TrimPrefix(name, pfx)
		name = strings.TrimLeft(name, "-_")
	}

	return name
}

const minimumInputs = 2

func assertValidInputs(fsys fs.FS, is []string) error {
	if len(is) < minimumInputs {
		return fmt.Errorf("given %d inputs but %d required", len(is), minimumInputs)
	}

	for _, inputPath := range is {
		stat, err := fs.Stat(fsys, inputPath)
		if err != nil {
			return fmt.Errorf("unable to use input %q: %w", inputPath, err)
		}
		if !stat.Mode().IsRegular() {
			return fmt.Errorf("unable to use input %q: not a regular file", inputPath)
		}
	}

	return nil
}

func assertValidOutput(fsys fs.FS, cli *CLI) error {
	if cli.GetOutputName() == "" {
		return fmt.Errorf("could not guess output name for %q: please provide one with -output", cli.Inputs)
	}
	return nil
}

func run(fsys fs.FS, cli *CLI) (io.ReadCloser, error) {
	var exes []nachos.Executable

	for _, inputPath := range cli.Inputs {
		f, err := fsys.Open(inputPath)
		if err != nil {
			return nil, fmt.Errorf("cannot open input %q: %w", inputPath, err)
		}
		exe, err := nachos.NewFileExe(f)
		if err != nil {
			return nil, fmt.Errorf("cannot use input %q: %w", inputPath, err)
		}
		exes = append(exes, exe)
	}

	return nachos.NewUniversalBinary(exes...)
}
