# BirdNET-Go AI Analyzer Fork

This repository is a fork of `tphakala/birdnet-go` with additional AI Analyzer functionality and related settings/UI.

## Goals

- Keep compatibility with upstream BirdNET-Go updates.
- Add AI analysis features without breaking base installation flow.
- Provide a simple custom installer entry point for this fork.

## Install This Fork

Run:

```bash
curl -fsSL https://raw.githubusercontent.com/bcardi0427/birdnet-go-aianalyzer/aianalyzer/main/install-aianalyzer.sh -o install-aianalyzer.sh
bash ./install-aianalyzer.sh
```

Review the downloaded script before running it if you want to inspect exactly what will execute.

## Proxmox Helper LXC Upgrade

If BirdNET-Go was installed with the Proxmox VE helper script, it usually runs as a native systemd service named `birdnet` with:

```text
WorkingDirectory=/opt/birdnet/data
ExecStart=/usr/local/bin/birdnet-go realtime
```

For that layout, use the LXC upgrade script from inside the LXC shell:

```bash
curl -fsSL https://raw.githubusercontent.com/bcardi0427/birdnet-go-aianalyzer/aianalyzer/main/scripts/install-aianalyzer-lxc.sh -o install-aianalyzer-lxc.sh
bash ./install-aianalyzer-lxc.sh
```

This script preserves `/opt/birdnet/data`, backs up `/usr/local/bin/birdnet-go`, and replaces only the binary used by `birdnet.service`.

Current requirement: this needs a GitHub release asset named `birdnet-go-linux-amd64`. Until that release binary exists, the script will stop with a clear message instead of changing the running service.

## Update Strategy

This fork tracks upstream using:

- `upstream`: `https://github.com/tphakala/birdnet-go.git`
- `origin`: `https://github.com/bcardi0427/birdnet-go-aianalyzer.git`
- custom branch: `aianalyzer/main`

Typical sync flow:

```bash
git fetch upstream
git checkout main
git merge upstream/main
git push origin main
git checkout aianalyzer/main
git merge main
git push origin aianalyzer/main
```

## Notes

- Upstream documentation remains valid for core setup and operations.
- Fork-specific behavior should be documented under `docs/aianalyzer/`.
- Contributions to this fork follow the upstream `CC BY-NC-SA 4.0` license and privacy-by-design expectations.
