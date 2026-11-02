import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'BatchFactoryNovelList.jsx'), 'utf8');
const apiSource = fs.readFileSync(path.join(here, '../../../shared/api/batchFactoryV11.js'), 'utf8');
const stylesheet = fs.readFileSync(path.join(here, '../shuihuo-production.css'), 'utf8');

test('shows the complete compiled comic prompt separately from the editable page description', () => {
  assert.match(apiSource, /export function getComicPagePromptPreview/);
  assert.match(source, /getComicPagePromptPreview/);
  assert.match(source, /分格导演描述/);
  assert.match(source, /最终完整出图提示词/);
  assert.match(source, /完整提示词会包含画风、全局前后缀、画幅与分格结构、负面限制/);
  assert.match(source, /referenceImageUrls/);
  assert.match(stylesheet, /\.batch-factory-comic-prompt-preview\s*\{/);
  assert.match(stylesheet, /\.batch-factory-comic-reference-summary\s*\{/);
});

test('edits and saves the director description and base image description independently', () => {
  assert.match(source, /const \[comicDirectorPrompt, setComicDirectorPrompt\] = useState\(''\)/);
  assert.match(source, /comicPrompt:\s*comicDirectorPrompt/);
  assert.match(source, /comicImagePrompt:\s*comicPrompt/);
  assert.match(source, /分格导演描述（可编辑）/);
  assert.match(source, /本页基础出图描述（可编辑）/);
});

test('loads and renders the final compiled prompt only for developer accounts', () => {
  assert.match(source, /getCurrentAccount/);
  assert.match(source, /isDeveloper={isDeveloper}/);
  assert.match(source, /promptKind !== 'visual' \|\| !isDeveloper/);
  assert.match(source, /isDeveloper \? <div className="batch-factory-comic-developer-preview">/);
});

test('renders structured comic references as previewable material chips', () => {
  assert.match(source, /<ComicReferenceChips/);
  assert.match(source, /referenceImages={comicPromptPreview\.referenceImages}/);
  assert.match(stylesheet, /\.batch-factory-comic-reference-strip\s*\{/);
});
