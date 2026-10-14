import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

const source = fs.readFileSync(path.resolve(process.cwd(), 'src/shared/api/novelFetch.js'), 'utf8');
const workshopSource = fs.readFileSync(path.resolve(process.cwd(), 'src/shared/api/novelFetchWorkshop.js'), 'utf8');

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
});
