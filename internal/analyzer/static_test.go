package analyzer

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAnalyzeBinary_ELF(t *testing.T) {
	// Open a known ELF binary
	data, err := os.ReadFile("/bin/ls")
	if err != nil {
		t.Fatalf("failed to read test binary: %v", err)
	}

	// Analyze the binary
	info, err := AnalyzeBinary(data)
	if err != nil {
		t.Fatalf("AnalyzeBinary failed: %v", err)
	}

	// Basic checks
	if info.Format != "ELF" {
		t.Errorf("expected ELF format, got %s", info.Format)
	}
	if info.Architecture == "" {
		t.Errorf("Architecture is empty")
	}
	if info.EntryPoint == 0 {
		t.Errorf("EntryPoint is zero")
	}
	if len(info.Sections) == 0 {
		t.Errorf("No sections found")
	}

	b, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		t.Logf("Failed to marshal info: %v", err)
	} else {
		t.Logf("Analyzed ELF binary:\n%s", string(b))
	}
}
