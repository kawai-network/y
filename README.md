# y

Kawai shared utilities and common packages for the Kawai ecosystem.

## Packages

### `paths`

Centralized data path configuration for all Kawai applications.

```go
import "github.com/kawai-network/y/paths"
```

#### Features

- **Cross-platform data directories**: Automatically detects platform (macOS, Windows, Linux)
  - macOS: `~/Library/Application Support/Kawai/`
  - Windows: `%APPDATA%\Kawai\`
  - Linux: `~/.config/Kawai/`
- **Development mode**: Uses `./data/` when running from terminal
- **Packaged app detection**: Detects when running from installed/packaged apps

#### Usage

```go
// Get base data directory
baseDir := paths.Base()

// Get specific paths
modelsDir := paths.Models()
libsDir := paths.Libraries()
dbPath := paths.Database()

// Set custom data directory (call before any path access)
paths.SetDataDir("/custom/path")
```

#### Available Paths

- `Base()` - Base data directory
- `Database()` - SQLite database path
- `DuckDB()` - DuckDB database path
- `Models()` - AI/ML models directory
- `Libraries()` - Shared libraries directory
- `Catalogs()` - Model catalogs directory
- `Templates()` - Chat templates directory
- `KBAssets()` - Knowledge base assets
- `FileBase()` - File uploads directory
- `ContributorLog()` - Contributor server log

#### AI/ML Specific Paths

- `Models()` - Unified models directory
- `ModelPath(huggingfaceURL)` - Parse HuggingFace URL to local path
- `Libraries()` - Shared libraries directory
- `StableDiffusionOutputs()` - SD output directory
- `StableDiffusionBin()` - SD binaries
- `StableDiffusionChecksums()` - SD checksums
- `StableDiffusionMetadata()` - SD metadata
- `WhisperModels()` - Whisper models
- `WhisperLib()` - Whisper library

#### Jarvis (Wallet) Paths

- `Jarvis()` - Base Jarvis directory
- `JarvisKeystores()` - Keystore directory
- `JarvisNetworks()` - Custom networks
- `JarvisAddressBook()` - Address book
- `JarvisSecrets()` - Secrets file
- `JarvisCache()` - Cache file

## Installation

```bash
go get github.com/kawai-network/y/paths
```

## License

MIT
