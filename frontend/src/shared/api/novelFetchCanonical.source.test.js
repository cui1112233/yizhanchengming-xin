import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

const source = fs.readFileSync(path.resolve(process.cwd(), 'src/shared/api/novelFetch.js'), 'utf8');
const workshopSource = fs.readFileSync(path.resolve(process.cwd(), 'src/shared/api/novelFetchWorkshop.js'), 'utf8');
const embeddedWorkbenchSource = fs.readFileSync(path.resolve(process.cwd(), 'public/batch-rewrite/app.js'), 'utf8');
const embeddedRuntimeSource = fs.readFileSync(path.resolve(process.cwd(), 'public/batch-rewrite/v78-novel-fetch-v2-runtime.js'), 'utf8');
const packagedRuntimeSource = fs.readFileSync(path.resolve(process.cwd(), '../public/batch-rewrite/v78-novel-fetch-v2-runtime.js'), 'utf8');
const embeddedV78Sources = [
  'v78-novel-fetch-v2.js',
  'v78-novel-fetch-v2-config.js',
  'v78-novel-fetch-v2-layout.js',
  'v78-novel-fetch-v2-run-controls.js',
].map((name) => fs.readFileSync(path.resolve(process.cwd(), `public/batch-rewrite/${name}`), 'utf8'));
const packagedV78Sources = [
  'v78-novel-fetch-v2.js',
  'v78-novel-fetch-v2-config.js',
  'v78-novel-fetch-v2-layout.js',
  'v78-novel-fetch-v2-run-controls.js',
].map((name) => fs.readFileSync(path.resolve(process.cwd(), `../public/batch-rewrite/${name}`), 'utf8'));

describe('novel fetch uses one canonical Go workshop submit API', () => {
  it('does not create a second batch-rewrite submit path', () => {
    expect(source).not.toContain('/api/batch-rewrite/web-submit/');
    expect(source).not.toContain('/api/web-submit/');
    expect(source).toContain("const WEB_SUBMIT_API = '/api/novel-fetch-workshop/web-submit'");
    for (const suffix of ['config', 'environment', 'sync-configs', 'sync-styles', 'test-visible', 'preview', 'submit']) {
      expect(source).toContain(`WEB_SUBMIT_API}/` + suffix);
    }
  });
});

describe('novel fetch workshop uses one canonical Go workshop API', () => {
  it('does not call the legacy batch-rewrite issue log path', () => {
    expect(workshopSource).not.toContain('/api/batch-rewrite/issues');
    expect(workshopSource).toContain("'/api/novel-fetch-workshop/issues'");
  });

  it('keeps the embedded workbench on the same canonical web-submit namespace', () => {
    expect(embeddedWorkbenchSource).not.toContain('/api/batch-rewrite/web-submit/');
    expect(embeddedWorkbenchSource).not.toContain('/api/web-submit/');
    expect(embeddedWorkbenchSource).toContain('const WEB_SUBMIT_API = "/api/novel-fetch-workshop/web-submit"');
    expect(embeddedRuntimeSource).not.toContain("'/api/web-submit/");
    expect(embeddedRuntimeSource).toContain("'/api/novel-fetch-workshop/web-submit'");
    expect(packagedRuntimeSource).not.toContain("'/api/web-submit/");
    expect(packagedRuntimeSource).toContain("'/api/novel-fetch-workshop/web-submit'");
  });

  it('keeps every embedded V78 API helper on the canonical Go workshop namespace', () => {
    for (const source of [embeddedWorkbenchSource, embeddedRuntimeSource, packagedRuntimeSource, ...embeddedV78Sources, ...packagedV78Sources]) {
      expect(source).not.toContain("'/api/batch-rewrite'");
      expect(source).not.toContain('"/api/batch-rewrite"');
    }
    for (const source of [...embeddedV78Sources, ...packagedV78Sources]) {
      expect(source).toContain('/api/novel-fetch-workshop');
    }
  });
});
