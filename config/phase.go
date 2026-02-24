package config

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kawai-network/contracts/kawaitoken"
	"github.com/kawai-network/contracts"
)

// Phase represents the current operational phase of the platform
type Phase int

const (
	Phase1 Phase = 1 // Free AI + KAWAI rewards
	Phase2 Phase = 2 // Paid AI (USDT) + Revenue sharing
)

// PhaseDetector manages phase detection and caching
type PhaseDetector struct {
	client       *ethclient.Client
	tokenAddress common.Address

	// Cache
	mu            sync.RWMutex
	cachedPhase   Phase
	lastCheck     time.Time
	cacheDuration time.Duration
}

// NewPhaseDetector creates a new phase detector
func NewPhaseDetector() (*PhaseDetector, error) {
	client, err := ethclient.Dial(contracts.MonadRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Monad: %w", err)
	}

	tokenAddr := common.HexToAddress(contracts.KawaiTokenAddress)

	return &PhaseDetector{
		client:        client,
		tokenAddress:  tokenAddr,
		cachedPhase:   Phase1,          // Default to Phase 1
		cacheDuration: 5 * time.Minute, // Cache for 5 minutes
	}, nil
}

// GetCurrentPhase returns the current phase (with caching)
func (pd *PhaseDetector) GetCurrentPhase(ctx context.Context) (Phase, error) {
	pd.mu.RLock()
	if time.Since(pd.lastCheck) < pd.cacheDuration {
		phase := pd.cachedPhase
		pd.mu.RUnlock()
		return phase, nil
	}
	pd.mu.RUnlock()

	// Cache expired, check blockchain
	phase, err := pd.detectPhase(ctx)
	if err != nil {
		return Phase1, err
	}

	pd.mu.Lock()
	pd.cachedPhase = phase
	pd.lastCheck = time.Now()
	pd.mu.Unlock()

	return phase, nil
}

// detectPhase checks blockchain to determine current phase
func (pd *PhaseDetector) detectPhase(ctx context.Context) (Phase, error) {
	kawaiToken, err := kawaitoken.NewKawaiToken(pd.tokenAddress, pd.client)
	if err != nil {
		return Phase1, fmt.Errorf("failed to load KawaiToken: %w", err)
	}

	callOpts := &bind.CallOpts{Context: ctx}

	// Get total supply
	totalSupply, err := kawaiToken.TotalSupply(callOpts)
	if err != nil {
		return Phase1, fmt.Errorf("failed to get total supply: %w", err)
	}

	// Get max supply (cap)
	maxSupply, err := kawaiToken.Cap(callOpts)
	if err != nil {
		return Phase1, fmt.Errorf("failed to get max supply: %w", err)
	}

	// Phase 2 triggers when totalSupply >= maxSupply
	if totalSupply.Cmp(maxSupply) >= 0 {
		log.Printf("🚀 Phase 2 Active: totalSupply=%s, maxSupply=%s", totalSupply.String(), maxSupply.String())
		return Phase2, nil
	}

	// Calculate percentage using big.Float to avoid overflow
	percentage := new(big.Float).Quo(
		new(big.Float).SetInt(totalSupply),
		new(big.Float).SetInt(maxSupply),
	)
	percentageFloat, _ := percentage.Float64()

	log.Printf("📊 Phase 1 Active: totalSupply=%s / maxSupply=%s (%.2f%%)",
		totalSupply.String(),
		maxSupply.String(),
		percentageFloat*100)

	return Phase1, nil
}

// IsPhase2 is a convenience method to check if Phase 2 is active
func (pd *PhaseDetector) IsPhase2(ctx context.Context) (bool, error) {
	phase, err := pd.GetCurrentPhase(ctx)
	if err != nil {
		return false, err
	}
	return phase == Phase2, nil
}

// ForceRefresh forces a cache refresh
func (pd *PhaseDetector) ForceRefresh(ctx context.Context) (Phase, error) {
	pd.mu.Lock()
	pd.lastCheck = time.Time{} // Reset cache
	pd.mu.Unlock()

	return pd.GetCurrentPhase(ctx)
}

// GetPhaseInfo returns detailed phase information
func (pd *PhaseDetector) GetPhaseInfo(ctx context.Context) (map[string]interface{}, error) {
	kawaiToken, err := kawaitoken.NewKawaiToken(pd.tokenAddress, pd.client)
	if err != nil {
		return nil, fmt.Errorf("failed to load KawaiToken: %w", err)
	}

	callOpts := &bind.CallOpts{Context: ctx}

	totalSupply, err := kawaiToken.TotalSupply(callOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get total supply: %w", err)
	}

	maxSupply, err := kawaiToken.Cap(callOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get max supply: %w", err)
	}

	phase, _ := pd.GetCurrentPhase(ctx)

	percentage := new(big.Float).Quo(
		new(big.Float).SetInt(totalSupply),
		new(big.Float).SetInt(maxSupply),
	)
	percentageFloat, _ := percentage.Float64()

	return map[string]interface{}{
		"current_phase":     int(phase),
		"phase_name":        fmt.Sprintf("Phase %d", phase),
		"total_supply":      totalSupply.String(),
		"max_supply":        maxSupply.String(),
		"supply_percentage": percentageFloat * 100,
		"is_phase_2":        phase == Phase2,
		"description":       getPhaseDescription(phase),
	}, nil
}

func getPhaseDescription(phase Phase) string {
	switch phase {
	case Phase1:
		return "Free AI usage with KAWAI token rewards"
	case Phase2:
		return "Paid AI usage (USDT) with revenue sharing to KAWAI holders"
	default:
		return "Unknown phase"
	}
}

// Global phase detector instance
var (
	globalDetector    *PhaseDetector
	globalDetectorErr error
	detectorOnce      sync.Once

	phaseDetectorFactory = NewPhaseDetector
)

// GetGlobalPhaseDetector returns the global phase detector instance
func GetGlobalPhaseDetector() (*PhaseDetector, error) {
	detectorOnce.Do(func() {
		globalDetector, globalDetectorErr = phaseDetectorFactory()
	})

	if globalDetectorErr != nil {
		return nil, globalDetectorErr
	}
	if globalDetector == nil {
		return nil, fmt.Errorf("phase detector is not initialized")
	}

	return globalDetector, nil
}

// IsPhase2Active is a convenience function using the global detector
func IsPhase2Active(ctx context.Context) (bool, error) {
	detector, err := GetGlobalPhaseDetector()
	if err != nil {
		return false, err
	}
	return detector.IsPhase2(ctx)
}
