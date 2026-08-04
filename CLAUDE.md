# RapidDeploy CLI (`rd`)

Go CLI (cobra) to manage a RapidDeploy server over its REST API. Module `github.com/MidVision/rd`, artifact `com.midvision.plugins:rdcli`. Public repo: https://github.com/MidVision/rd

## Architecture

- [main.go](main.go) just calls `cmd.Execute()`. Everything lives in `cmd/`.
- [cmd/root.go](cmd/root.go) defines `RootCmd` plus global `--debug`/`--quiet` flags. Each command file (`login.go`, `deploy.go`, `listServers.go`, ...) self-registers via `init() { RootCmd.AddCommand(...) }` — to add a command, copy an existing file's pattern.
- [cmd/rdclient.go](cmd/rdclient.go): `RDClient` REST client. `login` obtains a token from `<url>/ws/user/create/token` and persists `{url, token, username, password}` as JSON to `~/.rapiddeploy` (0600); every other command starts with `rdClient.loadLoginFile()`. All requests go through `rdClient.call()` → `util.go:call()` (5s timeout, exits the process on connection/HTTP errors).
- Responses are XML, unmarshalled into per-command structs (defined at the top of each command file); output is rendered with `tablewriter`. Some endpoints return HTML — the generic `Html/Div/Ul/Li` structs in rdclient.go parse those.
- [cmd/login.go](cmd/login.go): if no password is given with the default user (`mvadmin`), it tries cloud-default passwords in order: AWS instance-id (via IMDSv2), Azure `/etc/machine-id`, then `mvadmin`.
- No tests exist in this repo.

## Versioning — do not hardcode

`--version` reports `cmd.Version` ([cmd/root.go](cmd/root.go)), which defaults to `"development"` and is stamped at build time by Maven via
`-ldflags "-X github.com/MidVision/rd/cmd.Version=${project.version}"` (all five exec executions in [pom.xml](pom.xml)). It was once hardcoded (`1.7`) and drifted for years — never set a literal version in Go code; the Maven project version is the single source of truth.

## Build system

Maven orchestrates the Go build ([pom.xml](pom.xml)):

- `download-maven-plugin` fetches the **pinned Go toolchain** (`<go.version>` property) from go.dev, `maven-antrun-plugin` unpacks it into `./go/` (the Maven build directory, gitignored), then `exec-maven-plugin` cross-compiles five targets: mac (darwin/arm64), linux64, linux32, win64, win32 → `go/bin/<platform>/`.
- `mvn compile` builds all binaries; `mvn install` also installs them with per-platform classifiers (`-mac.bin`, `-linux64.bin`, `-win64.exe`, ...).
- **To upgrade Go**: change `<go.version>` in pom.xml (one line). Check both archives exist first: `go1.X.Y.darwin-arm64.tar.gz` and `go1.X.Y.linux-amd64.tar.gz` on go.dev (Jenkins builds on linux-amd64, local dev on mac). The `go 1.x` directive in [go.mod](go.mod) is the *minimum language version*, independent of the toolchain — it does not need to change with toolchain bumps.
- Local iteration without Maven: `go build -o rd main.go` works with any system Go ≥ the go.mod directive (binary reports `rd version development`).

## Releases

Releases are cut with the maven-release-plugin on internal CI: releasing `1.N` pushes two `[maven-release-plugin]` commits + tag `rdcli-1.N` to master and sets the development version to `1.(N+1)-SNAPSHOT`. Operational details live in `CLAUDE.local.md` (not committed).
