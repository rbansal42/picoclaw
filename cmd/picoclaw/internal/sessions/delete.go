package sessions

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newDeleteCommand(sessionsDir func() string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <id>",
		Short:   "Delete a session",
		Args:    cobra.ExactArgs(1),
		Example: "picoclaw sessions delete my-session",
		RunE: func(_ *cobra.Command, args []string) error {
			return sessionsDeleteCmd(sessionsDir(), args[0])
		},
	}

	return cmd
}

func sessionsDeleteCmd(sessionsDir, id string) error {
	filePath := findSessionFile(sessionsDir, id)
	if filePath == "" {
		return fmt.Errorf("session '%s' not found", id)
	}

	fmt.Printf("Delete session '%s'? (y/n): ", id)
	if !confirmPrompt() {
		fmt.Println("Cancelled")
		return nil
	}

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}

	fmt.Printf("Deleted session %s\n", id)
	return nil
}
