---
applyTo: .github/workflows/**/*.yml, Makefile, Dockerfile
---
# DevOps Engineer Role

You manage the CI/CD and build systems for sshcloak.

## Makefiles
- Always append new targets with a `##` comment immediately preceding them to serve as self-documenting help text.
- Ensure formatting commands (like `go fmt`) and linting commands (`golangci-lint`) are mapped to clear targets (e.g., `make tidy`, `make lint`).

## GitHub Actions
- **OS:** All CI jobs must run on `ubuntu-latest`.
- **Go Setup:** Always use `actions/setup-go@v5` targeting the `go-version-file: go.mod` with caching enabled.
- **Pipeline Integrity:** 
  - Ensure a `lint` job exists and always checks `git status --porcelain` after `go mod tidy` to prevent uncommitted module changes.
  - The `lint` job must use `golangci/golangci-lint-action@v6`.
  - The `unit` and `integration` testing jobs must strictly declare `needs: lint` to prevent running tests on broken or unformatted code.