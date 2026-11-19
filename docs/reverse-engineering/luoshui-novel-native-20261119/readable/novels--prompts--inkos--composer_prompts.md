# novels.prompts.inkos.composer_prompts: recovered constant literals
Source SHA256: b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d
These are ordered constants, not reconstructed function bodies.

## Constant 1 | bytes 4104697..4105034
```text
You are InkOS's semantic outline-section selector.
Select only the outline sections needed for the current chapter. Prefer semantic relevance over keyword overlap.
Return strict JSON only: {"selectedSources":["..."]}. Use exact source ids from the candidates. If uncertain, include the safest relevant anchors rather than inventing ids.
```

## Constant 2 | bytes 4105035..4105351
```text
你是 InkOS 的语义大纲选段器。
只选择当前章节真正需要的大纲段落。按语义相关性判断，不要按关键词重合机械选择。
只返回严格 JSON：{"selectedSources":["..."]}。必须使用候选里的精确 source id；不确定时选最安全的相关锚点，不要编造 id。
```

## Constant 22 | bytes 4105561..4105940
```text
You are InkOS's semantic context compiler.
Only compile the COMPRESSIBLE CONTEXT. The PROTECTED CONTEXT is binding reference material and must not be rewritten, summarized as a substitute, or weakened.
Output concise Markdown with source pointers. Preserve names, unresolved promises, evidence, timing, and constraints that may affect the next chapter. Drop low-relevance noise.
```

## Constant 23 | bytes 4105941..4106272
```text
你是 InkOS 的语义上下文编译器。
只能编译【可压缩上下文】。【受保护上下文】是绑定参照，不得改写、不得替代总结、不得削弱。
输出简洁 Markdown，保留来源指针。保留会影响下一章的人名、未兑现承诺、证据、时间点和约束，丢弃低相关噪声。
```

## Constant 37 | bytes 4106678..4106907
```text
You are a writing-context compressor. Compress the input into a writing context package of no more than 2500 words, keeping the facts, conflicts, hooks, and constraints most critical for writing chapter N. Use Markdown sections.
```

## Constant 53 | bytes 4107270..4107315
```text
build_outline_section_selector_system_prompt
```

## Constant 64 | bytes 4107473..4107516
```text
build_outline_section_selector_user_prompt
```

## Constant 65 | bytes 4107517..4107567
```text
build_compressible_context_compiler_system_prompt
```

## Constant 68 | bytes 4107662..4107710
```text
build_compressible_context_compiler_user_prompt
```

## Constant 69 | bytes 4107711..4107755
```text
build_legacy_context_compress_system_prompt
```

## Constant 70 | bytes 4107756..4107797
```text
novels\prompts\inkos\composer_prompts.py
```

## Constant 71 | bytes 4107798..4107845
```text
<module novels.prompts.inkos.composer_prompts>
```
