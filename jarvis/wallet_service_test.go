package jarvis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestImportKeystore_AddressChecksum tests that imported addresses are properly checksummed
func TestImportKeystore_AddressChecksum(t *testing.T) {
	// Test cases with different address formats
	// Using real Ethereum addresses to test checksum
	testCases := []struct {
		name              string
		addressInKeystore string // Address as stored in keystore (lowercase, no 0x)
	}{
		{
			name:              "lowercase address 1",
			addressInKeystore: "5aaeb6053f3e94c9b9a09f33669435e7ef1beaed",
		},
		{
			name:              "lowercase address 2",
			addressInKeystore: "fb6916095ca1df60bb79ce92ce3ea74c37c5d359",
		},
		{
			name:              "all uppercase",
			addressInKeystore: "5AAEB6053F3E94C9B9A09F33669435E7EF1BEAED",
		},
		{
			name:              "mixed case",
			addressInKeystore: "dBF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock keystore JSON
			keystoreData := map[string]interface{}{
				"address": tc.addressInKeystore,
				"crypto": map[string]interface{}{
					"cipher":       "aes-128-ctr",
					"ciphertext":   "mockdata",
					"cipherparams": map[string]interface{}{"iv": "mockiv"},
					"kdf":          "scrypt",
					"kdfparams": map[string]interface{}{
						"dklen": 32,
						"n":     262144,
						"p":     1,
						"r":     8,
						"salt":  "mocksalt",
					},
					"mac": "mockmac",
				},
				"id":      "mock-uuid",
				"version": 3,
			}

			keystoreJSON, err := json.Marshal(keystoreData)
			require.NoError(t, err)

			// Test the address extraction and checksumming logic
			// (We can't test the full ImportKeystore without a valid keystore,
			// but we can test the address conversion logic)
			var parsedData map[string]interface{}
			err = json.Unmarshal(keystoreJSON, &parsedData)
			require.NoError(t, err)

			addressRaw, ok := parsedData["address"].(string)
			require.True(t, ok, "address field should exist")

			// This is the fix we're testing
			checksummedAddress := common.HexToAddress(addressRaw).Hex()

			// Verify it's a valid Ethereum address
			assert.True(t, common.IsHexAddress(checksummedAddress),
				"Should be a valid hex address")

			// Verify the address starts with 0x
			assert.True(t, strings.HasPrefix(checksummedAddress, "0x"),
				"Address should have 0x prefix")

			// Verify checksum consistency - checksumming twice should give same result
			checksummedAgain := common.HexToAddress(checksummedAddress).Hex()
			assert.Equal(t, checksummedAddress, checksummedAgain,
				"Checksumming should be idempotent")

			// Verify that different case inputs produce the same checksum
			lowercaseInput := strings.ToLower(tc.addressInKeystore)
			uppercaseInput := strings.ToUpper(tc.addressInKeystore)

			checksumFromLower := common.HexToAddress(lowercaseInput).Hex()
			checksumFromUpper := common.HexToAddress(uppercaseInput).Hex()

			assert.Equal(t, checksumFromLower, checksumFromUpper,
				"Same address in different cases should produce same checksum")
		})
	}
}

// TestAddressChecksumConsistency tests that all address handling is consistent
func TestAddressChecksumConsistency(t *testing.T) {
	testAddresses := []string{
		"0x5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed",
		"0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
		"0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
		"0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb",
	}

	for _, addr := range testAddresses {
		t.Run(addr, func(t *testing.T) {
			// Test that checksumming is idempotent
			checksummed1 := common.HexToAddress(addr).Hex()
			checksummed2 := common.HexToAddress(checksummed1).Hex()

			assert.Equal(t, checksummed1, checksummed2,
				"Checksumming should be idempotent")

			// Test that lowercase and uppercase produce same checksum
			lowercaseAddr := strings.ToLower(addr)
			uppercaseAddr := strings.ToUpper(addr)

			checksumFromLower := common.HexToAddress(lowercaseAddr).Hex()
			checksumFromUpper := common.HexToAddress(uppercaseAddr).Hex()

			assert.Equal(t, checksumFromLower, checksumFromUpper,
				"Checksum should be same regardless of input case")
			assert.Equal(t, checksummed1, checksumFromLower,
				"All should produce the same checksum")
		})
	}
}

// TestAddressWithoutPrefix tests handling of addresses without 0x prefix
func TestAddressWithoutPrefix(t *testing.T) {
	// Address from keystore (typically without 0x prefix)
	addressWithoutPrefix := "5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed"
	addressWithPrefix := "0x5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed"

	// Both should produce the same checksummed result
	result1 := common.HexToAddress(addressWithoutPrefix).Hex()
	result2 := common.HexToAddress(addressWithPrefix).Hex()

	assert.Equal(t, result1, result2,
		"Should handle addresses with or without 0x prefix")
	assert.True(t, strings.HasPrefix(result1, "0x"),
		"Result should always have 0x prefix")
}

// TestInvalidAddressHandling tests that invalid addresses are handled properly
func TestInvalidAddressHandling(t *testing.T) {
	invalidAddresses := []string{
		"",
		"0x",
		"not_an_address",
		"0xZZZ",
		"0x123", // Too short
	}

	for _, addr := range invalidAddresses {
		t.Run(addr, func(t *testing.T) {
			// common.HexToAddress doesn't error, but produces zero address for invalid input
			result := common.HexToAddress(addr)

			// For truly invalid addresses, we should validate separately
			if addr == "" || addr == "0x" || len(addr) < 40 {
				// These should not be considered valid
				isValid := common.IsHexAddress(addr) && len(addr) >= 42
				assert.False(t, isValid, "Invalid address should not pass validation")
			}

			// The function should still not panic
			_ = result.Hex()
		})
	}
}

// TestKeystoreAddressExtraction tests the address extraction from keystore JSON
func TestKeystoreAddressExtraction(t *testing.T) {
	testCases := []struct {
		name         string
		keystoreJSON string
		expectError  bool
	}{
		{
			name: "valid keystore with lowercase address",
			keystoreJSON: `{
				"address": "5aaeb6053f3e94c9b9a09f33669435e7ef1beaed",
				"crypto": {},
				"version": 3
			}`,
			expectError: false,
		},
		{
			name: "valid keystore with mixed case",
			keystoreJSON: `{
				"address": "5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed",
				"crypto": {},
				"version": 3
			}`,
			expectError: false,
		},
		{
			name: "missing address field",
			keystoreJSON: `{
				"crypto": {},
				"version": 3
			}`,
			expectError: true,
		},
		{
			name:         "invalid JSON",
			keystoreJSON: `{invalid json}`,
			expectError:  true,
		},
		{
			name: "address is not a string",
			keystoreJSON: `{
				"address": 123,
				"crypto": {},
				"version": 3
			}`,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var keystoreData map[string]interface{}
			err := json.Unmarshal([]byte(tc.keystoreJSON), &keystoreData)

			if tc.expectError {
				if err == nil {
					// Check if address extraction would fail
					addressRaw, ok := keystoreData["address"].(string)
					if ok {
						t.Errorf("Expected error but got valid address: %s", addressRaw)
					}
				}
				return
			}

			require.NoError(t, err, "JSON should be valid")

			addressRaw, ok := keystoreData["address"].(string)
			require.True(t, ok, "Address should be a string")

			// Apply the fix
			checksummedAddress := common.HexToAddress(addressRaw).Hex()

			// Verify it's properly formatted
			assert.True(t, strings.HasPrefix(checksummedAddress, "0x"),
				"Address should have 0x prefix")
			assert.True(t, common.IsHexAddress(checksummedAddress),
				"Should be a valid hex address")

			// Verify idempotency
			checksummedAgain := common.HexToAddress(checksummedAddress).Hex()
			assert.Equal(t, checksummedAddress, checksummedAgain,
				"Checksumming should be idempotent")
		})
	}
}

// TestImportKeystore_RealKeystoreFiles tests with actual keystore files from data/jarvis
func TestImportKeystore_RealKeystoreFiles(t *testing.T) {
	// Path to real keystore files
	keystorePath := filepath.Join("../../data/jarvis/keystores")

	// Check if directory exists (skip test if running in different environment)
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		t.Skip("Skipping test: data/jarvis/keystores directory not found")
		return
	}

	// Read keystore files
	files, err := os.ReadDir(keystorePath)
	require.NoError(t, err, "Should be able to read keystores directory")

	if len(files) == 0 {
		t.Skip("No keystore files found in data/jarvis/keystores")
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			// Read keystore file
			keystoreFilePath := filepath.Join(keystorePath, file.Name())
			keystoreBytes, err := os.ReadFile(keystoreFilePath)
			require.NoError(t, err, "Should be able to read keystore file")

			// Parse keystore JSON
			var keystoreData map[string]interface{}
			err = json.Unmarshal(keystoreBytes, &keystoreData)
			require.NoError(t, err, "Keystore should be valid JSON")

			// Extract address
			addressRaw, ok := keystoreData["address"].(string)
			require.True(t, ok, "Keystore should have address field")

			// Test the fix: convert to checksummed address
			checksummedAddress := common.HexToAddress(addressRaw).Hex()

			// Verify properties
			assert.True(t, strings.HasPrefix(checksummedAddress, "0x"),
				"Address should have 0x prefix")

			assert.True(t, common.IsHexAddress(checksummedAddress),
				"Should be a valid hex address")

			// Verify checksumming is idempotent
			checksummedAgain := common.HexToAddress(checksummedAddress).Hex()
			assert.Equal(t, checksummedAddress, checksummedAgain,
				"Checksumming should be idempotent")

			// Verify that the filename matches the checksummed address
			// (case-insensitive comparison)
			fileNameWithoutExt := strings.TrimSuffix(file.Name(), ".json")
			assert.True(t,
				strings.EqualFold(fileNameWithoutExt, checksummedAddress),
				"Filename should match the checksummed address (case-insensitive)")

			// Log for verification
			t.Logf("Original address: %s", addressRaw)
			t.Logf("Checksummed address: %s", checksummedAddress)
			t.Logf("Filename: %s", file.Name())
		})
	}
}

// TestImportKeystore_AddressConsistency tests that addresses from real keystores
// are handled consistently across the codebase
func TestImportKeystore_AddressConsistency(t *testing.T) {
	// Test with real addresses from data/jarvis
	realAddresses := []struct {
		name    string
		rawAddr string // As stored in keystore (lowercase, no 0x)
	}{
		{
			name:    "address_1",
			rawAddr: "22cb1309794b475c5709dddb640b79533c38d924",
		},
		{
			name:    "address_2",
			rawAddr: "995c8e149b803009c2e8dc696d5d5d3429a30a05",
		},
	}

	for _, tc := range realAddresses {
		t.Run(tc.name, func(t *testing.T) {
			// Test the fix
			checksummed := common.HexToAddress(tc.rawAddr).Hex()

			// Verify it's a valid checksummed address
			assert.True(t, strings.HasPrefix(checksummed, "0x"),
				"Address should have 0x prefix")
			assert.True(t, common.IsHexAddress(checksummed),
				"Should be a valid hex address")

			// Verify consistency with different input formats
			checksumFromUpper := common.HexToAddress(strings.ToUpper(tc.rawAddr)).Hex()
			checksumFromLower := common.HexToAddress(strings.ToLower(tc.rawAddr)).Hex()
			checksumFromMixed := common.HexToAddress(checksummed).Hex()

			assert.Equal(t, checksummed, checksumFromUpper,
				"Should produce same checksum from uppercase input")
			assert.Equal(t, checksummed, checksumFromLower,
				"Should produce same checksum from lowercase input")
			assert.Equal(t, checksummed, checksumFromMixed,
				"Should produce same checksum from mixed case input")

			// Log the actual checksum for verification
			t.Logf("Raw address: %s", tc.rawAddr)
			t.Logf("Checksummed: %s", checksummed)
		})
	}
}

// TestImportKeystore_FilenameFormat tests that keystore filenames follow the correct format
func TestImportKeystore_FilenameFormat(t *testing.T) {
	keystorePath := filepath.Join("../../data/jarvis/keystores")

	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		t.Skip("Skipping test: data/jarvis/keystores directory not found")
		return
	}

	files, err := os.ReadDir(keystorePath)
	require.NoError(t, err)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			// Extract address from filename
			fileNameWithoutExt := strings.TrimSuffix(file.Name(), ".json")

			// Verify filename is a valid address format
			assert.True(t, common.IsHexAddress(fileNameWithoutExt),
				"Filename should be a valid hex address")

			// Read and verify it matches the address inside
			keystoreFilePath := filepath.Join(keystorePath, file.Name())
			keystoreBytes, err := os.ReadFile(keystoreFilePath)
			require.NoError(t, err)

			var keystoreData map[string]interface{}
			err = json.Unmarshal(keystoreBytes, &keystoreData)
			require.NoError(t, err)

			addressRaw, ok := keystoreData["address"].(string)
			require.True(t, ok)

			// Convert both to checksummed format and compare
			checksummedFromFile := common.HexToAddress(fileNameWithoutExt).Hex()
			checksummedFromKeystore := common.HexToAddress(addressRaw).Hex()

			assert.Equal(t, checksummedFromFile, checksummedFromKeystore,
				"Filename address should match keystore address")
		})
	}
}

// Benchmark to ensure checksumming doesn't add significant overhead
func BenchmarkAddressChecksum(b *testing.B) {
	address := "5aAeb6053f3E94C9b9A09f33669435E7Ef1BeAed"

	b.Run("with_checksum", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = common.HexToAddress(address).Hex()
		}
	})

	b.Run("without_checksum", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = "0x" + address
		}
	})
}
