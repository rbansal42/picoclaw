package update

import (
	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update picoclaw to the latest version",
		Long:  "Check GitHub for the latest release and offer to update in-place.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runUpdate(checkOnly)
		},
	}

	cmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Check only, don't prompt to update")

	return cmd
}
