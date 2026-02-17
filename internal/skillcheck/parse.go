package skillcheck

import (
	"regexp"
	"strings"
)

var (
	headingRe = regexp.MustCompile(`^(#{1,2})\s+(.+)$`)
	fenceRe   = regexp.MustCompile("^```(\\w*)\\s*$")
	linkRe    = regexp.MustCompile(`(!?)\[([^\]]*)\]\(([^)]+)\)`)
)

// ParseSkillDoc parses a SKILL.md file into a structured SkillDoc.
func ParseSkillDoc(content string) *SkillDoc {
	lines := strings.Split(content, "\n")
	doc := &SkillDoc{
		Lines: lines,
	}

	// Parse frontmatter
	doc.Frontmatter = parseFrontmatter(lines)
	if doc.Frontmatter.EndLine > 0 {
		doc.RawFrontmatter = strings.Join(lines[doc.Frontmatter.StartLine:doc.Frontmatter.EndLine+1], "\n")
	}

	// Parse sections, code blocks, and links
	parseSections(doc, lines)
	return doc
}

// parseFrontmatter extracts YAML frontmatter delimited by --- lines.
func parseFrontmatter(lines []string) Frontmatter {
	fm := Frontmatter{}
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fm
	}

	fm.StartLine = 0
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			fm.EndLine = i
			break
		}
	}
	if fm.EndLine == 0 {
		return fm
	}

	// Parse key-value pairs (simple YAML: key: value)
	for i := fm.StartLine + 1; i < fm.EndLine; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			fm.Name = value
		case "description":
			fm.Description = value
		default:
			fm.ExtraKeys = append(fm.ExtraKeys, key)
		}
	}

	return fm
}

// parseSections extracts headings, code blocks, and links from the document.
func parseSections(doc *SkillDoc, lines []string) {
	startLine := 0
	if doc.Frontmatter.EndLine > 0 {
		startLine = doc.Frontmatter.EndLine + 1
	}

	var currentSection *Section
	inCodeBlock := false
	var currentCodeBlock *CodeBlock

	for i := startLine; i < len(lines); i++ {
		line := lines[i]

		// Handle code block boundaries
		if m := fenceRe.FindStringSubmatch(line); m != nil {
			if !inCodeBlock {
				// Opening fence
				inCodeBlock = true
				sectionName := ""
				if currentSection != nil {
					sectionName = currentSection.Title
				}
				currentCodeBlock = &CodeBlock{
					Language:  m[1],
					Section:   sectionName,
					StartLine: i,
				}
				continue
			} else {
				// Closing fence
				inCodeBlock = false
				currentCodeBlock.EndLine = i
				doc.CodeBlocks = append(doc.CodeBlocks, *currentCodeBlock)
				currentCodeBlock = nil
				continue
			}
		}

		if inCodeBlock {
			if currentCodeBlock != nil {
				if currentCodeBlock.Content != "" {
					currentCodeBlock.Content += "\n"
				}
				currentCodeBlock.Content += line
			}
			continue
		}

		// Check for headings
		if m := headingRe.FindStringSubmatch(line); m != nil {
			// Close previous section
			if currentSection != nil {
				currentSection.EndLine = i - 1
				doc.Sections = append(doc.Sections, *currentSection)
			}

			level := len(m[1])
			currentSection = &Section{
				Title:     m[2],
				Level:     level,
				StartLine: i,
			}
			continue
		}

		// Accumulate content for current section
		if currentSection != nil {
			if currentSection.Content != "" {
				currentSection.Content += "\n"
			}
			currentSection.Content += line
		}

		// Extract links
		sectionName := ""
		if currentSection != nil {
			sectionName = currentSection.Title
		}
		for _, match := range linkRe.FindAllStringSubmatch(line, -1) {
			doc.Links = append(doc.Links, Link{
				Text:    match[2],
				URL:     match[3],
				IsImage: match[1] == "!",
				Section: sectionName,
				Line:    i,
			})
		}
	}

	// Close final section
	if currentSection != nil {
		currentSection.EndLine = len(lines) - 1
		doc.Sections = append(doc.Sections, *currentSection)
	}
}

// SectionNames returns the list of section titles from the parsed document.
func SectionNames(doc *SkillDoc) []string {
	names := make([]string, len(doc.Sections))
	for i, s := range doc.Sections {
		names[i] = s.Title
	}
	return names
}

// FindSection returns the first section matching the given title (case-insensitive).
func FindSection(doc *SkillDoc, title string) *Section {
	lower := strings.ToLower(title)
	for i := range doc.Sections {
		if strings.ToLower(doc.Sections[i].Title) == lower {
			return &doc.Sections[i]
		}
	}
	return nil
}

// InstructionsSection returns the Instructions section, or nil if not found.
func InstructionsSection(doc *SkillDoc) *Section {
	return FindSection(doc, "Instructions")
}

// CodeBlocksInSection returns all code blocks belonging to the named section.
func CodeBlocksInSection(doc *SkillDoc, sectionTitle string) []CodeBlock {
	lower := strings.ToLower(sectionTitle)
	var blocks []CodeBlock
	for _, cb := range doc.CodeBlocks {
		if strings.ToLower(cb.Section) == lower {
			blocks = append(blocks, cb)
		}
	}
	return blocks
}

// SectionForLine returns the section name for a given line number.
func SectionForLine(doc *SkillDoc, line int) string {
	if line >= doc.Frontmatter.StartLine && line <= doc.Frontmatter.EndLine {
		return "frontmatter"
	}
	for _, s := range doc.Sections {
		if line >= s.StartLine && line <= s.EndLine {
			return s.Title
		}
	}
	return "unknown"
}

// IsInCodeBlock returns true if the given line number falls within a code block.
func IsInCodeBlock(doc *SkillDoc, line int) bool {
	for _, cb := range doc.CodeBlocks {
		if line >= cb.StartLine && line <= cb.EndLine {
			return true
		}
	}
	return false
}
