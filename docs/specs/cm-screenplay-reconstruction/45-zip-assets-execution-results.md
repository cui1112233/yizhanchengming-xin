# ZIP第四批：资产解析、画风与保存

日期2026-11-17，继续计划41。本批只做原版调查、受控执行和规格，没有修改产品代码，没有更新/重启4001，没有付费文字、图片或视频请求。固定ZIP身份不变，不混用安装版新页面。

## 本批完成

- 24组原生解析场景：三类正常、空回复、非法JSON、模型异常、作品模型来源、重解析保留旧图、空列表及重复操作。
- 原生资产保存9组及新进程重开：明确整份替换与内部解析合并的差别。
- 11组原前端资产/画风/模型状态加3组合并规则，调用真实原生存储，再新进程重读。
- 9组有效图片入口参数记录：三类资产、三视图开关和参考画风图片；只截获参数并返回明确的模拟失败，没有出图。
- 48项记录校验通过。包括复现原版缺陷，不代表CM已修复或真实UI已验收。

## 已确认的关键规则

资产单独重解析读作品里保存的文字模型；给请求添加model不能代替保存配置。角色、场景、道具解析结果与图片/音色资产分开保存。正常重解析同名资产可保留旧图；公开资产保存却是整份替换，部分提交会清空其他类别。CM必须明确接口用途。

角色三视图会实际切换prompt并强制16:9，单人图、场景和道具使用作品比例。所选画风的本地参考图会转成style_images并带style_image_paths传到出图入口。本组纯文字画风描述没有出现在该入口的prompt；不能把“没有拼接文字”误说成“完全没传画风”，也不能把有名称当成参考图已发送。

恢复三份原生解析system正文：角色17476字符、场景11965、道具7578，共37019字符，已存readable-native-agent-prompts。当前后端prompt目录TXT精确摘要未命中，属于明确需继续对齐的范围；不据此断言运行时组合全文不存在。

## 本批新增修正规则

- ZD11：公开资产保存接受null/列表/字符串并清空资产，成功返回；CM必须类型校验，失败不改旧数据。
- ZD12：原前端成功读取空资产后仍保留旧资产；CM区分成功空结果和加载失败。
- ZD13：画风保存返回失败仍提示已切换；CM明确临时状态和持久化状态。
- ZD14：出图入口明确失败后，原生成Agent仍返回success=true且0张；CM按实际图片与失败项反馈。

原Xo/Zo已有真实失败检查并给保存警告，应保留这种有效设计。不能把所有原按钮一概判为假功能。

## 规格与源码差距

13-asset-parsing、14-style-and-three-view、15-asset-storage-and-reopen分别写明输入、输出、边界、失败提示和证据。CM assets_service.go:84–93按类型/解析程度选自己的Registry模板，:118向每块加入styledAssetContent；原版单独重解析的消息/配置来源需要逐项映射。现有模型、档位、稳定ID及保护不可为照搬而删除。实际4001本轮未验收。

站位、场景四视图尚未建立这份ZIP对应入口的证据，只确认本入口角色三视图。主screenplay_agents_run整条工作流、音色、媒体、真实窗口和最终供应商结果仍待核查，不能宣称资产全链路对齐。

## 证据

固定根目录：F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/07-parity-execution/zip-reconstruction-20261116/

analysis/native-api-isolated/asset-parse-v1至v4.json、asset-storage-v1.json、asset-storage-reopened-v1.json、asset-image-prompt-v2/v3.json；analysis/asset-frontend-storage-v1.json；analysis/asset-validation.json；readable-native-agent-prompts/manifest.json和三份正文。脚本在案目录scripts下。

REA实际分析wa/Xo/qn三份原函数切片：533节点、0解析失败；分析摘要ev_9e39d3b0b41a8b64dbee7260cc1ebf203c1088014e75625df2f10d54ef4ac3b1，父记录ev_225dbf2abaf6562838ed5089637f4af1c1a26ae90aa17175f87a53b62a44331d；语义关系不等于完整UI验收。

首次图片入口记录器误用同步函数导致await错误，v1保留为工具错误，不作原版缺陷依据；修正为异步后的v2/v3才用于结论。实际图片生成、供应商接受参数和效果均未验证。原生账号数据和网络访问保持禁止。
