package types

// ContributorStatus represents the operational status of a contributor.
type ContributorStatus string

const (
	// StatusOnline indicates the contributor is operational and accepting requests.
	StatusOnline ContributorStatus = "online"
	
	// StatusOffline indicates the contributor is not operational or unreachable.
	StatusOffline ContributorStatus = "offline"
)

// IsValid checks if the status is a valid value.
func (s ContributorStatus) IsValid() bool {
	return s == StatusOnline || s == StatusOffline
}

// String returns the string representation of the status.
func (s ContributorStatus) String() string {
	return string(s)
}

// HealthResponse represents the health status of a contributor.
type HealthResponse struct {
	Status         ContributorStatus `json:"status"`
	ActiveRequests int64             `json:"active_requests"`
	IsBusy         bool              `json:"is_busy"`
}