package analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/ashborn3/BinTraceBench/internal/syscalls"
)

func TraceBinary(filebytes []byte) ([]syscalls.SyscallEntry, error) {
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

	logs, err := ptraceBinaryPath(tmpfile.Name())
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func ptraceBinaryPath(path string) ([]syscalls.SyscallEntry, error) {
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

	var logs []syscalls.SyscallEntry
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
		entry := syscalls.SyscallEntry{
			Name: syscalls.SyscallNames[regs.Orig_rax],
			Args: []string{
				fmt.Sprintf("arg0=0x%x", regs.Rdi),
				fmt.Sprintf("arg1=0x%x", regs.Rsi),
				fmt.Sprintf("arg2=0x%x", regs.Rdx),
				fmt.Sprintf("arg3=0x%x", regs.R10),
				fmt.Sprintf("arg4=0x%x", regs.R8),
				fmt.Sprintf("arg5=0x%x", regs.R9),
			},
		}
		logs = append(logs, entry)

	}

	return logs, nil
}
