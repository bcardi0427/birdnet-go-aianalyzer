# Project Persistence Guardrails
You have a custom skill called `update-state-log`. 

- **Before executing any changes:** You must open and read `.agent/state_log.json` if it exists. Align your logic strictly with the `current_objective` and avoid patterns documented in `active_problems`.
- **When a fix is achieved:** Execute `update-state-log` passing your fix under `--milestone`, clear old blockers with `--clear_problems`, and set the exact target code segment under `--next`.
- **When an operation fails or you get stuck:** Do not try to write a new file from scratch. Update the log using `--problem` explaining what broke, so the context is preserved across crashes or workspace resets.