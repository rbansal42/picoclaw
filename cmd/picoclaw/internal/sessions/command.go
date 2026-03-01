package sessions

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/sipeed/picoclaw/cmd/picoclaw/internal"
)

func NewSessionsCommand() *cobra.Command {
	var sessionsDir string

	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage sessions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := internal.LoadConfig()
			if err != nil {
				return fmt.Errorf("error loading config: %w", err)
			}
			sessionsDir = filepath.Join(cfg.WorkspacePath(), "sessions")
			return nil
		},
	}

	cmd.AddCommand(
		newListCommand(func() string { return sessionsDir }),
		newShowCommand(func() string { return sessionsDir }),
		newDeleteCommand(func() string { return sessionsDir }),
		newClearCommand(func() string { return sessionsDir }),
	)

	return cmd
}
