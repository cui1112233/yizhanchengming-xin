# ZIP第六批：音色、配音、视频排队及导出入口

2026-11-17继续计划41。调用REA分析实际原函数，Craft登记证据，执行用户ZIP里的原生模块和原前端函数；本批没有修改产品或更新4001，也没有收费文字、图片或视频请求。

## 实际结果

音色绑定和单条配音已追到实际文件及保存：人物参考WAV→原上传/合成请求→自造WAV响应→原生写audio文件→原单条页面handler及kt→真实update_screenplay_shot→独立Python进程重读同一原数据库及WAV。这里的声音是自造静音，不是模型合成结果。

原单条页面6组、批量配音9组、导出按钮9组受控回放；原视频队列10组、单个Sora2API适配器9组。数量表示本批覆盖案例，不能换算成产品复刻完成率。

本批发现需避免的问题：音色ID不等于参考音频绑定；原单条页面生成成功早于保存，保存false/异常不会改为可见失败；原批量进度查询异常时仍保持生成中。视频查询success=true可能同时status=failed；completed也可能没有可用video_url，不能直接显示成片完成。

## 新增规格和证据

zip-baseline/17-voice-binding-and-tts.md、18-video-queue-and-provider.md、19-export-contract.md，分别写输入、输出、边界、失败、原文件位置及未验证部分。

固定证据根：

`F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/07-parity-execution/zip-reconstruction-20261116/`

原生记录在analysis/native-api-isolated/tts-binding-v3/v5/v6.json、tts-save-v1.json、tts-reopen-v1.json、video-queue-v1.json、video-protocol-v1.json；页面记录在analysis/media-frontend-v1.json、single-tts-frontend-v1.json。测试脚本在主案scripts/，可读原文位置及切片摘要见analysis/media-rea-slices/provenance.json。

REA证据ev_51c9a7a1cef846f3ad8a63664d1b61f6fc6093c787c847a067d8b55a87da3d28，父分析ev_b27926d0cfddaa7ac03cdc5d30d4e0fcc530a78429a496a4830892384c7f6017；3份切片787节点、0解析失败；存在未解析动态关系，非整包全覆盖。

18项记录校验通过（analysis/voice-media-validation.json）。Craft归档18份本批分析/回放记录及切片，证据ev_c04d42909ca00ee58472f1f1946c1d077a9c195810ad3a9eae577fa00a829153，目录摘要b81b9c4bedff53a4c9c24967c1d82a96afdda78d8d94052c09ccd12b0c666033。工具归档标签original-files表示读取当时的文件字节；该目录实际是分析者生成的记录，不是作者源码，不据此升级为供应商或完整运行证据。

补充任务构建试验后，最终记录校验23/23。新增独立目录analysis/voice-media-evidence-v2，共20文件，旧18文件目录保持不变；Craft完整归档证据ev_bdaa10eb293dd4b60c5de55851f9c1f96f7cdee4316c7d3bdbcfd3c8a618d2c3，摘要d53ab9123c424706fb0ed71337a72ec06974d51da2208287ee8c2b6d0db47ca3。同样只是分析记录及原文切片的摘要，不是产品验收。

## 未完成与下一段

追加验证了视频任务构建的实际字段与默认值、参考图标志透传、完成回调的页面事件，见18及video-task-storage-v4/v5.json。正确模型键是video_model_param，不能只照页面名称写video_model。回调与SQLite/新增project.db的关联仍需追踪，不能说完成回调已存入当前分镜。

真正的导出后台尚未经过合法完整初始化；没有生成/重开剪映工程，不以另一模板导出器替代。视频最终请求/参考素材组装、工作线程、下载入库、真实在途取消、批量配音工作线程、音色服务和窗口UI仍待补齐。

全范围仍包含ZIP规划/续写入口、站位与多视图、其余Agent组合、长原文，以及每页/每弹窗/图标的真实窗口对照。不能说链路全部对齐。继续补上述断点和差异表，按计划完成基线后实施，再做4001实际验收。
