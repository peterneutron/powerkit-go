# AGENTS: Working Guide for `powerkit-go`

This file is the source of truth for contributors/agents working in this repo.

## Scope and Intent
- Repo: `powerkit-go` only.
- Goal: provide a stable, safe Darwin power API and CLI.
- Priority order:
  1. Safety and correctness
  2. Backward compatibility for exported APIs
  3. Performance/ergonomics

## Project Layout
- `pkg/powerkit/`
  - Public API surface. Any exported behavior changes here are effectively user-facing changes.
- `internal/smc/`
  - SMC cgo bridge and raw read/write implementation.
- `internal/iokit/`
  - IOKit polling/streaming.
- `internal/os/`, `internal/powerd/`
  - OS helpers and assertion internals.
- `cmd/powerkit-cli/`
  - CLI split by command domains.

## Non-Negotiable Safety Rules
- Do not remove SMC data-size guardrails (`<= 32`) in `internal/smc/smc.go`.
- Do not reintroduce deferred C frees inside per-key loops in hot paths.
- Do not bypass root checks for mutating APIs.
- Preserve typed error semantics (`ErrPermissionRequired`, `ErrNotSupported`, `ErrTransientIO`) when extending behavior.

## Build and Verification
Run from repo root:
- `make tests`
- `make vet`
- `make lint`
- `make verify`

`make verify` must pass before proposing completion.

## API Design Rules
- New public APIs must be added under `pkg/powerkit`.
- Prefer additive changes; avoid breaking/removing exports unless explicitly requested.
- For new mutating APIs, provide context-aware variants (`FooContext`).
- Use typed errors for caller-actionable outcomes.

## CLI Rules
- Keep command parsing simple and split by domain file.
- Any new command must update:
  - `cmd/powerkit-cli/main.go` dispatch
  - help text
  - tests (if behavior is non-trivial)

## Testing Guidance
- Prefer table-driven tests.
- Add regression coverage when fixing bugs.
- High-value targets:
  - SMC raw size/decoder behavior
  - context cancellation paths
  - permission gating behavior

## Common Pitfalls
- This repo is Darwin-only; non-macOS build errors are expected.
- Some test/build paths require cgo + Xcode CLT.
- Lint failures are often gofmt/goimports/build-tag hygiene issues.

## Commit Guidance
- Use Conventional Commit style.
- Keep commits scoped and reviewable.
- Separate refactors from behavior changes when possible.
