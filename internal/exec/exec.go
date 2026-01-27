package exec

import (
	"os"
	"os/exec"
)

// AuditPrompt is the prompt used for security audits before publishing.
const AuditPrompt = `You are auditing this application for security issues before it is published to a public repository.

Perform the following checks:

1. **Current codebase scan**: Search all files for:
   - API keys, secrets, tokens, or credentials (look for patterns like API_KEY, SECRET, TOKEN, PASSWORD, etc.)
   - Personal information (emails, phone numbers, addresses)
   - Hardcoded URLs with embedded credentials
   - Private keys or certificates
   - Environment files (.env) that shouldn't be committed

2. **Git history scan**: Check the git history for any previously committed secrets that may have been removed:
   - Run: git log -p --all -S 'API_KEY|SECRET|TOKEN|PASSWORD|PRIVATE_KEY' --pickaxe-regex
   - Also check: git log -p --all -- '*.env' '.env*'
   - Look for any commits that added then removed sensitive data

3. **Configuration review**: Check for:
   - Proper .gitignore entries for sensitive files
   - Any configuration files that might contain secrets

Report your findings clearly, listing:
- Any issues found with file paths and line numbers
- Severity (critical/warning/info)
- Recommended remediation steps

If no issues are found, confirm the repository appears safe for publication.

IMPORTANT: 
- Output ONLY the markdown report. No preamble, no explanations, no follow-up questions—just the report itself.
- Format your response as valid markdown with proper headers, lists, and code blocks where appropriate.`

// ClaudeCmd builds an exec.Cmd for running claude with the given args.
// It falls back to running through the user's shell if claude is not in PATH.
func ClaudeCmd(args ...string) *exec.Cmd {
	if _, err := exec.LookPath("claude"); err == nil {
		return exec.Command("claude", args...)
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}

	// Build shell command with positional args
	// Format: shell -i -c 'claude "$@"' claude arg1 arg2 ...
	shellArgs := []string{"-i", "-c", `claude "$@"`, "claude"}
	shellArgs = append(shellArgs, args...)
	return exec.Command(shell, shellArgs...)
}
