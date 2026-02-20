package types

// RewardType represents the type of reward being distributed
type RewardType string

const (
	// RewardTypeMining represents mining rewards (KAWAI tokens)
	RewardTypeMining RewardType = "mining"

	// RewardTypeCashback represents deposit cashback rewards (KAWAI tokens)
	RewardTypeCashback RewardType = "cashback"

	// RewardTypeReferral represents referral rewards (KAWAI tokens)
	RewardTypeReferral RewardType = "referral"

	// RewardTypeRevenue represents revenue sharing rewards (stablecoin)
	RewardTypeRevenue RewardType = "revenue"
)

// String returns the string representation of RewardType
func (r RewardType) String() string {
	return string(r)
}

// IsKawaiReward returns true if the reward type is KAWAI token based
func (r RewardType) IsKawaiReward() bool {
	return r == RewardTypeMining || r == RewardTypeCashback || r == RewardTypeReferral
}

// IsStablecoinReward returns true if the reward type is stablecoin based
func (r RewardType) IsStablecoinReward() bool {
	return r == RewardTypeRevenue
}

// Decimals returns the number of decimals for this reward type
func (r RewardType) Decimals() int {
	if r.IsStablecoinReward() {
		return 6
	}
	return 18
}

// Normalize converts legacy reward types to their canonical form
// Note: No longer needed since all legacy types have been merged, but kept for compatibility
func (r RewardType) Normalize() RewardType {
	// All types are already canonical
	return r
}
