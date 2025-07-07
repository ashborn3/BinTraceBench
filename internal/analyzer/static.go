package analyzer

import (
	"bytes"
	"debug/elf"
	"fmt"
)

type BinaryInfo struct {
	Format       string   `json:"format"`
	Architecture string   `json:"architecture"`
	EntryPoint   uint64   `json:"entry_point"`
	Sections     []string `json:"sections"`
	Symbols      []string `json:"symbols,omitempty"`
}

func AnalyzeBinary(fileBytes []byte) (*BinaryInfo, error) {
	fileReader := bytes.NewReader(fileBytes)
	elfFile, err := elf.NewFile(fileReader)
	if err != nil {
		return nil, fmt.Errorf("error elf parsing: %s", err.Error())
	}

	sections := []string{}
	for _, sec := range elfFile.Sections {
		sections = append(sections, sec.Name)
	}

	symbols := []string{}
	if syms, err := elfFile.Symbols(); err == nil {
		for _, sym := range syms {
			symbols = append(symbols, sym.Name)
		}
	}

	return &BinaryInfo{
		Format:       "ELF",
		Architecture: elfFile.FileHeader.Machine.String(),
		EntryPoint:   elfFile.Entry,
		Sections:     sections,
		Symbols:      symbols,
	}, nil

}
