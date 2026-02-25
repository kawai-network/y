//go:build darwin

package hardware

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

// detectPlatformSpecs detects hardware specs on macOS
func (specs *HardwareSpecs) detectPlatformSpecs() {
	// Get total memory
	if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
		if memBytes, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64); err == nil {
			specs.TotalRAM = memBytes / (1024 * 1024 * 1024) // Convert to GB
		}
	}

	// Estimate available memory (roughly 80% of total, accounting for OS usage)
	specs.AvailableRAM = int64(float64(specs.TotalRAM) * 0.8)

	// Get CPU model
	if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
		specs.CPU = strings.TrimSpace(string(out))
	}

	// Try to detect GPU (basic detection)
	if out, err := exec.Command("system_profiler", "SPDisplaysDataType", "-json").Output(); err == nil {
		specs.parseGPUFromSystemProfiler(string(out))
	}
}

// parseGPUFromSystemProfiler parses GPU info from macOS system_profiler JSON output
func (specs *HardwareSpecs) parseGPUFromSystemProfiler(jsonStr string) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return
	}

	if displays, ok := data["SPDisplaysDataType"].([]interface{}); ok && len(displays) > 0 {
		if display, ok := displays[0].(map[string]interface{}); ok {
			if name, ok := display["sppci_model"].(string); ok {
				specs.GPUModel = name
			}
			// Note: macOS system_profiler doesn't easily provide VRAM info for all GPUs
			// For Apple Silicon, we can make educated guesses based on model
			if strings.Contains(specs.GPUModel, "Apple") {
				// Apple Silicon GPUs share system memory
				specs.GPUMemory = specs.TotalRAM / 4 // Conservative estimate
			}
		}
	}
}
