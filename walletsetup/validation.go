package walletsetup

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// ValidateKeystoreJSON validates keystore JSON format.
func ValidateKeystoreJSON(jsonStr string) (bool, string) {
	if strings.TrimSpace(jsonStr) == "" {
		return false, "Keystore JSON cannot be empty"
	}

	var ks struct {
		Address string `json:"address"`
		Crypto  struct {
			Kdf        string `json:"kdf"`
			Ciphertext string `json:"ciphertext"`
		} `json:"crypto"`
		Version int `json:"version"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &ks); err != nil {
		return false, fmt.Sprintf("Invalid JSON format: %v", err)
	}

	if ks.Address == "" {
		return false, "Missing required field: address"
	}
	if ks.Version == 0 {
		return false, "Missing or invalid required field: version"
	}
	if ks.Crypto.Kdf == "" && ks.Crypto.Ciphertext == "" {
		return false, "Invalid keystore: missing crypto information (kdf or ciphertext)"
	}

	return true, ""
}

// ValidatePrivateKey validates a hex private key and returns a reason on failure.
func ValidatePrivateKey(privateKey string) (bool, string) {
	privateKey = strings.TrimPrefix(privateKey, "0x")
	privateKey = strings.TrimPrefix(privateKey, "0X")

	if len(privateKey) != 64 {
		return false, fmt.Sprintf("invalid length: %d/64 characters", len(privateKey))
	}

	if _, err := hex.DecodeString(privateKey); err != nil {
		return false, "invalid hex characters"
	}

	return true, ""
}

// ValidateMnemonicBasic validates a 12 or 24-word mnemonic with lowercase words.
func ValidateMnemonicBasic(mnemonic string) (bool, string) {
	words := strings.Fields(mnemonic)
	if len(words) != 12 && len(words) != 24 {
		return false, fmt.Sprintf("invalid mnemonic: must be 12 or 24 words, got %d", len(words))
	}

	for _, word := range words {
		for _, r := range word {
			if r < 'a' || r > 'z' {
				return false, "mnemonic contains invalid words"
			}
		}
	}

	return true, ""
}

// PasswordStrength returns a score (0-100) and UI hints.
func PasswordStrength(password string) (int, string, string) {
	score := 0
	if len(password) >= 8 {
		score += 25
	}
	if len(password) >= 12 {
		score += 15
	}
	if len(password) >= 16 {
		score += 10
	}
	if len(password) > 0 {
		if strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			score += 15
		}
		if strings.ContainsAny(password, "0123456789") {
			score += 15
		}
		if strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
			score += 20
		}
	}

	score = min(score, 100)

	var label, color string
	switch {
	case score < 40:
		label = "Weak - Add more characters"
		color = "#ff4d4f"
	case score < 60:
		label = "Fair - Add uppercase or numbers"
		color = "#faad14"
	case score < 80:
		label = "Good - Add special characters"
		color = "#52c41a"
	default:
		label = "Strong password"
		color = "#1890ff"
	}

	return score, label, color
}
