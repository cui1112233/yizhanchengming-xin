# 音频混合导出链路验收记录

日期：2026-11-20

## 1. 原版对标

- REA JavaScript 应用分析父证据：`ev_0122c8437a7aadaa1c949cc2f5036d38dc280ccf2c13e3c103c492f74c92c2d1`。
- REA 摘要证据：`ev_91bf66c518a8f5af02e3c694d9f4d58525b801e8b4af702453513e2fccaafdc2`。
- 发布脚本直接核对：普通音频默认 `volume: .7`，口播默认 `volume: 1`，随机音乐音量范围 0–200%，普通音频轨和口播轨分别最多 10 条。
- `Audio-BeyA9XJk.js` 直接核对：选中音频后的右侧属性顺序为“音量 → 倍速 → 时长”；音量为 0–200%，倍速为 0.1–4，0% 就是静音。
- 主进程桥直接核对：`export-audio`、`export-audio-for-asr`、`export-audio-segments-for-asr`、`convert-audio-to-mp3`。
- `mcengine/index.cjs` 直接核对：根据 `MCE_LOADER` 装载 `mce_napi.node` 或 `mce_term.node`。

## 2. IDA 状态

已按 `ida-reverse` 官方 `start.ps1`、`open.ps1` 尝试打开 `mce_napi.node`。IDA MCP 监听端口为 13337，但 `tools/list` 持续超时。界面检查和 `supervisor.log` 证明根因是 `idalib` worker 以 255 退出，并显示：`IDADIR` 指向的安装不存在或不完整。当前 `F:\软件\IDA Professional 9.5` 虽然存在 `ida.exe`、`ida.dll`，但 idalib 自动化仍拒绝启动。

因此本步没有伪造 IDA 反编译结果。原生内部实现继续登记为阻塞项；已确认的业务合同只来自 REA、已解包发布脚本和受控本地 FFmpeg 实验。

## 3. 自动测试

- 前端：52/52 通过，其中新增音频属性顺序、范围、静音、倍速/时长联动、保存命令测试。
- Go：`go test ./...` 全部通过。
- `TestBuildRenderPlanMixesVideoOriginalSoundAndTimelineAudio`：检查原声、独立音频、70% 音量、500ms 延迟和 `amix` 参数。
- `TestFFmpegExportsPlayableOneSecondMP4`：真实保留视频原声，FFprobe 为 H.264/AAC，volumedetect 证明不是静音。
- `TestFFmpegMixesTimelineAudioWithOriginalSound`：真实视频原声 + 独立 880Hz 音频混合成功。
- Node 24 在 Windows 下完成 3005 模块转换后进程异常退出；项目固定脚本选择 Node 22.22.2 后构建成功。这是构建运行时差异，不是 React 编译错误。

## 4. 4040 实际操作

1. 通过 4040 已登录页面进入工程 `mcp_37280d23168cf54f1099a0ed48570677`。
2. 上传本地合成音频 `magic-cut-audio-880hz.m4a`，页面显示 `9.8 KB · 00:01.000 · 可用`。
3. 双击素材后，音频轨出现 `magic-cut-audio-880hz.m4a` 片段。
4. 导出名称填写“音频混合验收”，提交任务 `mct_f8f45491c40a76930512f078069e2985`。
5. 任务中心显示视频导出“已完成”、进度 100%，可下载成片。

## 5. 输出文件核对

- 导出记录：`mce_c6e6603725c217137f5d14d0b557a16d`。
- SHA-256：`fdaef9c0fe89d1d8a4854bfead09ec43d2e818468343637641b76bfa025841e4`。
- 时长：2.000 秒。
- 视频：H.264。
- 音频：AAC。
- 文件大小：25,209 字节。
- `mean_volume`: -20.1 dB。
- `max_volume`: -10.5 dB。

## 6. 音频属性界面实测

1. 在同一工程的时间线中选中音频片段，右侧原有空属性区切换成“音频属性”，没有跳到另一个页面。
2. 页面按原版顺序显示“音量、倍速、时长”，并明确显示“调到 0% 即静音”。
3. 把音量从 70% 改为 35%，倍速从 1 改为 2，时长自动从 1.00 秒联动为 0.50 秒。
4. 页面显示“已保存”；刷新后重新选择音频片段，仍读回 `35 / 2.0 / 0.50`。
5. MySQL 最新时间线快照为 revision 4，实际保存 `volume: 0.35`、`speed: 2`、`durationMs: 500`。

结论：这一条真实运行链已从“导出静音 AAC”提升为“保留视频原声并混合独立音轨”，同时补齐了编辑器右侧的基础音频属性入口。口播、TTS、波形、淡入淡出、降噪和变声等仍未完成，不能把 MC-017 或 MC-022 标成全部完成。

