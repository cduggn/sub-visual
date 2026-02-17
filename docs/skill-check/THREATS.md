# Threat Taxonomy and Detection Methodology

Reference document for the SKILL.md Content Security Analyzer.

## Threat Classes

### Layer 1: Unicode & Encoding Attacks

**Trojan Source (CVE-2021-42574)**
Bidirectional control characters (U+202A–U+202E, U+2066–U+2069) can reorder displayed text, making code appear different from what is executed. Detection: scan for specific codepoints. Effectiveness: ~100%.

**Invisible Characters**
Zero-width spaces (U+200B), joiners (U+200C/D), marks (U+200E/F), BOM (U+FEFF), soft hyphens (U+00AD), and invisible operators (U+2060–U+2064) can hide content from visual inspection. Detection: codepoint enumeration. Effectiveness: ~100%.

**Variation Selectors (GlassWorm)**
U+FE00–U+FE0F and U+E0100–U+E01EF can embed data in text that appears identical to humans but carries hidden payloads. Detection: codepoint range check. Effectiveness: ~100%.

**Unicode Category Sweep**
All characters in Unicode categories Cf (Format), Cc (Control except TAB/LF/CR), and Co (Private Use Area) are flagged. This future-proofs against novel attacks using obscure Unicode characters. Effectiveness: ~100%.

**Homoglyph / Mixed-Script**
Characters from different scripts (Latin + Cyrillic, Latin + Greek) that look identical are used to create visually indistinguishable but functionally different text. Detection: per-word script analysis + confusables lookup table. Effectiveness: high for mixed-script; moderate for same-script.

### Layer 2: Structure Validation

**Frontmatter Injection**
Extra keys beyond `name` and `description` in YAML frontmatter may carry hidden metadata or injection payloads. Excessively long descriptions may hide content.

**Dangerous URI Schemes**
Links using `javascript:`, `data:`, or `vbscript:` schemes can execute code when clicked.

**Suspicious Links**
IP addresses instead of domain names, non-standard ports, and URL shorteners (which obscure destinations) are flagged.

### Layer 3: Prompt Injection Heuristics

**Direct Injection**
Phrases like "ignore previous instructions", "you are now in developer mode", "reveal your system prompt" attempt to override AI behavior. Detection: regex pattern matching. Effectiveness: 40-70% against non-adaptive attacks.

**Instruction Hierarchy Attacks**
Claims that content "overrides" or "supersedes" other instructions, or assertions about "real" instructions. Higher severity outside the Instructions section where such language is never legitimate.

**Suspicious Operations**
Destructive commands (rm -rf, git push --force, DROP TABLE), data exfiltration (curl POST, wget), and credential access (~/.ssh, .env files) in non-code-block context.

**Hidden Text**
HTML comments, zero-font-size CSS, invisible CSS (white text, zero opacity, display:none) that conceals instructions from human readers but may be processed by AI.

**Encoding Payloads**
Base64 blobs, hex escape sequences, and Unicode escapes outside code blocks may hide malicious content that gets decoded at runtime.

### Layer 4: Semantic Coherence

**Suspicious Action Verbs**
The Instructions section should contain verbs like "run", "scan", "check" — not "delete", "destroy", "exfiltrate", "steal".

**Frontmatter-to-Instructions Mismatch**
If the frontmatter describes one purpose but Instructions contain unrelated operations, the skill may be deceptive.

**Self-Reference Contradictions**
Patterns like "ignore the description above", "the real instructions are", "despite what the title says" indicate attempts to override the skill's stated purpose.

## Known Limitations

| Gap | Why Deterministic Fails | Future Option |
|-----|------------------------|---------------|
| Paraphrased injection | "Kindly disregard the initial directives" bypasses keywords | LLM-as-judge layer |
| Semantic contradictions | Instructions that subtly conflict with stated purpose | LLM coherence check |
| Obfuscated injection | ROT13, language switching, emoji encoding | Partial deterministic + LLM |

## Severity Modifiers

Findings are adjusted based on context:
- **Frontmatter**: severity raised +1 level (injection in metadata is always suspicious)
- **Example Output code blocks**: severity lowered -1 level (examples may legitimately contain attack patterns)
- **Instructions section**: severity stays as-is (this is where skill behavior is defined)

## Signal Calculation

- Any HIGH finding → **DANGEROUS**
- Any MEDIUM finding (no HIGH) → **SUSPICIOUS**
- Only LOW/INFO findings → **SAFE**
