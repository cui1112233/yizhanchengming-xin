# Doubao Relay Standalone Window Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build and install a Relay executable whose window always loads bundled assets rather than a localhost development server.

**Architecture:** Tauri remains the desktop shell. Its frontend is built into `frontend/dist`, then `cargo tauri build` produces the only installable executable. A guarded installer accepts only that output and leaves the running installation unchanged if packaging or validation fails.

**Tech Stack:** Tauri 2, Rust, Preact/Vite, pnpm, PowerShell.

## Global Constraints

- The installed application must work while `localhost:1477` is unavailable.
- Never overwrite the installed executable with `cargo build` output.
- Preserve the old executable until the replacement passes package and runtime checks.
- Do not launch a real Doubao generation while verifying this fix.

---

### Task 1: Make Frontend Packaging Reproducible

**Files:**
- Modify: `executors/doubao-relay/pnpm-workspace.yaml`
- Modify: `executors/doubao-relay/package.json`
- Test: `executors/doubao-relay/frontend/dist/index.html`

**Interfaces:**
- Consumes: `frontend/package.json` build command.
- Produces: populated `frontend/dist/index.html` and the Vite-hashed JavaScript entry it references.

- [ ] **Step 1: Write a failing packaging probe**

```powershell
pnpm --dir frontend build
Get-Content frontend/dist/index.html -Raw
```

Expected: the current command fails because `esbuild` is blocked.

- [ ] **Step 2: Configure pnpm to approve only esbuild**

```yaml
allowBuilds:
  esbuild: true
```

Place this setting where the pnpm version invoked by the build command reads project settings.

- [ ] **Step 3: Reinstall and build the frontend**

```powershell
pnpm install
pnpm --dir frontend build
Get-Content frontend/dist/index.html -Raw
```

Expected: the command exits zero and `index.html` references a JavaScript file present in `frontend/dist/assets`.

### Task 2: Guard the Tauri Release Artifact

**Files:**
- Create: `executors/doubao-relay/scripts/build-standalone-win.ps1`
- Test: `executors/doubao-relay/scripts/build-standalone-win.ps1`

**Interfaces:**
- Consumes: `frontend/dist/index.html`, `src-tauri/tauri.conf.json`.
- Produces: `src-tauri/target/release/bundle/nsis/*.exe` or the configured Tauri bundle executable.

- [ ] **Step 1: Write a failing source-artifact check**

```powershell
if ($candidate -like '*target\\release\\doubao-relay.exe') {
  throw '普通 cargo build 产物不能安装'
}
```

- [ ] **Step 2: Implement guarded build script**

```powershell
& cargo tauri build
if (-not (Test-Path $frontendEntry)) { throw '内置界面入口缺失' }
if (-not (Test-Path $bundleExe)) { throw 'Tauri 安装包缺失' }
Copy-Item $bundleExe $outputExe
```

- [ ] **Step 3: Run guarded build**

```powershell
powershell -ExecutionPolicy Bypass -File scripts/build-standalone-win.ps1
```

Expected: output references the Tauri bundle, never `target/release/doubao-relay.exe`.

### Task 3: Install and Verify the Window

**Files:**
- Create: `executors/doubao-relay/scripts/install-standalone-win.ps1`
- Test: installed `C:\Users\Administrator\AppData\Local\一战晟铭执行器\doubao-relay.exe`

**Interfaces:**
- Consumes: output from `build-standalone-win.ps1`.
- Produces: installed, backed-up executable and a running Relay health endpoint.

- [ ] **Step 1: Implement atomic install guard**

```powershell
Copy-Item $installed "$installed.backup" -Force
Copy-Item $built "$installed.new" -Force
Move-Item $installed.new $installed -Force
```

The script must run only after the guarded build succeeds.

- [ ] **Step 2: Restart and probe health**

```powershell
Invoke-RestMethod http://127.0.0.1:8787/health
```

Expected: `{ "status": "ok" }`.

- [ ] **Step 3: Visually verify the execution window**

Open the installed executable while no process listens on port `1477`. The window must display the Relay application UI, not an Edge error page containing `localhost 拒绝连接`.

- [ ] **Step 4: Commit**

```bash
git add executors/doubao-relay docs/superpowers
git commit -m "fix: package relay with bundled frontend"
```
