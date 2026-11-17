# A02续查｜登录恢复、作品库与文件/生产请求

2026-11-17。状态：进行中。用户回复“可以了”后继续原步骤。

## 实际页面结果和中断

通过原浏览器句柄确认首页不再显示登录失效，随后进入`http://127.0.0.1:4001/comic-screenplay`，显示已登录账号和漫剧作品库。实际可见归档库、普通资产库、历史、创建漫剧、个人作品、团队作品、已归档，另有搜索和下拉。本轮没有创建/删除/归档/生成/保存作品。首次返回树尚未包含作品卡片，不能据此判定没有作品。

控件表追加CM-A02-026至032，均为只观察未点击。之后会话的CUA工具被移除，不能继续调用；可用浏览器入口恢复尝试返回`Invalid browser service environment`。没有读取会话/凭据文件、换账号绕过，也没有要求用户重复登录。完成摘要工具同时不可用，未伪造补调用。

本轮已确认登录恢复；剩余菜单实测的当前障碍是页面操作连接不可用。此前登录失效是历史时点，不继续当作当前状态。截图未独立保存，不能说视觉配对通过。

## 继续完成的请求链核查（源码，非新页面点击）

| 功能 | 确切源码入口 | 规则和仍待验部分 |
|---|---|---|
| 编号原文 | frontend/src/user/pages/ComicScreenplayFileTools.jsx:13–14；backend/internal/httpapi/screenplay_files_routes.go:34 | GET /api/comic-screenplays/projects/{id}/source-numbered，返回编号/摘要；页面回显和原文变化待验 |
| 导出分镜文件 | FileTools.jsx:15–20 | 重读工作区/原文，核对一致并要求用户确认编号后下载yzcm-storyboard-file-v1；不是媒体成片导出，实际下载重开待验 |
| 导入分镜文件 | FileTools.jsx:21–22；screenplay_files_routes.go:49 | JSON/8MB/格式/镜头/原文摘要校验，POST .../storyboards/import用expected_revision防覆盖；上传、取消、真实持久结果待验 |
| 生产状态/启动/暂停 | ComicScreenplayProductionPanel.jsx:14–20；frontend/src/shared/api/comicScreenplayWorkflow.js:12–23；screenplay_production_routes.go:44–77 | 分别调用production/status、start、pause；共享阅读不等于生成权限；暂停仅阻后续提交，不能承诺停止上游在途任务，实际队列待验 |
| 定时/重建 | Workflow.js:14–21；ProductionPanel.jsx:27–28 | 明确确认和未来时间校验；可能收费，本轮没有启动。真实队列/恢复待验 |
| 成片判断 | comic-screenplay-workflow.test.js和ProductionPanel.jsx | 镜头完成与真实交付地址的成片状态分开；真媒体播放/导出未验 |

缩写FileTools/ProductionPanel对应上列前端目录；Workflow.js指comicScreenplayWorkflow.js；后端均位于backend/internal/httpapi/。具体文件摘要保存在本轮证据，不以源码存在宣称全部已接通。

## 实际已有测试

后端4项定向测试全部通过：TestScreenplayNumberedFilesValidateSourceRevisionAndOwner、TestScreenplayImportRechecksExactOwnerAfterGrantRevocation、TestScreenplayFilesRecoveredProtocolRoundTripAndUntimedImage、TestScreenplayProductionRoutesSharedReadAndExplicitGenerate。

前端`node --test frontend/src/user/pages/comic-screenplay-workflow.test.js`6项通过，覆盖镜头/成片、媒体地址、时间线顺序、交付地址、任务版本、定时输入。本地内存/受控依赖，不是4001实际请求、真实MySQL任务或供应商生成。没有付费调用、没有产品修改/构建/重启。

证据目录：`Z/analysis/execution-a02-login-resumed-20261117-v1/`（Z见A01），包含后端JSONL、前端原始输出、source-test-ui-summary.json（版本、5文件摘要、界面观察与限制）。

Craft归档3份本轮分析记录：ev_f14f7e50759833be64d5e2be988e7701103147a60d2a627c20f57b9e7063bf90，目录摘要2ef5468e763ae1fd3088b082499e2b847ef9adde3db0133b9f9dade050b3feca。此处归档的是本地测试和观察记录，不是原ZIP作者源码或完整产品验收。

## 未完成及恢复起点

A02继续进行中，无需用户再次登录。页面操作连接恢复后从作品库继续：创建弹窗只打开取消→作品→剧本→分镜更多/批量/前缀→媒体/文件/生产抽屉。工具未恢复时继续原ZIP和组件内部链路核查，不假装已经点击或验收。
