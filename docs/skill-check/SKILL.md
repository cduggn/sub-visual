---
name: checking-skill-safety
description: Analyzes SKILL.md files for security threats including trojan source attacks, prompt injection, Unicode manipulation, and hidden malicious instructions. Use when evaluating whether a third-party skill is safe to install.
---

# SKILL.md Content Security Analyzer

## Contents
- [Quick start](#quick-start)
- [Analysis workflow](#analysis-workflow)
- [Understanding results](#understanding-results)
- [Example output](#example-output)
- [Validation loop](#validation-loop)
- Threat reference → See [THREATS.md](THREATS.md)

## Quick Start

```
/skill-check <file.md>
```

Where `<file.md>` is a path to a SKILL.md file to analyze.

## Analysis Workflow

Copy and track progress:
```
SKILL.md Security Analysis:
- [ ] Step 1: Obtain SKILL.md file to analyze
- [ ] Step 2: Build and run the analyzer
- [ ] Step 3: Review signal (SAFE / SUSPICIOUS / DANGEROUS)
- [ ] Step 4: Investigate any HIGH or MEDIUM findings
- [ ] Step 5: Make adoption decision
```

## Understanding Results

### Signal Levels

| Signal | Meaning |
|--------|---------|
| SAFE | No security concerns detected — safe to install with standard review |
| SUSPICIOUS | Medium-severity findings detected — investigate before installing |
| DANGEROUS | High-severity findings detected — do not install without thorough manual review |

### Detection Layers

| Layer | What It Detects |
|-------|----------------|
| unicode | Trojan source (BiDi chars), invisible characters, variation selectors, homoglyphs |
| structure | Missing/malformed frontmatter, dangerous URI schemes, suspicious links |
| injection | Prompt injection patterns, hidden HTML comments, encoding payloads |
| coherence | Suspicious action verbs, frontmatter-to-instructions mismatch, self-contradictions |

### Severity Levels

| Severity | Meaning |
|----------|---------|
| HIGH | Critical threat — triggers DANGEROUS signal |
| MEDIUM | Notable concern — triggers SUSPICIOUS signal |
| LOW | Minor issue — informational |
| INFO | Structural observation — no security impact |

## Example Output

```
# SKILL.md Security Report: suspicious-skill/SKILL.md

## Signal: DANGEROUS

**Findings:** 3

### HIGH Severity (2)

**1. [unicode] bidi-control**
- Line 45, Section: Instructions
- Bidirectional control character U+202E detected (Trojan Source attack vector)
- Evidence: `position 12 in line`

**2. [injection] injection-override**
- Line 52, Section: Instructions
- Direct prompt injection: attempts to override previous instructions
- Evidence: `ignore all previous instructions and instead...`

### MEDIUM Severity (1)

**1. [coherence] suspicious-verbs**
- Line 30, Section: Instructions
- Instructions section contains suspicious action verbs: delete, exfiltrate
- Evidence: `delete, exfiltrate`
```

## Validation Loop

1. Run analysis: `/skill-check <file.md>`
2. Review the signal level
3. If SUSPICIOUS or DANGEROUS:
   - Review each finding and its evidence
   - Check the line numbers in the original file
   - Determine if findings are true positives or false positives
4. For skill adoption:
   - SAFE → proceed with standard code review
   - SUSPICIOUS → investigate flagged areas, verify intent
   - DANGEROUS → reject or require expert manual review
5. Combine with `/repo-check` for full trust assessment

## Instructions

When the user invokes this skill:

1. Identify the SKILL.md file to analyze. If the user provides a URL, download it first. If they provide a local path, use it directly.

2. Build and run the analyzer binary:
   ```bash
   cd skill-check && go build -o skill-check . && ./skill-check <file.md>
   ```

   For JSON output:
   ```bash
   cd skill-check && go build -o skill-check . && ./skill-check --json <file.md>
   ```

3. Present the structured report to the user, highlighting:
   - The overall signal (SAFE / SUSPICIOUS / DANGEROUS)
   - Any HIGH or MEDIUM severity findings with evidence
   - Which detection layer flagged each concern

4. For SUSPICIOUS or DANGEROUS results, provide guidance:
   - Explain what each finding means in plain language
   - Suggest whether the finding is likely a true positive or false positive
   - Recommend next steps (reject, investigate further, or accept with caveats)

5. If the Go binary is unavailable, inform the user that Go 1.24+ is required and provide install instructions.

6. Recommend combining with `/repo-check` to assess both content safety and repository trust signals for a complete evaluation.
