# 2026-11-04 画风/预设种子差异清单

数据来源：本机 `data/system/presets.json`（published 条目）对比 embed `backend/internal/prompt/defaults/` 四个内置库。

## 统计

- 本机全部记录：157
- 本机 published：120
- 内置库记录：51
- 分类 1（线上必定缺失，本机有内置没有）：69
- 分类 2（版本落后，本机 version 更高）：7
- 已覆盖（内置已有同 id 且版本不低于本机）：44
- 本机 published 同 id 多版本重复：0

## 分类 1 / 2 明细（即本次要内置的种子）

| 分类 | id | name | slot | version | previewAssetId | 预览文件 | 内置预览 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1-线上必定缺失 | batch-character-h3 | H3 人物提示词 | batch.character-meta | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-character-meta | CM 资产提取提示词 | batch.character-meta | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-hook-adaptation | 爆款开头改编元提示词 | batch.hook-adaptation | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-opening-meta | 换开头元提示词 | batch.opening-meta | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-original-director | 原文直转导演元提示词 | batch.original-director | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-scene-meta | 场景提示词元提示词 | batch.scene-meta | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-video-h3-director | H3 导演分镜视频提示词 | batch.video-meta | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-video-meta | 分镜元提示词（自动组合） | batch.video-meta | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-video-sd | SD 视频提示词 | batch.video-meta | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-video-sd-fixed-shot | CM固定镜头提示词 | batch.video-meta | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-viral-director | 爆款开头导演元提示词 | batch.viral-director | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-visual-meta | 画面提示词元提示词 | batch.visual-meta | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | novel-analysis | 内容分析 | novel.analysis | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | novel-character | 人物卡 | novel.character | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | novel-outline | 分镜生成 | novel.outline | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | novel-fetch-hook | 爆款优化 | novel-fetch.hook | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | novel-fetch-induce | 诱导排查 | novel-fetch.induce | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-assets-h3 | H3 人物场景道具提取 | script.asset-extraction | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | batch-assets-sd-character-scene | SD 人物场景提示词 | script.asset-extraction | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-extract-assets | 人物场景道具提取 | script.asset-extraction | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-audio-match | 匹配音频规则 | script.audio-match | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-card-protocol | 统一外层分镜卡片协议 | script.card.protocol | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | script-character-focus | 星标人物聚焦规则 | script.character-focus | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-negative-general | 通用负面提示词 | script.constraint.negative | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-negative-sd-detail | 最终导出负面提示词 | script.constraint.negative | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-prefix-2d | 2D 动漫 | script.constraint.prefix | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-prefix-3d | 3D 国漫 | script.constraint.prefix | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-prefix-guoman | 国漫 | script.constraint.prefix | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-prefix-live-action | 真人实拍 | script.constraint.prefix | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-prefix-smart-unified | 智能统一 | script.constraint.prefix | 3 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-quality-4k | 4K 画质 | script.constraint.quality | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-quality-h3-shot-narrative | H3 镜头叙事约束 | script.constraint.quality | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-quality-sd-detail | SD 4K级细节画质 | script.constraint.quality | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-restriction-h3-visual-policy | 画面限制 | script.constraint.restriction | 3 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-restriction-no-overlay | 无关文字限制 | script.constraint.restriction | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | script-constraint-wrapper | 约束设置规则 | script.constraint.wrapper | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-continuous | 连续开头 | script.continuous | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-director-storyboard-master | 导演级分镜母版（小说面板同源） | script.director.master | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-extract | 人物场景提取 | script.extract | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-format-q版 | Q版模式 | script.format.q版 | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | script-format-screenplay | 剧情模式 | script.format.screenplay | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | script-format-shortdrama | 剧本模式 | script.format.shortdrama | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | script-format-shotlist | 分镜模式 | script.format.shotlist | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | script-format-storyboard | 画布模式 | script.format.storyboard | 2 |  | no-asset | n/a |
| 1-线上必定缺失 | script-general | 通用规则 | script.general | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-hook | 爆款开头 | script.hook | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-extract-novel-panel | 小说面板提取 | script.novel-panel-extract | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-quick-director-mode-balanced | 通用小说标准版 | script.quick-director | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-quick-director-mode-concise | 通用小说精简版 | script.quick-director | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-quick-director-mode-detailed | 通用小说较细版 | script.quick-director | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-quick-director-mode-example | 案例学习增强版 | script.quick-director | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-quick-director-mode-reference | 样例对照转换版 | script.quick-director | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-quick-director-mode-strict | 导演级完整成品版 | script.quick-director | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-quick-director-storyboard | 匹配音频 | script.quick-director | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | script-segmented | 分段开头 | script.segmented | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-asset-binding | 分镜资产绑定 | shuihuo.asset.binding | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-extract-characters | 提取人物 | shuihuo.asset.character-extraction | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-character-accessory-sheet | 三视图配饰设计 | shuihuo.asset.character-sheet | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-character-color-sheet | 配色图人物设计 | shuihuo.asset.character-sheet | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-character-expression-sheet | 角色表情多状态 | shuihuo.asset.character-sheet | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-character-single-view | 单视图 | shuihuo.asset.character-sheet | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-character-three-view | 三视图 | shuihuo.asset.character-sheet | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-extract-assets | 人物场景、道具提取 | shuihuo.asset.extraction | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-extract-props | 提取道具 | shuihuo.asset.prop-extraction | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-extract-scenes | 提取场景 | shuihuo.asset.scene-extraction | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-image-prompt | 画面提示词 | shuihuo.prompt.image | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-negative-prompt | 负面提示词 | shuihuo.prompt.negative | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-video-prompt | 视频提示词 | shuihuo.prompt.video | 1 |  | no-asset | n/a |
| 1-线上必定缺失 | shuihuo-smart-segmentation | 智能识别 | shuihuo.segmentation.smart | 1 |  | no-asset | n/a |
| 2-版本落后 | batch-comic-assets-cm-prop | CM·关键道具推导 | batch.comic.assets.prop | 2 | lea_BKKggo8utzN1PwIe | artifact-exists | n/a |
| 2-版本落后 | batch-comic-assets-cm-scene | CM·主要场景推导 | batch.comic.assets.scene | 2 | lea_BKKggo8utzN1PwIe | artifact-exists | n/a |
| 2-版本落后 | batch-comic-assets | CM·漫画资产统一推导 | batch.comic-assets | 4 | lea_BKKggo8utzN1PwIe | artifact-exists | n/a |
| 2-版本落后 | batch-comic-style-03 | 彩色2D手绘Q版条漫风 | batch.comic-style | 2 | lea_ZWy4UqgNh16SmVwx | artifact-exists | embedded |
| 2-版本落后 | batch-comic-style-07 | 治愈系Q版彩色网漫画风 | batch.comic-style | 2 | lea_DZg-36rwlduBI_aM | artifact-exists | embedded |
| 2-版本落后 | batch-comic-style-13 | Q版豆丁2D漫风 | batch.comic-style | 2 | lea_60sljNPSgUfP7ECq | artifact-exists | embedded |
| 2-版本落后 | batch-comic-style-16 | 精致火柴2DQ版漫风 | batch.comic-style | 3 | lea_BKKggo8utzN1PwIe | artifact-exists | embedded |

## 分类 3（仅内置有，禁止复活，仅供核对）

| id | name | slot | version | 本地状态 |
| --- | --- | --- | --- | --- |

## 重复 published id

无。


## 种子落盘

- 种子文件：\ackend/internal/prompt/defaults/admin_presets.json\（76 条 = 分类 1 的 69 条 + 分类 2 的 7 条）。
- 预览图：\ackend/internal/prompt/defaults/admin_previews/*.jpg\（4 张，最长边 800px、质量 82，合计 483,294 字节）。
- 合并规则（\EnsureComicPresets\）：缺 → 补；同 id 版本低 → 升级并归档旧 published；同 id 版本相等或更高 → 不动；仅内置有 → 不复活；重复 id+version → 按 published/更高 revision 收敛。

## 验证结果（2026-11-04，本机临时目录 + 4002/4003）

| 场景 | 总记录 | published | batch-factory published | 重复 id+version | 「Q版豆丁2D漫风」 |
| --- | --- | --- | --- | --- | --- |
| 本机真实文件（基准） | 157 | 120 | 60 | 0 | v2 |
| CaseA 空目录第一次启动 | 127 | 120 | 60 | 0 | v2 |
| CaseA 第二次启动（幂等） | 127 | 120 | 60 | 0 | v2，文件哈希不变 |
| CaseB 仅内置默认值（线上现状模拟） | 127 | 120 | 60 | 0 | v2 |

- \GET /api/presets?module=batch-factory\（4002 临时实例）→ \count=60\，命中 \Q版豆丁2D漫风\ 1 条：\id=batch-comic-style-13, version=2\。
- 预览图已写入临时 artifact 目录：\lea_60sljNPSgUfP7ECq.jpg\，118,853 字节。
- 关键 id 全部存在：\atch-comic-style-03/07/13/16\、\atch-character-meta\、\atch-comic-assets\。
- \go build ./...\ 通过；\internal/prompt\ 中预设相关测试全绿。唯一失败 \TestCatalogReloadsWhenPublishedFileChanges\ 属于并行会话对 \catalog.go\ 新增的 size+mtime 缓存，与本改动无关。