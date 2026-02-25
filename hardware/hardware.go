package hardware

import (
	"log"
	"runtime"
)

// HardwareSpecs represents system hardware specifications
type HardwareSpecs struct {
	TotalRAM     int64  // Total RAM in GB
	AvailableRAM int64  // Available RAM in GB
	CPU          string // CPU model
	CPUCores     int    // Number of CPU cores
	GPUMemory    int64  // GPU VRAM in GB (if available)
	GPUModel     string // GPU model
}

// DetectHardwareSpecs detects the current system's hardware specifications
func DetectHardwareSpecs() *HardwareSpecs {
	specs := &HardwareSpecs{}
	specs.CPUCores = runtime.NumCPU()

	// Platform-specific detection is implemented in platform files
	specs.detectPlatformSpecs()

	log.Printf("Detected hardware: RAM=%dGB (available=%dGB), CPU cores=%d, GPU=%s (VRAM=%dGB)",
		specs.TotalRAM, specs.AvailableRAM, specs.CPUCores, specs.GPUModel, specs.GPUMemory)

	return specs
}
