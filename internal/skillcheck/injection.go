package skillcheck

import (
	"fmt"
	"regexp"
	"strings"
)

type injectionPattern struct {
	re          *regexp.Regexp
	rule        string
	description string
	severity    string
}

var (
	// 3a. Direct injection patterns (HIGH severity)
	directInjectionPatterns = []injectionPattern{
		{
			re:          regexp.MustCompile(`(?i)(ignore|disregard|forget|override|bypass)\s+(all\s+)?(previous|prior|original|above|earlier)\s+(instructions?|prompts?|rules?|guidelines?|directives?)`),
			rule:        "injection-override",
			description: "Direct prompt injection: attempts to override previous instructions",
			severity:    SeverityHigh,
		},
		{
			re:          regexp.MustCompile(`(?i)you\s+are\s+now\s+(a|an|in)\s+`),
			rule:        "injection-role-switch",
			description: "Direct prompt injection: attempts to reassign AI role",
			severity:    SeverityHigh,
		},
		{
			re:          regexp.MustCompile(`(?i)(enter|switch\s+to|enable|activate)\s+(developer|admin|system|unrestricted|god|debug|root|sudo)\s*(mode|access)?`),
			rule:        "injection-mode-switch",
			description: "Direct prompt injection: attempts to switch to privileged mode",
			severity:    SeverityHigh,
		},
		{
			re:          regexp.MustCompile(`(?i)(system|administrative|root)\s+(override|access|privilege|escalation)`),
			rule:        "injection-privilege-escalation",
			description: "Direct prompt injection: attempts system/administrative override",
			severity:    SeverityHigh,
		},
		{
			re:          regexp.MustCompile(`(?i)(reveal|show|disclose|display|output|print|dump)\s+(your\s+)?(system\s+)?(prompt|instructions?|configuration|rules?|guidelines?|context)`),
			rule:        "injection-prompt-leak",
			description: "Direct prompt injection: attempts to extract system prompt or configuration",
			severity:    SeverityHigh,
		},
	}

	// 3b. Instruction hierarchy attacks (HIGH in non-Instructions sections)
	hierarchyPatterns = []injectionPattern{
		{
			re:          regexp.MustCompile(`(?i)(the\s+following|this)\s+(overrides?|supersedes?|replaces?)\s+(all|any|previous|prior|the\s+above)`),
			rule:        "injection-hierarchy-override",
			description: "Instruction hierarchy attack: claims to override other instructions",
			severity:    SeverityHigh,
		},
		{
			re:          regexp.MustCompile(`(?i)do\s+not\s+follow\s+(the\s+)?(above|previous|prior|other|original)`),
			rule:        "injection-hierarchy-contradict",
			description: "Instruction hierarchy attack: instructs to ignore other sections",
			severity:    SeverityHigh,
		},
		{
			re:          regexp.MustCompile(`(?i)(actual|real|true|hidden)\s+(instruction|task|goal|purpose|objective)\s+(is|are)\b`),
			rule:        "injection-hidden-intent",
			description: "Instruction hierarchy attack: claims different real intent",
			severity:    SeverityHigh,
		},
	}

	// 3c. Suspicious operations in Instructions (MEDIUM severity)
	suspiciousOps = []injectionPattern{
		{
			re:          regexp.MustCompile(`(?i)\brm\s+-rf\b`),
			rule:        "suspicious-destructive-rm",
			description: "Potentially destructive command: rm -rf",
			severity:    SeverityMedium,
		},
		{
			re:          regexp.MustCompile(`(?i)\bgit\s+push\s+--force\b`),
			rule:        "suspicious-force-push",
			description: "Potentially destructive command: git push --force",
			severity:    SeverityMedium,
		},
		{
			re:          regexp.MustCompile(`(?i)\b(drop|truncate)\s+(table|database|schema)\b`),
			rule:        "suspicious-drop-table",
			description: "Potentially destructive SQL command",
			severity:    SeverityMedium,
		},
		{
			re:          regexp.MustCompile(`(?i)\bdd\s+if=`),
			rule:        "suspicious-dd",
			description: "Potentially destructive command: dd (disk write)",
			severity:    SeverityMedium,
		},
		{
			re:          regexp.MustCompile(`(?i)\b(curl|wget|nc|netcat)\s+[^|]*[^a-zA-Z](--data|--post|-d|-X\s*POST)\b`),
			rule:        "suspicious-data-exfil",
			description: "Potential data exfiltration: HTTP POST to external endpoint",
			severity:    SeverityMedium,
		},
		{
			re:          regexp.MustCompile(`(?i)\b(printenv|\$HOME|\$SSH_KEY|~/\.ssh|~/\.aws|~/\.gnupg)\b`),
			rule:        "suspicious-env-access",
			description: "Suspicious access to environment variables or credential directories",
			severity:    SeverityMedium,
		},
		{
			re:          regexp.MustCompile(`(?i)\b(cat|less|head|tail|type)\s+[^\n]*(\.env|credentials|\.pem|\.key|id_rsa|\.aws/config|password|secret)\b`),
			rule:        "suspicious-credential-read",
			description: "Suspicious attempt to read credential or secret files",
			severity:    SeverityMedium,
		},
	}

	// 3d. Hidden text patterns (HIGH severity)
	hiddenTextPatterns = []injectionPattern{
		{
			re:          regexp.MustCompile(`<!--[\s\S]*?-->`),
			rule:        "hidden-html-comment",
			description: "Hidden HTML comment — may contain concealed instructions",
			severity:    SeverityHigh,
		},
		{
			re:          regexp.MustCompile(`(?i)font-size\s*:\s*0`),
			rule:        "hidden-zero-font",
			description: "Zero font-size CSS — text invisible to human readers",
			severity:    SeverityHigh,
		},
		{
			re:          regexp.MustCompile(`(?i)(color\s*:\s*(white|#fff(fff)?|rgb\(\s*255\s*,\s*255\s*,\s*255\s*\))|opacity\s*:\s*0\b|display\s*:\s*none)`),
			rule:        "hidden-invisible-css",
			description: "CSS that hides text (white text, zero opacity, or display:none)",
			severity:    SeverityHigh,
		},
	}

	// 3e. Encoding payloads (MEDIUM severity outside code blocks)
	encodingPatterns = []injectionPattern{
		{
			re:          regexp.MustCompile(`[A-Za-z0-9+/]{40,}={0,2}`),
			rule:        "encoding-base64",
			description: "Large Base64-encoded blob detected — may hide malicious content",
			severity:    SeverityMedium,
		},
		{
			re:          regexp.MustCompile(`(\\x[0-9a-fA-F]{2}){4,}`),
			rule:        "encoding-hex",
			description: "Hex escape sequence detected — may hide malicious content",
			severity:    SeverityMedium,
		},
		{
			re:          regexp.MustCompile(`(\\u[0-9a-fA-F]{4}){3,}`),
			rule:        "encoding-unicode-escape",
			description: "Unicode escape sequence detected — may hide malicious content",
			severity:    SeverityMedium,
		},
	}
)

// ScanInjection performs Layer 3: pattern-based prompt injection detection.
func ScanInjection(doc *SkillDoc) []Finding {
	var findings []Finding

	for lineNum, line := range doc.Lines {
		if strings.TrimSpace(line) == "---" {
			continue
		}

		section := SectionForLine(doc, lineNum)
		inCodeBlock := IsInCodeBlock(doc, lineNum)
		sectionLower := strings.ToLower(section)
		isExampleCodeBlock := inCodeBlock && strings.Contains(sectionLower, "example")

		// 3a. Direct injection patterns
		if !isExampleCodeBlock {
			for _, p := range directInjectionPatterns {
				if p.re.MatchString(line) {
					severity := p.severity
					severity = adjustSeverity(severity, section, inCodeBlock)
					findings = append(findings, Finding{
						Layer:       LayerInjection,
						Severity:    severity,
						Section:     section,
						Line:        lineNum + 1,
						Rule:        p.rule,
						Description: p.description,
						Evidence:    TruncateEvidence(strings.TrimSpace(line)),
					})
				}
			}
		}

		// 3b. Hierarchy attacks
		if !isExampleCodeBlock {
			for _, p := range hierarchyPatterns {
				if p.re.MatchString(line) {
					severity := p.severity
					if strings.EqualFold(section, "Instructions") {
						severity = SeverityMedium
					}
					severity = adjustSeverity(severity, section, inCodeBlock)
					findings = append(findings, Finding{
						Layer:       LayerInjection,
						Severity:    severity,
						Section:     section,
						Line:        lineNum + 1,
						Rule:        p.rule,
						Description: p.description,
						Evidence:    TruncateEvidence(strings.TrimSpace(line)),
					})
				}
			}
		}

		// 3c. Suspicious operations
		if !inCodeBlock {
			for _, p := range suspiciousOps {
				if p.re.MatchString(line) {
					findings = append(findings, Finding{
						Layer:       LayerInjection,
						Severity:    p.severity,
						Section:     section,
						Line:        lineNum + 1,
						Rule:        p.rule,
						Description: p.description,
						Evidence:    TruncateEvidence(strings.TrimSpace(line)),
					})
				}
			}
		}

		// 3d. Hidden text patterns
		for _, p := range hiddenTextPatterns {
			if p.re.MatchString(line) {
				findings = append(findings, Finding{
					Layer:       LayerInjection,
					Severity:    p.severity,
					Section:     section,
					Line:        lineNum + 1,
					Rule:        p.rule,
					Description: p.description,
					Evidence:    TruncateEvidence(strings.TrimSpace(line)),
				})
			}
		}

		// 3e. Encoding payloads
		if !inCodeBlock {
			for _, p := range encodingPatterns {
				if p.re.MatchString(line) {
					findings = append(findings, Finding{
						Layer:       LayerInjection,
						Severity:    p.severity,
						Section:     section,
						Line:        lineNum + 1,
						Rule:        p.rule,
						Description: p.description,
						Evidence:    TruncateEvidence(strings.TrimSpace(line)),
					})
				}
			}
		}
	}

	return findings
}

// adjustSeverity applies section-aware severity modifiers.
func adjustSeverity(severity string, section string, inCodeBlock bool) string {
	sectionLower := strings.ToLower(section)

	if sectionLower == "frontmatter" {
		return raiseSeverity(severity)
	}

	if strings.Contains(sectionLower, "example") && inCodeBlock {
		return lowerSeverity(severity)
	}

	return severity
}

func raiseSeverity(severity string) string {
	switch severity {
	case SeverityInfo:
		return SeverityLow
	case SeverityLow:
		return SeverityMedium
	case SeverityMedium:
		return SeverityHigh
	default:
		return severity
	}
}

func lowerSeverity(severity string) string {
	switch severity {
	case SeverityHigh:
		return SeverityMedium
	case SeverityMedium:
		return SeverityLow
	case SeverityLow:
		return SeverityInfo
	default:
		return severity
	}
}

// MultiLineHTMLCommentScan detects HTML comments that span multiple lines.
func MultiLineHTMLCommentScan(doc *SkillDoc) []Finding {
	var findings []Finding
	content := strings.Join(doc.Lines, "\n")
	re := regexp.MustCompile(`<!--[\s\S]*?-->`)

	for _, match := range re.FindAllStringIndex(content, -1) {
		startLine := strings.Count(content[:match[0]], "\n")
		section := SectionForLine(doc, startLine)
		comment := content[match[0]:match[1]]

		if strings.Contains(comment, "\n") {
			findings = append(findings, Finding{
				Layer:       LayerInjection,
				Severity:    SeverityHigh,
				Section:     section,
				Line:        startLine + 1,
				Rule:        "hidden-html-comment-multiline",
				Description: fmt.Sprintf("Multi-line HTML comment (%d lines) — may contain concealed instructions", strings.Count(comment, "\n")+1),
				Evidence:    TruncateEvidence(comment),
			})
		}
	}

	return findings
}
