package commands

import (
	"github.com/spf13/cobra"
)

// NewCommand returns a new instance of an manager command
func NewCommand() *cobra.Command {

	var command = &cobra.Command{
		Use:   "manager",
		Short: "manager cli provides a way to connect to manager server",
		Run: func(c *cobra.Command, args []string) {
			c.HelpFunc()(c, args)
		},
	}

	command.AddCommand(NewClusterCommand())

	return command
}
