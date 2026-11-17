# A02续查｜章节菜单、编辑入口与CM菜单调用

日期：2026-11-17。A02仍进行中。本轮新增的是静态链路和有限函数回放，不是新增一批已通过的页面点击。

## 页面实测遇到的实际情况

恢复原浏览器句柄后，实际URL已变为`http://127.0.0.1:4001/`，显示“登录已失效，请在首页重新登录后继续使用”。本轮没有登录、读取凭据、点击生成或改动作品。前轮25项CM页面观察仍属于前轮，不能称本轮重测通过。后续菜单点击需要恢复合法登录；其他源码调查继续进行，不把整个项目登记为无事可做。

## 原ZIP位置与实际规则

O为`Z/readable-supplied-frontend/pretty/index-CnO3GE3B.js`，Z固定路径见A01。已重新核对全文SHA256为fc78d66444269ecb4af9b6114c649e735c11c7770dbfa0cc1d262fd2f3168335。行号属于可读发布前端，不是作者工程源码。

| 编号 | 原入口与文件位置 | 输入、行为与边界 | CM对应及待验 |
|---|---|---|---|
| SRC-A02-001 | 合并章节按钮O120252–120253；st O117433–117437 | 切换合并选择模式，关闭批量选择并清空合并选择集合 | 不能把分镜工具“批量”直接当作此入口；章节管理的对应控件需实测 |
| SRC-A02-002 | 批量生成剧本按钮O120245；pt O117474–117478 | 进入批量选择模式，关闭合并模式并清空选择 | CM分镜批量按钮在StoryboardStage.jsx:40打开BatchTools，功能对象和原章节剧本批量不能仅凭名称归并 |
| SRC-A02-003 | 合并确认ct O117438–117472 | 少于2章提示；确认前说明不可撤销；按原序拼标题/正文，尝试合并镜头结果；取消不改内容；完成调用se保存 | 原源码合并不是本轮操作用户数据；CM保存及恢复仍待验，不能因原提示不可撤销就照搬数据风险 |
| SRC-A02-004 | 批量执行ht O117479–117496 | 无选择/全部空正文提示；过滤空章；确认提示已有剧本会覆盖；Promise.allSettled并发je；结束按fulfilled/rejected计数 | 原函数据Promise结果计数，不据此证明模型输出有效；CM须核对实际生成与保存，不能只检查任务没有抛错 |
| SRC-A02-005 | 添加章节O120261；Yt O117915–117921 | 进入新增章节的本地编辑状态；正文与确认提交在后续函数，尚未验证完整新增链 | 作品库及章节编辑的对应弹窗待登录后检查 |
| SRC-A02-006 | 选择风格O120195；Vn O118161–118177 | 打开风格选择并调用load_assets(styles)读取列表；所选文字/参考图进入生成的链路见既有画风规格，不把打开列表等同参数已发出 | CM画风入口前轮只观察，当前未选择/保存；A03继续核对 |
| SRC-A02-007 | 剧本编辑O120232–120234；pa O118473–118523 | 无projectPath发startCreate；有路径先check_dongman_permission；准备/读取分镜并进入编辑器；某些保存失败仍允许进入 | 只静态检查原权限入口，不运行/伪造授权。CM使用自己的权限和保存结果，编辑器真实保存重开待验 |

## CM现有菜单定位（源码观察，未在本轮点击）

- `frontend/src/user/pages/ComicScreenplayStoryboardStage.jsx:40`：更多分镜工具分为media→媒体与剪辑、production→自动制作任务、files→导入/导出、settings→模型与创作配置，统一交给onTools。还需追父组件处理及实际二级页面，不能判全部有效。
- 已补到父级`ComicScreenplayPage.jsx:100、648、654、657、659、660`：onTools绑定openTool；settings走openSettings，files渲染FileTools，media渲染MediaPanel，production渲染ProductionPanel；编辑入口设置集号后打开storyboard-editor。media/production打开后用hidden保留挂载，files只在当前工具时渲染，因此关闭重开状态不能假设一致。源码已定位具体组件，组件内部API/持久结果及登录后的页面仍待验。
- 同文件:50与:57：查看规划打开800px弹窗，按实际chapterId读取planning_text；此处没有触发生成。尚未通过ZIP入口配对。
- `frontend/src/user/pages/ComicScreenplayStoryboardSettings.jsx:49`及:58–63：前缀弹窗、搜索、新建、编辑、删除、导入/导出。选择/保存通过preferenceAction与commitConfig，并非只改页面状态；本轮未提交。
- 同文件:52：规划设置点击后可立即commitConfig修改plan_density。之后实测不能把“选择密度”当无副作用浏览，需要记录并恢复测试配置。
- `frontend/src/user/pages/comic-screenplay-storyboard-api.js:3–4`：偏好GET作品接口与POST动作接口；`backend/internal/httpapi/screenplay_storyboard_preferences.go:25、33、63`有相应处理入口。源码存在不等于当前登录/权限/持久保存全链路通过。

## 有限回放结果

仅执行原st/pt/ct/ht四个函数，使用自造章节、确认选择、保存记录替身和生成替身：10组、12项检查通过，包括模式互斥、少选提示、空内容、取消不生成/不合并、正常合并、部分生成抛错及运行标志复位。没有调用真实模型、原生保存或用户作品文件；不作为DOM/4001/供应商验收。

## 来源和分析证据

脚本：主案scripts/audit_tracker_a02_menus.py、replay_tracker_a02_menus.cjs。

证据目录：`Z/analysis/execution-a02-menus-20261117-v1/`，包含7个完整原函数切片、provenance.json（UTF16位置/行号/摘要）、bounded-label-search.json、cm-source-identities.json、menu-replay.json。

REA实际分析7切片，1355个AST节点、0解析失败；摘要ev_e8323ab9f1f4dae5dc54ec5f1664d881f08ecb40357115ef7fc81023b370cd1c，父分析ev_6aae3786e977f0f724626641f8f97eff0e7d9fa51b987ab018a293ef405bd4b2。动态引用仍未知，不能称完整应用调用图。REA分析在新增回放结果前进行，输入目录当时10文件；以provenance切片摘要定位，后续归档另记。

Craft已归档最终12份切片/分析/回放记录，证据ev_a1a54b4a8bd29050017d7f488073d5b00e7b3354a0de974a7c2e7a3729a0e66c，目录摘要8670f7611d2055f6d7cec5e39cc9d09434d7b54e0736a4a0e422f2daea69b549。这是分析资料身份，不是原窗口像素一致或真实模型功能通过。

对该可读JS全文精确搜索“规划设置、预设补充、查看规划、预设前缀、检查遗漏、开始分析”均为0。此结果仅限定这些确切文字；不证明功能不存在，不据此删CM功能，也不能用另一安装版本的相似标签补成ZIP证据。

## 未完成及恢复起点

当前CM菜单实测等待重新登录；原ZIP真实窗口仍未配对。onTools父级已补到具体组件，下一步追章节/编辑器内部请求与保存，恢复登录后按本表逐项打开页面、记录状态与截图。A02保持进行中；本轮未改产品、未更新4001。
