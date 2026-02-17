package github

import (
	"fmt"
	"os/exec"
)

// GhAPI runs `gh api <endpoint>` and returns the raw JSON output.
func GhAPI(endpoint string) ([]byte, error) {
	cmd := exec.Command("gh", "api", endpoint)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh api %s failed: %s", endpoint, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("gh api %s: %w", endpoint, err)
	}
	return out, nil
}

// CheckCLI verifies that the gh CLI is installed and available on PATH.
func CheckCLI() error {
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found. Install from https://cli.github.com/")
	}
	return nil
}
