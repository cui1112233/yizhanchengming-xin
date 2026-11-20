# 全篇短篇 parser、原始 prompts 与无效第二稿回退

2026-11-20，只读静态取证。未运行洛水程序/模块，未调用模型，未修改业务代码，未提交。原件 SHA-256：`b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d`。

复用既有 Ghidra 12.1.4 项目 `-readOnly -noanalysis`，没有重新 import。本目录 C/Listing 属分析员导出，**没有新的 REA Evidence ID**。`*.full.assembly.txt` 来自完整 Ghidra Function.body；较早 `*.assembly.txt` 只是 PE unwind 单区间，不用它声称完整函数。`selected-function-bindings.json` 保留已证实模块/名字/原生地址/构造调用/原字节范围；常量注释由既有模块常量池地址映射生成。`ghidra-*.log` 保留导出结果。

## 可直接实现的 parser 合同

### 标记块

`_extract_tagged_blocks` = `0x18014cdf0`。模式拼接为 `^\s*===\s*` + `re.escape(tag)` + `\s*===\s*$`，flags=IGNORECASE|MULTILINE。逐个 finditer；块从该 match.end() 后开始，lstrip('\n')；遇下一个 `^\s*===\s*[A-Z0-9_ ]+\s*===\s*$` 标记行则截断，否则到全文末尾。next-tag 同样 IGNORECASE|MULTILINE。返回所有块；不是以任意 `===` 子串切开。

证据：`0x18014cdf0.fallback.annotated.c` 常量#3/#5/#8和 finditer/end/start 分支；完整指令 `0x18014cdf0.full.assembly.txt`。

`_extract_last_nonempty_tagged_block` = `0x18014e500`：遍历上述所有块，取 block.strip()，只保留非空，最后返回列表[-1]；没有非空块返回空串。证据 `0x18014e500.fallback.annotated.c:49-214`，最后下标 -1 在 `:149` 附近，完整指令可查同地址文件。

### 标题规范化的精确顺序

`_normalize_title` = `0x1801519b0`：

```python
for line in raw.splitlines():
    cleaned = re.sub(r'^#+\s*', '', line).strip()
    if cleaned:
        return re.sub(r'^《(.+)》$', r'\1', cleaned).strip()
return ''
```

注意取第一条清理后非空行；不会把所有行拼成标题。`0x1801519b0.fallback.annotated.c:180-323`，splitlines/两次 strip 载入见 `0x1801519b0.full.assembly.txt:55,158,347`。

`_normalize_chapter_title` = `0x180152350`：先 normalize_title(raw)，再用 `^第\s*` + str(number) + `\s*章\s*` 替换空串，随后 strip；为空则 `第{number}章`。并非先去章节前缀再选标题行。证据 `0x180152350.fallback.annotated.c:56-203,255-293`。

### 正文清理

`_sanitize_chapter_content` = `0x180151160`：

```python
text = re.sub(r'^```(?:md|markdown)?\s*', '', raw, flags=re.IGNORECASE)
text = re.sub(r'```\s*$', '', text, flags=re.IGNORECASE)
text = re.sub(r'^\s*===\s*[A-Z0-9_ ]+\s*===\s*$', '', text,
              flags=re.IGNORECASE | re.MULTILINE)
return text.strip()
```

第三步 replacement 是空串，不是换行。`\s*` 可能消费换行，不应替换成只匹配空格的自创模式。证据 `0x180151160.fallback.annotated.c:46-91,108-151,171-267`；完整 Listing `0x180151160.full.assembly.txt:91,190,292,337,428` 分别载入前两次IGNORECASE、第三次IGNORECASE/MULTILINE、最终strip。

### 重复 TITLE 兜底

`_extract_duplicate_title_tagged_chapter_content` = `0x180150000`：构造指定 `CHAPTER {number} TITLE` 的完整标记模式（escape，IGNORECASE|MULTILINE）；finditer 后少于2个返回空串。否则取 **第2个 match**，不是最后一个。取其 end() 后文本，lstrip('\n')，以任意下一个 `CHAPTER 数字 TITLE|CONTENT` 或 `SHORT_FICTION_*` 标记为截止；取截止前部分并 strip。

该规则用于正文被误放到重复 TITLE 后的情况，不能泛化成“只要缺正文就任意拿标题段当正文”。证据 `0x180150000.fallback.annotated.c:352-445,452-510,676-730`，比较小于2、subscript=1、end/lstrip/next-tag和最终strip均有原件指令；flags也保存在完整 Listing。

### Markdown 兜底及确认的原版缺陷

`_extract_markdown_chapter_title` = `0x18014f130`：

```python
pattern = r'^##\s*(?:第\s*' + str(n) + r'\s*章\s*)?(.+)$'
m = re.compile(pattern, re.MULTILINE).search(raw)
return m.group(1).strip() if m else ''
```

`_extract_markdown_chapter_content` = `0x18014f860`：

```python
pattern = (r'^##\s*(?:第\s*' + str(n) + r'\s*章\s*)?.*$\n([\s\S]*?)'
           r'(?=^##\s*(?:第\s*' + str(n+1) + r'\s*章\s*)?.*$|$(?![\s\S]))')
m = re.compile(pattern, re.MULTILINE).search(raw)
return m.group(1).strip() if m else ''
```

这里整个“第 n 章”组确实可选，且使用 search 第一匹配，**可以让不同 n 反复匹配同一个首章**。不能改写成“原版精确寻找第n章”。本目录 `derived-markdown-regex-probe.json` 使用分析员构造的两章文字与这些解码模式做标准库 re 推导：n=2仍取第一章正文。此探针不导入/执行目标，仅说明观察到的正则语义。root 若修正该问题需标明适配；若逐字保持，则默认 prompts 应继续使用显式 CHAPTER blocks。

证据 `0x18014f130.fallback.annotated.c:55-126,181-245`；`0x18014f860.fallback.annotated.c:50-136,190-280`；对应完整 Listing 明确 compile/search/MULTILINE。

### 统计、缺章、完整性与原文

- `_count_zh_chars` = `0x1801533f0`，实际 `len(re.sub(r'\s+', '', text))`。名字虽含 zh，但**不是只数汉字**；英文、标点也计数。证据 `0x1801533f0.fallback.annotated.c:41-73`。
- `find_empty_short_fiction_chapters` = `0x180156ff0`，收集 content.strip() 为空的 chapter.number；不是按目标字数差触发缺章。其 C 从上一轮已保存证据复制到本目录。
- `validate_short_fiction_draft_for_final` = `0x180157690`，expected_chapters 非None时检查 len(chapters) 完全一致；再检查没有空章。没有45%字数阈值、逐章字数强校准或内容评分门槛。C保存在同目录，异常文案分别是“期望…实际…章”和“以下章节为空”。
- `parse_short_fiction_batch_draft` 原输入 raw **原样**留在 RawContent，不做全局 strip。原 C 在前轮目录 `../novel-short-recovery-20261120/0x180153e50.fallback.annotated.c:88` 仅赋值 local_170一次，返回结构字段直接使用该变量；标题/章节各自规范化不覆盖它。解析入口按 expected N 逐章构造，即使某章缺失也保留空条目，供补写检查。
- `estimate_short_fiction_max_tokens` = `0x180159910`，公式 `max(12288, math.ceil(chapter_count * chars_per_chapter * 2.2) + 4096)`，证据 `0x180159910.fallback.annotated.c:48-109`。

### render 格式

`render_short_fiction_draft_markdown` = `0x180158570`，parts 初始 `'# '+story_title`；hook为真时追加 `'## 开篇钩子\n\n'+opening_hook`；每章追加 `'## '+heading+'\n\n'+content`，最后以 `'\n\n'` join 非空 parts。正文此处直接使用 content，不重新sanitize或strip。

heading 由 `_format_chapter_heading`（`0x180152b30`）计算：title.strip()为空→`第n章`；若 `re.match('^第\s*'+n+'\s*章', stripped)`→直接返回stripped；否则 `第n章 `+stripped。证据 `0x180152b30.fallback.annotated.c:35-122,130-223,276-299`；render证据 `0x180158570.fallback.annotated.c:53-149,195-227,371-461`。

## 完整原始自然语言 prompts

文件全部由已解码原常量与成功导出的 builder 顺序构造；`{{...}}` 为动态值/循环说明，**不是目标原有模板语法**。所有 builder 最终按换行连接；章节 block helper 返回的四行直接进入列表，章节之间不额外自创空行。

| 内容 | 本目录文件 | 原件证据 |
|---|---|---|
| Writer 基底system | writer-system.txt | short_fiction_prompts常量#136 |
| Writer工艺要求，user内嵌 | writer-craft.txt | 同模块#138；writer与continuation调用_build_short_fiction_craft_prompt |
| 首稿user | writer-user-template.txt | 0x1801d6160.fallback.annotated.c:39-287 |
| 每章输出四行 | chapter-output-block.txt | 0x1801d5660.fallback.annotated.c:105-300 |
| 缺章补写user | continuation-user-template.txt | 0x1801d6860.fallback.annotated.c:80-408,452-630 |
| 正文审阅system | draft-review-system.txt | 常量#144 |
| 正文审阅user | draft-review-user-template.txt | 0x1801d7d70.fallback.annotated.c:31-129 |
| 打包system | package-system.txt | 常量#150 |
| 打包user | package-user-template.txt | 0x1801d8730.fallback.annotated.c:34-202 |

字段修剪差异已保留：writer、continuation、review中的 direction/outline/draft字段原样拼；package仅outline_markdown和draft_markdown strip，direction/draft_title原样。revision followup 的 review.strip() 合同仍见前轮证据。

额外system要求：`_writer_sys_prompt`（0x180161530）先调基底builder，再查 `self._ctx('writer').settings_store.custom_write_instructions('short_fiction')`；存在非空自定义值则追加 `\n\n## 用户自定义写作要求\n` + 自定义值。证据 `0x180161530.fallback.annotated.c:47-111,154-200`。仿写 addendum 后续由代理对首条system注入，与该用户自定义要求是不同来源。

`prompt-template-constants.json` 记录模板使用的完整文本与索引，可核对中文原字；没有前端硬编码或调用目标获取prompt。

## 无效第二稿的异常边界、采用稿件和最终状态

原 `_produce` 单次第二稿 try 区域覆盖：调用修订writer → parse revised_raw → validate完整性 → 写 drafts/v002.md。异常处理匹配 tuple `(ShortFictionDraftValidationError, Exception)`。

**第二项确为 Exception，不是 ValueError**：前轮完整 Listing 的0x1801688f2读取 IAT 0x180390790，再取该异常对象；原件PE导入表将此IAT绑定 `python312.dll!PyExc_Exception`，本轮独立记录在 `fallback-exception-import.json`。证据 `../novel-short-recovery-20261120/0x180165140.full.assembly.txt:3530-3612`。

因此原版范围很宽：修订网络调用、解析、校验甚至写第二稿的普通 Exception 均可能落入回退。它不等于“只有空章/无效格式才回退”。BaseException是否自定义/继承、AI客户端取消如何继承属于本轮未扩大结论的范围。

匹配异常后记录异常字符串、将采用稿件设回完整 v001，写 `reviews/draft-v002-warning.md`；文案包括“第二轮改稿未采用”“系统没有用不完整或解析失败的改稿覆盖完整首稿。”，日志 `Draft revision skipped, keeping v001: %s`。若警告文件写入或处理异常本身再次失败，会转向失败出口，不能声称始终成功。

- 正常第二稿校验成功：r13为parsed v002，0x180168e49将rsi设为r13。
- 捕获回退路径：0x180168dbf从[rbp-0x58]恢复v001，0x180168dc3将rsi设为该v001。
- 两路径在0x180168e7f后共同进入最终渲染/打包，使用同一采用稿件变量。
- 后续仍需打包成功；最终写status='completed'，有warning时加revision_warning，未见独立“带警告完成”状态枚举。

证据：前轮完整 Listing `:3613-3964`（异常处理/稿件合流）、`:3999-4100`（最终正文渲染与写入）、`:5284-5334`（completed/revision_warning）、`:5438`（status.json）。

**CM安全适配要求**：仅已确定的无效第二稿可以采用原首稿并保留警告；取消、可能已付费但结果未知、数据库/持久化失败等必须保留CM既有停止与恢复保护。不能把原版宽泛Exception捕获直接套到CM后冒称“一比一且安全”。本轮只取证，不修改业务实现。

## 明确未做的事

未执行目标/真实请求/UI；未重新验证外部provider运行时；未扩大其他模块。Markdown原版宽松正则缺陷已明确，不把意图当事实。已有 Ghidra database 只读打开；无新REA记录、无业务修改、无提交。

## 后续聚焦复核：温度、package、first block与补写预算更正

### 温度与预算

| 原版whole请求 | temperature | max_tokens | 直接原件证据 |
|---|---:|---:|---|
| 正文审阅 | 0.3 | 8192 | `../novel-short-recovery-20261120/0x180165140.full.assembly.txt:2673`，kwargs dict常量#153 |
| package打包 | 0.45 | 4096 | 同文件`:4378`，kwargs dict常量#179 |

**更正前轮补写预算误判**：首稿/第二稿修订按完整chapter_count估算；缺章completion按 `estimate_short_fiction_max_tokens(len(missing), chars_per_chapter)` 估算。它的用户提示仍写完整chapter_count，不应把提示里的总章数当作预算入参。

直接调用证据：同一完整 Listing `:1478-1500`（0x18016682f向helper传当前missing，[rbp-0x70]；0x18016685c载入原chars_per，[rbp]；0x18016686e调用estimate）。`0x180375370.fallback.c:10-17` 证明该helper调用底层长度函数FUN_180373860后用PyLong_FromSsize_t包装为Python整数；同一底层长度函数在已恢复_count_zh_chars用于字符串长度。missing变量在补写重解析后由find_empty返回并写回[rbp-0x70]，见前轮 Listing `:1740-1760`。

旧目录 `novel-short-recovery-20261120/FINDINGS.md` 中误写fullcount的句子已同步更正，保留本次更正说明；没有修改业务代码。

### First block语义：标题/hook与正文不同

本轮补导出 `_extract_tagged_block` = `0x18014e1c0`。它调用_extract_tagged_blocks后直接返回`blocks[0]`，没有跳过空首块。没有块则空串。证据 `0x18014e1c0.fallback.annotated.c:49-83`。每块在_extract_tagged_blocks内部加入返回列表之前已经strip，见 `0x18014cdf0.fallback.annotated.c:682-734`。

原parseShortDraft的SHORT_FICTION_TITLE、SHORT_FICTION_OPENING_HOOK/OPENING_HOOK、CHAPTER n TITLE均用first helper；只有CHAPTER n CONTENT使用last_nonempty helper。此区别会影响重复TITLE修复：第二个TITLE块可能含正文，不能让它覆盖第一章标题。

独立审查初版 `short_full_parser.go` 时发现上述first/last混用，已向root报告；root随后增加shortFullBlocks/First并切换相关字段。本代理未编辑业务源码。

### package parser完整标签合同

`parse_short_fiction_sales_package` = `0x180155eb0`，完整C导出成功，文件为 `0x180155eb0.fallback.annotated.c`。

```python
title = normalize_title(first('SHORT_FICTION_PACKAGE_TITLE')
                        or first('SHORT_FICTION_TITLE') or fallback_title)
intro = first('SHORT_FICTION_INTRO') or first('INTRO')
selling_raw = first('SHORT_FICTION_SELLING_POINTS') or first('SELLING_POINTS')
cover = first('SHORT_FICTION_COVER_PROMPT') or first('COVER_PROMPT')
selling_points = [re.sub(r'^\s*[-*]\s*','',line).strip()
                  for line in selling_raw.split('\n')
                  if re.sub(r'^\s*[-*]\s*','',line).strip()]
```

- `first`即上述strip后的首块，不是最后非空块。
- title fallback参数默认“未命名短篇”；调用者可传draft title。默认值直接来自既有parser常量#164，参数元数据#165。
- bullet只去**一次**行首 `-` 或 `*` 加附近空白；不剥数字序号、不剥 `•`；`--text`处理后仍保留`-text`。
- 空selling列表合法；原parser本身没有强制至少几条、非空intro/cover的验证门槛。
- 返回结构含title/intro/selling_points/cover_prompt/raw_content，raw_content为原raw。

行号：title优先顺序 `0x180155eb0.fallback.annotated.c:86-184`；intro aliases`:189-224`；selling aliases`:233-278`；cover aliases`:349-416`；返回字段`:439`；bullet与过滤`:514-590`。

### 最后独立审查记录

当前firsthelper修改后，已检查的标题/hook/正文规则没有其他明显阻断偏差。另向root报告package可选fallback未传时默认空标题与原默认“未命名短篇”的差异；root可统一默认，或在所有生产调用明确传draft.Title并记录为API适配。本代理不修改源码、不提交、不运行目标或模型。
