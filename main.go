package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/FreePeak/commitgen/pkg/commitrules"
	"github.com/urfave/cli/v2"
)

// Version information (set by GoReleaser).
var (
	version = "v0.1.3"
	commit  = "none"
	date    = "unknown"
	builtBy = "local"
)

// Error definitions.
var (
	ErrNotGitRepo          = errors.New("not in a git repository")
	ErrNoChangesFound      = errors.New("no changes found to analyze")
	ErrNoStagedFiles       = errors.New("no staged files found")
	ErrNoUntrackedFiles    = errors.New("no untracked files found")
	ErrUnsupportedProvider = errors.New("unsupported provider")
	ErrPermissionDenied    = errors.New("permission denied. Try: sudo commitgen install")
)

func main() {
	app := &cli.App{
		Name:    "commitgen",
		Version: version,
		Usage:   "AI-powered git commit message generator",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "provider",
				Usage: "AI provider to use (claude*, claudex2, claudex3, gemini, copilot)",
				Value: "claudex2", // Default to claudex2 like gitcommit function
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
								Usage: "AI provider to use (claude*, claudex2, claudex3, gemini, copilot)",
								Value: "claudex2",
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
								Usage: "AI provider to use (claude*, claudex2, claudex3, gemini, copilot)",
								Value: "claudex2",
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
								Usage: "AI provider to use (claude*, claudex2, claudex3, gemini, copilot)",
								Value: "claudex2",
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
			return ErrNotGitRepo
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
			return fmt.Errorf("%w: unknown mode: %s", ErrNoChangesFound, mode)
		}

		if err != nil {
			return err
		}

		if analysisInput == "" {
			return ErrNoChangesFound
		}

		provider := c.String("provider")
		if provider == "" {
			provider = "claudex2" // default provider like gitcommit function
		}

		commitMessage, err := callAIAPI(analysisInput, provider)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}

		commitMessage = commitrules.CleanCommitMessage(commitMessage)

		// Validate commit message format
		if err := commitrules.ValidateCommitMessage(commitMessage); err != nil {
			fmt.Printf("Warning: %s\n", err)
		}

		fmt.Printf("Generated commit message:\n\"%s\"\n\n", commitMessage)

		fmt.Print("Do you want to use this commit message? [y/N] ")
		var response string
		_, err = fmt.Scanln(&response)
		if err != nil {
			// Handle scan error (e.g., EOF)
			response = ""
		}

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

// validateFilePath validates that a file path is safe to use.
func validateFilePath(path string) bool {
	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return false
	}
	// Check for dangerous characters
	if strings.ContainsAny(path, "&|;<>()$`\"'") {
		return false
	}
	// Check for empty path
	if path == "" {
		return false
	}
	return true
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
		return "", ErrNoStagedFiles
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
		if !validateFilePath(file) {
			continue
		}
		if _, err := os.Stat(file); err == nil {
			analysisInput.WriteString(fmt.Sprintf("\n--- %s ---\n", file))
			//nolint:gosec // G204: file path is validated by validateFilePath()
			cmd = exec.Command("git", "diff", "--cached", "--unified=3", "--", file)
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
		return "", ErrNoChangesFound
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
			if !validateFilePath(file) {
				continue
			}
			if _, err := os.Stat(file); err == nil {
				analysisInput.WriteString(fmt.Sprintf("\n--- %s ---\n", file))
				//nolint:gosec // G204: file path is validated by validateFilePath()
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
			if !validateFilePath(file) {
				continue
			}
			if _, err := os.Stat(file); err == nil {
				analysisInput.WriteString(fmt.Sprintf("\n--- %s (new) ---\n", file))
				//nolint:gosec // G304: file path is validated by validateFilePath()
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
		return "", ErrNoUntrackedFiles
	}

	var analysisInput strings.Builder
	analysisInput.WriteString("=== UNTRACKED FILES ANALYSIS ===\n")
	files := strings.Split(untrackedFiles, "\n")
	analysisInput.WriteString(fmt.Sprintf("Files: %d\n", len(files)))
	analysisInput.WriteString(fmt.Sprintf("%s\n\n", strings.Join(files, " ")))
	analysisInput.WriteString("=== FILE CONTENTS ===\n")

	for _, file := range files {
		if !validateFilePath(file) {
			continue
		}
		if _, err := os.Stat(file); err == nil {
			analysisInput.WriteString(fmt.Sprintf("\n--- %s ---\n", file))
			//nolint:gosec // G304: file path is validated by validateFilePath()
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
	prompt := commitrules.GetPrompt(analysisInput)

	var cmd *exec.Cmd

	// Handle special providers that are typically defined as aliases
	switch provider {
	case "claudex2":
		// Expand the claudex2 alias with actual environment variables
		cmd = exec.Command("claude")
		cmd.Env = append(os.Environ(),
			"ANTHROPIC_BASE_URL=https://open.bigmodel.cn/api/anthropic",
			"ANTHROPIC_API_KEY=40574464bbb949aa8323462fc7018fb0.QAedKFcjo0dkzzX4",
			"ANTHROPIC_MODEL=glm-4.6",
		)
	case "claudex3":
		// Expand the claudex3 alias with actual environment variables
		cmd = exec.Command("claude")
		cmd.Env = append(os.Environ(),
			"ANTHROPIC_BASE_URL=https://open.bigmodel.cn/api/anthropic",
			"ANTHROPIC_API_KEY=fc40be21f0c942898c8c76ff8adbcb03.UAD9tIMBeng4BYGM",
			"ANTHROPIC_MODEL=glm-4.6",
		)
	default:
		// For other providers, try direct execution first
		cmd = exec.Command(provider)
	}

	cmd.Stdin = strings.NewReader(prompt)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to call %s API: %w", provider, err)
	}

	return strings.TrimSpace(string(output)), nil
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
		return ErrPermissionDenied
	}

	// Copy binary to install path
	if !validateFilePath(exePath) {
		return fmt.Errorf("%w: invalid executable path: %s", ErrPermissionDenied, exePath)
	}
	//nolint:gosec // G304: exePath is validated by validateFilePath()
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

	// Make it executable (0755 is appropriate for system binaries in /usr/local/bin)
	// This allows read and execute by all users, but write only by owner
	//nolint:gosec // G302: 0755 is appropriate for system binaries
	err = os.Chmod(installPath, 0o755)
	if err != nil {
		return err
	}

	fmt.Printf("Commitgen installed successfully to %s\n", installPath)
	fmt.Println("You can now use 'commitgen' from anywhere!")
	return nil
}
