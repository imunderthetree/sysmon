// internal/stats.go
package internal

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// SystemStats holds all system information
type SystemStats struct {
	CPU       CPUInfo    `json:"cpu"`
	Memory    MemoryInfo `json:"memory"`
	Disk      []DiskInfo `json:"disk"`
	Host      HostInfo   `json:"host"`
	Timestamp time.Time  `json:"timestamp"`
}

type CPUInfo struct {
	Usage     float64 `json:"usage"`
	Cores     int     `json:"cores"`
	ModelName string  `json:"model_name"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	Free        uint64  `json:"free"`
	Buffers     uint64  `json:"buffers"`
	Cached      uint64  `json:"cached"`
}

type DiskInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type HostInfo struct {
	Hostname      string `json:"hostname"`
	OS            string `json:"os"`
	Platform      string `json:"platform"`
	KernelVersion string `json:"kernel_version"`
	Uptime        uint64 `json:"uptime"`
}

// GetSystemStats collects all system statistics
func GetSystemStats() (*SystemStats, error) {
	stats := &SystemStats{
		Timestamp: time.Now(),
	}

	// Get CPU information
	cpuInfo, err := getCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}
	stats.CPU = cpuInfo

	// Get Memory information
	memInfo, err := getMemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}
	stats.Memory = memInfo

	// Get Disk information
	diskInfo, err := getDiskInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}
	stats.Disk = diskInfo

	// Get Host information
	hostInfo, err := getHostInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}
	stats.Host = hostInfo

	return stats, nil
}

func getCPUInfo() (CPUInfo, error) {
	var cpuInfo CPUInfo

	// Get CPU usage percentage (average over 1 second)
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return cpuInfo, err
	}
	if len(percentages) > 0 {
		cpuInfo.Usage = percentages[0]
	}

	// Get CPU count
	cpuInfo.Cores, err = cpu.Counts(true) // logical cores
	if err != nil {
		return cpuInfo, err
	}

	// Get CPU model information
	cpuInfos, err := cpu.Info()
	if err != nil {
		return cpuInfo, err
	}
	if len(cpuInfos) > 0 {
		cpuInfo.ModelName = cpuInfos[0].ModelName
	}

	return cpuInfo, nil
}

func getMemoryInfo() (MemoryInfo, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return MemoryInfo{}, err
	}

	return MemoryInfo{
		Total:       vmem.Total,
		Available:   vmem.Available,
		Used:        vmem.Used,
		UsedPercent: vmem.UsedPercent,
		Free:        vmem.Free,
		Buffers:     vmem.Buffers,
		Cached:      vmem.Cached,
	}, nil
}

func getDiskInfo() ([]DiskInfo, error) {
	partitions, err := disk.Partitions(false) // only physical partitions
	if err != nil {
		return nil, err
	}

	var diskInfos []DiskInfo
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			// Skip partitions we can't access
			continue
		}

		diskInfo := DiskInfo{
			Device:      partition.Device,
			Mountpoint:  partition.Mountpoint,
			Fstype:      partition.Fstype,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		}
		diskInfos = append(diskInfos, diskInfo)
	}

	return diskInfos, nil
}

func getHostInfo() (HostInfo, error) {
	hostStat, err := host.Info()
	if err != nil {
		return HostInfo{}, err
	}

	return HostInfo{
		Hostname:      hostStat.Hostname,
		OS:            hostStat.OS,
		Platform:      hostStat.Platform,
		KernelVersion: hostStat.KernelVersion,
		Uptime:        hostStat.Uptime,
	}, nil
}

// Helper functions for formatting
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func FormatUptime(seconds uint64) string {
	duration := time.Duration(seconds) * time.Second
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
