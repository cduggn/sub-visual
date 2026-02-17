package skillcheck

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Expected sections in a SKILL.md file (case-insensitive matching)
	expectedSections = map[string]bool{
		"contents":              true,
		"quick start":           true,
		"workflow":              true,
		"scan workflow":         true,
		"audit workflow":        true,
		"analysis workflow":     true,
		"understanding results": true,
		"example output":        true,
		"validation loop":       true,
		"instructions":          true,
		"patterns detected":     true,
	}

	// Dangerous URI schemes
	dangerousSchemes = regexp.MustCompile(`(?i)^(javascript|data|vbscript):`)

	// IP address in URLs
	ipInURL = regexp.MustCompile(`https?://\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)

	// Non-standard ports
	nonStandardPort = regexp.MustCompile(`https?://[^/]+:\d{2,5}[/\s]`)

	// URL shorteners
	urlShorteners = regexp.MustCompile(`(?i)https?://(bit\.ly|tinyurl\.com|t\.co|goo\.gl|ow\.ly|is\.gd|buff\.ly|rebrand\.ly)/`)
)

// ScanStructure performs Layer 2: validates SKILL.md structural integrity.
func ScanStructure(doc *SkillDoc) []Finding {
	var findings []Finding

	findings = append(findings, validateFrontmatter(doc)...)
	findings = append(findings, validateSections(doc)...)
	findings = append(findings, validateLinks(doc)...)

	return findings
}

// validateFrontmatter checks frontmatter for correctness and suspicious content.
func validateFrontmatter(doc *SkillDoc) []Finding {
	var findings []Finding
	fm := doc.Frontmatter

	// Missing frontmatter
	if fm.EndLine == 0 {
		findings = append(findings, Finding{
			Layer:       LayerStructure,
			Severity:    SeverityMedium,
			Section:     "frontmatter",
			Line:        1,
			Rule:        "missing-frontmatter",
			Description: "SKILL.md is missing YAML frontmatter (expected --- delimited block with name and description)",
		})
		return findings
	}

	// Missing name
	if fm.Name == "" {
		findings = append(findings, Finding{
			Layer:       LayerStructure,
			Severity:    SeverityMedium,
			Section:     "frontmatter",
			Line:        fm.StartLine + 1,
			Rule:        "missing-name",
			Description: "Frontmatter is missing required 'name' field",
		})
	}

	// Missing description
	if fm.Description == "" {
		findings = append(findings, Finding{
			Layer:       LayerStructure,
			Severity:    SeverityMedium,
			Section:     "frontmatter",
			Line:        fm.StartLine + 1,
			Rule:        "missing-description",
			Description: "Frontmatter is missing required 'description' field",
		})
	}

	// Extra keys
	if len(fm.ExtraKeys) > 0 {
		findings = append(findings, Finding{
			Layer:       LayerStructure,
			Severity:    SeverityLow,
			Section:     "frontmatter",
			Line:        fm.StartLine + 1,
			Rule:        "extra-frontmatter-keys",
			Description: fmt.Sprintf("Frontmatter contains unexpected keys: %s (expected only name and description)", strings.Join(fm.ExtraKeys, ", ")),
			Evidence:    strings.Join(fm.ExtraKeys, ", "),
		})
	}

	// Description too long
	if len(fm.Description) > 500 {
		findings = append(findings, Finding{
			Layer:       LayerStructure,
			Severity:    SeverityLow,
			Section:     "frontmatter",
			Line:        fm.StartLine + 1,
			Rule:        "long-description",
			Description: fmt.Sprintf("Frontmatter description is unusually long (%d chars, expected < 500)", len(fm.Description)),
		})
	}

	return findings
}

// validateSections checks section structure for unexpected patterns.
func validateSections(doc *SkillDoc) []Finding {
	var findings []Finding

	if len(doc.Sections) == 0 {
		findings = append(findings, Finding{
			Layer:       LayerStructure,
			Severity:    SeverityMedium,
			Section:     "document",
			Line:        1,
			Rule:        "no-sections",
			Description: "SKILL.md contains no markdown sections (expected headings like ## Contents, ## Instructions)",
		})
		return findings
	}

	for _, s := range doc.Sections {
		if s.Level > 2 {
			continue
		}
		titleLower := strings.ToLower(s.Title)
		isExpected := false
		for exp := range expectedSections {
			if strings.Contains(titleLower, exp) {
				isExpected = true
				break
			}
		}
		if !isExpected {
			findings = append(findings, Finding{
				Layer:       LayerStructure,
				Severity:    SeverityInfo,
				Section:     s.Title,
				Line:        s.StartLine + 1,
				Rule:        "unexpected-section",
				Description: fmt.Sprintf("Section '%s' is not a standard SKILL.md section", s.Title),
			})
		}
	}

	return findings
}

// validateLinks checks all links for suspicious patterns.
func validateLinks(doc *SkillDoc) []Finding {
	var findings []Finding

	for _, link := range doc.Links {
		url := link.URL

		if strings.HasPrefix(url, "#") {
			continue
		}

		if !strings.Contains(url, "://") && !strings.HasPrefix(url, "//") {
			continue
		}

		if dangerousSchemes.MatchString(url) {
			findings = append(findings, Finding{
				Layer:       LayerStructure,
				Severity:    SeverityHigh,
				Section:     link.Section,
				Line:        link.Line + 1,
				Rule:        "dangerous-uri-scheme",
				Description: "Link uses a dangerous URI scheme (javascript:, data:, or vbscript:)",
				Evidence:    TruncateEvidence(url),
			})
			continue
		}

		if ipInURL.MatchString(url) {
			findings = append(findings, Finding{
				Layer:       LayerStructure,
				Severity:    SeverityMedium,
				Section:     link.Section,
				Line:        link.Line + 1,
				Rule:        "ip-address-url",
				Description: "Link points to an IP address instead of a domain name",
				Evidence:    TruncateEvidence(url),
			})
		}

		if nonStandardPort.MatchString(url) {
			findings = append(findings, Finding{
				Layer:       LayerStructure,
				Severity:    SeverityLow,
				Section:     link.Section,
				Line:        link.Line + 1,
				Rule:        "non-standard-port",
				Description: "Link uses a non-standard port number",
				Evidence:    TruncateEvidence(url),
			})
		}

		if urlShorteners.MatchString(url) {
			findings = append(findings, Finding{
				Layer:       LayerStructure,
				Severity:    SeverityMedium,
				Section:     link.Section,
				Line:        link.Line + 1,
				Rule:        "url-shortener",
				Description: "Link uses a URL shortener — destination cannot be verified",
				Evidence:    TruncateEvidence(url),
			})
		}

		if strings.Contains(url, "://") && !strings.Contains(url, "github.com") {
			findings = append(findings, Finding{
				Layer:       LayerStructure,
				Severity:    SeverityInfo,
				Section:     link.Section,
				Line:        link.Line + 1,
				Rule:        "external-url",
				Description: "Link points to an external (non-GitHub) URL",
				Evidence:    TruncateEvidence(url),
			})
		}
	}

	return findings
}
