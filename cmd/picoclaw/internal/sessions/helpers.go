package sessions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// sessionData is a minimal struct for reading session JSON files.
// We only need the fields required for display — no dependency on pkg/session.
type sessionData struct {
	Key      string          `json:"key"`
	Messages json.RawMessage `json:"messages"`
	Summary  string          `json:"summary,omitempty"`
	Created  time.Time       `json:"created"`
	Updated  time.Time       `json:"updated"`
}

// sessionMessage is a minimal struct for reading individual messages
type sessionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type sessionEntry struct {
	id       string
	messages int
	modTime  time.Time
	size     int64
	corrupt  bool
}

// listSessionEntries reads the sessions directory and returns parsed entries
func listSessionEntries(sessionsDir string) ([]sessionEntry, error) {
	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no sessions found (sessions directory does not exist)")
	}

	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("error reading sessions directory: %w", err)
	}

	var entries []sessionEntry
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(sessionsDir, f.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			continue
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var sess sessionData
		if err := json.Unmarshal(data, &sess); err != nil {
			// Corrupt file — derive ID from filename
			id := strings.TrimSuffix(f.Name(), ".json")
			entries = append(entries, sessionEntry{
				id:      id,
				modTime: info.ModTime(),
				size:    info.Size(),
				corrupt: true,
			})
			continue
		}

		// Use the key from the JSON if present, otherwise derive from filename
		id := sess.Key
		if id == "" {
			id = strings.TrimSuffix(f.Name(), ".json")
		}

		var msgs []sessionMessage
		_ = json.Unmarshal(sess.Messages, &msgs)

		entries = append(entries, sessionEntry{
			id:       id,
			messages: len(msgs),
			modTime:  info.ModTime(),
			size:     info.Size(),
		})
	}

	return entries, nil
}

// findSessionFile locates the session file for a given ID.
// It first tries matching by the key inside the JSON, then falls back
// to matching by filename (with .json extension).
func findSessionFile(sessionsDir string, id string) string {
	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		return ""
	}

	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(sessionsDir, f.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var sess sessionData
		if err := json.Unmarshal(data, &sess); err != nil {
			// Corrupt file — match by filename
			name := strings.TrimSuffix(f.Name(), ".json")
			if name == id {
				return filePath
			}
			continue
		}

		if sess.Key == id {
			return filePath
		}
	}

	// Fallback: try direct filename match (id + .json or sanitized id + .json)
	direct := filepath.Join(sessionsDir, id+".json")
	if _, err := os.Stat(direct); err == nil {
		return direct
	}
	sanitized := filepath.Join(sessionsDir, strings.ReplaceAll(id, ":", "_")+".json")
	if _, err := os.Stat(sanitized); err == nil {
		return sanitized
	}

	return ""
}

func confirmPrompt() bool {
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func formatSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
	)
	switch {
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
