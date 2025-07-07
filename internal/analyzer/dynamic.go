package analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func TraceBinary(filebytes []byte) ([]string, error) {
	tmpfile, err := os.CreateTemp("", "bintracebench-*")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %s", err.Error())
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	_, err = tmpfile.Write(filebytes)
	if err != nil {
		return nil, fmt.Errorf("error writing bytes to temp file: %s", err.Error())
	}

	tmpfile.Chmod(0755)

	logs, err := ptraceBinaryPath(tmpfile.Name())
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func ptraceBinaryPath(path string) ([]string, error) {
	cmd := exec.Command(path)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Ptrace: true,
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting ptraced binary: %s", err.Error())
	}
	pid := cmd.Process.Pid

	if _, err := syscall.Wait4(pid, nil, 0, nil); err != nil {
		return nil, fmt.Errorf("initial wait failed: %w", err)
	}

	var logs []string
	for {
		if err := syscall.PtraceSyscall(pid, 0); err != nil {
			break
		}
		if _, err := syscall.Wait4(pid, nil, 0, nil); err != nil {
			break
		}
		if err := syscall.PtraceSyscall(pid, 0); err != nil {
			break
		}
		if _, err := syscall.Wait4(pid, nil, 0, nil); err != nil {
			break
		}

		var regs syscall.PtraceRegs
		if err := syscall.PtraceGetRegs(pid, &regs); err != nil {
			break
		}

		// Note: On x86_64 Linux, syscall number is in Orig_rax
		syscallNum := regs.Orig_rax
		logs = append(logs, fmt.Sprintf("syscall %d", syscallNum))
	}

	return logs, nil
}
