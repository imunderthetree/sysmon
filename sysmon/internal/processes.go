// internal/processes.go
package internal

import (
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo holds information about a single process
type ProcessInfo struct {
	PID         int32   `json:"pid"`
	Name        string  `json:"name"`
	Username    string  `json:"username"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemPercent  float32 `json:"mem_percent"`
	MemoryMB    uint64  `json:"memory_mb"`
	Status      string  `json:"status"`
	CreateTime  int64   `json:"create_time"`
	NumThreads  int32   `json:"num_threads"`
	CommandLine string  `json:"command_line"`
}

// ProcessStats holds process statistics and summaries
type ProcessStats struct {
	TotalProcesses int           `json:"total_processes"`
	RunningProcs   int           `json:"running_processes"`
	SleepingProcs  int           `json:"sleeping_processes"`
	TopCPU         []ProcessInfo `json:"top_cpu"`
	TopMemory      []ProcessInfo `json:"top_memory"`
	AllProcesses   []ProcessInfo `json:"all_processes"`
	Timestamp      time.Time     `json:"timestamp"`
}

// GetProcessStats collects information about all running processes
func GetProcessStats() (*ProcessStats, error) {
	stats := &ProcessStats{
		Timestamp: time.Now(),
	}

	// Get all process PIDs
	pids, err := process.Pids()
	if err != nil {
		return nil, err
	}

	var processes []ProcessInfo
	var runningCount, sleepingCount int

	// Collect information for each process
	for _, pid := range pids {
		proc, err := process.NewProcess(pid)
		if err != nil {
			continue // Process might have died, skip it
		}

		procInfo, err := getProcessInfo(proc)
		if err != nil {
			continue // Skip processes we can't access
		}

		processes = append(processes, procInfo)

		// Count by status
		switch procInfo.Status {
		case "R", "running":
			runningCount++
		case "S", "sleeping":
			sleepingCount++
		}
	}

	stats.TotalProcesses = len(processes)
	stats.RunningProcs = runningCount
	stats.SleepingProcs = sleepingCount
	stats.AllProcesses = processes

	// Get top processes by CPU
	stats.TopCPU = getTopProcesses(processes, "cpu", 10)

	// Get top processes by Memory
	stats.TopMemory = getTopProcesses(processes, "memory", 10)

	return stats, nil
}

// getProcessInfo extracts information from a process
func getProcessInfo(proc *process.Process) (ProcessInfo, error) {
	var info ProcessInfo

	// Basic info
	info.PID = proc.Pid

	// Process name
	if name, err := proc.Name(); err == nil {
		info.Name = name
	}

	// Username
	if username, err := proc.Username(); err == nil {
		info.Username = username
	} else {
		info.Username = "unknown"
	}

	// CPU percentage (this might take a moment)
	if cpuPercent, err := proc.CPUPercent(); err == nil {
		info.CPUPercent = cpuPercent
	}

	// Memory percentage
	if memPercent, err := proc.MemoryPercent(); err == nil {
		info.MemPercent = memPercent
	}

	// Memory info
	if memInfo, err := proc.MemoryInfo(); err == nil {
		info.MemoryMB = memInfo.RSS / 1024 / 1024 // Convert to MB
	}

	// Status
	if status, err := proc.Status(); err == nil {
		info.Status = strings.Join(status, ",")
	}

	// Create time
	if createTime, err := proc.CreateTime(); err == nil {
		info.CreateTime = createTime
	}

	// Number of threads
	if numThreads, err := proc.NumThreads(); err == nil {
		info.NumThreads = numThreads
	}

	// Command line (this might be long or fail for some processes)
	if cmdline, err := proc.Cmdline(); err == nil && len(cmdline) > 0 {
		info.CommandLine = cmdline
		// Truncate very long command lines
		if len(info.CommandLine) > 100 {
			info.CommandLine = info.CommandLine[:100] + "..."
		}
	} else {
		info.CommandLine = info.Name
	}

	return info, nil
}

// getTopProcesses returns the top N processes sorted by CPU or Memory usage
func getTopProcesses(processes []ProcessInfo, sortBy string, limit int) []ProcessInfo {
	// Make a copy to avoid modifying the original slice
	sorted := make([]ProcessInfo, len(processes))
	copy(sorted, processes)

	// Sort based on the criteria
	switch sortBy {
	case "cpu":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CPUPercent > sorted[j].CPUPercent
		})
	case "memory":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].MemPercent > sorted[j].MemPercent
		})
	}

	// Return top N processes
	if len(sorted) < limit {
		return sorted
	}
	return sorted[:limit]
}
