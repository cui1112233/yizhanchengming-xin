# novels.services.narrative_control: recovered constant literals
Source SHA256: b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d
These are ordered constants, not reconstructed function bodies.

## Constant 42 | bytes 4350324..4350602
```text
Narrative control: pre-sanitizes planning text before it reaches the writer.

Ported from InkOS packages/core/src/utils/narrative-control.ts.
Strips hook IDs, chapter references, and AI-flavor hedging words from
memo/intent/context text so the writer LLM never even sees them.

```
