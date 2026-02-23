package tools

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

// NewCLIPermissionFunc creates a PermissionFunc that prompts the user on a terminal.
//
// The optional onBefore and onAfter callbacks are called immediately before and
// after the interactive prompt is shown. Use them to stop/restart a spinner or
// temporarily exit readline raw mode so the prompt is visible and input works.
func NewCLIPermissionFunc(reader io.Reader, writer io.Writer, onBefore, onAfter func()) PermissionFunc {
	scanner := bufio.NewScanner(reader)
	return func(ctx context.Context, path string) (bool, error) {
		if onBefore != nil {
			onBefore()
		}
		fmt.Fprintf(writer, "\n⚠ Agent wants to access: %s\nAllow access to this directory? [y/N]: ", path)
		var answer string
		if scanner.Scan() {
			answer = strings.TrimSpace(strings.ToLower(scanner.Text()))
		} else if err := scanner.Err(); err != nil {
			if onAfter != nil {
				onAfter()
			}
			return false, err
		}
		if onAfter != nil {
			onAfter()
		}
		return answer == "y" || answer == "yes", nil
	}
}
