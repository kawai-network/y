package types

import "time"

// ReferralTrialRecord stores detailed information for a trial claim with referral
// This is used to:
// 1. Track trial claims in KV store (pkg/store)
// 2. Send to Telegram for double-verification audit trail (pkg/alert)
// Defined in separate package to avoid circular dependency
type ReferralTrialRecord struct {
	UserAddress     string    `json:"user_address"`
	ReferrerAddress string    `json:"referrer_address"`
	ReferralCode    string    `json:"referral_code"`
	TrialUSDT       string    `json:"trial_usdt"`     // Trial USDT amount (6 decimals)
	TrialKAWAI      string    `json:"trial_kawai"`    // Trial KAWAI amount (18 decimals)
	ReferrerBonus   string    `json:"referrer_bonus"` // Referrer bonus USDT (6 decimals)
	MachineID       string    `json:"machine_id"`     // Machine ID for anti-abuse
	IsReferral      bool      `json:"is_referral"`    // true if has referrer, false if solo
	Timestamp       time.Time `json:"timestamp"`
}
