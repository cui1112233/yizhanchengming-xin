---
name: ai-gateway-skill
description: 通过 YFAI (https://yf.token6688.com) 接入 AI 模型能力时使用: 给用户的代码 / 网站 / APP 加对话、生图、生视频、语音、音乐等功能, 或直接替用户调用模型生成内容。GT 兼容 (换 base_url 即用) + CL Messages / GT Responses 原生入口 + 图片 / 视频 / 音频统一异步入口。每次新对话先按「启动步骤」拉本站实时真值再动手。
---

# YFAI · AI 模型接入 Skill

> 这个文件只告诉你**去哪里拿真值、按什么规矩调**; 具体接口 / 模型 / 价格以本站实时接口返回为准, 不要背下来。

## 1. 平台与密钥

- 平台: YFAI · Base URL: `https://yf.token6688.com` · 网页文档: https://yf.token6688.com/api-docs
- 鉴权: `Authorization: Bearer <API_KEY>` (CL 入口也接受 `x-api-key`)。
- Key 从哪来 (按顺序): ① 项目根 `.env` 的 `API_KEY` ② 本文件末尾「密钥」一节记录的位置 ③ 都没有 → 让用户到网页右上角「管理密钥」创建, 存进 `.env`。
- **没有 Key 不调任何付费接口。** 公开接口 (`/v1/skills` `/v1/skills/guide`) 不需要 Key; 模型目录、模型详情与价格 (`/v1/skills/models`、`/v1/skills/models/{model}`、`/v1/skills/models/{model}/pricing`、`/v1/skills/pricing`) 要带 Key。

## 2. 启动步骤 (每次新对话都做, 顺序不变)

1. `GET https://yf.token6688.com/v1/skills` — 平台接口路由表 (分类 / path / method / 参数)。
2. `GET https://yf.token6688.com/v1/skills/guide` (加 `Accept-Language: zh-CN` 头; 英文发 en) — 每种调用模式的端点、请求示例、轮询流程。
3. 按需 `GET https://yf.token6688.com/v1/skills/models?type=chat|image|video|audio`, 再 `GET https://yf.token6688.com/v1/skills/models/{model}` 看参数 (合法值 = `params[].options[].value`), `GET https://yf.token6688.com/v1/skills/models/{model}/pricing` 看单个模型的跨渠道价格; 要全部模型的价格用 `GET https://yf.token6688.com/v1/skills/pricing` 一次取回。
4. 把 1~3 的要点整理进本文件同目录的 `API_MAP.md` (接口清单 / 模型清单 / 每类调用示例 / 计费口径 / Key 位置)。已存在且一致就跳过, 有差异就覆盖。
5. 然后才开始处理用户需求。

## 3. 该走哪个入口

| 需求 | 入口 | 备注 |
|---|---|---|
| 对话 / 写作 / 代码 (GT · CL · GM · GK 系列与国产模型 …) | `POST https://yf.token6688.com/v1/chat/completions` | GT 格式, 换 `model` 即换厂商; 默认带 `stream:true` (SSE): 非流式在生成完成前不返回任何内容, 推理模型 / 长输出容易先撞上客户端的等待上限; 需要一次性 JSON 时去掉 stream |
| 已用 CL SDK / CL Code | `POST https://yf.token6688.com/v1/messages` | 请求体保持 CL 原样 |
| 已用 GT Responses SDK / Codex | `POST https://yf.token6688.com/v1/responses` | 原生 Responses |
| 视频 / 音乐 / 语音 | `POST https://yf.token6688.com/v1/media/generate` → `GET https://yf.token6688.com/v1/tasks/{task_id}` | 请求体 `{"model","prompt","params":{…}}`, 轮询到 `is_final=true`, 结果在 `output_url`; 图片走这里是同步的 |
| 视频 (扁平参数写法) | `POST https://yf.token6688.com/v1/videos/generations` | 参数放**顶层**, 不用 params 信封; 支持 `client_request_id` 幂等 |
| 同步出图 | `POST https://yf.token6688.com/v1/images/generations` | 阻塞 40~50 秒; 链路超时 <60s 请改用 `POST https://yf.token6688.com/api/v1/model-runtime/invoke` 异步。返回 `data[].url` (可能是压缩图; 无损原图在条件字段 `original_url`; 极少数转存失败时改为 `b64_json`) + `usage.cost_rmb` / `usage.balance_rmb`; 收到 200 也要检查有无 `error` |
| 本地文件 → 公网 URL | `POST https://yf.token6688.com/v1/files` (multipart 字段 file, ≤50MB) | 返回 `url` 直接当参考素材 |
| 余额 | `GET https://yf.token6688.com/v1/skills/balance` | 付费调用前先查; `balance` / `available_balance` / `frozen` 是美元 "$x.xxxxxx" 字符串 (解析时去掉 $), 同额的算力 (⚡) 口径在 `credits` 里, 跟网页显示一致 |
| 消费明细 | `GET https://yf.token6688.com/v1/skills/usage` | 按次核算成本; 每条含 `model` / `channel` (渠道代号) / `cost` (⚡ 与美元双口径) / `tokens`。**金额带符号** (负=扣费 正=退款), 求和即净消费; 翻页带 `next_cursor` |

调度: 建密钥时可选策略 (智能 / 价格优先 / 稳定优先 / 速度优先) 或按模型固定渠道。单次请求覆盖: 图片 / 视频 / 音频 / `/v1/responses` 用请求头 `X-Schedule-Strategy`, `/v1/chat/completions` 用请求体字段 `strategy`; 取值只认 `balanced` (智能调度) / `cost_first` (价格优先) / `reliability_first` (稳定优先) / `latency_first` (速度优先) / `quality_first` (质量优先), 其它值按智能处理。固定渠道是严格路由, 不会切到未锁定的渠道。

- 同一个模型可能有多个上游渠道。
- 默认「智能调度」自动选最优, 并在失败时自动切换。
- 可以在「API 密钥」页为密钥指定调度策略, 或加请求头 `X-Schedule-Strategy` 覆盖单次请求 (取值见下方策略表)。

**钉渠道用哪个字段**

| 字段 | 用在哪 |
|---|---|
| `lane_id` | 下单时钉一条 |
| `lane_ids` | 下单时钉多条 |
| `pinned_lane_id` / `pinned_lane_ids` | 估价时对应的字段 |
| `X-Schedule-Strategy` | 请求头, 覆盖单次请求的调度策略 |
| 放在哪 | `/v1/media/generate` 放在 `params` 里, 其它入口放在请求体顶层 |

- 渠道号用估价响应的 `channels[].lane` 或 `/lanes` 列出的渠道号; 价格超过报价时失败响应给出的渠道号也可以 (见「计费与估价」)。
- 这些字段只用于选渠道, 不会转发给厂商。
- `channel_id` 是单向标识, 不能用来钉渠道。
- 钉多条只保证「只在这些渠道里选」, 不保证「只用第一条」: 按价格排序依次尝试, 前一条失败会切下一条。
- 不同渠道单价可能差好几倍 —— 要价格确定就只钉一条。
- 钉住的渠道是严格路由: 它限频、故障或审核拒绝时错误原样返回, 不会换到其它渠道。
- 需要自动容灾就用调度策略, 不要钉渠道。

**调度策略**

| 策略 | 说明 | X-Schedule-Strategy |
|---|---|---|
| 智能调度 | 综合价格、成功率、延迟自动选最优 (默认) | `balanced` |
| 价格优先 | 优先选择最便宜的渠道 | `cost_first` |
| 稳定优先 | 优先成功率高、最稳定的渠道 | `reliability_first` |
| 速度优先 | 只按响应速度排序、不设价格上限 (chat 比等首字, 生图/视频/音频比生成总时长) | `latency_first` |
| 质量优先 | 对话 / 图片优先满血渠道, 视频优先官方签发、码率高的渠道; 价格最多为最便宜渠道的 10 倍 | `quality_first` |

## 4. 异步任务怎么判

- 视频、音乐、语音, 以及走 `/api/v1/model-runtime/invoke` 的图片, 都是异步任务。
- 提交后立即拿到任务号 `task_id`, 再用 `GET /v1/tasks/{task_id}` 查询, 直到 `is_final` 为 true。
- 判断结束用 `is_final` (或 `status` 为 `completed` / `failed`), 不要比对其它字段的文字。
- 建议每 3 秒查一次, 查询总时长设为 60 分钟以上。
- 不要用 2~3 分钟的超时判失败: 停止查询不会让任务失败, 但你会拿不到已经生成的结果。
- `status` 的全部取值见「任务查询」页。

**每种状态怎么处理**

| 状态 | 含义 | 你该怎么做 |
|---|---|---|
| `pending` / `processing` | 长时间停留多半是上游排队, 任务不会丢 | 继续查; 不要重新提交同一个任务 (会重复扣费) |
| `processing` + `stage=downloading` | 结果还在转存到平台 CDN | 继续查, 不要把中间态当结果 |
| `completed` | `output_url` 就是结果地址 (`url` / `result_url` / `data[0].url` 是同一个值), 一定是平台 CDN 上的 https 绝对地址 | 取走结果 |
| `failed` | `error` 是失败原因, `error_class` 是机器可读的分类 | 看原因; 失败不收费, 提交时冻结的金额自动退回, 不用申请退款 |
| 410 `task_expired` | 任务记录过了保留期 | 任务结束后请及时取走结果 |

**真实耗时 (2026-08 线上 48 小时实测, 提交到结束)**

| 类型 | 中位 | p90 |
|---|---|---|
| 图片 | 1~3 分钟 | 高峰可到 60 分钟 |
| 视频 | 4~40 分钟 | 55~75 分钟 |

- 网络超时想安全重试, 用「防重复提交」。

## 5. 参数铁律 (每一条都有真实客户踩过)

- 枚举参数发选项的 `value`, 不是界面上的显示名。
- 合法值以 `param_schema` 里每个选项的 `value` 为准, 它可能是中文 (例如某档位 value 是「标准」, 界面显示「Mini」)。
- 网关认得出的写法 (显示名、大小写、常见同义词) 会换成合法值。
- 认不出的按没传处理、落到默认值 —— 表现为「参数没生效」, 且照常计费。
- 两种情况都会写进响应的 `param_adjustments`, 按它核对最省事。

- 参考图、参考视频、参考音频要用公网能直接下载的 http(s) 地址。
- 不支持相对路径、需要登录、防盗链或会过期的签名链接。
- 只有本地文件时先用 `POST /v1/files` 上传: 返回体里的 `url` 可以直接放进参考素材参数 (`images` / `videos` / `audios`) 或对话的 `attachments`, 它不是临时链接。
- 上传接口自己的 50 MB 上限只约束上传这一步。
- 每个模型能收几个素材、单个文件多大, 写在它的模型页参数表; 机器可读的是 `GET /v1/logical-models` 里该模型 `param_schema` 的 `max_items` / `max_bytes`。
- 不同模型差别很大, 别照搬别的模型的上限。
- 视频必须写提示词: 只传参考素材、不写 `prompt` 的请求会被拒绝。

- 视频与音频有三种请求体, 提交返回也各不相同 —— 别用同一段解析代码处理所有入口。
- 照抄另一个入口的示例, 参数会静默落到默认值, 而且照常计费。

| 入口 | 请求体 | 提交返回 |
|---|---|---|
| `/v1/media/generate` | `{model, prompt, params: {…}}` 信封 | `{code, msg, data: {task_id, status, is_final}}` 信封 (data 里可能还有兼容用的中文键, 忽略即可) |
| `/v1/videos/generations` | 参数直接放在顶层, 不收 `params` 信封 | 顶层 `{id, task_id, status}` |
| `/v1/audio/generations` | 参数直接放在顶层, 不收 `params` 信封 | `model_runtime.invoke` 信封, `task_id` 在顶层 |
| `/api/v1/model-runtime/invoke` | `{model_id, modality, operation, params}`, `prompt` 写在 `params` 里 | `model_runtime.invoke` 信封, `task_id` 在顶层 |
| 查询 `GET /v1/tasks/{task_id}` | — | 字段放在根上 (`id` / `status` / `is_final` / `output_url`) |

**提交前估价**

- 用 `POST /v1/pricing-estimate` 估价, 参数必须和提交时完全相同 (包括 `videos` / `images` / `audios` 和 `video_total_duration_sec`)。
- 带参考视频却不传各段时长之和, 估出来的价就不含参考视频。
- 例如真实客诉: 估价 ⚡4.5, 实扣 ⚡7.3 (2 段共 21 秒参考视频)。
- 实扣以服务端探测到的真实时长为准。

**估价与扣费字段**

| 字段 | 是什么 |
|---|---|
| `effective_total_rmb` | 估价响应里: 会走的那条渠道的实付价 |
| `max_effective_total_rmb` | 估价响应里: 它故障、换到备用渠道时的最高实付 |
| `price_basis` | 估价响应里: 价格构成拆分 |
| `required_freeze_rmb` | 提交异步任务时按整单上限预冻结的金额 |
| `actual_cost_micro_usd` | `GET /v1/tasks/{task_id}` 里的实扣 |

**冻结与结算**

- 提交异步任务时按整单上限预冻结 (`required_freeze_rmb`), 完成后按实际用量结算并退回差额。
- 可用余额不够整单冻结额时提交返回 402; 请保持至少一单完整费用的余额。

**API 内联素材传输费**

- API 调用在请求体里内联 base64 素材 (data URI / `input_audio` / 文件内联) 时, 按实际转发给模型厂商的素材字节另计传输费, 并在同一笔扣费里。
- 基础价 **⚡0.002/MB** (1MB = 1,048,576 字节), 实付以账单为准; 例: 一张 300KB 的截图约 ⚡0.0006。
- 素材经平台压缩后再转发的, 按压缩后的字节算; 失败不收, 网关自己的重试不重复收。
- 传公网链接, 或本地文件先 `POST /v1/files` 上传再传 url, 不收传输费; 网页端调用不收。
- 使用日志的扣费行会显示「含传输费 ⚡x (内联素材 y MB)」。
- 先用最便宜的模型 / 最短时长把链路跑通并打印 `usage` / 费用, 再换满血模型上量。

## 6. 交付前自检

- [ ] 密钥只在服务端 / .env, 没进前端或 git
- [ ] 用真实接口跑过一次最小示例, 记录了返回体和费用
- [ ] 异步任务按 `is_final` 收口, 有失败分支和幂等重试
- [ ] `API_MAP.md` 与本次拉到的接口一致

## 7. 长期规则

- 不假设任何接口存在, 以 `/v1/skills` 为准; 再拉发现有新增 / 变动就更新 `API_MAP.md`。
- 定期同步价格用 `GET /v1/skills/pricing` (一次返回全部模型的实付价), **不要**逐个模型调 `/v1/skills/models/{model}/pricing`。
- 价格变化慢: 客户端至少缓存 5 分钟 (响应头 `Cache-Control: private, max-age=300`), 再请求时把上次的 `ETag` 放进 `If-None-Match`, 没变化返回 304。价格因账号而异, 只缓存在自己这边, 不能放共享缓存 / CDN。
- 目录类接口 (`/v1/skills/models*`、`/v1/skills/pricing`) 每个 IP 限 20 次/秒, 不要并发批量拉。
- 401 → 提示用户到「管理密钥」检查或重建; 402 → 提示充值; 429 → 按 `Retry-After` 退避。

## 密钥

(由 AI 在完成第 2 步后填写: 例 “Key 在 .env 的 API_KEY”)

## 使用须知与免责声明

本 API 仅面向 YFAI 已注册用户开放, 用于在您自建的智能体 / 应用中调用本平台已接入的模型能力, 且仅限本人或本组织内部自用。禁止将 API 密钥或本平台接口转售、二次分发、封装后向不特定第三方提供服务, 或用于其他未经许可的外部接入; 禁止生成或传播违反法律法规的内容。经您的密钥发起的全部请求内容及生成结果由您自行负责; 平台对模型输出的准确性、完整性不作保证, 并有权在发现违规或异常调用时限流、停用密钥或终止服务。继续使用即视为已阅读并同意上述条款。
