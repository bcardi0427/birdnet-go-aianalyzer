# Project Persistence Guardrails
You have a custom skill called `update-state-log` which keeps the local `.agents/state_log.json` updated and automatically syncs it to a central Cloud Firestore database.

- **Before executing any changes:** You must open and read `.agents/state_log.json` if it exists. Align your logic strictly with the `current_objective` and avoid patterns documented in `active_problems`.
- **When starting a new objective:** Execute `update-state-log` with `--objective` and `--why` to log it both locally and to Cloud Firestore.
- **When a fix is achieved:** Execute `update-state-log` passing your fix under `--milestone`, clear old blockers with `--clear_problems`, and set the next execution target under `--next`.
- **When an operation fails or you get stuck:** Do not try to write a new file from scratch. Update the log using `--problem` explaining what broke, so the context is preserved across crashes, workspace resets, or handoffs to other agents.