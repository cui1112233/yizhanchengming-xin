# 原编辑器导出入口及未验证部分

2026-11-17。原前端O95746–95790的ll已直接受控运行9组。后台Api.export_to_jianying声明31111行，本轮尚未走过合法初始化和权限状态，未验证实际成品。

## 输入和页面规则

1. 没有分镜：warning“没有分镜数据”，不发请求。
2. 没有任何video_path：warning“没有分镜包含视频，请先生成视频”，不发请求。
3. 只要有一个video_path就发送全部分镜；没视频的项仍带空video_path。后端如何筛选未验证。
4. 请求为{project_path,shots,export_mode,params}；shots逐项包含id/sort_order/video_path/audio_path/voice_entries/subtitle_duration_us。
5. voice_entries从voice_text解析，先filter(audio_path)再map生成line_index、audio_path、character（默认旁白）。因此此处line_index是过滤后的序号；下游如何使用尚未验证，不能仅据此断言错位。
6. 模式3的front_trim_seconds/end_trim_seconds分别取Go/Yo；模式4取Qo/Jo；其他模式params为空。此处只确认数值分支，没有根据数字猜模式名称。

## 输出、失败和取消

- 开始loading=true、progress=0。桥接返回success=true：progress100、“导出完成”、成功消息、关闭弹窗。
- success=false：显示message或“导出失败”，弹窗保持打开；抛错显示“导出失败: ...”；finally loading=false。
- 没有原生桥接时显示“请在应用中使用此功能”。voice_text破损时voice_entries为空，仍可提交导出。
- 本函数无在途取消接口；弹窗关闭/文件选择取消后的原生结果、部分成功、路径冲突、可写性和真实产物均待验证。

## 明确的证据断点

初版隔离host缺少_auth，调用export_to_jianying停在初始化；这不是用户可见原版失败，也不能通过伪造授权放行来补测。后续应使用原合法初始化/会话，或继续静态追踪，不修改原授权判断。

ZIP另有根模块jianying_template_export.JianyingTemplateExporter.export_drafts（声明26行），但尚未证实它就是上述编辑器的调用后端。不能用另一导出器产物冒充本入口验收。

## CM实施约束

按真实配置和项目格式实现，进度必须对应真实处理；校验媒体可读、输出文件可用、导出路径明确；失败保留项目和有效媒体，允许重试导出；原版未确认的字段不得自行补成已对齐。当前既没有CM配对导出验收，也没有更新4001。

证据：analysis/media-frontend-v1.json；media-rea-slices/export.js及provenance.json；native-api-isolated/media-interface-metadata.json与voice-tts-export-v2.json。REA本次三份切片787个AST节点、0解析失败；语义图仍有动态未知，不能称整包分析完成。
