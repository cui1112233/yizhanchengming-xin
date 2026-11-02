// 日志页(docs/spec/06-webview-ui.md 页面 4):`log.recent` 拉最近 2000 条 + 订阅 `log_entry` 实时追加;
// 级别过滤 / 清空(只清本页视图) / 导出 ndjson(save 对话框 → `log.export`);附 CDP 调试面板(spec 09)。
import { signal } from '@preact/signals';
import { useEffect, useRef } from 'preact/hooks';
import * as api from '../api';
import { asRelayError } from '../api';
import type { LogEntry } from '../api';
import { CdpDebugPanel } from '../components/cdp-debug-panel';
import { cdpDebugEnabled, pushToast } from '../state';
import { appendCapped, filterByMinLevel, formatLogTime, rowKey } from '../view-model';
import './logs.css';

/** 与后端内存缓冲同容量。 */
export const LOG_CAP = 2000;
/** 与底部距离在此以内时视为"贴底",新日志来了自动滚到底。 */
const STICK_THRESHOLD_PX = 32;

export const LEVEL_OPTIONS: { value: string | null; label: string }[] = [
  { value: null, label: '全部' },
  { value: 'error', label: 'error' },
  { value: 'warn', label: 'warn 及以上' },
  { value: 'info', label: 'info 及以上' },
  { value: 'debug', label: 'debug 及以上' },
];

export const logEntries = signal<LogEntry[]>([]);
export const logMinLevel = signal<string | null>(null);
export const logsLoading = signal(false);
export const logsTab = signal<'app' | 'cdp'>('app');

export async function loadLogs() {
  if (logsLoading.value) return;
  logsLoading.value = true;
  try {
    const r = await api.logRecent(LOG_CAP);
    logEntries.value = r.entries;
  } catch (raw) {
    pushToast('error', `读取日志失败:${asRelayError(raw).message}`);
  } finally {
    logsLoading.value = false;
  }
}

export function appendLogEntry(e: LogEntry) {
  logEntries.value = appendCapped(logEntries.value, e, LOG_CAP);
}

/** 进日志页时同步 CDP 日志开关(设置页没打开过时信号还是初始值,tab 标题会误显示"未开启")。 */
export async function syncCdpEnabled() {
  try {
    const r = await api.debugCdpLog(1);
    cdpDebugEnabled.value = r.enabled;
  } catch {
    /* 读不到就保持原值 */
  }
}

function defaultExportName(now = new Date()): string {
  const p = (n: number) => String(n).padStart(2, '0');
  return `doubao-relay-logs-${now.getFullYear()}${p(now.getMonth() + 1)}${p(now.getDate())}-${p(now.getHours())}${p(now.getMinutes())}.ndjson`;
}

/** 导出:弹保存对话框 → 后端把缓冲(按当前级别过滤)写成 ndjson。取消对话框不提示。 */
export async function exportLogs() {
  let path: string | null;
  try {
    path = await api.saveFileDialog(defaultExportName(), 'ndjson');
  } catch (raw) {
    pushToast('error', `打开保存对话框失败:${asRelayError(raw).message}`);
    return;
  }
  if (!path) return;
  try {
    const r = await api.logExport(path, logMinLevel.value ?? undefined);
    pushToast('success', `已导出 ${r.count} 条日志到 ${r.path}`);
  } catch (raw) {
    pushToast('error', `导出日志失败:${asRelayError(raw).message}`);
  }
}

/** 结构化字段的紧凑展示:`k=v k2=v2`(字符串不加引号,其它 JSON 化)。 */
export function formatFields(fields: Record<string, unknown>): string {
  return Object.entries(fields)
    .map(([k, v]) => `${k}=${typeof v === 'string' ? v : JSON.stringify(v)}`)
    .join(' ');
}

function LogTable({ rows }: { rows: LogEntry[] }) {
  const scroller = useRef<HTMLDivElement>(null);
  const stick = useRef(true);
  useEffect(() => {
    const el = scroller.current;
    if (el && stick.current) el.scrollTop = el.scrollHeight;
  }, [rows.length]);
  const onScroll = () => {
    const el = scroller.current;
    if (!el) return;
    stick.current = el.scrollHeight - el.scrollTop - el.clientHeight <= STICK_THRESHOLD_PX;
  };
  return (
    <div class="log-scroller" ref={scroller} onScroll={onScroll}>
      <table class="table log-table">
        <thead>
          <tr>
            <th>时间</th>
            <th>级别</th>
            <th>来源</th>
            <th>消息</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((e) => {
            const extra = formatFields(e.fields ?? {});
            return (
              <tr key={rowKey(e)} data-level={e.level}>
                <td class="log-time">{formatLogTime(e.ts)}</td>
                <td class="log-level">
                  <span class={`level level-${e.level}`}>{e.level.toUpperCase()}</span>
                </td>
                <td class="log-target">{e.target}</td>
                <td class="log-message">
                  {e.message}
                  {extra && <span class="log-fields"> {extra}</span>}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

export function LogsPage() {
  useEffect(() => {
    void loadLogs();
    void syncCdpEnabled();
    return api.onEvent<LogEntry>('log_entry', appendLogEntry);
  }, []);
  const tab = logsTab.value;
  const all = logEntries.value;
  const rows = filterByMinLevel(all, logMinLevel.value);

  return (
    <section class="page page-logs">
      <header class="page-header">
        <h1 class="page-title">日志</h1>
        {tab === 'app' && (
          <div class="page-actions">
            <label class="select-label">
              级别
              <select
                class="select"
                value={logMinLevel.value ?? ''}
                onChange={(ev) => {
                  const v = (ev.currentTarget as HTMLSelectElement).value;
                  logMinLevel.value = v === '' ? null : v;
                }}
              >
                {LEVEL_OPTIONS.map((o) => (
                  <option key={o.label} value={o.value ?? ''}>
                    {o.label}
                  </option>
                ))}
              </select>
            </label>
            <button type="button" class="btn" disabled={all.length === 0} onClick={() => (logEntries.value = [])}>
              清空
            </button>
            <button type="button" class="btn" disabled={all.length === 0} onClick={exportLogs}>
              导出
            </button>
          </div>
        )}
      </header>

      <div class="tabs" role="tablist">
        <button
          type="button"
          role="tab"
          class={`tab${tab === 'app' ? ' active' : ''}`}
          aria-selected={tab === 'app'}
          onClick={() => (logsTab.value = 'app')}
        >
          应用日志
        </button>
        <button
          type="button"
          role="tab"
          class={`tab${tab === 'cdp' ? ' active' : ''}`}
          aria-selected={tab === 'cdp'}
          onClick={() => (logsTab.value = 'cdp')}
        >
          CDP 调试{cdpDebugEnabled.value ? '' : '(未开启)'}
        </button>
      </div>

      {tab === 'cdp' ? (
        <CdpDebugPanel />
      ) : all.length === 0 ? (
        <div class="empty-state" data-kind="no_logs">
          <p class="empty-title">还没有日志</p>
          <p class="empty-hint">操作账号或提交任务后,日志会实时出现在这里;也可以点击刷新重新读取</p>
          <button type="button" class="btn" disabled={logsLoading.value} onClick={loadLogs}>
            {logsLoading.value ? '读取中…' : '刷新'}
          </button>
        </div>
      ) : rows.length === 0 ? (
        <div class="empty-state" data-kind="filtered_out">
          <p class="empty-title">当前级别下没有日志</p>
          <p class="empty-hint">共 {all.length} 条,换一个级别试试</p>
          <button type="button" class="btn" onClick={() => (logMinLevel.value = null)}>
            显示全部
          </button>
        </div>
      ) : (
        <>
          <LogTable rows={rows} />
          <p class="muted table-footer">
            显示 {rows.length} / {all.length} 条,新日志自动追加到底部
          </p>
        </>
      )}
    </section>
  );
}
