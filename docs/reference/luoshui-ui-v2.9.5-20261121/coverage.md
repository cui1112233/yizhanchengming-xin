# 原版 UI 覆盖登记

采集日期：2026-11-21。原版：洛水 2.9.5，目录 F:\软件\洛水2026\洛水2026。1920×1050 视口。

已确认 101 个实拍状态，覆盖 16 个一级入口。插件入口 29 项，其中成功进入 25 项。每个状态附截图、页面文字、控件坐标、时间、来源 URL 和 SHA256，见 manifest.json。

## 范围与缺口

- 这是页面参考与截图交互 mock，不是已经实现或验证全部业务功能。热点只是已采集页面之间的参考跳转。
- 没有运行模型、生成图视频、提交作品、保存设置或删除数据。API 密钥输入在截图时遮挡，不采集输入框的值。
- 当前小说作品库为 0；正文生成、审稿结果、生产暂停/恢复/失败、导出成品等依赖作品或模型的状态未采集。
- 上传文件后的状态、付费/权限异常、全部下拉选项与错误组合未穷举；不能据此宣称全功能一比一完成。
- character-personal 与角色库截图相同，登记为 no-visible-transition，不计入已确认页面。
- 插件逐项结果见 plugin-coverage.json；未展开项不得作为已完成页面。

## 页面清单

| 分类 | 实拍 | 进入路径 | 状态 |
| --- | --- | --- | --- |
| 通用 | [首页](screenshots/home.png) | 程序启动→首页 | captured |
| 小说创作 | [小说创作·长篇原创](screenshots/novel-home.png) | 左侧→小说创作 | captured |
| 小说创作 | [小说创作·短篇原创](screenshots/novel-short.png) | 小说创作→短篇原创 | captured |
| 小说创作 | [仿写长篇·完整参数表单](screenshots/novel-long-imitation.png) | 小说创作→仿写长篇 | captured |
| 小说创作 | [仿写短篇·参数表单](screenshots/novel-short-imitation.png) | 小说创作→仿写短篇 | captured |
| 小说创作 | [番外创作·正传与类型](screenshots/novel-spinoff.png) | 小说创作→番外创作 | captured |
| 小说创作 | [番外创作·四种类型](screenshots/novel-spinoff-types.png) | 番外表单→番外类型 | captured |
| 小说创作 | [续写创作·导入章节](screenshots/novel-continuation.png) | 小说创作→续写创作 | captured |
| 小说创作 | [写作风格](screenshots/novel-styles.png) | 小说创作→写作风格 | captured |
| 小说创作 | [导入素材·默认仿写页](screenshots/novel-import.png) | 小说创作→导入素材 | captured |
| 小说创作 | [Agent设置·模型路由与控制](screenshots/novel-agent.png) | 小说创作→Agent设置 | captured |
| 小说创作 | [Agent设置·37项审稿维度（上）](screenshots/novel-agent-review.png) | Agent设置→下滚审稿维度 | captured |
| 小说创作 | [Agent设置·37项审稿维度（下）](screenshots/novel-agent-review-bottom.png) | Agent设置→审稿维度内滚到底 | captured |
| 小说创作 | [Agent设置·自定义审稿与写作提示词](screenshots/novel-agent-prompts.png) | Agent设置→底部 | captured |
| 小说创作 | [定时写作](screenshots/novel-scheduler.png) | 小说创作→定时写作 | captured |
| 小说创作 | [数据看板](screenshots/novel-dashboard.png) | 小说创作→数据看板 | captured |
| 小说创作 | [备忘录](screenshots/novel-memo.png) | 小说创作→备忘录 | captured |
| 资产与工具 | [角色库](screenshots/character-library.png) | 左侧→角色库 | captured |
| 资产与工具 | [个人添加入口（未观察到切换）](screenshots/character-personal.png) | 角色库→个人添加 | no-visible-transition |
| 资产与工具 | [音色库](screenshots/voice-library.png) | 左侧→音色库 | captured |
| 资产与工具 | [音色库·添加表单（未提交）](screenshots/voice-create.png) | 音色库→添加音色 | captured |
| 漫剧与剧本 | [漫剧解说·作品列表](screenshots/commentary-library.png) | 左侧→漫剧解说 | captured |
| 漫剧与剧本 | [漫剧解说·创建表单](screenshots/commentary-create.png) | 漫剧解说→创建漫剧 | captured |
| 漫剧与剧本 | [短剧工作室·作品库](screenshots/studio-library.png) | 左侧→短剧工作室 | captured |
| 漫剧与剧本 | [剧本创作·项目列表](screenshots/screenplay-library.png) | 左侧→剧本创作 | captured |
| 漫剧与剧本 | [剧本创作·剧本步骤](screenshots/screenplay-script.png) | 剧本列表→1212 | captured |
| 漫剧与剧本 | [剧本创作·资产步骤](screenshots/screenplay-assets.png) | 剧本步骤→下一步（未分析） | captured |
| 漫剧与剧本 | [剧本创作·模型配置](screenshots/screenplay-models.png) | 剧本创作→模型配置 | captured |
| 漫剧与剧本 | [资产·场景](screenshots/screenplay-scenes.png) | 资产→场景 | captured |
| 漫剧与剧本 | [资产·道具空状态](screenshots/screenplay-props.png) | 资产→道具 | captured |
| 漫剧与剧本 | [资产·3D导演台](screenshots/screenplay-director.png) | 资产→3D导演台 | captured |
| 漫剧与剧本 | [资产·AI音色](screenshots/screenplay-voices.png) | 资产→AI音色 | captured |
| 漫剧与剧本 | [剧本创作·分镜步骤](screenshots/screenplay-shots.png) | 剧本创作→分镜 | captured |
| 漫剧与剧本 | [选择分镜词·完整列表](screenshots/screenplay-shot-presets.png) | 分镜→分镜词 | captured |
| 漫剧与剧本 | [分镜·查看已生成脚本](screenshots/screenplay-shot-script.png) | 分镜→查看脚本 | captured |
| 漫剧与剧本 | [分镜编辑器·现有项目](screenshots/screenplay-shot-editor.png) | 分镜→编辑（已加载3个分镜） | captured |
| 画布与剪辑 | [momo画布·项目列表](screenshots/canvas-library.png) | 左侧→momo画布 | captured |
| 画布与剪辑 | [momo画布·节点编辑器](screenshots/canvas-editor.png) | 画布列表→1212 | captured |
| 画布与剪辑 | [画布·Agent面板](screenshots/canvas-agent.png) | 画布→Agent | captured |
| 画布与剪辑 | [画布·文件页](screenshots/canvas-file-menu.png) | 画布→文件 | captured |
| 画布与剪辑 | [画布·资产管理](screenshots/canvas-assets.png) | 画布→资产管理 | captured |
| 画布与剪辑 | [绘画·对话创作](screenshots/drawing.png) | 左侧→绘画 | captured |
| 画布与剪辑 | [绘画·尺寸与数量选项](screenshots/drawing-options.png) | 绘画→尺寸/数量（只查看） | captured |
| 画布与剪辑 | [绘画·提示词列表](screenshots/drawing-presets.png) | 绘画→提示词 | captured |
| 画布与剪辑 | [AI-CUT·工程列表](screenshots/aicut-library.png) | 左侧→AI-CUT | captured |
| 画布与剪辑 | [AI-CUT·剪辑工作台](screenshots/aicut-editor.png) | AI-CUT工程列表→23 | captured |
| 画布与剪辑 | [AI-CUT·转场](screenshots/aicut-transitions.png) | AI-CUT→转场 | captured |
| 画布与剪辑 | [AI-CUT·文字稿](screenshots/aicut-transcript.png) | AI-CUT→文字稿 | captured |
| 画布与剪辑 | [AI-CUT·音频](screenshots/aicut-audio.png) | AI-CUT→音频 | captured |
| 画布与剪辑 | [AI-CUT·视觉效果](screenshots/aicut-visual.png) | AI-CUT→视觉 | captured |
| 画布与剪辑 | [AI-CUT·字幕](screenshots/aicut-subtitles.png) | AI-CUT→字幕 | captured |
| 画布与剪辑 | [AI-CUT·Skills](screenshots/aicut-skills.png) | AI-CUT→SKILLS | captured |
| 画布与剪辑 | [AI-CUT·Agent设置](screenshots/aicut-agent-settings.png) | AI-CUT Agent→设置 | captured |
| 画布与剪辑 | [AI-CUT·Agent运行检查器](screenshots/aicut-inspector.png) | AI-CUT→Agent运行检查器 | captured |
| 资产与工具 | [插件中心](screenshots/plugins.png) | 左侧→插件 | captured |
| 资产与工具 | [我的资产](screenshots/my-assets.png) | 左侧→我的资产 | captured |
| 通用 | [反馈列表](screenshots/feedback.png) | 左侧→反馈 | captured |
| 通用 | [反馈·提交表单（未提交）](screenshots/feedback-form.png) | 反馈→提交反馈（仅打开表单） | captured |
| 通用 | [在线客服](screenshots/support.png) | 左侧→在线客服 | captured |
| 通用 | [教程入口](screenshots/tutorial.png) | 左侧→教程 | captured |
| 通用 | [社区](screenshots/community.png) | 左侧→社区 | captured |
| 通用 | [全局设置·首屏](screenshots/settings.png) | 顶栏→全局设置 | captured |
| 设置 | [全局设置·基础设置](screenshots/settings-basic.png) | 全局设置→基础设置 | captured |
| 设置 | [全局设置·API接口（密钥遮挡）](screenshots/settings-api.png) | 全局设置→API接口 | captured |
| 设置 | [全局设置·小说规则](screenshots/settings-novel.png) | 全局设置→小说规则 | captured |
| 设置 | [全局设置·配音设置](screenshots/settings-voice.png) | 全局设置→配音设置 | captured |
| 设置 | [全局设置·Comfyui](screenshots/settings-comfyui.png) | 全局设置→Comfyui | captured |
| 设置 | [全局设置·环境监测](screenshots/settings-environment.png) | 全局设置→环境监测（未安装） | captured |
| 设置 | [全局设置·数据备份](screenshots/settings-backup.png) | 全局设置→数据备份（未执行） | captured |
| 通用 | [消息中心](screenshots/notifications.png) | 顶栏→消息中心 | captured |
| 通用 | [用户协议·只读](screenshots/agreement.png) | 顶栏→用户协议（未接受） | captured |
| 通用 | [社群入口](screenshots/groups.png) | 顶栏→社群（仅查看） | captured |
| 通用 | [工作模式切换](screenshots/workspace-switch.png) | 左上角→工作模式 | captured |
| Agent代理 | [Agent代理·工作区](screenshots/agent-workspace.png) | 工作模式→Agent代理 | captured |
| Agent代理 | [Agent代理·自动任务](screenshots/agent-tasks.png) | Agent代理→自动任务（未启动） | captured |
| Agent代理 | [Agent代理·技能管理](screenshots/agent-skills.png) | Agent代理→Agent技能 | captured |
| Agent代理 | [Agent代理·更多工具菜单](screenshots/agent-more.png) | Agent代理→更多 | captured |
| 插件详情 | [资产参考·插件页面](screenshots/plugin-01.png) | 插件中心→去使用资产参考NEW文生图/多图参考快速出图 | captured |
| 插件详情 | [角色修改·插件页面](screenshots/plugin-02.png) | 插件中心→去使用角色修改批量修改角色名称属性 | captured |
| 插件详情 | [封面生成·插件页面](screenshots/plugin-03.png) | 插件中心→去使用封面生成AI智能生成精美封面图 | captured |
| 插件详情 | [角色三视图·插件页面](screenshots/plugin-04.png) | 插件中心→去使用角色三视图一键生成人物三视图 | captured |
| 插件详情 | [AI-分段·插件页面](screenshots/plugin-05.png) | 插件中心→去使用AI-分段NEW按主轨道 / 文本轨道 / AI对照剧本分集切割 | captured |
| 插件详情 | [AI-CUT插件版·插件页面](screenshots/plugin-06.png) | 插件中心→去使用AI-CUT插件版NEWAI-CUT草稿插件版 | captured |
| 插件详情 | [AI-Shot·插件页面](screenshots/plugin-07.png) | 插件中心→去使用AI-ShotNEW按剧本和字幕从成片里抽宣传高光，只在片段交界下刀 | captured |
| 插件详情 | [AI-Upscale·插件页面](screenshots/plugin-08.png) | 插件中心→去使用AI-UpscaleNEW批量超分视频。支持火山、云端、Picsart CLI 和本地显卡 | captured |
| 插件详情 | [AI-Split·插件页面](screenshots/plugin-09.png) | 插件中心→去使用AI-SplitNEW按画面变化切开剪映主轨道上的每个视频 | captured |
| 插件详情 | [分镜大师·插件页面](screenshots/plugin-10.png) | 插件中心→去使用分镜大师智能生成分镜脚本 | captured |
| 插件详情 | [草稿合并·插件页面](screenshots/plugin-11.png) | 插件中心→去使用草稿合并多选剪映草稿按顺序合成一份 | captured |
| 插件详情 | [剪映关键帧·插件页面](screenshots/plugin-12.png) | 插件中心→去使用剪映关键帧为剪映草稿添加关键帧动画 | captured |
| 插件详情 | [AI-Subtitle·插件页面](screenshots/plugin-13.png) | 插件中心→去使用AI-SubtitleNEW对照剧本修正剪映语音识别错别字 | captured |
| 插件详情 | [AI-Filler·插件页面](screenshots/plugin-14.png) | 插件中心→去使用AI-FillerNEW对照剧本删掉单独成行的呵、嗯、啊、哼、哈哈、哦哦、呵呵 | captured |
| 插件详情 | [AI-双语字幕·插件页面](screenshots/plugin-15.png) | 插件中心→去使用AI-双语字幕NEW把剪映字幕译成另一种语言，新建一条位置更高的字幕轨 | captured |
| 插件详情 | [AI-BGM·插件页面](screenshots/plugin-16.png) | 插件中心→去使用AI-BGMNEW按剧本和字幕从曲库选乐，写入剪映同一条 BGM 轨 | captured |
| 插件详情 | [字幕对轴·插件页面](screenshots/plugin-17.png) | 插件中心→去使用字幕对轴字幕时间轴对齐调整 | captured |
| 插件详情 | [声音克隆·插件页面](screenshots/plugin-18.png) | 插件中心→去使用声音克隆AI克隆个性化音色 | captured |
| 插件详情 | [字幕合并·插件页面](screenshots/plugin-19.png) | 插件中心→去使用字幕合并多条字幕智能合并处理 | captured |
| 插件详情 | [音色设计·插件页面](screenshots/plugin-20.png) | 插件中心→去使用音色设计AI智能设计角色音色 | captured |
| 插件详情 | [小说转剧·插件页面](screenshots/plugin-21.png) | 插件中心→去使用小说转剧中长篇·纯对话版 | captured |
| 插件详情 | [AI改文·插件页面](screenshots/plugin-22.png) | 插件中心→去使用AI改文智能改写文本，优化表达 | captured |
| 插件详情 | [爆款开头·插件页面](screenshots/plugin-23.png) | 插件中心→去使用爆款开头生成吸引人的开头文案 | captured |
| 插件详情 | [女频性转·插件页面](screenshots/plugin-24.png) | 插件中心→去使用女频性转女频小说性别转换改写 | captured |
| 插件详情 | [格式化文本·插件页面](screenshots/plugin-25.png) | 插件中心→去使用格式化文本文本格式整理与规范化 | captured |

## 插件登记

- 去使用资产参考NEW文生图/多图参考快速出图：captured；仅进入界面，未运行处理
- 去使用角色修改批量修改角色名称属性：captured；仅进入界面，未运行处理
- 去使用封面生成AI智能生成精美封面图：captured；仅进入界面，未运行处理
- 去使用角色三视图一键生成人物三视图：captured；仅进入界面，未运行处理
- 去使用AI-分段NEW按主轨道 / 文本轨道 / AI对照剧本分集切割：captured；仅进入界面，未运行处理
- 去使用AI-CUT插件版NEWAI-CUT草稿插件版：captured；仅进入界面，未运行处理
- 去使用AI-ShotNEW按剧本和字幕从成片里抽宣传高光，只在片段交界下刀：captured；仅进入界面，未运行处理
- 去使用AI-UpscaleNEW批量超分视频。支持火山、云端、Picsart CLI 和本地显卡：captured；仅进入界面，未运行处理
- 去使用AI-SplitNEW按画面变化切开剪映主轨道上的每个视频：captured；仅进入界面，未运行处理
- 去使用分镜大师智能生成分镜脚本：captured；仅进入界面，未运行处理
- 去使用草稿合并多选剪映草稿按顺序合成一份：captured；仅进入界面，未运行处理
- 去使用剪映关键帧为剪映草稿添加关键帧动画：captured；仅进入界面，未运行处理
- 去使用AI-SubtitleNEW对照剧本修正剪映语音识别错别字：captured；仅进入界面，未运行处理
- 去使用AI-FillerNEW对照剧本删掉单独成行的呵、嗯、啊、哼、哈哈、哦哦、呵呵：captured；仅进入界面，未运行处理
- 去使用AI-双语字幕NEW把剪映字幕译成另一种语言，新建一条位置更高的字幕轨：captured；仅进入界面，未运行处理
- 去使用AI-BGMNEW按剧本和字幕从曲库选乐，写入剪映同一条 BGM 轨：captured；仅进入界面，未运行处理
- 去使用字幕对轴字幕时间轴对齐调整：captured；仅进入界面，未运行处理
- 去使用声音克隆AI克隆个性化音色：captured；仅进入界面，未运行处理
- 去使用字幕合并多条字幕智能合并处理：captured；仅进入界面，未运行处理
- 去使用音色设计AI智能设计角色音色：captured；仅进入界面，未运行处理
- 去使用小说转剧中长篇·纯对话版：captured；仅进入界面，未运行处理
- 去使用AI改文智能改写文本，优化表达：captured；仅进入界面，未运行处理
- 去使用爆款开头生成吸引人的开头文案：captured；仅进入界面，未运行处理
- 去使用女频性转女频小说性别转换改写：captured；仅进入界面，未运行处理
- 去使用格式化文本文本格式整理与规范化：captured；仅进入界面，未运行处理
- 去使用长篇拆书长篇小说智能拆分章节：unresolved；已尝试进入；原版宿主窗口最小化，CDP截图超时/不可用；恢复WebView边界与再次采集仍失败，未伪造截图。
- 去使用世界观构建小说世界观设定：unresolved；已尝试进入；原版宿主窗口最小化，CDP截图超时/不可用；恢复WebView边界与再次采集仍失败，未伪造截图。
- 去使用文档合并按章节规则批量合并TXT：unresolved；已尝试进入；原版宿主窗口最小化，CDP截图超时/不可用；恢复WebView边界与再次采集仍失败，未伪造截图。
- 去使用书名生成智能生成吸引人的书名：unresolved；已尝试进入；原版宿主窗口最小化，CDP截图超时/不可用；恢复WebView边界与再次采集仍失败，未伪造截图。
