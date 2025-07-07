package inspector

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ProcInfo struct {
	PID      int    `json:"pid"`
	Command  string `json:"command"`
	Cmdline  string `json:"cmdline"`
	UID      int    `json:"uid"`
	State    string `json:"state"`
	CPUTime  string `json:"cpu_time"`
	MemoryKB int    `json:"memory_kb"`
}

func GetProcInfo(pid int) (*ProcInfo, error) {
	base := fmt.Sprintf("/proc/%d", pid)

	statusBytes, err := os.ReadFile(base + "/status")
	if err != nil {
		return nil, fmt.Errorf("error reading /proc/%d/status: %s", pid, err.Error())
	}
	status := parseStatus(string(statusBytes))

	cmdLineBytes, err := os.ReadFile(base + "/cmdline")
	if err != nil {
		return nil, fmt.Errorf("error reading /proc/%d/cmdline: %s", pid, err.Error())
	}

	cmdline := strings.ReplaceAll(string(cmdLineBytes), "\u0000", " ")

	statbytes, err := os.ReadFile(base + "/stat")
	if err != nil {
		return nil, fmt.Errorf("error reading /proc/%d/stat: %s", pid, err.Error())
	}

	fields := strings.Fields(string(statbytes))
	utime, _ := strconv.Atoi(fields[13])
	stime, _ := strconv.Atoi(fields[14])
	cputime := fmt.Sprintf("%d ticks", utime+stime)

	return &ProcInfo{
		PID:      pid,
		Command:  status["Name"],
		Cmdline:  cmdline,
		UID:      atoi(status["Uid"]),
		State:    status["State"],
		CPUTime:  cputime,
		MemoryKB: atoi(status["VmRSS"]),
	}, err
}

func parseStatus(input string) map[string]string {
	lines := strings.Split(input, "\n")
	out := make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			out[key] = val
		}
	}
	return out
}

func atoi(s string) int {
	i, _ := strconv.Atoi(strings.Fields(s)[0]) // take only first part
	return i
}
