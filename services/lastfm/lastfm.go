package lastfm

import (
	"fmt"
	"os"

	"github.com/shkh/lastfm-go/lastfm"
)

// Lastfm is the external Lastfm service implementation.
type Lastfm struct {
	api *lastfm.Api
}

// NewLastfm creates a Lastfm instance.
func NewLastfm() (*Lastfm, error) {
	clientID := os.Getenv("LASTFM_CLIENT_ID")
	clientSecret := os.Getenv("LASTFM_CLIENT_SECRET")

	if len(clientID) == 0 || len(clientSecret) == 0 {
		return nil, fmt.Errorf("please set LASTFM_CLIENT_ID and LASTFM_CLIENT_SECRET environment variables")
	}

	return &Lastfm{
		api: lastfm.New(clientID, clientSecret),
	}, nil
}

// Name returns the human readable service name.
func (l *Lastfm) Name() string {
	return "Last.fm"
}

// CreateAuthURL returns an authorization URL to authorize the integration.
func (l *Lastfm) CreateAuthURL() string {
	// Not an actual web server (yet).
	redirectURL := "https://admirer.test"

	return l.api.GetAuthRequestUrl(redirectURL)
}

// Authenticate takes an authorization code and authenticates the user.
func (l *Lastfm) Authenticate(oauthCode string) error {
	if err := l.api.LoginWithToken(oauthCode); err != nil {
		return fmt.Errorf("failed to parse Last.fm token")
	}

	return nil
}

// GetUsername requests and returns the username of the logged in user.
func (l *Lastfm) GetUsername() (string, error) {
	user, err := l.api.User.GetInfo(lastfm.P{})
	if err != nil {
		return "", fmt.Errorf("failed to read Last.fm profile data")
	}

	return user.Name, nil
}
