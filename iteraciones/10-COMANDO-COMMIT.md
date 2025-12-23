# Iteracion 10: Comando Commit

## Objetivos

- Implementar comando `goreview commit`
- Generar mensajes de commit con IA
- Soportar Conventional Commits
- Integracion con git commit

## Tiempo Estimado: 4 horas

---

## Commit 10.1: Crear estructura del comando commit

**Mensaje de commit:**
```
feat(cli): add commit command structure

- Add commit.go command file
- Define flags for commit generation
- Support interactive and non-interactive modes
```

### `goreview/cmd/commit.go`

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate AI-powered commit messages",
	Long: `Generate commit messages using AI analysis of staged changes.

Examples:
  # Generate and show commit message
  goreview commit

  # Generate and commit directly
  goreview commit --execute

  # Generate with specific type
  goreview commit --type feat

  # Amend last commit with new message
  goreview commit --amend`,
	RunE: runCommit,
}

func init() {
	rootCmd.AddCommand(commitCmd)

	// Execution flags
	commitCmd.Flags().BoolP("execute", "e", false, "Execute git commit with generated message")
	commitCmd.Flags().Bool("amend", false, "Amend the last commit")

	// Message customization
	commitCmd.Flags().StringP("type", "t", "", "Force commit type (feat, fix, docs, etc.)")
	commitCmd.Flags().StringP("scope", "s", "", "Force commit scope")
	commitCmd.Flags().Bool("breaking", false, "Mark as breaking change")
	commitCmd.Flags().StringP("body", "b", "", "Additional commit body")
	commitCmd.Flags().String("footer", "", "Commit footer (e.g., 'Closes #123')")

	// Output flags
	commitCmd.Flags().Bool("dry-run", false, "Show message without committing")
	commitCmd.Flags().BoolP("verbose", "v", false, "Show detailed analysis")
}

func runCommit(cmd *cobra.Command, args []string) error {
	fmt.Println("Commit command - implementation follows")
	return nil
}
```

---

## Commit 10.2: Implementar generacion de mensaje

**Mensaje de commit:**
```
feat(cli): implement commit message generation

- Analyze staged diff
- Generate Conventional Commit message
- Support type/scope overrides
- Handle breaking changes
```

### Actualizar `goreview/cmd/commit.go`:

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/git"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
)

func runCommit(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize git repo
	gitRepo, err := git.NewRepository(".")
	if err != nil {
		return fmt.Errorf("initializing git: %w", err)
	}

	// Get staged diff
	diff, err := gitRepo.GetStagedDiff(ctx)
	if err != nil {
		return fmt.Errorf("getting staged diff: %w", err)
	}

	if len(diff.Files) == 0 {
		return fmt.Errorf("no staged changes found. Stage changes with 'git add' first")
	}

	// Initialize provider
	provider, err := providers.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("initializing provider: %w", err)
	}
	defer provider.Close()

	// Generate commit message
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Fprintf(os.Stderr, "Analyzing %d files...\n", len(diff.Files))
	}

	diffText := formatDiffForCommit(diff)
	message, err := provider.GenerateCommitMessage(ctx, diffText)
	if err != nil {
		return fmt.Errorf("generating commit message: %w", err)
	}

	// Apply overrides
	message = applyCommitOverrides(cmd, message)

	// Add body and footer
	body, _ := cmd.Flags().GetString("body")
	footer, _ := cmd.Flags().GetString("footer")
	if body != "" || footer != "" {
		message = buildFullMessage(message, body, footer)
	}

	// Dry run - just show message
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Println("Generated commit message:")
		fmt.Println("─────────────────────────")
		fmt.Println(message)
		fmt.Println("─────────────────────────")
		return nil
	}

	// Execute commit
	execute, _ := cmd.Flags().GetBool("execute")
	amend, _ := cmd.Flags().GetBool("amend")

	if execute || amend {
		return executeGitCommit(message, amend)
	}

	// Default: print message for user to copy
	fmt.Println(message)
	return nil
}

func formatDiffForCommit(diff *git.Diff) string {
	var sb strings.Builder
	for _, file := range diff.Files {
		sb.WriteString(fmt.Sprintf("File: %s (%s)\n", file.Path, file.Status))
		for _, hunk := range file.Hunks {
			sb.WriteString(hunk.Header + "\n")
			for _, line := range hunk.Lines {
				prefix := " "
				if line.Type == git.LineAddition {
					prefix = "+"
				} else if line.Type == git.LineDeletion {
					prefix = "-"
				}
				sb.WriteString(prefix + line.Content + "\n")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func applyCommitOverrides(cmd *cobra.Command, message string) string {
	commitType, _ := cmd.Flags().GetString("type")
	scope, _ := cmd.Flags().GetString("scope")
	breaking, _ := cmd.Flags().GetBool("breaking")

	// Parse existing message
	parts := parseConventionalCommit(message)

	// Apply overrides
	if commitType != "" {
		parts.Type = commitType
	}
	if scope != "" {
		parts.Scope = scope
	}
	if breaking {
		parts.Breaking = true
	}

	// Rebuild message
	return parts.String()
}

type conventionalParts struct {
	Type        string
	Scope       string
	Breaking    bool
	Description string
}

func parseConventionalCommit(message string) *conventionalParts {
	parts := &conventionalParts{}

	// Simple parsing - find type(scope): description
	line := strings.Split(message, "\n")[0]

	if idx := strings.Index(line, ":"); idx > 0 {
		prefix := line[:idx]
		parts.Description = strings.TrimSpace(line[idx+1:])

		if strings.HasSuffix(prefix, "!") {
			parts.Breaking = true
			prefix = prefix[:len(prefix)-1]
		}

		if paren := strings.Index(prefix, "("); paren > 0 {
			parts.Type = prefix[:paren]
			parts.Scope = strings.Trim(prefix[paren:], "()")
		} else {
			parts.Type = prefix
		}
	} else {
		parts.Type = "chore"
		parts.Description = message
	}

	return parts
}

func (p *conventionalParts) String() string {
	var sb strings.Builder
	sb.WriteString(p.Type)
	if p.Scope != "" {
		sb.WriteString("(" + p.Scope + ")")
	}
	if p.Breaking {
		sb.WriteString("!")
	}
	sb.WriteString(": ")
	sb.WriteString(p.Description)
	return sb.String()
}

func buildFullMessage(subject, body, footer string) string {
	var parts []string
	parts = append(parts, subject)
	if body != "" {
		parts = append(parts, "", body)
	}
	if footer != "" {
		parts = append(parts, "", footer)
	}
	return strings.Join(parts, "\n")
}

func executeGitCommit(message string, amend bool) error {
	args := []string{"commit", "-m", message}
	if amend {
		args = append(args, "--amend")
	}

	gitCmd := exec.Command("git", args...)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	gitCmd.Stdin = os.Stdin

	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Commit created successfully\n")
	return nil
}
```

---

## Commit 10.3: Agregar modo interactivo

**Mensaje de commit:**
```
feat(cli): add interactive commit mode

- Allow editing message before commit
- Show file summary
- Confirm before committing
```

### `goreview/cmd/commit_interactive.go`

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/git"
)

// InteractiveCommit handles the interactive commit flow.
func InteractiveCommit(diff *git.Diff, generatedMessage string) (string, bool, error) {
	// Show summary
	fmt.Println("\n┌─────────────────────────────────────┐")
	fmt.Println("│         Commit Summary              │")
	fmt.Println("├─────────────────────────────────────┤")

	for _, file := range diff.Files {
		icon := getStatusIcon(file.Status)
		fmt.Printf("│ %s %-33s│\n", icon, truncate(file.Path, 33))
	}

	fmt.Println("├─────────────────────────────────────┤")
	fmt.Printf("│ Total: %d files changed             │\n", len(diff.Files))
	fmt.Println("└─────────────────────────────────────┘")

	// Show generated message
	fmt.Println("\nGenerated commit message:")
	fmt.Println("─────────────────────────────────────")
	fmt.Println(generatedMessage)
	fmt.Println("─────────────────────────────────────")

	// Ask for action
	fmt.Println("\nOptions:")
	fmt.Println("  [c] Commit with this message")
	fmt.Println("  [e] Edit message")
	fmt.Println("  [r] Regenerate message")
	fmt.Println("  [q] Quit without committing")
	fmt.Print("\nChoice [c/e/r/q]: ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(strings.ToLower(choice))

	switch choice {
	case "c", "":
		return generatedMessage, true, nil
	case "e":
		edited, err := editMessage(generatedMessage)
		if err != nil {
			return "", false, err
		}
		return edited, true, nil
	case "r":
		return "", false, fmt.Errorf("regenerate requested")
	case "q":
		return "", false, nil
	default:
		return "", false, fmt.Errorf("invalid choice: %s", choice)
	}
}

func getStatusIcon(status git.FileStatus) string {
	switch status {
	case git.FileAdded:
		return "+"
	case git.FileModified:
		return "~"
	case git.FileDeleted:
		return "-"
	case git.FileRenamed:
		return ">"
	default:
		return "?"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func editMessage(message string) (string, error) {
	fmt.Println("\nEnter new commit message (empty line to finish):")
	fmt.Println("Current message shown below. Press Enter to keep or type new message.")
	fmt.Println("─────────────────────────────────────")
	fmt.Println(message)
	fmt.Println("─────────────────────────────────────")
	fmt.Print("New message (or press Enter to keep): ")

	reader := bufio.NewReader(os.Stdin)
	newMessage, _ := reader.ReadString('\n')
	newMessage = strings.TrimSpace(newMessage)

	if newMessage == "" {
		return message, nil
	}
	return newMessage, nil
}
```

---

## Commit 10.4: Tests del comando commit

**Mensaje de commit:**
```
test(cli): add commit command tests

- Test message parsing
- Test conventional commit formatting
- Test override application
```

### `goreview/cmd/commit_test.go`

```go
package cmd

import (
	"testing"
)

func TestParseConventionalCommit(t *testing.T) {
	tests := []struct {
		input    string
		wantType string
		wantScope string
		wantDesc string
		wantBreaking bool
	}{
		{
			input:    "feat(auth): add login endpoint",
			wantType: "feat",
			wantScope: "auth",
			wantDesc: "add login endpoint",
		},
		{
			input:    "fix: resolve memory leak",
			wantType: "fix",
			wantScope: "",
			wantDesc: "resolve memory leak",
		},
		{
			input:    "feat(api)!: change response format",
			wantType: "feat",
			wantScope: "api",
			wantDesc: "change response format",
			wantBreaking: true,
		},
		{
			input:    "chore: update dependencies",
			wantType: "chore",
			wantScope: "",
			wantDesc: "update dependencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parts := parseConventionalCommit(tt.input)

			if parts.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", parts.Type, tt.wantType)
			}
			if parts.Scope != tt.wantScope {
				t.Errorf("Scope = %q, want %q", parts.Scope, tt.wantScope)
			}
			if parts.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", parts.Description, tt.wantDesc)
			}
			if parts.Breaking != tt.wantBreaking {
				t.Errorf("Breaking = %v, want %v", parts.Breaking, tt.wantBreaking)
			}
		})
	}
}

func TestConventionalPartsString(t *testing.T) {
	tests := []struct {
		parts *conventionalParts
		want  string
	}{
		{
			parts: &conventionalParts{Type: "feat", Description: "add feature"},
			want:  "feat: add feature",
		},
		{
			parts: &conventionalParts{Type: "fix", Scope: "api", Description: "fix bug"},
			want:  "fix(api): fix bug",
		},
		{
			parts: &conventionalParts{Type: "feat", Scope: "core", Breaking: true, Description: "breaking change"},
			want:  "feat(core)!: breaking change",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.parts.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildFullMessage(t *testing.T) {
	tests := []struct {
		subject string
		body    string
		footer  string
		want    string
	}{
		{
			subject: "feat: add feature",
			want:    "feat: add feature",
		},
		{
			subject: "feat: add feature",
			body:    "This is the body",
			want:    "feat: add feature\n\nThis is the body",
		},
		{
			subject: "fix: fix bug",
			body:    "Detailed description",
			footer:  "Closes #123",
			want:    "fix: fix bug\n\nDetailed description\n\nCloses #123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			got := buildFullMessage(tt.subject, tt.body, tt.footer)
			if got != tt.want {
				t.Errorf("buildFullMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exactly10!", 10, "exactly10!"},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := truncate(tt.s, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
			}
		})
	}
}
```

---

## Resumen de la Iteracion 10

### Commits:
1. `feat(cli): add commit command structure`
2. `feat(cli): implement commit message generation`
3. `feat(cli): add interactive commit mode`
4. `test(cli): add commit command tests`

### Archivos:
```
goreview/cmd/
├── commit.go
├── commit_interactive.go
└── commit_test.go
```

---

## Siguiente Iteracion

Continua con: **[11-COMANDO-DOC.md](11-COMANDO-DOC.md)**
