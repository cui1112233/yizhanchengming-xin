# ZIP功能链路：已确认规则与明确断点

本文件承接00/01，属于正在建立的规格，尚未覆盖全部127个接口。O为ZIP恢复前端主文件；编译模块的api.py/database.py行号仅为编译声明信息，不是已恢复的Python源码。

## Z-CREATE：新建作品

前端原函数Ke位于O:124718–124806。输入非空作品名称；内容每行去首尾空白并过滤空行，再拼接换行。发送`create_screenplay`，字段为name、content、script_mode、style、aspect_ratio、content_type、script_master；该分支script_master固定true。

成功返回时从data.id/path更新本地作品。内容非空且有真实path时，按原函数的章节正则识别章/回/节/篇/卷/部及Chapter，创建章节记录并调用`save_screenplay_chapters`。这不是此前安装版“第N集”分集规则的证明。

边界：名称去空白后为空，不进入该保存分支；正文空值是否应阻止创建，当前前端片段不能证明后台规则。前端在创建接口返回success=false时没有明确阻止后续本地插入和成功提示，需作为原版缺陷候选记录，不复制成CM假成功。后台创建、取消及实际文件尚需独立验证。

## Z-MODEL-CONFIG：模型配置保存和加载

界面O:122721–122802。点击保存调用na(false)，O:118280–118293。请求为`save_model_config({project_path,config:{...Jn}})`。加载ta位于O:118261–118279，调用`load_model_config(projectPath)`。

原生实际函数：Api.load_model_config声明api.py:22160；Api.save_model_config声明api.py:22179。已用本ZIP原生函数、隔离自造项目及OWN_*模型标识验证：配置写入项目的`model_config.json`，重读保持给定字段；缺文件加载返回success=true/data={}；缺project_path保存返回success=false/“项目路径不存在”；非对象config字符串仍会被保存并原样重读。没有调用供应商。

原前端默认值Jn：imageGenModel=gemini-3.1-flash-image-preview，imageEditModel=jimeng-4.6，voiceDesignModel=minimax，dubbingModel=index-tts2，characterThreeView=true，showOriginalText=true。它们是原件观察值，不应复制为我们不存在的服务；CM按自己的模型能力映射。

空加载结果边界：原ta对characterThreeView无条件`!!e.data.characterThreeView`，data={}会把该开关改为false，即使初始化为true。需原样记录，不凭初始化值宣称实际首次打开仍为true。

失败：未改动的原na在保存抛错后仍关闭弹窗并显示“模型配置已保存”；success=false和缺接口也一样。10组受控前端函数试验已确认。此为原版缺陷，与用户明确要求“不许假功能、假链路”冲突，CM应真实显示失败并保留编辑内容，登记为有理由的修正差异，不宣传完全一致。

原生保存还接受字符串配置，属于类型边界缺陷。CM应在服务端校验对象和实际字段，失败不覆盖旧有效设置；不得照搬无效设置写入。

证据：`analysis/model-config-controlled-original.json`、`analysis/native-api-isolated/model-config-native-results.json`。前者是真实原前端函数、受控桥接；后者是真实原生配置函数和实际文件。不是完整桌面窗口的按钮点击验收。

## Z-STORYBOARD：原文逐场分镜请求

O:117206–117353实际组装`screenplay_master_storyboard_inference`。字段包括project_path、original_content、scenes/all_scenes、model、prompt_id/content/name、chapter_index/name、style_desc，以及首轮/重试错峰参数。其中首轮base120ms/jitter450ms/max2200ms，重试min250ms/max800ms。

另一入口O:117108–117145还传request_timeout_sec、request_max_retries、http_transport_retries、min_request_interval_sec、skip_failed_scene_retry，不能把两入口直接合并。具体参数表达式和调用位置已写入native-calls.json。

返回后前端检查取消标志，读取failed_scenes及asset_sync，更新场景状态并进入保存。后端实际并发、最终模型消息、自动重试次数、取消是否写入等仍待本ZIP原生函数验证。另一安装版的8轮规划/续写实验不能填入这一规格。
