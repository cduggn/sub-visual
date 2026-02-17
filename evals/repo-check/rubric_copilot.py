import subprocess
import os
import re

# Model for rubric judging — lighter model is fine for pass/fail evaluation
MODEL = os.environ.get("COPILOT_RUBRIC_MODEL", "gpt-5-mini")


def _parse_token_count(s):
    """Convert token counts like '14.3k' or '2.1m' to integers."""
    s = s.strip()
    if s.lower().endswith("k"):
        return int(float(s[:-1]) * 1000)
    if s.lower().endswith("m"):
        return int(float(s[:-1]) * 1000000)
    return int(float(s))


def _parse_seconds(s):
    """Parse duration strings like '2s', '1m30s', '150ms' to float seconds."""
    s = s.strip()
    if s.endswith("ms"):
        return float(s[:-2]) / 1000
    m = re.match(r"(?:(\d+)m)?(\d+)s", s)
    if m:
        mins = int(m.group(1) or 0)
        secs = int(m.group(2))
        return mins * 60 + secs
    return float(re.sub(r"[^\d.]", "", s)) if s else 0


def _strip_usage_block(raw_output):
    """
    Remove the trailing usage stats block from CLI output.
    Returns just the response content for rubric evaluation.
    """
    marker_idx = raw_output.find("Total usage est:")
    if marker_idx == -1:
        return raw_output
    return raw_output[:marker_idx].rstrip()


def call_api(prompt, options, context):
    """
    Rubric judge provider — routes llm-rubric evaluation through gh copilot
    using gpt-5-mini. Receives a judging prompt from promptfoo and returns
    Copilot's pass/fail assessment.

    Usage stats are stripped from the output so the rubric judge response
    is clean for promptfoo's pass/fail parsing.
    """
    try:
        # Option A: gh copilot explain with model flag (if supported)
        cmd = ["gh", "copilot", "explain", "--model", MODEL, prompt]

        # Option B: If your setup uses `gh models run` instead, swap to:
        # cmd = ["gh", "models", "run", MODEL, prompt]

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=60,
            cwd=os.environ.get("TEST_CWD", os.getcwd()),
        )

        raw_output = result.stdout + result.stderr

        if result.returncode != 0 and not raw_output.strip():
            return {"error": f"Rubric judge failed with code {result.returncode}"}

        content = _strip_usage_block(raw_output)
        return {"output": content}

    except subprocess.TimeoutExpired:
        return {"error": "Rubric judge timed out"}
    except Exception as e:
        return {"error": str(e)}
