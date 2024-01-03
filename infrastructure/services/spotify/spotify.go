//go:generate mockgen -source spotify.go -destination spotify_mock.go -package spotify

package spotify

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dietrichm/admirer/domain"
	"github.com/dietrichm/admirer/infrastructure/config"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

// Authenticator is our interface for a Spotify authenticator.
type Authenticator interface {
	SetAuthInfo(clientID, secretKey string)
	AuthURLWithOpts(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(string, ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	NewClient(token *oauth2.Token) spotify.Client
}

// Client is our interface for a Spotify client.
type Client interface {
	CurrentUser() (*spotify.PrivateUser, error)
	Token() (*oauth2.Token, error)
	CurrentUsersTracksOpt(opt *spotify.Options) (*spotify.SavedTrackPage, error)
	SearchOpt(query string, t spotify.SearchType, opt *spotify.Options) (*spotify.SearchResult, error)
	AddTracksToLibrary(ids ...spotify.ID) error
	GetPlaylistTracksOpt(id spotify.ID, opt *spotify.Options, fields string) (*spotify.PlaylistTrackPage, error)
	CreatePlaylistForUser(userID, playlistName, description string, public bool) (*spotify.FullPlaylist, error)
	ReplacePlaylistTracks(playlistID spotify.ID, trackIDs ...spotify.ID) error
	AddTracksToPlaylist(playlistID spotify.ID, trackIDs ...spotify.ID) (snapshotID string, err error)
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

	authenticator := spotify.NewAuthenticator("", spotify.ScopeUserReadPrivate, spotify.ScopeUserLibraryRead, spotify.ScopeUserLibraryModify)
	authenticator.SetAuthInfo(clientID, clientSecret)

	service := &Spotify{
		authenticator: &authenticator,
		secrets:       secrets,
	}
	service.authenticateFromSecrets(secrets)

	return service, nil
}

// Name returns the human-readable service name.
func (s *Spotify) Name() string {
	return "Spotify"
}

// Authenticated returns whether the service is logged in.
func (s *Spotify) Authenticated() bool {
	return s.client != nil
}

// CreateAuthURL returns an authorization URL to authorize the integration.
func (s *Spotify) CreateAuthURL(redirectURL string) string {
	redirectOption := oauth2.SetAuthURLParam("redirect_uri", redirectURL)
	return s.authenticator.AuthURLWithOpts("", redirectOption)
}

// CodeParam is the query parameter name used in the authentication callback.
func (s *Spotify) CodeParam() string {
	return "code"
}

// Authenticate takes an authorization code and authenticates the user.
func (s *Spotify) Authenticate(code string, redirectURL string) error {
	redirectOption := oauth2.SetAuthURLParam("redirect_uri", redirectURL)
	token, err := s.authenticator.Exchange(code, redirectOption)
	if err != nil {
		return fmt.Errorf("failed to authenticate on Spotify: %w", err)
	}

	client := s.authenticator.NewClient(token)
	s.client = &client

	return nil
}

// GetUsername requests and returns the username of the logged-in user.
func (s *Spotify) GetUsername() (string, error) {
	user, err := s.client.CurrentUser()
	if err != nil {
		return "", fmt.Errorf("failed to read Spotify profile data: %w", err)
	}

	return user.DisplayName, nil
}

// GetLovedTracks returns loved tracks from the external service.
func (s *Spotify) GetLovedTracks(limit int, page int) (tracks []domain.Track, err error) {
	offset := (page - 1) * limit
	options := &spotify.Options{
		Limit:  &limit,
		Offset: &offset,
	}

	result, err := s.client.CurrentUsersTracksOpt(options)
	if err != nil {
		return tracks, fmt.Errorf("failed to read Spotify loved tracks: %w", err)
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

// LoveTrack marks a track as loved on the external service.
func (s *Spotify) LoveTrack(track domain.Track) error {
	query := fmt.Sprintf("artist:%q track:%q", track.Artist, track.Name)
	query = strings.ReplaceAll(query, `\"`, "")

	limit := 1
	options := &spotify.Options{
		Limit: &limit,
	}

	result, err := s.client.SearchOpt(query, spotify.SearchTypeTrack, options)
	if err != nil {
		return fmt.Errorf("failed to search track on Spotify: %w", err)
	}

	if len(result.Tracks.Tracks) == 0 {
		return nil
	}

	trackID := result.Tracks.Tracks[0].ID
	if err := s.client.AddTracksToLibrary(trackID); err != nil {
		return fmt.Errorf("failed to mark track as loved on Spotify: %w", err)
	}

	return nil
}

// Close persists any state before quitting the application.
func (s *Spotify) Close() error {
	if !s.Authenticated() {
		return nil
	}

	newToken, err := s.client.Token()
	if err != nil {
		return fmt.Errorf("failed to save Spotify secrets: %w", err)
	}

	if err := s.persistToken(newToken); err != nil {
		return fmt.Errorf("failed to save Spotify secrets: %w", err)
	}

	return nil
}

func (s *Spotify) persistToken(token *oauth2.Token) error {
	s.secrets.Set("token_type", token.TokenType)
	s.secrets.Set("access_token", token.AccessToken)
	s.secrets.Set("expiry", token.Expiry.Format(time.RFC3339))
	s.secrets.Set("refresh_token", token.RefreshToken)

	if err := s.secrets.Save(); err != nil {
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

func (s *Spotify) GetUserId() (string, error) {
	user, err := s.client.CurrentUser()
	if err != nil {
		return "", fmt.Errorf("failed to read Spotify profile data: %w", err)
	}

	return user.ID, nil
}

func (s *Spotify) DumpDiscoverWeeklyTracksToNewPlaylist(writer io.Writer) error {
	userId, err := s.GetUserId()
	fmt.Fprintln(writer, "UserID: ", userId)
	if err != nil {
		return fmt.Errorf("failed to read Spotify profile data: %w", err)
	}

	year, week := time.Now().UTC().ISOWeek()
	playlistName := fmt.Sprintf("Discover Weekly #%d %d", week, year)
	playlistDescription := fmt.Sprintf("Backup of the Discover Weekly playlist for %d week in %d.", week, year)
	playlist, err := s.client.CreatePlaylistForUser(userId, playlistName, playlistDescription, true)
	if err != nil {
		return fmt.Errorf("failed to create Spotify playlist: %w", err)
	}

	offset := 0
	limit := 50
	tracksOpt := &spotify.Options{Limit: &limit, Offset: &offset}

	for page := 1; ; page++ {
		searchLimit := 1
		searchOpt := &spotify.Options{Limit: &searchLimit}
		sp, err := s.client.SearchOpt("Discover Weekly", spotify.SearchTypePlaylist, searchOpt)
		if err != nil {
			return fmt.Errorf("failed to search Discover Weekly playlist: %w", err)
		}
		if len(sp.Playlists.Playlists) == 0 {
			return fmt.Errorf("playlist Discover Weekly not found")
		}
		playlistID := sp.Playlists.Playlists[0].ID
		fmt.Fprintln(writer, "PlaylistID: ", playlistID)
		tracks, err := s.client.GetPlaylistTracksOpt(playlistID, tracksOpt, "")
		if err != nil {
			return fmt.Errorf("failed to read Spotify playlist data: %w", err)
		}
		_, err = fmt.Fprintf(writer, "Playlist has %d total tracks\n", tracks.Total)
		if err != nil {
			return err
		}

		trackCount := len(tracks.Tracks)
		fmt.Fprintf(writer, "Page %d has %d tracks\n", page, trackCount)
		if trackCount == 0 {
			break
		}

		var trackIDs []spotify.ID

		for _, track := range tracks.Tracks {
			trackIDs = append(trackIDs, track.Track.SimpleTrack.ID)
			fmt.Fprintln(writer, track.Track.String())
		}

		if page == 1 {
			if err := s.client.ReplacePlaylistTracks(playlist.ID, trackIDs...); err != nil {
				return fmt.Errorf("failed to replace tracks in playlist: %w", err)
			}
		} else {
			if _, err := s.client.AddTracksToPlaylist(playlist.ID, trackIDs...); err != nil {
				return fmt.Errorf("failed to add tracks to playlist: %w", err)
			}
		}

		if trackCount < limit {
			break
		}
		offset += trackCount
	}
	return nil
}
