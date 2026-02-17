package repocheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"
)

// scorecardAPIResponse matches the JSON from api.securityscorecards.dev
type scorecardAPIResponse struct {
	Score  float64 `json:"score"`
	Checks []struct {
		Name  string  `json:"name"`
		Score float64 `json:"score"`
	} `json:"checks"`
}

// fetchScorecardAPI tries the public OpenSSF Scorecard REST API first.
func fetchScorecardAPI(owner, repo string) (*ScorecardResult, error) {
	url := fmt.Sprintf("https://api.securityscorecards.dev/projects/github.com/%s/%s", owner, repo)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("scorecard API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repo not tracked by OpenSSF Scorecard")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scorecard API returned status %d", resp.StatusCode)
	}

	var raw scorecardAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("scorecard API parse error: %w", err)
	}

	return parseScorecardResponse(&raw), nil
}

// runScorecardCLI falls back to the local scorecard CLI.
func runScorecardCLI(owner, repo string) (*ScorecardResult, error) {
	if _, err := exec.LookPath("scorecard"); err != nil {
		return nil, fmt.Errorf("scorecard CLI not installed (install: go install github.com/ossf/scorecard/v5/cmd/scorecard@latest)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "scorecard",
		fmt.Sprintf("--repo=github.com/%s/%s", owner, repo),
		"--format=json",
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("scorecard command failed: %w", err)
	}

	var raw scorecardAPIResponse
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("scorecard output parse error: %w", err)
	}

	return parseScorecardResponse(&raw), nil
}

func parseScorecardResponse(raw *scorecardAPIResponse) *ScorecardResult {
	result := &ScorecardResult{
		Score:  raw.Score,
		Checks: make(map[string]float64),
	}
	for _, c := range raw.Checks {
		result.Checks[c.Name] = c.Score
	}
	return result
}

// RunScorecard tries the public API first, then falls back to the CLI.
// Returns (result, reason). Result is nil if both fail; reason explains why.
func RunScorecard(owner, repo string, logWriter io.Writer) (*ScorecardResult, string) {
	// Try the public REST API first (covers top ~1M repos, no install needed)
	result, err := fetchScorecardAPI(owner, repo)
	if err == nil {
		return result, ""
	}
	apiReason := err.Error()
	fmt.Fprintf(logWriter, "scorecard API: %v — trying CLI fallback\n", err)

	// Fall back to local CLI
	result, err = runScorecardCLI(owner, repo)
	if err == nil {
		return result, ""
	}
	cliReason := err.Error()
	fmt.Fprintf(logWriter, "scorecard CLI: %v\n", err)

	return nil, fmt.Sprintf("API: %s; CLI: %s", apiReason, cliReason)
}
