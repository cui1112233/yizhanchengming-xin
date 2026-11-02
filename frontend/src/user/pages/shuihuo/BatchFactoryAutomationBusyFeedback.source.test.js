import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { describe, it, expect } from 'vitest';

const here = dirname(fileURLToPath(import.meta.url));
const source = readFileSync(join(here, 'BatchFactoryNovelList.jsx'), 'utf8');

describe('runAutomationAction 可见反馈', () => {
  it('不存在静默吞点击的裸 return', () => {
    expect(source).not.toContain('if (!batch?.id || automationBusy) return;');
  });
  it('批次未加载时给出错误提示', () => {
    expect(source).toContain('页面批次尚未加载，请刷新后重试');
  });
  it('busy 时给出可见提示', () => {
    expect(source).toContain('上一次操作仍在执行，请稍候');
  });
  it('有 30 秒看门狗复位', () => {
    expect(source).toContain('automationBusyAtRef');
    expect(source).toMatch(/30\s*\*\s*1000/);
    expect(source).toContain('上一次请求超时，已重置状态');
  });
  it('提交期间有 loading 常驻提示', () => {
    expect(source).toContain('正在提交自动化任务');
  });
});
