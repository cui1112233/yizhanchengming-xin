# AI-CUT AI 自动剪辑验收记录（2026-11-11）

## 本次范围

- 在专业剪辑器顶栏加入“AI 自动剪辑”入口。
- 素材、文本模型、剪辑目标、目标时长、画面比例和字幕/转场/滤镜/BGM 开关组成输入。
- 模型只生成可审核的剪辑草案，不直接修改时间线。
- 用户可以逐项启用/停用操作，再明确点击“应用到时间线”。
- 应用支持版本冲突提示、重复请求幂等和一次撤销；取消草案不修改时间线。
- 系统提示词只在后端 `internal/prompt` 中维护，前端没有硬编码系统提示词。

## 同组输入与边界结果

| 场景 | 输入 | 预期/结果 |
| --- | --- | --- |
| 正常草案 | 1 个视频素材、有效文本模型、默认目标、60 秒、9:16 | 浏览器中素材选择后“生成剪辑草案”可执行；后端路由测试验证返回 ready 草案，应用后时间线版本递增 |
| 空素材 | 不选择素材 | 浏览器中按钮禁用；接口返回“请至少选择一个素材” |
| 空目标 | 目标只含空白 | 前端 API 在发送网络请求前返回“剪辑目标不能为空” |
| 重复请求 | 相同 clientRequestId 连续提交 | 返回同一草案，不重复创建、不重复调用模型 |
| 错误操作 | 模型不存在、模型 JSON 非法、操作不在白名单 | 分别返回模型未配置、草案格式无效或草案内容无效，不修改时间线 |
| 时间线已改变 | 草案 baseRevision 与当前 revision 不同 | 草案变为 stale，提示“时间线已更新，草案已过期，请重新生成” |
| 取消 | 关闭抽屉或取消 ready 草案 | 浏览器确认抽屉关闭；后端确认草案 canceled，时间线不变 |
| 重复应用 | 同一草案重复点击应用 | 返回已应用结果，不重复写入时间线 |
| 撤销 | 应用草案后执行一次撤销 | 恢复应用前的时间线版本 |

## 真实运行验收

- 地址：`http://127.0.0.1:4001/ai-cut`
- 运行进程：`F:\脚本测试\一战晟铭-v88\backend\qiantie-local.exe`
- 浏览器实际看到顶栏“AI 自动剪辑”，点击后出现跟随深色主题的右侧草案抽屉。
- 浏览器实际验证：空素材时按钮禁用；选择视频素材后按钮启用；关闭抽屉后不可见。
- 没有执行付费模型生成；真实模型调用留给用户选择已配置模型后主动点击。
- 真实运行时曾发现 MySQL JSON 参数被当作二进制，日志报 `Error 3144`；已改为 UTF-8 JSON 字符串入库，并加入参数类型回归测试。

## 自动化证据

```text
go test ./internal/aicut ./internal/prompt ./internal/httpapi ./internal/storage
ok qiantie/backend/internal/aicut
ok qiantie/backend/internal/prompt
ok qiantie/backend/internal/httpapi
ok qiantie/backend/internal/storage

node --test src/shared/api/aicutEditPlan.test.js src/user/components/AICutAutoEditDrawer.source.test.js src/user/pages/aicutTimeline.test.js src/user/pages/aicutWorkbench.test.js src/user/pages/aicut-controls.contract.test.js src/user/pages/aicut-workbench.layout.test.js
63 tests passed

npm run build
vite build passed

powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1
Go static assets prepared
```

## 独立实现边界

本功能依据公开交互和本项目的数据结构独立实现，没有复制洛水原代码、密钥、私有接口或专有资源。模型调用走一战晟铭现有的用户模型配置，草案操作只允许后端白名单中的剪辑能力。
