# 仿写长篇：用户截图暴露的未交付项

本次是重新核对，不是产品修复。当前运行产品仍6571f506。截图中的专用弹窗未实现；输入区变紧凑不能算完成此入口。原基线features.json的E03已经记为P0缺失，不能因CORE01或输入框样式完成而关闭。

固定原包：F:/软件/洛水2026/洛水2026/dist/assets/index-C1aA08SG.js，SHA256 5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a。以下恢复文件不是原开发仓库；native私有提示词/算法仍未恢复。

| 观察 | 原版文件/行号 | 当前缺口 |
|---|---|---|
| 六入口图标两列，仿写长篇打开仿写创作modal | NovelStudioSidebar.readable.js；NovelImportPage.readable.js:11-14、129-133 | NovelCreationPage.jsx:290只切activeMode，侧栏仍单列无对应入口图标 |
| 暂定标题可留空；原文必须非空且去空白后至少100个UTF16单位 | NovelImportPage.readable.js:162、300-305、512-533 | 通用创作入口先要求作品；长篇仿写只要求referenceText非空，无100门槛 |
| 默认每章5000、目标100章及快捷chips | NovelImportPage.readable.js:56-95、597-640 | 专用目标章数和快选未接入 |
| TXT/DOCX读取、多章识别、最近3/5/10和自选章节 | NovelImportPage.readable.js:238-261、549-579 | 现有通用素材导入不能算这个弹窗的自动回填与选章已实现 |
| 读者偏好、原创故事想法、仿写方向、补充信息、文风参考等更多字段 | NovelImportPage.readable.js:367-381及对应模板 | 当前通用模式DTO没有这些独立字段 |
| 提交完整参数，参考长文按50000分块，携带文风统计和维度 | NovelImportPage.readable.js:300-382 | backend/internal/novelcreation/validation.go:83-85、100-122只保留referenceText，未知字段被丢弃 |
| 父页面绑定当前sessionId/model再执行并处理结果 | NovelStudioPage.readable.js:513-539 | 不能用新弹窗调用通用SendMessage后宣称原创建链相同 |

已复用真实Craft记录ev_63329cef65a23ea168249c352d8eb2015d355052a877c1a2a4a411d08cc6c514（js_function_explain，NovelImportPage.ie），本轮core_evidence_browse再次确认。针对完整脚本调用core_feature_find返回not_supported（该工具按原生代码处理），未将其计作成功JS分析。直接读取既有恢复文件核查了实际字段和值；没有调用模型/运行原应用生成。

实现验收必须整体通过：点击入口→原顺序专用弹窗→输入/上传/选章/字数校验→所有字段冻结保存→当前自家模型实际收到对应输入→创建结果/检查点可重开。取消保留原页面草稿；重复点击不重复建书；错误保留表单；未知结果只查询。先补专用创建DTO和持久参数，再接表单和后续步骤，不能只画壳。
