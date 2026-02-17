package skillcheck

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Analyze is the main entry point for SKILL.md security analysis.
// It parses the content, runs all 4 detection layers, and returns a report.
func Analyze(content, filePath string) *SkillReport {
	doc := ParseSkillDoc(content)

	var findings []Finding
	findings = append(findings, ScanUnicode(doc)...)
	findings = append(findings, ScanStructure(doc)...)
	findings = append(findings, ScanInjection(doc)...)
	findings = append(findings, MultiLineHTMLCommentScan(doc)...)
	findings = append(findings, ScanCoherence(doc)...)

	return &SkillReport{
		File:         filePath,
		Signal:       ComputeSignal(findings),
		FindingCount: len(findings),
		Findings:     findings,
		Sections:     SectionNames(doc),
	}
}

// FormatJSON returns the report as indented JSON.
func FormatJSON(report *SkillReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatMarkdown returns the report as a markdown string.
func FormatMarkdown(report *SkillReport) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# SKILL.md Security Report: %s\n\n", report.File))
	b.WriteString(fmt.Sprintf("## Signal: %s\n\n", report.Signal))

	if report.FindingCount == 0 {
		b.WriteString("No security concerns detected.\n\n")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("**Findings:** %d\n\n", report.FindingCount))

	// Group findings by severity
	severities := []string{SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}
	for _, sev := range severities {
		var sevFindings []Finding
		for _, f := range report.Findings {
			if f.Severity == sev {
				sevFindings = append(sevFindings, f)
			}
		}
		if len(sevFindings) == 0 {
			continue
		}

		b.WriteString(fmt.Sprintf("### %s Severity (%d)\n\n", sev, len(sevFindings)))

		for i, f := range sevFindings {
			b.WriteString(fmt.Sprintf("**%d. [%s] %s**\n", i+1, f.Layer, f.Rule))
			b.WriteString(fmt.Sprintf("- Line %d, Section: %s\n", f.Line, f.Section))
			b.WriteString(fmt.Sprintf("- %s\n", f.Description))
			if f.Evidence != "" {
				b.WriteString(fmt.Sprintf("- Evidence: `%s`\n", f.Evidence))
			}
			b.WriteString("\n")
		}
	}

	// Parsed sections
	if len(report.Sections) > 0 {
		b.WriteString("### Parsed Sections\n\n")
		for _, s := range report.Sections {
			b.WriteString(fmt.Sprintf("- %s\n", s))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// ComputeSignal determines the composite signal from findings.
func ComputeSignal(findings []Finding) string {
	for _, f := range findings {
		if f.Severity == SeverityHigh {
			return SignalDangerous
		}
	}
	for _, f := range findings {
		if f.Severity == SeverityMedium {
			return SignalSuspicious
		}
	}
	return SignalSafe
}
