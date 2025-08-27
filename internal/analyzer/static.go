package analyzer

import (
	"bytes"
	"debug/elf"
	"fmt"
)

type BinaryInfo struct {
	Format        string        `json:"format"`
	Class         string        `json:"class"`
	Data          string        `json:"data"`
	Version       uint8         `json:"version"`
	OSABI         string        `json:"osabi"`
	ABIVersion    uint8         `json:"abi_version"`
	Type          string        `json:"type"`
	Machine       string        `json:"machine"`
	EntryPoint    uint64        `json:"entry_point"`
	Sections      []SectionInfo `json:"sections"`
	Segments      []SegmentInfo `json:"segments"`
	Symbols       []string      `json:"symbols,omitempty"`
	DynamicNeeded []string      `json:"dynamic_needed,omitempty"`
	SecurityNotes []string      `json:"security_notes,omitempty"`
}

type SectionInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Flags     string `json:"flags"`
	Addr      uint64 `json:"addr"`
	Offset    uint64 `json:"offset"`
	Size      uint64 `json:"size"`
	Entsize   uint64 `json:"entsize"`
	Addralign uint64 `json:"addralign"`
}

type SegmentInfo struct {
	Type   string `json:"type"`
	Flags  string `json:"flags"`
	Vaddr  uint64 `json:"vaddr"`
	Paddr  uint64 `json:"paddr"`
	Filesz uint64 `json:"filesz"`
	Memsz  uint64 `json:"memsz"`
	Align  uint64 `json:"align"`
}

func AnalyzeBinary(fileBytes []byte) (*BinaryInfo, error) {
	fileReader := bytes.NewReader(fileBytes)
	elfFile, err := elf.NewFile(fileReader)
	if err != nil {
		return nil, fmt.Errorf("error elf parsing: %s", err.Error())
	}

	// ELF class and data
	class := "Unknown"
	switch elfFile.Class {
	case elf.ELFCLASS32:
		class = "ELF32"
	case elf.ELFCLASS64:
		class = "ELF64"
	}
	data := "Unknown"
	switch elfFile.Data {
	case elf.ELFDATA2LSB:
		data = "LittleEndian"
	case elf.ELFDATA2MSB:
		data = "BigEndian"
	}

	// OSABI
	osabi := elfFile.OSABI.String()
	// Type
	typ := elfFile.Type.String()
	// Machine
	machine := elfFile.Machine.String()

	// Sections
	var sections []SectionInfo
	for _, sec := range elfFile.Sections {
		sections = append(sections, SectionInfo{
			Name:      sec.Name,
			Type:      sec.Type.String(),
			Flags:     sec.Flags.String(),
			Addr:      sec.Addr,
			Offset:    sec.Offset,
			Size:      sec.Size,
			Entsize:   sec.Entsize,
			Addralign: sec.Addralign,
		})
	}

	// Segments (Program Headers)
	var segments []SegmentInfo
	for _, ph := range elfFile.Progs {
		segments = append(segments, SegmentInfo{
			Type:   ph.Type.String(),
			Flags:  ph.Flags.String(),
			Vaddr:  ph.Vaddr,
			Paddr:  ph.Paddr,
			Filesz: ph.Filesz,
			Memsz:  ph.Memsz,
			Align:  ph.Align,
		})
	}

	// Symbols
	var symbols []string
	if syms, err := elfFile.Symbols(); err == nil {
		for _, sym := range syms {
			symbols = append(symbols, sym.Name)
		}
	}

	// Dynamic needed (shared libraries)
	var dynamicNeeded []string
	if dynSec := elfFile.SectionByType(elf.SHT_DYNAMIC); dynSec != nil {
		dyns, _ := elfFile.DynString(elf.DT_NEEDED)
		dynamicNeeded = append(dynamicNeeded, dyns...)
	}

	// Security notes (e.g., GNU property, RELRO, etc.)
	var securityNotes []string
	for _, sec := range elfFile.Sections {
		if sec.Name == ".note.gnu.property" {
			securityNotes = append(securityNotes, "GNU property present")
		}
		if sec.Name == ".note.gnu.build-id" {
			securityNotes = append(securityNotes, "Build ID present")
		}
		if sec.Name == ".got.plt" {
			securityNotes = append(securityNotes, "GOT/PLT present (dynamic linking)")
		}
	}

	return &BinaryInfo{
		Format:        "ELF",
		Class:         class,
		Data:          data,
		Version:       uint8(elfFile.Version),
		OSABI:         osabi,
		ABIVersion:    elfFile.ABIVersion,
		Type:          typ,
		Machine:       machine,
		EntryPoint:    elfFile.Entry,
		Sections:      sections,
		Segments:      segments,
		Symbols:       symbols,
		DynamicNeeded: dynamicNeeded,
		SecurityNotes: securityNotes,
	}, nil
}
