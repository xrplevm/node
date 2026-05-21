---
name: security-audit
description: Run a full security audit across all focus areas in parallel, reporting findings to peersyst/security
disable-model-invocation: true
allowed-tools: Bash(gh *) Read Grep Glob Agent
---

# Security Audit

Read `.claude/security/SECURITY_SCOPE.md` and `.claude/security/THREAT_MODEL.md`. Then spawn one `security-reviewer` agent per focus area defined in the "Audit Focus Areas" table, running them all in parallel.

Each agent should:

1. Query existing issues first:
   ```bash
   gh issue list -R peersyst/security --label "repo:xrplevm/node" --state all --json number,title,labels,state
   ```
2. Audit its assigned area by reading the components listed in scope
3. For each new finding, create an issue in `peersyst/security` following the reporting workflow in SECURITY_SCOPE.md (labels: `repo:xrplevm/node`, severity, `pending`; area goes in issue body)
4. Skip any finding that already exists as `pending`, `accepted`, or `discarded`
5. For closed (done) findings, verify the fix is still in place — reopen only if regressed

After all agents complete, summarize the total findings created per area.
