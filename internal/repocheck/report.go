package repocheck

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatJSON returns the report as indented JSON.
func FormatJSON(report *TrustReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatMarkdown returns the report as a markdown string.
func FormatMarkdown(report *TrustReport) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Repository Trust Report: %s/%s\n\n", report.Owner, report.Repo))
	b.WriteString(fmt.Sprintf("**URL:** https://github.com/%s/%s\n\n", report.Owner, report.Repo))
	b.WriteString(fmt.Sprintf("## Overall Trust Signal: %s\n\n", report.TrustSignal))

	// Warnings
	if len(report.Warnings) > 0 {
		b.WriteString("### Warnings\n\n")
		for _, w := range report.Warnings {
			b.WriteString(fmt.Sprintf("- %s\n", w))
		}
		b.WriteString("\n")
	}

	// Popularity
	b.WriteString(fmt.Sprintf("## Popularity — %s\n\n", report.Popularity.Score))
	b.WriteString("| Metric | Value |\n")
	b.WriteString("|--------|-------|\n")
	b.WriteString(fmt.Sprintf("| Stars | %d |\n", report.Popularity.Stars))
	b.WriteString(fmt.Sprintf("| Forks | %d |\n", report.Popularity.Forks))
	b.WriteString(fmt.Sprintf("| Watchers | %d |\n", report.Popularity.Watchers))
	b.WriteString("\n")

	// Activity
	b.WriteString(fmt.Sprintf("## Activity — %s\n\n", report.Activity.Score))
	b.WriteString("| Metric | Value |\n")
	b.WriteString("|--------|-------|\n")
	b.WriteString(fmt.Sprintf("| Last commit | %s (%d days ago) |\n", report.Activity.LastCommitDate, report.Activity.DaysSinceLastCommit))
	b.WriteString(fmt.Sprintf("| Commits (90d) | %d |\n", report.Activity.CommitsLast90d))
	b.WriteString(fmt.Sprintf("| Open issues | %d |\n", report.Activity.OpenIssues))
	b.WriteString(fmt.Sprintf("| Open PRs | %d |\n", report.Activity.OpenPRs))
	b.WriteString(fmt.Sprintf("| Contributors | %d |\n", report.Activity.Contributors))
	b.WriteString(fmt.Sprintf("| Repo age | %d days |\n", report.Activity.RepoAgeDays))
	b.WriteString(fmt.Sprintf("| Archived | %v |\n", report.Activity.Archived))
	b.WriteString("\n")

	// Security
	b.WriteString(fmt.Sprintf("## Security — %s\n\n", report.Security.Score))
	b.WriteString("| Metric | Value |\n")
	b.WriteString("|--------|-------|\n")
	if report.Security.HasLicense {
		b.WriteString(fmt.Sprintf("| License | %s (%s) |\n", report.Security.LicenseName, report.Security.LicenseSPDX))
	} else {
		b.WriteString("| License | None detected |\n")
	}
	if report.Security.ScorecardAvailable {
		b.WriteString(fmt.Sprintf("| OpenSSF Scorecard | %.1f/10 |\n", report.Security.ScorecardScore))
		if len(report.Security.ScorecardChecks) > 0 {
			b.WriteString("\n**Scorecard Checks:**\n\n")
			b.WriteString("| Check | Score |\n")
			b.WriteString("|-------|-------|\n")
			for name, score := range report.Security.ScorecardChecks {
				b.WriteString(fmt.Sprintf("| %s | %.1f |\n", name, score))
			}
		}
	} else {
		b.WriteString("| OpenSSF Scorecard | Not available |\n")
		if report.Security.ScorecardReason != "" {
			b.WriteString(fmt.Sprintf("| | *%s* |\n", report.Security.ScorecardReason))
		}
		b.WriteString("| | *Security score capped at MODERATE without scorecard data* |\n")
	}
	b.WriteString("\n")

	// Maturity
	b.WriteString(fmt.Sprintf("## Maturity — %s\n\n", report.Maturity.Score))
	b.WriteString("| Metric | Value |\n")
	b.WriteString("|--------|-------|\n")
	b.WriteString(fmt.Sprintf("| Releases | %d |\n", report.Maturity.ReleasesCount))
	if report.Maturity.LatestRelease != "" {
		b.WriteString(fmt.Sprintf("| Latest release | %s |\n", report.Maturity.LatestRelease))
	}
	b.WriteString(fmt.Sprintf("| README | %v |\n", report.Maturity.HasReadme))
	b.WriteString(fmt.Sprintf("| Contributing guide | %v |\n", report.Maturity.HasContributing))
	b.WriteString(fmt.Sprintf("| Code of conduct | %v |\n", report.Maturity.HasCodeOfConduct))
	b.WriteString(fmt.Sprintf("| CI/CD | %v |\n", report.Maturity.HasCI))
	b.WriteString(fmt.Sprintf("| Description | %v |\n", report.Maturity.HasDescription))
	b.WriteString(fmt.Sprintf("| Topics | %d |\n", report.Maturity.TopicsCount))
	b.WriteString("\n")

	return b.String()
}
