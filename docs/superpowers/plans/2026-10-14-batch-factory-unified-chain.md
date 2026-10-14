# 批量工厂统一链路收口实施计划

> Native execution in this session.

**Goal:** 将批量工厂收口为“智能统一 1 次、资产 1 次、视频 1 次；TTS 时长始终参与时间轴；约束只影响最终 Prompt 展示/拼接”的统一链路。

**Architecture:** 保留现有 Go batchfactoryv11 服务和 React 批量工厂界面，不重做存储模型。通过配置归一化固定新任务的音频时间计划，通过 SD 视频合同把原文、时间计划、资产和导演画面放进单次视频提示词请求，并保留旧任务的历史兼容路径。约束层仍由系统最终编译，不再触发 AI 重跑。

**Spec:** docs/批量工厂完整链路思维导图.md

## Global Constraints

- 后端逻辑只改 Go backend/internal/batchfactoryv11，前端只改 React frontend/src/user/pages/shuihuo。
- 不覆盖工作区现有未提交改动。
- 历史任务保留兼容行为；新任务默认启用完整配音时间计划。
- 不把提示词硬编码到前端。
- 约束开关变化不调用文本 AI、不重新生成导演分镜。
- 分镜真实时长由完整配音和逐行探针决定，视频模型档位单独保存。
- 只提交本次定向文件，不提交构建产物、日志和其他脏文件。

## Review Focus

- 原声模式仍必须生成 TTS 时长计划，只在最终合成阶段保留原声。
- 关闭基础设定、智能统一、画质约束或负面词时，分镜卡即时隐藏对应段，不触发 AI。
- SD 新流程只做一次视频提示词文本请求；原文行、资产和时长契约不能被 AI 改写。
- 旧任务不能因读取默认配置而静默升级。
- 失败的提示词结构必须在视频请求前阻断。

### Task 1：音频来源与时间计划边界

**Files:** backend/internal/batchfactoryv11/director_service.go; frontend/src/user/pages/shuihuo/batchFactoryAudioSource.js; frontend/src/user/pages/shuihuo/BatchFactoryEngineSettingsDrawer.jsx; frontend/src/user/pages/shuihuo/BatchFactoryBookSettingsModal.jsx

- 新任务默认启用音频时间计划；原声只关闭最终 TTS 合入，不关闭完整 TTS 与逐行测量。
- 去除用户需要手动打开“分镜规划跟随配音”的误导入口，保留旧字段兼容读取。
- 增加测试：原声模式仍要求有效时间计划；TTS/原声只改变最终合成选择。

### Task 2：SD 单次视频提示词合同

**Files:** backend/internal/batchfactoryv11/sd_story_script.go; backend/internal/batchfactoryv11/director_service.go; backend/internal/batchfactoryv11/director_contract.go

- 新 SD 流程把原文、时间计划、资产、智能统一和选中视频预设一次性提交给视频提示词 AI。
- 返回结果直接包含故事画面和固定镜头信息，系统绑定原文行、镜头数量、镜头顺序、覆盖范围和时长。
- 新流程不再先调用一次故事剧本再调用一次导演；历史任务仍读取旧快照。
- 增加测试：一次请求、原文契约不变、时长不被 AI 覆盖。

### Task 3：最终分镜卡和约束开关

**Files:** backend/internal/batchfactoryv11/final_prompt.go; frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx

- 分镜卡只展示系统编译后的最终 VIDEO Prompt。
- 基础设定控制统一人物，智能统一控制统一风格，画质约束控制最终导出画质约束，负面提示词控制最终导出负面提示词。
- 开关切换即时更新，不调用生成接口。
- 保留 H3/SD 预设选择差异，但使用同一条阶段链路。

### Task 4：验证与发布

- 运行定向 Go 测试。
- 运行定向前端测试和构建。
- 只暂存本次修改文件，提交到 v88。
- 重启本地 Go embed 服务并验证 :4001/batch-factory。
