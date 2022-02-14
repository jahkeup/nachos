package nachos

import (
	"bytes"
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
)

// alignment defines the data alignment of each executable within the universal
// binary.
var alignment = struct {
	bits      int
	blocksize int
}{
	// 14 bits is used as its the large of two options: 12 or 14, depending on
	// the platform.
	bits: 14,
	// blocksize is the aligning size an executable will be aligned to.
	blocksize: 1 << 14,
}

const (
	// fat32MaxSize is the maximum size that is addressable by the Fat32 header.
	fat32MaxSize = 1 << 32
)

// fatEndianness is the ByteOrder required for writing out universal binary
// headers.
var fatEndianness = binary.BigEndian

// headerEndianness is the ByteOrder used by Mach-O binaries.
var headerEndianness = binary.LittleEndian

// Executable provides data for making ooey gooey universal binaries.
type Executable interface {
	// header returns the executable's parsed Mach-O header.
	header() macho.FileHeader
	// reader provides the entirety of an executable.
	reader() io.Reader
	// size is the total size of the executable.
	size() int
}

var _ Executable = (*fileExe)(nil)

// fileExe is a wrapper for a file (or any io/fs.File) backed executable.
type fileExe struct {
	file     fs.File
	fileSize int
	hdr      macho.FileHeader
}

func (f *fileExe) header() macho.FileHeader { return f.hdr }
func (f *fileExe) size() int                { return f.fileSize }
func (f *fileExe) reader() io.Reader        { return f.file }

// NewFileExe reads a provided executable file to be built into a universal
// binary (from os.OpenFile or some io/fs file).
func NewFileExe(f fs.File) (*fileExe, error) {
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	hdr := macho.FileHeader{}
	sz := int64(binary.Size(hdr))

	var rd io.Reader
	switch v := f.(type) {
	case io.ReaderAt:
		rd = io.NewSectionReader(v, 0, sz)
	default:
		rd = io.LimitReader(v, sz)
	}

	if err := binary.Read(rd, headerEndianness, &hdr); err != nil {
		return nil, fmt.Errorf("cannot read header: %w", err)
	}

	return &fileExe{
		file:     f,
		fileSize: int(stat.Size()),
		hdr:      hdr,
	}, nil
}

// universalBinary provides a universal binary as a readable stream.
type universalBinary struct {
	io.Reader
	close func() error
}

// Close calls the inner closer to tidy up any resources created.
func (u *universalBinary) Close() error {
	if u.close == nil {
		return nil
	}
	return u.close()
}

// NewUniversalBinary composes a set of executables into a single binary.
func NewUniversalBinary(executables ...Executable) (io.ReadCloser, error) {
	headers := []macho.FileHeader{}
	for _, exe := range executables {
		headers = append(headers, exe.header())
	}

	if hasDuplicateTargets(headers) {
		return nil, errors.New("duplicate targets provided")
	}

	// Prepare collection of executables to be written together. Each needs a
	// calculated and aligned offset which determines where it'll land.
	type entry struct {
		exe    Executable
		header macho.FatArchHeader
		pad    int
	}

	// First executable is offset to the first block's alignment.
	offset := alignment.blocksize

	entries := []entry{}
	for i := range executables {
		exe := executables[i]
		ent := entry{
			exe:    exe,
			header: newFatArchHeader(offset, exe),
		}

		if over := exe.size() % alignment.blocksize; over != 0 {
			ent.pad = alignment.blocksize - over
		}

		// Confirm the file can fit within the addressable bounds.
		offset += exe.size()
		if offset >= fat32MaxSize {
			// TODO: add support for Fat64
			return nil, errors.New("exceeds maximum file size")
		}
		// Pad the next entry.
		offset += ent.pad

		entries = append(entries, ent)
	}

	// Buffer headers to read out from the beginning of the file stream.
	buf := bytes.NewBuffer(nil)
	err := binary.Write(buf, fatEndianness, fatHeader{
		Magic:    macho.MagicFat,
		NFatArch: uint32(len(entries)),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot write binary header: %w", err)
	}

	for _, ent := range entries {
		// Append the entry to the trailing entries following the FatHeader.
		err = binary.Write(buf, fatEndianness, ent.header)
		if err != nil {
			return nil, fmt.Errorf("cannot write binary entry: %w", err)
		}
	}

	stream := []io.Reader{buf}
	pad := alignment.blocksize - buf.Len()
	zero := &zeroReader{}
	for _, ent := range entries {
		// Padding was necessary for alignment, so fill the space with zeros
		// when read.
		if pad != 0 {
			stream = append(stream, io.LimitReader(zero, int64(pad)))
		}
		// Add the file itself to the stream.
		stream = append(stream, ent.exe.reader())
		// set the padding for a "next" element, if needed during next
		// iteration.
		pad = ent.pad
	}

	return &universalBinary{
		Reader: io.MultiReader(stream...),
		close: func() error {
			return nil
		},
	}, nil
}

var _ io.Reader = (*zeroReader)(nil)

// zeroReader provides a zero-cost padding reader. It provides an endless stream
// of zeros. Read carefully :)
type zeroReader struct{}

func (*zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

// newFatArchHeader builds a universal binary target header for the given
// executable. The offset must be provided from the file's origin where the
// executable is expected to be written.
func newFatArchHeader(offset int, exe Executable) macho.FatArchHeader {
	return macho.FatArchHeader{
		Cpu:    exe.header().Cpu,
		SubCpu: exe.header().SubCpu,
		Offset: uint32(offset),
		Size:   uint32(exe.size()),
		Align:  uint32(alignment.bits),
	}
}

// hasDuplicateTargets checks for any duplicate targets in the provided set of
// headers.
func hasDuplicateTargets(headers []macho.FileHeader) bool {
	seen := map[uint64]bool{}
	for _, h := range headers {
		key := uint64(h.Cpu) + uint64(h.SubCpu)<<32
		if dup := seen[key]; dup {
			return true
		} else {
			seen[key] = true
		}
	}

	return false
}

// fatHeader is the type's shape as defined by the Mach-O specification.
type fatHeader struct {
	// Magic is the leading bits of the header that must be magically just so.
	Magic uint32
	// NFatArch is the number of target architecture entries that will follow
	// this in the universal binary's header.
	NFatArch uint32
}
