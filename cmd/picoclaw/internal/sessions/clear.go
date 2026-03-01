package sessions

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newClearCommand(sessionsDir func() string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Delete all sessions",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return sessionsClearCmd(sessionsDir())
		},
	}

	return cmd
}

func sessionsClearCmd(sessionsDir string) error {
	entries, err := listSessionEntries(sessionsDir)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No sessions found")
		return nil
	}

	fmt.Printf("Delete all %d sessions? (y/n): ", len(entries))
	if !confirmPrompt() {
		fmt.Println("Cancelled")
		return nil
	}

	deleted := 0
	for _, e := range entries {
		filePath := findSessionFile(sessionsDir, e.id)
		if filePath == "" {
			continue
		}
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Error deleting session '%s': %v\n", e.id, err)
			continue
		}
		deleted++
	}

	fmt.Printf("Cleared %d session(s).\n", deleted)
	return nil
}
