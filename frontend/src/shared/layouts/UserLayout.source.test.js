import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const source = fs.readFileSync(path.join(path.dirname(fileURLToPath(import.meta.url)), 'UserLayout.jsx'), 'utf8');

test('global header status expires instead of permanently retaining API polling failures', () => {
  assert.match(source, /expiresAt/);
  assert.match(source, /window\.setTimeout\(\(\) =>/);
  assert.match(source, /setGlobalStatus\(current => current\.expiresAt === expiresAt \? \{ text: '', tone: 'idle'/);
});

test('avatar enters profile through the client route without opening the menu first', () => {
  const avatarButton = source.slice(
    source.indexOf('className="legacy-sidebar-avatar"'),
    source.indexOf('</button>', source.indexOf('className="legacy-sidebar-avatar"'))
  );

  assert.match(avatarButton, /window\.history\.pushState\(\{\}, '', '\/profile'\)/);
  assert.match(avatarButton, /window\.dispatchEvent\(new PopStateEvent\('popstate'\)\)/);
  assert.doesNotMatch(avatarButton, /setAccountCenterOpen\(true\)/);
  assert.doesNotMatch(avatarButton, /window\.location\.assign\('\/profile'\)/);
});

test('routes account-center pages through the dedicated layout while the avatar records a business origin', () => {
  assert.match(source, /captureAccountCenterOrigin/);
  assert.match(source, /<AccountCenterLayout/);
  assert.doesNotMatch(source, /accountCenterOpen/);
});
