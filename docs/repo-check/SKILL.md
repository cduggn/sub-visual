---
name: checking-repo-trust
description: Analyzes GitHub repositories for trust and quality signals including popularity, activity, security posture (OpenSSF Scorecard), and maturity. Use when evaluating whether to adopt or depend on an open-source repository or skill.
---

# Repository Trust Analyzer

## Contents
- [Quick start](#quick-start)
- [Analysis workflow](#analysis-workflow)
- [Understanding results](#understanding-results)
- [Example output](#example-output)
- [Validation loop](#validation-loop)
- Scoring methodology → See [SCORING.md](SCORING.md)

## Quick Start

```
/repo-check <target>
```

Where `<target>` is one of:
- `owner/repo` — shorthand (e.g. `facebook/react`)
- `https://github.com/owner/repo` — full URL
- `https://github.com/owner/repo/blob/branch/path/SKILL.md` — SKILL.md URL

## Analysis Workflow

Copy and track progress:
```
Repository Trust Analysis:
- [ ] Step 1: Parse input to extract owner/repo
- [ ] Step 2: Verify gh CLI authentication
- [ ] Step 3: Build and run the analyzer
- [ ] Step 4: Review trust signal and category scores
- [ ] Step 5: Review any warnings
- [ ] Step 6: Make adoption decision
```

## Understanding Results

### Trust Signal Levels

| Signal | Meaning |
|--------|---------|
| HIGH | Strong trust indicators across categories — safe to adopt with standard review |
| MODERATE | Acceptable but review flagged areas — investigate warnings before adopting |
| LOW | Significant concerns — manual deep review required before adoption |

### Categories Analyzed

| Category | Weight | What It Measures |
|----------|--------|-----------------|
| Activity | 35% | Commit recency, frequency, contributor count, archive status |
| Security | 25% | OpenSSF Scorecard (if available), license presence |
| Popularity | 20% | Stars, forks, watchers |
| Maturity | 20% | README, contributing guide, CI/CD, releases, code of conduct |

## Example Output

```
# Repository Trust Report: facebook/react

**URL:** https://github.com/facebook/react

## Overall Trust Signal: HIGH

## Popularity — HIGH

| Metric | Value |
|--------|-------|
| Stars | 230000 |
| Forks | 47000 |
| Watchers | 6500 |

## Activity — HIGH

| Metric | Value |
|--------|-------|
| Last commit | 2026-02-15T10:30:00Z (1 days ago) |
| Commits (90d) | 350 |
| Open issues | 900 |
| Contributors | 1800 |

## Security — MODERATE

| Metric | Value |
|--------|-------|
| License | MIT License (MIT) |
| OpenSSF Scorecard | Not available |

## Maturity — HIGH

| Metric | Value |
|--------|-------|
| Releases | 100 |
| README | true |
| Contributing guide | true |
| CI/CD | true |
```

## Validation Loop

1. Run analysis: `/repo-check <target>`
2. Review the overall trust signal
3. If MODERATE or LOW:
   - Check the warnings section for specific concerns
   - Review the lowest-scoring categories
   - Investigate flagged areas manually
4. For skill adoption:
   - HIGH → proceed with standard code review
   - MODERATE → deeper review of flagged areas, check issue tracker
   - LOW → consider alternatives, require team sign-off
5. Re-run periodically to track changes

## Instructions

When the user invokes this skill:

1. Parse the input to extract `owner/repo` from the provided target (shorthand, URL, or SKILL.md URL).

2. Check that `gh` CLI is authenticated:
   ```bash
   gh auth status
   ```

3. Build and run the analyzer binary:
   ```bash
   cd repo-check && go build -o repo-check . && ./repo-check <target>
   ```

   For JSON output:
   ```bash
   cd repo-check && go build -o repo-check . && ./repo-check --json <target>
   ```

4. Present the structured report to the user, highlighting:
   - The overall trust signal (HIGH/MODERATE/LOW)
   - Any warnings that need attention
   - Category-level scores for deeper analysis

5. If the Go binary is unavailable (build fails, Go not installed), fall back to manual `gh api` calls:
   ```bash
   # Repo metadata
   gh api repos/{owner}/{repo}

   # Commit activity
   gh api repos/{owner}/{repo}/stats/participation

   # Community profile
   gh api repos/{owner}/{repo}/community/profile

   # Releases
   gh api repos/{owner}/{repo}/releases?per_page=100
   ```

   Then manually assess trust signals based on the raw data and the scoring methodology in [SCORING.md](SCORING.md).

6. Provide a recommendation based on the trust signal and the user's context (adopting a dependency vs. evaluating a skill vs. general assessment).
