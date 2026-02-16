// Package paths provides centralized data path configuration.
// All application data storage paths should be resolved through this package.
//
// Platform-specific data directories:
//   - macOS:   ~/Library/Application Support/Kawai/
//   - Windows: %APPDATA%\Kawai\
//   - Linux:   ~/.config/Kawai/ (or $XDG_CONFIG_HOME/Kawai/)
//
// Development mode (running from terminal): ./data/
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	dataDir     string
	initialized bool
	mu          sync.RWMutex
)

// SetDataDir sets custom data directory. Must be called before any path access.
// For development, call SetDataDir("data") early in main().
func SetDataDir(dir string) {
	mu.Lock()
	defer mu.Unlock()
	dataDir = dir
	initialized = false // Reset to re-initialize
}

func ensureInit() {
	mu.Lock()
	defer mu.Unlock()
	if initialized {
		return
	}
	if dataDir == "" {
		if IsPackaged() {
			// Running from packaged app - use platform-specific user data directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				dataDir = "data" // Fallback
			} else {
				switch runtime.GOOS {
				case "darwin":
					dataDir = filepath.Join(homeDir, "Library", "Application Support", "Kawai")
				case "windows":
					// Prefer APPDATA, fallback to LOCALAPPDATA
					if appData := os.Getenv("APPDATA"); appData != "" {
						dataDir = filepath.Join(appData, "Kawai")
					} else if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
						dataDir = filepath.Join(localAppData, "Kawai")
					} else {
						dataDir = filepath.Join(homeDir, "AppData", "Roaming", "Kawai")
					}
				case "linux":
					// Follow XDG Base Directory specification
					if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
						dataDir = filepath.Join(xdgConfig, "Kawai")
					} else {
						dataDir = filepath.Join(homeDir, ".config", "Kawai")
					}
				default:
					// Other Unix-like systems
					dataDir = filepath.Join(homeDir, ".config", "Kawai")
				}
			}
		} else {
			// Running from terminal or development - use relative path
			dataDir = "data"
		}
	}
	_ = os.MkdirAll(dataDir, 0755)
	initialized = true
}

// Base returns the base data directory
func Base() string {
	ensureInit()
	return dataDir
}

// IsPackaged returns true if running from a packaged app bundle
func IsPackaged() bool {
	execPath, err := os.Executable()
	if err != nil {
		return false
	}
	execDir := filepath.Dir(execPath)
	execPathLower := strings.ToLower(execPath)

	switch runtime.GOOS {
	case "darwin":
		// macOS: Check for .app bundle structure
		// Path: /path/to/App.app/Contents/MacOS/binary
		if filepath.Base(filepath.Dir(filepath.Dir(execPath))) == "Contents" {
			return true
		}
		// CLI installs via curl|sh to ~/.local/bin or /usr/local/bin
		homeDir, _ := os.UserHomeDir()
		if homeDir != "" {
			localBin := filepath.Join(homeDir, ".local", "bin") + string(filepath.Separator)
			if strings.HasPrefix(execPath, localBin) {
				return true
			}
		}
		if strings.HasPrefix(execPath, "/usr/local/bin/") ||
			strings.HasPrefix(execPath, "/opt/homebrew/bin/") {
			return true
		}
	case "windows":
		// Windows: Check for resources directory or Program Files
		if filepath.Base(execDir) == "resources" || filepath.Base(filepath.Dir(execDir)) == "resources" {
			return true
		}
		// Also check if installed in Program Files (case-insensitive)
		if strings.Contains(execPathLower, "program files") {
			return true
		}
		// CLI installs to %LOCALAPPDATA%\bin or %USERPROFILE%\.local\bin
		homeDir, _ := os.UserHomeDir()
		if homeDir != "" {
			localBin := filepath.Join(homeDir, ".local", "bin") + string(filepath.Separator)
			if strings.HasPrefix(execPath, localBin) {
				return true
			}
		}
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			localBin := filepath.Join(localAppData, "bin") + string(filepath.Separator)
			if strings.HasPrefix(execPath, localBin) {
				return true
			}
		}
	case "linux":
		// Linux: Check for resources directory or standard install paths
		if filepath.Base(execDir) == "resources" || filepath.Base(filepath.Dir(execDir)) == "resources" {
			return true
		}
		// Check common Linux install paths
		if strings.HasPrefix(execPath, "/usr/") || strings.HasPrefix(execPath, "/opt/") {
			return true
		}
		// CLI installs via curl|sh to ~/.local/bin
		homeDir, _ := os.UserHomeDir()
		if homeDir != "" {
			localBin := filepath.Join(homeDir, ".local", "bin") + string(filepath.Separator)
			if strings.HasPrefix(execPath, localBin) {
				return true
			}
		}
	}

	return false
}

// =============================================================================
// Application-level paths
// =============================================================================

// Database returns path to main SQLite database
func Database() string { return filepath.Join(Base(), "veridium.db") }

// DuckDB returns path to DuckDB database
func DuckDB() string { return filepath.Join(Base(), "duckdb.db") }

// KBAssets returns path to knowledge base assets directory
func KBAssets() string { return filepath.Join(Base(), "kb-assets") }

// FileBase returns path to file uploads directory
func FileBase() string { return filepath.Join(Base(), "files") }

// ContributorLog returns path to contributor server log file
func ContributorLog() string { return filepath.Join(Base(), "logs", "contributor.log") }

// =============================================================================
// Jarvis-specific paths (blockchain/wallet)
// =============================================================================

// Jarvis returns the base directory for Jarvis data
func Jarvis() string { return filepath.Join(Base(), "jarvis") }

// JarvisKeystores returns path to keystore directory
func JarvisKeystores() string { return filepath.Join(Jarvis(), "keystores") }

// JarvisNetworks returns path to custom networks directory
func JarvisNetworks() string { return filepath.Join(Jarvis(), "networks") }

// JarvisAddressBookDB returns path to DuckDB addressbook database
func JarvisAddressBookDB() string { return filepath.Join(Jarvis(), "addressbook.duckdb") }

// JarvisAddressBookHash returns path to addressbook hash file
func JarvisAddressBookHash() string { return filepath.Join(Jarvis(), "addressbook.hash") }

// JarvisCache returns path to cache file
func JarvisCache() string { return filepath.Join(Jarvis(), "cache.json") }

// JarvisAddressBook returns path to addresses.json
func JarvisAddressBook() string { return filepath.Join(Jarvis(), "addresses.json") }

// JarvisSecrets returns path to secrets.json
func JarvisSecrets() string { return filepath.Join(Jarvis(), "secrets.json") }

// =============================================================================
// AI/ML paths (models and libraries) - centralized at user path level
// =============================================================================

// Models returns path to AI/ML models directory (unified for all model types)
// All models are organized by {author}/{repo}/ structure from HuggingFace URLs
func Models() string { return filepath.Join(Base(), "models") }

// ModelPath returns the full path for a model based on its HuggingFace URL
// Extracts author/repo from URL and creates: {Base}/models/{author}/{repo}/
// Example: https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin
//
//	-> {Base}/models/ggerganov/whisper.cpp/
func ModelPath(huggingfaceURL string) (string, error) {
	// Parse URL to extract author/repo
	// Format: https://huggingface.co/{author}/{repo}/resolve/...
	parts := strings.Split(huggingfaceURL, "/")
	if len(parts) < 5 {
		return "", fmt.Errorf("invalid huggingface URL format: %s", huggingfaceURL)
	}

	// Find huggingface.co index
	hfIndex := -1
	for i, part := range parts {
		if strings.Contains(part, "huggingface.co") {
			hfIndex = i
			break
		}
	}

	if hfIndex == -1 || hfIndex+2 >= len(parts) {
		return "", fmt.Errorf("invalid huggingface URL: %s", huggingfaceURL)
	}

	author := parts[hfIndex+1]
	repo := parts[hfIndex+2]

	// Validate author and repo are not empty
	if author == "" || repo == "" {
		return "", fmt.Errorf("empty author or repo in URL: %s", huggingfaceURL)
	}

	// Prevent path traversal attacks - reject paths with ".." or absolute paths
	if strings.Contains(author, "..") || strings.Contains(repo, "..") ||
		filepath.IsAbs(author) || filepath.IsAbs(repo) ||
		strings.ContainsAny(author, "/\\") || strings.ContainsAny(repo, "/\\") {
		return "", fmt.Errorf("invalid path components in URL: %s", huggingfaceURL)
	}

	return filepath.Join(Models(), author, repo), nil
}

// Libraries returns path to shared libraries directory
func Libraries() string { return filepath.Join(Base(), "libraries") }

// Catalogs returns path to model catalogs directory
func Catalogs() string { return filepath.Join(Base(), "catalogs") }

// Templates returns path to chat templates directory
func Templates() string { return filepath.Join(Base(), "templates") }

// =============================================================================
// Stable Diffusion specific paths
// =============================================================================

// StableDiffusionOutputs returns path to SD generated images output directory
func StableDiffusionOutputs() string { return filepath.Join(Base(), "outputs", "stable-diffusion") }

// StableDiffusionBin returns path to SD binary directory
func StableDiffusionBin() string { return filepath.Join(Libraries(), "stable-diffusion", "bin") }

// StableDiffusionChecksums returns path to SD checksums directory
func StableDiffusionChecksums() string {
	return filepath.Join(Libraries(), "stable-diffusion", "checksums")
}

// StableDiffusionMetadata returns path to SD metadata directory
func StableDiffusionMetadata() string {
	return filepath.Join(Libraries(), "stable-diffusion", "metadata")
}

// =============================================================================
// Whisper specific paths
// =============================================================================

// WhisperModels returns path to Whisper models directory
// Used by github.com/kawai-network/whisper for model storage
func WhisperModels() string { return filepath.Join(Models(), "whisper") }

// WhisperLib returns path to Whisper library directory
// Used by github.com/kawai-network/whisper for gowhisper library storage
func WhisperLib() string { return filepath.Join(Libraries(), "whisper") }

// =============================================================================
// Deprecated: Legacy node-specific paths (for backward compatibility)
// These will be removed in future versions. Use the non-Node versions instead.
// =============================================================================

// Node returns the base directory - deprecated, use Base() instead
// Deprecated: Use Base() directly
func Node() string { return Base() }

// NodeModels returns path to AI/ML models directory - deprecated, use Models() instead
// Deprecated: Use Models() instead
func NodeModels() string { return Models() }

// NodeLibraries returns path to shared libraries directory - deprecated, use Libraries() instead
// Deprecated: Use Libraries() instead
func NodeLibraries() string { return Libraries() }

// NodeCatalogs returns path to model catalogs directory - deprecated, use Catalogs() instead
// Deprecated: Use Catalogs() instead
func NodeCatalogs() string { return Catalogs() }

// NodeTemplates returns path to chat templates directory - deprecated, use Templates() instead
// Deprecated: Use Templates() instead
func NodeTemplates() string { return Templates() }
