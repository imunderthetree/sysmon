// internal/network.go
package internal

import (
	"fmt"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/net"
)

// NetworkInterface holds information about a network interface
type NetworkInterface struct {
	Name        string    `json:"name"`
	BytesSent   uint64    `json:"bytes_sent"`
	BytesRecv   uint64    `json:"bytes_recv"`
	PacketsSent uint64    `json:"packets_sent"`
	PacketsRecv uint64    `json:"packets_recv"`
	Errin       uint64    `json:"errin"`
	Errout      uint64    `json:"errout"`
	Dropin      uint64    `json:"dropin"`
	Dropout     uint64    `json:"dropout"`
	Speed       uint64    `json:"speed"` // Interface speed in Mbps
	IsUp        bool      `json:"is_up"`
	HasTraffic  bool      `json:"has_traffic"`
	LastUpdate  time.Time `json:"last_update"`
}

// NetworkStats holds all network statistics
type NetworkStats struct {
	Interfaces   []NetworkInterface `json:"interfaces"`
	TotalSent    uint64             `json:"total_sent"`
	TotalRecv    uint64             `json:"total_recv"`
	ActiveIfaces int                `json:"active_interfaces"`
	Connections  int                `json:"connections"`
	Timestamp    time.Time          `json:"timestamp"`
}

// NetworkSpeed holds speed calculations
type NetworkSpeed struct {
	Interface    string    `json:"interface"`
	UploadKBps   float64   `json:"upload_kbps"`
	DownloadKBps float64   `json:"download_kbps"`
	Timestamp    time.Time `json:"timestamp"`
}

// Global variables to track previous readings for speed calculation
var (
	previousNetStats map[string]NetworkInterface
	lastNetworkRead  time.Time
)

// GetNetworkStats collects network interface statistics
func GetNetworkStats() (*NetworkStats, error) {
	stats := &NetworkStats{
		Timestamp: time.Now(),
	}

	// Get network IO counters per interface
	ioCounters, err := net.IOCounters(true) // true = per interface
	if err != nil {
		return nil, fmt.Errorf("failed to get network IO counters: %w", err)
	}

	var interfaces []NetworkInterface
	var totalSent, totalRecv uint64
	var activeCount int

	// Process each interface
	for _, counter := range ioCounters {
		iface := NetworkInterface{
			Name:        counter.Name,
			BytesSent:   counter.BytesSent,
			BytesRecv:   counter.BytesRecv,
			PacketsSent: counter.PacketsSent,
			PacketsRecv: counter.PacketsRecv,
			Errin:       counter.Errin,
			Errout:      counter.Errout,
			Dropin:      counter.Dropin,
			Dropout:     counter.Dropout,
			LastUpdate:  time.Now(),
		}

		// Check if interface has any traffic (indicates it's active)
		iface.HasTraffic = (counter.BytesSent > 0 || counter.BytesRecv > 0)
		iface.IsUp = iface.HasTraffic // Simple heuristic for "up" status

		// Skip loopback and inactive interfaces for totals
		if !isLoopbackInterface(counter.Name) && iface.HasTraffic {
			totalSent += counter.BytesSent
			totalRecv += counter.BytesRecv
			activeCount++
		}

		interfaces = append(interfaces, iface)
	}

	// Sort interfaces by total traffic (most active first)
	sort.Slice(interfaces, func(i, j int) bool {
		totalI := interfaces[i].BytesSent + interfaces[i].BytesRecv
		totalJ := interfaces[j].BytesSent + interfaces[j].BytesRecv
		return totalI > totalJ
	})

	stats.Interfaces = interfaces
	stats.TotalSent = totalSent
	stats.TotalRecv = totalRecv
	stats.ActiveIfaces = activeCount

	// Get connection count
	connections, err := getConnectionCount()
	if err == nil {
		stats.Connections = connections
	}

	return stats, nil
}

// GetNetworkSpeeds calculates current network speeds
func GetNetworkSpeeds() ([]NetworkSpeed, error) {
	currentStats, err := GetNetworkStats()
	if err != nil {
		return nil, err
	}

	var speeds []NetworkSpeed
	now := time.Now()

	// Initialize previous stats if first run
	if previousNetStats == nil {
		previousNetStats = make(map[string]NetworkInterface)
		lastNetworkRead = now

		// Store current stats for next calculation
		for _, iface := range currentStats.Interfaces {
			previousNetStats[iface.Name] = iface
		}

		return speeds, nil // Return empty speeds for first run
	}

	// Calculate time difference
	timeDiff := now.Sub(lastNetworkRead).Seconds()
	if timeDiff <= 0 {
		return speeds, nil
	}

	// Calculate speeds for each interface
	for _, current := range currentStats.Interfaces {
		if previous, exists := previousNetStats[current.Name]; exists {
			// Calculate bytes per second
			sentDiff := float64(current.BytesSent - previous.BytesSent)
			recvDiff := float64(current.BytesRecv - previous.BytesRecv)

			speed := NetworkSpeed{
				Interface:    current.Name,
				UploadKBps:   (sentDiff / timeDiff) / 1024, // Convert to KB/s
				DownloadKBps: (recvDiff / timeDiff) / 1024, // Convert to KB/s
				Timestamp:    now,
			}

			// Only include interfaces with significant traffic
			if speed.UploadKBps > 0.1 || speed.DownloadKBps > 0.1 {
				speeds = append(speeds, speed)
			}
		}
	}

	// Update previous stats for next calculation
	for _, iface := range currentStats.Interfaces {
		previousNetStats[iface.Name] = iface
	}
	lastNetworkRead = now

	// Sort by total speed (highest first)
	sort.Slice(speeds, func(i, j int) bool {
		totalI := speeds[i].UploadKBps + speeds[i].DownloadKBps
		totalJ := speeds[j].UploadKBps + speeds[j].DownloadKBps
		return totalI > totalJ
	})

	return speeds, nil
}

// getConnectionCount returns the number of active network connections
func getConnectionCount() (int, error) {
	connections, err := net.Connections("all")
	if err != nil {
		return 0, err
	}

	// Count only established connections
	established := 0
	for _, conn := range connections {
		if conn.Status == "ESTABLISHED" {
			established++
		}
	}

	return established, nil
}

// isLoopbackInterface checks if an interface is a loopback interface
func isLoopbackInterface(name string) bool {
	loopbackNames := []string{"lo", "lo0", "Loopback"}
	for _, loName := range loopbackNames {
		if name == loName {
			return true
		}
	}
	return false
}

// GetTopNetworkInterfaces returns the most active network interfaces
func GetTopNetworkInterfaces(interfaces []NetworkInterface, limit int) []NetworkInterface {
	// Filter out loopback and inactive interfaces
	var active []NetworkInterface
	for _, iface := range interfaces {
		if !isLoopbackInterface(iface.Name) && iface.HasTraffic {
			active = append(active, iface)
		}
	}

	// Sort by total traffic
	sort.Slice(active, func(i, j int) bool {
		totalI := active[i].BytesSent + active[i].BytesRecv
		totalJ := active[j].BytesSent + active[j].BytesRecv
		return totalI > totalJ
	})

	if len(active) < limit {
		return active
	}
	return active[:limit]
}

// FormatNetworkSpeed formats network speed for display
func FormatNetworkSpeed(kbps float64) string {
	if kbps >= 1024*1024 {
		return fmt.Sprintf("%.1f GB/s", kbps/(1024*1024))
	} else if kbps >= 1024 {
		return fmt.Sprintf("%.1f MB/s", kbps/1024)
	} else if kbps >= 1 {
		return fmt.Sprintf("%.1f KB/s", kbps)
	} else {
		return fmt.Sprintf("%.0f B/s", kbps*1024)
	}
}

// FormatNetworkBytes formats network byte counts for display
func FormatNetworkBytes(bytes uint64) string {
	return FormatBytes(bytes) // Reuse the existing FormatBytes function
}
