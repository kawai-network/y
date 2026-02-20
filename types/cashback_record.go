package types

import "time"

// CashbackRecord stores detailed cashback information for a deposit
// This is used to:
// 1. Track cashback in KV store (pkg/store)
// 2. Send to Telegram for double-verification audit trail (pkg/alert)
// Defined in separate package to avoid circular dependency
type CashbackRecord struct {
	UserAddress    string    `json:"user_address"`
	TxHash         string    `json:"tx_hash"`
	DepositAmount  string    `json:"deposit_amount"`  // USDT amount (6 decimals)
	CashbackAmount string    `json:"cashback_amount"` // KAWAI amount (18 decimals)
	RateBPS        uint64    `json:"rate_bps"`        // Rate in basis points (e.g., 1000 = 10%)
	Tier           uint64    `json:"tier"`            // Cashback tier (1-4)
	IsFirstTime    bool      `json:"is_first_time"`   // First-time deposit bonus
	Period         uint64    `json:"period"`          // Settlement period
	Timestamp      time.Time `json:"timestamp"`
}
