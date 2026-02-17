package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// fileContent represents the GitHub API response for a file's content.
type fileContent struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// FetchFileContent downloads a file from a GitHub repository via the API.
func FetchFileContent(owner, repo, branch, path string) (string, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", owner, repo, path, branch)
	data, err := GhAPI(endpoint)
	if err != nil {
		return "", fmt.Errorf("fetching file content: %w", err)
	}

	var fc fileContent
	if err := json.Unmarshal(data, &fc); err != nil {
		return "", fmt.Errorf("parsing file content response: %w", err)
	}

	if fc.Encoding != "base64" {
		return "", fmt.Errorf("unexpected encoding %q (expected base64)", fc.Encoding)
	}

	decoded, err := base64.StdEncoding.DecodeString(fc.Content)
	if err != nil {
		return "", fmt.Errorf("decoding base64 content: %w", err)
	}

	return string(decoded), nil
}
