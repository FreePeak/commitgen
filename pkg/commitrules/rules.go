package commitrules

import (
	"fmt"
	"regexp"
	"strings"
)

// CommitRule defines the structure for commit message rules
type CommitRule struct {
	Type        string
	Description string
	Examples    []string
}

// CommitRules holds all available commit types and their rules
var CommitRules = map[string]CommitRule{
	"feat": {
		Type:        "feat",
		Description: "A new feature",
		Examples:    []string{"feat(core): add user authentication service", "feat(ui): implement dark mode toggle"},
	},
	"fix": {
		Type:        "fix",
		Description: "A bug fix",
		Examples:    []string{"fix(api): resolve null pointer in validation", "fix(ui): correct button alignment"},
	},
	"docs": {
		Type:        "docs",
		Description: "Documentation only changes",
		Examples:    []string{"docs(readme): update installation instructions", "docs(api): add endpoint documentation"},
	},
	"style": {
		Type:        "style",
		Description: "Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)",
		Examples:    []string{"style(utils): format code with prettier", "style(ui): fix indentation"},
	},
	"refactor": {
		Type:        "refactor",
		Description: "A code change that neither fixes a bug nor adds a feature",
		Examples:    []string{"refactor(utils): extract validation logic", "refactor(api): simplify request handling"},
	},
	"test": {
		Type:        "test",
		Description: "Adding missing tests or correcting existing tests",
		Examples:    []string{"test(core): add unit tests for user service", "test(api): fix integration tests"},
	},
	"chore": {
		Type:        "chore",
		Description: "Other changes that don't modify src or test files",
		Examples:    []string{"chore(deps): update dependencies", "chore(build): update build configuration"},
	},
}

// GetCommitTypes returns all available commit types
func GetCommitTypes() []string {
	var types []string
	for commitType := range CommitRules {
		types = append(types, commitType)
	}
	return types
}

// GetPrompt generates the commit message prompt based on analysis input
func GetPrompt(analysisInput string) string {
	commitTypesList := strings.Join(GetCommitTypes(), ", ")

	prompt := fmt.Sprintf(`You are a commit message generator. Your ONLY task is to output a single conventional commit message.

FORMAT: type(scope): description
RULES:
- Maximum 50 characters total
- Types: %s
- Extract scope from file paths (api, ui, core, scripts, pkg, etc.)
- Use lowercase, present tense, imperative mood
- No periods, quotes, or extra text

EXAMPLE OUTPUTS:
feat(core): add user authentication
fix(api): resolve null pointer exception
docs(readme): update installation guide
style(ui): format button components
refactor(db): simplify query logic
test(auth): add unit tests for login
chore(deps): update go modules

CRITICAL: Respond with ONLY the commit message. No explanations, no quotes, no "Here is the commit message:", no extra text whatsoever.

Git diff to analyze:
%s`, commitTypesList, analysisInput)

	return prompt
}

// CleanCommitMessage cleans and formats the generated commit message
func CleanCommitMessage(message string) string {
	// Remove quotes and extra whitespace
	message = strings.Trim(message, `"'`)
	message = strings.TrimSpace(message)

	// Split into lines and take the first one
	lines := strings.Split(message, "\n")
	firstLine := strings.TrimSpace(lines[0])

	// Try to extract a proper commit message from descriptive text
	if strings.Contains(firstLine, ":") {
		// Check if it looks like a proper commit message (type(scope): description)
		parts := strings.SplitN(firstLine, ":", 2)
		if len(parts) == 2 {
			typePart := strings.TrimSpace(parts[0])
			descPart := strings.TrimSpace(parts[1])
			if len(typePart) > 0 && len(descPart) > 0 && len(firstLine) < 72 {
				return firstLine
			}
		}
	}

	// Look for patterns that might contain a commit message
	patterns := []string{
		`commit message:\s*"?([^"]+)"?`,           // "commit message: 'feat: add feature'"
		`should be:\s*"?([^"]+)"?`,              // "should be: 'fix: resolve bug'"
		`message is:\s*"?([^"]+)"?`,             // "message is: 'docs: update readme'"
		`message:\s*"?([^"]+)"?`,                // "message: 'docs: update readme'"
		`^\s*([a-z]+\([^)]+\):\s*[^.]+)`,       // Direct match at start
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(firstLine)
		if len(matches) > 1 {
			candidate := strings.TrimSpace(matches[1])
			if strings.Contains(candidate, ":") && len(candidate) < 72 {
				return strings.Trim(candidate, `"'`)
			}
		}
	}

	// If no good pattern found, but first line looks like a commit message, use it
	if strings.Contains(firstLine, ":") && len(firstLine) < 72 {
		return firstLine
	}

	// Last resort: if we can't extract a good commit message, return the cleaned first line
	return firstLine
}

// ValidateCommitMessage validates if a commit message follows the conventional format
func ValidateCommitMessage(message string) error {
	message = strings.TrimSpace(message)

	// Check basic format type(scope): description
	parts := strings.SplitN(message, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("commit message must follow format: type(scope): description")
	}

	// Check if type is valid
	typeAndScope := strings.TrimSpace(parts[0])
	scopeParts := strings.SplitN(typeAndScope, "(", 2)
	if len(scopeParts) == 0 {
		return fmt.Errorf("commit message must have a type")
	}

	commitType := scopeParts[0]
	if _, exists := CommitRules[commitType]; !exists {
		return fmt.Errorf("invalid commit type: %s. Valid types: %s", commitType, strings.Join(GetCommitTypes(), ", "))
	}

	// Check length
	if len(message) > 72 {
		return fmt.Errorf("commit message is too long: %d characters (maximum: 72)", len(message))
	}

	if len(message) > 50 {
		fmt.Printf("Warning: Commit message is %d characters (recommended: <50)\n", len(message))
	}

	return nil
}
