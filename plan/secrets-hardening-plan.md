# Secrets Hardening Implementation Checklist

This checklist tracks implementation of env-first secrets hardening for BirdNET-Go, based on the original plan in this file.

Date: 2026-05-18

## Core Security Rules

- [ ] Enforce secret resolution order: `env > secret-file > plaintext(legacy)`.
- [ ] Treat `config.yaml` as reference storage, not secret-value storage, in recommended flows.
- [ ] Keep backward compatibility while warning on plaintext usage.
- [ ] Never expose resolved secret values in logs, API responses, or UI payloads.

---

## Phase 1: Inventory and Secret Map

- [ ] Inventory all secret-bearing config fields.
- [ ] Confirm AI key coverage (`ai.apiKey`).
- [ ] Confirm eBird key coverage (`realtime.ebird.apiKey`).
- [ ] Confirm weather provider key coverage (`realtime.weather.*.apiKey`).
- [ ] Confirm existing secret-enabled paths (webhook auth and others) remain compatible.
- [ ] Produce secret field matrix:
  - [ ] Config path
  - [ ] Current persistence behavior
  - [ ] Target resolver behavior

---

## Phase 2: Config and Validation Extensions

- [ ] Add missing `...File` config fields (YAML + JSON tags) where needed.
- [ ] Add `ai.apiKeyFile` (if missing).
- [ ] Add `realtime.ebird.apiKeyFile` (if missing).
- [ ] Add `realtime.weather.openWeather.apiKeyFile` (if missing).
- [ ] Add `realtime.weather.wunderground.apiKeyFile` (if missing).
- [ ] Validate enabled features require at least one key source:
  - [ ] Inline value
  - [ ] File reference
  - [ ] Env reference via `${VAR}`
- [ ] Keep validation errors explicit and user-actionable.

---

## Phase 3: Runtime Secret Resolution Wiring

- [ ] Wire all target components through `internal/secrets` resolver pattern.
- [ ] Use `secrets.Resolve` / `secrets.MustResolve` at credential use points.
- [ ] Preserve file-over-value precedence.
- [ ] Preserve env expansion for `${VAR}` and `${VAR:-default}` patterns.
- [ ] Keep plaintext fallback for compatibility only.
- [ ] Emit warning when plaintext fallback is used.
- [ ] Ensure diagnostics log source class only (`env`, `file`, `plaintext`) and never value content.

---

## Phase 4: Settings Persistence Controls

- [ ] Preserve submitted secret references (`${ENV_VAR}` and `...File`) without resolving/writing plaintext.
- [ ] Prevent resolved runtime secrets from being persisted back into config.
- [ ] If plaintext secret is submitted, allow for v1 compatibility and emit migration warning.
- [ ] Add guidance text in responses/logs for migration to env/file references.
- [ ] Keep strict plaintext-reject mode as future optional toggle (not required for v1).

---

## Phase 5: API/UI Behavior and Redaction

### Backend

- [ ] Ensure settings APIs never return raw secret values.
- [ ] Preserve/redouble-check redaction semantics and key round-trip behavior.
- [ ] Confirm support output/snapshots exclude secret values.

### Frontend

- [ ] Add/verify guidance encouraging env vars and secret files.
- [ ] Keep API key fields non-destructive with placeholder/redacted UX.
- [ ] Ensure saving settings does not accidentally replace secret references with plaintext.

---

## Phase 6: Observability and Safe Warnings

- [ ] Add warning events when plaintext secrets are detected.
- [ ] Include only field identifier/path in warning output.
- [ ] Exclude secret value content from warnings.
- [ ] Include remediation recommendation in warning text.

---

## Phase 7: Tests

### Backend Tests

- [ ] Test precedence: env expansion, file precedence, plaintext fallback.
- [ ] Test validation for enabled features with no key source.
- [ ] Test save/load behavior does not leak resolved secrets into persisted config.
- [ ] Test settings API redaction behavior.
- [ ] Test logging/privacy behavior to confirm scrubbing.

### Frontend Tests

- [ ] Test API key placeholder and non-destructive save behavior.
- [ ] Test env/file guidance visibility and messaging.

---

## Phase 8: Documentation

- [ ] Document env-var setup examples (Docker/Dokploy).
- [ ] Document mounted secret-file examples (Docker/K8s).
- [ ] Document migration path from plaintext keys to `${ENV_VAR}` / `...File`.
- [ ] Add troubleshooting notes for unreadable secret files and missing env vars.

---

## Migration and Compatibility

- [ ] Keep plaintext secrets functional for v1 to avoid breaking existing users.
- [ ] Emit warnings for plaintext usage at runtime/settings save.
- [ ] Provide clear migration guidance in docs and logs.
- [ ] Define strict mode acceptance criteria for future release.

---

## Risks and Mitigations

- [ ] Missing env var handling produces clear startup/runtime errors.
- [ ] Unreadable secret file handling produces explicit permission/path errors.
- [ ] Settings save-flow regressions are covered by regression tests.

---

## Definition of Done (v1)

- [ ] Recommended secret flows do not require plaintext API keys in `config.yaml`.
- [ ] Env/file references work across AI, eBird, and weather integrations.
- [ ] Secret values do not appear in API responses, logs, or support outputs.
- [ ] Existing plaintext-based users remain functional with warning-driven migration path.
- [ ] Docs and tests are updated and passing for hardened behavior.
