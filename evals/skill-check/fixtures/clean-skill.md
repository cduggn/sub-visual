---
name: formatting-code
description: Formats source code files using standard formatters for Python, Go, and JavaScript. Use when you want to auto-format code before committing.
---

# Code Formatter

## Contents
- [Quick start](#quick-start)
- [Workflow](#workflow)
- [Understanding results](#understanding-results)
- [Instructions](#instructions)

## Quick Start

```
/format [--check] [--ecosystem python|go|js]
```

## Workflow

```
Code Formatting:
- [ ] Step 1: Detect ecosystems
- [ ] Step 2: Install formatters if missing
- [ ] Step 3: Run formatter
- [ ] Step 4: Review changes
- [ ] Step 5: Commit formatted code
```

## Understanding Results

| Status | Meaning |
|--------|---------|
| Formatted | Files were modified to match style |
| Clean | Files already match style |
| Error | Formatter encountered issues |

## Instructions

When the user invokes this skill:

1. Detect which ecosystems exist by checking for source files:
   ```bash
   find . -name "*.py" -not -path "*/venv/*" | head -5
   find . -name "*.go" | head -5
   find . -name "*.js" -not -path "*/node_modules/*" | head -5
   ```

2. Install missing formatters:
   ```bash
   which black || pip install black
   which gofmt || echo "gofmt ships with Go"
   which prettier || npm install -g prettier
   ```

3. Run the appropriate formatter:
   ```bash
   black .
   gofmt -w .
   prettier --write "**/*.js"
   ```

4. Present the results showing which files were formatted.

5. Suggest committing the changes if any files were modified.
