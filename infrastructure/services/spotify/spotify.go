//go:generate mockgen -source spotify.go -destination spotify_mock.go -package spotify

package spotify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dietrichm/admirer/domain"
	"github.com/dietrichm/admirer/infrastructure/config"
	"github.com/zmb3/spotify/v2"
	"github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// Authenticator is our interface for a Spotify authenticator.
type Authenticator interface {
	AuthURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, token *oauth2.Token) *http.Client
}

// Client is our interface for a Spotify client.
type Client interface {
	CurrentUser(ctx context.Context) (*spotify.PrivateUser, error)
	Token() (*oauth2.Token, error)
	CurrentUsersTracks(ctx context.Context, opts ...spotify.RequestOption) (*spotify.SavedTrackPage, error)
	Search(ctx context.Context, query string, t spotify.SearchType, opts ...spotify.RequestOption) (*spotify.SearchResult, error)
	AddTracksToLibrary(ctx context.Context, ids ...spotify.ID) error
	GetPlaylistItems(ctx context.Context, playlistID spotify.ID, opts ...spotify.RequestOption) (*spotify.PlaylistItemPage, error)
	CreatePlaylistForUser(ctx context.Context, userID, playlistName, description string, public bool, collaborative bool) (*spotify.FullPlaylist, error)
	ReplacePlaylistTracks(ctx context.Context, playlistID spotify.ID, trackIDs ...spotify.ID) error
	AddTracksToPlaylist(ctx context.Context, playlistID spotify.ID, trackIDs ...spotify.ID) (snapshotID string, err error)
	CurrentUsersTopArtists(ctx context.Context, opts ...spotify.RequestOption) (*spotify.FullArtistPage, error)
	CurrentUsersTopTracks(ctx context.Context, opts ...spotify.RequestOption) (*spotify.FullTrackPage, error)
	GetRecommendations(ctx context.Context, seeds spotify.Seeds, trackAttributes *spotify.TrackAttributes, opts ...spotify.RequestOption) (*spotify.Recommendations, error)
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

	authenticator := spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		spotifyauth.WithRedirectURL(""),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopeUserLibraryRead,
			spotifyauth.ScopeUserLibraryModify,
			spotifyauth.ScopePlaylistModifyPublic,
			spotifyauth.ScopeUserTopRead,
		),
	)

	service := &Spotify{
		authenticator: authenticator,
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
	return s.authenticator.AuthURL("", redirectOption)
}

// CodeParam is the query parameter name used in the authentication callback.
func (s *Spotify) CodeParam() string {
	return "code"
}

// Authenticate takes an authorization code and authenticates the user.
func (s *Spotify) Authenticate(code string, redirectURL string) error {
	ctx := context.Background()
	redirectOption := oauth2.SetAuthURLParam("redirect_uri", redirectURL)
	token, err := s.authenticator.Exchange(ctx, code, redirectOption)
	if err != nil {
		return fmt.Errorf("failed to authenticate on Spotify: %w", err)
	}

	client := s.authenticator.Client(ctx, token)
	s.client = spotify.New(client)

	return nil
}

// GetUsername requests and returns the username of the logged-in user.
func (s *Spotify) GetUsername() (string, error) {
	ctx := context.Background()
	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read Spotify profile data: %w", err)
	}

	return user.DisplayName, nil
}

// GetLovedTracks returns loved tracks from the external service.
func (s *Spotify) GetLovedTracks(limit int, page int) (tracks []domain.Track, err error) {
	ctx := context.Background()
	offset := (page - 1) * limit
	options := []spotify.RequestOption{spotify.Limit(limit), spotify.Offset(offset)}

	result, err := s.client.CurrentUsersTracks(ctx, options...)
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
	ctx := context.Background()
	query := fmt.Sprintf("artist:%q track:%q", track.Artist, track.Name)
	query = strings.ReplaceAll(query, `\"`, "")

	options := []spotify.RequestOption{
		spotify.Limit(1),
	}

	result, err := s.client.Search(ctx, query, spotify.SearchTypeTrack, options...)
	if err != nil {
		return fmt.Errorf("failed to search track on Spotify: %w", err)
	}

	if len(result.Tracks.Tracks) == 0 {
		return nil
	}

	trackID := result.Tracks.Tracks[0].ID
	if err := s.client.AddTracksToLibrary(ctx, trackID); err != nil {
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
	ctx := context.Background()
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

	client := s.authenticator.Client(ctx, token)
	s.client = spotify.New(client)
}

func (s *Spotify) GetUserId() (string, error) {
	ctx := context.Background()
	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read Spotify profile data: %w", err)
	}

	return user.ID, nil
}

func (s *Spotify) DumpDiscoverWeeklyTracksToNewPlaylist(writer io.Writer) error {
	ctx := context.Background()
	userId, err := s.GetUserId()
	if err != nil {
		return fmt.Errorf("failed to read Spotify profile data: %w", err)
	}
	fmt.Fprintln(writer, "UserID: ", userId)

	year, week := time.Now().UTC().ISOWeek()
	playlistName := fmt.Sprintf("Discover Weekly #%d %d", week, year)
	playlistDescription := fmt.Sprintf("Backup of the Discover Weekly playlist for %d week in %d.", week, year)
	playlist, err := s.client.CreatePlaylistForUser(ctx, userId, playlistName, playlistDescription, true, false)
	if err != nil {
		return fmt.Errorf("failed to create Spotify playlist: %w", err)
	}

	offset := 0
	limit := 50
	tracksOpt := []spotify.RequestOption{spotify.Limit(limit), spotify.Offset(offset), spotify.AdditionalTypes(spotify.TrackAdditionalType)}

	for page := 1; ; page++ {
		searchLimit := 1
		searchOpt := []spotify.RequestOption{spotify.Limit(searchLimit)}
		sp, err := s.client.Search(ctx, "Discover Weekly", spotify.SearchTypePlaylist, searchOpt...)
		if err != nil {
			return fmt.Errorf("failed to search Discover Weekly playlist: %w", err)
		}
		if len(sp.Playlists.Playlists) == 0 {
			return fmt.Errorf("playlist Discover Weekly not found")
		}
		playlistID := sp.Playlists.Playlists[0].ID
		fmt.Fprintln(writer, "PlaylistID: ", playlistID)
		tracks, err := s.client.GetPlaylistItems(ctx, playlistID, tracksOpt...)
		if err != nil {
			return fmt.Errorf("failed to read Spotify playlist data: %w", err)
		}
		_, err = fmt.Fprintf(writer, "Playlist has %d total tracks\n", tracks.Total)
		if err != nil {
			return err
		}

		trackCount := len(tracks.Items)
		fmt.Fprintf(writer, "Page %d has %d tracks\n", page, trackCount)
		if trackCount == 0 {
			break
		}

		var trackIDs []spotify.ID
		for _, track := range tracks.Items {
			trackIDs = append(trackIDs, track.Track.Track.ID)
			fmt.Fprintln(writer, track.Track.Track.String())
		}

		if page == 1 {
			if err := s.client.ReplacePlaylistTracks(ctx, playlist.ID, trackIDs...); err != nil {
				return fmt.Errorf("failed to replace tracks in playlist: %w", err)
			}
		} else {
			if _, err := s.client.AddTracksToPlaylist(ctx, playlist.ID, trackIDs...); err != nil {
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

func (s *Spotify) DiscoverDailyPlaylist(writer io.Writer) error {
	ctx := context.Background()
	userId, err := s.GetUserId()
	if err != nil {
		return fmt.Errorf("failed to read Spotify profile data: %w", err)
	}
	timeRanges := []spotify.Range{spotify.LongTermRange, spotify.MediumTermRange, spotify.ShortTermRange}
	timeRange := timeRanges[rand.Intn(len(timeRanges))]
	topRequestOptions := []spotify.RequestOption{spotify.Timerange(timeRange), spotify.Limit(50)}
	topArtists, err := s.client.CurrentUsersTopArtists(ctx, topRequestOptions...)
	if err != nil {
		return fmt.Errorf("failed to get current user top artist from Spotify: %w", err)
	}
	var topArtistIDs []spotify.ID
	for _, artist := range topArtists.Artists {
		topArtistIDs = append(topArtistIDs, artist.ID)
	}
	//randomize top artist IDs
	rand.Shuffle(len(topArtistIDs), func(i, j int) { topArtistIDs[i], topArtistIDs[j] = topArtistIDs[j], topArtistIDs[i] })

	topTracks, err := s.client.CurrentUsersTopTracks(ctx, topRequestOptions...)
	if err != nil {
		return fmt.Errorf("failed to get current user top tracks from Spotify: %w", err)
	}
	var topTrackIDs []spotify.ID
	for _, track := range topTracks.Tracks {
		topTrackIDs = append(topTrackIDs, track.ID)
	}
	//shuffle top track IDs
	rand.Shuffle(len(topArtistIDs), func(i, j int) { topTrackIDs[i], topTrackIDs[j] = topTrackIDs[j], topTrackIDs[i] })

	opts := []spotify.RequestOption{spotify.Limit(100)}
	recommendedTracks, err := s.client.GetRecommendations(ctx,
		spotify.Seeds{
			//Max seeds count is 5!
			Artists: topArtistIDs[:2], // first 2 artists
			Tracks:  topTrackIDs[:3],  // first 3 tracks
		},
		//TODO: Add more options like Tempo, Acousticness, Danceability, Energy, Instrumentalness, Liveness, Valence
		spotify.NewTrackAttributes(),
		opts...)
	if err != nil {
		return fmt.Errorf("failed to get recommendations from Spotify: %w", err)
	}

	var trackIDs []spotify.ID
	for _, track := range recommendedTracks.Tracks {
		trackIDs = append(trackIDs, track.ID)
		fmt.Fprintln(writer, track.String())
	}

	date := time.Now().UTC().Format("02-01-2006")
	playlistName := fmt.Sprintf("Discover Daily %s", date)
	playlistDescription := fmt.Sprintf("Discover Daily playlist for %s from recomendations with options: %s", date, timeRange)
	playlist, err := s.client.CreatePlaylistForUser(ctx, userId, playlistName, playlistDescription, true, false)
	if err != nil {
		return fmt.Errorf("failed to create Spotify playlist: %w", err)
	}

	if _, err := s.client.AddTracksToPlaylist(ctx, playlist.ID, trackIDs...); err != nil {
		return fmt.Errorf("failed to add tracks to playlist: %w", err)
	}

	return nil
}
