package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (cli *CLI) buildVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "version",
		Short:                 fmt.Sprintf("Get version of %s CLI", cli.Name),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			showSuccess(cli.version)
		},
	}

	return cmd
}
