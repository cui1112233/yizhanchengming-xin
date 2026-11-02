// Task 5.1:设置页 / 日志页 / 快捷键的纯函数视图模型。
import { describe, expect, it } from 'vitest';
import {
  appendCapped,
  apiExamples,
  filterByMinLevel,
  formatLogTime,
  gcSummary,
  levelRank,
  maskToken,
  PER_ACCOUNT_MAX_FIXED,
  rowKey,
  saveResultMessage,
  settingsApiExamples,
  shortcutFor,
  shouldClearClipboard,
  urlHost,
  validateSettings,
} from './view-model';

describe('maskToken', () => {
  it('长 token 只露首尾各 4 位', () => {
    const t = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef';
    expect(maskToken(t)).toBe('0123…cdef');
  });
  it('短 token 全部打码,空串给占位', () => {
    expect(maskToken('abcdefgh')).toBe('••••••••');
    expect(maskToken('abc')).toBe('•••');
    expect(maskToken('')).toBe('(未生成)');
  });
});

describe('apiExamples', () => {
  const all = (port = 8790, token = 'tok', host?: string) =>
    apiExamples(port, token, host).map((e) => e.command).join('\n');

  it('覆盖全部 OpenAI 兼容接口,端口来自参数', () => {
    const s = all();
    for (const path of [
      'http://127.0.0.1:8790/health',
      'http://127.0.0.1:8790/v1/models',
      'http://127.0.0.1:8790/v1/accounts',
      'http://127.0.0.1:8790/v1/accounts/switch',
      'http://127.0.0.1:8790/v1/images/generations',
      'http://127.0.0.1:8790/v1/video/generations',
      'http://127.0.0.1:8790/v1/video/generations/<id>',
    ]) {
      expect(s).toContain(path);
    }
    expect(s).not.toContain('/api/v1/');
    expect(s).toContain('"model":"doubao-image-4.5"');
    expect(s).toContain('"model":"doubao-video-fast"');
    expect(s).toContain('"ref_image":"https://example.com/ref.png"');
  });

  it('除健康检查外都带 Bearer 头;POST 带 JSON 头', () => {
    const ex = apiExamples(8787, 'tok');
    expect(ex[0].title).toBe('账号列表');
    expect(ex[1].title).toBe('切换账号');
    for (const e of ex) {
      if (e.title === '健康检查') expect(e.command).not.toContain('Authorization');
      else expect(e.command).toContain("-H 'Authorization: Bearer tok'");
    }
    const posts = ex.filter((e) => e.command.includes(' -d '));
    expect(posts.length).toBe(4);
    for (const e of posts) expect(e.command).toContain("-H 'Content-Type: application/json'");
  });

  it('token 里的单引号会被转义,不破坏 shell 引号', () => {
    expect(all(8787, "a'b")).toContain("Bearer a'\\''b'");
  });
});

describe('validateSettings', () => {
  it('合法值无错误', () => {
    expect(
      validateSettings({
        port: 8787,
        globalMax: 3,
        perAccountMax: 1,
        cdpPort: 9707,
        resultRetentionHours: 24,
        metadataRetentionDays: 7,
      }),
    ).toEqual({});
  });
  it('端口 / 并发 / 保留期越界给出中文提示', () => {
    const errs = validateSettings({
      port: 0,
      globalMax: 0,
      perAccountMax: 2.5,
      cdpPort: 70000,
      resultRetentionHours: 0,
      metadataRetentionDays: -1,
    });
    expect(Object.keys(errs).sort()).toEqual(
      ['cdpPort', 'globalMax', 'metadataRetentionDays', 'perAccountMax', 'port', 'resultRetentionHours'].sort(),
    );
    expect(errs.port).toMatch(/1.*65535/);
  });
  it('NaN(输入框清空)也算错误', () => {
    expect(validateSettings({ port: Number.NaN }).port).toBeTruthy();
  });
});

describe('saveResultMessage', () => {
  it('不需要重启 → 已保存;需要重启 → 提示重启', () => {
    expect(saveResultMessage(false)).toBe('已保存');
    expect(saveResultMessage(true)).toMatch(/重启/);
  });
});

describe('shouldClearClipboard', () => {
  it('剪贴板仍是刚复制的 token 才清', () => {
    expect(shouldClearClipboard('tok', 'tok')).toBe(true);
    expect(shouldClearClipboard('其它内容', 'tok')).toBe(false);
    expect(shouldClearClipboard(null, 'tok')).toBe(false);
    expect(shouldClearClipboard('', 'tok')).toBe(false);
  });
});

describe('gcSummary', () => {
  it('三项计数拼成一句中文', () => {
    expect(gcSummary({ resultFilesRemoved: 2, tasksRemoved: 1, uploadsRemoved: 0 })).toBe(
      '清理完成:删除 2 个结果文件、1 条任务记录、0 条上传',
    );
  });
});

describe('日志过滤 / 追加', () => {
  const e = (level: string, ts = 0) => ({ ts, level, target: 't', message: level, fields: {} });

  it('levelRank:error 最严重,未知级别最不严重', () => {
    expect(levelRank('error')).toBeLessThan(levelRank('warn'));
    expect(levelRank('warn')).toBeLessThan(levelRank('info'));
    expect(levelRank('info')).toBeLessThan(levelRank('debug'));
    expect(levelRank('debug')).toBeLessThan(levelRank('trace'));
    expect(levelRank('WARN')).toBe(levelRank('warn'));
    expect(levelRank('weird')).toBeGreaterThan(levelRank('trace'));
  });

  it('filterByMinLevel:null 不过滤;warn 只留 warn/error', () => {
    const list = [e('debug'), e('info'), e('warn'), e('error')];
    expect(filterByMinLevel(list, null)).toBe(list);
    expect(filterByMinLevel(list, 'warn').map((x) => x.level)).toEqual(['warn', 'error']);
    expect(filterByMinLevel(list, 'info').map((x) => x.level)).toEqual(['info', 'warn', 'error']);
  });

  it('appendCapped:超过上限丢最旧', () => {
    const list = [e('a', 1), e('b', 2)];
    expect(appendCapped(list, e('c', 3), 2).map((x) => x.ts)).toEqual([2, 3]);
    expect(appendCapped(list, e('c', 3), 10).map((x) => x.ts)).toEqual([1, 2, 3]);
  });

  it('formatLogTime:本地时间 HH:mm:ss.SSS', () => {
    const d = new Date(2026, 9, 16, 10, 33, 5, 7);
    expect(formatLogTime(d.getTime())).toBe('10:33:05.007');
  });
});

describe('shortcutFor', () => {
  const key = (k: string, mods: Partial<KeyboardEvent> = {}) =>
    ({ key: k, metaKey: false, ctrlKey: false, altKey: false, shiftKey: false, ...mods }) as KeyboardEvent;

  it('macOS 用 Cmd,其它平台用 Ctrl', () => {
    expect(shortcutFor(key(',', { metaKey: true }), true)).toBe('settings');
    expect(shortcutFor(key(',', { ctrlKey: true }), true)).toBeNull();
    expect(shortcutFor(key(',', { ctrlKey: true }), false)).toBe('settings');
    expect(shortcutFor(key(',', { metaKey: true }), false)).toBeNull();
  });

  it('五个快捷键映射;大小写不敏感;Alt/Shift 组合不算', () => {
    expect(shortcutFor(key('l', { metaKey: true }), true)).toBe('logs');
    expect(shortcutFor(key('L', { metaKey: true }), true)).toBe('logs');
    expect(shortcutFor(key('r', { metaKey: true }), true)).toBe('refresh_accounts');
    expect(shortcutFor(key('n', { metaKey: true }), true)).toBe('new_task');
    expect(shortcutFor(key('k', { metaKey: true }), true)).toBe('search_tasks');
    expect(shortcutFor(key('k', { metaKey: true, shiftKey: true }), true)).toBeNull();
    expect(shortcutFor(key('k', { metaKey: true, altKey: true }), true)).toBeNull();
    expect(shortcutFor(key('x', { metaKey: true }), true)).toBeNull();
    expect(shortcutFor(key('l'), true)).toBeNull();
  });
});

describe('apiExamples 主机', () => {
  it('按 server.host 生成 URL,IPv6 加方括号,通配地址换回环', () => {
    expect(apiExamples(8787, 't', '::1')[0].command).toContain('http://[::1]:8787/v1/accounts');
    expect(apiExamples(8787, 't', '127.0.0.1')[0].command).toContain('http://127.0.0.1:8787/');
    expect(urlHost('0.0.0.0')).toBe('127.0.0.1');
    expect(urlHost('::')).toBe('[::1]');
    expect(urlHost('[::1]')).toBe('[::1]');
    expect(urlHost('')).toBe('127.0.0.1');
  });
});

describe('settingsApiExamples', () => {
  const token = '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef';
  const server = { host: '127.0.0.1', port: 8787, bearerToken: token };

  it('token 未显示时示例里只有打码值,绝不含明文 token', () => {
    const s = settingsApiExamples(server, false).map((e) => e.command).join('\n');
    expect(s).not.toContain(token);
    expect(s).toContain("Bearer 0123…cdef'");
  });

  it('点了"显示"后示例才带完整 token', () => {
    expect(settingsApiExamples(server, true)[0].command).toContain(`Bearer ${token}'`);
  });

  it('走 server.host', () => {
    expect(settingsApiExamples({ ...server, host: '::1' }, false)[0].command).toContain('http://[::1]:8787/');
  });
});

describe('validateSettings 每账号并发', () => {
  it('只接受固定值 1', () => {
    expect(PER_ACCOUNT_MAX_FIXED).toBe(1);
    expect(validateSettings({ perAccountMax: 1 })).toEqual({});
    expect(validateSettings({ perAccountMax: 3 }).perAccountMax).toMatch(/固定为 1/);
    expect(validateSettings({ perAccountMax: 0 }).perAccountMax).toBeDefined();
  });
});

describe('rowKey', () => {
  it('同一对象 key 稳定,不同对象 key 不同;封顶裁掉旧行后留下的行 key 不变', () => {
    const a = { ts: 1 };
    const b = { ts: 1 };
    expect(rowKey(a)).toBe(rowKey(a));
    expect(rowKey(a)).not.toBe(rowKey(b));
    let list = [a, b];
    const before = list.map(rowKey);
    const c = { ts: 2 };
    list = appendCapped(list, c, 2);
    expect(list.map(rowKey)).toEqual([before[1], rowKey(c)]);
  });
});
