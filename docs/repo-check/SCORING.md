# Scoring Methodology

Reference document for the repository trust scoring system.

## Category Scoring

### Popularity

| Signal | Criteria |
|--------|----------|
| HIGH | stars >= 1000, OR stars >= 500 AND forks >= 100 |
| MODERATE | stars >= 100, OR stars >= 50 AND forks >= 20 |
| LOW | < 100 stars |

### Activity

| Signal | Criteria |
|--------|----------|
| HIGH | last commit <= 30 days AND commits in last 90d >= 10 |
| MODERATE | last commit <= 180 days AND commits in last 90d >= 3 |
| LOW | last commit > 180 days, OR repository is archived |

### Security

**With OpenSSF Scorecard available:**

| Signal | Criteria |
|--------|----------|
| HIGH | scorecard score >= 7.0 AND has license |
| MODERATE | scorecard score >= 4.0 AND has license |
| LOW | scorecard score < 4.0, OR no license |

**Without OpenSSF Scorecard:**

| Signal | Criteria |
|--------|----------|
| HIGH | N/A — capped at MODERATE without scorecard |
| MODERATE | has license |
| LOW | no license |

### Maturity

Points system (max 6):
- README present: +1
- Contributing guide present: +1
- Code of conduct present: +1
- CI/CD configured: +1
- Any releases: +1
- 5+ releases: +1

| Signal | Criteria |
|--------|----------|
| HIGH | >= 5 points |
| MODERATE | >= 3 points |
| LOW | < 3 points |

## Composite Trust Signal

Weighted average using: Activity 35%, Security 25%, Popularity 20%, Maturity 20%.

Signal values: HIGH = 3, MODERATE = 2, LOW = 1.

| Composite Score | Trust Signal |
|-----------------|-------------|
| >= 2.5 | HIGH |
| >= 1.8 | MODERATE |
| < 1.8 | LOW |

### Overrides

- Activity scored LOW → composite capped at MODERATE
- Security scored LOW (with scorecard data available) → composite capped at MODERATE

### Weight Rationale

- **Activity (35%)**: An unmaintained repo is the highest risk — security issues won't be patched, bugs won't be fixed.
- **Security (25%)**: License and security posture directly affect adoption safety.
- **Popularity (20%)**: Community validation provides signal but can lag reality.
- **Maturity (20%)**: Documentation and release practices indicate project health and usability.

## OpenSSF Scorecard Reference

The [OpenSSF Scorecard](https://securityscorecards.dev/) evaluates open-source projects on security practices:

| Check | What It Measures |
|-------|-----------------|
| Binary-Artifacts | No checked-in binaries |
| Branch-Protection | Branch protection rules |
| CI-Tests | CI test presence |
| CII-Best-Practices | CII badge |
| Code-Review | Code review enforcement |
| Contributors | Active contributors |
| Dangerous-Workflow | No dangerous workflow patterns |
| Dependency-Update-Tool | Dependabot/Renovate usage |
| Fuzzing | Fuzz testing |
| License | License presence |
| Maintained | Maintenance activity |
| Packaging | Published packages |
| Pinned-Dependencies | Pinned dependency versions |
| SAST | Static analysis usage |
| Security-Policy | SECURITY.md present |
| Signed-Releases | Signed release artifacts |
| Token-Permissions | Minimal token permissions |
| Vulnerabilities | Known vulnerability count |

### Installing Scorecard

```bash
go install github.com/ossf/scorecard/v5/cmd/scorecard@latest
```

## Prerequisites

### gh CLI (required)

```bash
# macOS
brew install gh

# Linux
sudo apt install gh  # or snap, or see https://cli.github.com/

# Authenticate
gh auth login
```

### Go (required for building)

```bash
# See https://go.dev/dl/ for latest
# Minimum version: 1.24
go version
```
