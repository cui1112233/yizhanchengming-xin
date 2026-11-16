# 视频排队与一个供应商适配器的协议

2026-11-17。固定ZIP原生方法实际执行，网络替身仅记录请求并返回自造结果。没有创建供应商任务、没有视频下载、没有启动队列工作线程。

## 输入和输出

原`screenplay.sora2_queue.Sora2TaskQueue.add_task`声明103行：shot_id、project_path、shot_data、generation_mode（默认text_to_video）、orientation（landscape）、duration（10）、video_model（sora2_kongyang）。Api.add_video_task声明29009行；原页面调用见O95108、O95404。声明行来自编译元数据，不是Python源码。

队列受控10组：不存在状态→入队→queued→重复入队拒绝→取消→查询无任务→重复取消拒绝→取消后重试产生新task_id→取消全部。源记录`analysis/native-api-isolated/video-queue-v1.json`。

| 情况 | 原实际输出 | CM规则 |
|---|---|---|
| 同镜已在队列 | false，分镜已有任务在队列中 | 重复点击复用/展示当前任务，不重复扣费 |
| 排队 | success,status=queued,progress=0,queue_position | 页面区别排队和生成，展示位置 |
| 取消排队 | success；后续查询success=false/status=none | 不将none自动解释为完成；CM需持久记录取消，且允许重新提交 |
| 找不到任务 | false，无任务/未找到任务 | 提示实际原因并查询已保存结果，不能永远停在生成中 |

原队列本轮只验证内存行为，不证明重启恢复、在途请求取消或跨用户任务隔离。CM应采用现有MySQL任务记录和Redis协调，不能直接复用原进程内队列。

## 实际捕获的一种视频协议

`screenplay.sora2_api.Sora2API.submit_task`声明130行，query_task201行。9组记录`analysis/native-api-isolated/video-protocol-v1.json`，仅适用于这一适配器。

- POST原默认base_url下`/v1/video/create`：images、model、orientation、prompt、size、duration、watermark。本例是portrait、15秒、large、false、空images、自造model/prompt。总超时600秒。
- GET`/v1/video/query`，params.id，查询超时30秒。
- 根层task_id被接受，嵌套data.task_id在本例被拒绝为未返回任务ID。HTTP429显示提交失败。
- queued/processing/completed/failed按实际status透传。failed时查询外层success仍true，只表示查询成功，生成是否成功必须读status和fail_reason。
- completed根层video_url正确返回；本例嵌套data.video_url未提取，出现completed但空地址。未知status也会透传。

CM必须按自己配置的渠道协议解析，不照抄原默认网址或厂商型号。已完成还必须取得可用媒体、保存成功，再显示可导出。此处未验证下载器、工作线程、完成回调写库、全部视频通道，也不构成CM实际视频验收。

## 仍需补齐

Api._build_screenplay_video_task_data声明23961行，新增v4/v5记录的实际执行与字典读取追踪：它保留完整shot，不在该层把资产ID解析为图片；读取aspect_ratio、generation_mode、video_model_param、duration、vidu_concurrency、vidu_offpeak_mode，默认16:9/image_to_video/seedance_2/5/3/false。仅传video_model字段无效，正确键video_model_param才成为任务video_model；正确字段本例得到portrait/15秒/text_to_video/OWN_MODEL。no_storyboard_image和only_storyboard_image作为标志透传，shot.image_path仍在，因此不能认为该层已删除或上传参考图。

_on_video_task_complete声明34355行，在自造结果及接收evaluate_js的窗口替身下发出window.onVideoTaskComplete；现有SQLite分镜记录未更新，另出现project.db。原注册回调在O76256，不能据此直接认定它就是ScriptEditor当前页的保存入口。需要继续确认各页面/任务系统是否共用，以及其实际保存文件。不存在的video_path与存在但非视频的自造文件都不能冒充真实媒体验收；也不能将测试host或数据路径不匹配定为原版缺陷。

需继续追镜头/资产/画风/参考图如何进入最终供应商请求、媒体如何下载入库、取消与迟到回调如何处理，并做同一输入原版/CM对照。只有任务构建、排队和协议测试不能写成视频制作全链路已通。v1/v2新建fixture未获得有效shot，属于测试准备失败；未计入通过。
