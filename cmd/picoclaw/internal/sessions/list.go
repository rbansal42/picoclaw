package sessions

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newListCommand(sessionsDir func() string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all sessions",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return sessionsListCmd(sessionsDir())
		},
	}

	return cmd
}

func sessionsListCmd(sessionsDir string) error {
	entries, err := listSessionEntries(sessionsDir)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No sessions found")
		return nil
	}

	// Sort by modification time, most recent first
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].modTime.After(entries[j].modTime)
	})

	fmt.Println("Sessions:")
	fmt.Printf("  %-30s %8s  %s\n", "ID", "Messages", "Last Modified")
	for _, e := range entries {
		msgStr := fmt.Sprintf("%d", e.messages)
		if e.corrupt {
			msgStr = "(corrupt)"
		}
		fmt.Printf("  %-30s %8s  %s\n", e.id, msgStr, e.modTime.Format("2006-01-02 15:04"))
	}

	fmt.Printf("\n%d session(s) found\n", len(entries))
	return nil
}
