package sandbox

import (
	"fmt"
	"time"
)

type Config struct {
	// Execution limits
	MaxExecutionTime time.Duration // Maximum time binary can run
	MaxMemory        string        // Memory limit (e.g., "32M")
	MaxCPUQuota      string        // CPU quota (e.g., "10%")
	MaxTasks         int           // Maximum number of tasks/processes

	// Filesystem limits
	MaxFileSize   int64  // Maximum file size in bytes
	TempDirPrefix string // Prefix for temporary directories
}

func DefaultConfig() *Config {
	return &Config{
		MaxExecutionTime: 30 * time.Second, // 30 second timeout
		MaxMemory:        "32M",            // 32MB memory limit
		MaxCPUQuota:      "10%",            // 10% CPU quota
		MaxTasks:         10,               // Max 10 processes
		MaxFileSize:      50 * 1024 * 1024, // 50MB file size limit
		TempDirPrefix:    "bintracebench-sandbox",
	}
}

func (c *Config) Validate() error {
	if c.MaxExecutionTime <= 0 {
		return fmt.Errorf("MaxExecutionTime must be positive")
	}
	if c.MaxFileSize <= 0 {
		return fmt.Errorf("MaxFileSize must be positive")
	}
	if c.MaxTasks <= 0 {
		return fmt.Errorf("MaxTasks must be positive")
	}
	return nil
}
