package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/cduggn/sub-visual/internal/github"
	"github.com/cduggn/sub-visual/internal/skillcheck"
)

func main() {
	jsonFlag := flag.Bool("json", false, "output as JSON instead of markdown")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: skill-check [--json] <target>\n\n")
		fmt.Fprintf(os.Stderr, "  target: path to a SKILL.md file, or a GitHub blob URL\n")
		fmt.Fprintf(os.Stderr, "  --json: output as JSON instead of markdown\n")
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	target := flag.Arg(0)
	var content string
	var displayPath string

	if isURL(target) {
		// Parse the GitHub URL
		gt, err := github.ParseTarget(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if !gt.IsBlob {
			fmt.Fprintf(os.Stderr, "Error: URL must be a GitHub blob URL pointing to a specific file\n")
			fmt.Fprintf(os.Stderr, "Example: https://github.com/owner/repo/blob/main/SKILL.md\n")
			os.Exit(1)
		}

		// Check that gh CLI is available for API calls
		if err := github.CheckCLI(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(2)
		}

		// Fetch the file content from GitHub
		fetched, err := github.FetchFileContent(gt.Owner, gt.Repo, gt.Branch, gt.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not fetch file — %v\n", err)
			os.Exit(3)
		}
		content = fetched
		displayPath = target
	} else {
		// Read from local file
		data, err := os.ReadFile(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not read file — %v\n", err)
			os.Exit(1)
		}
		content = string(data)
		displayPath = target
	}

	// Run analysis
	report := skillcheck.Analyze(content, displayPath)

	// Output
	if *jsonFlag {
		out, err := skillcheck.FormatJSON(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)
	} else {
		fmt.Print(skillcheck.FormatMarkdown(report))
	}

	// Exit with non-zero status for dangerous findings
	if report.Signal == skillcheck.SignalDangerous {
		os.Exit(2)
	}
}

// isURL returns true if the target looks like a URL.
func isURL(target string) bool {
	return strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://")
}
