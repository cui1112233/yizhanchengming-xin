# V88 public runtime topology

The active application release path is the direct Go binary flow in
`deploy/v88-direct`. Git `v88` is the maintained source authority; the server
is a runtime copy. Docker/Compose and GitHub Actions are not application
deployment entry points.

`runtime-topology.manifest.json` remains a historical, non-secret topology
record. It must not be used to start or replace the application.

## Evidence and ownership

The manifest records a source-reference snapshot, not the current production
release. Current releases are identified by the exact Git SHA embedded in the
Go binary and exposed by `/api/runtime-build-info`.

The direct release package contains the Linux Go binary, systemd unit,
activation script, and runtime identity manifest. The activation script switches
the `current` symlink, restarts `qiantie.service`, checks health and exact SHA,
and restores `previous` on failure.

The only permitted secret material is deploy-time environment values and certificate/session storage outside Git. Secret overlays are supplied at deployment and kept outside Git. Passwords, tokens, private keys, populated environment files, host-local mount paths, certificates, login/session content, database contents and business media must not be copied into the manifest or tracked deployment files. Persistent mounts are represented by purpose only; backup and restore procedures must retain their real storage separately.

## Service roles

| Runtime role | Active owner | Dependencies |
| --- | --- | --- |
| Go application | `/opt/qiantie/v88/direct/current/go/qiantie` via `qiantie.service` | MySQL, Redis, TOS, FFmpeg |
| Nginx | Host reverse proxy to `127.0.0.1:3000` | Go application |
| MySQL / Redis / TOS | Existing infrastructure | Go application |

Nginx owns the public ports and proxies to the host Go listener. The Go binary
itself owns the application port and embeds the React frontend.

## Validation and retirement

Run the direct deployment contract tests and `go test ./...` before release.

Retirement of old runtime artifacts still requires authenticated route evidence,
durable data/media readback, a restorable backup, and a known rollback point.
