# C2/C3 当前原版纯本地规则：静态证据与可用合同

日期：2026-11-17。只读检查原安装包；本次只新增审计和恢复产物，没有启动原应用、导入 `.pyd`、执行原 helper、调用 AI 或修改业务代码。

## 已闭合的部分

**文风统计指纹 `Tln` 的完整公式、六条修辞规则和 `qln` 八维定义现已恢复。** `qln` 是八项语义分析说明，并不是八个由统计函数直接计算的分数。不要把两个结构合成自创的“八维打分器”。后写校验、AI 味检测和 45% 长度门槛的原始算法仍然未知。

### 证据身份

- 当前安装前端归档：`_local-archive-20261028/luoshui-novel-audit-20261117/installed-original/index-C1aA08SG.js`。
- SHA256：`5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a`。
- 解析方法：Babel 只解析文字，抽取 AST；手写 custom-base64 + RC4 字符串解码器，只接受字面量字符串表、整数、加减法、固定解码包装器。没有 `eval`、`Function`、原包 import、原函数调用或原应用运行。
- 原字符串表 `Rln` 位于 UTF16 `[8848582,8857072)`；原解码器 `zln` 从 UTF16 `8862678` 开始；原旋转校验逻辑位于 UTF16 约 `8830200` 开始。字面表轮转 **450** 次后，原校验算式得到 **107024**。
- 恢复脚本：`_local-archive-20261028/luoshui-novel-audit-20261117/decode-style-static.cjs`。
- 可读完整证据：`.../recovered/frontend-current/style-helpers.decoded.js`。
- 全部片段 UTF16/UTF8 半开范围：`.../recovered/frontend-current/style-helpers-static-evidence.json`。
- `C2-SF-001` 至 `C2-SF-004` 是本地报告标签，不是 REA 服务器签发的 evidence id。原文件哈希和具体范围才是内容身份。

| 本地标签 | 恢复对象 | 原 bundle UTF16 半开范围 | 原 bundle UTF8 半开范围 |
|---|---|---|---|
| C2-SF-001 | `Tln(e,t)` 整个函数 | `[8858592,8862678)` | `[9118387,9122509)` |
| C2-SF-002 | `wln` 六个正则对象及数组 | `[8845914,8846570)` | `[9105441,9106265)` |
| C2-SF-003 | `qln` 数组 | `[8858548,8858592)` | `[9118343,9118387)` |
| C2-SF-004 | 八个维度对象赋值 | 见 JSON 中分段范围；中间穿插字符串表 | 同左 |

### Tln 原样行为合同

输入 `e` 为文本，`t` 为可选来源名称。前端调用点是 `NovelImportPage.readable.js:367` 和 `:475`。

1. 句子：`e.split(/[。！？\n]/).map(x => x.trim()).filter(x => x.length > 0)`。英文 `.?!` 不作句末切分；中文省略号、分号也不作切分。
2. 段落：`e.split(/\n\s*\n/).map(x => x.trim()).filter(x => x.length > 0)`。单个换行切句，但不一定切段。
3. 每句或每段长度：移除 `/\s+/g` 后取 JavaScript `.length`，即 **UTF16 码元数**。句子中被 split 去掉的句末标点不算在句长内；段长内仍保留标点。
4. `avgSentenceLength`：所有句长之和 / 句数，空数组为 0；`Math.round(avg * 10) / 10`。
5. `sentenceLengthStdDev`：句数 > 1 时取 `sqrt(sum((length - avg)^2) / 句数)`，否则 0；`Math.round(stddev * 10) / 10`。这是总体标准差，分母不是 `n-1`。
6. `avgParagraphLength`：段长均值，空数组为 0；`Math.round(avg)`。
7. `paragraphLengthRange`：`{min: Math.min(...段长), max: Math.max(...段长)}`，空数组为 `{min:0,max:0}`。
8. `vocabularyDiversity`：先 `e.replace(/[\s\n\r，。！？、：；""''（）【】《》\d]/g, "")` 得到 G。G 非空时 `new Set(G).size / G.length`，否则 0；最后 `Math.round(value * 1000) / 1000`。**分子是 JavaScript 字符串迭代得到的 Unicode 码点去重数，分母仍是 UTF16 长度**。不是分词词汇多样性；不要自行扩大标点过滤集，原正则保留英文标点、弯引号等。
9. `topPatterns`：对前述已 trim 的句子，若原句 `.length >= 2`，取 `.slice(0,2)` 作为句首，累计次数；`Object.entries(counts).sort((a,b)=>b[1]-a[1]).slice(0,5).filter(([,n])=>n>=3).map(([prefix,n])=>prefix+"...("+n+"次)")`。注意先取前 5 再过滤；前两个字符按 UTF16 切；排序保持 JS 稳定排序及 `Object.entries` 的键顺序语义。
10. `rhetoricalFeatures`：依次使用下表六个 `/g` 正则对完整文本执行 `e.match(regex)`。匹配数 **>= 2** 时加入 `name + "(" + count + "处)"`。保留原顺序和非重叠匹配规则。
11. `sourceName`：`t || undefined`，不是无条件空串。

整个返回对象只有：`avgSentenceLength`, `sentenceLengthStdDev`, `avgParagraphLength`, `paragraphLengthRange`, `vocabularyDiversity`, `topPatterns`, `rhetoricalFeatures`, `sourceName`。

原函数带有两个混淆死分支：一个比较固定同值字符串 `SSoWv`；另一个比较同一属性 `KpIzC`。在恢复的常量对象里二者均不会进入含 `_0x...` 的分支。可读证据保留这些原代码，报告未将死分支误当成所需外部依赖。

### 六条修辞规则，逐字来自原包

| 顺序 | name | regex |
|---|---|---|
| 1 | 比喻(像/如/仿佛) | `/[像如仿佛似](?:是\|同\|一般\|一样)/g` |
| 2 | 排比 | `/[，。；]([^，。；]{2,6})[，。；]\1/g` |
| 3 | 反问 | `/难道\|怎么可能\|岂不是\|何尝不/g` |
| 4 | 夸张 | `/天崩地裂\|惊天动地\|翻天覆地\|震耳欲聋/g` |
| 5 | 拟人 | `/[风雨雪月花树草石](?:在\|像\|仿佛).*?(?:笑\|哭\|叹\|呻\|吟\|怒\|舞)/g` |
| 6 | 短句节奏 | `/[。！？][^。！？]{1,8}[。！？]/g` |

表格为 Markdown 显示对 `|` 作了转义。实现时以 `style-helpers.decoded.js` 中的原始正则文字为准。例如比喻原正则的首组就是一个字符类，不是把“仿佛”当两字整体匹配；短句节奏会消耗两侧句末符，不能擅自改成重叠统计。

### qln 原八维定义和顺序

| key | label | desc |
|---|---|---|
| narrativeVoice | 叙事声音与语气 | 冷峻/热烈/讽刺/温情/...，附1-2个原文例句 |
| dialogueStyle | 对话风格 | 角色说话的共性特征：句子长短、口头禅倾向、方言痕迹、对话节奏 |
| sceneDescription | 场景描写特征 | 五感偏好、意象选择、描写密度、环境与情绪的关联方式 |
| transitions | 转折与衔接手法 | 场景如何切换、时间跳跃的处理方式、段落间的过渡特征 |
| pacing | 节奏特征 | 长短句分布、段落长度偏好、高潮/舒缓的交替方式 |
| diction | 词汇偏好 | 高频特色用词、比喻/修辞倾向、口语化程度 |
| emotionalExpression | 情绪表达方式 | 直白抒情 vs 动作外化、内心独白的频率和风格 |
| distinctiveHabits | 独特习惯 | 任何值得模仿的个人写作习惯 |

这些是原 UI 的字段说明，不是原私有长提示词。`Tln` 本体只用本地字符串处理、数学运算、Set 和正则，没有 HTTP、模型或桥接调用。这个结论仅限已完整恢复的前端 helper。

### 可以直接由公式构造的独立验证案例

以下预期由数学和代码文字推导，没有运行原函数取得 golden：

- 空串：平均句长/标准差/平均段长/多样性均 0，范围 `{min:0,max:0}`，两个特征数组为空，空来源输出 undefined。
- `甲乙。甲乙。甲乙。`：平均句长 2，标准差 0；只有一段且段长 9，范围 `{min:9,max:9}`；多样性 `2/6 -> 0.333`；topPatterns `['甲乙...(3次)']`；短句节奏仅非重叠命中 1 次，不应输出该项。
- `甲。乙乙乙！`：句长 `[1,3]`，均值 2，总体标准差 1；段长 6；多样性 `2/4 -> 0.5`。
- `😀。😀。`：句长 `[2,2]`，平均句长 2；段长 6；G 为两个 emoji，分子 1、UTF16 分母 4，多样性 0.25。不能以 rune 长度直接替换所有 JS 长度。
- `甲。\n\n乙乙。`：句长 `[1,2]`，均值 1.5，标准差 0.5；段长 `[2,3]`，平均段长 `Math.round(2.5)=3`，范围 `{min:2,max:3}`。

## 尚未闭合的部分

- `post_write_validator` 的真正函数体、检查项、清洗/修复顺序、阈值和返回结构。
- `ai_tells_detector` 的真正检测词库、算法、评分、改写和返回结构。
- 长度偏差 >45% 的精确计数字符规则、分母、比较符、目标范围及触发后的动作。

当前 `.pyd` SHA256 `0471b2a77ee1aad06f4dd30e58b2515ed89b3a20d780ad5ae0f64ab3e4552853` 已由 REA 静态 open 核对。`post_write_validator` 字符串 `0x180391670` 及其引用 `0x18028de9f`、`ai_tells_detector` 字符串 `0x1803914d8` 仅沿用已保存索引作为后续入口，**不属于函数体证据**。按本轮优先完成可闭合前端合同的指示，没有继续耗时原生反编译。不得以本报告给这些原生算法填入自创实现。

## 交付边界

本次交付是当前包的静态原版规则证据，不声称原应用运行验收或全部小说 Agent 一致。八维和统计 helper 的定义缺口已闭合；原生后写、AI 味和 45% 门槛仍须保留 unknown。后续实现应按精确公式另写实现、用独立构造案例验证，不能执行原包 helper 当作测试依赖。
