# ZIP原生分镜生成规格

范围：固定ZIP的Api.screenplay_master_storyboard_inference，编译声明api.py:20281。声明行号不是完整Python源码。本批使用原模块和原解析器，模型客户端替换为不会联网的模拟客户端；配置只有OWN_*测试值。未改原鉴权方法、未读取账户数据。24组原生场景见analysis/native-api-isolated/storyboard-matrix-v1至v4.json。

## 输入和校验

输入project_path、original_content、scenes、all_scenes、model、prompt_content/id/name、chapter_index/name、style_desc；可选错峰和请求重试参数。原前端批量入口O:117206–117426，单个失败场景入口O:117091–117185。

原生空值返回success=false：无路径“请提供作品路径”；无原文“请提供章节原文”；空场景“请提供有效的剧本场景列表”；空提示词“未找到分镜提示词内容”；缺gptApiKey“请先在全局设置中配置 GPT API Key”。这些边界在发送模型前拒绝。CM使用自己的文字模型配置，不能照抄此密钥名称或服务地址作为唯一配置入口。

scenes是本次生成目标；all_scenes是整章场景上下文，单个重试仍可提供全章场景。来源场景支持_source_scene_index/number/name等字段。不能把重试列表中的局部序号直接当作整章序号，已发现原版此处缺陷，见下文。

## 实际模型消息

截获的是原ScreenplayGPTAPI最终交给OpenAI客户端的参数，不是根据界面推测：

- model为所选测试模型；temperature=1、stream=false、max_tokens=20000。
- system为所选分镜提示词，替换画风和原生支持的场景占位符。
- user依次为【章节原文】、整章原文、【当前剧本场景JSON（仅1个场景）】、清理后的当前场景JSON、【处理要求】及生成该场景完整分镜JSON的要求。
- 首轮和失败补充各自使用这一消息结构。fail-once用例中两次消息完全相同，没有额外编造“修复提示词”。

原生支持的CURRENT_SCENE_CONTENT/POSITION、TOTAL_SCENE_COUNT、PREVIOUS_SCENE_CONTENT、NEXT_SCENE_CONTENT和style已用真实调用验证。测试中当前内容为“场景内容：Alice enters.”，总场景数2，前段“无上一片段内容”，后段为Bob场景。测试字符串current_scene/scene_index/all_scenes并非已确认占位符，原样保留，不能扩写成支持列表。

原提示词164已实际送入这一受控原生调用，模拟回复成功解析；这只证明消息传入，不证明真实模型遵循全部164协议。原文摘要和正文文件见03/43。CM提示词必须放后端prompt模块，前端提交选择ID和业务参数；提示词恢复成功不等于它已注册、已绑定或最终请求一致。

## 并发、重试和进度

两个场景各发一次请求，结果按来源场景排列。原前端宣告全场景并发，并传首轮120/450/2200ms、补充250/800ms的错峰参数。受控试验使用零错峰，并跳过休眠，所以不证明真实并发吞吐和时间间隔。

首轮失败的场景再单独补充一次：模型异常、空回复、非法JSON、空分镜及顶层数组在本组输入下均在两次边界调用后失败；fail-once第二次成功；两场中一场持续失败产生3次边界调用，保留另一场有效镜头。skip_failed_scene_retry=true时不补充，只有一次边界调用。

注意三种重试分开：场景自动补充、ScreenplayGPTAPI级重试、SDK/HTTP内部重试。原前端单场重试传request_max_retries=1、http_transport_retries=0、timeout=300、min_interval=0；批量入口未传这些覆盖值。未覆盖时截获OpenAI构造max_retries=3、timeout=300。模拟客户端不执行SDK网络重试，因此边界调用次数不是生产环境真实HTTP次数。

进度通过window.onScreenplayAIProgress：init/parallel、scene_completed、retry_scene_started及对应补充结果、asset_sync_start、finalizing。记录保留完整参数。CM要让用户看见每场状态和失败原因，不能把“处理了全部场景”显示为“全部成功”。

## 输出、解析和保存

成功data含parsed_data.storyboard、scene_summaries、failed_scenes、retried_scenes、asset_sync、初次/最终失败数及mode=storyboard_master_parallel。有部分有效镜头时success仍可为true并带失败列表；全部失败success=false，也保留失败场景原输入和原因。

本组验证：storyboard对象、storyboards对象、JSON代码围栏可解析；顶层数组未被该原生入口接受。不要用前端Qe支持数组来推断原生也支持数组。id/scene_id/shot_number统一从1重排，并添加_scene_index/number/name/local_order。CM应保留自身稳定数据库ID，只对显示顺序重排，避免换序改变作品身份。

原生本身写ai_raw_responses/storyboard_master_ch1_时间.json，含原回复、阶段、场景摘要及失败计数。最终chapter_0_storyboard.json和chapters.json由原前端另行调用保存，见12；不能把原回复日志当作已保存可编辑分镜。

资产：输出存在asset_sync入口。本批带顶层characters/scenes/props的简化模拟回复仍被汇总为资产空列表；返回的updated_files路径在本组空资产例中并未实际创建文件。没有证明全部资产自动同步，需要继续调查资产识别协议和真实文件，不能据返回文件名宣称资产链路完成。

## 取消和原版缺陷

在模拟模型返回前实际调用screenplay_agents_cancel和screenplay_ai_chat_cancel，分别返回“没有正在运行的Agent/调度”，该次原生分镜函数仍产生成功结果和原回复文件。原前端取消确认后设置自己的标志，等返回时丢弃新镜头并恢复界面。此实验不证明在途HTTP真正终止；CM取消应明确停止新增任务、停止补充，保存已经确认的成果，并说明在途请求状态。

重做第二场的确证缺陷：输入_source_scene_index=1，原生scene_summaries.scene_index=1，却给输出镜头_scene_index=0。将这份真实原生结果交给原tt/et合并，再用真实原生文件保存并重开，第一场被替换，第二场旧内容仍保留。属于原版跨层数据错误，CM必须按稳定来源场景ID合并，不能照搬。

失败提示应有可理解的文字，同时保留原模型/解析原因供查看；不得直接把Python内部异常作为唯一提示。全失败、部分失败、已生成未保存、已取消必须分开。

证据关联：原函数O:117091、117206；辅助清理/上下文/替换编译声明api.py:20015/20099/20177；模型封装ai_inference.py:224/265；原生24组、前端15组及38项记录校验。所有付费供应商调用为0。
