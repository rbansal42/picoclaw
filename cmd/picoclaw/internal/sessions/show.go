package sessions

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newShowCommand(sessionsDir func() string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show <id>",
		Short:   "Show session details",
		Args:    cobra.ExactArgs(1),
		Example: "picoclaw sessions show my-session",
		RunE: func(_ *cobra.Command, args []string) error {
			return sessionsShowCmd(sessionsDir(), args[0])
		},
	}

	return cmd
}

func sessionsShowCmd(sessionsDir, id string) error {
	filePath := findSessionFile(sessionsDir, id)
	if filePath == "" {
		return fmt.Errorf("session '%s' not found", id)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("reading session: %w", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading session: %w", err)
	}

	var sess sessionData
	if err := json.Unmarshal(data, &sess); err != nil {
		fmt.Printf("Session: %s\n", id)
		fmt.Printf("Size: %s\n", formatSize(info.Size()))
		fmt.Printf("Last Modified: %s\n", info.ModTime().Format("2006-01-02 15:04"))
		fmt.Println("Status: corrupt (invalid JSON)")
		return nil
	}

	var msgs []sessionMessage
	_ = json.Unmarshal(sess.Messages, &msgs)

	fmt.Printf("Session: %s\n", sess.Key)
	fmt.Printf("Messages: %d\n", len(msgs))
	fmt.Printf("Last Modified: %s\n", info.ModTime().Format("2006-01-02 15:04"))
	fmt.Printf("Size: %s\n", formatSize(info.Size()))

	if len(msgs) > 0 {
		fmt.Println()
		start := len(msgs) - 3
		if start < 0 {
			start = 0
		}
		fmt.Println("Last messages:")
		for _, m := range msgs[start:] {
			content := strings.TrimSpace(m.Content)
			content = strings.ReplaceAll(content, "\n", " ")
			if len(content) > 80 {
				content = content[:77] + "..."
			}
			fmt.Printf("  [%s] %s\n", m.Role, content)
		}
	}

	return nil
}
