//go:generate mockgen -source services.go -destination services_mock.go -package domain

package domain

// Service is the external service interface.
type Service interface {
	Name() string
	Authenticated() bool
	CreateAuthURL(redirectURL string) string
	CodeParam() string
	Authenticate(code string, redirectURL string) error
	GetUsername() (string, error)
	GetLovedTracks(limit int, page int) ([]Track, error)
	LoveTrack(track Track) error
	Close() error
}

// ServiceLoader loads service instances by name.
type ServiceLoader interface {
	ForName(serviceName string) (Service, error)
	Names() []string
}
