package github

import (
	"fmt"
	"regexp"
)

// GitHubTarget represents a parsed GitHub URL or shorthand target.
type GitHubTarget struct {
	Owner  string
	Repo   string
	Branch string // set for blob URLs
	Path   string // set for blob URLs
	IsBlob bool
}

var (
	// owner/repo shorthand
	shorthandRe = regexp.MustCompile(`^([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)$`)

	// https://github.com/owner/repo (with optional trailing slash or .git)
	repoURLRe = regexp.MustCompile(`^https?://github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+?)(?:\.git)?/?$`)

	// https://github.com/owner/repo/blob/branch/path...
	blobURLRe = regexp.MustCompile(`^https?://github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)/blob/([^/]+)/(.+)$`)

	// https://github.com/owner/repo/tree/branch/path... (also extract owner/repo)
	treeURLRe = regexp.MustCompile(`^https?://github\.com/([a-zA-Z0-9_.-]+)/([a-zA-Z0-9_.-]+)/tree/`)
)

// ParseTarget extracts owner, repo, and optionally branch+path from supported input formats.
func ParseTarget(target string) (*GitHubTarget, error) {
	// Try blob URL first (most specific)
	if m := blobURLRe.FindStringSubmatch(target); m != nil {
		return &GitHubTarget{
			Owner:  m[1],
			Repo:   m[2],
			Branch: m[3],
			Path:   m[4],
			IsBlob: true,
		}, nil
	}

	// Tree URL (extract owner/repo only)
	if m := treeURLRe.FindStringSubmatch(target); m != nil {
		return &GitHubTarget{
			Owner: m[1],
			Repo:  m[2],
		}, nil
	}

	// Plain repo URL
	if m := repoURLRe.FindStringSubmatch(target); m != nil {
		return &GitHubTarget{
			Owner: m[1],
			Repo:  m[2],
		}, nil
	}

	// Shorthand
	if m := shorthandRe.FindStringSubmatch(target); m != nil {
		return &GitHubTarget{
			Owner: m[1],
			Repo:  m[2],
		}, nil
	}

	return nil, fmt.Errorf("unrecognized target format: %s\nExpected: owner/repo, https://github.com/owner/repo, or a GitHub blob URL", target)
}

// IsGitHubURL returns true if the target string looks like a GitHub URL.
func IsGitHubURL(target string) bool {
	return blobURLRe.MatchString(target) || repoURLRe.MatchString(target) || treeURLRe.MatchString(target)
}
