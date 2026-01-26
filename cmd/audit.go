package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

const auditPrompt = `You are auditing this application for security issues before it is published to a public repository.

Perform the following checks:

1. **Current codebase scan**: Search all files for:
   - API keys, secrets, tokens, or credentials (look for patterns like API_KEY, SECRET, TOKEN, PASSWORD, etc.)
   - Personal information (emails, phone numbers, addresses)
   - Hardcoded URLs with embedded credentials
   - Private keys or certificates
   - Environment files (.env) that shouldn't be committed

2. **Git history scan**: Check the git history for any previously committed secrets that may have been removed:
   - Run: git log -p --all -S 'API_KEY\|SECRET\|TOKEN\|PASSWORD\|PRIVATE_KEY' --pickaxe-regex
   - Also check: git log -p --all -- '*.env' '.env*'
   - Look for any commits that added then removed sensitive data

3. **Configuration review**: Check for:
   - Proper .gitignore entries for sensitive files
   - Any configuration files that might contain secrets

Report your findings clearly, listing:
- Any issues found with file paths and line numbers
- Severity (critical/warning/info)
- Recommended remediation steps

If no issues are found, confirm the repository appears safe for publication.`

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

		return execClaudeAudit(cwd, auditPrompt)
	},
}

func execClaudeAudit(dir, prompt string) error {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH: %w", err)
	}

	cmd := exec.Command(claudePath, "-p", prompt)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(auditCmd)
}
