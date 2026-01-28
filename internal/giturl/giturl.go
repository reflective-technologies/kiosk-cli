package giturl

import "strings"

// ExtractOrgRepo extracts org/repo from supported Git URLs.
func ExtractOrgRepo(gitURL string) string {
	gitURL = strings.TrimSuffix(gitURL, ".git")
	for _, prefix := range []string{
		"https://github.com/",
		"https://gitlab.com/",
		"https://bitbucket.org/",
		"git@github.com:",
		"git@gitlab.com:",
		"git@bitbucket.org:",
	} {
		if strings.HasPrefix(gitURL, prefix) {
			return strings.TrimPrefix(gitURL, prefix)
		}
	}
	return ""
}
