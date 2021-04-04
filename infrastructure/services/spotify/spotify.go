//go:generate mockgen -source spotify.go -destination spotify_mock.go -package spotify

package spotify

import (
	"errors"
	"os"
	"time"

	"github.com/dietrichm/admirer/domain"
	"github.com/dietrichm/admirer/infrastructure/config"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

// Authenticator is our interface for a Spotify authenticator.
type Authenticator interface {
	SetAuthInfo(clientID, secretKey string)
	AuthURL(state string) string
	Exchange(string, ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	NewClient(token *oauth2.Token) spotify.Client
}

// Client is our interface for a Spotify client.
type Client interface {
	CurrentUser() (*spotify.PrivateUser, error)
	Token() (*oauth2.Token, error)
	CurrentUsersTracksOpt(opt *spotify.Options) (*spotify.SavedTrackPage, error)
}

// Spotify is the external Spotify service implementation.
type Spotify struct {
	authenticator Authenticator
	client        Client
	secrets       config.Config
}

// NewSpotify creates a Spotify instance.
func NewSpotify(secrets config.Config) (*Spotify, error) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	if len(clientID) == 0 || len(clientSecret) == 0 {
		return nil, errors.New("please set SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables")
	}

	// Not an actual web server (yet).
	redirectURL := "https://admirer.test"
	authenticator := spotify.NewAuthenticator(redirectURL, spotify.ScopeUserReadPrivate, spotify.ScopeUserLibraryRead)
	authenticator.SetAuthInfo(clientID, clientSecret)

	service := &Spotify{
		authenticator: &authenticator,
		secrets:       secrets,
	}
	service.authenticateFromSecrets(secrets)

	return service, nil
}

// Name returns the human readable service name.
func (s *Spotify) Name() string {
	return "Spotify"
}

// Authenticated returns whether the service is logged in.
func (s *Spotify) Authenticated() bool {
	if s.client != nil {
		return true
	}
	return false
}

// CreateAuthURL returns an authorization URL to authorize the integration.
func (s *Spotify) CreateAuthURL() string {
	return s.authenticator.AuthURL("")
}

// Authenticate takes an authorization code and authenticates the user.
func (s *Spotify) Authenticate(code string) error {
	token, err := s.authenticator.Exchange(code)
	if err != nil {
		return errors.New("failed to parse Spotify token")
	}

	client := s.authenticator.NewClient(token)
	s.client = &client

	return nil
}

// GetUsername requests and returns the username of the logged in user.
func (s *Spotify) GetUsername() (string, error) {
	user, err := s.client.CurrentUser()
	if err != nil {
		return "", errors.New("failed to read Spotify profile data")
	}

	return user.DisplayName, nil
}

// GetLovedTracks returns loved tracks from the external service.
func (s *Spotify) GetLovedTracks(limit int) (tracks []domain.Track, err error) {
	options := &spotify.Options{
		Limit: &limit,
	}

	result, err := s.client.CurrentUsersTracksOpt(options)
	if err != nil {
		return
	}

	for _, resultTrack := range result.Tracks {
		track := domain.Track{
			Artist: resultTrack.Artists[0].Name,
			Name:   resultTrack.Name,
		}
		tracks = append(tracks, track)
	}
	return
}

// Close persists any state before quitting the application.
func (s *Spotify) Close() error {
	if !s.Authenticated() {
		return nil
	}

	newToken, err := s.client.Token()
	if err != nil {
		return err
	}

	if err := s.persistToken(newToken); err != nil {
		return err
	}

	return nil
}

func (s *Spotify) persistToken(token *oauth2.Token) error {
	s.secrets.Set("token_type", token.TokenType)
	s.secrets.Set("access_token", token.AccessToken)
	s.secrets.Set("expiry", token.Expiry.Format(time.RFC3339))
	s.secrets.Set("refresh_token", token.RefreshToken)

	if err := s.secrets.WriteConfig(); err != nil {
		return err
	}
	return nil
}

func (s *Spotify) authenticateFromSecrets(secrets config.Config) {
	if !secrets.IsSet("token_type") {
		return
	}

	expiryTime, err := time.Parse(time.RFC3339, secrets.GetString("expiry"))
	if err != nil {
		return
	}

	token := &oauth2.Token{
		TokenType:    secrets.GetString("token_type"),
		AccessToken:  secrets.GetString("access_token"),
		Expiry:       expiryTime,
		RefreshToken: secrets.GetString("refresh_token"),
	}

	client := s.authenticator.NewClient(token)
	s.client = &client
}
