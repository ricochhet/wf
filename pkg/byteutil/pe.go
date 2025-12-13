package byteutil

import (
	"debug/pe"
	"encoding/binary"
	"errors"
	"io"
	"os"

	"github.com/ricochhet/pkg/errutil"
)

var (
	ErrSectionHeaderIsSizeZero = errors.New("section header size is 0")
	ErrSectionIsNil            = errors.New("section is nil")
)

// COFFHeader.
// 0x50, 0x45 = PE.
// COFF_START_BYTES_LEN == len(COFFStartBytes).
var COFFStartBytes = []byte{0x50, 0x45, 0x00, 0x00}

const (
	COFFStartBytesLen = 4
	COFFHeaderSize    = 20
)

// OptionalHeader64.
// https://github.com/golang/go/blob/master/src/debug/pe/pe.go
// uint byte size of OptionalHeader64 without magic mumber(2 bytes) or data directory(128 bytes).
// OptionalHeader64 size is 240.
// (110).
var OH64ByteSize = binary.Size(OptionalHeader64X110{})

// DataDirectory.
// 16 entries * 8 bytes / entry.
const (
	DataDirSize      = 128
	DataDirEntrySize = 8
)

// SectionHeader32.
// https://github.com/golang/go/blob/master/src/debug/pe/section.go
// uint byte size of SectionHeader32 without name(8 bytes) or characteristics(4 bytes).
// (28).
var SH32ByteSize = binary.Size(SectionHeader32X28{})

const (
	SH32EntrySize           = 64
	SH32NameSize            = 8
	SH32CharacteristicsSize = 4
)

type PE struct{}

// Data structure.
type Data struct {
	Bytes []byte
	PE    pe.File
}

// Section structure (.ooa).
type Section struct {
	ContentID   string
	OEP         uint64
	EncBlocks   []EncBlock
	ImageBase   uint64
	SizeOfImage uint32
	ImportDir   DataDir
	IATDir      DataDir
	RelocDir    DataDir
}

// Import structure (.ooa).
type Import struct {
	Characteristics uint32
	Timedatestamp   uint32
	ForwarderChain  uint32
	Name            uint32
	FThunk          uint32
}

// Thunk structure (.ooa).
type Thunk struct {
	Function uint32
	DataAddr uint32
}

// DataDir structure (.ooa).
type DataDir struct {
	VA   uint32
	Size uint32
}

// EncBlock structure (.ooa).
type EncBlock struct {
	VA          uint32
	RawSize     uint32
	VirtualSize uint32
	Unk         uint32
	CRC         uint32
	Unk2        uint32
	CRC2        uint32
	Pad         uint32
	FileOffset  uint32
	Pad2        uint64
	Pad3        uint32
}

type OptionalHeader64X110 struct {
	MajorLinkerVersion          uint8
	MinorLinkerVersion          uint8
	SizeOfCode                  uint32
	SizeOfInitializedData       uint32
	SizeOfUninitializedData     uint32
	AddressOfEntryPoint         uint32
	BaseOfCode                  uint32
	ImageBase                   uint64
	SectionAlignment            uint32
	FileAlignment               uint32
	MajorOperatingSystemVersion uint16
	MinorOperatingSystemVersion uint16
	MajorImageVersion           uint16
	MinorImageVersion           uint16
	MajorSubsystemVersion       uint16
	MinorSubsystemVersion       uint16
	Win32VersionValue           uint32
	SizeOfImage                 uint32
	SizeOfHeaders               uint32
	CheckSum                    uint32
	Subsystem                   uint16
	DllCharacteristics          uint16
	SizeOfStackReserve          uint64
	SizeOfStackCommit           uint64
	SizeOfHeapReserve           uint64
	SizeOfHeapCommit            uint64
	LoaderFlags                 uint32
	NumberOfRvaAndSizes         uint32
}

type SectionHeader32X28 struct {
	VirtualSize          uint32
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLineNumbers uint32
	NumberOfRelocations  uint16
	NumberOfLineNumbers  uint16
}

var errRVALessThanOne = errors.New("rva is equal to '-1'")

// Open opens a file at the specified path and returns Data.
func (p *PE) Open(path string) (*Data, error) {
	data := &Data{}

	file, err := os.Open(path)
	if err != nil {
		return nil, errutil.New("os.Open", err)
	}

	pe, err := pe.NewFile(file)
	if err != nil {
		file.Close()
		return nil, errutil.New("pe.newFile", err)
	}

	b, err := io.ReadAll(file)
	if err != nil {
		file.Close()
		return nil, errutil.New("io.ReadAll", err)
	}

	data.Bytes = b
	data.PE = *pe

	return data, nil
}

// COFFHeaderOffset searches for the specified bytes in the file.
func (p *PE) COFFHeaderOffset(data []byte) (int, error) {
	offset, err := Index(data, COFFStartBytes)
	if err != nil {
		return -1, errutil.WithFrame(err)
	}

	return offset, nil
}

// DDBytes reads the data directory entry at the specified offset.
func (p *PE) DDBytes(data []byte) ([]byte, error) {
	offset, err := p.COFFHeaderOffset(data)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	start := offset + COFFStartBytesLen + COFFHeaderSize + OH64ByteSize
	end := offset + COFFStartBytesLen + COFFHeaderSize + OH64ByteSize + DataDirSize

	return data[start:end], nil
}

// DDEntryOffset reads the offset of the data directory entry at the specified address.
func (p *PE) DDEntryOffset(data []byte, addr, size uint32) (int, error) {
	dir, err := p.DDBytes(data)
	if err != nil {
		return -1, errutil.New("ReadDDBytes", err)
	}

	b := make([]byte, DataDirEntrySize)
	binary.LittleEndian.PutUint32(b[:4], addr)
	binary.LittleEndian.PutUint32(b[4:], size)
	rva, err := Index(dir, b)

	if err != nil || rva == -1 {
		if err == nil {
			return -1, errRVALessThanOne
		}

		return -1, errutil.New("Index", err)
	}

	offset, err := p.COFFHeaderOffset(data)
	if err != nil {
		return -1, errutil.New("ReadCOFFHeaderOffset", err)
	}

	return offset + COFFStartBytesLen + COFFHeaderSize + OH64ByteSize + rva, nil
}

// SHSize reads the section header bytes at the specified offset.
func (p *PE) SHSize(file pe.File) (int, error) {
	size := len(file.Sections) * SH32EntrySize

	if size == 0 {
		return -1, ErrSectionHeaderIsSizeZero
	}

	return size, nil
}

// SHBytes reads the section header entry at the specified offset.
func (p *PE) SHBytes(data []byte, size int) ([]byte, error) {
	offset, err := p.COFFHeaderOffset(data)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	index := offset + COFFStartBytesLen + COFFHeaderSize + OH64ByteSize + DataDirSize

	return data[index : index+size], nil
}

// SHEntryOffset reads the offset of the specified section header entry.
func (p *PE) SHEntryOffset(data []byte, address int) (int, error) {
	offset, err := p.COFFHeaderOffset(data)
	if err != nil {
		return -1, errutil.WithFrame(err)
	}

	return offset + COFFStartBytesLen + COFFHeaderSize + OH64ByteSize + DataDirSize + address, nil
}

// SectionBytes reads the specified section bytes.
func (p *PE) SectionBytes(file *Data, sectionVirtualAddress, sectionSize uint32) ([]byte, error) {
	var s *pe.Section

	for _, section := range file.PE.Sections {
		if sectionVirtualAddress >= section.VirtualAddress &&
			sectionVirtualAddress < section.VirtualAddress+section.Size {
			s = section
			break
		}
	}

	if s == nil {
		return nil, ErrSectionIsNil
	}

	offset := sectionVirtualAddress - s.VirtualAddress + s.Offset

	return file.Bytes[offset : offset+sectionSize], nil
}

// Import reads the import section.
func (p *PE) Import(reader io.Reader) (Import, error) {
	var d Import

	err := binary.Read(reader, binary.LittleEndian, &d)

	return d, errutil.WithFrame(err)
}

// Thunk reads the thunk section.
func (p *PE) Thunk(reader io.Reader) (Thunk, error) {
	var d Thunk

	err := binary.Read(reader, binary.LittleEndian, &d)

	return d, errutil.WithFrame(err)
}

// DataDir reads the encryption block.
func (p *PE) DataDir(reader io.Reader) (DataDir, error) {
	var d DataDir

	err := binary.Read(reader, binary.LittleEndian, &d)

	return d, errutil.WithFrame(err)
}

// EncBlock reads the encryption block.
func (p *PE) EncBlock(reader io.Reader) (EncBlock, error) {
	var d EncBlock

	err := binary.Read(reader, binary.LittleEndian, &d)

	return d, errutil.WithFrame(err)
}
