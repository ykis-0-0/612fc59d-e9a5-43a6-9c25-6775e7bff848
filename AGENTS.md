# Instructions for Agents

## Architecture
- Use Go version 1.25 to match Traefik upstream

### Project Structure
- Majority of code should be under `realip-zoning/`

### Agent Scope
- High-level overview are and should be in README.md.
- Keep this file focused on implementation guidance for coding agents.
- Ongoing status and open items belong in PROGRESS-for-agents.md.

## Development Guidelines

### General Coding
- Keep changes small and incremental.
- Prefer clear, minimal Go code and standard formatting (gofmt).
- Follow Traefik plugin conventions and middleware patterns.

### Dependencies and Docs
- Avoid adding new dependencies unless necessary.
- If adding docs, keep them concise and focused on plugin behavior.

### LLM Change Fences
- All changes should be made within comment fences (either added or provided) for review purposes.
    - Start the comment fence with `// LLM: <one-liner description of change>`; and
    - Complete with `// LLM-END`.
- Do only one "transformation" or "action" for each fenced change stanza, you may add multiple stanzas for a change requested. You may also split existing fences into smaller ones if deemed required.

### File Naming
- When naming new files, use the pattern `%s%1d%1d_%s.go` with following placeholders:
    - `%s`: "t" for test files, empty otherwise.
    - first `%1d`: one-digit role number, currently used:
        - 0: core plugin interface to Traefik
        - 1: configuration structures, typings, and relevant transformations
        - 3: CIDR fetchers
        - 7: utility functions
    - seconds`%1d`: one-digit one-based incremental number for ordering, except use "0" for base, common, or support code
    - `%s`: concise description
    - Exception: "t00"-series are for integration tests
