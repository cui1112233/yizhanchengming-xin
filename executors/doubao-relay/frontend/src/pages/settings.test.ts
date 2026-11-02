// 设置页数据逻辑:分组补丁、保存后的 toast / 配置刷新、复制 token 60s 后清剪贴板、重新生成 token、立即清理。
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { AppConfig } from '../api';
import { appConfig, cdpDebugEnabled, toasts } from '../state';

const invoke = vi.fn<(cmd: string, args?: unknown) => Promise<unknown>>();
const clipboard = { text: '' as string | null };
vi.mock('@tauri-apps/api/core', () => ({ invoke: (cmd: string, args?: unknown) => invoke(cmd, args) }));
vi.mock('@tauri-apps/api/event', () => ({ listen: vi.fn(() => Promise.resolve(() => {})) }));
vi.mock('@tauri-apps/plugin-clipboard-manager', () => ({
  writeText: vi.fn(async (t: string) => {
    clipboard.text = t;
  }),
  readText: vi.fn(async () => clipboard.text ?? ''),
}));
vi.mock('@tauri-apps/plugin-dialog', () => ({ save: vi.fn(async () => null) }));
vi.mock('../components/confirm-dialog', () => ({ confirm: vi.fn(async () => true) }));

const settings = await import('./settings');
const {
  CLIPBOARD_CLEAR_MS,
  GROUP_FIELDS,
  copyToken,
  draft,
  draftFrom,
  loadSettings,
  mergeSavedGroup,
  patchFor,
  registerVersionClick,
  resetDeveloperModeForTest,
  regenerateToken,
  runGc,
  saveGroup,
} = settings;

const TOKEN = 'a'.repeat(64);

function config(patch: Partial<AppConfig> = {}): AppConfig {
  return {
    server: { enabled: false, host: '127.0.0.1', port: 8787, bearerToken: TOKEN },
    concurrency: { globalMax: 3, perAccountMax: 1, queueDepthMax: 50 },
    storage: { resultRetentionHours: 24, metadataRetentionDays: 7, uploadExpiryHours: 1 },
    logging: { level: 'info', fileMaxSizeMb: 10, fileRetentionDays: 7, cdpDebugLog: false },
    doubao: { cdpPort: 9707, cdpAddress: '127.0.0.1', autoConnectOnStartup: true, uiCompatCheckOnStartup: true },
    ui: { minimizeToTrayOnClose: true, startMinimized: false },
    v88Executor: { baseUrl: '', executorId: '', executorToken: '' },
    ...patch,
  };
}

let current: AppConfig;

beforeEach(() => {
  resetDeveloperModeForTest();
  current = config();
  invoke.mockReset();
  invoke.mockImplementation(async (cmd, args) => {
    switch (cmd) {
      case 'app_get_config':
        return current;
      case 'app_get_version':
        return { version: '0.1.0', buildTarget: 'macos', tauriVersion: '2.12.4', rustcVersion: '1.94.0' };
      case 'app_save_config': {
        const patch = (args as { patch: Record<string, Record<string, unknown>> }).patch;
        const needsRestart = patch.server?.port !== undefined || patch.doubao?.cdpPort !== undefined;
        current = Object.fromEntries(
          Object.entries(current).map(([g, v]) => [g, { ...(v as object), ...(patch[g] ?? {}) }]),
        ) as AppConfig;
        return { ok: true, needsRestart };
      }
      case 'app_regenerate_token':
        current = { ...current, server: { ...current.server, bearerToken: 'b'.repeat(64) } };
        return { bearerToken: current.server.bearerToken };
      case 'maintenance_gc':
        return { resultFilesRemoved: 2, tasksRemoved: 1, uploadsRemoved: 0 };
      default:
        throw new Error(`unexpected ${cmd}`);
    }
  });
  appConfig.value = null;
  draft.value = null;
  cdpDebugEnabled.value = false;
  toasts.value = [];
  clipboard.text = '';
});

describe('developer mode', () => {
  it('unlocks developer mode only on the seventh version click', () => {
    for (let i = 1; i < 7; i += 1) {
      expect(registerVersionClick(1000 + i * 100).unlocked).toBe(false);
    }
    expect(registerVersionClick(1700).unlocked).toBe(true);
  });
});

afterEach(() => {
  vi.useRealTimers();
});

describe('loadSettings / draftFrom / patchFor', () => {
  it('读配置 + 版本,草稿按配置初始化,CDP 开关信号同步', async () => {
    current = config({ logging: { level: 'debug', fileMaxSizeMb: 10, fileRetentionDays: 7, cdpDebugLog: true } });
    await loadSettings();
    expect(appConfig.value).toEqual(current);
    expect(draft.value).toEqual(draftFrom(current));
    expect(draft.value?.level).toBe('debug');
    expect(cdpDebugEnabled.value).toBe(true);
    expect(settings.version.value?.version).toBe('0.1.0');
  });

  it('每组补丁只含该组字段', () => {
    const d = draftFrom(config());
    expect(patchFor('server', d)).toEqual({ server: { enabled: false, port: 8787 } });
    expect(patchFor('concurrency', d)).toEqual({ concurrency: { globalMax: 3, perAccountMax: 1 } });
    expect(patchFor('storage', d)).toEqual({ storage: { resultRetentionHours: 24, metadataRetentionDays: 7 } });
    expect(patchFor('logging', d)).toEqual({ logging: { level: 'info', cdpDebugLog: false } });
    expect(patchFor('advanced', d)).toEqual({
      doubao: { cdpPort: 9707, autoConnectOnStartup: true },
      ui: { minimizeToTrayOnClose: true },
    });
  });
});

describe('saveGroup', () => {
  it('合法 → 调 app_save_config 并重读配置;不需重启时 success toast', async () => {
    await loadSettings();
    draft.value = { ...draft.value!, level: 'debug', cdpDebugLog: true };
    expect(await saveGroup('logging')).toBe(true);
    expect(invoke).toHaveBeenCalledWith('app_save_config', {
      patch: { logging: { level: 'debug', cdpDebugLog: true } },
    });
    expect(appConfig.value?.logging.level).toBe('debug');
    expect(cdpDebugEnabled.value).toBe(true);
    expect(toasts.value.map((t) => t.kind)).toEqual(['success']);
    expect(toasts.value[0].message).toBe('已保存');
  });

  it('端口改动 → needsRestart → 提示重启', async () => {
    await loadSettings();
    draft.value = { ...draft.value!, port: 9000 };
    expect(await saveGroup('server')).toBe(true);
    expect(appConfig.value?.server.port).toBe(9000);
    expect(toasts.value[0].message).toMatch(/重启/);
  });

  it('前端校验不过 → 不调后端,错误 toast', async () => {
    await loadSettings();
    draft.value = { ...draft.value!, port: 0 };
    expect(await saveGroup('server')).toBe(false);
    expect(invoke).not.toHaveBeenCalledWith('app_save_config', expect.anything());
    expect(toasts.value[0].kind).toBe('info');
    expect(toasts.value[0].message).toMatch(/65535/);
  });

  it('后端拒绝 → 错误 toast 带 reason,草稿保留', async () => {
    await loadSettings();
    invoke.mockImplementationOnce(async () => {
      throw { code: 'request_malformed', message: '配置取值不合法', details: { reason: 'server.port 不能为 0' } };
    });
    draft.value = { ...draft.value!, globalMax: 5 };
    expect(await saveGroup('concurrency')).toBe(false);
    expect(toasts.value[0].message).toContain('server.port 不能为 0');
    expect(draft.value?.globalMax).toBe(5);
  });
});

describe('保存一组不丢其他组的未保存编辑', () => {
  it('改了日志组 CDP 开关又去保存高级组 → 高级组刷新,CDP 勾选仍保留(10-live-probe Task 5.1 #2)', async () => {
    await loadSettings();
    draft.value = { ...draft.value!, cdpDebugLog: true, globalMax: 7, cdpPort: 9800 };
    expect(await saveGroup('advanced')).toBe(true);
    expect(appConfig.value?.doubao.cdpPort).toBe(9800);
    expect(appConfig.value?.logging.cdpDebugLog).toBe(false);
    expect(draft.value?.cdpPort).toBe(9800);
    expect(draft.value?.cdpDebugLog).toBe(true);
    expect(draft.value?.globalMax).toBe(7);
  });

  it('mergeSavedGroup 只用新配置覆盖该组字段;GROUP_FIELDS 覆盖所有草稿字段且与 patchFor 对应', () => {
    const prev = { ...draftFrom(config()), port: 1234, level: 'trace' };
    const fresh = config({ logging: { level: 'warn', fileMaxSizeMb: 10, fileRetentionDays: 7, cdpDebugLog: true } });
    const merged = mergeSavedGroup('logging', prev, fresh);
    expect(merged.level).toBe('warn');
    expect(merged.cdpDebugLog).toBe(true);
    expect(merged.port).toBe(1234);
    const all = Object.values(GROUP_FIELDS).flat().sort();
    expect(all).toEqual(Object.keys(draftFrom(config())).sort());
    expect(mergeSavedGroup('server', null, fresh)).toEqual(draftFrom(fresh));
  });

  it('重新生成 token 不动草稿', async () => {
    await loadSettings();
    draft.value = { ...draft.value!, globalMax: 9 };
    expect(await regenerateToken()).toBe(true);
    expect(draft.value?.globalMax).toBe(9);
  });
});

describe('copyToken', () => {
  it('写入剪贴板;60s 后剪贴板仍是它则清空', async () => {
    vi.useFakeTimers();
    expect(await copyToken(TOKEN)).toBe(true);
    expect(clipboard.text).toBe(TOKEN);
    expect(toasts.value[0].message).toContain('60 秒');
    await vi.advanceTimersByTimeAsync(CLIPBOARD_CLEAR_MS - 1);
    expect(clipboard.text).toBe(TOKEN);
    await vi.advanceTimersByTimeAsync(1);
    expect(clipboard.text).toBe('');
  });

  it('60s 内用户复制了别的内容 → 不动剪贴板', async () => {
    vi.useFakeTimers();
    await copyToken(TOKEN);
    clipboard.text = '用户自己复制的东西';
    await vi.advanceTimersByTimeAsync(CLIPBOARD_CLEAR_MS);
    expect(clipboard.text).toBe('用户自己复制的东西');
  });

  it('空 token 不复制', async () => {
    expect(await copyToken('')).toBe(false);
    expect(clipboard.text).toBe('');
  });
});

describe('regenerateToken / runGc', () => {
  it('确认后调 app_regenerate_token,配置里的 token 更新并显示', async () => {
    await loadSettings();
    expect(await regenerateToken()).toBe(true);
    expect(appConfig.value?.server.bearerToken).toBe('b'.repeat(64));
    expect(settings.tokenVisible.value).toBe(true);
    expect(toasts.value[0].kind).toBe('success');
  });

  it('立即清理 → toast 汇总三项计数', async () => {
    expect(await runGc()).toBe(true);
    expect(toasts.value[0].message).toBe('清理完成:删除 2 个结果文件、1 条任务记录、0 条上传');
  });
});
