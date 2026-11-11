import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const globalCSS = fs.readFileSync(path.join(dirname, 'global.css'), 'utf8');
const agentCSS = fs.readFileSync(path.join(dirname, 'agent-workspace.css'), 'utf8');

const themeSelectors = {
  dark: ':root',
  light: "[data-theme='light']"
};

const requiredVariables = [
  '--text-primary',
  '--text-secondary',
  '--text-tertiary',
  '--text-disabled',
  '--text-link',
  '--text-danger',
  '--text-warning',
  '--text-success',
  '--surface-base',
  '--surface-panel',
  '--surface-input',
  '--surface-elevated',
  '--border-default',
  '--border-strong',
  '--theme-primary',
  '--theme-primary-hover',
  '--theme-primary-active',
  '--text-on-accent',
  '--canvas-background',
  '--canvas-grid',
  '--node-surface',
  '--edge-color',
  '--viewport-background',
  '--timeline-background'
];

function cssBlock(selector) {
  const escaped = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const match = globalCSS.match(new RegExp(`${escaped}\\s*\\{([^}]*)\\}`, 's'));
  assert.ok(match, `missing theme block ${selector}`);
  return match[1];
}

for (const [id, selector] of Object.entries(themeSelectors)) {
  test(`${id} CSS theme defines the complete semantic workspace contract`, () => {
    const block = cssBlock(selector);
    for (const variable of requiredVariables) {
      assert.match(block, new RegExp(`${variable}\\s*:`), `${id} must define ${variable}`);
    }
  });
}

test('global stylesheet ships exactly the dark default and the light palette', () => {
  const declared = [...globalCSS.matchAll(/\[data-theme='([^']+)'\]\s*\{/g)].map(match => match[1]);
  assert.deepEqual([...new Set(declared)], ['light']);
});

test('Agent canvas and Director Desk consume semantic theme surfaces', () => {
  assert.match(agentCSS, /--agent-canvas:\s*var\(--surface-base\)/);
  assert.match(agentCSS, /--agent-panel:\s*var\(--surface-panel\)/);
  assert.match(agentCSS, /--director-bg:\s*var\(--legacy-bg\)/);
  assert.match(agentCSS, /background:\s*var\(--agent-canvas\)/);
});
