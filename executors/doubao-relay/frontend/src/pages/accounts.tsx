// 账号页(docs/spec/06-webview-ui.md 页面 1)。
// 数据流:signals(state.ts)← 命令返回值 / 事件 payload;本页只负责触发命令与渲染。
import { useEffect } from 'preact/hooks';
import * as api from '../api';
import { asRelayError, isDoubaoUnavailable } from '../api';
import type { RelayError } from '../api';
import { confirm } from '../components/confirm-dialog';
import {
  accountList,
  accountsError,
  accountsLoading,
  doubaoStatus,
  pushToast,
  reconnecting,
  switching,
} from '../state';
import { accountRows, emptyStateFor } from '../view-model';

const AMBIGUOUS_HINT = '同名,不支持通过 API 调用';

/** 这两种错误等价于"豆包未连接",状态栏与空状态已经在表达,不再弹 toast。 */
const SILENT_DOUBAO_CODES = new Set(['doubao_cdp_disconnected', 'doubao_not_running']);

/**
 * 读账号列表失败时要不要静默:只有"豆包没连上"这一类(错误码本身说明断线,或状态栏已显示未连接)
 * 才交给空状态表达;已连接却读不到(`doubao_ui_changed` 豆包改版、`doubao_not_ready` evaluate 超时等)
 * 必须让用户看到后端带"下一步建议"的中文 message。
 */
export function isSilentAccountsError(e: RelayError, status = doubaoStatus.value): boolean {
  if (!isDoubaoUnavailable(e)) return false;
  if (SILENT_DOUBAO_CODES.has(e.code)) return true;
  return emptyStateFor(status, null) === 'not_connected';
}

/** 错误文案:message 加上 `doubao_ui_changed` 的 `failed_selectors`(排查改版时有用)。 */
export function describeAccountsError(e: RelayError): string {
  const d = e.details as { failed_selectors?: unknown; failedSelectors?: unknown } | undefined;
  const raw = d && typeof d === 'object' ? (d.failed_selectors ?? d.failedSelectors) : undefined;
  const failed = Array.isArray(raw) ? raw.filter((s): s is string => typeof s === 'string') : [];
  return failed.length > 0 ? `${e.message}(未命中:${failed.join('、')})` : e.message;
}

/** 读取进行中又来了 `force` 请求:先记下,本次结束后再强刷一次,不能静默丢掉。 */
let pendingForce = false;

/** 读账号列表;`force` 跳过后端 2s 缓存。豆包未连/未登录由空状态表达,不弹 toast。 */
export async function loadAccounts(force: boolean) {
  if (accountsLoading.value) {
    pendingForce ||= force;
    return;
  }
  accountsLoading.value = true;
  try {
    accountList.value = force ? await api.accountRefresh() : await api.accountList();
    accountsError.value = null;
  } catch (raw) {
    const e = asRelayError(raw);
    if (e.code === 'doubao_not_logged_in') {
      accountList.value = { current: null, accounts: [], readAt: Date.now() };
      accountsError.value = null;
    } else if (isSilentAccountsError(e)) {
      accountList.value = null;
      accountsError.value = null;
    } else if (isDoubaoUnavailable(e)) {
      // 已连接却读不到:列表不可信,空状态里显示原因,错误 toast 常驻
      accountList.value = null;
      accountsError.value = describeAccountsError(e);
      pushToast('error', `读取账号失败:${accountsError.value}`);
    } else {
      pushToast('error', `读取账号失败:${e.message}`);
    }
  } finally {
    accountsLoading.value = false;
    if (pendingForce) {
      pendingForce = false;
      void loadAccounts(true);
    }
  }
}

/** "重连豆包" / "连接豆包":`doubao.reconnect(force=false)`;成功后顺手刷新账号。 */
export async function reconnectDoubao() {
  if (reconnecting.value) return;
  reconnecting.value = true;
  try {
    const r = await api.doubaoReconnect(false);
    const text: Record<string, string> = {
      already_connected: '豆包已连接',
      reconnected: '已重新连接豆包',
      restarted: '已重启豆包并连接',
      started: '已启动豆包并连接',
    };
    pushToast('success', text[r.action] ?? `豆包连接完成(${r.action})`);
    doubaoStatus.value = await api.doubaoStatus();
    void loadAccounts(true);
  } catch (raw) {
    const e = asRelayError(raw);
    pushToast('error', `连接豆包失败:${e.message}`);
    try {
      doubaoStatus.value = await api.doubaoStatus();
    } catch {
      /* 状态读不到就保持原值 */
    }
  } finally {
    reconnecting.value = false;
  }
}

async function switchTo(name: string) {
  if (switching.value) return;
  const ok = await confirm({
    title: '切换账号',
    message: `确定切换到账号“${name}”吗?豆包页面会重新加载,进行中的任务会被中断。`,
    confirmLabel: '切换',
  });
  // 等确认框期间可能已经有别的切换开始(如 Enter 误触发第二次调用),再查一次防重入
  if (!ok || switching.value) return;
  switching.value = name;
  try {
    const r = await api.accountSwitch(name);
    // 成功 toast 由 account_switched 事件统一发;这里只兜底没收到事件的情况
    if (!r.ok) pushToast('error', `切换到“${name}”未成功,请刷新后重试`);
  } catch (raw) {
    const e = asRelayError(raw);
    pushToast('error', `切换账号失败:${e.message}`);
  } finally {
    switching.value = null;
    void loadAccounts(false);
  }
}

function Skeleton() {
  return (
    <div class="skeleton" aria-busy="true" aria-label="加载中">
      <div class="skeleton-line w-40" />
      <div class="skeleton-line" />
      <div class="skeleton-line" />
      <div class="skeleton-line w-70" />
    </div>
  );
}

function EmptyState({ kind }: { kind: 'not_connected' | 'not_logged_in' | 'empty' }) {
  const s = doubaoStatus.value;
  if (kind === 'not_connected') {
    const exited = s !== null && !s.running;
    const busy = reconnecting.value || (s?.reconnecting ?? false);
    return (
      <div class="empty-state" data-kind={kind}>
        <p class="empty-title">{exited ? '豆包已退出,等待重启' : '豆包未连接'}</p>
        <p class="empty-hint">
          {exited
            ? '豆包进程已不在运行,点击下方按钮由 relay 重新启动并连接'
            : s?.reconnecting
              ? '正在尝试重新连接豆包,请稍候'
              : '连接后才能读取账号列表'}
        </p>
        <button type="button" class="btn btn-primary" disabled={busy} onClick={reconnectDoubao}>
          {busy ? '连接中…' : exited ? '重连豆包' : '连接豆包'}
        </button>
      </div>
    );
  }
  if (kind === 'not_logged_in') {
    return (
      <div class="empty-state" data-kind={kind}>
        <p class="empty-title">豆包尚未登录任何账号</p>
        <p class="empty-hint">请先在豆包中登录,然后点击刷新</p>
        <button type="button" class="btn" onClick={() => loadAccounts(true)}>
          刷新
        </button>
      </div>
    );
  }
  const err = accountsError.value;
  return (
    <div class="empty-state" data-kind={kind} data-error={err ? true : undefined}>
      <p class="empty-title">{err ? '读取账号列表失败' : '还没有读到账号列表'}</p>
      <p class="empty-hint">{err ?? '点击刷新从豆包重新读取'}</p>
      <button type="button" class="btn" onClick={() => loadAccounts(true)}>
        刷新
      </button>
    </div>
  );
}

export function AccountsPage() {
  const status = doubaoStatus.value;
  const list = accountList.value;
  const loading = accountsLoading.value;
  const rows = accountRows(list);
  const empty = emptyStateFor(status, list);
  const current = list?.accounts.find((a) => a.isCurrent) ?? null;
  const currentName = current?.name ?? list?.current ?? null;

  useEffect(() => {
    if (list === null && status?.cdpConnected) void loadAccounts(false);
  }, [status?.cdpConnected]);

  return (
    <section class="page page-accounts">
      <header class="page-header">
        <h1 class="page-title">账号</h1>
        <div class="page-actions">
          <button
            type="button"
            class="btn"
            disabled={loading || !status?.cdpConnected}
            onClick={() => loadAccounts(true)}
          >
            {loading ? '读取中…' : '刷新'}
          </button>
          <button
            type="button"
            class="btn"
            disabled={reconnecting.value}
            onClick={reconnectDoubao}
          >
            {reconnecting.value ? '连接中…' : '重连豆包'}
          </button>
        </div>
      </header>

      {empty === 'not_connected' ? (
        <EmptyState kind="not_connected" />
      ) : loading && list === null ? (
        <Skeleton />
      ) : empty ? (
        <EmptyState kind={empty} />
      ) : (
        <>
          <p class="current-account">
            当前账号:{' '}
            {currentName ? (
              <strong>
                {currentName}
                {current?.subscription ? ` (${current.subscription})` : ''}
              </strong>
            ) : (
              <span class="muted">未知</span>
            )}
          </p>
          <table class="table accounts-table">
            <thead>
              <tr>
                <th>名字</th>
                <th>套餐</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row, i) => {
                const isSwitching = switching.value === row.name;
                return (
                  <tr
                    key={`${row.name}-${i}`}
                    class={row.isCurrent ? 'row-current' : undefined}
                    data-ambiguous={row.isAmbiguous || undefined}
                  >
                    <td>
                      {row.isCurrent && (
                        <span class="star" aria-label="当前账号">
                          ★{' '}
                        </span>
                      )}
                      {row.isAmbiguous && (
                        <span class="warn" title={AMBIGUOUS_HINT} aria-label="同名账号">
                          ⚠{' '}
                        </span>
                      )}
                      {row.name}
                    </td>
                    <td>{row.subscription ?? <span class="muted">—</span>}</td>
                    <td>
                      {row.isCurrent ? (
                        <span class="badge badge-current">当前</span>
                      ) : row.isAmbiguous ? (
                        <span class="muted">{AMBIGUOUS_HINT}</span>
                      ) : isSwitching ? (
                        <span class="switching">
                          <span class="spinner" aria-hidden="true" />
                          切换中…
                        </span>
                      ) : (
                        <button
                          type="button"
                          class="btn btn-small"
                          disabled={switching.value !== null || !row.canSwitch}
                          onClick={() => switchTo(row.name)}
                        >
                          切换
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
          <p class="muted table-footer">共 {rows.length} 个已登录账号。</p>
        </>
      )}
    </section>
  );
}
