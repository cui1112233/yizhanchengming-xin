# 魔剪音频混合导出链路规格

更新日期：2026-11-20

## 本步范围

本步只闭环“导入音频素材 → 放入音频轨 → 与视频原声混合 → 导出 AAC 音轨”这一条真实链路，不把尚未验证的 TTS、录音、变声和高级音效冒充完成。

## 原版静态证据

| 环节 | 原版发布包观察 | 本步合同 |
| --- | --- | --- |
| 提取音频 | `ExportAudioPanel-Dfg5SDYr.js` 使用 `fileType: "export_audio"`，把素材标记为 `material_type: "audio"` | 音频素材必须按音频类型进入素材库和音频轨 |
| 普通音频轨 | 新增普通音频时写入 `mediaElements.audio_list`，默认 `volume: .7` | 新增普通音频片段默认音量为 70% |
| 口播轨 | 新增口播时写入 `mediaElements.microphone_list`，默认 `volume: 1` | 口播默认音量为 100%；本步先保留数据合同，口播 UI 后续单独闭环 |
| 数量限制 | 普通音频轨和口播轨分别提示上限 10 条 | 后续轨道管理必须分别限制 10 条，本步不把单条基础轨伪装成完整轨道管理 |
| 随机音乐 | 音量控件范围 0–200，并换算为 `volume / 100` | 片段音量数据允许 0–2；0 代表静音，不可被当成“未填写” |
| 本地处理 | `main.context.v2.js` 暴露 `export-audio`、`export-audio-for-asr`、`export-audio-segments-for-asr`、`convert-audio-to-mp3` | Web 重写由 Go + FFmpeg 在本机处理，不调用原版服务或付费接口 |
| 原生边界 | `mcengine/index.cjs` 根据 `MCE_LOADER` 装载 `mce_napi.node` 或 `mce_term.node` | IDA 只用于核对原生边界；本步不复制原版二进制代码 |

REA 本批静态图证据：`ev_0122c8437a7aadaa1c949cc2f5036d38dc280ccf2c13e3c103c492f74c92c2d1`、`ev_91bf66c518a8f5af02e3c694d9f4d58525b801e8b4af702453513e2fccaafdc2`。图覆盖为部分覆盖，具体参数同时由已解包的发布脚本直接核对。

## 4040 当前差距

`api/internal/export/service.go` 目前只收集视频轨，并无条件生成 `anullsrc` 静音 AAC。因此已有界面虽能放入音频素材，导出文件仍会丢掉视频原声和独立音轨。

## 本步实现合同

1. 时间线片段增加可选 `volume`，范围为 0–2；不传时视频原声默认 1，普通音频默认 0.7。
2. 视频片段有音频流时保留原声；没有音频流时只为该视频片段补等长静音，保证片段拼接时声音轴不移位。
3. 独立音频片段按 `sourceStartMs` 裁切，按 `speed` 变速，按 `startMs` 延迟到时间线位置，再按 `volume` 调整音量。
4. 视频原声和所有独立音频片段通过 FFmpeg `amix` 混合，最终仍输出 H.264 + AAC 的 MP4。
5. 输出必须通过真实 FFmpeg/FFprobe 验证，并用音量检测证明不是静音文件。

## 本步不宣称完成

- 口播录制、TTS、波形、淡入淡出、声道映射、降噪、变声。
- 轨道增删、锁定、静音按钮和每类 10 轨的完整 UI。
- 原版 mcengine 内部算法的一比一复制；只对标可观察输入输出合同。

