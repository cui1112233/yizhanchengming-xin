# 漫剧制作补齐进展：实现与本地验收记录

日期：2026-11-14。原始审计保留，本文件只追加当前 V88 工作区实现与证据。**不能称全量一比一；没有付费供应商全链或像素一致性结论。**

源码、模拟测试、Go embed构建、本地页面验收、供应商实跑分别记录。逐行“测试”链接表示测试源码。下方补记记录最终冻结身份与本地验收；其余项目仍区分实现和供应商实跑。

已观察前端相关35项测试通过；已观察HTTP Screenplay相关测试通过。后台生产同事报告模块连续10轮通过。后续新增工具仍需最终集成回归。这些证据均不能替代实际供应商执行。

## 必须保留的边界

- 逐项状态代表当前独立实现范围，不代表原软件完全等价。
- P07推理档位未接；S09只有受控单动作助手；Q05不承诺取消已受理上游任务。
- M08主要项目级时间线渲染，独立分集成片仍未完成或验证。
- 剪映草稿仅基础兼容结构，新版剪映实际导入未验证；M12仅本系统资料包。
- 跨作者状态表和既有内容表通过同项目锁与修订号协调，仍非一次跨表事务；数据库中断可能部分内容已保存，应刷新核对，不自动付费重试。
- 没有新增付费供应商测试、没有公网部署或远程推送；前端构建及本地运行验收记录待主任务追加。

## 57项逐项记录

| 编号 | 能力 | 原审计 | 当前实现范围 | 变化与限制 | 源码与测试证据 |
| --- | --- | --- | --- | --- | --- |
| P01 | 本人项目列表与打开 | 已接通 | 源码及模拟接通 | 本人项目读取保留；旧项目路由统一检查权限后解析真实拥有者。 | [源码](../../backend/internal/httpapi/screenplay_routes.go)、[测试](../../backend/internal/httpapi/screenplay_routes_test.go) |
| P02 | 团队作品与权限 | 未接通 | 源码及模拟接通 | 按真实已启用账号分享；查看、编辑、生成分别授权，生成默认关闭，整项目删除仅拥有者；共享模型使用须明确generate授权。 | [源码](../../backend/internal/httpapi/screenplay_authoring_workflow.go)、[测试](../../backend/internal/httpapi/screenplay_authoring_workflow_test.go) |
| P03 | 搜索、时间排序、网格/列表 | 已接通 | 已有实现，本轮回归 | 搜索、日期排序、网格/列表保留；切个人/团队/归档的迟到请求已防覆盖。未做大库分页或性能验收。 | [源码](../../backend/internal/httpapi/screenplay_routes.go)、[测试](../../backend/internal/httpapi/screenplay_routes_test.go) |
| P04 | 历史、归档及恢复 | 部分 | 源码及模拟接通 | 独立归档标记、列表和恢复已实现；归档先暂停后台生产并取消定时，归档期间禁止生成，恢复不自动重启；本人列表过滤归档。近期历史仍是项目历史，不是永久删除恢复。 | [源码](../../backend/internal/httpapi/screenplay_authoring_workflow.go)、[测试](../../backend/internal/httpapi/screenplay_authoring_workflow_test.go) |
| P05 | 新版/旧版工作台切换 | 未接通 | 部分实现 | 保存classic/studio布局偏好并切布局，复用同一真实数据和接口；没有证据证明两套原版工作台所有交互及像素一致。 | [源码](../../backend/internal/screenplay/authoring_service.go)、[测试](../../backend/internal/screenplay/authoring_test.go) |
| P06 | 直接输入、TXT/DOCX 导入 | 已接通 | 已有实现，本轮回归 | 原文直接输入及TXT/DOCX解析保留；新建一次保存，工作区上传与自动保存受编辑权限限制。 | [源码](../../backend/internal/httpapi/screenplay_routes.go)、[测试](../../backend/internal/httpapi/screenplay_routes_test.go) |
| P07 | 文本模型及提示词选择 | 部分 | 部分实现 | 账号文本模型和后端提示词选择保留；High等项目级推理档位仍未接入，不能称全部参数一致。 | [源码](../../backend/internal/httpapi/screenplay_routes.go)、[测试](../../backend/internal/httpapi/screenplay_routes_test.go) |
| P08 | 画风库、预览及后续传递 | 已接通 | 已有实现，本轮回归 | 使用本系统画风目录及后端描述，不复制原软件资源；图片质量和相同样例效果没有供应商实跑证明。 | [源码](../../backend/internal/screenplay/assets_service.go)、[测试](../../backend/internal/screenplay/assets_service_test.go) |
| P09 | 四档资产解析 | 已接通 | 已有实现，本轮回归 | 四档解析提示词和后端映射保留，不据此推断内容质量相同。 | [源码](../../backend/internal/screenplay/assets_service.go)、[测试](../../backend/internal/screenplay/assets_service_test.go) |
| P10 | 画幅及项目设置 | 已接通 | 已有实现，本轮回归 | 项目比例/模型/风格等设置有保存、修订号和权限；关键设置变化会使后续生产失效。比例集合未宣称与原版相同。 | [源码](../../backend/internal/httpapi/screenplay_routes.go)、[测试](../../backend/internal/httpapi/screenplay_routes_test.go) |
| P11 | 显示配置持久化 | 未接通 | 部分实现 | 项目级保存布局、默认标签、字号、导航折叠、说明显示；尚未实现任意面板宽度/全部原版显示选项。 | [源码](../../backend/internal/screenplay/authoring_service.go)、[测试](../../backend/internal/screenplay/authoring_test.go) |
| S01 | 原文查看、修改、保存 | 已接通 | 已有实现，本轮回归 | 原文持久化保留；同原文自动保存与生成共用进行中的请求，避免双写冲突；原文变化从剧本步骤起失效并暂停生产。 | [源码](../../backend/internal/httpapi/screenplay_authoring_workflow.go)、[测试](../../backend/internal/httpapi/screenplay_authoring_workflow_test.go) |
| S02 | 章节列表与逐章存取 | 未接通 | 源码及模拟接通 | 独立章节ID、标题、source_text、script_text及作者状态revision存MySQL；刷新可读，过期修订号拒绝覆盖。 | [源码](../../backend/internal/screenplay/authoring_service.go)、[测试](../../backend/internal/screenplay/authoring_test.go) |
| S03 | AI 章节计划 | 未接通 | 源码及模拟接通 | 实际配置文本服务+后端规划提示词，返回章节列表后保存并可继续编辑；模拟验证，不是供应商规划质量验收。 | [源码](../../backend/internal/screenplay/authoring_service.go)、[测试](../../backend/internal/screenplay/authoring_test.go) |
| S04 | 小说转剧本/导演级剧本 | 已接通 | 已有实现，本轮回归 | A1/B实际适配器及版本持久化保留；没有付费比较相同小说改编质量。 | [源码](../../backend/internal/screenplay/screenplay_service.go)、[测试](../../backend/internal/screenplay/screenplay_service_test.go) |
| S05 | 逐章/逐集生成结果单独管理 | 部分 | 源码及模拟接通 | 逐章保存生成结果，可手工编辑；按章节规范集号后发布活动剧本和结构化场次供主流程使用。 | [源码](../../backend/internal/screenplay/authoring_service.go)、[测试](../../backend/internal/screenplay/authoring_test.go) |
| S06 | 手工改剧本、版本历史与恢复 | 已接通 | 已有实现，本轮回归 | 不可变剧本版本与恢复保留；新旧写入口共用作者锁，修改后清除旧审核并暂停后续生产。 | [源码](../../backend/internal/screenplay/screenplay_service.go)、[测试](../../backend/internal/screenplay/screenplay_service_test.go) |
| S07 | 原文编号、分镜文件处理 | 未接通 | 源码接口与界面已接，待最终集成验收 | 新增编号原文、源SHA和本系统分镜文件导出/导入界面；源摘要、分镜修订号及归属校验有测试；不是原软件专有文件兼容保证。 | [源码](../../backend/internal/httpapi/screenplay_routes.go)、[测试](../../backend/internal/httpapi/screenplay_routes_test.go) |
| S08 | 剧本内 AI 对话 | 未接通 | 源码及模拟接通 | 作品原文/章节/剧本/资产/场次和历史会话进入配置文本模型，持久化用户与助手消息；不再只是跳转链接。 | [源码](../../backend/internal/screenplay/authoring_service.go)、[测试](../../backend/internal/screenplay/authoring_test.go) |
| S09 | 剧本 Skill Agent 与会话恢复 | 未接通 | 部分实现 | 实现受控单动作助手：document/chapter/asset/shot，先展示再明确确认；支持会话恢复和陈旧提案拒绝。未实现原版完整Skill Agent、多工具自主步骤链或任意动作。 | [源码](../../backend/internal/screenplay/authoring_service.go)、[测试](../../backend/internal/screenplay/authoring_test.go) |
| S10 | 改写后关联内容失效/重建 | 部分 | 部分实现 | 新旧内容改动清除下游审核、按阶段标stale并暂停持久生产；请求哈希阻止静默复用旧镜头，重建须明确确认。依赖传播目前按阶段，不是逐章到逐资产的精细依赖图。 | [源码](../../backend/internal/httpapi/screenplay_authoring_workflow.go)、[测试](../../backend/internal/httpapi/screenplay_authoring_workflow_test.go) |
| A01 | 三类资产提取 | 已接通 | 已有实现，本轮回归 | 三类资产文本提取/分批/合并/失败提示保留；仍需真实供应商验收。 | [源码](../../backend/internal/screenplay/assets_service.go)、[测试](../../backend/internal/screenplay/assets_service_test.go) |
| A02 | 新增、编辑、删除资产 | 已接通 | 已有实现，本轮回归 | 资产真实CRUD增加共享edit控制，修改后失效后续步骤。 | [源码](../../backend/internal/screenplay/assets_service.go)、[测试](../../backend/internal/screenplay/assets_service_test.go) |
| A03 | 画风与解析档位生效 | 已接通 | 已有实现，本轮回归 | 后端风格与四档请求测试保留；不从模拟正确推断实际出图一致。 | [源码](../../backend/internal/screenplay/assets_service.go)、[测试](../../backend/internal/screenplay/assets_service_test.go) |
| A04 | 人物/场景/道具参考图上传与绑定 | 未接通 | 源码及模拟接通 | 上传资产/镜头素材并选择绑定，参考图ID经拥有者素材签名进入实际生成请求；模拟验证归属与参考传递。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| A05 | 资产图片批量生成、编辑与历史 | 未接通 | 部分实现 | 图片生成/上传/历史列表/版本选择/重做已有实际服务与持久结果；当前界面主要逐对象生成，未补齐全资产一键批量生产和任意图像编辑。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| A06 | 音色目录及角色配音绑定 | 部分 | 部分实现 | 保存角色、旁白或镜头音色ID并传入实际TTS，产物可试听；新增默认TTS六种内置音色目录，自定义服务无目录时明确提示填写真实ID。未宣称覆盖全部供应商音色或逐音色目录试听。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| A07 | 资产图/集合图拆分与结果整理 | 未接通 | 部分实现，待本地实包验收 | 新增按明确行列拆分集合图、按顺序绑定目标的后端与媒体Tools界面，具有网格拆分/预校验测试；不是自动语义识别人脸或人物分割，最终本地素材实测待补。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| B01 | 按单集/全部集生成分镜 | 已接通 | 已有实现，本轮回归 | 单集和全部集生成保留，配置模型调用与保存有模拟测试。 | [源码](../../backend/internal/screenplay/storyboard_service.go)、[测试](../../backend/internal/screenplay/storyboard_service_test.go) |
| B02 | 多分镜引擎及配置 | 部分 | 部分实现 | 手动页可选两种已有分镜引擎；后台生产仍按固定优先引擎选择，原版全部高级参数未补齐。 | [源码](../../backend/internal/screenplay/storyboard_service.go)、[测试](../../backend/internal/screenplay/storyboard_service_test.go) |
| B03 | 场次增删改、同集排序 | 已接通 | 已有实现，本轮回归 | 场次增删改和同集排序保留，编辑权限及作者锁已覆盖旧入口。 | [源码](../../backend/internal/httpapi/screenplay_authoring_workflow.go)、[测试](../../backend/internal/httpapi/screenplay_authoring_workflow_test.go) |
| B04 | 分镜条目编辑、拆分、合并 | 已接通 | 已有实现，本轮回归 | 逐条编辑/拆分/合并保留，新旧接口持锁避免人工修改与后台单步交错。 | [源码](../../backend/internal/screenplay/storyboard_service.go)、[测试](../../backend/internal/screenplay/storyboard_service_test.go) |
| B05 | 多选、批量调整分镜 | 未接通 | 源码及模拟接通 | 多选、统一时长、复制、删除和排序有真实批量接口；按分镜revision原子CAS，校验归属/选中ID/时长；未宣称全部原版批改模式。 | [源码](../../backend/internal/screenplay/authoring_service.go)、[测试](../../backend/internal/screenplay/authoring_test.go) |
| B06 | 镜头图片生成/重做/版本选择 | 未接通 | 部分实现 | 镜头图片走已有图片服务，历史结果可选择为该镜头当前版本；主要逐对象操作，原版一键多镜头批量和编辑模式未全面覆盖。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| B07 | 镜头音频、旁白与对白回填 | 未接通 | 源码及模拟接通 | 镜头/角色音色、独立台词或旁白TTS和选中音轨进入时间线；字幕及混音细节仍待供应商端到端验收。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| B08 | 参考资产实际进入镜头媒体请求 | 未接通 | 源码及模拟接通 | 稳定素材ID与实际签名参考图进入图片/视频请求，视频绑定引用参与请求哈希；模拟测试覆盖签名刷新不重复提交。 | [源码](../../backend/internal/httpapi/screenplay_media_runner.go)、[测试](../../backend/internal/httpapi/screenplay_media_runner_test.go) |
| M01 | 视频生成通道提交与轮询 | 部分 | 源码及模拟接通 | 真实Submit/Poll适配器保留，新增逐镜头持久状态及任务ID；未付费运行供应商全链。 | [源码](../../backend/internal/httpapi/screenplay_media_runner.go)、[测试](../../backend/internal/httpapi/screenplay_media_runner_test.go) |
| M02 | 图生视频及角色一致性参考 | 未接通 | 源码及模拟接通 | 已有参考图选择传入图生视频请求；这只证明参数链路接通，不证明人物一致性质量。 | [源码](../../backend/internal/httpapi/screenplay_media_runner.go)、[测试](../../backend/internal/httpapi/screenplay_media_runner_test.go) |
| M03 | 逐镜头任务状态、重试与媒体替换 | 部分 | 部分实现 | 保存逐镜头提交/查询/失败/结果和请求版本，状态表与媒体选择可用；不明确的提交标uncertain阻止重提，不会假称可自动恢复所有供应商异常。 | [源码](../../backend/internal/httpapi/screenplay_media_runner.go)、[测试](../../backend/internal/httpapi/screenplay_media_runner_test.go) |
| M04 | 可控旁白/角色 TTS 生产 | 未接通 | 源码及模拟接通 | 独立TTS使用角色/镜头音色、文本及语速，经真实服务保存音频；模拟验证，无新增付费配音实跑。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| M05 | ASR 字幕识别与字幕编辑 | 未接通 | 源码及模拟接通 | 复用ASR服务和AI-CUT字幕轨，支持逐行时间/正文编辑保存；依赖账号字幕服务配置，没有宣称所有字幕引擎等价。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| M06 | 粗剪、裁切、缩放与顺序编辑 | 未接通 | 部分实现 | 片段顺序、时间裁切、开始时间、缩放存入实际时间线；界面未提供全部原版空间裁切/复杂粗剪工具。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| M07 | 漫剧结果进入 AI-CUT 的编辑闭环 | 未接通 | 源码及模拟接通 | 已生成素材可整理为AI-CUT视频/音频/字幕轨，项目拥有者可打开编辑；共享成员在漫剧工作台编辑，不扩展独立AI-CUT权限。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| M08 | 分集成片合并与渲染 | 未接通 | 部分实现 | 真实时间线可提交已有导出队列，后台Finalize等待实际可访问交付；当前主要项目级时间线按集号串接镜头，并未实现/验证每集独立成片输出、逐集标题片尾等完整交付。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| M09 | 媒体包完整收集与下载 | 部分 | 源码及模拟接通 | 打包实际媒体、时间线、字幕和相对路径；资料包下载通过带登录请求获取Blob。大包、TOS/外部媒体全链仍待真实验收。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| M10 | 导出 JSON 项目包 | 已接通 | 已有实现，本轮回归 | 原本JSON数据导出保留并如实命名；与新增含媒体的ZIP资料包分开，不称洛水原生工程。 | [源码](../../backend/internal/httpapi/screenplay_routes.go)、[测试](../../backend/internal/httpapi/screenplay_routes_test.go) |
| M11 | 剪映草稿/可继续编辑工程 | 未接通 | 部分实现 | 独立映射基础剪映草稿轨道、时间单位、素材路径并带说明；引用完整性有测试，未在实际新版剪映导入验证，不能宣称原生完全兼容。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| M12 | 工程包导入与还原 | 未接通 | 部分实现，待本地实包验收 | 新增本系统媒体资料包导入，校验ZIP路径、体积、素材和时间线并重新归属当前拥有者；恢复范围是媒体素材与剪辑时间线，不还原原文、章节/全剧本、团队权限、模型凭证或上游生产任务，亦不是洛水或任意剪映工程完整还原。 | [源码](../../backend/internal/httpapi/screenplay_media_workbench.go)、[测试](../../backend/internal/httpapi/screenplay_media_workbench_test.go) |
| Q01 | 剧本医生/预设医生/分镜医生 | 部分 | 源码及模拟接通 | 三个质量医生调用真实后端提示词并保存pass/block/issues；缺少有效判定失败，不再固定通过。基础结构检查与AI审核区分，无真实供应商内容质量比较。 | [源码](../../backend/internal/screenplay/production_runtime.go)、[测试](../../backend/internal/screenplay/production_recovery_test.go) |
| Q02 | 九步批次状态及日志 | 已接通 | 已有实现，本轮回归 | MySQL九步状态/日志保留，新增持久生产任务、逐镜头记录与后台状态面板。 | [源码](../../backend/internal/screenplay/production_runtime.go)、[测试](../../backend/internal/screenplay/production_recovery_test.go) |
| Q03 | 长篇持续自动推进 | 部分 | 源码及模拟接通 | Redis后台按单元持续推进，不再依赖页面64次循环；测试覆盖超过64单元和运行时替换恢复。长篇供应商负载未实跑。 | [源码](../../backend/internal/screenplay/production_runtime.go)、[测试](../../backend/internal/screenplay/production_recovery_test.go) |
| Q04 | 关页/切页后后台继续 | 未接通 | 源码及模拟接通 | 后台运行上下文独立于start请求，关页不终止worker；Redis任务与MySQL记录支持重启续跑。尚待本地真实进程重启/浏览器验收补记。 | [源码](../../backend/internal/screenplay/production_runtime.go)、[测试](../../backend/internal/screenplay/production_recovery_test.go) |
| Q05 | 暂停/取消供应商任务 | 部分 | 部分实现 | 暂停阻止后续提交，已有供应商任务恢复时续查；适配器未支持上游取消则不能停止已受理任务或承诺停止计费。取消浏览器等待也不能声称结果必不保存。 | [源码](../../backend/internal/screenplay/production_runtime.go)、[测试](../../backend/internal/screenplay/production_recovery_test.go) |
| Q06 | 失败后的逐镜头续查/复用 | 部分 | 源码及模拟接通 | 已知TaskID及请求版本持久化，失败续查复用；未知提交不自动重提，修改后重建需确认；导入已生成视频失败可免费重试导入。 | [源码](../../backend/internal/httpapi/screenplay_media_runner.go)、[测试](../../backend/internal/httpapi/screenplay_media_runner_test.go) |
| Q07 | 定时制作与调度管理 | 未接通 | 源码及模拟接通 | 新增未来时间定时、取消计划、到期原子认领和后台恢复；当前是每项目一个计划，不宣称原版任意复杂调度规则。 | [源码](../../backend/internal/screenplay/production_runtime.go)、[测试](../../backend/internal/screenplay/production_recovery_test.go) |
| Q08 | 生产完成口径 | 部分 | 源码及模拟接通 | 无成片只能clips_ready，completed必须有Finalize产物和可访问交付地址；页面区分镜头就绪与成片交付，仍需真实媒体实跑验收。 | [源码](../../backend/internal/screenplay/production_runtime.go)、[测试](../../backend/internal/screenplay/production_recovery_test.go) |
| Q09 | 相同样例全链路效果一致 | 未验证 | 未验证 | 未运行相同样例的两边付费模型/媒体/字幕/导入恢复全链；原软件内部页面像素对照未完成。不得称全量一比一或给完成百分比。 | [源码](../../backend/internal/screenplay/production_runtime.go)、[测试](../../backend/internal/screenplay/production_recovery_test.go) |

完整源码与测试路径、原审计SHA256及各项验证标记见[同名JSON](2026-11-14-screenplay-implementation-followup.json)。JSON中只对实际验收范围标记 `local_runtime_verified` 并写明范围；其余仍为false。

## 本地构建与运行验收

- 应用源码：`817cca3bb18645e28912952ef493660d7aca72a1`；核心实现`7a863f62`，真实MySQL修复`900026ab`。
- 本地Go embed程序：`_local-archive-20261028/screenplay-input-style/candidate-817cca3b/qiantie-screenplay-local.exe`，4001 PID64004；前端`/assets/user-p7TYt-Jj.js`；health正常，运行身份一致。
- 相关Go三个包测试通过；前端34项测试、React构建、冻结Go embed构建通过。没有宣称全部仓库测试或vet通过；storage既有传锁告警未改。
- 真实MySQL首次验收捕获JSON参数字符集错误；修复后22项API验收通过，覆盖章节发布、编号原文、分镜导入及批量时长、显示偏好、排期/取消、归档/恢复、媒体读取及旁白绑定。
- 登录页面改章节标题并保存、选镜头改5秒、切换经典/完整布局并保存；重启后标题、时长、字号及布局均保留。
- [页面截图](../../_local-archive-20261028/screenplay-input-style/workflow-final-chapters.png)、[API记录](../../_local-archive-20261028/screenplay-input-style/workflow-runtime-acceptance.json)、[最终记录](../../_local-archive-20261028/screenplay-input-style/workflow-final-acceptance.json)。示例作品为“漫剧流程验收 · 窗外的故事”。
- 模型付费调用0次，未推送Git、未发布公网。镜头生产/参考图/TTS/ASR/成片合成等供应商过程仍主要以模拟适配器和本地文件测试验证；未做新的付费整链实跑，剪映新版导入及完整像素对照未验证。
- 冻结包保留原工作区其他既有改动，清单在candidate目录`existing-dirty-baseline.txt`；没有把这些变动混入本任务提交。
