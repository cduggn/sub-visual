package repocheck

// TrustReport is the top-level output of a repository trust analysis.
type TrustReport struct {
	Owner       string            `json:"owner"`
	Repo        string            `json:"repo"`
	URL         string            `json:"url"`
	TrustSignal string            `json:"trust_signal"`
	Warnings    []string          `json:"warnings"`
	Popularity  PopularityMetrics `json:"popularity"`
	Activity    ActivityMetrics   `json:"activity"`
	Security    SecurityMetrics   `json:"security"`
	Maturity    MaturityMetrics   `json:"maturity"`
}

type PopularityMetrics struct {
	Stars    int    `json:"stars"`
	Forks    int    `json:"forks"`
	Watchers int    `json:"watchers"`
	Score    string `json:"score"`
}

type ActivityMetrics struct {
	LastCommitDate      string `json:"last_commit_date"`
	DaysSinceLastCommit int    `json:"days_since_last_commit"`
	CommitsLast90d      int    `json:"commits_last_90d"`
	OpenIssues          int    `json:"open_issues"`
	OpenPRs             int    `json:"open_prs"`
	Contributors        int    `json:"contributors"`
	RepoAgeDays         int    `json:"repo_age_days"`
	Archived            bool   `json:"archived"`
	Score               string `json:"score"`
}

type SecurityMetrics struct {
	ScorecardAvailable bool               `json:"scorecard_available"`
	ScorecardReason    string             `json:"scorecard_reason"`
	ScorecardScore     float64            `json:"scorecard_score"`
	ScorecardChecks    map[string]float64 `json:"scorecard_checks"`
	LicenseName        string             `json:"license_name"`
	LicenseSPDX        string             `json:"license_spdx"`
	HasLicense         bool               `json:"has_license"`
	Score              string             `json:"score"`
}

type MaturityMetrics struct {
	ReleasesCount    int    `json:"releases_count"`
	LatestRelease    string `json:"latest_release"`
	HasReadme        bool   `json:"has_readme"`
	HasContributing  bool   `json:"has_contributing"`
	HasCodeOfConduct bool   `json:"has_code_of_conduct"`
	HasCI            bool   `json:"has_ci"`
	HasDescription   bool   `json:"has_description"`
	TopicsCount      int    `json:"topics_count"`
	Score            string `json:"score"`
}

// ScorecardResult holds parsed output from the OpenSSF Scorecard.
type ScorecardResult struct {
	Score  float64            `json:"score"`
	Checks map[string]float64 `json:"checks"`
}
