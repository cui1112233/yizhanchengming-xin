# 小说创作测试：2026-11-20

结论：不能验收为“已完整对齐、能真实输出成品”。当前已确定对话动作分流有缺口；其他测试结果和失败现场在本目录登记。

## 实际服务和页面

- 4001 `/health`返回ok；已登录的`/novel-creation`可以打开。实际服务为Go，构建身份b88a02d7，详见runtime.json；不把当前Git中的文档提交号当实际服务版本。
- 浏览器进入前的历史任务“链路实测-1120-anti2”仍暂停、5%、0章，显示上游结果待确认。没有继续/重试该任务。
- 仿写短篇：空原文/99字时开始按钮禁用，100字启用，清空后重新禁用。测试文字已清空，没有点击开始生成；取消后弹窗关闭，作品可见数量仍为3。
- Agent设置：加载完成后37项、默认选中29项；未修改或保存设置。
- 设置首次打开短暂显示“0/0、目录加载失败”，数据到达后恢复为29/37。属于加载提示的真实问题，不能把该短暂画面当目录永久缺失。源码AgentPanel初始化catalog=[]，与空目录提示缺少明确加载状态有关，具体修复后再验。
- 短篇原创、番外入口可以切换模式；本轮没有触发它们的真实生产。结束时已回到原历史暂停任务。

## 已复现的动作分流缺口

用Go overlay临时注入诊断测试，不修改业务源码、不访问模型/数据库。

- 同一份格式正确的回复，action.type=create_book → completed，建立pending action。
- 仅将类型改成short_run → invalid_output，未建立pending action，提示“模型回复格式不正确，已保留原始结果，未创建作品。”

诊断测试的PASS表示**成功复现不兼容**，不是“短篇功能通过”。详见action-probe.log及action_probe_test.go.txt。它仅隔离测试动作类型，不将构造的JSON冒充原版实际模型回复。

原版IDA证据见规格10：AgentSessionService.confirm_action有short_run→run_short_fiction等独立分派；CM当前Studio回复解析只接受create_book，确认随后建立long_original。这不是只有UI不同。

## 测试范围

- 后端：`go test -short -count=1 -json ./internal/novelcreation ./internal/prompt/novelcreation`，665项通过（含子测试）、41项跳过。未设置专用隔离MySQL条件的用例不计通过；本轮不能据此声称真实MySQL端到端通过。
- HTTP：`go test -short -count=1 -json -run=(NovelCreation|NovelStudio) ./internal/httpapi`，8项通过。
- 前端：8个指定测试文件，含浏览器固定数据测试、界面来源/进度/章节导入测试。所有模拟页面使用假响应，不代表上游模型实跑。最终数量见summary.json和frontend.tap。
- 前端最终结果：94项，76通过、18失败，0跳过/取消，整轮退出码1。约28分钟；包含多次浏览器读取超时，不把耗时当产品处理速度。
- 18项失败中，4项是浏览器读取超时，14项是页面断言失败（部分断言内也有等待超时）；后14项仍需区分真实功能问题与固定数据/旧预期问题，不能直接忽略。
- 文档：verify_docs.py与verify_bridge_contracts.py均通过；只证明哈希、范围、条目和引用，不证明全功能正确。

真实模型本轮0次，未生成图片/视频，未改产品逻辑，未部署。历史服务未重启。

## 前端失败的处理规则

保留全部失败，不能为了通过而直接删除断言。当前已看到工具页切换、内部定时会话创建、无作品限制、侧栏宽度、资料入口、草稿模式隔离等失败。

无作品限制用例要求风格/Agent等一律“请先选择作品”，与当前账号默认Agent设置设计有冲突：这部分须按已经确认的使用规则修订测试。其他失败还需逐项查实际页面、固定数据接口和断言，不能全部归因于旧测试，也不能将全部失败直接说成线上功能损坏。

最终测试汇总：summary.json；原始日志backend.jsonl、http.jsonl、frontend.tap；模拟浏览器现场位置记录于summary.json的case，目录为`_local-archive-20261028/novel-task7-browser/`。

frontend.tap提交副本仅去除了各行末尾空格，所有结果文字保留；原始字节副本和SHA见frontend-log-provenance.json，原件位于`_local-archive-20261028/novel-contract-test-20261120/frontend.raw.log`。

`summary.json.frontend.failedTests`逐项保留18项失败的名称、位置和首个错误。分类只是“读取浏览器超时”或“断言失败、待区分产品与固定数据问题”，没有擅自豁免失败。定时和文件专用测试现场还分别在`_local-archive-20261028/novel-scheduler-browser/`及各测试源文件声明的目录。
