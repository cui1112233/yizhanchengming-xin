// doubao_status_changed 处理:断线清列表,但账号切换期间(豆包 reload 造成的短暂断线)保留"切换中"行。
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { AppConfig, DoubaoStatus, TaskDetail } from './api';
import { accountList, appConfig, appConfigLoading, doubaoStatus, switching, toasts } from './state';

const invoke = vi.fn<(cmd: string, args?: unknown) => Promise<unknown>>();
vi.mock('@tauri-apps/api/core', () => ({ invoke: (cmd: string, args?: unknown) => invoke(cmd, args) }));
vi.mock('@tauri-apps/api/event', () => ({ listen: vi.fn(() => Promise.resolve(() => {})) }));

const { applyDoubaoStatus, applyTaskCompleted, bootstrap } = await import('./app');

const PAIRED_CONFIG: AppConfig = {
  server: { enabled: true, host: '127.0.0.1', port: 8787, bearerToken: 'token' },
  concurrency: { globalMax: 3, perAccountMax: 1, queueDepthMax: 50 },
  storage: { resultRetentionHours: 24, metadataRetentionDays: 7, uploadExpiryHours: 1 },
  logging: { level: 'info', fileMaxSizeMb: 10, fileRetentionDays: 7, cdpDebugLog: false },
  doubao: { cdpPort: 9707, cdpAddress: '127.0.0.1', autoConnectOnStartup: true, uiCompatCheckOnStartup: true },
  v88Executor: { baseUrl: 'http://localhost:4001', executorId: 'lex_1', executorToken: 'executor-token' },
  ui: { minimizeToTrayOnClose: true, startMinimized: false },
};

function status(patch: Partial<DoubaoStatus> = {}): DoubaoStatus {
  return {
    running: true,
    cdpConnected: true,
    cdpPort: 9707,
    doubaoVersionDetected: '2.32.5',
    uiCompatCheck: 'ok',
    uiCompatDetails: {
      status: 'ok',
      accountButtonFound: true,
      inputFound: true,
      sendButtonFound: true,
      checkedAt: 1,
    },
    chatTargets: 1,
    reconnecting: false,
    lastCheckAt: 1,
    ...patch,
  };
}

const LIST = {
  current: '她是翻哥',
  accounts: [
    { name: '她是翻哥', subscription: null, label: '她是翻哥', isCurrent: true, isAmbiguous: false },
    { name: '李四', subscription: null, label: '李四', isCurrent: false, isAmbiguous: false },
  ],
  readAt: 1,
};

beforeEach(() => {
  invoke.mockReset();
  invoke.mockImplementation(async (cmd) => {
    if (cmd === 'app_get_config') return PAIRED_CONFIG;
    if (cmd === 'app_get_version') return { version: '1.3.0' };
    if (cmd === 'app_api_status') return { enabled: true, listening: true, host: '127.0.0.1', port: 8787, error: null };
    if (cmd === 'doubao_status') return status();
    throw new Error(`unexpected ${cmd}`);
  });
  doubaoStatus.value = status();
  accountList.value = LIST;
  switching.value = null;
  toasts.value = [];
  appConfig.value = null;
  appConfigLoading.value = true;
});

describe('applyDoubaoStatus', () => {
  it('更新状态;CDP 断开且没有切换进行中 → 清空列表', () => {
    const s = status({ cdpConnected: false, reconnecting: true });
    applyDoubaoStatus(s);
    expect(doubaoStatus.value).toBe(s);
    expect(accountList.value).toBeNull();
    expect(toasts.value).toEqual([]);
  });

  it('切换进行中遇到心跳断线(豆包 reload)→ 保留列表,"切换中"行不被未连接空状态替换', () => {
    switching.value = '李四';
    applyDoubaoStatus(status({ cdpConnected: false, reconnecting: true }));
    expect(accountList.value).toEqual(LIST);
  });

  it('切换进行中但豆包进程退出 → 仍清空列表', () => {
    switching.value = '李四';
    applyDoubaoStatus(status({ running: false, cdpConnected: false }));
    expect(accountList.value).toBeNull();
  });

  it('CDP 仍连着时不动列表', () => {
    applyDoubaoStatus(status({ lastCheckAt: 2 }));
    expect(accountList.value).toEqual(LIST);
  });

  it('running 从 true 变 false 时弹一次错误 toast;之后保持 false 不再重复', () => {
    applyDoubaoStatus(status({ running: false, cdpConnected: false }));
    expect(toasts.value.map((t) => t.kind)).toEqual(['error']);
    applyDoubaoStatus(status({ running: false, cdpConnected: false, lastCheckAt: 2 }));
    expect(toasts.value).toHaveLength(1);
  });

  it('启动时读取 V88 配对配置，账号页不会因尚未打开设置而误报未配对', async () => {
    await bootstrap();
    expect(invoke).toHaveBeenCalledWith('app_get_config', undefined);
    expect(appConfig.value).toEqual(PAIRED_CONFIG);
    expect(appConfigLoading.value).toBe(false);
  });

  it('任务准备失败时，右下角提示直接带上页面的具体原因', () => {
    applyTaskCompleted({
      taskId: '01M5D3H2GG',
      status: 'failed',
      error: {
        code: 'generation_preparing_failed',
        message: '生成准备阶段失败,请检查豆包当前页面状态后重试',
        details: { js_exception: '未找到所选生成比例。' },
      },
    } as TaskDetail);
    expect(toasts.value.at(-1)?.message).toBe('任务 01M5D3H2GG 失败:未找到所选生成比例。');
  });
});
