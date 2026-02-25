//go:build linux

package hardware

import (
	"os/exec"
	"strconv"
	"strings"
)

// detectPlatformSpecs detects hardware specs on Linux
func (specs *HardwareSpecs) detectPlatformSpecs() {
	// Get memory info from /proc/meminfo
	if out, err := exec.Command("cat", "/proc/meminfo").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if memKB, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
						specs.TotalRAM = memKB / (1024 * 1024) // Convert KB to GB
					}
				}
			} else if strings.HasPrefix(line, "MemAvailable:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if memKB, err := strconv.ParseInt(fields[1], 10, 64); err == nil {
						specs.AvailableRAM = memKB / (1024 * 1024) // Convert KB to GB
					}
				}
			}
		}
	}

	// If MemAvailable not found, estimate as 80% of total
	if specs.AvailableRAM == 0 {
		specs.AvailableRAM = int64(float64(specs.TotalRAM) * 0.8)
	}

	// Get CPU info
	if out, err := exec.Command("cat", "/proc/cpuinfo").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					specs.CPU = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	// Try to detect NVIDIA GPU
	if out, err := exec.Command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader,nounits").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 {
			parts := strings.Split(lines[0], ",")
			if len(parts) >= 2 {
				specs.GPUModel = strings.TrimSpace(parts[0])
				if vramMB, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil {
					specs.GPUMemory = vramMB / 1024 // Convert MB to GB
				}
			}
		}
	}
}
