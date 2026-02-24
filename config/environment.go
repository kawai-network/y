package config

import (
	"fmt"
	"strings"

	"github.com/kawai-network/contracts"
)

// Environment represents the deployment environment
type Environment string

const (
	EnvironmentTestnet Environment = "testnet"
	EnvironmentMainnet Environment = "mainnet"
)

// Config holds the application configuration
type Config struct {
	Environment Environment
	IsTestnet   bool
	IsMainnet   bool
	ChainID     uint64
	NetworkName string
}

var currentConfig *Config

func configFromRPCURL(rpcURL string) (*Config, error) {
	cfg := &Config{}
	normalizedRPCURL := strings.ToLower(rpcURL)

	if strings.Contains(normalizedRPCURL, "testnet") {
		cfg.Environment = EnvironmentTestnet
		cfg.IsTestnet = true
		cfg.IsMainnet = false
		cfg.ChainID = 10143
		cfg.NetworkName = "Monad Testnet"
		return cfg, nil
	}

	// Default Monad public RPC host maps to mainnet.
	if strings.Contains(normalizedRPCURL, "mainnet") || strings.Contains(normalizedRPCURL, "rpc.monad.xyz") {
		cfg.Environment = EnvironmentMainnet
		cfg.IsTestnet = false
		cfg.IsMainnet = true
		cfg.ChainID = 143
		cfg.NetworkName = "Monad Mainnet"
		return cfg, nil
	}

	return nil, fmt.Errorf("unable to determine environment from RPC URL: %s", rpcURL)
}

// Initialize sets up the configuration based on environment variables
// This should be called once at application startup
func Initialize() error {
	rpcURL := contracts.MonadRpcUrl
	cfg, err := configFromRPCURL(rpcURL)
	if err != nil {
		return err
	}

	currentConfig = cfg
	return nil
}

// Get returns the current configuration
// Panics if Initialize() has not been called
func Get() *Config {
	if currentConfig == nil {
		panic("config not initialized - call config.Initialize() at startup")
	}
	return currentConfig
}

// IsTestnet returns true if running on testnet
func IsTestnet() bool {
	return Get().IsTestnet
}

// IsMainnet returns true if running on mainnet
func IsMainnet() bool {
	return Get().IsMainnet
}

// GetEnvironment returns the current environment
func GetEnvironment() Environment {
	return Get().Environment
}

// GetChainID returns the chain ID for the current environment
func GetChainID() uint64 {
	return Get().ChainID
}

// GetNetworkName returns the network name for the current environment
func GetNetworkName() string {
	return Get().NetworkName
}

// ValidateForProduction checks if configuration is safe for production
func ValidateForProduction() error {
	cfg := Get()

	if cfg.IsMainnet {
		// Mainnet-specific validations
		if contracts.StablecoinAddress == "0x3AE05118C5B75b1B0b860ec4b7Ec5095188D1CCc" {
			return fmt.Errorf("CRITICAL: Still using MockStablecoin address on mainnet! Update STABLECOIN_ADDRESS in .env.mainnet")
		}

		// Verify USDC address format
		if !strings.HasPrefix(contracts.StablecoinAddress, "0x") {
			return fmt.Errorf("invalid stablecoin address format: %s", contracts.StablecoinAddress)
		}

		// Check that we're not using testnet RPC
		if strings.Contains(contracts.MonadRpcUrl, "testnet") {
			return fmt.Errorf("CRITICAL: Using testnet RPC URL on mainnet configuration")
		}
	}

	return nil
}
