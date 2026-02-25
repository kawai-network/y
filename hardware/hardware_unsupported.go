//go:build !darwin && !linux && !windows

package hardware

// detectPlatformSpecs stub for unsupported platforms
func (specs *HardwareSpecs) detectPlatformSpecs() {
	// Fallback values for unsupported platforms
	specs.TotalRAM = 8
	specs.AvailableRAM = 6
	specs.CPU = "Unknown CPU"
	specs.GPUModel = "Unknown GPU"
	specs.GPUMemory = 0
}
