package types

import "time"

// JobRewardRecord stores detailed reward split for a single job
// This is used to:
// 1. Generate 9-field Merkle leaves for MiningRewardDistributor (pkg/store)
// 2. Send to Telegram for double-verification audit trail (pkg/alert)
// Defined in separate package to avoid circular dependency
type JobRewardRecord struct {
	Timestamp          time.Time `json:"timestamp"`
	ContributorAddress string    `json:"contributor_address"`
	UserAddress        string    `json:"user_address"`
	ReferrerAddress    string    `json:"referrer_address"`  // Empty if non-referral
	DeveloperAddress   string    `json:"developer_address"` // From GetRandomTreasuryAddress()

	ContributorAmount string `json:"contributor_amount"` // Contributor reward amount (85% or 90% of total)
	DeveloperAmount   string `json:"developer_amount"`   // Developer reward amount (5% of total)
	UserAmount        string `json:"user_amount"`        // User reward amount (5% of total)
	AffiliatorAmount  string `json:"affiliator_amount"`  // Affiliator reward amount (5% of total or 0)

	TokenUsage  int64  `json:"token_usage"`
	RewardType  RewardType `json:"reward_type"` // "mining", "cashback", "referral", "stablecoin"
	HasReferrer bool   `json:"has_referrer"`

	// For tracking settlement
	SettledPeriodID int64 `json:"settled_period_id,omitempty"`
	IsSettled       bool  `json:"is_settled,omitempty"`
}
