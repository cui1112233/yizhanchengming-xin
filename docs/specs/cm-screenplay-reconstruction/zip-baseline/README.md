# ZIP专属规格基线

唯一输入为`F:/软件/洛水2026.zip`，SHA256 `934e43c03115587777050970332deeb0cf653e1c0bfb80b180c9e386f89bcf8c`。当前处于调查和规格建立阶段，尚未完成全范围spec和完整复刻。

- [00 来源及版本边界](00-scope-and-source.md)
- [01 页面、控件和图标地图](01-page-contracts.md)
- [02 已确认功能链路](02-function-contracts.md)
- [03 原版提示词恢复与对比](03-prompt-contracts.md)
- [04 未解决问题](04-open-questions.md)
- [08 新建、分章、保存和重开](08-creation-and-chapters.md)
- [09 提示词选择、排序和收藏](09-prompt-selector.md)
- [10 原版缺陷及修复边界](10-original-defects.md)
- [11 分镜原生生成、消息、重试和结果](11-storyboard-native-generation.md)
- [12 分镜按钮、取消和保存重开](12-storyboard-frontend-storage.md)
- [13 资产解析和原生提示词](13-asset-parsing.md)
- [14 画风和三视图](14-style-and-three-view.md)
- [15 资产保存、合并与重开](15-asset-storage-and-reopen.md)
- [第四批执行结果](../45-zip-assets-execution-results.md)
- [第三批执行结果](../44-zip-storyboard-execution-results.md)
- [本批执行结果](../43-zip-creation-selector-results.md)
- [实际执行结果](../42-zip-baseline-execution-results.md)

固定证据目录：

`F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/07-parity-execution/zip-reconstruction-20261116/`

| 子目录 | 内容 |
|---|---|
| `full-extracted/洛水2026/` | REA完整解包并逐份复核摘要的11014个文件 |
| `readable-supplied-frontend/pretty/` | 可读前端代码，非作者工程 |
| `readable-supplied-prompts/` | 84份原生接口实际返回的提示词正文和清单 |
| `readable-supplied-icons/` | 130份原SVG及引用清单，2个动态组件未提取 |
| `analysis/` | 来源、映射、比较、受控运行记录和校验报告 |

所有静态候选、原函数受控运行、产品实际功能和UI像素对照分开裁定。代码存在、文件一致、数量增加均不等于完整功能通过。
