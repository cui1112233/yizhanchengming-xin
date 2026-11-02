# 批量工厂统一链路收口实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将批量工厂的预设选择、资产生成、时间计划、导演分镜和最终 VIDEO Prompt 统一为一条由 Go 后端控制的可追踪链路。

**Architecture:** 批次快照保存唯一有效配置；Go 后端按配置选择 SD/H3 资产预设和视频预设。AI 只负责资产与故事/导演结果，时间计划和最终分镜卡由系统编译；前端只读取有效配置和编译结果，不再自行拼接提示词。

**Tech Stack:** Go、MySQL、React、现有 Batch Factory V11/V12 API、Go embed 发布产物。

**Spec:** 本次对话确认的“统一合并完整功能”方案，以及 `SD_AI完整处理流程与新系统接入指令.md` 中的阶段边界。

## Global Constraints

- `v88` 是唯一维护主线，所有改动先进入 Git。
- 系统预设提示词统一由后端 prompt 模块提供，前端不得硬编码。
- TTS/原声只决定最终合成音轨，不改变时间计划、资产或视频提示词流程。
- 约束开关只决定最终分镜卡是否注入对应内容；关闭开关不得触发 AI 重生成。
- SD/H3 选择决定对应预设协议和输出格式，不能由视频模型名称反向覆盖。

## Review Focus

- SD 视频提示词必须输出“场景/原文行/总时长/故事剧本/环境/镜头/时长”结构。
- 资产提示词必须将人物名单、事实、外形、关系作为一次资产预设调用的输入边界。
- 关闭基础设定、画质约束、画面限制或负面提示词后，最终卡只隐藏对应层。
- 重新生成导演分镜才允许重新调用视频提示词；只改开关必须即时编译。
- 每行真实配音时长必须进入时间计划，不能使用固定 3.5 秒替代。

### Task 1: 统一批次配置快照与预设选择

**Files:**
- Modify: `backend/internal/batchfactoryv11/settings.go`
- Modify: `backend/internal/batchfactoryv11/*config*` and related HTTP handlers
- Test: `backend/internal/batchfactoryv11/settings_test.go`

**Interfaces:**
- Produces an effective snapshot containing `asset_preset_id`, `video_preset_id`, `smart_unified`, `base_setup`, `quality`, `picture_limit`, `negative`, and `audio_source`.

- [ ] Write failing tests proving SD/H3 selection is independent from video model name and that audio source only affects final mix mode.
- [ ] Run the focused Go tests and verify the expected failures.
- [ ] Implement one effective snapshot resolver used by config reads, director execution, and final prompt compilation.
- [ ] Run focused tests and the existing batch-factory settings tests.
- [ ] Commit the configuration contract.

### Task 2: Unify asset and director execution boundaries

**Files:**
- Modify: `backend/internal/batchfactoryv11/director_service.go`
- Modify: `backend/internal/batchfactoryv11/visual_prompt_service.go`
- Modify: `backend/internal/batchfactoryv11/sd_story_script.go`
- Test: `backend/internal/batchfactoryv11/director_service_test.go`, `sd_story_script_test.go`

**Interfaces:**
- Asset stage makes one preset request and persists the resulting asset pool.
- Director stage consumes the frozen timeline, assets, selected video preset, and optional smart-unified result.

- [ ] Add failing tests for one asset call, SD/H3 video preset routing, and line-duration preservation.
- [ ] Verify failures.
- [ ] Route both stages through the effective snapshot and reject legacy mixed-format output.
- [ ] Verify focused tests.
- [ ] Commit the execution-boundary changes.

### Task 3: Make the final storyboard compiler the only display source

**Files:**
- Modify: `backend/internal/batchfactoryv11/final_prompt.go`
- Modify: `backend/internal/batchfactoryv11/director_output.go`
- Modify: `backend/internal/httpapi/batch_factory_v11_compiler.go`
- Test: `backend/internal/batchfactoryv11/final_prompt_test.go`, `director_output_test.go`

**Interfaces:**
- `CompileFinalPrompt` returns the final card text, injected layers, references, audio mix mode, and compile trace.

- [ ] Add failing tests for switch-controlled injection and exact SD scene/shot headings.
- [ ] Verify failures.
- [ ] Compile optional layers without AI calls and preserve @ asset references without embedding the @ marker in the visible capsule label.
- [ ] Verify all compiler tests.
- [ ] Commit the final compiler changes.

### Task 4: Remove frontend prompt authority and align UI actions

**Files:**
- Modify: `frontend/src/shared/api/batchFactoryV11.js`
- Modify: `frontend/src/user/pages/batch-factory-v11/BatchFactoryV11UiPage.jsx`
- Modify: `frontend/src/user/pages/batch-factory-v11/BatchFactoryV11Workbench.jsx`
- Test: `frontend/src/user/pages/batch-factory-v11/*.test.js`

**Interfaces:**
- Frontend renders only server `effective-settings` and `final-prompt` results.
- Save config, regenerate director, compile-only switch changes, and edit-cancel are distinct actions.

- [ ] Add failing source tests for action separation and @ asset insertion/deletion as whole chips.
- [ ] Verify failures.
- [ ] Bind UI actions to the corresponding Go endpoints and remove local prompt assembly paths.
- [ ] Verify frontend tests and build.
- [ ] Commit the UI authority cleanup.

### Task 5: Runtime verification and embed build

**Files:**
- Modify: relevant Go embed/build wiring only if verification exposes a gap.
- Test: existing Go and frontend suites plus local browser smoke flow.

- [ ] Run the focused Go and frontend tests.
- [ ] Rebuild the Go local executable with embedded frontend assets.
- [ ] Start local 4001 and verify: select SD → save → regenerate once → toggle constraints without regenerating → final card changes only by selected layers.
- [ ] Record the verified commit SHA and remaining public deployment boundary.
- [ ] Commit the plan implementation and build evidence.
