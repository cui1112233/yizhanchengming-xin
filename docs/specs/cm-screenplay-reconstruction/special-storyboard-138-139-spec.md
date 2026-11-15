# 分镜词138/139：原版可选图片与续写边界

2026-11-15。这里区分提示词文字、原前端实际处理、受控派生I/O、native未恢复部分；不调用模型、不初始化真实项目/账户，不把自动续写设计冒充原版事实。

证据目录：`_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/07-parity-execution/special-storyboard-138-139-original/`。恢复前端`01-readable-frontend/pretty/index-DszfAkKD.js` SHA256 `8c5145ee0f26f594c35c6ca954e24d87dbbb44f5506029343f211a374992e492`。`source-excerpts/manifest.json`保存18组原行范围、逐片段SHA及提示词原SHA；下文JS行号全部是此恢复前端，不是native伪代码行。

独立Craft`reconstruction.cany`记录finding/unknown；可导入`special-storyboard-craft-bundle.json` SHA256 `b779d6009927331f28b67029c74ddb8dbce388fcaecb71a00aa75b23519a54df`。source inventory `ev_35f3fda6e3eed33b4a9b205e7214ed958f30f469f8fd1861c9fc3dc178057c5d`，既有I/O/拒绝记录inventory `ev_7aacbebe1298a2a6019f83c2f3d6d6e8b6d48c6be6ce9197fcb1fe6b897e22f4`。Craft只登记已保存证据，没有再次运行native或把哈希登记冒充native执行。

## 139：提示词冲突不等于软件拒绝空image_prompt

`03-prompts/prompt-139.txt`418–440的结构含image_prompt（428），但508起最后“输出结果”的storyboard示例从515的source_elements、characters/scene/voice_actor/emotion/line到video_desc，没有image_prompt。示例同时具有`json{`前缀及JSON字符串中的展示换行，不能把整个示例字节直接当合法HTTP JSON fixture，更不能自行推断native自动修复。

原前端实际证据如下：

| 链路 | 精确处理 | 可得结论 |
| --- | --- | --- |
| 生成返回440645–660 | 从z.data.parsed_data取数组/storyboard/storyboards；直接T→`{storyboard:T,plan:al(e)}`，调用save_chapter_storyboard_result | 前端不校验image_prompt必填；native若返回缺字段的row，前端保存请求不补造图片 |
| 显示La438751–769 | JSON.parse已保存chapter结果，取array或storyboard；只整理characters/scene/props/lines | 缺image不会使整条被过滤；没有此处补图逻辑 |
| 章节分镜UI453961–454031 | source_elements非空显示来源；`String(image_prompt||scene_desc||"").trim()`非空才渲染图片提示词块 | 缺图有视频/台词仍显示，图片区隐藏；不是拒绝整条 |
| 手动编辑or440942–954 | 深拷贝row后image_prompt=`n.image_prompt||n.scene_prompt||n.scene_desc||""`，video_desc类似fallback | 空image是可表示的编辑状态 |
| 手动保存ir440967–982 | 保留其他row字段，image_prompt=`n.image_prompt||""`，组成storyboard并save_chapter_storyboard_result | 接受空串传给保存桥；未验证native最终落盘 |
| 导入Ur441460–507 | 仅Er441399–401要求顶层为array/storyboard/storyboards；每row扩展`...n`，image_prompt=`n.image_prompt||n.prompt||n.imagePrompt||""` | 没有image、scene、video、line非空校验；甚至全空row也能传保存，是宽松前端，不代表有效模型生成 |
| 导出qr441234–257 | 有image_prompt才导出“画面”；有video_desc才导出“视频”；来源单独保留 | 不需要把video复制成image才能导出缺图row |

因此：**139“原前端与保存桥输入允许缺image_prompt/空image_prompt”已有实际代码和派生I/O证据**。把139仅因prose/最后示例不一致而永久判为不支持，不能准确描述原前端。更强命题“native主生成parser也一定接受/存储所有缺图返回”仍未恢复；不得以这次前端证据声称完整native生成通过。

## 可验证前端I/O

`replay_frontend.mjs`只运行原Er441399–401与Ur441460–507的派生完整函数；独立vm内使用合成章节和fake `save_chapter_storyboard_result`，没有页面bootstrap/原桥/native/模型。五个fixture结果在`frontend-import-io.json`：

1. 139缺image_prompt→返回shotCount1，保存row.image_prompt=""。
2. image_prompt显式空→同样成功传保存空串。
3. image_prompt存在→原样传保存，不改正文。
4. 其他关键字段也缺/空→同样传保存空字段，证明Ur宽松，**不是应当允许无效生成结果的产品要求**。
5. 138合成partial row+顶层self_check/remaining_content→row/source_elements保留，但Ur仅存`{storyboard:u}`，不保存顶层报告/剩余字段。

5/5断言通过。characters均用数组，未覆盖原co字符归一化分支；未证明raw字符串JSON提取、报告剥离或native持久化。fake API是合成捕获，不是原服务返回。

## 已授权native纯helper尝试：拒绝后停止

已有G11 `asset-check-original/native-boundary-probe.json`只证明scheduler方法名存在，包括`_normalize_generated_storyboard_shot`、`_extract_storyboard_items_from_payload`，没有这两个方法的签名或输出fixture。G11资产提示词builder/请求捕获也不是分镜parser证据。

本轮经root明确授权沿同类隔离方式，只计划导入`screenplay.screenplay_agents`、读两方法签名、用`__new__`纯helperfixture；设置合成data目录，拒绝api/auth导入、socket/subprocess、任何SQLite连接、真实账户/配置读、原件及案例外写。实际导入时触发`Api/auth imports forbidden`并立即停止，**没有签名、没有实例化、没有执行这两个helper**。结果`native-pure-storyboard-probe.json` status=stopped、signatures=[]、records=[]，audit_denials有1条import拒绝。未尝试放宽、移除拒绝或重试绕过。此记录没有保存具体被拒绝依赖名，不能进一步猜它是否只是通用库还是账户模块。

Python audit仅覆盖Python介导I/O，不是native系统调用级隔离；脚本限制在记录中明确。没有重新导入已被拒的api模块，没有Ghidra慢扫、没有原软件窗口/真实模型调用。native主parser、helper返回、最终落盘依然未知。

## 138：提示词真实要求

`prompt-138.txt`229–264给storyboard JSON：id、source_elements(string[])、characters、scene、image_prompt（四宫格关键帧）、voice_actor、emotion、line、video_desc（六宫格Grid）。source_elements来源为script element_id或原文段落编号。440–450说明拆分/截断机制；447要求未处理剧情等用户“继续”，449长输入只处理前20%-30%。625再次明确每次仅前20%-30%与等待“继续”。不能把少量已输出镜头自动标整集完整。

590–615自检在JSON外：场景、元素/原文句数→分镜数→比例、台词遗漏0、ANTI-PPT、Grid1黑屏、Grid1–6完整、连贯性、映射、**未处理剧情/剩余概述等待继续**。这是文字报告，不是已恢复的结构化continuation cursor，不能凭报告样式发明native内部状态字段。

## 138当前入口里的“继续”不是已恢复剧情续写

原当前分镜入口请求440594–618带project_path、original_content、scenes/all_scenes、model、prompt_id/name/prefix、planning_text、chapter_index/name、style_desc、source_mode、context_history_count等；没有显示从138报告提取remaining cursor/next element的参数。代码对所有选中prompt共用，未见138专用分支。

确实存在“继续/补全”，但证据限定为失败场景重试：

- DOM459762–796标题“失败场景补全”，按钮“补全全部失败场景”，onClick=vl（459786），disabled取failedScenes是否还有非success。
- vl440112–229筛选failedScenes.uiStatus!=success（440117），缺场景时从scriptResult.scenes补回；无失败项warning。调用cl440075–106，只把失败scene子集再次送`screenplay_master_storyboard_inference`，skip_failed_scene_retry=true；并非发送文字“继续”，没有从“未处理剧情”定位原文剩余。
- 返回rows经ul440066–74按_scene_index覆盖旧同场景结果或直接拼接并重编号；这是一种重试合并，不能当138百分比续写累积协议。
- 单场ml440253+也是失败场景AI补全，不是原文20%-30%后半段续写。

另有su442503–523的“写不完自动续写”，调用的是**screenplay_generate_chapter_plan**，读取complete、plan_chars、rounds，只涉及规划；440580–592还允许规划不完整后仍带入分镜。不能把规划自动续写拿来宣称138分镜正文自动续写已恢复。

## source_elements、自检与剩余的保存/显示

原生成端前端只取parsed_data中的rows，再存`{storyboard:T,plan:al(e)}`（440645–660）。每row source_elements不被剥掉；La保留，DOM453971–74 `join("、")`展示来源。手动导入Ur将数组项String+trim+filter，字符串则按中英文分隔/空白切分（441467–75）；导出qr441243–44同样显示来源。**没有当前前端source_elements覆盖率、去重、连续性或遗漏检测**的已恢复证据，来源存在并不等于整段已处理完。

现有JSON结构nl439996–440021仅恢复storyboard与plan，未恢复completed/remaining/report专用状态。完整恢复前端的literal-search对“未处理剧情”“分镜自检”无命中，结果保存在manifest；这只覆盖静态字面量，不能证明native或动态内容不存在。原前端章节保存不保留top-level自检/remaining字段；native是否另存原raw/report未知。

## 结论与CM边界

139可以依据原前端证据将image_prompt视为可空/可缺，不用编造图片内容；要支持生成仍应按选定协议验证真正需要的video、source、台词等，并保留原字段，不把宽松空row当成功结果。此处是兼容性决策依据，尚不是native端到端或真实模型验收。

138目前只恢复提示词的部分输出要求，以及一般失败重试/规划续写；**没有恢复“20%-30%→等待用户继续→处理剩余→跨轮不丢不重→全篇完成”的可执行原版链**。CM若新增显式部分草稿、剩余元素、继续按钮、报告保存或自动续写，这都是必须单独设计、标注和验收的CM适配，不能仅凭模板一句话称为原版自动续写。保持unsupported或明确partial能力直到完成该门槛，比把首轮当整集完成更符合现有证据。
