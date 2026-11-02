// CDP 调试面板(spec 09 "CDP 日志调试面板"):`debug.cdp_log` 拉最近条目 + 订阅 `cdp_log_entry` 实时追加。
// 只在设置页勾了"开启 CDP 日志"后有数据;未开启时给出去设置页的入口。
import { signal } from '@preact/signals';
import { useEffect, useRef, useState } from 'preact/hooks';
import * as api from '../api';
import { asRelayError } from '../api';
import type { CdpLogEntry } from '../api';
import { cdpDebugEnabled, pushToast, route } from '../state';
import { appendCapped, formatLogTime, rowKey } from '../view-model';

/** 面板最多保留的条数(后端缓冲 2000;面板只看最近的)。 */
export const CDP_PANEL_CAP = 500;
/** 单元格里 JSON 预览的最大字符数。 */
const PREVIEW_CHARS = 160;

export const cdpEntries = signal<CdpLogEntry[]>([]);

/** 一条 CDP 日志的"内容"列:send 看 params,recv 响应看 result,事件看 event。 */
export function cdpPayload(e: CdpLogEntry): unknown {
  if (e.direction === 'send') return e.params;
  return e.event !== undefined ? e.event : e.result;
}

export function previewJson(v: unknown, max = PREVIEW_CHARS): string {
  if (v === undefined) return '';
  let s: string;
  try {
    s = JSON.stringify(v) ?? '';
  } catch {
    s = String(v);
  }
  return s.length > max ? `${s.slice(0, max)}…` : s;
}

/** 距底部多少像素以内视为"在底部",新条目到来时才自动滚动。 */
const STICK_THRESHOLD_PX = 24;

/** 是否在底部附近(用户上翻在看旧条目时不自动滚)。导出便于单测。 */
export function isNearBottom(el: { scrollHeight: number; scrollTop: number; clientHeight: number }): boolean {
  return el.scrollHeight - el.scrollTop - el.clientHeight <= STICK_THRESHOLD_PX;
}

/** 内容列:折叠时只算一次预览,完整 JSON 只在展开后才序列化。 */
function PayloadCell({ entry }: { entry: CdpLogEntry }) {
  const [open, setOpen] = useState(false);
  const payload = cdpPayload(entry);
  return (
    <details onToggle={(ev) => setOpen((ev.currentTarget as HTMLDetailsElement).open)}>
      <summary>{previewJson(payload) || <span class="muted">(空)</span>}</summary>
      {open && <pre>{JSON.stringify(payload ?? null, null, 2)}</pre>}
    </details>
  );
}

export async function loadCdpLog() {
  try {
    const r = await api.debugCdpLog(CDP_PANEL_CAP);
    cdpEntries.value = r.entries;
    cdpDebugEnabled.value = r.enabled;
  } catch (raw) {
    pushToast('error', `读取 CDP 日志失败:${asRelayError(raw).message}`);
  }
}

export function CdpDebugPanel() {
  const scroller = useRef<HTMLDivElement>(null);
  const stick = useRef(true);
  useEffect(() => {
    void loadCdpLog();
    return api.onEvent<CdpLogEntry>('cdp_log_entry', (e) => {
      cdpEntries.value = appendCapped(cdpEntries.value, e, CDP_PANEL_CAP);
    });
  }, []);
  const entries = cdpEntries.value;
  useEffect(() => {
    const el = scroller.current;
    if (el && stick.current) el.scrollTop = el.scrollHeight;
  }, [entries]);
  const onScroll = () => {
    const el = scroller.current;
    if (el) stick.current = isNearBottom(el);
  };

  if (!cdpDebugEnabled.value) {
    return (
      <div class="empty-state" data-kind="cdp_disabled">
        <p class="empty-title">CDP 日志未开启</p>
        <p class="empty-hint">调试豆包改版时,在设置 → 日志里勾选“开启 CDP 日志”,豆包的每条 CDP 收发会实时显示在这里</p>
        <button type="button" class="btn" onClick={() => (route.value = 'settings')}>
          去设置
        </button>
      </div>
    );
  }
  return (
    <div class="cdp-panel" ref={scroller} onScroll={onScroll}>
      <div class="cdp-toolbar">
        <span class="muted">共 {entries.length} 条(最多保留 {CDP_PANEL_CAP} 条,完整记录见 cdp-debug.log)</span>
        <button type="button" class="btn btn-small" onClick={() => (cdpEntries.value = [])}>
          清空
        </button>
      </div>
      {entries.length === 0 ? (
        <p class="muted">还没有 CDP 消息,操作一下账号或任务后会实时出现</p>
      ) : (
        <table class="table log-table cdp-table">
          <thead>
            <tr>
              <th>时间</th>
              <th>方向</th>
              <th>方法</th>
              <th>内容</th>
            </tr>
          </thead>
          <tbody>
            {entries.map((e) => (
              <tr key={rowKey(e)} data-direction={e.direction}>
                <td class="log-time">{formatLogTime(e.ts)}</td>
                <td class="cdp-direction">{e.direction === 'send' ? '→ 发送' : '← 接收'}</td>
                <td class="log-target">{e.method ?? <span class="muted">—</span>}</td>
                <td class="cdp-payload">
                  <PayloadCell entry={e} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
