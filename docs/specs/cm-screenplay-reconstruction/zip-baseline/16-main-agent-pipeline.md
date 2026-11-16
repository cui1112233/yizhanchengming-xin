# ZIP主Agent调度规格

固定入口Api.screenplay_agents_run，编译声明api.py:21376；ScreenplayAgentScheduler.run_all_agents声明screenplay_agents.py:369。声明行号不是完整反编译源码。本批保留原公开接口、调度、解析和写文件逻辑，用模拟文字客户端、拒绝出图的最终边界和进度记录窗口运行。没有付费生成。

## 输入

原AIChat初始化请求O:119002附近：project_path、content、model、prompt_content/name、script_mode、aspect_ratio、art_style、style_desc、selected_style、segments、image_gen_model、image_edit_model、voice_model、voice_prompt_id、character_three_view、pipeline。缺路径直接失败“缺少项目路径”。

该入口model直接用于摘要、角色/场景/道具解析和剧本生成；本组最终请求均为OWN_SELECTED_TEXT。与13中“单独重解析从作品配置读model”不同，不能混成一条取值规则。script-only捕获system为OWN_SCRIPT_RULE OWN_STYLE_DESC，证明选中剧本提示词和画风文字进入实际请求。

前端对原文按换行组装约1000字符片段，单条超长行不会仅因超过阈值被强行切断；同时传完整content与segments。本批输入短原文，不证明长篇分段、并发和全部跨段合并行为。

## 步骤顺序

前端短内容默认text_input→line_numbering→synopsis→character_parse→scene_parse→prop_parse；长内容默认不含line_numbering。已有initialPipeline时使用用户保存的流程。另有可选screenplay、character、scene、prop、voice、director、shot_numbering、dialogue_check等入口，本批未全测所有组合。

实际验证：

| 配置 | 实际结果 |
|---|---|
| 摘要→角色→场景→道具解析 | 依次执行，4次文字请求，4项completed |
| 道具→场景→角色→摘要 | 依所选顺序执行，4次请求 |
| 角色解析重复两次 | 发2次请求，最终状态对象只保留该key最后状态 |
| 只text_input | 不发模型请求，外层完成，没有生成成果 |
| 未知步骤 | 被跳过，不发请求，但仍返回所有Agent执行完成 |
| 空content | 四个文字步骤均failed，没有模型请求，外层仍success=true |
| 摘要→角色失败→场景→道具 | 角色failed，其他继续completed，保留部分文件，外层仍success=true |
| 单独screenplay | 1次文字请求，保存原回复和ai_results结果 |
| 仅line_numbering，或line_numbering后screenplay | 本输入前者无模型/无步骤事件，后者仅剧本步骤有事件；不能据此认定所有编号能力不存在 |
| 摘要/三类解析后screenplay | 5次文字请求，都使用所选model，相关步骤completed |
| screenplay后角色/场景/道具生图 | 1次文字请求、3次被明确拒绝的出图边界；原版三个图片步骤仍completed，属于错误完成标记 |

同配置再提交一遍会再次执行4次文字请求，本组共8次，没有自动复用。重复步骤与重复提交要区别：用户明确编排的重复步骤可以有意义，但同一待完成任务的重复提交应避免重复计费。

## 输出和失败状态

外层返回success/message及agents_result。各条含name、status、message、result。未选择的步骤仍可在状态表中保留pending，不代表它已经排队或需要执行。

本批解析步骤结果为result.characters/scenes/props，摘要为result.title/synopsis等；不是统一嵌套result.data。原前端完成分支中存在读取result.data的代码，但实时回调$r直接读正确结构，且本批回放确实显示资产。不能只改返回数据包装，或省略进度事件后还宣称等效。

外层success不能作为全部成功标准。CM必须依据选中的有效步骤逐项汇总：全部成功、部分成功、全部失败、已取消分别返回，保留失败原因和有效成果。未知步骤应明确拒绝；只有输入步骤应告知没有执行生成。

## 取消

第一条模拟文字请求返回前调用真实screenplay_agents_cancel，返回已取消；主流程最终success=false/用户取消，后续请求没有启动，总计1次。已完成的synopsis_result.json仍留存。不能把取消等同于回滚删除全部成果。

本批证明停止后续调度，不证明真实供应商在途HTTP已终止。CM取消要停止新增步骤/补充、处理晚到结果，并保留明确已保存成果。voice等其他调度器取消另测。

## 文件及恢复

摘要和三类解析各自保存结果文件，生成资产写agent_assets.json；剧本生成写ai_raw_responses和ai_results。公开Agent调用本身不写前端聊天记录。解析流程没有统一ai_result时，单独load_screenplay_ai_result会失败，但screenplay_load_agent_assets仍能读取解析资产。

原前端实时更新后还需Jo保存综合结果、ya保存历史。已将正常、部分失败、单独剧本三类原事件回放给未改写的原函数，调用真实保存，再换新进程读取；结果与历史一致。这里是分层串接实验，未渲染完整桌面窗口。

证据：native-api-isolated/agent-pipeline-v2.json的13组、v3.json的2组，agent-pipeline-reopen-v2.json、agent-progress-replay-v2.json、agent-ui-save-reopen-v1.json。正常JSON回复为受控样本，不保证真实模型输出质量或全部提示词协议。
