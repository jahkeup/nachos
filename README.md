# nachos: a library to FAT-en Mach-O binaries

`nachos` is a library that provides a Mach-O FAT binary creation API.

The library creates a universal, or FAT, binary for a set of given Mach-O binaries, compiled for a number of platforms - currently `ppc`, `x86_64` (or `amd64`), and `arm64`.

## example


``` go
// Read Mach-O headers from files to be nachoed into a universal binary.
files := []fs.File{arm, amd}
exes := []nachos.Executable{}
for i := range files {
    exe, err := nachos.NewFileExe(files[i])
    if err != nil {
        panic(err)
    }
    exes = append(exes, exe)
}

// Build a universal binary from the individual binaries.
res, err := nachos.NewUniversalBinary(exes...)
if err != nil { /* TODO */ }

// Write to an output file, the universal binary file itself.
io.Copy(out, res)
```

