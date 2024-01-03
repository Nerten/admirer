package commands

import (
	"fmt"
	"github.com/dietrichm/admirer/infrastructure/config"
	"github.com/dietrichm/admirer/infrastructure/services/spotify"
	"github.com/spf13/cobra"
	"io"
	"regexp"
	"strings"
)

func init() {
	rootCommand.AddCommand(dumpCommand)
}

var dumpCommand = &cobra.Command{
	Use:   "dump",
	Short: "Back up your Spotify Discover Weekly playlist for the current week",
	RunE: func(command *cobra.Command, args []string) error {
		return dump(config.SecretsLoader, command.OutOrStdout())
	},
}

func dump(secretsLoader config.Loader, writer io.Writer) error {
	serviceName := "spotify"
	replaceRegex := regexp.MustCompile("[^a-zA-Z0-9]")
	internalServiceName := strings.ToLower(replaceRegex.ReplaceAllString(serviceName, ""))

	secrets, err := secretsLoader.Load("secrets-" + internalServiceName)
	if err != nil {
		return err
	}

	service, err := spotify.NewSpotify(secrets)
	if err != nil {
		return err
	}

	defer service.Close()

	if !service.Authenticated() {
		return fmt.Errorf("not logged in on %s", service.Name())
	}

	return service.DumpDiscoverWeeklyTracksToNewPlaylist(writer)
}
