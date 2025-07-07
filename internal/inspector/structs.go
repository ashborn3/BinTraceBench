package inspector

type ProcInfo struct {
	PID      int    `json:"pid"`
	Command  string `json:"command"`
	Cmdline  string `json:"cmdline"`
	UID      int    `json:"uid"`
	State    string `json:"state"`
	CPUTime  string `json:"cpu_time"`
	MemoryKB int    `json:"memory_kb"`
}

type OpenFile struct {
	FD     string `json:"fd"`
	Target string `json:"target"`
	Type   string `json:"type"`
}
