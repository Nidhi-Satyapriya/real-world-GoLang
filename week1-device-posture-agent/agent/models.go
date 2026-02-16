package main

import "time"

// DeviceStatus represents the health status of a device
type DeviceStatus struct {
	Hostname     string    `json:"hostname"`
	IP           string    `json:"ip"`
	DiskUsage    float64   `json:"disk_usage"`
	Status       string    `json:"status"`
	Timestamp    time.Time `json:"timestamp"`
	Message      string    `json:"message,omitempty"`
}

// HealthStatus constants
const (
	StatusHealthy   = "HEALTHY"
	StatusUnhealthy = "UNHEALTHY"
	DiskThreshold   = 90.0 // Threshold percentage for unhealthy status
)
