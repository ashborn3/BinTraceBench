package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type BenchResult struct {
	ExitCode  int   `json:"exit_code"`
	RuntimeMS int64 `json:"runtime_ms"`
	Success   bool  `json:"success"`
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
