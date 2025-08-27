package analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/ashborn3/BinTraceBench/internal/syscalls"
)

func TraceBinary(filebytes []byte) ([]VerboseSyscallEntry, error) {
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

type VerboseSyscallEntry struct {
	PID       int      `json:"pid"`
	Name      string   `json:"name"`
	Number    uint64   `json:"number"`
	Args      []string `json:"args"`
	Return    string   `json:"return,omitempty"`
	Timestamp string   `json:"timestamp"`
	Event     string   `json:"event"` // "entry" or "exit"
}

func ptraceBinaryPath(path string) ([]VerboseSyscallEntry, error) {
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

	var logs []VerboseSyscallEntry
	inSyscall := false
	for {
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

		now := time.Now().Format("2006-01-02 15:04:05.000000")
		if !inSyscall {
			// Syscall entry
			args := []string{
				fmt.Sprintf("RDI=0x%x", regs.Rdi),
				fmt.Sprintf("RSI=0x%x", regs.Rsi),
				fmt.Sprintf("RDX=0x%x", regs.Rdx),
				fmt.Sprintf("R10=0x%x", regs.R10),
				fmt.Sprintf("R8=0x%x", regs.R8),
				fmt.Sprintf("R9=0x%x", regs.R9),
			}
			// Try to decode pointer arguments for open/execve
			name := humanSyscallName(regs.Orig_rax)
			if name == "open" || name == "openat" {
				// First arg is filename pointer
				filename := readStringFromChild(pid, uintptr(regs.Rdi))
				args[0] = fmt.Sprintf("filename=\"%s\" (0x%x)", filename, regs.Rdi)
			} else if name == "execve" {
				filename := readStringFromChild(pid, uintptr(regs.Rdi))
				argv := readStringArrayFromChild(pid, uintptr(regs.Rsi))
				args[0] = fmt.Sprintf("filename=\"%s\" (0x%x)", filename, regs.Rdi)
				args[1] = fmt.Sprintf("argv=%v (0x%x)", argv, regs.Rsi)
			}
			entry := VerboseSyscallEntry{
				PID:       pid,
				Name:      name,
				Number:    regs.Orig_rax,
				Args:      args,
				Timestamp: now,
				Event:     "entry",
			}
			logs = append(logs, entry)
		} else {
			// Syscall exit
			entry := VerboseSyscallEntry{
				PID:       pid,
				Name:      humanSyscallName(regs.Orig_rax),
				Number:    regs.Orig_rax,
				Return:    fmt.Sprintf("0x%x", regs.Rax),
				Timestamp: now,
				Event:     "exit",
			}
			logs = append(logs, entry)
		}
		inSyscall = !inSyscall
	}

	return logs, nil
}

// Read a null-terminated string from the traced process's memory
func readStringFromChild(pid int, addr uintptr) string {
	var data []byte
	for i := 0; i < 256; i++ { // limit max string length
		var tmp [1]byte
		_, err := syscall.PtracePeekData(pid, addr+uintptr(i), tmp[:])
		if err != nil || tmp[0] == 0 {
			break
		}
		data = append(data, tmp[0])
	}
	return string(data)
}

// Read a null-terminated array of string pointers from the traced process's memory
func readStringArrayFromChild(pid int, addr uintptr) []string {
	var result []string
	for j := 0; j < 16; j++ { // limit max argv
		var ptrBuf [8]byte
		_, err := syscall.PtracePeekData(pid, addr+uintptr(j*8), ptrBuf[:])
		if err != nil {
			break
		}
		ptr := uintptr(0)
		for k := 0; k < 8; k++ {
			ptr |= uintptr(ptrBuf[k]) << (8 * k)
		}
		if ptr == 0 {
			break
		}
		s := readStringFromChild(pid, ptr)
		result = append(result, s)
	}
	return result
}

// Helper to get a human-readable syscall name
func humanSyscallName(num uint64) string {
	if int(num) < len(syscalls.SyscallNames) {
		return syscalls.SyscallNames[num]
	}
	return fmt.Sprintf("syscall_%d", num)
}
