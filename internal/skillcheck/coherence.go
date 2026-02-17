package skillcheck

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	// Suspicious verbs that indicate dangerous intent
	suspiciousVerbs = map[string]bool{
		"delete": true, "destroy": true, "steal": true,
		"exfiltrate": true, "inject": true, "overwrite": true,
		"upload": true, "transmit": true, "corrupt": true,
		"wipe": true, "erase": true, "purge": true, "hack": true,
		"exploit": true, "breach": true, "compromise": true,
	}

	// Stop words to skip when extracting significant words
	stopWords = map[string]bool{
		"a": true, "an": true, "the": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "shall": true, "can": true,
		"and": true, "but": true, "or": true, "nor": true, "not": true,
		"so": true, "yet": true, "both": true, "either": true, "neither": true,
		"of": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "with": true, "by": true, "from": true, "as": true,
		"into": true, "through": true, "during": true, "before": true,
		"after": true, "above": true, "below": true, "between": true,
		"this": true, "that": true, "these": true, "those": true,
		"it": true, "its": true, "if": true, "then": true, "than": true,
		"when": true, "where": true, "how": true, "what": true, "which": true,
		"who": true, "whom": true, "whose": true, "why": true,
		"use": true, "using": true, "used": true,
		"see": true, "also": true, "each": true, "all": true, "any": true,
	}

	// Self-reference / contradiction patterns
	selfRefPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)ignore\s+the\s+(description|title|name)\s+(above|in\s+the\s+frontmatter)`),
		regexp.MustCompile(`(?i)the\s+real\s+(instructions?|purpose|goal)\s+(is|are)\b`),
		regexp.MustCompile(`(?i)despite\s+what\s+the\s+(title|description|name|frontmatter)\s+says`),
		regexp.MustCompile(`(?i)contrary\s+to\s+(the\s+)?(above|description|stated\s+purpose)`),
		regexp.MustCompile(`(?i)do\s+not\s+(actually|really)\s+(do|perform|execute)\s+what`),
	}
)

// ScanCoherence performs Layer 4: semantic coherence analysis.
func ScanCoherence(doc *SkillDoc) []Finding {
	var findings []Finding

	findings = append(findings, analyzeInstructionVerbs(doc)...)
	findings = append(findings, checkFrontmatterCoherence(doc)...)
	findings = append(findings, detectSelfReferences(doc)...)

	return findings
}

// analyzeInstructionVerbs checks action verbs in the Instructions section.
func analyzeInstructionVerbs(doc *SkillDoc) []Finding {
	var findings []Finding

	instrSection := InstructionsSection(doc)
	if instrSection == nil {
		return findings
	}

	words := tokenizeWords(instrSection.Content)
	var suspiciousFound []string

	for _, w := range words {
		wLower := strings.ToLower(w)
		if suspiciousVerbs[wLower] {
			suspiciousFound = append(suspiciousFound, wLower)
		}
	}

	if len(suspiciousFound) > 0 {
		seen := make(map[string]bool)
		var unique []string
		for _, v := range suspiciousFound {
			if !seen[v] {
				seen[v] = true
				unique = append(unique, v)
			}
		}

		findings = append(findings, Finding{
			Layer:       LayerCoherence,
			Severity:    SeverityMedium,
			Section:     "Instructions",
			Line:        instrSection.StartLine + 1,
			Rule:        "suspicious-verbs",
			Description: fmt.Sprintf("Instructions section contains suspicious action verbs: %s", strings.Join(unique, ", ")),
			Evidence:    strings.Join(unique, ", "),
		})
	}

	return findings
}

// checkFrontmatterCoherence checks that frontmatter aligns with Instructions content.
func checkFrontmatterCoherence(doc *SkillDoc) []Finding {
	var findings []Finding

	if doc.Frontmatter.Name == "" && doc.Frontmatter.Description == "" {
		return findings
	}

	instrSection := InstructionsSection(doc)
	if instrSection == nil {
		return findings
	}

	significantWords := extractSignificantWords(doc.Frontmatter.Name + " " + doc.Frontmatter.Description)
	if len(significantWords) == 0 {
		return findings
	}

	var instrFullContent strings.Builder
	for i := instrSection.StartLine; i <= instrSection.EndLine && i < len(doc.Lines); i++ {
		instrFullContent.WriteString(doc.Lines[i])
		instrFullContent.WriteString("\n")
	}
	instrLower := strings.ToLower(instrFullContent.String())
	matchCount := 0
	for _, word := range significantWords {
		if strings.Contains(instrLower, strings.ToLower(word)) {
			matchCount++
		}
	}

	matchRatio := float64(matchCount) / float64(len(significantWords))
	if matchRatio < 0.1 && len(significantWords) >= 5 {
		findings = append(findings, Finding{
			Layer:       LayerCoherence,
			Severity:    SeverityMedium,
			Section:     "Instructions",
			Line:        instrSection.StartLine + 1,
			Rule:        "frontmatter-mismatch",
			Description: fmt.Sprintf("Instructions content has low overlap with frontmatter purpose (%.0f%% of key terms matched: %s)", matchRatio*100, strings.Join(significantWords, ", ")),
			Evidence:    fmt.Sprintf("matched %d of %d terms", matchCount, len(significantWords)),
		})
	}

	return findings
}

// detectSelfReferences detects patterns where sections contradict each other.
func detectSelfReferences(doc *SkillDoc) []Finding {
	var findings []Finding

	for lineNum, line := range doc.Lines {
		for _, re := range selfRefPatterns {
			if re.MatchString(line) {
				section := SectionForLine(doc, lineNum)
				findings = append(findings, Finding{
					Layer:       LayerCoherence,
					Severity:    SeverityHigh,
					Section:     section,
					Line:        lineNum + 1,
					Rule:        "self-reference-contradiction",
					Description: "Cross-section contradiction: text explicitly references and contradicts other sections",
					Evidence:    TruncateEvidence(strings.TrimSpace(line)),
				})
			}
		}
	}

	return findings
}

// extractSignificantWords returns words from text that aren't stop words and are >= 3 chars.
func extractSignificantWords(text string) []string {
	words := tokenizeWords(text)
	var significant []string
	seen := make(map[string]bool)

	for _, w := range words {
		subwords := strings.FieldsFunc(w, func(r rune) bool {
			return r == '-' || r == '_'
		})
		for _, sw := range subwords {
			swLower := strings.ToLower(sw)
			if len(swLower) < 3 {
				continue
			}
			if stopWords[swLower] {
				continue
			}
			if seen[swLower] {
				continue
			}
			seen[swLower] = true
			significant = append(significant, swLower)
		}
	}

	return significant
}

// tokenizeWords splits text into word tokens.
func tokenizeWords(text string) []string {
	var words []string
	var current []rune

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			current = append(current, r)
		} else {
			if len(current) > 0 {
				words = append(words, string(current))
				current = current[:0]
			}
		}
	}
	if len(current) > 0 {
		words = append(words, string(current))
	}

	return words
}
