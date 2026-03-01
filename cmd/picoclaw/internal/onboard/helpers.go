package onboard

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/cmd/picoclaw/internal"
	"github.com/sipeed/picoclaw/pkg/auth"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/providers"
)

func onboard() error {
	configPath := internal.GetConfigPath()

	// Check for existing config
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config already exists at %s\n", configPath)
		fmt.Print("Overwrite? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Show welcome and provider menu
	fmt.Printf("\n%s Welcome to PicoClaw! Let's set up your AI provider.\n\n", internal.Logo)
	fmt.Println("Which provider would you like to use?")
	fmt.Println()
	fmt.Println("  1. Ollama (local, free — requires Ollama installed)")
	fmt.Println("  2. Anthropic Claude (API key or OAuth subscription)")
	fmt.Println("  3. OpenAI GPT (API key required)")
	fmt.Println("  4. DeepSeek (API key required)")
	fmt.Println("  5. Google Gemini (OAuth — free tier available)")
	fmt.Println("  6. OpenRouter (API key — access to many models)")
	fmt.Println()

	choice := promptChoice("Choice [1-6]: ", 1, 6)

	cfg := config.DefaultConfig()
	var verified bool

	switch choice {
	case 1:
		verified = setupOllama(cfg)
	case 2:
		verified = setupAnthropic(cfg)
	case 3:
		verified = setupOpenAI(cfg)
	case 4:
		verified = setupDeepSeek(cfg)
	case 5:
		verified = setupGoogleGemini(cfg)
	case 6:
		verified = setupOpenRouter(cfg)
	}

	// Save config
	if err := config.SaveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// Copy workspace templates
	workspace := cfg.WorkspacePath()
	createWorkspaceTemplates(workspace)

	// Print results
	fmt.Println()
	fmt.Printf("  [ok] Config written to %s\n", configPath)
	fmt.Printf("  [ok] Workspace initialized at %s\n", workspace)

	if verified {
		fmt.Printf("  [ok] Provider verified\n")
	}

	fmt.Printf("\n%s PicoClaw is ready!\n\n", internal.Logo)
	fmt.Println("Try it: picoclaw agent -m \"Hello!\"")
	fmt.Println()
	fmt.Println("If something isn't working:")
	fmt.Println("  picoclaw doctor          Diagnose problems")
	fmt.Println("  picoclaw doctor --fix    Auto-fix what it can")

	return nil
}

// promptChoice asks the user for a number between min and max, re-prompting on invalid input.
func promptChoice(prompt string, min, max int) int {
	for {
		fmt.Print(prompt)
		var input string
		fmt.Scanln(&input)
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		// Parse as number
		var n int
		if _, err := fmt.Sscanf(input, "%d", &n); err != nil || n < min || n > max {
			fmt.Printf("Please enter a number between %d and %d.\n", min, max)
			continue
		}
		return n
	}
}

// promptString asks for a string input. Returns empty string if user just hits enter.
func promptString(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

// setModelAPIKey sets the API key on the model entry matching the given modelName.
func setModelAPIKey(cfg *config.Config, modelName, apiKey string) {
	for i := range cfg.ModelList {
		if cfg.ModelList[i].ModelName == modelName {
			cfg.ModelList[i].APIKey = apiKey
			return
		}
	}
}

// setupOllama configures Ollama as the provider and checks if it's running.
func setupOllama(cfg *config.Config) bool {
	cfg.Agents.Defaults.Model = "llama3"

	fmt.Println("\nChecking if Ollama is running...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:11434/v1/models", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("  [!] Ollama doesn't seem to be running at localhost:11434")
		fmt.Println()
		fmt.Println("  To install Ollama:")
		fmt.Println("    https://ollama.com/download")
		fmt.Println()
		fmt.Println("  Then start it:")
		fmt.Println("    ollama serve")
		fmt.Println()
		fmt.Println("  And pull a model:")
		fmt.Println("    ollama pull llama3")
		fmt.Println()
		fmt.Println("  Config will be saved — just start Ollama when ready.")
		return false
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("  [ok] Ollama is running at localhost:11434")
		return true
	}

	fmt.Printf("  [!] Ollama returned HTTP %d — it may not be fully ready\n", resp.StatusCode)
	return false
}

// setupAnthropic configures Anthropic Claude, offering OAuth or API key options.
func setupAnthropic(cfg *config.Config) bool {
	cfg.Agents.Defaults.Model = "claude-sonnet-4.6"

	fmt.Println("\nHow would you like to authenticate with Anthropic?")
	fmt.Println("  1. Claude Max/Pro subscription (OAuth — free inference)")
	fmt.Println("  2. API key (pay-per-use)")
	fmt.Println()

	choice := promptChoice("Choice [1-2]: ", 1, 2)

	switch choice {
	case 1:
		return setupAnthropicOAuth(cfg)
	case 2:
		return setupAnthropicAPIKey(cfg)
	}
	return false
}

// setupAnthropicOAuth runs the Claude Max/Pro OAuth flow.
func setupAnthropicOAuth(cfg *config.Config) bool {
	fmt.Println("\nOpening browser for Anthropic login...")

	cred, err := auth.LoginAnthropicOAuth(auth.AnthropicOAuthMax)
	if err != nil {
		fmt.Printf("  [!] OAuth login failed: %v\n", err)
		fmt.Println("      You can try again later with: picoclaw auth login --provider anthropic")
		return false
	}

	if err := auth.SetCredential("anthropic", cred); err != nil {
		fmt.Printf("  [!] Failed to save credentials: %v\n", err)
		return false
	}

	// Update config to use OAuth for anthropic models
	for i := range cfg.ModelList {
		if isAnthropicModel(cfg.ModelList[i].Model) {
			cfg.ModelList[i].AuthMethod = "oauth"
			break
		}
	}

	fmt.Print("  [ok] Anthropic login successful")
	if cred.Email != "" {
		fmt.Printf(" (%s)", cred.Email)
	}
	if cred.SubscriptionType != "" {
		fmt.Printf(" [%s]", cred.SubscriptionType)
	}
	fmt.Println()
	return true
}

// setupAnthropicAPIKey collects and validates an Anthropic API key.
func setupAnthropicAPIKey(cfg *config.Config) bool {
	fmt.Println()
	fmt.Println("  Get your API key at: https://console.anthropic.com/settings/keys")
	fmt.Println()

	apiKey := promptString("Enter your Anthropic API key: ")
	if apiKey == "" {
		fmt.Println("  [!] No API key provided — you can add it later in config.json")
		return false
	}

	setModelAPIKey(cfg, "claude-sonnet-4.6", apiKey)

	return validateAPIKey("Anthropic", apiKey, func(key string) *http.Request {
		req, _ := http.NewRequest("GET", "https://api.anthropic.com/v1/models", nil)
		req.Header.Set("x-api-key", key)
		req.Header.Set("anthropic-version", "2023-06-01")
		return req
	})
}

// setupOpenAI configures OpenAI GPT with an API key.
func setupOpenAI(cfg *config.Config) bool {
	cfg.Agents.Defaults.Model = "gpt-5.2"

	fmt.Println()
	fmt.Println("  Get your API key at: https://platform.openai.com/api-keys")
	fmt.Println()

	apiKey := promptString("Enter your OpenAI API key: ")
	if apiKey == "" {
		fmt.Println("  [!] No API key provided — you can add it later in config.json")
		return false
	}

	setModelAPIKey(cfg, "gpt-5.2", apiKey)

	return validateAPIKey("OpenAI", apiKey, func(key string) *http.Request {
		req, _ := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
		req.Header.Set("Authorization", "Bearer "+key)
		return req
	})
}

// setupDeepSeek configures DeepSeek with an API key.
func setupDeepSeek(cfg *config.Config) bool {
	cfg.Agents.Defaults.Model = "deepseek-chat"

	fmt.Println()
	fmt.Println("  Get your API key at: https://platform.deepseek.com/")
	fmt.Println()

	apiKey := promptString("Enter your DeepSeek API key: ")
	if apiKey == "" {
		fmt.Println("  [!] No API key provided — you can add it later in config.json")
		return false
	}

	setModelAPIKey(cfg, "deepseek-chat", apiKey)

	return validateAPIKey("DeepSeek", apiKey, func(key string) *http.Request {
		req, _ := http.NewRequest("GET", "https://api.deepseek.com/v1/models", nil)
		req.Header.Set("Authorization", "Bearer "+key)
		return req
	})
}

// setupGoogleGemini configures Google Gemini via Antigravity OAuth.
func setupGoogleGemini(cfg *config.Config) bool {
	cfg.Agents.Defaults.Model = "gemini-flash"

	fmt.Println("\nOpening browser for Google login...")

	oauthCfg := auth.GoogleAntigravityOAuthConfig()
	cred, err := auth.LoginBrowser(oauthCfg)
	if err != nil {
		fmt.Printf("  [!] Google login failed: %v\n", err)
		fmt.Println("      You can try again later with: picoclaw auth login --provider google-antigravity")
		return false
	}

	cred.Provider = "google-antigravity"

	// Fetch project ID for Cloud Code Assist
	projectID, err := providers.FetchAntigravityProjectID(cred.AccessToken)
	if err != nil {
		fmt.Printf("  [!] Could not fetch project ID: %v\n", err)
		fmt.Println("      You may need Google Cloud Code Assist enabled on your account.")
	} else {
		cred.ProjectID = projectID
	}

	if err := auth.SetCredential("google-antigravity", cred); err != nil {
		fmt.Printf("  [!] Failed to save credentials: %v\n", err)
		return false
	}

	// Update config for antigravity
	cfg.Providers.Antigravity.AuthMethod = "oauth"
	for i := range cfg.ModelList {
		if isAntigravityModel(cfg.ModelList[i].Model) {
			cfg.ModelList[i].AuthMethod = "oauth"
			break
		}
	}

	fmt.Print("  [ok] Google login successful")
	if cred.Email != "" {
		fmt.Printf(" (%s)", cred.Email)
	}
	fmt.Println()
	return true
}

// setupOpenRouter configures OpenRouter with an API key.
func setupOpenRouter(cfg *config.Config) bool {
	cfg.Agents.Defaults.Model = "openrouter-auto"

	fmt.Println()
	fmt.Println("  Get your API key at: https://openrouter.ai/keys")
	fmt.Println()

	apiKey := promptString("Enter your OpenRouter API key: ")
	if apiKey == "" {
		fmt.Println("  [!] No API key provided — you can add it later in config.json")
		return false
	}

	// Set key on all openrouter model entries
	for i := range cfg.ModelList {
		if strings.HasPrefix(cfg.ModelList[i].ModelName, "openrouter") {
			cfg.ModelList[i].APIKey = apiKey
		}
	}

	return validateAPIKey("OpenRouter", apiKey, func(key string) *http.Request {
		req, _ := http.NewRequest("GET", "https://openrouter.ai/api/v1/models", nil)
		req.Header.Set("Authorization", "Bearer "+key)
		return req
	})
}

// validateAPIKey makes a test request to verify the API key works.
// Returns true if the key is valid, false otherwise (but doesn't fail hard).
func validateAPIKey(providerName, apiKey string, buildReq func(string) *http.Request) bool {
	fmt.Println("\nVerifying API key...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := buildReq(apiKey)
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("  [!] Could not reach %s (network error)\n", providerName)
		fmt.Println("      Key saved — double-check your connection later.")
		return false
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("  [ok] API key valid\n")
		return true
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		fmt.Printf("  [!] API key may be invalid (HTTP %d)\n", resp.StatusCode)
		fmt.Println("      Key saved — double-check it and try again.")
		return false
	}

	// Some APIs return non-200 for model listing but key might still work
	fmt.Printf("  [!] %s returned HTTP %d (key may still be valid)\n", providerName, resp.StatusCode)
	return false
}

// isAnthropicModel checks if a model string belongs to the anthropic provider.
func isAnthropicModel(model string) bool {
	return model == "anthropic" || strings.HasPrefix(model, "anthropic/")
}

// isAntigravityModel checks if a model string belongs to the antigravity provider.
func isAntigravityModel(model string) bool {
	return model == "antigravity" ||
		model == "google-antigravity" ||
		strings.HasPrefix(model, "antigravity/") ||
		strings.HasPrefix(model, "google-antigravity/")
}

func createWorkspaceTemplates(workspace string) {
	err := copyEmbeddedToTarget(workspace)
	if err != nil {
		fmt.Printf("Error copying workspace templates: %v\n", err)
	}
}

func copyEmbeddedToTarget(targetDir string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("creating target directory: %w", err)
	}

	// Walk through all files in embed.FS
	err := fs.WalkDir(embeddedFiles, "workspace", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Read embedded file
		data, err := embeddedFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading embedded file %s: %w", path, err)
		}

		relPath, err := filepath.Rel("workspace", path)
		if err != nil {
			return fmt.Errorf("getting relative path for %s: %w", path, err)
		}

		// Build target file path
		targetPath := filepath.Join(targetDir, relPath)

		// Ensure target file's directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", filepath.Dir(targetPath), err)
		}

		// Write file
		if err := os.WriteFile(targetPath, data, 0o644); err != nil {
			return fmt.Errorf("writing file %s: %w", targetPath, err)
		}

		return nil
	})

	return err
}
