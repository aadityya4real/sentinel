package agent

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// SystemCollector gathers metrics from the operating system.
type SystemCollector struct {
	cpuSampleInterval time.Duration
}

// NewSystemCollector creates a system collector that samples CPU usage for the supplied interval.
func NewSystemCollector(cpuSampleInterval time.Duration) (*SystemCollector, error) {
	if cpuSampleInterval <= 0 {
		return nil, fmt.Errorf("cpu sample interval must be greater than zero")
	}

	return &SystemCollector{cpuSampleInterval: cpuSampleInterval}, nil
}

// Collect gathers CPU, memory, disk, host, uptime, and timestamp metrics.
func (c *SystemCollector) Collect(ctx context.Context) (Metrics, error) {
	cpuUsage, err := cpu.PercentWithContext(ctx, c.cpuSampleInterval, false)
	if err != nil {
		return Metrics{}, fmt.Errorf("collect cpu usage: %w", err)
	}
	if len(cpuUsage) != 1 {
		return Metrics{}, fmt.Errorf("collect cpu usage: expected one aggregate value, got %d", len(cpuUsage))
	}

	memory, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return Metrics{}, fmt.Errorf("collect memory usage: %w", err)
	}

	hostInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return Metrics{}, fmt.Errorf("collect host information: %w", err)
	}
	disks, err := collectDiskUsage(ctx)
	if err != nil {
		return Metrics{}, err
	}

	return Metrics{
		CPUUsagePercent: cpuUsage[0],
		Memory: MemoryUsage{
			TotalBytes:     memory.Total,
			UsedBytes:      memory.Used,
			UsedPercent:    memory.UsedPercent,
			AvailableBytes: memory.Available,
		},
		Disks:         disks,
		Hostname:      hostInfo.Hostname,
		OS:            hostInfo.Platform,
		UptimeSeconds: hostInfo.Uptime,
		Timestamp:     time.Now().UTC(),
	}, nil
}

func collectDiskUsage(ctx context.Context) ([]DiskUsage, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("list disk partitions: %w", err)
	}

	usage := make([]DiskUsage, 0, len(partitions))
	seenPaths := make(map[string]struct{}, len(partitions))
	for _, partition := range partitions {
		// Ignore partitions the operating system reports but does not expose as
		// readable filesystems; the API only needs the usable mounted disks.
		if partition.Mountpoint == "" {
			continue
		}
		if _, exists := seenPaths[partition.Mountpoint]; exists {
			continue
		}

		stats, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			continue
		}
		if stats.Total == 0 || stats.Used > stats.Total || math.IsNaN(stats.UsedPercent) || math.IsInf(stats.UsedPercent, 0) || stats.UsedPercent < 0 || stats.UsedPercent > 100 {
			continue
		}

		seenPaths[partition.Mountpoint] = struct{}{}
		usage = append(usage, DiskUsage{
			Path:        partition.Mountpoint,
			Filesystem:  partition.Fstype,
			TotalBytes:  stats.Total,
			UsedBytes:   stats.Used,
			UsedPercent: stats.UsedPercent,
		})
	}
	if len(usage) == 0 {
		return nil, fmt.Errorf("collect disk usage: no readable mounted filesystems")
	}

	return usage, nil
}
