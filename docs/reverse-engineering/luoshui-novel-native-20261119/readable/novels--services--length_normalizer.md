# novels.services.length_normalizer: recovered constant literals
Source SHA256: b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d
These are ordered constants, not reconstructed function bodies.

## Constant 67 | bytes 4338967..4339384
```text
Normalize chapter length if outside hard range.

    Args:
        content: Full chapter text (may include # title).
        target_word_count: Target character/word count.
        chat_fn: LLM caller accepting (messages, kwargs) -> response text.
        language: "zh" or "en".
        chapter_intent: Optional chapter intent for context.

    Returns:
        LengthNormResult with possibly adjusted content.
    
```

## Constant 68 | bytes 4339385..4339567
```text
Chapter length normalization via LLM expand/compress.

Ported from InkOS length-normalizer.ts.
Called after writer output, before the review cycle.
Single-pass only (no recursion).

```

## Constant nested | bytes 4340657..4340671
```text
system_prompt
```

## Constant nested | bytes 4340672..4340684
```text
user_prompt
```
