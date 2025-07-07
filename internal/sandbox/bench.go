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
	ExitCode  int      `json:"exit_code"`
	RuntimeMS int64    `json:"runtime_ms"`
	Success   bool     `json:"success"`
	Syscalls  []string `json:"syscalls,omitempty"`
}

func RunBenchmark(filebytes []byte) (*BenchResult, error) {
	tmpfile, err := os.CreateTemp("", "bintracebench-*")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(filebytes)
	if err != nil {
		return nil, fmt.Errorf("error writing bytes to temp file: %s", err.Error())
	}

	tmpfile.Chmod(0755)
	tmpfile.Close()

	cmd := exec.Command("unshare",
		"--mount", "--uts", "--ipc", "--net", "--pid", "--fork", "--user",
		"--map-root-user",
		tmpfile.Name(),
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
	tmpfile, err := os.CreateTemp("", "bintracebench-*")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(filebytes)
	if err != nil {
		return nil, fmt.Errorf("error writing bytes to temp file: %s", err.Error())
	}

	tmpfile.Chmod(0755)
	tmpfile.Close()

	cmd := exec.Command("unshare",
		"--mount", "--uts", "--ipc", "--net", "--pid", "--fork", "--user",
		"--map-root-user",
		"./bintracer", tmpfile.Name(),
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr // optional: show tracer errors

	start := time.Now()
	err = cmd.Run()
	elapsed := time.Since(start)

	lines := strings.Split(stdout.String(), "\n")
	var logs []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			idx, _ := strconv.Atoi(line)
			logs = append(logs, syscalls.SyscallNames[uint64(idx)])
		}
	}

	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}

	return &BenchResult{
		ExitCode:  exitCode,
		RuntimeMS: elapsed.Milliseconds(),
		Success:   exitCode == 0,
		Syscalls:  logs,
	}, nil
}
