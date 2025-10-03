package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
)

// Version information (set by GoReleaser).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "local"
)

// AI provider constants.
const (
	ProviderClaude  = "claude"
	ProviderGemini  = "gemini"
	ProviderCopilot = "copilot"
)

func main() {
	app := &cli.App{
		Name:    "commitgen",
		Version: version,
		Usage:   "AI-powered git commit message generator",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "provider",
				Usage: "AI provider to use (claude*, gemini, copilot)",
				Value: ProviderClaude,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "commit",
				Aliases: []string{"c"},
				Usage:   "Generate commit message from changes",
				Subcommands: []*cli.Command{
					{
						Name:    "staged",
						Aliases: []string{"s"},
						Usage:   "Generate from staged files",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "provider",
								Usage: "AI provider to use (claude*, gemini, copilot)",
								Value: ProviderClaude,
							},
						},
						Action: generateCommitMessage("staged"),
					},
					{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "Generate from all changes",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "provider",
								Usage: "AI provider to use (claude*, gemini, copilot)",
								Value: ProviderClaude,
							},
						},
						Action: generateCommitMessage("all"),
					},
					{
						Name:    "untracked",
						Aliases: []string{"u"},
						Usage:   "Generate from untracked files",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "provider",
								Usage: "AI provider to use (claude*, gemini, copilot)",
								Value: ProviderClaude,
							},
						},
						Action: generateCommitMessage("untracked"),
					},
				},
			},
			{
				Name:   "install",
				Usage:  "Install commitgen to /usr/local/bin",
				Action: installBinary,
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show version information",
				Action: func(c *cli.Context) error {
					fmt.Printf("Commitgen %s\n", version)
					fmt.Printf("Commit: %s\n", commit)
					fmt.Printf("Built: %s\n", date)
					fmt.Printf("Built by: %s\n", builtBy)
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			return generateCommitMessage("staged")(c)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func generateCommitMessage(mode string) cli.ActionFunc {
	return func(c *cli.Context) error {
		// Check if we're in a git repository
		if !isGitRepo() {
			return fmt.Errorf("not in a git repository")
		}

		var analysisInput string
		var err error

		switch mode {
		case "staged":
			analysisInput, err = analyzeStagedChanges()
		case "all":
			analysisInput, err = analyzeAllChanges()
		case "untracked":
			analysisInput, err = analyzeUntrackedFiles()
		default:
			return fmt.Errorf("unknown mode: %s", mode)
		}

		if err != nil {
			return err
		}

		if analysisInput == "" {
			return fmt.Errorf("no changes found to analyze")
		}

		provider := c.String("provider")
		if provider == "" {
			provider = ProviderClaude // default provider
		}

		commitMessage, err := callAIAPI(analysisInput, provider)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}

		commitMessage = cleanCommitMessage(commitMessage)

		if len(commitMessage) > 72 {
			fmt.Printf("Warning: Commit message is %d characters (recommended: <72)\n", len(commitMessage))
		}

		fmt.Printf("Generated commit message:\n\"%s\"\n\n", commitMessage)

		fmt.Print("Do you want to use this commit message? [y/N] ")
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			if err := executeCommit(mode, commitMessage); err != nil {
				return err
			}
			fmt.Println("Committed successfully!")
		} else {
			fmt.Println("Commit cancelled.")
		}

		return nil
	}
}

func isGitRepo() bool {
	_, err := exec.Command("git", "rev-parse", "--git-dir").CombinedOutput()
	return err == nil
}

func analyzeStagedChanges() (string, error) {
	// Get staged files
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	stagedFiles := strings.TrimSpace(string(output))
	if stagedFiles == "" {
		return "", fmt.Errorf("no staged files found")
	}

	var analysisInput strings.Builder
	analysisInput.WriteString("=== STAGED CHANGES ANALYSIS ===\n")
	files := strings.Split(stagedFiles, "\n")
	analysisInput.WriteString(fmt.Sprintf("Files changed: %d\n", len(files)))
	analysisInput.WriteString(fmt.Sprintf("Files: %s\n\n", strings.Join(files, " ")))

	// Get diff stats
	cmd = exec.Command("git", "diff", "--cached", "--stat")
	output, _ = cmd.Output()
	analysisInput.WriteString("=== DIFF ===\n")
	analysisInput.Write(output)
	analysisInput.WriteString("\n=== DETAILED CHANGES ===\n")

	// Get detailed diff for each file
	for _, file := range files {
		if file == "" {
			continue
		}
		if _, err := os.Stat(file); err == nil {
			analysisInput.WriteString(fmt.Sprintf("\n--- %s ---\n", file))
			cmd = exec.Command("git", "diff", "--cached", "--unified=3", file)
			output, _ := cmd.Output()
			if len(output) > 2000 {
				output = output[:2000]
			}
			analysisInput.Write(output)
		}
	}

	return analysisInput.String(), nil
}

func analyzeAllChanges() (string, error) {
	// Get modified files
	cmd := exec.Command("git", "diff", "--name-only")
	modifiedOutput, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Get untracked files
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	untrackedOutput, err := cmd.Output()
	if err != nil {
		return "", err
	}

	modifiedFiles := strings.TrimSpace(string(modifiedOutput))
	untrackedFiles := strings.TrimSpace(string(untrackedOutput))

	if modifiedFiles == "" && untrackedFiles == "" {
		return "", fmt.Errorf("no changes found")
	}

	var analysisInput strings.Builder
	analysisInput.WriteString("=== ALL CHANGES ANALYSIS ===\n")

	if modifiedFiles != "" {
		files := strings.Split(modifiedFiles, "\n")
		analysisInput.WriteString(fmt.Sprintf("Modified files: %d\n", len(files)))
		analysisInput.WriteString("=== MODIFIED FILES ===\n")
		analysisInput.WriteString(fmt.Sprintf("%s\n\n", strings.Join(files, " ")))
		analysisInput.WriteString("=== MODIFICATIONS ===\n")

		for _, file := range files {
			if file == "" {
				continue
			}
			if _, err := os.Stat(file); err == nil {
				analysisInput.WriteString(fmt.Sprintf("\n--- %s ---\n", file))
				cmd = exec.Command("git", "diff", "--unified=3", file)
				output, _ := cmd.Output()
				if len(output) > 2000 {
					output = output[:2000]
				}
				analysisInput.Write(output)
			}
		}
	}

	if untrackedFiles != "" {
		files := strings.Split(untrackedFiles, "\n")
		analysisInput.WriteString("\n=== UNTRACKED FILES ===\n")
		analysisInput.WriteString(fmt.Sprintf("%s\n\n", strings.Join(files, " ")))
		analysisInput.WriteString("=== FILE CONTENTS ===\n")

		for _, file := range files {
			if file == "" {
				continue
			}
			if _, err := os.Stat(file); err == nil {
				analysisInput.WriteString(fmt.Sprintf("\n--- %s (new) ---\n", file))
				content, _ := os.ReadFile(file)
				if len(content) > 2000 {
					content = content[:2000]
				}
				analysisInput.Write(content)
			}
		}
	}

	return analysisInput.String(), nil
}

func analyzeUntrackedFiles() (string, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	untrackedFiles := strings.TrimSpace(string(output))
	if untrackedFiles == "" {
		return "", fmt.Errorf("no untracked files found")
	}

	var analysisInput strings.Builder
	analysisInput.WriteString("=== UNTRACKED FILES ANALYSIS ===\n")
	files := strings.Split(untrackedFiles, "\n")
	analysisInput.WriteString(fmt.Sprintf("Files: %d\n", len(files)))
	analysisInput.WriteString(fmt.Sprintf("%s\n\n", strings.Join(files, " ")))
	analysisInput.WriteString("=== FILE CONTENTS ===\n")

	for _, file := range files {
		if file == "" {
			continue
		}
		if _, err := os.Stat(file); err == nil {
			analysisInput.WriteString(fmt.Sprintf("\n--- %s ---\n", file))
			content, _ := os.ReadFile(file)
			if len(content) > 2000 {
				content = content[:2000]
			}
			analysisInput.Write(content)
		}
	}

	return analysisInput.String(), nil
}

func callAIAPI(analysisInput, provider string) (string, error) {
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

	var cmd *exec.Cmd

	switch {
	case strings.HasPrefix(provider, ProviderClaude):
		cmd = exec.Command("claude")
	case provider == ProviderGemini:
		cmd = exec.Command("gemini")
	case provider == ProviderCopilot:
		cmd = exec.Command("copilot")
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}

	cmd.Stdin = strings.NewReader(prompt)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to call %s API: %w", provider, err)
	}

	return strings.TrimSpace(string(output)), nil
}

func cleanCommitMessage(message string) string {
	// Remove quotes and extra whitespace
	message = strings.Trim(message, `"'`)
	message = strings.TrimSpace(message)

	// Remove any remaining explanatory text
	lines := strings.Split(message, "\n")
	if len(lines) > 1 {
		message = strings.TrimSpace(lines[0])
	}

	return message
}

func executeCommit(mode, commitMessage string) error {
	var cmd *exec.Cmd

	switch mode {
	case "staged":
		cmd = exec.Command("git", "commit", "-m", commitMessage)
	case "all", "untracked":
		// First stage all changes
		if err := exec.Command("git", "add", ".").Run(); err != nil {
			return err
		}
		cmd = exec.Command("git", "commit", "-m", commitMessage)
	}

	return cmd.Run()
}

func installBinary(c *cli.Context) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	installPath := "/usr/local/bin/commitgen"

	// Check if we have permission to write to /usr/local/bin
	if _, err := os.Stat("/usr/local/bin"); os.IsPermission(err) {
		return fmt.Errorf("permission denied. Try: sudo commitgen install")
	}

	// Copy binary to install path
	source, err := os.Open(exePath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(installPath)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	// Make it executable
	err = os.Chmod(installPath, 0755)
	if err != nil {
		return err
	}

	fmt.Printf("Commitgen installed successfully to %s\n", installPath)
	fmt.Println("You can now use 'commitgen' from anywhere!")
	return nil
}
