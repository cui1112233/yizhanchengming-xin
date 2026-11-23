# Nine-Function Art Generation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce nine clear, consistent 3D function illustrations and one 3×3 review board for the online function catalog before integrating them into the mini-program.

**Architecture:** Treat the images as design-source assets first, separate from the 2 MiB mini-program package. A manifest locks the nine concepts and filenames; ImageGen produces transparent high-resolution PNGs; a deterministic normalization step aligns visible bounds and creates the review board. No homepage code or routes change until the user approves the board.

**Tech Stack:** OpenAI ImageGen, PNG with alpha, FFmpeg/FFprobe, Markdown design evidence.

## Global Constraints

- The catalog contains exactly nine entries: `private-custom`, `novel-creation`, `batch-review`, `shuihuo-production`, `comic-production`, `ai-cut`, `cm-canvas`, `voice-studio`, `plugins`.
- Source images remain high-resolution PNGs with transparent backgrounds; package optimization happens only after visual approval.
- Every image uses the same cute premium 3D style, three-quarter view, upper-left soft light, cyan rim light, purple fill light, and warm pink highlights.
- No image contains text, letters, logos, emoji, phone frames, screenshots, or UI labels.
- Visible subject bounds target 72%–78% of a square canvas with a shared bottom baseline.
- This plan does not modify mini-program routes, function pages, backend code, or production deployment.

---

### Task 1: Lock the nine-image manifest

**Files:**
- Create: `docs/design-assets/miniprogram-nine-functions/manifest.json`
- Create: `docs/design-assets/miniprogram-nine-functions/README.md`

**Interfaces:**
- Consumes: the approved nine-function design spec.
- Produces: filenames and exact semantic prompts used by every generation call.

- [ ] **Step 1: Create the manifest with nine unique entries**

The JSON keys and filenames must be:

```json
{
  "private-custom": "01-private-custom.png",
  "novel-creation": "02-novel-creation.png",
  "batch-review": "03-batch-review.png",
  "shuihuo-production": "04-shuihuo-production.png",
  "comic-production": "05-comic-production.png",
  "ai-cut": "06-ai-cut.png",
  "cm-canvas": "07-cm-canvas.png",
  "voice-studio": "08-voice-studio.png",
  "plugins": "09-plugins.png"
}
```

- [ ] **Step 2: Validate count and uniqueness**

Run:

```powershell
$m = Get-Content -Raw docs/design-assets/miniprogram-nine-functions/manifest.json | ConvertFrom-Json
$p = $m.PSObject.Properties
if ($p.Count -ne 9) { throw "Expected 9 images" }
if (($p.Value | Sort-Object -Unique).Count -ne 9) { throw "Duplicate filenames" }
```

Expected: no output and exit code `0`.

- [ ] **Step 3: Commit the manifest**

```powershell
git add docs/design-assets/miniprogram-nine-functions/manifest.json docs/design-assets/miniprogram-nine-functions/README.md
git commit -m "docs: lock nine-function art manifest"
```

### Task 2: Generate the nine transparent source illustrations

**Files:**
- Create: `docs/design-assets/miniprogram-nine-functions/source/01-private-custom.png`
- Create: `docs/design-assets/miniprogram-nine-functions/source/02-novel-creation.png`
- Create: `docs/design-assets/miniprogram-nine-functions/source/03-batch-review.png`
- Create: `docs/design-assets/miniprogram-nine-functions/source/04-shuihuo-production.png`
- Create: `docs/design-assets/miniprogram-nine-functions/source/05-comic-production.png`
- Create: `docs/design-assets/miniprogram-nine-functions/source/06-ai-cut.png`
- Create: `docs/design-assets/miniprogram-nine-functions/source/07-cm-canvas.png`
- Create: `docs/design-assets/miniprogram-nine-functions/source/08-voice-studio.png`
- Create: `docs/design-assets/miniprogram-nine-functions/source/09-plugins.png`

**Interfaces:**
- Consumes: manifest concepts and the approved 3D reference direction.
- Produces: one transparent source PNG per function.

- [ ] **Step 1: Generate the content-creation trio**

Use the shared style prompt plus these subjects: magical open book and feather pen; purple storybook with star bookmark; three stacked books with a scanning ring and check marks.

- [ ] **Step 2: Generate the production trio**

Use the same style prompt plus these subjects: glowing clapperboard with batch light trails; curved film strip containing an anime frame; smart scissors cutting a luminous editing timeline.

- [ ] **Step 3: Generate the AI-tools trio**

Use the same style prompt plus these subjects: friendly robot with floating connected workflow cards; studio microphone with sound waves and music notes; three interlocking 3D extension modules with connector ports.

- [ ] **Step 4: Verify alpha and dimensions**

Run FFprobe on all nine files and require PNG plus an alpha-capable pixel format. Reject any file with a baked checkerboard, black background, text, or missing subject.

### Task 3: Normalize visible bounds without reducing source clarity

**Files:**
- Create: `docs/design-assets/miniprogram-nine-functions/normalized/*.png`

**Interfaces:**
- Consumes: nine generated source PNGs.
- Produces: nine square review assets with consistent visual size and baseline.

- [ ] **Step 1: Detect non-transparent bounds**

Use FFmpeg `alphaextract,cropdetect` for each source and record the detected crop rectangle in `README.md`.

- [ ] **Step 2: Normalize onto square canvases**

Scale each detected subject to fit within 76% of a square canvas and center it horizontally with the shared bottom baseline. Preserve the source PNGs unchanged.

- [ ] **Step 3: Verify the visible-size tolerance**

Calculate normalized visible bounds and require the largest-to-smallest visible width and height variance to remain within approximately 8%.

### Task 4: Build and inspect the 3×3 review board

**Files:**
- Create: `docs/design-assets/miniprogram-nine-functions/nine-functions-board.png`
- Create: `docs/design-assets/miniprogram-nine-functions/design-qa.md`

**Interfaces:**
- Consumes: nine normalized PNGs.
- Produces: a single review image and a finding list for user approval.

- [ ] **Step 1: Compose the board**

Place the images in manifest order on a dark navy `3 x 3` board with equal cells and gutters. Add labels outside the image-safe areas only on the review board; source PNGs remain text-free.

- [ ] **Step 2: Review the five fidelity surfaces**

Check subject meaning, visible size, style consistency, lighting/material consistency, and alpha-edge quality. Record every mismatch as P1, P2, or P3 in `design-qa.md`.

- [ ] **Step 3: Regenerate any P1/P2 outlier**

Replace only the failing source asset, rerun normalization, rebuild the board, and record the before/after result.

- [ ] **Step 4: Present the board for user approval**

Show `nine-functions-board.png` and link the nine source files. Do not integrate them into `miniprogram/src` until the user approves the board.

### Task 5: Commit the approved design-source package

**Files:**
- Add: `docs/design-assets/miniprogram-nine-functions/**`

**Interfaces:**
- Consumes: the user-approved board and nine source images.
- Produces: a stable source-art checkpoint for the later mini-program integration plan.

- [ ] **Step 1: Run the final asset inventory check**

Require exactly nine source PNGs, nine normalized PNGs, one board, one manifest, one README, and one QA report.

- [ ] **Step 2: Commit only the design-source package**

```powershell
git add docs/design-assets/miniprogram-nine-functions
git commit -m "design: add nine-function illustration set"
```

- [ ] **Step 3: Stop before integration**

Report the approved asset paths and open a separate implementation cycle for homepage catalog, routes, and package-size handling.
