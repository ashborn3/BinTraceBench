// internal/tools/bintracer.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/ashborn3/BinTraceBench/internal/syscalls"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: bintracer <binary> [args...]")
	}
	binary := os.Args[1]
	args := os.Args[2:]

	cmd := exec.Command(binary, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Ptrace: true,
	}

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start: %v", err)
	}

	pid := cmd.Process.Pid
	var status syscall.WaitStatus

	// Wait for initial PTRACE_TRACEME stop
	_, err = syscall.Wait4(pid, &status, 0, nil)
	if err != nil {
		log.Fatalf("Initial wait failed: %v", err)
	}

	err = syscall.PtraceSetOptions(pid, syscall.PTRACE_O_TRACESYSGOOD)
	if err != nil {
		log.Fatalf("PtraceSetOptions failed: %v", err)
	}

	for {
		err = syscall.PtraceSyscall(pid, 0)
		if err != nil {
			break
		}

		_, err = syscall.Wait4(pid, &status, 0, nil)
		if err != nil || status.Exited() || status.Signaled() {
			break
		}

		// Get syscall number
		var regs syscall.PtraceRegs
		syscall.PtraceGetRegs(pid, &regs)

		name := syscalls.SyscallNames[regs.Orig_rax]
		if name == "" {
			name = fmt.Sprintf("syscall_%d", regs.Orig_rax)
		}
		fmt.Println(name)

		syscall.PtraceSyscall(pid, 0)
		syscall.Wait4(pid, &status, 0, nil)
	}

	cmd.Wait()
}
