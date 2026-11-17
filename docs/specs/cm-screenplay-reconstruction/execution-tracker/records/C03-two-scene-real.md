# C03 两个短场景真实试生成（未通过）

2026-11-17，用户授权最多2次文字请求，不生成图片或视频、不自动补试。实际2次，剩余0次。

输入：自写《雨夜的一盏灯》两场短剧本，手工保存v1；项目comic-90c40c413b4d6345a7f9e85a，作品名“并发实测 · 雨夜的一盏灯”。自己的默认文字模型custom-anti-2-gemini-3-8-flash（界面标签[aistudio]gemini-3.8-flash，不能据此认定官方模型身份），分镜模板luoshui_storyboard_164。不带规划、关闭资产补充和历史依赖；未启动额外规划调用。

输出：两个场景均返回HTTP503/model_not_found“当前模型暂不可用”，保存分镜组0个。请求标识分别为202611171353549551198258268d9d6JdEsZUVI和202611171353544950569798268d9d6QV7p1Pbv。请求记录及每场状态保存为unknown，页面显示“需要核对”，继续未完成按钮禁用；未自动重发、未擅自切模型。不能将此算作正常链路通过。

页面实操：创建时检查必填提示词，选择本系统四项模板；创建作品、直接输入已有剧本、保存新版本、切到分镜、选择不带规划、点击开始分镜。失败后关闭弹窗→返回项目库→重新打开同名作品，手写两场正文仍完整。留存测试作品，未删除。

边界修复：检查发现底层通道自带重试/兼容回退会扩大真实请求数；f416827e3f068c45ece63e7a0f1354cc73e377ca给场景请求加SingleAttempt，并在解析通道前传入现有单次调用上下文；输入指纹加入scene-single-attempt-v1。其他文字流程不改。先观察解析层和通道层标志为false的失败测试，再修复通过。整组prompt/screenplay641个pass事件（含子测试）、HTTP定向8个、单次传输9个通过。测试替身通过不代表真实生成成功。

4001已加载上述补丁，PID15324，git_sha与runtime_manifest_sha均f416827e3f068c45ece63e7a0f1354cc73e377ca，保留user-CgpcDZny.js。隔离冻结上轮5b1c9ead快照并只覆盖5个本次后端文件，未混入共享目录其他修改；没有公网发布。

待修观察：顶部显示内部模型ID；手工两场剧本分镜卡片显示0场但后端实际拆出2场；不带规划时底部仍出现“最多8轮”规划费用说明。均登记为待核对，未冒充已修复。503下unknown过于保守，需按明确失败/不明结果区分并另测，不能直接清掉记录重复收费。

下一步：先核实可用文字通道及上述错误状态设计，再取得新的生成预算补验成功回复→解析→保存→编辑回读；图片视频成片及原版全部按钮仍未通过。

证据目录：_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/07-parity-execution/zip-reconstruction-20261116/analysis/execution-c03-two-scene-real-20261117。run-budget.json、trial-state-finished.json、real-result-summary.json、real-generation-failure.jpg及三组测试日志为本轮依据。
