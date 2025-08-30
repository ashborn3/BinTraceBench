package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/ashborn3/BinTraceBench/internal/syscalls"
)

func RunBenchmarkSecure(filebytes []byte, config *Config) (*BenchResult, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}
	if err := ValidateBinaryWithConfig(filebytes, config); err != nil {
		return nil, fmt.Errorf("binary validation failed: %v", err)
	}

	tmpPath, cleanup, err := CreateSecureTempFileWithConfig(filebytes, "benchmark", config)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), config.MaxExecutionTime)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemd-run",
		"--scope",
		"--quiet", // Reduce output noise
		"-p", "MemoryMax="+config.MaxMemory,
		"-p", "CPUQuota="+config.MaxCPUQuota,
		"-p", "TasksMax="+strconv.Itoa(config.MaxTasks),
		"-p", "PrivateTmp=yes", // Isolated /tmp
		"-p", "NoNewPrivileges=yes", // Prevent privilege escalation
		"unshare",
		"--mount", "--uts", "--ipc", "--net", "--pid", "--fork", "--user",
		"--map-root-user",
		tmpPath,
	)

	start := time.Now()
	err = cmd.Run()
	elapsed := time.Since(start)

	exitCode := 0
	var errorMsg string

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg = "execution timeout"
			exitCode = 124 // Standard timeout exit code
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			errorMsg = err.Error()
			exitCode = -1
		}
	}

	result := &BenchResult{
		ExitCode:  exitCode,
		RuntimeMS: elapsed.Milliseconds(),
		Success:   exitCode == 0,
	}

	if errorMsg != "" {
		result.ErrorMessage = errorMsg
	}

	return result, nil
}

func RunBenchmarkWithTraceSecure(filebytes []byte, config *Config) (*BenchResult, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	if err := ValidateBinaryWithConfig(filebytes, config); err != nil {
		return nil, fmt.Errorf("binary validation failed: %v", err)
	}

	tmpPath, cleanup, err := CreateSecureTempFileWithConfig(filebytes, "benchmark-trace", config)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), config.MaxExecutionTime)
	defer cancel()

	cmd := exec.CommandContext(ctx, "systemd-run",
		"--scope",
		"--quiet",
		"-p", "MemoryMax="+config.MaxMemory,
		"-p", "CPUQuota="+config.MaxCPUQuota,
		"-p", "TasksMax="+strconv.Itoa(config.MaxTasks),
		"-p", "PrivateTmp=yes",
		"-p", "NoNewPrivileges=yes",
		"unshare",
		"--mount", "--uts", "--ipc", "--net", "--pid", "--fork", "--user",
		"--map-root-user",
		"./bintracer", tmpPath,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	start := time.Now()
	err = cmd.Run()
	elapsed := time.Since(start)

	lines := strings.Split(stdout.String(), "\n")
	var logs []syscalls.SyscallEntry
	var cgroupName, invocationID string
	var errorMsg string

	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(firstLine, "Running as unit:") {
			parts := strings.Split(firstLine, ";")
			if len(parts) >= 2 {
				unitPart := strings.TrimPrefix(strings.TrimSpace(parts[0]), "Running as unit:")
				cgroupName = strings.TrimSpace(unitPart)
				invIDPart := strings.TrimPrefix(strings.TrimSpace(parts[1]), "invocation ID:")
				invocationID = strings.TrimSpace(invIDPart)
			}
		}
	}

	// Parse syscall logs
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) != "" {
			regVals := strings.Split(line, " ")
			if len(regVals) > 0 {
				idx, parseErr := strconv.Atoi(regVals[0])
				if parseErr == nil {
					logs = append(logs, syscalls.SyscallEntry{
						Name: syscalls.SyscallNames[uint64(idx)],
						Args: regVals[1:],
					})
				}
			}
		}
	}

	exitCode := 0
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			errorMsg = "execution timeout"
			exitCode = 124
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			errorMsg = err.Error()
			exitCode = -1
		}
	}

	result := &BenchResult{
		CGroup:       cgroupName,
		InvocationID: invocationID,
		ExitCode:     exitCode,
		RuntimeMS:    elapsed.Milliseconds(),
		Success:      exitCode == 0,
		Syscalls:     logs,
	}

	if errorMsg != "" {
		result.ErrorMessage = errorMsg
	}

	return result, nil
}
