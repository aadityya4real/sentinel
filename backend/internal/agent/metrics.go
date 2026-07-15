// Package agent collects host-level infrastructure metrics for Sentinel.
package agent

import "time"

// Metrics is a point-in-time snapshot of host resource usage.
type Metrics struct {
	CPUUsagePercent float64     `json:"cpu_usage_percent"`
	Memory          MemoryUsage `json:"memory"`
	Disks           []DiskUsage `json:"disks"`
	Hostname        string      `json:"hostname"`
	OS              string      `json:"os"`
	UptimeSeconds   uint64      `json:"uptime_seconds"`
	Timestamp       time.Time   `json:"timestamp"`
}

// MemoryUsage describes the host's physical memory consumption in bytes.
type MemoryUsage struct {
	TotalBytes     uint64  `json:"total_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	UsedPercent    float64 `json:"used_percent"`
	AvailableBytes uint64  `json:"available_bytes"`
}

// DiskUsage describes the consumed capacity of a mounted filesystem.
type DiskUsage struct {
	Path        string  `json:"path"`
	Filesystem  string  `json:"filesystem"`
	TotalBytes  uint64  `json:"total_bytes"`
	UsedBytes   uint64  `json:"used_bytes"`
	UsedPercent float64 `json:"used_percent"`
}
