package repocheck

const (
	SignalHigh     = "HIGH"
	SignalModerate = "MODERATE"
	SignalLow      = "LOW"
)

// ScorePopularity scores the popularity metrics.
func ScorePopularity(p *PopularityMetrics) string {
	if p.Stars >= 1000 || (p.Stars >= 500 && p.Forks >= 100) {
		p.Score = SignalHigh
	} else if p.Stars >= 100 || (p.Stars >= 50 && p.Forks >= 20) {
		p.Score = SignalModerate
	} else {
		p.Score = SignalLow
	}
	return p.Score
}

// ScoreActivity scores the activity metrics.
func ScoreActivity(a *ActivityMetrics) string {
	if a.Archived {
		a.Score = SignalLow
		return a.Score
	}
	if a.DaysSinceLastCommit <= 30 && a.CommitsLast90d >= 10 {
		a.Score = SignalHigh
	} else if a.DaysSinceLastCommit <= 180 && a.CommitsLast90d >= 3 {
		a.Score = SignalModerate
	} else {
		a.Score = SignalLow
	}
	return a.Score
}

// ScoreSecurity scores the security metrics.
func ScoreSecurity(s *SecurityMetrics) string {
	if s.ScorecardAvailable {
		if s.ScorecardScore >= 7.0 && s.HasLicense {
			s.Score = SignalHigh
		} else if s.ScorecardScore >= 4.0 && s.HasLicense {
			s.Score = SignalModerate
		} else {
			s.Score = SignalLow
		}
	} else {
		// Without scorecard, cap at MODERATE
		if s.HasLicense {
			s.Score = SignalModerate
		} else {
			s.Score = SignalLow
		}
	}
	return s.Score
}

// ScoreMaturity scores the maturity metrics.
func ScoreMaturity(m *MaturityMetrics) string {
	points := 0
	if m.HasReadme {
		points++
	}
	if m.HasContributing {
		points++
	}
	if m.HasCodeOfConduct {
		points++
	}
	if m.HasCI {
		points++
	}
	if m.ReleasesCount > 0 {
		points++
	}
	if m.ReleasesCount >= 5 {
		points++
	}

	if points >= 5 {
		m.Score = SignalHigh
	} else if points >= 3 {
		m.Score = SignalModerate
	} else {
		m.Score = SignalLow
	}
	return m.Score
}

func signalValue(s string) float64 {
	switch s {
	case SignalHigh:
		return 3
	case SignalModerate:
		return 2
	case SignalLow:
		return 1
	default:
		return 0
	}
}

// ComputeComposite calculates the weighted composite trust signal.
// Weights: Activity 35%, Security 25%, Popularity 20%, Maturity 20%.
func ComputeComposite(report *TrustReport) string {
	weighted := signalValue(report.Activity.Score)*0.35 +
		signalValue(report.Security.Score)*0.25 +
		signalValue(report.Popularity.Score)*0.20 +
		signalValue(report.Maturity.Score)*0.20

	signal := SignalLow
	if weighted >= 2.5 {
		signal = SignalHigh
	} else if weighted >= 1.8 {
		signal = SignalModerate
	}

	// Override: Activity LOW caps at MODERATE
	if report.Activity.Score == SignalLow && signal == SignalHigh {
		signal = SignalModerate
	}

	// Override: Security LOW (with scorecard data) caps at MODERATE
	if report.Security.Score == SignalLow && report.Security.ScorecardAvailable && signal == SignalHigh {
		signal = SignalModerate
	}

	return signal
}

// CollectWarnings generates human-readable warnings based on the report.
func CollectWarnings(report *TrustReport) []string {
	var w []string

	if report.Activity.Archived {
		w = append(w, "Repository is archived — no longer maintained")
	}
	if report.Activity.DaysSinceLastCommit > 365 {
		w = append(w, "No commits in over a year")
	} else if report.Activity.DaysSinceLastCommit > 180 {
		w = append(w, "No commits in over 6 months")
	}
	if report.Popularity.Stars < 10 {
		w = append(w, "Very low popularity (< 10 stars)")
	}
	if !report.Security.HasLicense {
		w = append(w, "No license detected — usage rights unclear")
	}
	if report.Security.ScorecardAvailable && report.Security.ScorecardScore < 4.0 {
		w = append(w, "OpenSSF Scorecard score below 4.0 — review security practices")
	}
	if !report.Security.ScorecardAvailable {
		w = append(w, "OpenSSF Scorecard not available — security score capped at MODERATE")
	}
	if report.Activity.Contributors == 1 {
		w = append(w, "Single contributor — bus factor risk")
	}
	if report.Activity.RepoAgeDays < 90 {
		w = append(w, "Repository is less than 90 days old")
	}
	if !report.Maturity.HasReadme {
		w = append(w, "No README found")
	}

	return w
}
