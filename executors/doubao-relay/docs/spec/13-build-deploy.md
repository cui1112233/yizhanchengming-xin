# 13 · 构建与发布

## 项目结构

```
doubao-relay/
├── docs/
│   └── spec/                     本 spec
├── src-tauri/
│   ├── Cargo.toml
│   ├── tauri.conf.json
│   ├── build.rs
│   ├── icons/
│   │   ├── icon.ico              Windows
│   │   ├── icon.icns             macOS
│   │   └── icon.png              PNG for Linux (保留)
│   └── src/
│       ├── main.rs
│       ├── lib.rs
│       ├── cdp/
│       ├── doubao/
│       │   └── scripts/
│       │       ├── doubao-generation-page.js
│       │       ├── doubao-result-reader.js
│       │       └── doubao-video-decoder.js
│       ├── platform/
│       ├── storage/
│       ├── http_server/
│       ├── config.rs
│       └── commands.rs
├── frontend/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── index.html
│   └── src/
│       ├── main.ts
│       ├── router.ts
│       └── pages/
├── migrations/                   SQLx migrations
│   ├── 20261016_initial.sql
│   └── ...
├── scripts/
│   ├── sign-windows.ps1
│   ├── sign-macos.sh
│   └── notarize-macos.sh
├── .github/
│   └── workflows/
│       ├── ci.yml                lint + test
│       └── release.yml           build + sign + publish
├── Cargo.toml                    workspace
├── package.json                  用于 frontend 构建编排
├── README.md
├── LICENSE
└── CHANGELOG.md
```

## 开发环境

### 前置

- **Rust**: stable 1.80+(用 rustup 装)
- **Node.js**: 20+ LTS
- **pnpm**: 9+(替代 npm)
- **系统工具**:
  - Windows: Visual Studio Build Tools + WebView2 Runtime
  - macOS: Xcode Command Line Tools

### 首次装

```bash
cargo install tauri-cli --version "^2.0"
cd frontend && pnpm install
cd ..
sqlx database create && sqlx migrate run
```

### 本地跑

```bash
cargo tauri dev
```

- 自动启 vite dev server
- 自动启 Rust 后端 + Tauri window
- 代码改动热重载(前端),Rust 改动自动 rebuild

### 单元测试

```bash
cargo test --workspace
cd frontend && pnpm test
```

### 集成测试

v0.1 不做自动化集成测试(难做,豆包不可控)。改用:

- 手动测试清单(见 14-acceptance.md)
- 关键模块 mock 测试:mock CDP server + mock SQLite

## 构建

### 开发构建

```bash
cargo tauri build --debug
```

输出:
- Windows: `src-tauri/target/debug/bundle/msi/doubao-relay_0.1.0_x64_en-US.msi`
- macOS: `src-tauri/target/debug/bundle/dmg/doubao-relay_0.1.0_x64.dmg`

### 发布构建

```bash
# Windows (在 Windows 机器上跑)
cargo tauri build --target x86_64-pc-windows-msvc

# macOS (在 Mac 机器上跑;universal binary)
cargo tauri build --target universal-apple-darwin
```

**不交叉编译**(交叉编译 Tauri 很折腾,直接用对应平台的 CI runner 更简单)。

### 产物

| 平台 | 文件 | 大小估计 |
|---|---|---|
| Windows | `doubao-relay_<ver>_x64_en-US.msi` | 10-20 MB |
| Windows | `doubao-relay_<ver>_x64-setup.exe` | 10-20 MB |
| macOS | `doubao-relay_<ver>_universal.dmg` | 15-30 MB |

## 代码签名

### Windows

```powershell
# 用 EV Code Signing Certificate
$cert = Get-ChildItem Cert:\CurrentUser\My\<thumbprint>
signtool sign `
    /sha1 $cert.Thumbprint `
    /tr http://timestamp.digicert.com `
    /td sha256 /fd sha256 `
    doubao-relay_0.1.0_x64_en-US.msi
```

### macOS

```bash
# 签名
codesign --sign "Developer ID Application: <Your Name> (TEAMID)" \
         --deep --timestamp --options runtime \
         doubao-relay.app

# DMG 打包
hdiutil create -volname "doubao-relay" -srcfolder doubao-relay.app \
               -ov -format UDZO doubao-relay_0.1.0_universal.dmg

# 公证
xcrun notarytool submit doubao-relay_0.1.0_universal.dmg \
      --apple-id <apple-id> --team-id <team-id> --password <app-password> \
      --wait

# 装订公证票据
xcrun stapler staple doubao-relay_0.1.0_universal.dmg
```

**首次发布建议**:至少先搞定 macOS 公证,Windows 可以等到有用户抱怨 SmartScreen 时再签。

## CI/CD

### `ci.yml`(每个 PR)

```yaml
jobs:
  lint-rust:
    - cargo fmt --check
    - cargo clippy -- -D warnings
  test-rust:
    - cargo test --workspace
  lint-frontend:
    - pnpm lint
    - pnpm typecheck
  test-frontend:
    - pnpm test
```

matrix 跑 ubuntu-latest + macos-latest + windows-latest。

### `release.yml`(打 tag 触发)

```yaml
on:
  push:
    tags: ['v*']

jobs:
  build-windows:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v4
      - setup-rust / node / pnpm
      - cargo tauri build --target x86_64-pc-windows-msvc
      - sign (optional)
      - upload artifact

  build-macos:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
      - setup-rust / node / pnpm
      - cargo tauri build --target universal-apple-darwin
      - sign + notarize (optional)
      - upload artifact

  publish-release:
    needs: [build-windows, build-macos]
    runs-on: ubuntu-latest
    steps:
      - download all artifacts
      - create GitHub Release
      - upload assets
```

## 版本号

**semver**: `MAJOR.MINOR.PATCH`

- `0.x.y` — Alpha,API 可能破坏性变更
- `1.0.0` — 稳定版,HTTP API 从此遵守 semver
- `2.0.0` — 破坏性变更(URL 加 `/v2/`)

Rust 和前端的版本号 MUST 同步:

- `Cargo.toml` 的 `[package] version`
- `tauri.conf.json` 的 `"version"`
- `frontend/package.json` 的 `"version"`

有一个 `scripts/bump-version.sh` 保持三处同步。

## 自动更新(opt-in)

**v0.1 不做内置自动更新**(Tauri 的 updater 需要签名 feed,复杂)。
**v0.2+** 可考虑:
- GET `https://api.github.com/repos/<org>/doubao-relay/releases/latest`
- 发现新版本在 UI 显示 banner
- 用户点"下载",跳到 GitHub release 页

## 卸载

### Windows (MSI)

通过"添加/删除程序"卸载。MSI 自动清理:
- 程序文件
- 开始菜单项
- 开机自启项(如有)

**保留**:
- `%APPDATA%\doubao-relay\`(用户数据)

### macOS (DMG)

拖 Applications 里的 Doubao-Relay.app 到废纸篓。

**保留**:
- `~/Library/Application Support/doubao-relay/`

### 清理用户数据(手动)

提供 `uninstall-user-data.sh/ps1` 脚本,一键清:

```bash
# macOS
rm -rf ~/Library/Application\ Support/doubao-relay/
rm -rf ~/Library/Logs/doubao-relay/
```

```powershell
# Windows
Remove-Item -Recurse $env:APPDATA\doubao-relay\
```

## 发布 checklist

每次打 tag 前:

```
□ 版本号三处同步 (Cargo.toml / tauri.conf.json / package.json)
□ CHANGELOG.md 更新
□ 跑 cargo clippy -- -D warnings 零警告
□ 跑 pnpm typecheck 零错误
□ 手动跑一遍 14-acceptance.md 的 P0 测试
□ Windows 和 macOS 两边都跑一次 tauri build 成功
□ 签名(如果有证书)
□ README 版本号 badge 更新
□ 打 git tag v0.x.y && git push --tags
□ 等 CI 自动发 release
□ 写 release note(复用 CHANGELOG 相应段)
□ (如有用户群)发公告
```

## 依赖升级

每月跑一次:

```bash
cargo update
cargo audit                       # 查已知漏洞
cd frontend && pnpm up --latest
```

破坏性升级(major 版本)单独 PR,带充分测试。
