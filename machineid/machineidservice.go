package machineid

// Service provides access to machine identification functionality
type Service struct{}

// GetID returns the platform specific machine ID of the current host OS
func (m *Service) GetID() (string, error) {
	return ID()
}

// GetProtectedID returns a hashed version of the machine ID in a cryptographically secure way,
// using a fixed, application-specific key.
// Internally, this function calculates HMAC-SHA256 of the application ID, keyed by the machine ID.
func (m *Service) GetProtectedID(appID string) (string, error) {
	return ProtectedID(appID)
}
