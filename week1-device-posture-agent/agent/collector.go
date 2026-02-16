package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// SystemCollector handles collection of system information
type SystemCollector struct{}

// NewSystemCollector creates a new SystemCollector instance
func NewSystemCollector() *SystemCollector {
	return &SystemCollector{}
}

// GetHostname retrieves the system hostname
func (sc *SystemCollector) GetHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %w", err)
	}
	return hostname, nil
}

// GetLocalIP retrieves the local IP address (non-loopback)
func (sc *SystemCollector) GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, addr := range addrs {
		// Check if it's an IP address (not a network interface)
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			// Get IPv4 address
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid local IP address found")
}

// GetDiskUsage retrieves disk usage percentage based on OS
func (sc *SystemCollector) GetDiskUsage() (float64, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		return sc.getDiskUsageUnix()
	case "windows":
		return sc.getDiskUsageWindows()
	default:
		return 0, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// getDiskUsageUnix gets disk usage for Unix-like systems (macOS, Linux)
func (sc *SystemCollector) getDiskUsageUnix() (float64, error) {
	cmd := exec.Command("df", "-h", "/")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute df command: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected df output format")
	}

	// Parse the second line (actual disk info)
	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return 0, fmt.Errorf("unexpected df output fields")
	}

	// The 5th field is the usage percentage (e.g., "45%")
	usageStr := strings.TrimSuffix(fields[4], "%")
	usage, err := strconv.ParseFloat(usageStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse disk usage: %w", err)
	}

	return usage, nil
}

// getDiskUsageWindows gets disk usage for Windows systems
func (sc *SystemCollector) getDiskUsageWindows() (float64, error) {
	cmd := exec.Command("wmic", "logicaldisk", "get", "size,freespace,caption")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to execute wmic command: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected wmic output format")
	}

	// Parse C: drive info (usually the third line)
	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) >= 3 && strings.HasPrefix(fields[0], "C:") {
			free, err1 := strconv.ParseFloat(fields[1], 64)
			total, err2 := strconv.ParseFloat(fields[2], 64)
			
			if err1 != nil || err2 != nil {
				continue
			}

			used := total - free
			usage := (used / total) * 100
			return usage, nil
		}
	}

	return 0, fmt.Errorf("failed to parse Windows disk usage")
}

// CollectDeviceStatus collects all system information and determines health status
func (sc *SystemCollector) CollectDeviceStatus() (*DeviceStatus, error) {
	hostname, err := sc.GetHostname()
	if err != nil {
		return nil, err
	}

	ip, err := sc.GetLocalIP()
	if err != nil {
		return nil, err
	}

	diskUsage, err := sc.GetDiskUsage()
	if err != nil {
		return nil, err
	}

	// Determine health status based on disk usage
	status := StatusHealthy
	message := "All systems operational"
	
	if diskUsage > DiskThreshold {
		status = StatusUnhealthy
		message = fmt.Sprintf("Critical: Disk usage at %.2f%% (threshold: %.0f%%)", diskUsage, DiskThreshold)
	}

	return &DeviceStatus{
		Hostname:  hostname,
		IP:        ip,
		DiskUsage: diskUsage,
		Status:    status,
		Timestamp: time.Now(),
		Message:   message,
	}, nil
}
