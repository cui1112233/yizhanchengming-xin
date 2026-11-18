# 仿写长篇专项验收

范围：根据用户给出的仿写创作截图，补专用入口、表单、参数保存、拆解与蓝图/续写约束。本文件不把小说六模式、资料文件对象、原生私有提示词或全部界面判为已复刻。

## 原版依据

原包：`F:/软件/洛水2026/洛水2026/dist/assets/index-C1aA08SG.js`，SHA256 `5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a`。

留存可读文件：`_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/NovelImportPage.readable.js`。默认值与输入在56—95行，上传及最近5章在238—261行，提交合同在300—382行，完整表单在512—750行。`NovelStudioPage.readable.js` 513行后为父页面会话/模型绑定。Craft实际函数证据 `ev_63329cef65a23ea168249c352d8eb2015d355052a877c1a2a4a411d08cc6c514`；本轮继续使用已打开的相同哈希目标，并更新实施计划。

## 验证登记

| 检查 | 已有证据 | 边界 |
|---|---|---|
| 表单与字段 | 5000字/100章、可空标题、原文/上传/选章/偏好/想法/方向/补充/文风/备忘录；78项受控浏览器回归通过 | 假接口，不当作模型成品 |
| 取消和恢复 | 保留原会话、草稿；重复点击保护；未知结果只GET核对；404后同编号明确重发；切账号迟到回复不抢页面 | 不自动重跑不明结果 |
| 长参考草稿 | IndexedDB大文本及账号隔离测试 | 本机磁盘/浏览器故障仍须明确提示 |
| 持久保存 | 真实隔离MySQL创建/快照/摘要/回滚/回放/标题更新测试PASS159.81s | 编译于分段调用及128字符内部哈希修复之前，仅证明当时短参考链路 |
| 参考分段调用 | >50000参考实际请求检查、分块独立调用记录、已完成块复用、unknown停止、保存失败复用raw | 本地假模型；CM独立算法 |
| 长辅助输入 | 不缩原上限，长要求/文风分段提炼；保存摘要及源指纹；可在作品内查看摘要 | 摘要不保证保留原文的所有细节；完整原件仍保存 |
| 4001实际页面 | 候选a207e34f、入口user-KUxtTP8z、进程54332；实际上传8章默认最近5，自选1/3按源顺序回填、取消/重开/刷新恢复均通过 | 未点击有效创建，不冒充真实模型生成 |
| 弹层取消后可点击 | 实际elementFromPoint命中按钮、取消后再次打开通过；新增命中检查的两项页面测试通过 | 先前仅调用DOM.click不足以证明真实鼠标可达 |

日志统一放在 `_local-archive-20261028/long-imitation-delivery-20261118/`；完整受控回归原日志在 `_local-archive-20261028/long-imitation-browser/full-final-regression.log`。SQL原日志在 `_local-archive-20261028/novel-parity-20261117/novelparity20261117f4621da7f650-20261118-190523-test.log`，测试库及临时用户清理后均为0。128字符外部编号通过新SQL mock及内存边界测试，不混记为旧真实SQL覆盖。

## 实现与证据定位

- 表单、默认值、完整参数及原请求恢复：`frontend/src/user/pages/novel-creation/LongImitationDialog.jsx:18`，上传`:99`，提交`:150`；`longImitationDraftStore.js:1` 为按账号隔离的IndexedDB记录。
- 原子创建与完整快照：`backend/internal/novelcreation/long_imitation.go:131`；`long_imitation_store.go` 负责MySQL创建/查询合同；原始字段没有写在前端系统提示词里。
- 逐块八维拆解和分层汇总：`backend/internal/novelcreation/long_imitation_pipeline.go:61`。小参数/要求提炼/缓存与超限入口：`long_imitation_chunking.go:46`。后续模型请求由 `pipeline.go` 的callJSON传有界brief与已保存分析，系统定义位于 `backend/internal/prompt/novelcreation/long_imitation.go:5`。
- 只读重开与摘要展示：`frontend/src/user/pages/novel-creation/LongImitationContext.jsx:8`；服务器summary仅按需读取全量参考原文，不每次自动加载。
- 取消后残留遮挡的原因和修复：`frontend/src/user/pages/NovelCreationPage.css:503`。原 `display:none!important` 截断Antd退场动画，留下全屏wrap；现在保留动画，仅退场modal/wrap不接收鼠标。实际修复前命中ant-modal-wrap、修复后命中按钮且再次打开成功，见 `actual-ui-checks.json` 与 `frontend-pointer-focused.log`。

同一最终后端源码的专项/领域/相关HTTP日志均为 `backend-*-final-20261118-194333.log`，全部exit0。包含>50000实际两块内容顺序、5块4路分层、分析/要求/蓝图路由继承、初回包/修复回包的语义及raw未知字段/截断超限；跨Retry的provider调用计数保持1或2，不出现第三次收费。源码独立复审指出的超限P2已关闭。

候选包和index/二进制hash见 `runtime-manifest.json`，真实运行身份见 `runtime-verified.json`。在全站队列状态均未运行后，仅停止已确认的本地旧进程52688，启动归档候选包；不是公开部署。免费接口空值拒绝、纯preview前后作品数均为1（`free-api-checks.json`）。UI上传/选章不创建作品，验收文件与内容已从草稿移除。原截图及同1920×1080实际对照分别为 `original-reference.png` 和 `final-default-1920.jpg`；正常窗口尺寸截图为 `final-default.jpg`，临时viewport已复原。保留统一玻璃主题，未宣称像素完全相同。

## 最后一轮回归状态

弹层修复前完整受控回归78/78通过；加入真实鼠标命中判断后的两项LIM页面复验通过。弹层修复后完整重复回归另记于 `post-overlay-full-regression.log`。其中三项专用弹窗用例的第一次失败停在初始草稿读取：参考输入仍空且disabled、仅有文风/memo GET、没有POST，未进入账号切换/预览等待测操作。另有浏览器未开放检查端口的环境失败。只对浏览器启动及初次IndexedDB准备增加45秒等待；启动可明确配到60秒，业务断言没有改。

测试启动参数参照[Playwright官方Chromium启动设置](https://github.com/microsoft/playwright/blob/main/packages/playwright-core/src/server/chromium/chromiumSwitches.ts)，避免临时隐藏测试窗口受后台限速和Edge兼容重启影响；没有修改用户正常浏览器的设置。第一轮 `cold-draft-targeted-recheck.log` 仍失败并保留原日志；后续 `chromium-isolated-recheck.log` 三项全部通过，`chromium-page-failures-recheck.log` 三项通过、一项仍未启动，`remaining-intent-recheck.log` 单项通过，`compact-settings-recheck.log` 最后一项通过。

最终登记：弹层修复后的完整首轮 **78项，70通过、8失败、0跳过，exit1**；其中3项初始草稿准备超时、5项浏览器检查端口未就绪。八项均已逐项补跑通过。`combined-browser-verification.json` 按原测试名逐项对照原日志和最后通过证据，最终 **78项都有通过记录，未解决0**。这是补跑后的完整覆盖，**不是一次干净的78/78全量运行**；原失败日志保留。测试准备时间与隔离进程参数变化没有修改应用代码或业务断言。该78项覆盖当前实现的回归范围，不等于原版全部功能都已经复刻。

## 当前未证明

没有使用新增文字、图片、视频额度；没有真实生成5万字正文，也没有对原生私有提示词或native内部算法作一致性承诺。原本地统计公式及八维定义有来源，实际AI拆解、长要求提炼和长篇生成提示词为后端prompt模块中的CM独立实现。其他五模式及CORE-02/检查点/资料侧栏继续按总登记表逐项推进。
