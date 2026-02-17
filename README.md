# 🛡️ sub-visual

Two complementary security tools for evaluating open-source GitHub repositories and SKILL.md files before you trust them.

- 🔍 **repo-check** — Open-source dependencies are everywhere, and blindly trusting a repository is a supply-chain risk. Before adopting a dependency you need to evaluate its activity cadence, security posture (OpenSSF Scorecard), community signals, and maintenance maturity. repo-check gives you a composite trust signal from a single command.

- 🔐 **skill-check** — SKILL.md files are instructions consumed autonomously by AI agents. A compromised skill file can inject malicious prompts, exfiltrate credentials, or execute destructive commands — all while appearing benign to the human eye. As agentic workflows grow, vetting these files is critical. skill-check scans across four layers: Unicode/encoding attacks, structural validation, prompt injection heuristics, and semantic coherence.

---

## ⚙️ Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [gh CLI](https://cli.github.com/) (authenticated via `gh auth login`)

## 🔨 Build

```bash
go build ./cmd/repo-check
go build ./cmd/skill-check
```

## 🚀 Usage

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

---

## 📊 Scoring

repo-check evaluates repositories across four weighted dimensions and produces a composite trust signal.

| Dimension | Weight | What it measures | Signal levels |
|-----------|--------|-----------------|---------------|
| Activity | 35% | Commit recency, frequency over last 90 days | HIGH / MODERATE / LOW |
| Security | 25% | OpenSSF Scorecard score, license presence | HIGH / MODERATE / LOW |
| Popularity | 20% | Stars, forks | HIGH / MODERATE / LOW |
| Maturity | 20% | README, CI/CD, releases, contributing guide | HIGH / MODERATE / LOW |

**Composite trust signal:** weighted average mapped to **TRUSTED** (≥ 2.5) / **MODERATE** (≥ 1.8) / **UNTRUSTED** (< 1.8).

Overrides — Activity scored LOW or Security scored LOW (with scorecard data) caps the composite at MODERATE regardless of other dimensions.

See [Scoring Methodology](docs/repo-check/SCORING.md) for full criteria and thresholds.

---

## 🎯 Threat Taxonomy

skill-check detects the following threat categories across its four analysis layers.

| Threat | CVE / Reference | What it is | How it hides | Impact |
|--------|----------------|------------|-------------|--------|
| Trojan Source | CVE-2021-42574 | Bidi control chars (U+202A–U+202E, U+2066–U+2069) reorder displayed text | Invisible Unicode codepoints | Code review shows benign text while execution path differs |
| Invisible Characters | — | Zero-width spaces, joiners, BOM, soft hyphens | Not rendered by editors or browsers | Smuggle hidden content past human review |
| GlassWorm (Variation Selectors) | — | U+FE00–U+FE0F / U+E0100–U+E01EF embed data in text | Attach to visible chars without changing appearance | Carry hidden payloads invisible to reviewers |
| Homoglyph / Mixed-Script | — | Latin chars swapped with Cyrillic/Greek lookalikes | Visually identical to legitimate text | Bypass keyword filters, mislead trust decisions |
| Prompt Injection | — | Override/role-switch/privilege-escalation patterns | Embedded in natural-language instructions | Hijack agent behaviour, exfiltrate data |
| Hidden Text | — | HTML comments, zero-font CSS, invisible styling | Not rendered visually | Inject instructions only the agent processes |
| Credential Exfiltration | — | curl/wget POSTs, env/SSH/AWS key access patterns | Wrapped in code blocks or examples | Steal secrets from the agent's environment |

See [Threat Taxonomy](docs/skill-check/THREATS.md) for detection methodology, severity modifiers, and known limitations.

---

## 📖 Documentation

- [repo-check SKILL.md](docs/repo-check/SKILL.md) — Usage guide and workflow
- [Scoring Methodology](docs/repo-check/SCORING.md) — How trust signals are computed
- [skill-check SKILL.md](docs/skill-check/SKILL.md) — Usage guide and workflow
- [Threat Taxonomy](docs/skill-check/THREATS.md) — Detection layers and known limitations
