# Fork Maintenance Guide

This is a fork of [anthropics/anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) maintained by Nordlys Labs.

## Remote Configuration

This repository has two remotes configured:
- **origin**: `https://github.com/Nordlys-Labs/anthropic-sdk-go.git` (your fork)
- **upstream**: `https://github.com/anthropics/anthropic-sdk-go.git` (original repo)

## Merging Upstream Changes

To sync with the upstream Anthropic SDK:

```bash
# 1. Revert module paths to upstream format
./scripts/fork-revert

# 2. Fetch and merge upstream changes
git fetch upstream
git merge upstream/main

# 3. Resolve any conflicts

# 4. Re-apply fork module paths
./scripts/fork-apply

# 5. Test the changes
go test ./...

# 6. Commit and push
git add .
git commit -m "Merge upstream changes from anthropics/anthropic-sdk-go"
git push origin main
```

## Scripts

### `./scripts/fork-apply`
Converts all import paths from `github.com/anthropics/anthropic-sdk-go` to `github.com/Nordlys-Labs/anthropic-sdk-go`.

Run this after merging upstream changes.

### `./scripts/fork-revert`
Converts all import paths from `github.com/Nordlys-Labs/anthropic-sdk-go` back to `github.com/anthropics/anthropic-sdk-go`.

Run this before merging upstream changes to minimize merge conflicts.

## What Gets Changed

The fork scripts update:
- `go.mod` module declaration
- `examples/go.mod` module and replace directives
- All Go import statements across the codebase
- Documentation references in README.md and CONTRIBUTING.md

## Why This Approach?

By reverting paths before merging upstream, we minimize merge conflicts since our code structure matches upstream's. After merging, we re-apply our fork's module path.

This makes it easy to:
- Stay up-to-date with upstream improvements
- Contribute changes back to upstream (by reverting before creating PRs)
- Maintain our fork-specific customizations
