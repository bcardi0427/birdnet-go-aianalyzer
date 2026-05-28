# Name
update-state-log

# Description
Use this skill to log milestones, document ongoing environment errors, or checkpoint project progression. It ensures future workspace agents understand historical solutions and don't re-invent previous bug fixes.

# Goal
Maintains a sliding-window memory database inside `.agent/state_log.json` detailing the architectural scope, error states, and history.

# Instructions
Extract parameters from the environment context or explicit user command and pass them directly to the underlying CLI script:
- If a new goal is set, pass `--objective` and `--why`.
- If an operation or compilation script fails, pass the exact error log via `--problem`.
- If a problem is fixed, pass `--clear_problems` and log the fix details via `--milestone`.
- Always extrapolate the immediate execution path into `--next`.

# CLI Invocation
python3 .agent/skills/update-state-log/index.py {{args}}

# Constraints
- Keep strings direct and functional.
- Do not make up successful milestones if a build/compile step is throwing errors.