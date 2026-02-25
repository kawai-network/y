//go:build windows

package hardware

import (
	"os/exec"
	"strconv"
	"strings"
)

// detectPlatformSpecs detects hardware specs on Windows
func (specs *HardwareSpecs) detectPlatformSpecs() {
	// Get total memory using wmic
	if out, err := exec.Command("wmic", "computersystem", "get", "TotalPhysicalMemory", "/value").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "TotalPhysicalMemory=") {
				memStr := strings.TrimPrefix(line, "TotalPhysicalMemory=")
				memStr = strings.TrimSpace(memStr)
				if memBytes, err := strconv.ParseInt(memStr, 10, 64); err == nil {
					specs.TotalRAM = memBytes / (1024 * 1024 * 1024) // Convert to GB
				}
			}
		}
	}

	// Estimate available memory
	specs.AvailableRAM = int64(float64(specs.TotalRAM) * 0.8)

	// Get CPU info
	if out, err := exec.Command("wmic", "cpu", "get", "name", "/value").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Name=") {
				specs.CPU = strings.TrimPrefix(line, "Name=")
				specs.CPU = strings.TrimSpace(specs.CPU)
				break
			}
		}
	}
}
