# 音色绑定、单条与批量配音

2026-11-17，唯一基准仍为用户ZIP。本文的O指可读前端`readable-supplied-frontend/pretty/index-CnO3GE3B.js`。证据根见README；N是`full-extracted/洛水2026/modules/api.pyd`的编译方法声明行，不代表已恢复Python函数体。

## 输入、入口与顺序

| 编号 | 输入/操作 | 原链路与实际结果 | 边界和失败 |
|---|---|---|---|
| VT01 | 人物及选定音色 | O93915附近加载人物，O93928读取get_screenplay_voices；原SQLite.characters含bound_audio_path，voice_config保存character_id/name/id/settings | 仅在agent_assets.json写人物或仅保存voice_id不等于绑定参考音频。`tts-binding-v3/v5/v6.json`中id-only明确返回角色没有配置参考音频 |
| VT02 | 单行shot_id、line_index、character、text、emotion、strength、speed | O96545起原单条按钮调用tts_generate_single（N26875）；默认旁白/平静/中等。ScriptEditor此入口未发volume | 空白正文在页面提示请先输入配音文本，不请求；原生empty-text返回文本为空。错误显示error，finally复位_generating并保存voice_text |
| VT03 | 实际绑定参考WAV | 真实原生先POST本地/v1/upload_audio，再POST/v2/synthesize；捕获text/audio_path/speed/emo_vector，平静中等的本例向量末项0.5 | 本地默认地址127.0.0.1:8000是原版部署约定，CM必须用自己的已配置服务。v5在合成边界明确中止；v6响应为自造静音WAV，不是语音质量或供应商验证 |
| VT04 | 合成响应音频字节 | v6原生写作品audio/分镜1_OWN_ALICE_台词0_*.wav并返回success/audio_path；单次原生调用未自动修改该行voice_text | 不能收到success就认为音频路径已入库。原单条页面写入i.audio_path，再由kt调用update_screenplay_shot（O93982–93990） |
| VT05 | 路径保存与重开 | 原单条函数及kt受控执行6组；成功请求送入真实update_screenplay_shot，再在新Python进程打开原数据库，voice_text完全相等且WAV可读（1600帧/16000Hz/单声道） | 原kt忽略success=false，异常只console；原单条提示成功早于保存。CM必须区分已生成/未保存，保存失败允许重试保存，不能付费重生成来修保存 |
| VT06 | 批量模式all/incomplete及全部分镜voice_text | O95127–95182：解析每镜voice_text，有text才入队；incomplete跳过已有audio_path；保留原数组line_index；传tts_batch_start(project,{voice_entries,mode}) | malformed JSON被跳过；无符合项页面warning并不请求。原生空列表success,total0，idle progress running=false，idle stop返回无运行任务 |
| VT07 | 批量进度/停止 | 页面每1000ms读取tts_batch_progress，running=false后停止轮询、重载分镜，按failed显示成功/部分失败；停止入口fo见O95352–95362 | 进度请求抛错时仅console，Qa仍true；受控回放已复现。CM应保留任务ID、提示查询失败并可刷新/恢复，不把异常认成完成或永远禁止重试。真实在途取消及批量写库尚未验证 |
| VT08 | 音色设计文本 | O101024使用voice_design_generate(文字模型,127,正文)，O101026解析；N15702 parser受控输入characters，正常项会规范name/role/prompt并补足preview_text | broken string和{}也success=true/data[]。必须以实际可用条目判断，不将空数组显示为设计成功；生成音色/克隆服务未实测 |

## 验证与证据

- `analysis/native-api-isolated/voice-tts-export-v2.json`：解析、空文本、批量空列表/空闲，及不完整export host。后者是测试准备不足，不是原版缺陷。
- `tts-binding-v3.json`：真实参考绑定到上传请求；v4缺headers是替身不完整；v5到合成请求；v6自造WAV到实际文件。v1/v2缺绑定不能代替v6结果。
- `analysis/single-tts-frontend-v1.json`：6组原单条handler/kt，包括生成失败、生成异常、保存false、保存异常、空白。不是DOM测试。
- `analysis/native-api-isolated/tts-save-v1.json`、`tts-reopen-v1.json`：原API真实保存与独立进程重读。
- `analysis/media-frontend-v1.json`：9组批量handler及9组导出handler。原文切片及摘要在media-rea-slices/provenance.json。

这些验证没有真实合成语音、没有批量执行线程、没有真实窗口操作。CM对应界面、模型目录及完整链路尚未成套验收。
