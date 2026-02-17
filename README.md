# sub-visual

Fused repository trust analyzer + SKILL.md content security checker.

Two complementary security tools for evaluating open-source GitHub repositories and skill files:

- **repo-check** — Analyzes GitHub repositories for trust signals across popularity, activity, security (OpenSSF Scorecard), and maturity
- **skill-check** — Detects security threats in SKILL.md files including trojan source attacks, prompt injection, Unicode manipulation, and hidden instructions

## Prerequisites

- [Go 1.24+](https://go.dev/dl/)
- [gh CLI](https://cli.github.com/) (authenticated via `gh auth login`)

## Build

```bash
go build ./cmd/repo-check
go build ./cmd/skill-check
```

## Usage

### repo-check

```bash
# By shorthand
./repo-check facebook/react

# By URL
./repo-check https://github.com/pallets/flask

# By SKILL.md blob URL
./repo-check https://github.com/owner/repo/blob/main/SKILL.md

# JSON output
./repo-check --json owner/repo
```

### skill-check

```bash
# Local file
./skill-check path/to/SKILL.md

# GitHub blob URL
./skill-check https://github.com/owner/repo/blob/main/SKILL.md

# JSON output
./skill-check --json path/to/SKILL.md
```

Exit code 2 indicates a **DANGEROUS** signal.

## Project Structure

```
sub-visual/
├── cmd/
│   ├── repo-check/main.go       # CLI entry point for repository trust analysis
│   └── skill-check/main.go      # CLI entry point for SKILL.md security analysis
├── internal/
│   ├── github/                   # Shared GitHub API utilities
│   │   ├── api.go                # GhAPI(), CheckCLI()
│   │   ├── target.go             # ParseTarget(), GitHubTarget struct
│   │   └── fetch.go              # FetchFileContent()
│   ├── repocheck/                # Repository trust analysis engine
│   │   ├── types.go              # TrustReport, metrics structs
│   │   ├── fetch.go              # GitHub API data fetchers
│   │   ├── scoring.go            # Scoring logic and composite signal
│   │   ├── scorecard.go          # OpenSSF Scorecard integration
│   │   └── report.go             # JSON and Markdown formatters
│   └── skillcheck/               # SKILL.md security analysis engine
│       ├── types.go              # SkillReport, Finding, SkillDoc
│       ├── parse.go              # Markdown/frontmatter parser
│       ├── unicode.go            # Layer 1: Unicode/trojan source detection
│       ├── confusables.go        # Homoglyph mapping table
│       ├── structure.go          # Layer 2: Structure validation
│       ├── injection.go          # Layer 3: Prompt injection detection
│       ├── coherence.go          # Layer 4: Semantic coherence analysis
│       └── report.go             # Analyze(), formatters, signal computation
├── docs/
│   ├── repo-check/               # SKILL.md + SCORING.md
│   └── skill-check/              # SKILL.md + THREATS.md
└── evals/
    ├── repo-check/               # promptfoo eval config
    └── skill-check/              # promptfoo eval config + test fixtures
```

## Documentation

- [repo-check SKILL.md](docs/repo-check/SKILL.md) — Usage guide and workflow
- [Scoring Methodology](docs/repo-check/SCORING.md) — How trust signals are computed
- [skill-check SKILL.md](docs/skill-check/SKILL.md) — Usage guide and workflow
- [Threat Taxonomy](docs/skill-check/THREATS.md) — Detection layers and known limitations
