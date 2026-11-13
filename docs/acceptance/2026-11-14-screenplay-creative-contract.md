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

前端49/49测试通过，含真实Chrome组件交互和请求观察；`npm run build`成功。Go screenplay/storage/luoshui全套、prompt的Creative/Luoshui聚焦、httpapi的screenplay聚焦通过。真实MySQL8.4.11、utf8mb4独立临时库验证两次迁移、重读冻结正文、修改库不影响项目、原JSON与可编辑正文保真、账号隔离、冲突和清空；临时测试库已删除。全httpapi曾在路由修复落盘前的旧编译快照触发模式冲突，修复后的相关全套已通过；不声称全仓库测试全绿。

## 本地运行与未完成验收

4001在本轮被其他工作切换为批量书籍版本，因此保留该进程。新版运行在 `http://127.0.0.1:4004/comic-screenplay`：PID20344（记录时）、构建ef766a3a、入口`/assets/user-cXejC1Le.js`，Go embed、健康检查及runtime-build-info均核实。使用独立数据库`qiantie_screenplay_review_20261114`、独立Redis14号库，只复制当前账户已有登录记录，不复制生产任务。

真实签名API已验证四类目录、四选择保存重读、错误用途、冲突、清空恢复和工作区同步。原目录10套画风仅导入该验证数据库当前用户私人库，10张受保护预览全部读回成功，没有全站公开。记录在案卷`runtime-creative-acceptance.json`；Craft运行证据`ev_81acce63c4b92081e989bcbb688d289a241b15b3b921404f51fd104d0720640d`记录验收脚本退出0，不是原版生成捕获。

**页面验收等待登录。**4004与4001的浏览器缓存分开，新地址已打开登录窗口；没有读取浏览器口令或绕过登录。不能把组件浏览器测试或签名API当成已登录页面验收。

**原版付费生成/取消和最终HTTP尚未对照。**原版AX可以读取1212当前配置，但点击报窗口几何不可用、截图FrameArrived超时。四次文本授权已记录，本轮实际文字/图片/视频付费请求均为0。当前完成的是证据范围内的配对实现与自动化/数据库/API验收，不能声称洛水全部页面、全部运行行为已一模一样。
