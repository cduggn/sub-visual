package repocheck

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cduggn/sub-visual/internal/github"
)

// RepoMetadata holds the fields we extract from the repos/ endpoint.
type RepoMetadata struct {
	Stars       int      `json:"stargazers_count"`
	Forks       int      `json:"forks_count"`
	Watchers    int      `json:"subscribers_count"`
	OpenIssues  int      `json:"open_issues_count"`
	Archived    bool     `json:"archived"`
	CreatedAt   string   `json:"created_at"`
	PushedAt    string   `json:"pushed_at"`
	Description string   `json:"description"`
	Topics      []string `json:"topics"`
	License     *struct {
		Name   string `json:"name"`
		SPDXID string `json:"spdx_id"`
	} `json:"license"`
}

// FetchRepoMetadata fetches core repository metadata via the GitHub API.
func FetchRepoMetadata(owner, repo string) (*RepoMetadata, error) {
	data, err := github.GhAPI(fmt.Sprintf("repos/%s/%s", owner, repo))
	if err != nil {
		return nil, err
	}
	var meta RepoMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parsing repo metadata: %w", err)
	}
	return &meta, nil
}

// FetchCommitActivity returns commit count for the last 90 days using the
// stats/participation endpoint. Returns -1 if data is unavailable.
func FetchCommitActivity(owner, repo string) int {
	endpoint := fmt.Sprintf("repos/%s/%s/stats/participation", owner, repo)
	data, err := github.GhAPI(endpoint)
	if err != nil {
		// GitHub returns 202 when stats are being computed; retry once.
		if strings.Contains(err.Error(), "202") || strings.Contains(err.Error(), "Accepted") {
			time.Sleep(2 * time.Second)
			data, err = github.GhAPI(endpoint)
			if err != nil {
				return -1
			}
		} else {
			return -1
		}
	}

	var stats struct {
		All []int `json:"all"`
	}
	if err := json.Unmarshal(data, &stats); err != nil {
		return -1
	}

	// Last 13 weeks ≈ 90 days
	total := 0
	start := len(stats.All) - 13
	if start < 0 {
		start = 0
	}
	for _, c := range stats.All[start:] {
		total += c
	}
	return total
}

// FetchContributors returns the contributor count by parsing the Link header.
func FetchContributors(owner, repo string) int {
	cmd := exec.Command("gh", "api", "-i",
		fmt.Sprintf("repos/%s/%s/contributors?per_page=1&anon=true", owner, repo))
	out, err := cmd.Output()
	if err != nil {
		return -1
	}

	output := string(out)

	// Parse Link header for last page number
	linkRe := regexp.MustCompile(`page=(\d+)>; rel="last"`)
	if m := linkRe.FindStringSubmatch(output); m != nil {
		count, _ := strconv.Atoi(m[1])
		return count
	}

	// No Link header means single page; count the JSON array elements
	bodyIdx := strings.Index(output, "\n\n")
	if bodyIdx < 0 {
		bodyIdx = strings.Index(output, "\r\n\r\n")
	}
	if bodyIdx >= 0 {
		body := strings.TrimSpace(output[bodyIdx:])
		var arr []json.RawMessage
		if json.Unmarshal([]byte(body), &arr) == nil {
			return len(arr)
		}
	}

	return 1
}

// FetchOpenPRs returns the count of open pull requests.
func FetchOpenPRs(owner, repo string) int {
	data, err := github.GhAPI(fmt.Sprintf("search/issues?q=repo:%s/%s+type:pr+state:open&per_page=1", owner, repo))
	if err != nil {
		return -1
	}
	var result struct {
		TotalCount int `json:"total_count"`
	}
	if json.Unmarshal(data, &result) != nil {
		return -1
	}
	return result.TotalCount
}

// ReleaseInfo holds release count and latest release date.
type ReleaseInfo struct {
	Count  int
	Latest string
}

// FetchReleases returns release information for the repository.
func FetchReleases(owner, repo string) ReleaseInfo {
	data, err := github.GhAPI(fmt.Sprintf("repos/%s/%s/releases?per_page=100", owner, repo))
	if err != nil {
		return ReleaseInfo{}
	}

	var releases []struct {
		PublishedAt string `json:"published_at"`
	}
	if json.Unmarshal(data, &releases) != nil {
		return ReleaseInfo{}
	}

	info := ReleaseInfo{Count: len(releases)}
	if len(releases) > 0 {
		info.Latest = releases[0].PublishedAt
	}
	return info
}

// CommunityProfile holds community health indicators.
type CommunityProfile struct {
	HasReadme        bool
	HasContributing  bool
	HasCodeOfConduct bool
	HasCI            bool
}

// FetchCommunityProfile returns community health indicators for the repository.
func FetchCommunityProfile(owner, repo string) CommunityProfile {
	data, err := github.GhAPI(fmt.Sprintf("repos/%s/%s/community/profile", owner, repo))
	if err != nil {
		return CommunityProfile{}
	}

	var raw struct {
		Files struct {
			Readme        interface{} `json:"readme"`
			Contributing  interface{} `json:"contributing"`
			CodeOfConduct interface{} `json:"code_of_conduct"`
		} `json:"files"`
	}
	if json.Unmarshal(data, &raw) != nil {
		return CommunityProfile{}
	}

	profile := CommunityProfile{
		HasReadme:        raw.Files.Readme != nil,
		HasContributing:  raw.Files.Contributing != nil,
		HasCodeOfConduct: raw.Files.CodeOfConduct != nil,
	}

	// Check for CI by looking at workflows
	wfData, err := github.GhAPI(fmt.Sprintf("repos/%s/%s/actions/workflows?per_page=1", owner, repo))
	if err == nil {
		var wf struct {
			TotalCount int `json:"total_count"`
		}
		if json.Unmarshal(wfData, &wf) == nil && wf.TotalCount > 0 {
			profile.HasCI = true
		}
	}

	return profile
}

// DaysSince computes the number of days between a date string and now.
func DaysSince(dateStr string) int {
	layouts := []string{time.RFC3339, "2006-01-02T15:04:05Z"}
	for _, layout := range layouts {
		t, err := time.Parse(layout, dateStr)
		if err == nil {
			days := time.Since(t).Hours() / 24
			return int(math.Round(days))
		}
	}
	return -1
}
