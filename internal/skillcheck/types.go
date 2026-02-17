package skillcheck

// SkillReport is the top-level output of a SKILL.md security analysis.
type SkillReport struct {
	File         string    `json:"file"`
	Signal       string    `json:"signal"`
	FindingCount int       `json:"finding_count"`
	Findings     []Finding `json:"findings"`
	Sections     []string  `json:"sections"`
}

// Finding represents a single security concern detected in the SKILL.md file.
type Finding struct {
	Layer       string `json:"layer"`
	Severity    string `json:"severity"`
	Section     string `json:"section"`
	Line        int    `json:"line"`
	Rule        string `json:"rule"`
	Description string `json:"description"`
	Evidence    string `json:"evidence"`
}

// SkillDoc represents a parsed SKILL.md file.
type SkillDoc struct {
	Frontmatter    Frontmatter
	Sections       []Section
	CodeBlocks     []CodeBlock
	Links          []Link
	Lines          []string // raw lines for line-number references
	RawFrontmatter string   // raw frontmatter text (for scanning)
}

// Frontmatter holds the YAML frontmatter key-value pairs.
type Frontmatter struct {
	Name        string
	Description string
	ExtraKeys   []string // keys beyond name and description
	StartLine   int
	EndLine     int
}

// Section represents a markdown heading and its content.
type Section struct {
	Title     string
	Level     int    // 1 for #, 2 for ##
	StartLine int
	EndLine   int
	Content   string // text content (excluding code blocks)
}

// CodeBlock represents a fenced code block.
type CodeBlock struct {
	Language  string
	Content   string
	Section   string // which section this block belongs to
	StartLine int
	EndLine   int
}

// Link represents a markdown link or image reference.
type Link struct {
	Text    string
	URL     string
	IsImage bool
	Section string
	Line    int
}

// Severity constants
const (
	SeverityHigh   = "HIGH"
	SeverityMedium = "MEDIUM"
	SeverityLow    = "LOW"
	SeverityInfo   = "INFO"
)

// Signal constants
const (
	SignalSafe       = "SAFE"
	SignalSuspicious = "SUSPICIOUS"
	SignalDangerous  = "DANGEROUS"
)

// Layer constants
const (
	LayerUnicode   = "unicode"
	LayerStructure = "structure"
	LayerInjection = "injection"
	LayerCoherence = "coherence"
)
