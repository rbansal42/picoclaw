package doctor

import (
	"github.com/spf13/cobra"
)

func NewDoctorCommand() *cobra.Command {
	var fix bool

	cmd := &cobra.Command{
		Use:     "doctor",
		Aliases: []string{"d"},
		Short:   "Diagnose common problems",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(fix)
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "Attempt to automatically fix problems")

	return cmd
}
