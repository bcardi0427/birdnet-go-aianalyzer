# Implementation Checklist - Configuration & LLM Settings Fixes

## 1. Setup & Baseline Verification
- [x] Run baseline unit tests to verify existing behavior:
  - [x] `go test -v -run TestClone ./internal/conf/...`
  - [x] `go test -v ./internal/api/v2/...`

## 2. Configuration & Storage Hardening
- [x] Modify `persistMigration` in `internal/conf/migrations.go` to clone settings and encrypt secrets before saving.
- [x] Modify `ensureSessionSecret` in `internal/conf/migrations.go` to clone settings and encrypt secrets before saving.
- [x] Verify fix by running existing unit tests.

## 3. LLM Provider Reversion Fix
- [x] Remove duplicate `AI` field from `RealtimeSettings` struct in `internal/conf/config.go`.
- [x] Verify fix by compiling and running tests.

## 4. Verification & Testing
- [x] Create/run custom integration test to verify secrets are saved encrypted on disk (i.e. prefixed with `enc:v1:`).
- [x] Verify settings endpoint and encryption together under `internal/api/v2/...`.
