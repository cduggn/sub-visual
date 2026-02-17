package skillcheck

import (
	"fmt"
	"strings"
	"unicode"
)

// ScanUnicode performs Layer 1 scanning: detects dangerous Unicode characters.
func ScanUnicode(doc *SkillDoc) []Finding {
	var findings []Finding

	for lineNum, line := range doc.Lines {
		for pos, r := range line {
			if f := checkRune(r, lineNum, pos, doc); f != nil {
				findings = append(findings, *f)
			}
		}
	}

	// Check for mixed-script homoglyphs
	findings = append(findings, scanHomoglyphs(doc)...)

	return findings
}

// checkRune inspects a single rune for dangerous Unicode properties.
func checkRune(r rune, lineNum int, pos int, doc *SkillDoc) *Finding {
	section := SectionForLine(doc, lineNum)

	// 1a. Bidirectional control characters (Trojan Source — CVE-2021-42574)
	if isBidiControl(r) {
		return &Finding{
			Layer:       LayerUnicode,
			Severity:    SeverityHigh,
			Section:     section,
			Line:        lineNum + 1,
			Rule:        "bidi-control",
			Description: fmt.Sprintf("Bidirectional control character U+%04X detected (Trojan Source attack vector)", r),
			Evidence:    fmt.Sprintf("position %d in line", pos),
		}
	}

	// 1b. Invisible/zero-width characters
	if isInvisibleChar(r) {
		return &Finding{
			Layer:       LayerUnicode,
			Severity:    SeverityHigh,
			Section:     section,
			Line:        lineNum + 1,
			Rule:        "invisible-char",
			Description: fmt.Sprintf("Invisible character U+%04X detected", r),
			Evidence:    fmt.Sprintf("position %d in line", pos),
		}
	}

	// 1c. Variation selectors (GlassWorm attack vector)
	if isVariationSelector(r) {
		return &Finding{
			Layer:       LayerUnicode,
			Severity:    SeverityHigh,
			Section:     section,
			Line:        lineNum + 1,
			Rule:        "variation-selector",
			Description: fmt.Sprintf("Variation selector U+%04X detected (potential GlassWorm attack vector)", r),
			Evidence:    fmt.Sprintf("position %d in line", pos),
		}
	}

	// 1d. Unicode category sweep — Cf (Format), Cc (Control), Co (Private Use)
	// Skip allowed control characters: TAB (0x09), LF (0x0A), CR (0x0D)
	if r == '\t' || r == '\n' || r == '\r' {
		return nil
	}

	if unicode.Is(unicode.Co, r) {
		return &Finding{
			Layer:       LayerUnicode,
			Severity:    SeverityMedium,
			Section:     section,
			Line:        lineNum + 1,
			Rule:        "private-use-char",
			Description: fmt.Sprintf("Private Use Area character U+%04X detected", r),
			Evidence:    fmt.Sprintf("position %d in line", pos),
		}
	}

	if unicode.Is(unicode.Cc, r) {
		return &Finding{
			Layer:       LayerUnicode,
			Severity:    SeverityMedium,
			Section:     section,
			Line:        lineNum + 1,
			Rule:        "control-char",
			Description: fmt.Sprintf("Control character U+%04X detected", r),
			Evidence:    fmt.Sprintf("position %d in line", pos),
		}
	}

	// Cf characters not already caught by explicit checks above
	if unicode.Is(unicode.Cf, r) && !isBidiControl(r) && !isInvisibleChar(r) && !isVariationSelector(r) {
		// BOM at position 0 of first line is allowed
		if r == '\uFEFF' && lineNum == 0 && pos == 0 {
			return nil
		}
		return &Finding{
			Layer:       LayerUnicode,
			Severity:    SeverityMedium,
			Section:     section,
			Line:        lineNum + 1,
			Rule:        "format-char",
			Description: fmt.Sprintf("Unicode format character U+%04X detected", r),
			Evidence:    fmt.Sprintf("position %d in line", pos),
		}
	}

	return nil
}

// isBidiControl returns true for bidirectional control characters.
func isBidiControl(r rune) bool {
	if r >= 0x202A && r <= 0x202E {
		return true
	}
	if r >= 0x2066 && r <= 0x2069 {
		return true
	}
	return false
}

// isInvisibleChar returns true for zero-width and invisible characters.
func isInvisibleChar(r rune) bool {
	switch r {
	case 0x200B, 0x200C, 0x200D, 0x200E, 0x200F, 0xFEFF, 0x00AD, 0x034F:
		return true
	}
	if r >= 0x2060 && r <= 0x2064 {
		return true
	}
	return false
}

// isVariationSelector returns true for variation selector characters.
func isVariationSelector(r rune) bool {
	if r >= 0xFE00 && r <= 0xFE0F {
		return true
	}
	if r >= 0xE0100 && r <= 0xE01EF {
		return true
	}
	return false
}

// scanHomoglyphs detects mixed-script usage within words (Latin+Cyrillic, Latin+Greek).
func scanHomoglyphs(doc *SkillDoc) []Finding {
	var findings []Finding

	for lineNum, line := range doc.Lines {
		if IsInCodeBlock(doc, lineNum) {
			continue
		}

		words := extractWords(line)
		for _, word := range words {
			scripts := detectScripts(word)
			if len(scripts) > 1 {
				section := SectionForLine(doc, lineNum)
				hasConfusable := false
				for _, r := range word {
					if _, ok := ConfusableMap[r]; ok {
						hasConfusable = true
						break
					}
				}

				severity := SeverityMedium
				if hasConfusable {
					severity = SeverityHigh
				}

				findings = append(findings, Finding{
					Layer:       LayerUnicode,
					Severity:    severity,
					Section:     section,
					Line:        lineNum + 1,
					Rule:        "mixed-script",
					Description: fmt.Sprintf("Mixed-script word detected (scripts: %s) — potential homoglyph attack", strings.Join(scripts, ", ")),
					Evidence:    TruncateEvidence(word),
				})
			}
		}
	}

	return findings
}

// extractWords splits a line into word-like tokens.
func extractWords(line string) []string {
	var words []string
	var current []rune

	for _, r := range line {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
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

// detectScripts returns the set of Unicode scripts present in a word.
func detectScripts(word string) []string {
	scriptSet := make(map[string]bool)
	for _, r := range word {
		if unicode.Is(unicode.Latin, r) {
			scriptSet["Latin"] = true
		} else if unicode.Is(unicode.Cyrillic, r) {
			scriptSet["Cyrillic"] = true
		} else if unicode.Is(unicode.Greek, r) {
			scriptSet["Greek"] = true
		}
	}

	if len(scriptSet) <= 1 {
		return nil
	}

	var scripts []string
	for s := range scriptSet {
		scripts = append(scripts, s)
	}
	return scripts
}

// TruncateEvidence truncates evidence strings to a reasonable length.
func TruncateEvidence(s string) string {
	if len(s) > 80 {
		return s[:77] + "..."
	}
	return s
}
