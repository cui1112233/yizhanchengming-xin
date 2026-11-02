# Changelog

格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/),版本号遵守 [SemVer](https://semver.org/lang/zh-CN/)
(`0.x` 为 Alpha,HTTP API 可能有破坏性变更,见 `docs/spec/13-build-deploy.md`「版本号」)。

发版时把 `[Unreleased]` 下的条目挪到新版本段落,`scripts/bump-version.sh <ver>` 同步版本号;
`release.yml` 会校验 tag、三处版本号与本文件的 `## [<ver>]` 段落一致,并把该段落作为 GitHub Release 说明。

## [Unreleased]

## [1.3.0] - 2026-10-28

### 修复
- V88 领取任务后只读取豆包页面当前显示的账号，并在本地入队、生成、结果回传全程沿用已确认账号，不再因账号缓存过期打开菜单，避免干扰紧接着的模型和比例设置。
- 优先使用新版豆包页面直接显示的模型、比例按钮，旧版总设置菜单只作为兼容回退，减少页面改版导致的连续生成失败。

## [1.1.0] - 2026-10-19

### 新增
- V88 配对、心跳、任务领取、租约续期和结果回传。
- 从 V88 网页执行图片与视频生成任务。

### 安全
- V88 执行器凭证与本地 HTTP API Bearer Token 完全分离。

## [1.0.0] - 2026-10-19

### 新增
- OpenAI 兼容接口:`/v1/models`、`/v1/accounts`、`/v1/accounts/switch`、同步 `/v1/images/generations`、
  异步 `/v1/video/generations` + `GET /v1/video/generations/{id}`;参考图用 `ref_image`(data URI 或图片网址);
  结果返回豆包 CDN 原始地址(`url` / `video_url`)与本地 `local_url`;鉴权支持 Bearer 与 `?key=`。
- 模型 ID:`doubao-image-4.0/4.5/5.0-flash`、`doubao-video-mini/fast`;新增 Seedream 4.0。
- 设置页列出全部接口调用示例;侧栏左下角显示 relay 版本与豆包版本;应用 logo。

### 修复(适配豆包 2.32.6)
- 参考图上传入口改为"+"→"添加图片 / 上传素材"后才生成,自动打开。
- 首次上传素材的"安全确认"弹窗自动确认。
- 视频技能"请先确认参数"话术自动回复"继续生成"。
- 绑定前记录的其它对话历史模板在绑定后丢弃。
- 状态栏 API 状态显示真实监听状态;开发端口改为 1477。

## [0.1.0] - 未发布

首个版本:通过 CDP 操控本机豆包桌面端(2.32.5 macOS 实测),对外提供本地生图 / 生视频 HTTP API。

### 新增

- **豆包连接**:自动发现 / 启动带 `--remote-debugging-port=9707` 的豆包,target 白名单 + 监听进程路径校验;
  10s 心跳、断线自动重连、豆包退出检测;UI 兼容性自检(`ok` / `warn` / `error` / `unknown`)。
- **账号**:读取豆包账号菜单(单行 / 两行账号文本、无头像账号行),按名字切换账号并验证;
  任务运行中拒绝切换(`account_busy`)。
- **生图管线**:按模型 / 比例准备豆包"图像生成"技能,填 prompt、发送并绑定提交身份;
  捕获原图、多轮确认(`auto` 代回复 / `manual` 等待用户确认)、拒绝 / 配额 / 中断识别;结果按内容判定格式下载落盘。
- **生视频(部分)**:"视频生成"技能的准备与提交、视频候选直链捕获;本地解密视频 token(AES-128-CBC,与豆包原脚本逐向量一致)
  与无水印直链过滤。时长滑块设置、参考图上传、视频结果落盘尚未完成(Phase 3)。
- **任务调度**:严格 FIFO 队列(`global_max=3`、每账号 1 个、队列上限 50),各阶段超时、取消、重启恢复。
- **HTTP API**:`http://127.0.0.1:8787/api/v1`,Bearer 鉴权(常量时间比较),`/health` 免鉴权;
  账号、上传、生成、任务查询 / 取消 / 确认、结果下载(Range)、豆包状态 / 重连;统一错误体 `{error:{code,message,details}}`。
- **持久化**:SQLite(任务、阶段历史、上传、webhook 日志),结果文件保留 24h、元数据 7d、上传 1h,每小时 GC。
- **桌面端**:Tauri 2 + Preact,账号页(实时状态栏、切换确认框)。
- **发布工程**:GitHub Actions CI(macOS + Windows)、tag 触发的发布流水线、macOS 签名 / 公证脚本、
  Windows 签名脚本、版本号同步脚本。

### 已知限制

- 只在 macOS + 豆包 2.32.5 上做过真机验证;Windows 代码只做过交叉编译检查。
- 只有 1 个有效账号的环境下,"切换账号成功 → 页面 reload → 重连"路径只有 mock 覆盖。
- 详见 `docs/superpowers/plans/2026-10-16-doubao-relay-full.md`「已知问题」与 `docs/analysis/10-live-probe-macos.md`。
