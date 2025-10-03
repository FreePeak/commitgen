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
	version = "v0.1.4" // test comment
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
	app := createApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func createApp() *cli.App {
	return &cli.App{
		Name:    "commitgen",
		Version: version,
		Usage:   "AI-powered git commit message generator",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "provider",
				Usage: "AI provider to use (claude*, claude, claude, gemini, copilot)",
				Value: "claude", // Default to claude like gitcommit function
			},
		},
		Commands: []*cli.Command{
			createCommitCommand(),
			{
				Name:   "install",
				Usage:  "Install commitgen to /usr/local/bin",
				Action: installBinary,
			},
			createVersionCommand(),
		},
		Action: func(c *cli.Context) error {
			return generateCommitMessage("staged")(c)
		},
	}
}

func createCommitCommand() *cli.Command {
	return &cli.Command{
		Name:    "commit",
		Aliases: []string{"c"},
		Usage:   "Generate commit message from changes",
		Subcommands: []*cli.Command{
			createStagedCommand(),
			createAllCommand(),
			createUntrackedCommand(),
		},
	}
}

func createStagedCommand() *cli.Command {
	return &cli.Command{
		Name:    "staged",
		Aliases: []string{"s"},
		Usage:   "Generate from staged files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "provider",
				Usage: "AI provider to use (claude*, claude, claude, gemini, copilot)",
				Value: "claude",
			},
		},
		Action: generateCommitMessage("staged"),
	}
}

func createAllCommand() *cli.Command {
	return &cli.Command{
		Name:    "all",
		Aliases: []string{"a"},
		Usage:   "Generate from all changes",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "provider",
				Usage: "AI provider to use (claude*, claude, claude, gemini, copilot)",
				Value: "claude",
			},
		},
		Action: generateCommitMessage("all"),
	}
}

func createUntrackedCommand() *cli.Command {
	return &cli.Command{
		Name:    "untracked",
		Aliases: []string{"u"},
		Usage:   "Generate from untracked files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "provider",
				Usage: "AI provider to use (claude*, claude, claude, gemini, copilot)",
				Value: "claude",
			},
		},
		Action: generateCommitMessage("untracked"),
	}
}

func createVersionCommand() *cli.Command {
	return &cli.Command{
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
	}
}

func generateCommitMessage(mode string) cli.ActionFunc {
	return func(cliContext *cli.Context) error {
		if !isGitRepo() {
			return ErrNotGitRepo
		}

		analysisInput, err := getAnalysisInput(mode)
		if err != nil {
			return err
		}

		provider := getProvider(cliContext)
		commitMessage, err := callAIAPI(analysisInput, provider)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}

		commitMessage = commitrules.CleanCommitMessage(commitMessage)
		validateAndShowWarning(commitMessage)

		if confirmCommit(commitMessage) {
			return executeCommit(mode, commitMessage)
		}
		fmt.Println("Commit cancelled.")
		return nil
	}
}

func getAnalysisInput(mode string) (string, error) {
	switch mode {
	case "staged":
		return analyzeStagedChanges()
	case "all":
		return analyzeAllChanges()
	case "untracked":
		return analyzeUntrackedFiles()
	default:
		return "", fmt.Errorf("%w: unknown mode: %s", ErrNoChangesFound, mode)
	}
}

func getProvider(cliContext *cli.Context) string {
	provider := cliContext.String("provider")
	if provider == "" {
		provider = "claude"
	}
	return provider
}

func validateAndShowWarning(commitMessage string) {
	if err := commitrules.ValidateCommitMessage(commitMessage); err != nil {
		fmt.Printf("Warning: %s\n", err)
	}
}

func confirmCommit(commitMessage string) bool {
	fmt.Printf("Generated commit message:\n\"%s\"\n\n", commitMessage)
	fmt.Print("Do you want to use this commit message? [y/N] ")

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		response = ""
	}

	confirmed := strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
	if confirmed {
		fmt.Println("Committed successfully!")
	}
	return confirmed
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
		return "", fmt.Errorf("failed to get staged files: %w", err)
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
	modifiedFiles, untrackedFiles, err := getModifiedAndUntrackedFiles()
	if err != nil {
		return "", err
	}

	if modifiedFiles == "" && untrackedFiles == "" {
		return "", ErrNoChangesFound
	}

	var analysisInput strings.Builder
	analysisInput.WriteString("=== ALL CHANGES ANALYSIS ===\n")

	if modifiedFiles != "" {
		addModifiedFilesToAnalysis(&analysisInput, modifiedFiles)
	}

	if untrackedFiles != "" {
		addUntrackedFilesToAnalysis(&analysisInput, untrackedFiles)
	}

	return analysisInput.String(), nil
}

func getModifiedAndUntrackedFiles() (string, string, error) {
	cmd := exec.Command("git", "diff", "--name-only")
	modifiedOutput, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get modified files: %w", err)
	}

	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	untrackedOutput, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get untracked files: %w", err)
	}

	return strings.TrimSpace(string(modifiedOutput)), strings.TrimSpace(string(untrackedOutput)), nil
}

func addModifiedFilesToAnalysis(analysisInput *strings.Builder, modifiedFiles string) {
	files := strings.Split(modifiedFiles, "\n")
	fmt.Fprintf(analysisInput, "Modified files: %d\n", len(files))
	analysisInput.WriteString("=== MODIFIED FILES ===\n")
	fmt.Fprintf(analysisInput, "%s\n\n", strings.Join(files, " "))
	analysisInput.WriteString("=== MODIFICATIONS ===\n")

	for _, file := range files {
		if !validateFilePath(file) {
			continue
		}
		if _, err := os.Stat(file); err == nil {
			addFileDiffToAnalysis(analysisInput, file)
		}
	}
}

func addUntrackedFilesToAnalysis(analysisInput *strings.Builder, untrackedFiles string) {
	files := strings.Split(untrackedFiles, "\n")
	analysisInput.WriteString("\n=== UNTRACKED FILES ===\n")
	fmt.Fprintf(analysisInput, "%s\n\n", strings.Join(files, " "))
	analysisInput.WriteString("=== FILE CONTENTS ===\n")

	for _, file := range files {
		if !validateFilePath(file) {
			continue
		}
		if _, err := os.Stat(file); err == nil {
			addFileContentToAnalysis(analysisInput, file)
		}
	}
}

func addFileDiffToAnalysis(analysisInput *strings.Builder, file string) {
	fmt.Fprintf(analysisInput, "\n--- %s ---\n", file)
	cmd := exec.Command("git", "diff", "--unified=3", file)
	output, _ := cmd.Output()
	if len(output) > 2000 {
		output = output[:2000]
	}
	analysisInput.Write(output)
}

func addFileContentToAnalysis(analysisInput *strings.Builder, file string) {
	fmt.Fprintf(analysisInput, "\n--- %s (new) ---\n", file)
	//nolint:gosec // G304: file path is validated by validateFilePath()
	content, _ := os.ReadFile(file)
	if len(content) > 2000 {
		content = content[:2000]
	}
	analysisInput.Write(content)
}

func analyzeUntrackedFiles() (string, error) {
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get untracked files: %w", err)
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
	case "claude":
		// Expand the claude alias with actual environment variables
		cmd = exec.Command("claude")
		cmd.Env = append(os.Environ(),
			"ANTHROPIC_BASE_URL=https://open.bigmodel.cn/api/anthropic",
			"ANTHROPIC_API_KEY=REDACTED_API_KEY",
			"ANTHROPIC_MODEL=glm-4.6",
		)
	case "claude":
		// Expand the claude alias with actual environment variables
		cmd = exec.Command("claude")
		cmd.Env = append(os.Environ(),
			"ANTHROPIC_BASE_URL=https://open.bigmodel.cn/api/anthropic",
			"ANTHROPIC_API_KEY=REDACTED_API_KEY",
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
			return fmt.Errorf("failed to stage changes: %w", err)
		}
		cmd = exec.Command("git", "commit", "-m", commitMessage)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

func installBinary(c *cli.Context) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
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
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if closeErr := source.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close source file: %v\n", closeErr)
		}
	}()

	destination, err := os.Create(installPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if closeErr := destination.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close destination file: %v\n", closeErr)
		}
	}()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Make it executable (0755 is appropriate for system binaries in /usr/local/bin)
	// This allows read and execute by all users, but write only by owner
	//nolint:gosec // G302: 0755 is appropriate for system binaries
	err = os.Chmod(installPath, 0o755)
	if err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	fmt.Printf("Commitgen installed successfully to %s\n", installPath)
	fmt.Println("You can now use 'commitgen' from anywhere!")
	return nil
}
