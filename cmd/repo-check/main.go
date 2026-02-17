package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/cduggn/sub-visual/internal/github"
	"github.com/cduggn/sub-visual/internal/repocheck"
)

// logWriter is where non-fatal diagnostic messages go (stderr).
var logWriter io.Writer = os.Stderr

func main() {
	jsonFlag := flag.Bool("json", false, "output as JSON instead of markdown")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: repo-check [--json] <target>\n\n")
		fmt.Fprintf(os.Stderr, "  target: owner/repo, GitHub URL, or SKILL.md URL\n")
		fmt.Fprintf(os.Stderr, "  --json: output as JSON instead of markdown\n")
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Check that gh CLI is available
	if err := github.CheckCLI(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}

	target := flag.Arg(0)
	gt, err := github.ParseTarget(target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	owner, repo := gt.Owner, gt.Repo

	report := &repocheck.TrustReport{
		Owner: owner,
		Repo:  repo,
		URL:   fmt.Sprintf("https://github.com/%s/%s", owner, repo),
	}

	// Fetch repo metadata (required)
	meta, err := repocheck.FetchRepoMetadata(owner, repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not fetch repository — %v\n", err)
		os.Exit(3)
	}

	// Populate popularity
	report.Popularity = repocheck.PopularityMetrics{
		Stars:    meta.Stars,
		Forks:    meta.Forks,
		Watchers: meta.Watchers,
	}

	// Populate activity
	commits90d := repocheck.FetchCommitActivity(owner, repo)
	if commits90d < 0 {
		commits90d = 0
		fmt.Fprintln(logWriter, "Warning: could not fetch commit activity stats")
	}

	lastCommitDate := meta.PushedAt
	daysSinceCommit := repocheck.DaysSince(lastCommitDate)
	if daysSinceCommit < 0 {
		daysSinceCommit = 0
	}

	repoAge := repocheck.DaysSince(meta.CreatedAt)
	if repoAge < 0 {
		repoAge = 0
	}

	contributors := repocheck.FetchContributors(owner, repo)
	if contributors < 0 {
		contributors = 0
		fmt.Fprintln(logWriter, "Warning: could not fetch contributor count")
	}

	openPRs := repocheck.FetchOpenPRs(owner, repo)
	if openPRs < 0 {
		openPRs = 0
		fmt.Fprintln(logWriter, "Warning: could not fetch open PR count")
	}

	report.Activity = repocheck.ActivityMetrics{
		LastCommitDate:      lastCommitDate,
		DaysSinceLastCommit: daysSinceCommit,
		CommitsLast90d:      commits90d,
		OpenIssues:          meta.OpenIssues,
		OpenPRs:             openPRs,
		Contributors:        contributors,
		RepoAgeDays:         repoAge,
		Archived:            meta.Archived,
	}

	// Populate security
	report.Security = repocheck.SecurityMetrics{
		ScorecardChecks: make(map[string]float64),
	}
	if meta.License != nil {
		report.Security.HasLicense = true
		report.Security.LicenseName = meta.License.Name
		report.Security.LicenseSPDX = meta.License.SPDXID
	}

	sc, scReason := repocheck.RunScorecard(owner, repo, logWriter)
	if sc != nil {
		report.Security.ScorecardAvailable = true
		report.Security.ScorecardScore = sc.Score
		report.Security.ScorecardChecks = sc.Checks
	} else {
		report.Security.ScorecardReason = scReason
	}

	// Populate maturity
	releases := repocheck.FetchReleases(owner, repo)
	community := repocheck.FetchCommunityProfile(owner, repo)

	report.Maturity = repocheck.MaturityMetrics{
		ReleasesCount:    releases.Count,
		LatestRelease:    releases.Latest,
		HasReadme:        community.HasReadme,
		HasContributing:  community.HasContributing,
		HasCodeOfConduct: community.HasCodeOfConduct,
		HasCI:            community.HasCI,
		HasDescription:   meta.Description != "",
		TopicsCount:      len(meta.Topics),
	}

	// Score all categories
	repocheck.ScorePopularity(&report.Popularity)
	repocheck.ScoreActivity(&report.Activity)
	repocheck.ScoreSecurity(&report.Security)
	repocheck.ScoreMaturity(&report.Maturity)

	// Composite signal and warnings
	report.TrustSignal = repocheck.ComputeComposite(report)
	report.Warnings = repocheck.CollectWarnings(report)

	// Output
	if *jsonFlag {
		out, err := repocheck.FormatJSON(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)
	} else {
		fmt.Print(repocheck.FormatMarkdown(report))
	}
}
