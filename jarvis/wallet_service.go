package jarvis

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"

	"github.com/kawai-network/x/blockchain"
	"github.com/kawai-network/x/store"
	"github.com/kawai-network/y/jarvis/accounts"
	"github.com/kawai-network/y/jarvis/accounts/types"
	"github.com/kawai-network/y/jarvis/util/account"
	"github.com/kawai-network/y/machineid"
	"github.com/kawai-network/y/paths"
)

// WalletInfo represents a wallet in the list
type WalletInfo struct {
	Address     string `json:"address"`
	Description string `json:"description"`
	IsActive    bool   `json:"isActive"`
}

// WalletStatus represents the current state of the wallet
type WalletStatus struct {
	HasWallet bool         `json:"hasWallet"`
	IsLocked  bool         `json:"isLocked"`
	Address   string       `json:"address"`
	Wallets   []WalletInfo `json:"wallets"`
}

// WalletService handles local wallet management using jarvis/accounts
type WalletService struct {
	mu             sync.RWMutex
	currentAccount *account.Account
	address        string
	kvStore        *store.KVStore
	holderRegistry *blockchain.HolderRegistry
}

// NewWalletService creates a new wallet service
func NewWalletService(dataDir string, kvStore *store.KVStore) *WalletService {
	var holderRegistry *blockchain.HolderRegistry
	if kvStore != nil {
		holderRegistry = blockchain.NewHolderRegistry(kvStore)
	}
	return &WalletService{
		kvStore:        kvStore,
		holderRegistry: holderRegistry,
	}
}

// HasWallet checks if a wallet already exists
func (s *WalletService) HasWallet() bool {
	accs := accounts.GetAccounts()
	return len(accs) > 0
}

// GetWallets returns a list of all stored wallets
func (s *WalletService) GetWallets() []WalletInfo {
	s.mu.RLock()
	currentAddr := s.address
	s.mu.RUnlock()

	accs := accounts.GetAccounts()
	result := make([]WalletInfo, 0, len(accs))
	for addr, acc := range accs {
		result = append(result, WalletInfo{
			Address:     addr,
			Description: acc.Desc,
			IsActive:    currentAddr == addr,
		})
	}
	return result
}

// GetAPIKey returns the API key for the current wallet (generating one if needed)
func (s *WalletService) GetAPIKey() (string, error) {
	if s.address == "" {
		return "", nil // No active wallet
	}
	if s.kvStore == nil {
		return "", errors.New("kvstore not available")
	}

	ctx := context.Background()
	// 1. Try reverse lookup
	apiKey, err := s.kvStore.GetAPIKeyByAddress(ctx, s.address)
	if err == nil && apiKey != "" {
		return apiKey, nil
	}

	// 2. Generate new key
	newKey, err := s.kvStore.CreateAPIKey(ctx, s.address)
	if err != nil {
		return "", fmt.Errorf("failed to create api key: %w", err)
	}

	return newKey.Key, nil
}

// GenerateMnemonic creates a new 12-word bip39 mnemonic
func (s *WalletService) GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

// CreateWallet creates a new wallet (supports multiple wallets)
func (s *WalletService) CreateWallet(password string, mnemonic string, description string) (string, error) {
	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return "", errors.New("invalid mnemonic")
	}

	// Derive private key from mnemonic
	seed := bip39.NewSeed(mnemonic, "")
	masterKey, err := crypto.ToECDSA(seed[:32])
	if err != nil {
		return "", fmt.Errorf("failed to derive key: %v", err)
	}

	privBytes := crypto.FromECDSA(masterKey)
	privHex := fmt.Sprintf("%x", privBytes)
	address := crypto.PubkeyToAddress(masterKey.PublicKey).Hex()

	// Check if wallet already exists
	accs := accounts.GetAccounts()
	if _, exists := accs[address]; exists {
		return "", errors.New("wallet with this address already exists")
	}

	// Store keystore
	keystorePath, err := accounts.StorePrivateKeyWithKeystore(privHex, password)
	if err != nil {
		return "", fmt.Errorf("failed to store keystore: %v", err)
	}

	// Verify keystore
	verifiedAddr, err := accounts.VerifyKeystore(keystorePath)
	if err != nil {
		return "", fmt.Errorf("failed to verify keystore: %v", err)
	}

	// Use default description if empty
	if description == "" {
		description = fmt.Sprintf("Wallet %d", len(accs)+1)
	}

	// Store account metadata
	accDesc := types.AccDesc{
		Address: verifiedAddr,
		Kind:    "keystore",
		Keypath: keystorePath,
		Desc:    description,
	}
	if err := accounts.StoreAccountRecord(accDesc); err != nil {
		return "", fmt.Errorf("failed to store account record: %v", err)
	}

	// Auto-unlock the new wallet
	acc, err := accounts.UnlockKeystoreAccountWithPassword(accDesc, password)
	if err != nil {
		return "", fmt.Errorf("failed to unlock account: %v", err)
	}

	s.mu.Lock()
	s.currentAccount = acc
	s.address = address
	s.mu.Unlock()

	// Register holder in registry
	s.registerHolderAsync(address)

	return address, nil
}

// SetupWallet creates a new keystore from a password and mnemonic (first wallet only)
func (s *WalletService) SetupWallet(password string, mnemonic string, name string) (string, error) {
	if s.HasWallet() {
		return "", errors.New("wallet already exists")
	}
	if name == "" {
		name = "My Wallet"
	}
	return s.CreateWallet(password, mnemonic, name)
}

// SwitchWallet switches to a different wallet by address
func (s *WalletService) SwitchWallet(address string, password string) (string, error) {
	accs := accounts.GetAccounts()
	accDesc, exists := accs[address]
	if !exists {
		return "", errors.New("wallet not found")
	}

	acc, err := accounts.UnlockKeystoreAccountWithPassword(accDesc, password)
	if err != nil {
		return "", errors.New("invalid password")
	}

	s.mu.Lock()
	s.currentAccount = acc
	s.address = acc.AddressHex()
	s.mu.Unlock()

	// Register holder in registry
	s.registerHolderAsync(s.address)

	return s.address, nil
}

// UnlockWallet decrypts the keystore and loads it into memory
func (s *WalletService) UnlockWallet(password string) (string, error) {
	accs := accounts.GetAccounts()
	if len(accs) == 0 {
		return "", errors.New("no wallet found")
	}

	// Get the first account if none is active
	var accDesc types.AccDesc
	for _, acc := range accs {
		accDesc = acc
		break
	}

	acc, err := accounts.UnlockKeystoreAccountWithPassword(accDesc, password)
	if err != nil {
		return "", errors.New("invalid password")
	}

	s.mu.Lock()
	s.currentAccount = acc
	s.address = acc.AddressHex()
	s.mu.Unlock()

	// Register holder in registry
	s.registerHolderAsync(s.address)

	return s.address, nil
}

// LockWallet clears the private key from memory
func (s *WalletService) LockWallet() {
	s.mu.Lock()
	s.currentAccount = nil
	s.address = ""
	s.mu.Unlock()
}

// DeleteWallet removes a wallet from storage
func (s *WalletService) DeleteWallet(address string) error {
	accs := accounts.GetAccounts()
	accDesc, exists := accs[address]
	if !exists {
		return errors.New("wallet not found")
	}

	// Cannot delete active wallet
	if s.address == address {
		return errors.New("cannot delete active wallet, switch to another wallet first")
	}

	// Delete keystore file
	if err := os.Remove(accDesc.Keypath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete keystore: %v", err)
	}

	// Delete account record (metadata file)
	metadataPath := filepath.Join(paths.Jarvis(), fmt.Sprintf("%s.json", address))
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete metadata: %v", err)
	}

	return nil
}

// ExportKeystore returns the keystore JSON content for a wallet
func (s *WalletService) ExportKeystore(address string) (string, error) {
	accs := accounts.GetAccounts()
	accDesc, exists := accs[address]
	if !exists {
		return "", errors.New("wallet not found")
	}

	content, err := os.ReadFile(accDesc.Keypath)
	if err != nil {
		return "", fmt.Errorf("failed to read keystore: %v", err)
	}

	return string(content), nil
}

// ImportKeystore imports a keystore from JSON content
func (s *WalletService) ImportKeystore(keystoreJSON string, password string, description string) (string, error) {
	// Validate keystore by trying to decrypt it
	var keystoreData map[string]interface{}
	if err := json.Unmarshal([]byte(keystoreJSON), &keystoreData); err != nil {
		return "", errors.New("invalid keystore JSON format")
	}

	// Get address from keystore
	addressRaw, ok := keystoreData["address"].(string)
	if !ok {
		return "", errors.New("keystore missing address field")
	}
	// Convert to checksummed address (EIP-55)
	address := common.HexToAddress(addressRaw).Hex()

	// Check if wallet already exists
	accs := accounts.GetAccounts()
	if _, exists := accs[address]; exists {
		return "", errors.New("wallet with this address already exists")
	}

	// Save keystore to file
	keystoreDir := paths.JarvisKeystores()
	os.MkdirAll(keystoreDir, 0755)
	keystorePath := filepath.Join(keystoreDir, fmt.Sprintf("%s.json", address))

	if err := os.WriteFile(keystorePath, []byte(keystoreJSON), 0600); err != nil {
		return "", fmt.Errorf("failed to save keystore: %v", err)
	}

	// Use default description if empty
	if description == "" {
		description = fmt.Sprintf("Imported Wallet %d", len(accs)+1)
	}

	// Store account metadata
	accDesc := types.AccDesc{
		Address: address,
		Kind:    "keystore",
		Keypath: keystorePath,
		Desc:    description,
	}
	if err := accounts.StoreAccountRecord(accDesc); err != nil {
		return "", fmt.Errorf("failed to store account record: %v", err)
	}

	// Verify by unlocking
	acc, err := accounts.UnlockKeystoreAccountWithPassword(accDesc, password)
	if err != nil {
		// Rollback: delete the keystore file and metadata
		_ = os.Remove(keystorePath)
		_ = os.Remove(filepath.Join(paths.Jarvis(), fmt.Sprintf("%s.json", address)))
		return "", errors.New("invalid password for keystore")
	}

	s.mu.Lock()
	s.currentAccount = acc
	s.address = address
	s.mu.Unlock()

	// Register holder in registry
	s.registerHolderAsync(address)

	return address, nil
}

// ImportPrivateKey imports a wallet from a private key
func (s *WalletService) ImportPrivateKey(privateKeyHex string, password string, description string) (string, error) {
	// Remove 0x prefix if present
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	// Validate private key length
	if len(privateKeyHex) != 64 {
		return "", errors.New("invalid private key: must be 64 hex characters")
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %v", err)
	}

	// Derive address from private key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("failed to derive public key")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	// Check if wallet already exists
	accs := accounts.GetAccounts()
	if _, exists := accs[address]; exists {
		return "", errors.New("wallet with this address already exists")
	}

	// Create encrypted keystore
	ks := keystore.NewKeyStore(paths.JarvisKeystores(), keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.ImportECDSA(privateKey, password)
	if err != nil {
		return "", fmt.Errorf("failed to create keystore: %v", err)
	}

	// Get keystore file path
	keystorePath := account.URL.Path

	// Use default description if empty
	if description == "" {
		description = fmt.Sprintf("Imported Wallet %d", len(accs)+1)
	}

	// Store account metadata
	accDesc := types.AccDesc{
		Address: address,
		Kind:    "keystore",
		Keypath: keystorePath,
		Desc:    description,
	}
	if err := accounts.StoreAccountRecord(accDesc); err != nil {
		// Rollback: delete the keystore file
		_ = os.Remove(keystorePath)
		return "", fmt.Errorf("failed to store account record: %v", err)
	}

	// Unlock and set as current account
	acc, err := accounts.UnlockKeystoreAccountWithPassword(accDesc, password)
	if err != nil {
		// Rollback
		_ = os.Remove(keystorePath)
		_ = os.Remove(filepath.Join(paths.Jarvis(), fmt.Sprintf("%s.json", address)))
		return "", fmt.Errorf("failed to unlock imported wallet: %v", err)
	}

	s.mu.Lock()
	s.currentAccount = acc
	s.address = address
	s.mu.Unlock()

	// Register holder in registry
	s.registerHolderAsync(address)

	return address, nil
}

// UpdateWalletDescription updates the description of a wallet
func (s *WalletService) UpdateWalletDescription(address string, description string) error {
	accs := accounts.GetAccounts()
	accDesc, exists := accs[address]
	if !exists {
		return errors.New("wallet not found")
	}

	accDesc.Desc = description
	return accounts.StoreAccountRecord(accDesc)
}

// GetStatus returns the current wallet status
func (s *WalletService) GetStatus() WalletStatus {
	s.mu.RLock()
	isLocked := s.currentAccount == nil
	address := s.address
	s.mu.RUnlock()

	return WalletStatus{
		HasWallet: s.HasWallet(),
		IsLocked:  isLocked,
		Address:   address,
		Wallets:   s.GetWallets(),
	}
}

// GetCurrentAddress returns the current active wallet address
func (s *WalletService) GetCurrentAddress() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.address
}

// IsUnlocked returns true if the wallet is currently unlocked
func (s *WalletService) IsUnlocked() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentAccount != nil
}

// GetCurrentAccountAddress returns the current account address if unlocked
func (s *WalletService) GetCurrentAccountAddress() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.currentAccount != nil {
		return s.currentAccount.AddressHex()
	}
	return ""
}

// SignMessage signs a message with the private key
func (s *WalletService) SignMessage(message string) (string, error) {
	s.mu.RLock()
	acc := s.currentAccount
	s.mu.RUnlock()

	if acc == nil {
		return "", errors.New("wallet is locked")
	}

	return acc.SignMessage(message)
}

// GetTransactOpts creates a bind.TransactOpts for the current account
// This is used for blockchain transactions
//
//wails:ignore - Not used in frontend, only used by backend services
func (s *WalletService) GetTransactOpts(chainId *big.Int) (*bind.TransactOpts, error) {
	s.mu.RLock()
	acc := s.currentAccount
	s.mu.RUnlock()

	if acc == nil {
		return nil, errors.New("wallet is locked")
	}

	return &bind.TransactOpts{
		From: acc.Address(),
		Signer: func(addr common.Address, tx *ethtypes.Transaction) (*ethtypes.Transaction, error) {
			_, signedTx, err := acc.SignTx(tx, chainId)
			return signedTx, err
		},
	}, nil
}

// registerHolderAsync registers the current wallet address as a KAWAI holder (non-blocking)
func (s *WalletService) registerHolderAsync(address string) {
	if s.holderRegistry == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.holderRegistry.RegisterHolder(ctx, common.HexToAddress(address), "desktop"); err != nil {
			// Log but don't fail - holder registration is best-effort
			fmt.Printf("⚠️ Failed to register holder: %v\n", err)
		}
	}()
}

// AutoClaimTrialIfNeeded checks if user needs to claim trial and claims it automatically
// This should be called after successful wallet unlock
// Returns: (claimed bool, usdtAmount float64, kawaiAmount string, error)
func (s *WalletService) AutoClaimTrialIfNeeded(referralCode string) (bool, float64, string, error) {
	if s.kvStore == nil {
		return false, 0, "0", errors.New("kvstore not available")
	}

	s.mu.RLock()
	address := s.address
	s.mu.RUnlock()

	if address == "" {
		return false, 0, "0", errors.New("no active wallet")
	}

	ctx := context.Background()

	// Get machine ID for anti-abuse (required)
	machineID, err := s.getMachineID()
	if err != nil {
		return false, 0, "0", fmt.Errorf("machine id unavailable: %w", err)
	}

	// Check if trial already claimed (address or machine)
	hasClaimed, err := s.kvStore.HasClaimedTrial(ctx, address, machineID)
	if err != nil {
		return false, 0, "0", fmt.Errorf("failed to check trial status: %w", err)
	}

	if hasClaimed {
		return false, 0, "0", nil // Already claimed, no error
	}

	// Claim trial with optional referral code
	usdtAmount, kawaiAmount, err := s.kvStore.ClaimFreeTrialWithReferral(ctx, address, machineID, referralCode)
	if err != nil {
		return false, 0, "0", fmt.Errorf("failed to claim trial: %w", err)
	}

	// Convert micro USDT to USDT
	usdtFloat := float64(usdtAmount) / 1_000_000

	return true, usdtFloat, kawaiAmount, nil
}

// UserBalanceInfo represents user balance data formatted for UI display
type UserBalanceInfo struct {
	Address         string `json:"address"`
	USDTBalance     string `json:"usdt_balance"`  // Formatted (e.g., "5.00")
	USDTMicro       string `json:"usdt_micro"`    // Raw micro USDT value
	KawaiBalance    string `json:"kawai_balance"` // In wei
	TrialClaimed    bool   `json:"trial_claimed"`
	ReferrerAddress string `json:"referrer_address"` // Empty if no referrer
	HasReferrer     bool   `json:"has_referrer"`
}

// GetUserBalanceInfo retrieves and formats user balance for UI display
// Returns nil if no wallet is active or kvstore is not available
func (s *WalletService) GetUserBalanceInfo() (*UserBalanceInfo, error) {
	if s.kvStore == nil {
		return nil, errors.New("kvstore not available")
	}

	s.mu.RLock()
	address := s.address
	s.mu.RUnlock()

	if address == "" {
		return nil, nil // No active wallet - not an error, just no data
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	balance, err := s.kvStore.GetUserBalance(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	// Convert micro USDT to display format (2 decimal places)
	usdtMicro := new(big.Int)
	if balance.USDTBalance != "" {
		usdtMicro.SetString(balance.USDTBalance, 10)
	}

	// Convert to USDT with 2 decimals: divide by 10^4 to get cents, then format
	usdtCents := new(big.Int).Div(usdtMicro, big.NewInt(10000))
	usdtFloat := float64(usdtCents.Int64()) / 100.0

	return &UserBalanceInfo{
		Address:         address,
		USDTBalance:     fmt.Sprintf("%.2f", usdtFloat),
		USDTMicro:       balance.USDTBalance,
		KawaiBalance:    balance.KawaiBalance,
		TrialClaimed:    balance.TrialClaimed,
		ReferrerAddress: balance.ReferrerAddress,
		HasReferrer:     balance.ReferrerAddress != "" && balance.ReferrerAddress != "0x0000000000000000000000000000000000000000",
	}, nil
}

// GetBalanceInUSDT returns the USDT balance as a float value
// Useful for quick balance checks in the UI
//
//wails:ignore - Not currently used in frontend, available for future use
func (s *WalletService) GetBalanceInUSDT() (float64, error) {
	info, err := s.GetUserBalanceInfo()
	if err != nil {
		return 0, err
	}
	if info == nil {
		return 0, nil
	}

	var balance float64
	_, err = fmt.Sscanf(info.USDTBalance, "%f", &balance)
	if err != nil {
		return 0, fmt.Errorf("failed to parse balance: %s", info.USDTBalance)
	}
	return balance, nil
}

// BalanceCheckResult represents the result of a balance sufficiency check
//
//wails:ignore - Not currently used in frontend, available for future use
type BalanceCheckResult struct {
	HasSufficientBalance bool   `json:"has_sufficient_balance"`
	CurrentBalanceUSDT   string `json:"current_balance_usdt"`  // Formatted (e.g., "5.00")
	CurrentBalanceMicro  string `json:"current_balance_micro"` // Raw micro USDT
	EstimatedCostUSDT    string `json:"estimated_cost_usdt"`   // Formatted
	EstimatedCostMicro   string `json:"estimated_cost_micro"`  // Raw micro USDT
	EstimatedTokens      int64  `json:"estimated_tokens"`
	Message              string `json:"message"`
}

// CheckBalanceForAI checks if user has sufficient balance for estimated AI usage
// This is useful for pre-flight checks in the UI before making AI calls
// estimatedTokens: rough estimate of tokens needed (4 chars = 1 token)
//
//wails:ignore - Not currently used in frontend, available for future use
func (s *WalletService) CheckBalanceForAI(estimatedTokens int64) (*BalanceCheckResult, error) {
	if s.kvStore == nil {
		return nil, errors.New("kvstore not available")
	}

	s.mu.RLock()
	address := s.address
	s.mu.RUnlock()

	if address == "" {
		return &BalanceCheckResult{
			HasSufficientBalance: false,
			Message:              "No wallet connected",
		}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get current balance
	balance, err := s.kvStore.GetUserBalance(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	currentMicro := new(big.Int)
	if balance.USDTBalance != "" {
		currentMicro.SetString(balance.USDTBalance, 10)
	}

	// Calculate estimated cost
	cost := store.CalculateUsageCost(estimatedTokens)

	// Format balances for display
	currentCents := new(big.Int).Div(currentMicro, big.NewInt(10000))
	currentFloat := float64(currentCents.Int64()) / 100.0

	costCents := new(big.Int).Div(cost, big.NewInt(10000))
	costFloat := float64(costCents.Int64()) / 100.0

	result := &BalanceCheckResult{
		HasSufficientBalance: currentMicro.Cmp(cost) >= 0,
		CurrentBalanceUSDT:   fmt.Sprintf("%.2f", currentFloat),
		CurrentBalanceMicro:  currentMicro.String(),
		EstimatedCostUSDT:    fmt.Sprintf("%.2f", costFloat),
		EstimatedCostMicro:   cost.String(),
		EstimatedTokens:      estimatedTokens,
	}

	if result.HasSufficientBalance {
		result.Message = fmt.Sprintf("Sufficient balance: %s USDT available, %s USDT needed",
			result.CurrentBalanceUSDT, result.EstimatedCostUSDT)
	} else {
		result.Message = fmt.Sprintf("Insufficient balance: %s USDT available, %s USDT needed. Please deposit to continue.",
			result.CurrentBalanceUSDT, result.EstimatedCostUSDT)
	}

	return result, nil
}

// getMachineID returns the machine ID for anti-abuse protection
func (s *WalletService) getMachineID() (string, error) {
	return machineid.ProtectedID("veridium-desktop")
}
