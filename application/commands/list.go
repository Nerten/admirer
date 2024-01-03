package commands

import (
	"fmt"
	"io"

	"github.com/dietrichm/admirer/domain"
	"github.com/dietrichm/admirer/infrastructure/services"
	"github.com/spf13/cobra"
)

func init() {
	listCommand.Flags().IntVarP(&limit, "limit", "l", 10, "Limit the number of tracks to be displayed. Specify 0 to output all tracks without limitations. In this case, the default limit for a group of tracks will be 50 (note: important for accurate page counting)")
	listCommand.Flags().IntVarP(&page, "page", "p", 1, "Page number to start displaying from")
	rootCommand.AddCommand(listCommand)
}

var listCommand = &cobra.Command{
	Use:   "list <service>",
	Short: "List loved tracks on specified service",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		return list(services.AvailableServices, limit, page, command.OutOrStdout(), args)
	},
}

func list(serviceLoader domain.ServiceLoader, limit int, page int, writer io.Writer, args []string) error {
	serviceName := args[0]

	continuously := false
	if limit == 0 {
		limit = 50
		continuously = true
	}

	service, err := serviceLoader.ForName(serviceName)
	if err != nil {
		return err
	}

	defer service.Close()

	if !service.Authenticated() {
		return fmt.Errorf("not logged in on %s", service.Name())
	}

	for ; ; page++ {
		tracks, err := service.GetLovedTracks(limit, page)
		if err != nil {
			return err
		}

		for _, track := range tracks {
			fmt.Fprintln(writer, track.String())
		}
		if !continuously || len(tracks) < limit {
			break
		}
	}

	return nil
}
