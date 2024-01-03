package commands

import (
	"fmt"
	"io"

	"github.com/dietrichm/admirer/domain"
	"github.com/dietrichm/admirer/infrastructure/services"
	"github.com/spf13/cobra"
)

func init() {
	syncCommand.Flags().IntVarP(&limit, "limit", "l", 10, "Limit number of tracks for syncing. Specify 0 to sync all tracks without limitations. In this case, the default limit for a group of tracks will be 50 (note: important for accurate page counting)")
	syncCommand.Flags().IntVarP(&page, "page", "p", 1, "Page number to start syncing from")
	rootCommand.AddCommand(syncCommand)
}

var syncCommand = &cobra.Command{
	Use:   "sync <source-service> <target-service>",
	Short: "Sync recently loved tracks from one service to another",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(command *cobra.Command, args []string) error {
		return sync(services.AvailableServices, limit, page, command.OutOrStdout(), args)
	},
}

func sync(serviceLoader domain.ServiceLoader, limit int, page int, writer io.Writer, args []string) error {
	sourceServiceName := args[0]
	targetServiceName := args[1]

	continuously := false
	if limit == 0 {
		limit = 50
		continuously = true
	}

	sourceService, err := serviceLoader.ForName(sourceServiceName)
	if err != nil {
		return err
	}

	targetService, err := serviceLoader.ForName(targetServiceName)
	if err != nil {
		return err
	}

	defer sourceService.Close()
	defer targetService.Close()

	if !sourceService.Authenticated() {
		return fmt.Errorf("not logged in on %s", sourceService.Name())
	}

	if !targetService.Authenticated() {
		return fmt.Errorf("not logged in on %s", targetService.Name())
	}

	for ; ; page++ {
		tracks, err := sourceService.GetLovedTracks(limit, page)
		if err != nil {
			return err
		}

		for _, track := range tracks {
			if err := targetService.LoveTrack(track); err != nil {
				return err
			}

			fmt.Fprintln(writer, "Synced:", track.String())
		}
		if !continuously || len(tracks) < limit {
			break
		}
	}

	return nil
}
