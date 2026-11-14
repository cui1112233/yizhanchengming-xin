# 漫剧创作配对：实现与验收记录

本轮代码主提交 `ef766a3a`；后续 `8dddcf1b`、`554b552e`固定四份正文原始字节，防止换系统检出后摘要失配。只提交本轮文件，未推送、未发布公网。完整规格见 `docs/superpowers/specs/2026-11-14-screenplay-creative-contract.md`。

## 原版证据与结论

|结论|证据、文件和定位|适用边界|
|---|---|---|
|样例1212走新版剧本入口，不能混用旧版前端取正文合同|原bundle `dist/assets/index-DszfAkKD.js:56:2152346`；`editor-entry-mapping.md:11–25`记录info、父组件选路和原偏移；Craft `ev_2d4506da486b5507bb2adb83372dde00a4c1c81fa8a4f11d2dcc0b93725ac4b0`|原包SHA256 `2e5c09685ff5bc10160480f963936b56bac38a57a9e9b6bd199492f9437df1e8`；排版片段是派生物，不是原源码|
|剧本144、分镜307、角色213、场景289独立保存与调用|`original-contract-notes.md:15–57`；`fragments/scope-193.js:40–60`→`screenplay_master_inference`；S195:1–33→分镜桥接；H045/H046→model_config快照；同上Craft源证据|可证实桥接参数，原后端最终HTTP地址/报文仍未捕获|
|四套选中正文确实不同于现有固定模板|`isolated-database/decrypted-manifest.json:1–38`；`module-interface.json`记录`database.py`编译元数据：decrypt首行43、get_prompt_content_decrypted首行1310；Craft运行证据 `ev_520a9afb1b0fb31427daf04bcf6c4e164fe6b8eba3c5b049305794d07c940c77`|只在禁止网络、禁止目录外写入的分析副本调用公开方法；没有读取或输出密钥；编译元数据不是恢复的源码正文|
|画风预览图不能默认作为生成参考图|`original-contract-notes.md:47–59`与`_l()`派生片段：只保留风格文字字段；同上Craft源证据|用户明确选择的参考图继续发送；不推断原版隐藏后端行为|
|取消后的迟到响应需要拦截|`original-contract-notes.md:69–71`；S193:61、78–95检查cancelled/章节集合并finally清理；同上Craft源证据|这不是上游退款或HTTP已终止的证明|

上述案卷文件统一位于 `_local-archive-20261028/luoshui-screenplay-contract-20261114/`。REA成功分析了S193单函数载体：`rea-tiny-summary.json`，父证据 `ev_20aedc264ce5fa6d5f6b72a77e32e1db974e4d0d8be6ccf324c9ab8a76c1cc5c`；只覆盖该载体、存在161条语义未知，不能声称完整应用调用图。大范围JS图和native字符串分析超时均如实保留。

## 独立实现落点

- `backend/internal/prompt/creative.go:45`：四套逐字正文、真实名称、用途和摘要；前端不包含系统正文。
- `backend/internal/screenplay/creative_chapter.go:12,38`：所选剧本正文、三类资产变量替换、原始JSON与可编辑剧本分开保存。
- `backend/internal/screenplay/storyboard_service.go:64`：独立分镜选择决定V9/V7/V3协议、保存引擎和时长校验；没有静默换词。
- `backend/internal/httpapi/screenplay_creative_image.go:13`：按目标人物/场景读取名字、描述、资产提示词，与对应设计模板、画风文字、核心要求和本次补充共同组装最终输入。
- `backend/internal/screenplay/mysql_creative_store.go:98`：私人风格、四套正文冻结快照和原项目兼容字段原子保存；按owner隔离、revision防覆盖。8800109为独立配置表，8800110保存原始生成结果。
- `backend/internal/screenplay/batch_executor.go:345`：后续审核/视频步骤按所选协议读取分镜，包含V9。
- `frontend/src/user/pages/ComicScreenplayCreationFields.jsx:43,81`、`comic-screenplay-creative.js:38,60`、`ComicScreenplayMediaPanel.jsx:13`：四选择独立、清空风格、弹窗取消不应用、分步创建失败补存同一作品。媒体页按项目图片生成/编辑模型切换，明确本次临时覆盖。

## 已验证

|用例|结果|
|---|---|
|正常|模拟适配器断言真正收到所选正文、画风、正确目标描述；144 JSON转可编辑场次；307可保存30秒V9；媒体模型B→C→临时D→恢复C|
|空值|无画风可清空并重读；无正文/错误用途在供应商调用前失败；人物/场景有描述时无需手填补充要求；空结构不覆盖已有剧本|
|重复|同一创建草稿复用作品ID；配置失败补存；图片幂等请求不二次提交；旧revision冲突|
|错误|错用途、失效模型、无权限、失效预览引用、非法JSON分别阻止或提示；V7词不会按V9上限保存|
|取消|画风弹窗取消保留原选择；切作品忽略迟到响应；A1/B/分镜模拟服务在取消后返回内容，均不保存新结果|

前期前端49/49测试通过，含真实Chrome组件交互和请求观察；本轮三步UI追加检查以文末更新为准。`npm run build`成功。Go screenplay/storage/luoshui全套、prompt的Creative/Luoshui聚焦、httpapi的screenplay聚焦通过。真实MySQL8.4.11、utf8mb4独立临时库验证两次迁移、重读冻结正文、修改库不影响项目、原JSON与可编辑正文保真、账号隔离、冲突和清空；临时测试库已删除。全httpapi曾在路由修复落盘前的旧编译快照触发模式冲突，修复后的相关全套已通过；不声称全仓库测试全绿。

## 本地真实页面与UI对照更新

用户已登录4004，登录等待已解除。实际页面验收已完成：剧本→资产→缺图确认→分镜；四类提示词名称与私人画风；修改配置后取消、重开仍为已保存值；站位手工新增保存、刷新后名称/集数/场景/说明重读。测试项目creative-review-20261114使用明确标记的手工剧本、人物、场景、道具和站位，不冒充AI生成结果。

### 为什么之前不是一比一

之前仍保留七页签、重复入口和常驻右设置栏，与原版三阶段结构不同。本轮已经改为两行页头、左侧配置胶囊、居中三步、1196px剧本卡区、235px底部下一步；配置放入1176px弹窗；资产分人物/场景/道具/站位/AI音色五类；分镜按集展示，Agent按需展开。保留本项目品牌及既定玻璃主题，不能将其称为原版逐像素复刻。

实际调用Craft Anything读取并操作原版窗口，已走通下一步、站位、缺图确认、分镜、Agent展开/收起。详细证据ID、原件行列/偏移、派生文件行号见 docs/audit/2026-11-14-luoshui-screenplay-ui-walkthrough.md；不是仅阅读技能说明。

站位有独立状态/版本和GET/PUT、分析、生图、上传、标注改图接口；图片参考按标注→人物选择顺序→场景，16:9并传明确尺寸。原版隐藏分析算法尚未取得，分析规则在Go prompt模块独立实现，不能冒充原版同一算法。供应商边界由模拟Provider验收，没有消耗图片或视频额度。

### 文字调用失败与边界

本轮实际提交过1次真实文本逻辑生成请求，但失败；脚本未保留原始HTTP状态/错误体，不能认定具体失败原因。内部尝试次数未知，保守按用户授权的4次全部预留，本轮不再追加请求；图片/视频请求0。案卷text-budget.json为准，runtime-creative-acceptance.json的0仅指该无模型验收脚本，不是整轮调用总数。

独立确认并修复：缺省max_tokens由0改为16000；144模板要求JSON后附转换自检，现已兼容该指定尾部，并保留原始结果、单独提取可编辑正文；A1分段逐段解析，不吞后续集数。正常、非法尾部、取消和旧数据保留由离线测试验证，尚未再次用真实供应商确认成功。

### 未完成的一比一范围

原版像素截图仍超时，没有完成同尺寸逐像素对照；原版正常生成/取消的最终HTTP和产物未取得。Agent多会话、Skill、子会话全部行为尚未对齐。不能宣布全软件、全页面、全功能已100%相同。只提交本地、未推送或发布公网；4001属于其他工作，保持原进程。

## 最终本地构建与实图

最终功能提交 c77c16f6（前序三步/后端主提交f951e77c），4004 PID11560，candidate-044247，Go embed入口/assets/user-BYj46LWB.js；health=true，runtime-build-info实际读回c77c16f6。独立MySQL qiantie_screenplay_review_20261114、Redis14，未重启4001。

根审重新运行前端65/65、0skip、0fail；UI构建成功；Go Positioning/ScreenplayTextGeneratorOutputBudget/Professional144聚焦通过。最终review无剩余新P1/P2，不等于全仓库全部测试。

最新真实页面还验证：默认站位页无长前缀表单；卡片进入960px编辑窗；改变站位说明后取消重开仍为原文；新增取消仍1张卡；分析760px先选模式、批量936px先选条目，均取消未执行；分镜0结果编辑器禁用、缺文本模型有明确说明。已恢复正常窗口尺寸并保留预览。

案卷final-ui-acceptance.json记录本地读回构建及逐项操作；final-ui-tests.log为根审65项输出。实图ui-story-final.png（1450参考宽度）、ui-standing-modal-final.png（站位编辑）、ui-review-final.png（恢复正常窗口尺寸）。图片是我们的真实页面，不能当成原版截图。

## UI/icon后续更新

2026-11-14追加原SVG图标和工具布局（74164055）及短视口弹窗修复（6d4de5fe），最新4004 PID55608、入口user-BGoeDwqT.js。66页面+5图标检查通过；实际图标尺寸、五资产、站位footer可见及截图见docs/audit/2026-11-14-screenplay-ui-icons.md。本节更新优先于前面的阶段性构建号，不扩大原版像素/Agent未完成范围。


## 2026-11-14 第二批二级界面验收

功能提交c71112f2，4004实际运行PID66996，entry /assets/user-DQpfXXLl.js；独立qiantie_screenplay_review_20261114/Redis14。首批配置03a009ff与第二批均本地完成，未公网发布。

补齐原三栏完整模型配置、12+6模板、角色6类22字段210选项、真实音色切换、全局收藏/最近/前缀、四密度与规划准备窗。系统提示词位于Go prompt模块，最终模拟请求核对对应真实选项；没有用前端硬编码伪装默认系统正文。生成与规划各自持久请求，回包丢失先读状态，部分失败只补未完成集，已收补资产结果本地落库失败不重复收费；明确新输入与续跑分开，取消及切作品后的旧确认不会提交。

整合：Go三整包通过；前端135/136首次通过，1个Chrome超时文件重跑5/5通过，独立严格JSON信号/确认专项14/14通过。前端build+Go embed通过，最终复核无本批P1/P2。日志在案卷model-config-20261114/integrated-deep-*.log。

实际MySQL/API：前缀保存重读、稳定ID去重、全局/项目读取、重复选择取消、409冲突、空值不写、账号隔离、attempt列表无raw、音色库路由、测试前缀清理还原共10项通过。实际浏览器已查看模型配置、角色设计、分镜词、准备窗和Agent草稿。真实生成attempt仍0，本轮文字/图片/音色供应商调用0。

未验证项：原native隐藏规划/资产分配算法、84条混淆模型与其可用账号、原版自动截图像素对比、全部Agent Skills和深层编辑器。当前账号音色服务未配置，页面已明确禁用理由。保持品牌玻璃主题，不声明全软件一比一。
