package commands

import (
	"fmt"

	"github.com/dietrichm/admirer/services"
	"github.com/spf13/cobra"
)

func init() {
	rootCommand.AddCommand(loginCommand)
}

var loginCommand = &cobra.Command{
	Use:   "login <service> [oauth-code]",
	Short: "Log in on external service",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(command *cobra.Command, args []string) error {
		serviceLoader := new(services.DefaultServiceLoader)
		service, err := serviceLoader.ForName(args[0])
		if err != nil {
			return err
		}

		writer := command.OutOrStdout()

		if len(args) < 2 {
			fmt.Fprintln(writer, service.Name(), "authentication URL:", service.CreateAuthURL())
			return nil
		}

		if err := service.Authenticate(args[1]); err != nil {
			return err
		}

		username, err := service.GetUsername()
		if err != nil {
			return err
		}

		fmt.Fprintln(writer, "Logged in on", service.Name(), "as", username)
		return nil
	},
}
