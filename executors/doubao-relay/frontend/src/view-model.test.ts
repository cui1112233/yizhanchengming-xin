import { describe, expect, it } from 'vitest';
import { accountRows, apiStatusText, doubaoDot, versionLines, doubaoStatusText, emptyStateFor } from './view-model';
import type { AccountList, DoubaoStatus } from './api';

describe('doubaoDot', () => {
  it('豆包运行且 CDP 已连 → green', () => {
    expect(doubaoDot({ running: true, cdpConnected: true })).toBe('green');
  });

  it('豆包未运行或 CDP 未连 → red', () => {
    expect(doubaoDot({ running: false, cdpConnected: false })).toBe('red');
    expect(doubaoDot({ running: true, cdpConnected: false })).toBe('red');
    expect(doubaoDot({ running: false, cdpConnected: true })).toBe('red');
  });

  it('重连中 → yellow,优先级高于 running/cdpConnected', () => {
    expect(doubaoDot({ running: true, cdpConnected: true, reconnecting: true })).toBe('yellow');
    expect(doubaoDot({ running: false, cdpConnected: false, reconnecting: true })).toBe('yellow');
  });
});

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

function list(
  accounts: Array<Partial<AccountList['accounts'][number]> & { name: string }>,
  current: string | null = accounts.find((a) => a.isCurrent)?.name ?? null,
): AccountList {
  return {
    current,
    accounts: accounts.map((a) => ({
      subscription: null,
      label: a.name,
      isCurrent: false,
      isAmbiguous: false,
      ...a,
    })),
    readAt: 1,
  };
}

describe('doubaoStatusText', () => {
  it('未知状态 / 未连接 / 重连中 / 已退出 / 已连接 各有文案', () => {
    expect(doubaoStatusText(null)).toBe('状态未知');
    expect(doubaoStatusText(status())).toBe('已连接');
    expect(doubaoStatusText(status({ reconnecting: true, cdpConnected: false }))).toBe('重连中');
    expect(doubaoStatusText(status({ running: false, cdpConnected: false }))).toBe(
      '豆包已退出,等待重启',
    );
    expect(doubaoStatusText(status({ running: true, cdpConnected: false }))).toBe('未连接');
  });
});

describe('accountRows', () => {
  it('当前账号标 isCurrent 且不可切换,其它账号可切换', () => {
    const rows = accountRows(
      list([
        { name: '她是翻哥', subscription: '体验套餐', label: '她是翻哥\n体验套餐', isCurrent: true },
        { name: '李四', subscription: '高级套餐', label: '李四 高级套餐' },
      ]),
    );
    expect(rows).toEqual([
      {
        name: '她是翻哥',
        subscription: '体验套餐',
        isCurrent: true,
        isAmbiguous: false,
        canSwitch: false,
      },
      { name: '李四', subscription: '高级套餐', isCurrent: false, isAmbiguous: false, canSwitch: true },
    ]);
  });

  it('同名账号 isAmbiguous 且不可切换,即使不是当前账号', () => {
    const rows = accountRows(
      list([
        { name: '张三', subscription: '标准套餐', isCurrent: true },
        { name: '李四', subscription: '高级套餐', isAmbiguous: true },
        { name: '李四', subscription: '标准套餐', isAmbiguous: true },
      ]),
    );
    expect(rows.map((r) => r.canSwitch)).toEqual([false, false, false]);
    expect(rows.map((r) => r.isAmbiguous)).toEqual([false, true, true]);
  });

  it('保持后端顺序;套餐缺失为 null;后端没标 isCurrent 时按 current 名字兜底', () => {
    const rows = accountRows(list([{ name: '王五' }, { name: '赵六' }], '赵六'));
    expect(rows.map((r) => r.name)).toEqual(['王五', '赵六']);
    expect(rows[0].subscription).toBeNull();
    expect(rows[0].isCurrent).toBe(false);
    expect(rows[1].isCurrent).toBe(true);
    expect(rows[1].canSwitch).toBe(false);
  });

  it('列表为 null 时返回空数组', () => {
    expect(accountRows(null)).toEqual([]);
  });
});

describe('emptyStateFor', () => {
  it('没有状态 / CDP 未连 / 豆包未运行 → not_connected', () => {
    expect(emptyStateFor(null, null)).toBe('not_connected');
    expect(emptyStateFor(status({ cdpConnected: false }), null)).toBe('not_connected');
    expect(emptyStateFor(status({ running: false, cdpConnected: false }), list([{ name: 'a' }]))).toBe(
      'not_connected',
    );
    expect(emptyStateFor(status({ reconnecting: true, cdpConnected: false }), null)).toBe(
      'not_connected',
    );
  });

  it('已连但列表为空且无当前账号 → not_logged_in', () => {
    expect(emptyStateFor(status(), list([], null))).toBe('not_logged_in');
  });

  it('已连但尚未读到列表,或列表为空却有当前账号 → empty', () => {
    expect(emptyStateFor(status(), null)).toBe('empty');
    expect(emptyStateFor(status(), list([], '她是翻哥'))).toBe('empty');
  });

  it('已连且有账号 → null(正常显示表格)', () => {
    expect(emptyStateFor(status(), list([{ name: '她是翻哥', isCurrent: true }]))).toBeNull();
  });
});

describe('apiStatusText', () => {
  it('shows address when listening, otherwise 未启动', () => {
    expect(apiStatusText({ listening: true, host: '127.0.0.1', port: 8787 })).toBe('监听中 127.0.0.1:8787');
    expect(apiStatusText({ listening: true, host: '::1', port: 9 })).toBe('监听中 [::1]:9');
    expect(apiStatusText({ listening: false, host: '127.0.0.1', port: 8787 })).toBe('未启动');
    expect(apiStatusText(null)).toBe('未启动');
  });
});

describe('versionLines', () => {
  it('shows relay version and detected Doubao version', () => {
    expect(versionLines('0.1.0', '2.32.6')).toEqual(['一战晟铭执行器 v0.1.0', '豆包 2.32.6']);
  });
  it('omits unknown parts', () => {
    expect(versionLines('0.1.0', null)).toEqual(['一战晟铭执行器 v0.1.0']);
    expect(versionLines(null, null)).toEqual([]);
  });
});
