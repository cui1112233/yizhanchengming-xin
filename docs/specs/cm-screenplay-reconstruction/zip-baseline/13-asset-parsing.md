# ZIP资产单独重解析规格

范围为固定ZIP公开接口screenplay_character_parse_agent、screenplay_scene_parse_agent、screenplay_prop_parse_agent，编译声明api.py:21489、21435、21543；内部调用ScreenplayAgentScheduler的角色/场景/道具解析函数，声明screenplay_agents.py:3553、3427、3678。这里不是恢复出的完整Python源码，也不代表screenplay_agents_run整条工作流已验证。

## 输入与模型来源

公开入口读取project_path和content。试验同时传model、style_desc、art_style、selected_style，但这些字段没有成为本入口最终模型来源。未保存作品模型时，截获模型为gpt-4o-mini；给全局gptModel=OWN_GLOBAL_MODEL依然如此。通过真实save_screenplay_ai_config把作品model保存为OWN_SAVED_MODEL后，三个入口最终都使用OWN_SAVED_MODEL。不能把此现象误判成完全忽略模型选择，也不能直接在CM硬编码原版默认模型。

本入口的解析system使用原生内置提示词；user为要求分析小说原文、提取对应资产、返回JSON及中文字段要求，加完整传入content。未观察到把本组style_desc或保存style.desc直接拼进这三个解析system/user。出图画风另有传递方式，见14。工作流入口可使用不同配置，必须另测。

## 原生提示词已恢复

从未经改写的原生解析函数交给模型客户端的真实参数提取，不是重新编写：

| 类型 | 字符数 | SHA256 |
|---|---:|---|
| 角色 | 17476 | 3f48b4d2bdf3829d3df86b5413114f1e058d3101cc0a595ee9dc003d9052ec52 |
| 场景 | 11965 | 947e1a48ae0147231064c388be8670b64cbe99dff925a63c39cc29ded74596ba |
| 道具 | 7578 | 75735262bbf476f63c16d8f2804c2e4b0b7241aa42347c22ec9c5e6f61098d0f |

固定证据根目录readable-native-agent-prompts/内有characters-system.txt、scenes-system.txt、props-system.txt及manifest.json，共37019字符。它们是这次原生调用的system原文，不与此前数据库恢复的84条提示词混为同一来源。

在当前backend/internal/prompt下TXT文件逐字节查找没有完全一致文件；这只证明该搜索范围没有完全相同文本，不证明所有Go内联提示词或运行时组合都缺失。CM当前assets_service.go:84–93按类型/解析程度从Registry取自己的模板，:118另加styledAssetContent；需要逐条映射并比较最终消息，不能用“有角色提取模板”代替原文对齐。

## 输出和失败

模型返回characters/scenes/props对应数组，解析成功返回success=true、data中该数组和“成功解析N个…”提示。本批正常回复保存对应character_parse_result.json、scene_parse_result.json、prop_parse_result.json和agent_assets.json。

空回复返回“AI未返回…解析结果”；非法JSON返回“AI返回JSON格式错误”及解析原因；模型抛错保留错误消息；合法JSON但对应数组为空返回“AI未返回有效的…数据”。v1错误用例没有预存旧资产；v4空列表用例已确认旧资产保留。24组场景包括三类正常/空回复/非法JSON/模型错误、全局与作品模型来源、保留旧图片、空列表和重复解析。

成功重解析同名资产时，native内部保存会合并保留旧image_path，更新提示词等文字字段；其他类别资产仍保留。连续相同解析没有增加重复条目。此行为与公开screenplay_save_agent_assets的整份替换不同，见15，不能混用。

失败必须保留已有资产和引用图片。重复操作应明确是重新调用模型，CM应阻止同一待完成请求重复计费；已有已完成结果不应无提示地自动重新解析。解析成功不等于图片、音色或最终成品已生成。

## 待验收

本批模型客户端为模拟，原生解析/保存函数和文件实际执行；未调用付费供应商。单独重解析的空content、取消、主工作流并发及部分失败、全部提示词协议和真实窗口还未全测；不能宣称资产全流程完成。
