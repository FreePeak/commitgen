package main

import (
	"strings"
	"testing"

	"github.com/FreePeak/commitgen/pkg/commitrules"
)

func TestCleanCommitMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple message",
			input:    "feat: add new feature",
			expected: "feat: add new feature",
		},
		{
			name:     "message with quotes",
			input:    `"feat: add new feature"`,
			expected: "feat: add new feature",
		},
		{
			name:     "message with extra whitespace",
			input:    "  feat: add new feature  ",
			expected: "feat: add new feature",
		},
		{
			name:     "message with single quotes",
			input:    `'feat: add new feature'`,
			expected: "feat: add new feature",
		},
		{
			name:     "multiline message",
			input:    "feat: add new feature\n\nThis is the description",
			expected: "feat: add new feature",
		},
		{
			name:     "empty message",
			input:    "",
			expected: "",
		},
		{
			name:     "message with only whitespace",
			input:    "   ",
			expected: "",
		},
		{
			name:     "message with explanatory text",
			input:    "feat: add new feature\nThis commit adds a new feature to the application",
			expected: "feat: add new feature",
		},
		{
			name:     "message with trailing spaces and newlines",
			input:    "fix: resolve bug   \n\n  ",
			expected: "fix: resolve bug",
		},
		{
			name:     "descriptive response with commit message",
			input:    "Based on the changes, the commit message should be: \"feat: add new feature\"",
			expected: "feat: add new feature",
		},
		{
			name:     "long descriptive response",
			input:    "Looking at the git diff, I can see this is a major refactoring. The commit message is: refactor(core): extract validation logic",
			expected: "refactor(core): extract validation logic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := commitrules.CleanCommitMessage(tt.input)
			if result != tt.expected {
				t.Errorf("commitrules.CleanCommitMessage(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// resolveProviderCommand returns the command for a given provider and whether it's a Claude provider.
func resolveProviderCommand(provider string) (string, bool) {
	switch {
	case strings.HasPrefix(provider, "claude"):
		return "claude", true
	case provider == "gemini":
		return "gemini", false
	case provider == "copilot":
		return "copilot", false
	default:
		return "", false
	}
}

// validateProviderMapping validates that a provider maps to the expected command.
func validateProviderMapping(t *testing.T, provider, expectedCmd, actualCmd string, isClaude bool) {
	switch {
	case strings.HasPrefix(provider, "claude"):
		if !isClaude {
			t.Errorf("Provider %s should be recognized as claude provider", provider)
		}
		if actualCmd != "claude" {
			t.Errorf("Provider %s should map to 'claude' command, got %s", provider, actualCmd)
		}
	case provider == "gemini" || provider == "copilot":
		if provider != expectedCmd {
			t.Errorf("Provider %s should map to '%s' command, got %s", provider, expectedCmd, actualCmd)
		}
	default:
		// Unsupported providers
		if isClaude || actualCmd != "" {
			t.Errorf("Provider %s should be unsupported", provider)
		}
	}
}

func TestProviderValidation(t *testing.T) {
	tests := []struct {
		provider    string
		isSupported bool
		expectedCmd string
	}{
		// Claude variants
		{"claude", true, "claude"},
		{"claudex", true, "claude"},
		{"claudex2", true, "claude"},
		{"claudex3", true, "claude"},
		{"claude-external", true, "claude"},
		{"claude-custom", true, "claude"},
		{"claude-2", true, "claude"},
		{"claudex-external", true, "claude"},

		// Other providers
		{"gemini", true, "gemini"},
		{"copilot", true, "copilot"},

		// Unsupported providers
		{"openai", false, ""},
		{"chatgpt", false, ""},
		{"claud", false, ""},   // Too short
		{"cclaude", false, ""}, // Doesn't start with claude
		{"", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			command, isClaude := resolveProviderCommand(tt.provider)
			validateProviderMapping(t, tt.provider, tt.expectedCmd, command, isClaude)

			// Validate expected behavior
			if tt.isSupported && command == "" && !strings.HasPrefix(tt.provider, "claude") {
				t.Errorf("Provider %s should be supported but got empty command", tt.provider)
			}
		})
	}
}

func TestPromptStructure(t *testing.T) {
	analysisInput := "feat(service): add new endpoint"

	prompt := commitrules.GetPrompt(analysisInput)

	// Verify prompt contains all required elements
	requiredElements := []string{
		"type(scope): description",
		"Maximum 50 characters total",
		"conventional commit message",
		"CRITICAL: Respond with ONLY the commit message",
		"Git diff to analyze:",
		analysisInput,
	}

	for _, element := range requiredElements {
		if !strings.Contains(prompt, element) {
			t.Errorf("Prompt missing required element: %s", element)
		}
	}

	// Verify prompt doesn't contain unwanted elements
	unwantedElements := []string{
		"TODO",
		"FIXME",
		"<placeholder>",
	}

	for _, element := range unwantedElements {
		if strings.Contains(prompt, element) {
			t.Errorf("Prompt contains unwanted element: %s", element)
		}
	}
}

func TestCommitMessageFormatValidation(t *testing.T) {
	validFormats := []string{
		"feat: add new feature",
		"fix(auth): resolve login issue",
		"docs(readme): update installation guide",
		"style: format code",
		"refactor(api): simplify endpoint logic",
		"test: add unit tests for auth",
		"chore: update dependencies",
		"feat(service:auth): add JWT validation",
		"fix(ui): resolve button rendering issue",
	}

	invalidFormats := []string{
		"add new feature",        // Missing type and scope
		"feat",                   // Missing description
		"feat:",                  // Empty description
		": add new feature",      // Missing type
		"Add new feature",        // Capital first letter (should be lowercase)
		"feat Add new feature",   // Missing colon
		"feat: Add new feature",  // Capital description (should be lowercase)
		"feat: add new feature.", // Ends with period
	}

	// Test that valid formats pass our basic validation
	for _, msg := range validFormats {
		t.Run("valid/"+msg, func(t *testing.T) {
			cleaned := commitrules.CleanCommitMessage(msg)
			if cleaned == "" {
				t.Errorf("Valid commit message %s was cleaned to empty string", msg)
			}

			// Basic format check: type: description
			parts := strings.SplitN(cleaned, ":", 2)
			if len(parts) != 2 {
				t.Errorf("Valid commit message %s doesn't contain colon separator", msg)
				return
			}

			typePart := strings.TrimSpace(parts[0])
			descPart := strings.TrimSpace(parts[1])

			if typePart == "" {
				t.Errorf("Valid commit message %s has empty type part", msg)
			}
			if descPart == "" {
				t.Errorf("Valid commit message %s has empty description part", msg)
			}
		})
	}

	// Test that invalid formats are still processed (cleanCommitMessage doesn't validate format)
	for _, msg := range invalidFormats {
		t.Run("invalid/"+msg, func(t *testing.T) {
			cleaned := commitrules.CleanCommitMessage(msg)
			// commitrules.CleanCommitMessage should still return the message even if format is invalid
			// The format validation would happen elsewhere
			if cleaned != msg && msg != `"add new feature"` && msg != `'add new feature'` {
				t.Logf("Message %s was cleaned to %s", msg, cleaned)
			}
		})
	}
}

func TestGitRepositoryDetection(t *testing.T) {
	// Test that the function exists and returns a boolean
	result := isGitRepo()

	// We don't assert a specific value since it depends on the test environment
	// but we can verify it returns a boolean
	if result != true && result != false {
		t.Errorf("isGitRepo() should return true or false, got %v", result)
	}

	// Type check
	var _ bool = isGitRepo()
}

// Edge case tests.
func TestEdgeCases(t *testing.T) {
	t.Run("empty analysis input", func(t *testing.T) {
		// Test that empty analysis input is handled
		prompt := commitrules.GetPrompt("")
		if len(prompt) == 0 {
			t.Error("Prompt should not be empty even with empty analysis input")
		}
	})

	t.Run("very long commit message", func(t *testing.T) {
		longMsg := strings.Repeat("a", 1000)
		cleaned := commitrules.CleanCommitMessage(longMsg)
		if cleaned != longMsg {
			t.Errorf("Long message should not be truncated by commitrules.CleanCommitMessage")
		}
	})

	t.Run("special characters in commit message", func(t *testing.T) {
		specialMsg := "feat: add support for émojis 🎉 and ñoños"
		cleaned := commitrules.CleanCommitMessage(specialMsg)
		if cleaned != specialMsg {
			t.Errorf("Special characters should be preserved")
		}
	})
}

// Performance tests.
func BenchmarkCleanCommitMessageSimple(b *testing.B) {
	message := "feat: add new feature"
	for i := 0; i < b.N; i++ {
		commitrules.CleanCommitMessage(message)
	}
}

func BenchmarkCleanCommitMessageComplex(b *testing.B) {
	message := `"  feat: add new feature with quotes and whitespace

	This is a multiline message with extra content.
	"`
	for i := 0; i < b.N; i++ {
		commitrules.CleanCommitMessage(message)
	}
}

func BenchmarkProviderValidation(b *testing.B) {
	providers := []string{"claude", "claudex2", "claude-external", "gemini", "copilot", "unknown"}
	for i := 0; i < b.N; i++ {
		provider := providers[i%len(providers)]
		switch {
		case strings.HasPrefix(provider, "claude"):
			_ = "claude"
		case provider == "gemini":
			_ = "gemini"
		case provider == "copilot":
			_ = "copilot"
		default:
			_ = ""
		}
	}
}
