# Auto-Map N8 Audio Category Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reuse saved per-book AI gender/style classifications to automatically choose a matching N8 audio category when the entire batch has one unambiguous shared mapping.

**Architecture:** Add a pure frontend mapping helper that reads `book.sourceMetadata.gender`, `style`, and `classifyStatus`, maps only explicit supported gender/style combinations to exact N8 category names/IDs, and returns a category only when every batch book is classified and resolves to the same N8 category. Wire it into the unified comic publication settings so inferred values update only while the category source is automatic; any user category selection becomes an explicit manual override and is not overwritten by later book metadata changes. Existing audio auto-selection continues to react to the selected organization/category.

**Tech Stack:** React, JavaScript ES modules, Node built-in test runner, existing Ant Design Select.

## Global Constraints

- Reuse saved AI metadata only; do not call an AI model again from publication settings.
- Do not guess when metadata is missing, unclassified, low-confidence/pending, unsupported, ambiguous, or inconsistent across books.
- Preserve user manual category selection and existing specified-audio selection behavior.
- Match N8 categories by the live category directory ID/name; use explicit deterministic mappings only.
- Do not commit or publish unless explicitly requested.

---

### Task 1: Define strict batch-to-N8 category mapping

**Files:**
- Create: `frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.js`
- Create: `frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.test.js`

**Interfaces:**
- Produces: `inferBatchN8AudioCategory(books, n8Categories)` returning `{ category, reason }`; `category` is one matching directory item or `null`; `reason` is a stable string enum.

- [ ] **Step 1: Write failing tests**

```js
import assert from 'node:assert/strict';
import test from 'node:test';
import { inferBatchN8AudioCategory } from './batchFactoryAudioCategory.js';

const categories = [
  { id: 9, name: '女-爽文（打脸）', status: 'ENABLE' },
  { id: 10, name: '古风-女悲伤', status: 'ENABLE' },
  { id: 11, name: '古风-男悲伤', status: 'ENABLE' },
  { id: 12, name: '古风-爽文（打脸）', status: 'ENABLE' }
];

test('maps a fully classified all-female revenge batch to the exact N8 category', () => {
  const result = inferBatchN8AudioCategory([
    { sourceMetadata: { classifyStatus: 'classified', gender: '女频', style: '现代爽文' } },
    { sourceMetadata: { classifyStatus: 'classified', gender: '女频', style: '女主打脸爽文' } }
  ], categories);
  assert.equal(result.category?.id, 9);
  assert.equal(result.reason, 'matched');
});

test('does not infer when books disagree, metadata is incomplete, or style is unsupported', () => {
  assert.equal(inferBatchN8AudioCategory([
    { sourceMetadata: { classifyStatus: 'classified', gender: '女频', style: '爽文' } },
    { sourceMetadata: { classifyStatus: 'classified', gender: '男频', style: '爽文' } }
  ], categories).reason, 'mixed');
  assert.equal(inferBatchN8AudioCategory([
    { sourceMetadata: { classifyStatus: 'pending_confirmation', gender: '女频', style: '爽文' } }
  ], categories).reason, 'unclassified');
  assert.equal(inferBatchN8AudioCategory([
    { sourceMetadata: { classifyStatus: 'classified', gender: '女频', style: '现代悬疑' } }
  ], categories).reason, 'unsupported');
});
```

- [ ] **Step 2: Run the test and verify the expected failure**

Run: `node --test src/user/pages/shuihuo/batchFactoryAudioCategory.test.js` from `frontend/`.
Expected: FAIL because the helper module does not exist.

- [ ] **Step 3: Implement the minimal explicit mapper**

Implement normalization for gender values `女频`/`女` and `男频`/`男`; recognize only explicit style tokens for `爽文`/`打脸` or `悲伤`. Map female revenge to exact N8 category name `女-爽文（打脸）`, female sadness to `古风-女悲伤`, male sadness to `古风-男悲伤`, and male revenge to `古风-爽文（打脸）`. Require `classifyStatus === 'classified'`, non-empty books, all books resolve to the same mapping, and exactly one enabled N8 directory item matches the mapped name. Do not infer male/female from missing metadata or silently select the closest category.

- [ ] **Step 4: Run the helper tests and verify pass**

Run: `node --test src/user/pages/shuihuo/batchFactoryAudioCategory.test.js`.
Expected: PASS.

- [ ] **Step 5: Commit only the helper and its tests when commit is explicitly requested**

```bash
git add frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.js frontend/src/user/pages/shuihuo/batchFactoryAudioCategory.test.js
git commit -m "feat: infer shared N8 audio category"
```

### Task 2: Wire inference into unified publication settings with manual lock

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js`

**Interfaces:**
- Consumes: `inferBatchN8AudioCategory(books, n8Categories)`.
- Reads: `batch.books` and each book's saved `sourceMetadata`.
- Persists with `comicPublishSettings`: `audioCategorySelectionMode: 'auto' | 'manual'`.

- [ ] **Step 1: Add failing assertions**

```js
assert.match(source, /inferBatchN8AudioCategory/);
assert.match(source, /audioCategorySelectionMode/);
assert.match(source, /根据批次 AI 分类自动匹配/);
assert.match(source, /分类不一致|分类未完成|需手动选择/);
```

- [ ] **Step 2: Run the source test and verify failure**

Run: `node --test src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js` from `frontend/`.
Expected: FAIL because the inference helper and source-mode state are not wired in.

- [ ] **Step 3: Implement automatic mode and manual override**

Pass `batch.books` through the unified modal into `BatchFactoryComicPublishSettingsForm`. When `audioCategorySelectionMode` is not `manual`, compute the inference from saved metadata and the current N8 category directory. If it returns a category, update `audioCategoryId`, `audioCategoryName`, and set mode `auto`; if inference becomes unsupported/mixed/unclassified, clear only values previously selected by auto mode and show the corresponding non-blocking explanation. On either category Select `onChange`, set mode `manual` and persist the selected category. Never overwrite a manual category due to subsequent classification metadata updates. Existing audio candidate effect then reloads on inferred category ID changes and continues to honor saved audio ID.

- [ ] **Step 4: Run focused source and mapper tests**

Run: `node --test src/user/pages/shuihuo/batchFactoryAudioCategory.test.js src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js` from `frontend/`.
Expected: PASS.

- [ ] **Step 5: Commit only the integration files when commit is explicitly requested**

```bash
git add frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.jsx frontend/src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js
git commit -m "feat: auto-fill shared comic audio category"
```

### Task 3: Verification

**Files:**
- Modify only files implicated by a failing test.

- [ ] **Step 1: Run focused frontend tests**

Run: `node --test src/user/pages/shuihuo/batchFactoryAudioCategory.test.js src/user/pages/shuihuo/BatchFactoryUnifiedSettingsModal.source.test.js src/user/pages/shuihuo/batchFactoryComicPublishAccounts.test.js` from `frontend/`.
Expected: PASS.

- [ ] **Step 2: Build frontend**

Run: `npm run build` from `frontend/`.
Expected: successful Vite production build.

- [ ] **Step 3: Validate representative cases**

Confirm: all books female + revenge map to live N8 category 9; mixed genders do not auto-select; any book lacking `classifyStatus: classified` does not auto-select; unsupported style does not auto-select; manual selection survives book metadata updates; automatic selection change refreshes the specified-audio candidates.
