import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { describe, it, expect } from 'vitest';

const here = dirname(fileURLToPath(import.meta.url));
const source = readFileSync(join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

describe('工作台常显自动化状态行', () => {
  it('含常显状态行元素', () => {
    expect(source).toContain('data-testid="automation-status-line"');
  });
  it('状态行使用 automationStatus 快照', () => {
    expect(source).toMatch(/automationStatus\?\.state/);
    expect(source).toContain('最近错误');
  });
  it('覆盖运行中与失败/阻塞计数', () => {
    expect(source).toContain('运行中');
    expect(source).toContain('失败/阻塞');
  });
});
