package binding

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/kawai-network/contracts/cashbackdistributor"
	"github.com/kawai-network/contracts/kawaitoken"
	"github.com/kawai-network/contracts/miningdistributor"
	"github.com/kawai-network/contracts/mockstablecoin"
	"github.com/kawai-network/contracts/otcmarket"
	"github.com/kawai-network/contracts/referraldistributor"
	"github.com/kawai-network/contracts/revenuedistributor"
	"github.com/kawai-network/contracts/vault"
	"github.com/kawai-network/y/jarvis/util"
	"github.com/kawai-network/y/jarvis/util/reader"
)

// ResolveAddress resolves a string (hex or name) to a common.Address using Jarvis
func ResolveAddress(addrStr string) (common.Address, error) {
	addr, _, err := util.GetAddressFromString(addrStr)
	if err != nil {
		return common.Address{}, err
	}
	return common.HexToAddress(addr), nil
}

// KawaiToken wraps the generated KawaiToken binding
func KawaiToken(addrStr string, r *reader.EthReader) (*kawaitoken.KawaiToken, error) {
	addr, err := ResolveAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address %s: %w", addrStr, err)
	}
	backend := NewJarvisBackend(r)
	return kawaitoken.NewKawaiToken(addr, backend)
}

// OTCMarket wraps the generated OTCMarket binding
func OTCMarket(addrStr string, r *reader.EthReader) (*otcmarket.OTCMarket, error) {
	addr, err := ResolveAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address %s: %w", addrStr, err)
	}
	backend := NewJarvisBackend(r)
	return otcmarket.NewOTCMarket(addr, backend)
}

// Vault wraps the generated PaymentVault binding
func Vault(addrStr string, r *reader.EthReader) (*vault.PaymentVault, error) {
	addr, err := ResolveAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address %s: %w", addrStr, err)
	}
	backend := NewJarvisBackend(r)
	return vault.NewPaymentVault(addr, backend)
}

// RevenueDistributor wraps the generated RevenueDistributor binding for revenue sharing to KAWAI holders
func RevenueDistributor(addrStr string, r *reader.EthReader) (*revenuedistributor.RevenueDistributor, error) {
	addr, err := ResolveAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address %s: %w", addrStr, err)
	}
	backend := NewJarvisBackend(r)
	return revenuedistributor.NewRevenueDistributor(addr, backend)
}

// MiningRewardDistributor wraps the generated MiningRewardDistributor binding
// This contract supports referral-based mining rewards with flexible developer addresses
func MiningRewardDistributor(addrStr string, r *reader.EthReader) (*miningdistributor.MiningRewardDistributor, error) {
	addr, err := ResolveAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address %s: %w", addrStr, err)
	}
	backend := NewJarvisBackend(r)
	return miningdistributor.NewMiningRewardDistributor(addr, backend)
}

// CashbackDistributor wraps the generated DepositCashbackDistributor binding
// This contract distributes KAWAI cashback rewards for USDT deposits
func CashbackDistributor(addrStr string, r *reader.EthReader) (*cashbackdistributor.DepositCashbackDistributor, error) {
	addr, err := ResolveAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address %s: %w", addrStr, err)
	}
	backend := NewJarvisBackend(r)
	return cashbackdistributor.NewDepositCashbackDistributor(addr, backend)
}

// ReferralRewardDistributor wraps the generated ReferralRewardDistributor binding
// This contract distributes KAWAI referral commission rewards
func ReferralRewardDistributor(addrStr string, r *reader.EthReader) (*referraldistributor.ReferralRewardDistributor, error) {
	addr, err := ResolveAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address %s: %w", addrStr, err)
	}
	backend := NewJarvisBackend(r)
	return referraldistributor.NewReferralRewardDistributor(addr, backend)
}

// Stablecoin wraps any ERC-20 stablecoin token (MockStablecoin on testnet, USDC on mainnet)
// Uses KawaiToken binding since it implements standard ERC-20 interface
func Stablecoin(addrStr string, r *reader.EthReader) (*kawaitoken.KawaiToken, error) {
	addr, err := ResolveAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve stablecoin address %s: %w", addrStr, err)
	}
	backend := NewJarvisBackend(r)
	return kawaitoken.NewKawaiToken(addr, backend)
}

// MockStablecoin wraps the MockStablecoin contract (testnet only)
func MockStablecoin(addrStr string, r *reader.EthReader) (*mockstablecoin.MockStablecoin, error) {
	addr, err := ResolveAddress(addrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve MockStablecoin address %s: %w", addrStr, err)
	}
	backend := NewJarvisBackend(r)
	return mockstablecoin.NewMockStablecoin(addr, backend)
}
