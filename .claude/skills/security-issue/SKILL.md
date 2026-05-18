---
name: security-issue
description: Create a security finding in the private peersyst/security repo with proper labels and format. Use when reporting a security vulnerability, concern, or observation found during audits or ad-hoc review.
disable-model-invocation: true
allowed-tools: Bash(gh *) Read Grep
---

# Security Issue Skill

Create a security finding in `peersyst/security`. Gather the following information from the user or from context:

1. **Severity**: One of `critical`, `high`, `medium`, `low`, `info`
2. **Area**: The focus area (e.g., POA module, Authorization & access control, EVM integration, IBC & cross-chain, etc.)
3. **Title**: Short description of the finding
4. **Location**: File path and line numbers
5. **Finding**: Description of the vulnerability or concern
6. **Impact**: What could go wrong and under what conditions
7. **Recommendation**: Suggested fix or mitigation
8. **Evidence**: Code snippets, reasoning, or proof of concept (optional)

If any required field is missing, ask the user before proceeding.

## Before creating

Check for duplicates first:

```bash
gh issue list -R peersyst/security --label "repo:xrplevm/node" --state all --json number,title,labels,state -q '.[] | select(.title | test("KEYWORD"; "i"))'
```

Replace KEYWORD with a relevant term from the title. If a duplicate exists, inform the user and ask whether to proceed.

## Create the issue

```bash
gh issue create -R peersyst/security \
  --title "[SEVERITY_UPPER] Title here" \
  --label "repo:xrplevm/node,SEVERITY,pending" \
  --body "$(cat <<'EOF'
## Area

<focus area>

## Finding

<finding description>

## Location

<file:line>

## Impact

<impact description>

## Recommendation

<recommendation>

## Evidence

<code snippets or reasoning>
EOF
)"
```

## After creating

Report the issue URL back to the user.
