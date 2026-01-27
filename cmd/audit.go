package cmd

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/glamour"
	kioskexec "github.com/reflective-technologies/kiosk-cli/internal/exec"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit the current directory for security issues before publishing",
	Long: `Run a security audit on the current directory to check for:
- API keys, secrets, and credentials in the codebase
- Personal information that shouldn't be published
- Git history containing previously committed secrets

This command runs Claude with an audit-focused prompt and prints the results.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		return execClaudeAudit(cwd, kioskexec.AuditPrompt)
	},
}

func execClaudeAudit(dir, prompt string) error {
	cmd := kioskexec.ClaudeCmd("-p", prompt)
	cmd.Dir = dir

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	if term.IsTerminal(int(os.Stdout.Fd())) {
		dots := []string{".  ", ".. ", "..."}
		i := 0
		ticker := time.NewTicker(400 * time.Millisecond)
		defer ticker.Stop()

		fmt.Print("Running security audit.")
	loop:
		for {
			select {
			case err := <-done:
				fmt.Print("\r\033[K")
				if err != nil {
					return err
				}
				break loop
			case <-ticker.C:
				i = (i + 1) % len(dots)
				fmt.Printf("\rRunning security audit%s", dots[i])
			}
		}
	} else {
		if err := <-done; err != nil {
			return err
		}
	}

	output := stdout.String()

	fmt.Println("Audit results:")
	fmt.Println()

	if term.IsTerminal(int(os.Stdout.Fd())) {
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(80),
		)
		if err == nil {
			rendered, err := renderer.Render(output)
			if err == nil {
				fmt.Print(rendered)
				return nil
			}
		}
	}

	fmt.Print(output)
	return nil
}

func init() {
	rootCmd.AddCommand(auditCmd)
}
