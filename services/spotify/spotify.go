package spotify

import (
	"fmt"
	"os"

	"github.com/zmb3/spotify"
)

// Login runs the CLI procedure for logging in on Spotify.
func Login(oauthCode string) {
	authenticator := createSpotifyAuthenticator()

	if len(oauthCode) == 0 {
		fmt.Println("Spotify authentication URL: " + createSpotifyAuthURL(&authenticator))
		return
	}

	client := callbackSpotify(&authenticator, oauthCode)

	user, err := client.CurrentUser()
	if err != nil {
		panic("Failed to read Spotify profile data.")
	}

	fmt.Println("Logged in on Spotify as " + user.DisplayName)
}

func createSpotifyAuthenticator() spotify.Authenticator {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	if len(clientID) == 0 || len(clientSecret) == 0 {
		panic("Please set SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables.")
	}

	// Not an actual web server (yet).
	redirectURL := "https://admirer.test"
	authenticator := spotify.NewAuthenticator(redirectURL, spotify.ScopeUserReadPrivate)
	authenticator.SetAuthInfo(clientID, clientSecret)

	return authenticator
}

func createSpotifyAuthURL(authenticator *spotify.Authenticator) string {
	return authenticator.AuthURL("")
}

func callbackSpotify(authenticator *spotify.Authenticator, code string) spotify.Client {
	token, err := authenticator.Exchange(code)
	if err != nil {
		panic("Failed to parse Spotify token.")
	}

	return authenticator.NewClient(token)
}
