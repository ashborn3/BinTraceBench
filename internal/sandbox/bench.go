package sandbox

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ashborn3/BinTraceBench/internal/syscalls"
)

type BenchResult struct {
	CGroup       string                  `json:"c_group"`
	InvocationID string                  `json:"invocation_id"` // Question: what's the use for this?
	ExitCode     int                     `json:"exit_code"`
	RuntimeMS    int64                   `json:"runtime_ms"`
	Success      bool                    `json:"success"`
	ErrorMessage string                  `json:"error_message,omitempty"`
	Syscalls     []syscalls.SyscallEntry `json:"syscalls,omitempty"`
}

func RunBenchmark(filebytes []byte) (*BenchResult, error) {
	if err := ValidateBinary(filebytes); err != nil {
		return nil, fmt.Errorf("binary validation failed: %v", err)
	}

	tmpPath, cleanup, err := CreateSecureTempFile(filebytes, "benchmark")
	if err != nil {
		return nil, err
	}
	defer cleanup()

	cmd := exec.Command("unshare",
		"--mount", "--uts", "--ipc", "--net", "--pid", "--fork", "--user",
		"--map-root-user",
		tmpPath,
	)

	start := time.Now()
	err = cmd.Run()
	elapsed := time.Since(start)

	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	return &BenchResult{
		ExitCode:  exitCode,
		RuntimeMS: elapsed.Milliseconds(),
		Success:   exitCode == 0,
	}, nil
}

func RunBenchmarkWithTrace(filebytes []byte) (*BenchResult, error) {
	if err := ValidateBinary(filebytes); err != nil {
		return nil, fmt.Errorf("binary validation failed: %v", err)
	}

	tmpPath, cleanup, err := CreateSecureTempFile(filebytes, "benchmark-trace")
	if err != nil {
		return nil, err
	}
	defer cleanup()

	cmd := exec.Command( // give config options to user later
		"systemd-run",
		"--scope",
		"-p", "MemoryMax=32M",
		"-p", "CPUQuota=10%",
		"-p", "TasksMax=10",
		"unshare",
		"--mount", "--uts", "--ipc", "--net", "--pid", "--fork", "--user",
		"--map-root-user",
		"./bintracer.out", tmpPath,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr // optional: show tracer errors

	start := time.Now()
	err = cmd.Run()
	elapsed := time.Since(start)
	lines := strings.Split(stdout.String(), "\n")
	var logs []syscalls.SyscallEntry
	var cgroupName, invocationID string

	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(firstLine, "Running as unit:") {
			// "Running as unit: run-r03369332ebd9417c89fbefe2f3853162.scope; invocation ID: 84495dcd32c040348d67aeeffa22b498"
			parts := strings.Split(firstLine, ";")
			if len(parts) >= 2 {
				unitPart := strings.TrimPrefix(strings.TrimSpace(parts[0]), "Running as unit:")
				cgroupName = strings.TrimSpace(unitPart)
				invIDPart := strings.TrimPrefix(strings.TrimSpace(parts[1]), "invocation ID:")
				invocationID = strings.TrimSpace(invIDPart)
			}
		}
	}

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) != "" {
			regVals := strings.Split(line, " ")
			idx, _ := strconv.Atoi(regVals[0])
			logs = append(logs, syscalls.SyscallEntry{
				Name: syscalls.SyscallNames[uint64(idx)],
				Args: regVals[1:],
			})
		}
	}

	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	return &BenchResult{
		CGroup:       cgroupName,
		InvocationID: invocationID,
		ExitCode:     exitCode,
		RuntimeMS:    elapsed.Milliseconds(),
		Success:      exitCode == 0,
		Syscalls:     logs,
	}, nil
}
