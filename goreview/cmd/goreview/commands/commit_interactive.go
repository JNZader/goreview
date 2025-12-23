package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/JNZader/goreview/goreview/internal/git"
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
		edited := editMessage(generatedMessage)
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

func editMessage(message string) string {
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
		return message
	}
	return newMessage
}
