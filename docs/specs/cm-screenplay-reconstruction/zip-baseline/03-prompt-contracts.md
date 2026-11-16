# ZIP提示词：正文恢复及当前CM对照

## 来源和实际恢复方法

完整REA解包的`洛水2026/data/app.db`只读查询得到84条prompts记录。读取时只选择提示词ID、名称、分类、描述和内容，没有读取账户配置。存储内容全部为不可直接阅读的编码文本，不能把该次目录查询称为正文已恢复。

随后将提示词记录单独复制到隔离SQLite，使用该ZIP自己的Api.get_prompt_content_decrypted及Database.get_prompt_content_decrypted实际读取。API声明api.py:12124，数据库声明database.py:1301；这是编译声明信息。没有替换认证方法，没有使用假提示词正文，没有开放网络或供应商调用。缺失ID实际返回success=false/“提示词不存在”。

84条全部得到非空可读中文正文，总计508357字符。原样UTF8正文已保存于本案`readable-supplied-prompts/prompt-<原ID>.txt`，manifest.json记录原编号、名称、类别、正文摘要、来源函数和自身对应项。正文没有进行品牌替换或重新编写，作为原件证据保留。

## 选择器与范围

本地84条包含多个软件模块，不等于84条全部属于漫剧的同一选择器。其中“剧本分镜”19条。O:118784附近的AIChat通过`get_all_prompts("剧本分镜",1,1000,"")`取候选并映射label/name和value/id；这还不证明后端真实返回顺序或联网同步后的完整集合。

O:125846的创建入口按布尔条件请求“剧本大师”或“小说转剧”。数据库分别有2条和4条。前端条件、分页、搜索、收藏和最近使用仍需逐项对齐，不得把全部84条混进分镜选择器。

## 对照结果及边界

1. 对CM当前`screenplay_recovered_catalog.json`的63项目录按source_id核对，有21项同编号、同名称，正文逐字节一致；包括18条剧本分镜、2条剧本大师、1条音色设计。
2. 扩大搜索到后端prompt模块所有txt文件，共35条ZIP正文有完全相同的文本文件。存在正文不等于已被正确UI/接口使用。
3. 其余项仅表示在该搜索边界内未发现完全相同的txt或目录项；不能直接宣布缺少同名功能，仍需检查组合提示词、其他模块和调用绑定。
4. Craft实际对比了“CM recovered63目录快照”与“ZIP84正文目录快照”，证据`ev_62b5a66d94369403a020a3e2dad5c11dbfa9eaa658d1e55ec3f9779ab6ee7261`。21文件一致、63仅在ZIP侧、42仅在CM recovered侧、0同路径正文变化。这个范围不是CM全部提示词，更不是功能差异数量。

对照文件：`analysis/zip-prompt-body-comparison.csv`、`analysis/zip-prompt-body-comparison-summary.json`、`analysis/prompt-body-diffs/`、`analysis/craft-prompt-body-comparison.json`。早先的opaque目录查询记录保留为历史步骤，不代表正文仍无法恢复。

## 接入规则

每项必须继续确认：选择器类型→原ID→保存配置→读取完整正文→占位符替换→自己的模型请求→解析器→保存。系统正文只能进入后端prompt模块，前端不得硬编码；品牌替换、自己的模型映射和输出格式修正均需单独记录来源及理由。

当前没有将这84条整体加入产品，也没有证明它们全都适用于剧本创作。未完成调用链和输出约束之前，不按名称猜测绑定，也不把本次正文比对标成完整功能验收。
