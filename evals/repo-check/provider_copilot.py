import subprocess
import os
import re

# Model for skill execution queries
MODEL = os.environ.get("COPILOT_MODEL", "gpt-5.2-codex")


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


def _parse_usage(raw_output):
    """
    Separate Copilot response content from the trailing usage stats block.

    Returns (content, metadata_dict, token_usage_dict).

    Expected usage block format:
        Total usage est: 1 Premium request
        API time spent:  2s
        Total session time: 9s
        Total code changes: 0+ 0-
        Breakdown by AI model:
          gpt-5.2-codex    14.3k in, 21 out, 1.7k cached ( Est. 1 Premium request)
    """
    metadata = {}
    token_usage = {}

    # Find the start of the usage block
    marker_idx = raw_output.find("Total usage est:")
    if marker_idx == -1:
        return raw_output, metadata, token_usage

    content = raw_output[:marker_idx].rstrip()
    usage_block = raw_output[marker_idx:]

    # Total usage est
    m = re.search(r"Total usage est:\s*(.+)", usage_block)
    if m:
        metadata["total_usage_est"] = m.group(1).strip()

    # API time spent
    m = re.search(r"API time spent:\s*(\S+)", usage_block)
    if m:
        metadata["api_time_seconds"] = _parse_seconds(m.group(1))

    # Total session time
    m = re.search(r"Total session time:\s*(\S+)", usage_block)
    if m:
        metadata["session_time_seconds"] = _parse_seconds(m.group(1))

    # Total code changes
    m = re.search(r"Total code changes:\s*(\d+)\+\s*(\d+)-", usage_block)
    if m:
        metadata["code_additions"] = int(m.group(1))
        metadata["code_deletions"] = int(m.group(2))

    # Model breakdown lines:
    #   gpt-5.2-codex    14.3k in, 21 out, 1.7k cached ( Est. 1 Premium request)
    model_pattern = re.compile(
        r"^\s+(\S+)\s+([\d.]+[km]?)\s+in,\s+([\d.]+[km]?)\s+out,\s+([\d.]+[km]?)\s+cached",
        re.MULTILINE | re.IGNORECASE,
    )
    model_lines = model_pattern.findall(usage_block)

    models_breakdown = []
    for model_name, t_in, t_out, t_cached in model_lines:
        prompt_tokens = _parse_token_count(t_in)
        completion_tokens = _parse_token_count(t_out)
        cached_tokens = _parse_token_count(t_cached)
        models_breakdown.append(
            {
                "model": model_name,
                "tokens_in": prompt_tokens,
                "tokens_out": completion_tokens,
                "tokens_cached": cached_tokens,
            }
        )

    if models_breakdown:
        # Primary model stats go into promptfoo's tokenUsage
        primary = models_breakdown[0]
        token_usage = {
            "prompt": primary["tokens_in"],
            "completion": primary["tokens_out"],
            "cached": primary["tokens_cached"],
            "total": primary["tokens_in"] + primary["tokens_out"],
        }
        metadata["model"] = primary["model"]

        # Include full breakdown when multiple models are used
        if len(models_breakdown) > 1:
            metadata["model_breakdown"] = models_breakdown

    return content, metadata, token_usage


def call_api(prompt, options, context):
    """
    Invoke GitHub Copilot CLI with the test query.
    Parses usage statistics from output and returns them as structured metadata
    so they appear in the promptfoo results table.
    """
    try:
        # Option A: gh copilot suggest with model flag (if supported)
        cmd = ["gh", "copilot", "suggest", "-t", "shell", "--model", MODEL, prompt]

        # Option B: If your setup uses `gh models run` instead, swap to:
        # cmd = ["gh", "models", "run", MODEL, prompt]

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=120,
            cwd=os.environ.get("TEST_CWD", os.getcwd()),
        )

        raw_output = result.stdout + result.stderr

        if result.returncode != 0 and not raw_output.strip():
            return {"error": f"gh copilot exited with code {result.returncode}"}

        content, metadata, token_usage = _parse_usage(raw_output)

        response = {"output": content}
        if token_usage:
            response["tokenUsage"] = token_usage
        if metadata:
            response["metadata"] = metadata

        return response

    except subprocess.TimeoutExpired:
        return {"error": "gh copilot timed out after 120s"}
    except FileNotFoundError:
        return {
            "error": "gh copilot CLI not found. Install with: gh extension install github/gh-copilot"
        }
    except Exception as e:
        return {"error": str(e)}
