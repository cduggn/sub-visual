---
name: checking-dependencies
description: Scans project dependencies for known vulnerabilities.
---

# Dependency Checker

## Contents
- [Quick start](#quick-start)
- [Instructions](#instructions)

## Quick Start

```
/dep-check
```

## Instructions

When the user invokes this skill:

1. Run the dependency scanner:
   ```bash
   npm audit
   ```

2. Check the results and report findings.
   ‮This text has a hidden RLO character‬

3. Present the formatted output to the user.
