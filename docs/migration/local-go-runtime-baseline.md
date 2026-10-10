# Local Go runtime baseline

Date: 2026-10-10
Branch: `v88`
Baseline source SHA before this runtime-skeleton work: `2263245f2f361dbdc1af6aaea3d818ea042fa712`

## Verified

- `go test ./...` passes from `backend/`.
- Go health behavior is covered: MySQL probe failure returns HTTP 503; healthy probe returns HTTP 200.
- Redis runtime status distinguishes unconfigured, invalid, unreachable, and reachable states without exposing credentials.
- Static handler tests cover `/`, hashed assets, missing assets, and missing frontend files.
- `scripts/build-go-local.ps1 -SkipGoBuild` copies `frontend/dist` into `backend/web/dist`.
- `go build -o qiantie-local.exe ./cmd/qiantie` succeeds.

## Not yet verified on this machine

- The Go service was not started because the current shell did not provide the required MySQL DSN and bridge secret.
- `scripts/check-local-go.ps1` correctly failed against the unused default port `4000` with connection errors.
- Real MySQL/Redis connectivity, authenticated browser flow, external providers, TOS, and ECS/public deployment remain outside this local skeleton check.
- Existing Node/Vite processes and unrelated uncommitted changes were left untouched.

## Runtime boundary

Go is the intended single local runtime entry point. React remains a build input; Node/npm remains a development/build tool. MySQL remains the durable business-data source, while Redis is only a runtime queue/lock dependency.
