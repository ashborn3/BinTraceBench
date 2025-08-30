package sandbox

import (
	"bytes"
	"debug/elf"
	"fmt"
	"os"
	"path/filepath"
)

const (
	MaxFileSize = 50 * 1024 * 1024 // 50MB max file size
)

func ValidateBinary(data []byte) error {
	return ValidateBinaryWithConfig(data, DefaultConfig())
}

func ValidateBinaryWithConfig(data []byte, config *Config) error {
	if len(data) == 0 {
		return fmt.Errorf("empty file")
	}

	if int64(len(data)) > config.MaxFileSize {
		return fmt.Errorf("file too large: %d bytes (max %d)", len(data), config.MaxFileSize)
	}

	// Check if it's an ELF file
	if len(data) < 4 {
		return fmt.Errorf("file too small to be a valid binary")
	}

	// ELF magic number check
	if !bytes.HasPrefix(data, []byte{0x7f, 'E', 'L', 'F'}) {
		return fmt.Errorf("not a valid ELF binary")
	}

	// Try to parse as ELF to ensure it's valid
	reader := bytes.NewReader(data)
	_, err := elf.NewFile(reader)
	if err != nil {
		return fmt.Errorf("invalid ELF file: %v", err)
	}

	return nil
}

func CreateSecureTempFile(data []byte, prefix string) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "bintracebench-sandbox-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	tempFile := filepath.Join(tempDir, prefix+"-binary")

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// restricted permissions
	err = os.WriteFile(tempFile, data, 0700) // Owner read/write/execute only
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to write temp file: %v", err)
	}

	return tempFile, cleanup, nil
}

func CreateSecureTempFileWithConfig(data []byte, prefix string, config *Config) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", config.TempDirPrefix+"-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	tempFile := filepath.Join(tempDir, prefix+"-binary")

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	err = os.WriteFile(tempFile, data, 0700) // Owner read/write/execute only
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to write temp file: %v", err)
	}

	return tempFile, cleanup, nil
}
