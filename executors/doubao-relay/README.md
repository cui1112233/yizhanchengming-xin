# doubao-relay

![version](https://img.shields.io/badge/version-1.3.0-blue) ![license](https://img.shields.io/badge/license-MIT-green)

<img src="assets/logo.png" width="96" align="right" alt="logo">

把本机**已登录的豆包桌面端**的生图 / 生视频能力,通过本地 HTTP API(含 OpenAI 兼容接口)对外提供的桌面客户端。
基于 Tauri 2(Rust + Preact),通过 CDP 操作豆包界面,结果返回豆包 CDN 的无水印原图 / 原片地址。

变更记录见 [CHANGELOG.md](./CHANGELOG.md)。

## 安装

| 平台 | 文件 | 说明 |
|---|---|---|
| macOS 12+(Intel / Apple 芯片) | `doubao-relay_<版本>_universal.dmg` | 打开后把 `doubao-relay.app` 拖进"应用程序" |
| Windows 10 / 11 x64 | `doubao-relay_<版本>_windows_x64.zip` | 解压后直接运行 `doubao-relay.exe`(需系统自带的 WebView2) |

未签名的包首次打开会被系统拦下:macOS 在 Finder 里右键 → 打开;Windows SmartScreen 选"更多信息 → 仍要运行"。

## 使用

1. 先安装并登录**豆包桌面端**(可登录多个账号)。
2. 打开 doubao-relay,它会自动用调试端口(`127.0.0.1:9707`)连接豆包;豆包在运行但没开调试端口时会自动重启它
   (登录状态不受影响)。关闭 doubao-relay 不会关闭豆包。
3. 左下角状态栏显示豆包连接、API 监听状态,以及 **relay 版本和豆包版本**(报问题时请一起截图)。
4. 设置页可复制 Bearer token,并列出所有接口的调用示例。

数据目录:macOS `~/Library/Application Support/doubao-relay/`(日志在 `~/Library/Logs/doubao-relay/`),
Windows `%APPDATA%\doubao-relay\`。卸载不会删除这些目录。

**注意**:任务运行时请不要在豆包里打字、切换账号或新建对话;单个豆包实例下任务按顺序执行。

## API(OpenAI 兼容)

地址 `http://127.0.0.1:8787`,鉴权 `Authorization: Bearer <token>` 或 `?key=<token>`(`/health` 免鉴权)。

| 接口 | 说明 |
|---|---|
| `GET /health` | 存活检查,返回 `version` |
| `GET /v1/accounts` | 已登录账号与当前账号(`?refresh=1` 强制重读) |
| `POST /v1/accounts/switch` | `{"account":"张三"}` 切换账号(有任务在跑时 409) |
| `GET /v1/models` | 可用模型 |
| `POST /v1/images/generations` | **同步**生图,出图后返回 |
| `POST /v1/video/generations` | **异步**生视频,立即返回 `id` |
| `GET /v1/video/generations/{id}` | 查询视频任务 |

模型:`doubao-image-4.0` / `doubao-image-4.5`(默认)/ `doubao-image-5.0-flash`,
`doubao-video-mini` / `doubao-video-fast`(默认)。

请求字段:`prompt`(必填)、`model`、`ratio`(或 OpenAI 风格 `size`)、`account`(不传用豆包当前账号)、
`ref_image`(参考图,图片网址或 `data:image/png;base64,...`,多张用数组)、视频另有 `duration`(秒,默认 5)。
生成请求里带 `account` 会自动切换账号,一般不需要单独调切换接口。

```bash
TOKEN="<设置页复制的 token>"

# 生图(同步)
curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"model":"doubao-image-4.5","prompt":"一只蓝色小鸟站在樱花枝上","ratio":"1:1"}' \
  http://127.0.0.1:8787/v1/images/generations
# → {"created":…,"task_id":"…","data":[{"url":"<豆包 CDN 原图>","local_url":"…","width":2048,"height":2048,…}]}

# 生视频(异步)+ 参考图
curl -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"model":"doubao-video-fast","prompt":"人物在海边散步","ref_image":"https://example.com/a.png","ratio":"16:9","duration":5}' \
  http://127.0.0.1:8787/v1/video/generations
# → {"id":"…","status":"pending",…}
curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:8787/v1/video/generations/<id>
# 完成后 → {"status":"completed","data":[{"video_url":"<豆包 CDN 原片>","local_url":"…","duration":5.0,…}]}
```

- `url` / `video_url` 是豆包 CDN 地址,带签名会过期;`local_url` 是 relay 本地备份(同一个 token 鉴权)。
- 错误统一为 `{"error":{"code","message","details"}}`;生成失败(如豆包拒绝)返回 502,`details.task_id` 可用于排查。
- 豆包弹出的"安全确认"和"请先确认参数"会自动确认。

原始任务接口(`/api/v1/generate`、`/api/v1/tasks/{id}` 等)仍可用,规格见 `docs/spec/04-http-api.md`。

## 开发

前置:Rust 1.94、Node 20+、pnpm 9、`cargo install tauri-cli --version "^2.12" --locked`。

```bash
(cd frontend && pnpm install)
cargo tauri dev --no-watch      # 开发运行(前端 :1477);--no-watch 避免改代码时自动重启打断任务

cargo fmt --all --check && cargo clippy --workspace --all-targets -- -D warnings && cargo test --workspace
(cd frontend && pnpm typecheck && pnpm test -- --run)
cargo test live_ -- --ignored --nocapture   # 需要真豆包的测试(干跑,不消耗配额)
```

给 AI 编程助手的说明见 [AGENTS.md](./AGENTS.md)。设计文档:`docs/spec/`(做什么)、`docs/analysis/`(怎么做与真机探测记录)。

## 打包发布

```bash
bash scripts/bump-version.sh 1.0.1   # 同步版本号,并在 CHANGELOG.md 加 ## [1.0.1] 段落
bash scripts/package.sh              # 产物在 dist/:universal DMG + Windows x64 zip
```

Windows 包在 macOS 上用 `cargo-xwin` 交叉编译(需 `rustup target add x86_64-pc-windows-msvc`、
`cargo install cargo-xwin`、`brew install llvm`)。也可打 `v*` tag 走 `.github/workflows/release.yml` 在 CI 上构建。

## 版权与免责

- 本项目不包含豆包的代码或资源;`docs/analysis/reference/` 与注入脚本来自对沈浪 Ai 画布的逆向研究,仓库须保持私有。
- 通过模拟客户端操作使用豆包**可能违反豆包用户协议**,账号被封等后果由使用者承担。
- 只在用户已登录的豆包客户端上做前端操作,不绕过鉴权、不读取账号凭证、不伪造接口签名。
