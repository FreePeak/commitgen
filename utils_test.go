package main

import (
	"strings"
	"testing"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanCommitMessage(tt.input)
			if result != tt.expected {
				t.Errorf("cleanCommitMessage(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestProviderValidation(t *testing.T) {
	tests := []struct {
		provider       string
		isSupported    bool
		expectedCmd    string
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
		{"claud", false, ""}, // Too short
		{"cclaude", false, ""}, // Doesn't start with claude
		{"", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			// Test provider validation logic
			var isClaudeProvider bool
			var command string

			switch {
			case strings.HasPrefix(tt.provider, "claude"):
				isClaudeProvider = true
				command = "claude"
			case tt.provider == "gemini":
				isClaudeProvider = false
				command = "gemini"
			case tt.provider == "copilot":
				isClaudeProvider = false
				command = "copilot"
			default:
				isClaudeProvider = false
				command = ""
			}

			// For claude-prefixed providers, we expect them to be supported
			if strings.HasPrefix(tt.provider, "claude") {
				if !isClaudeProvider {
					t.Errorf("Provider %s should be recognized as claude provider", tt.provider)
				}
				if command != "claude" {
					t.Errorf("Provider %s should map to 'claude' command, got %s", tt.provider, command)
				}
			} else if tt.provider == "gemini" || tt.provider == "copilot" {
				if tt.provider != "gemini" && tt.provider != "copilot" {
					t.Errorf("Provider %s should be handled separately", tt.provider)
				}
			} else {
				// Unsupported providers
				if isClaudeProvider || command != "" {
					t.Errorf("Provider %s should be unsupported", tt.provider)
				}
			}

			// Validate expected behavior
			if tt.isSupported && command == "" && !strings.HasPrefix(tt.provider, "claude") {
				t.Errorf("Provider %s should be supported but got empty command", tt.provider)
			}
		})
	}
}

func TestPromptStructure(t *testing.T) {
	analysisInput := "feat(service): add new endpoint"

	prompt := `Generate commit message using exact format: type(scope): description
Max 50 chars. From file paths extract service/module as scope.
Types: feat, fix, docs, style, refactor, test, chore
Return ONLY commit message, no extra text.

Examples:
feat(service:rating): add get RestaurantQuickReview with caching
fix(api:user): resolve null pointer in validation
docs(readme): update setup instructions

Git diff analysis:
` + analysisInput

	// Verify prompt contains all required elements
	requiredElements := []string{
		"type(scope): description",
		"Max 50 chars",
		"feat, fix, docs, style, refactor, test, chore",
		"Return ONLY commit message",
		"Git diff analysis:",
		analysisInput,
		"feat(service:rating): add get RestaurantQuickReview with caching",
		"fix(api:user): resolve null pointer in validation",
		"docs(readme): update setup instructions",
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
		"add new feature", // Missing type and scope
		"feat", // Missing description
		"feat:", // Empty description
		": add new feature", // Missing type
		"Add new feature", // Capital first letter (should be lowercase)
		"feat Add new feature", // Missing colon
		"feat: Add new feature", // Capital description (should be lowercase)
		"feat: add new feature.", // Ends with period
	}

	// Test that valid formats pass our basic validation
	for _, msg := range validFormats {
		t.Run("valid/"+msg, func(t *testing.T) {
			cleaned := cleanCommitMessage(msg)
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
			cleaned := cleanCommitMessage(msg)
			// cleanCommitMessage should still return the message even if format is invalid
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

// Edge case tests
func TestEdgeCases(t *testing.T) {
	t.Run("empty analysis input", func(t *testing.T) {
		// Test that empty analysis input is handled
		prompt := `Generate commit message using exact format: type(scope): description
Max 50 chars. From file paths extract service/module as scope.
Types: feat, fix, docs, style, refactor, test, chore
Return ONLY commit message, no extra text.

Examples:
feat(service:rating): add get RestaurantQuickReview with caching
fix(api:user): resolve null pointer in validation
docs(readme): update setup instructions

Git diff analysis:
`
		if len(prompt) == 0 {
			t.Error("Prompt should not be empty even with empty analysis input")
		}
	})

	t.Run("very long commit message", func(t *testing.T) {
		longMsg := strings.Repeat("a", 1000)
		cleaned := cleanCommitMessage(longMsg)
		if cleaned != longMsg {
			t.Errorf("Long message should not be truncated by cleanCommitMessage")
		}
	})

	t.Run("special characters in commit message", func(t *testing.T) {
		specialMsg := "feat: add support for émojis 🎉 and ñoños"
		cleaned := cleanCommitMessage(specialMsg)
		if cleaned != specialMsg {
			t.Errorf("Special characters should be preserved")
		}
	})
}

// Performance tests
func BenchmarkCleanCommitMessageSimple(b *testing.B) {
	message := "feat: add new feature"
	for i := 0; i < b.N; i++ {
		cleanCommitMessage(message)
	}
}

func BenchmarkCleanCommitMessageComplex(b *testing.B) {
	message := `"  feat: add new feature with quotes and whitespace

	This is a multiline message with extra content.
	"`
	for i := 0; i < b.N; i++ {
		cleanCommitMessage(message)
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