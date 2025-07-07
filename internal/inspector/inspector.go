package inspector

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

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

func GetOpenFiles(pid int) ([]OpenFile, error) {
	fdDir := fmt.Sprintf("/proc/%d/fd", pid)
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return nil, fmt.Errorf("error reading file desc directory: %s", err.Error())
	}

	var openfiles []OpenFile
	for _, entry := range entries {
		linkPath := fmt.Sprintf("%s/%s", fdDir, entry.Name())
		target, err := os.Readlink(linkPath)
		if err != nil {
			target = "unreadable"
		}

		var filetype string
		switch {
		case strings.HasPrefix(target, "socket:"):
			filetype = "socket"
		case strings.HasPrefix(target, "pipe:"):
			filetype = "pipe"
		case strings.HasPrefix(target, "/"):
			filetype = "file"
		default:
			filetype = "other"
		}

		openfiles = append(openfiles, OpenFile{
			FD:     entry.Name(),
			Target: target,
			Type:   filetype,
		})
	}

	return openfiles, nil
}

func GetNetworkConnections(pid int) ([]NetConnection, error) {
	var conns []NetConnection
	protos := []string{"tcp", "udp", "tcp6", "udp6"}

	for _, proto := range protos {
		path := fmt.Sprintf("/proc/%d/net/%s", pid, proto)
		data, err := os.ReadFile(path)
		if err != nil {
			continue // process may not have connections in this proto
		}

		lines := strings.Split(string(data), "\n")
		if len(lines) <= 1 {
			continue
		}

		for _, line := range lines[1:] {
			fields := strings.Fields(line)
			if len(fields) < 10 {
				continue
			}

			local := parseIPPort(fields[1])
			remote := parseIPPort(fields[2])
			state := tcpState(fields[3])

			conns = append(conns, NetConnection{
				Protocol: proto,
				Local:    local,
				Remote:   remote,
				State:    state,
			})
		}
	}
	return conns, nil
}
