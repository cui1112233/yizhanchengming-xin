// 设置页(docs/spec/06-webview-ui.md 页面 5):六个分组,每组独立"保存"(`app.save_config` 深合并补丁)。
// token:打码显示 / 复制(60s 后若剪贴板仍是它则清空,spec 12)/ 重新生成(立刻生效);"立即清理"调 `maintenance.gc`。
import { signal } from '@preact/signals';
import { useEffect } from 'preact/hooks';
import type { ComponentChildren, JSX } from 'preact';
import * as api from '../api';
import { asRelayError } from '../api';
import type { AppConfig, ConfigPatch, VersionInfo } from '../api';
import { confirm } from '../components/confirm-dialog';
import { apiStatus, appConfig, cdpDebugEnabled, pushToast } from '../state';
import {
  gcSummary,
  maskToken,
  saveResultMessage,
  settingsApiExamples,
  shouldClearClipboard,
  validateSettings,
} from '../view-model';
import type { SettingsNumbers } from '../view-model';
import './settings.css';

/** 复制 token 后多久清剪贴板(spec 12:60 秒)。 */
export const CLIPBOARD_CLEAR_MS = 60_000;

export const RESULT_RETENTION_OPTIONS = [1, 6, 12, 24, 48, 72, 168];
export const METADATA_RETENTION_OPTIONS = [1, 3, 7, 14, 30];
export const LOG_LEVEL_OPTIONS = ['error', 'warn', 'info', 'debug', 'trace'];
export const DEVELOPER_MODE_KEY = 'yizhanchengming.executor.developerMode';
const VERSION_CLICK_WINDOW_MS = 1500;

function storedDeveloperMode(): boolean {
  try {
    return globalThis.localStorage?.getItem(DEVELOPER_MODE_KEY) === 'enabled';
  } catch {
    return false;
  }
}

export const developerMode = signal(storedDeveloperMode());
let versionClickCount = 0;
let lastVersionClickAt = 0;

export function registerVersionClick(now = Date.now()): { unlocked: boolean; remaining: number } {
  if (developerMode.value) return { unlocked: true, remaining: 0 };
  if (lastVersionClickAt > 0 && now - lastVersionClickAt > VERSION_CLICK_WINDOW_MS) versionClickCount = 0;
  lastVersionClickAt = now;
  versionClickCount += 1;
  const remaining = Math.max(0, 7 - versionClickCount);
  if (remaining === 0) {
    developerMode.value = true;
    try {
      globalThis.localStorage?.setItem(DEVELOPER_MODE_KEY, 'enabled');
    } catch {
      // 浏览器禁用本地存储时，本次运行仍可使用开发者模式。
    }
    return { unlocked: true, remaining: 0 };
  }
  return { unlocked: false, remaining };
}

export function resetDeveloperModeForTest() {
  developerMode.value = false;
  versionClickCount = 0;
  lastVersionClickAt = 0;
  try {
    globalThis.localStorage?.removeItem(DEVELOPER_MODE_KEY);
  } catch {
    // 测试环境可能没有 localStorage。
  }
}

export type Group = 'server' | 'concurrency' | 'storage' | 'logging' | 'advanced';

/** 表单草稿(从配置拷贝;数字框允许暂时为 NaN,由 validateSettings 挡住)。 */
export type Draft = {
  serverEnabled: boolean;
  port: number;
  globalMax: number;
  perAccountMax: number;
  resultRetentionHours: number;
  metadataRetentionDays: number;
  level: string;
  cdpDebugLog: boolean;
  cdpPort: number;
  autoConnectOnStartup: boolean;
  minimizeToTrayOnClose: boolean;
};

export const draft = signal<Draft | null>(null);
export const settingsLoading = signal(false);
export const saving = signal<Group | null>(null);
export const tokenVisible = signal(false);
export const regenerating = signal(false);
export const cleaning = signal(false);
export const version = signal<VersionInfo | null>(null);
export const v88BaseUrl = signal('');
export const v88PairCode = signal('');
export const v88Pairing = signal(false);

export function draftFrom(c: AppConfig): Draft {
  return {
    serverEnabled: c.server.enabled,
    port: c.server.port,
    globalMax: c.concurrency.globalMax,
    perAccountMax: c.concurrency.perAccountMax,
    resultRetentionHours: c.storage.resultRetentionHours,
    metadataRetentionDays: c.storage.metadataRetentionDays,
    level: c.logging.level,
    cdpDebugLog: c.logging.cdpDebugLog,
    cdpPort: c.doubao.cdpPort,
    autoConnectOnStartup: c.doubao.autoConnectOnStartup,
    minimizeToTrayOnClose: c.ui.minimizeToTrayOnClose,
  };
}

/** 某一组的补丁(只含该组字段)。 */
export function patchFor(group: Group, d: Draft): ConfigPatch {
  switch (group) {
    case 'server':
      return { server: { enabled: d.serverEnabled, port: d.port } };
    case 'concurrency':
      return { concurrency: { globalMax: d.globalMax, perAccountMax: d.perAccountMax } };
    case 'storage':
      return {
        storage: {
          resultRetentionHours: d.resultRetentionHours,
          metadataRetentionDays: d.metadataRetentionDays,
        },
      };
    case 'logging':
      return { logging: { level: d.level, cdpDebugLog: d.cdpDebugLog } };
    case 'advanced':
      return {
        doubao: { cdpPort: d.cdpPort, autoConnectOnStartup: d.autoConnectOnStartup },
        ui: { minimizeToTrayOnClose: d.minimizeToTrayOnClose },
      };
  }
}

/** 某一组要校验的数字字段。 */
export function numbersFor(group: Group, d: Draft): SettingsNumbers {
  switch (group) {
    case 'server':
      return { port: d.port };
    case 'concurrency':
      return { globalMax: d.globalMax, perAccountMax: d.perAccountMax };
    case 'storage':
      return { resultRetentionHours: d.resultRetentionHours, metadataRetentionDays: d.metadataRetentionDays };
    case 'advanced':
      return { cdpPort: d.cdpPort };
    default:
      return {};
  }
}

/** 每组在草稿里对应的字段。 */
export const GROUP_FIELDS: Record<Group, (keyof Draft)[]> = {
  server: ['serverEnabled', 'port'],
  concurrency: ['globalMax', 'perAccountMax'],
  storage: ['resultRetentionHours', 'metadataRetentionDays'],
  logging: ['level', 'cdpDebugLog'],
  advanced: ['cdpPort', 'autoConnectOnStartup', 'minimizeToTrayOnClose'],
};

/** 保存某组后刷新草稿:只用服务端新值覆盖该组字段,其他组未保存的编辑原样保留。 */
export function mergeSavedGroup(group: Group, prev: Draft | null, c: AppConfig): Draft {
  const fresh = draftFrom(c);
  if (!prev) return fresh;
  const next: Draft = { ...prev };
  for (const k of GROUP_FIELDS[group]) {
    (next as Record<keyof Draft, unknown>)[k] = fresh[k];
  }
  return next;
}

function setAppConfig(c: AppConfig) {
  appConfig.value = c;
  cdpDebugEnabled.value = c.logging.cdpDebugLog;
}

function applyConfig(c: AppConfig) {
  setAppConfig(c);
  draft.value = draftFrom(c);
  v88BaseUrl.value = c.v88Executor.baseUrl;
}

export async function pairV88(): Promise<boolean> {
  if (v88Pairing.value) return false;
  if (!v88BaseUrl.value.trim() || !v88PairCode.value.trim()) {
    pushToast('info', '请填写 V88 地址和配对码');
    return false;
  }
  v88Pairing.value = true;
  try {
    const result = await api.appPairV88(v88BaseUrl.value, v88PairCode.value);
    v88PairCode.value = '';
    const config = await api.appGetConfig();
    setAppConfig(config);
    pushToast('success', `已连接 V88，执行器 ${result.executorId}`);
    return true;
  } catch (raw) {
    pushToast('error', `连接 V88 失败:${asRelayError(raw).message}`);
    return false;
  } finally {
    v88Pairing.value = false;
  }
}

export async function chooseAvailablePort(): Promise<boolean> {
  try {
    const result = await api.appFindAvailablePort();
    setDraft('port', result.port);
    pushToast('success', `已选择可用端口 ${result.port}，保存并重启后生效`);
    return true;
  } catch (raw) {
    pushToast('error', `查找可用端口失败：${asRelayError(raw).message}`);
    return false;
  }
}

export async function loadSettings() {
  if (settingsLoading.value) return;
  settingsLoading.value = true;
  try {
    const [c, v] = await Promise.all([api.appGetConfig(), api.appGetVersion()]);
    applyConfig(c);
    version.value = v;
  } catch (raw) {
    pushToast('error', `读取配置失败:${asRelayError(raw).message}`);
  } finally {
    settingsLoading.value = false;
  }
}

/** 保存一组:前端校验 → `app.save_config` → 用新配置刷新草稿 → toast(需要重启时提示)。返回是否成功。 */
export async function saveGroup(group: Group): Promise<boolean> {
  const d = draft.value;
  if (!d || saving.value) return false;
  const errs = validateSettings(numbersFor(group, d));
  if (Object.keys(errs).length > 0) {
    // 表单下方已有 inline 红字,这里只给一个会自动消失的提示,不常驻
    pushToast('info', Object.values(errs)[0] ?? '取值不合法');
    return false;
  }
  saving.value = group;
  try {
    const r = await api.appSaveConfig(patchFor(group, d));
    const c = await api.appGetConfig();
    setAppConfig(c);
    draft.value = mergeSavedGroup(group, draft.value, c);
    pushToast(r.needsRestart ? 'info' : 'success', saveResultMessage(r.needsRestart));
    return true;
  } catch (raw) {
    const e = asRelayError(raw);
    const reason = (e.details as { reason?: string } | undefined)?.reason;
    pushToast('error', `保存失败:${e.message}${reason ? `(${reason})` : ''}`);
    return false;
  } finally {
    saving.value = null;
  }
}

let clearTimer: ReturnType<typeof setTimeout> | null = null;

/** 60s 到点:剪贴板里仍是这个 token 才清空。导出便于单测。 */
export async function clearClipboardIfStillToken(token: string): Promise<boolean> {
  let current: string | null = null;
  try {
    current = await api.clipboardRead();
  } catch {
    current = null;
  }
  if (!shouldClearClipboard(current, token)) return false;
  try {
    await api.clipboardWrite('');
    return true;
  } catch {
    return false;
  }
}

export async function copyToken(token: string): Promise<boolean> {
  if (!token) return false;
  try {
    await api.clipboardWrite(token);
  } catch (raw) {
    pushToast('error', `复制失败:${asRelayError(raw).message}`);
    return false;
  }
  pushToast('success', '已复制 token,60 秒后会自动从剪贴板清除');
  if (clearTimer) clearTimeout(clearTimer);
  clearTimer = setTimeout(() => {
    clearTimer = null;
    void clearClipboardIfStillToken(token);
  }, CLIPBOARD_CLEAR_MS);
  return true;
}

export async function regenerateToken(): Promise<boolean> {
  if (regenerating.value) return false;
  const ok = await confirm({
    title: '重新生成 Bearer Token',
    message: '旧 token 会立刻失效,所有使用旧 token 的调用方都需要更新。确定重新生成吗?',
    confirmLabel: '重新生成',
  });
  if (!ok || regenerating.value) return false;
  regenerating.value = true;
  try {
    const r = await api.appRegenerateToken();
    const c = appConfig.value;
    // 只换 token:草稿里各组未保存的编辑不动
    if (c) appConfig.value = { ...c, server: { ...c.server, bearerToken: r.bearerToken } };
    tokenVisible.value = true;
    pushToast('success', '已重新生成 token,旧 token 已失效');
    return true;
  } catch (raw) {
    pushToast('error', `重新生成失败:${asRelayError(raw).message}`);
    return false;
  } finally {
    regenerating.value = false;
  }
}

export async function runGc(): Promise<boolean> {
  if (cleaning.value) return false;
  cleaning.value = true;
  try {
    const r = await api.maintenanceGc();
    pushToast('success', gcSummary(r));
    return true;
  } catch (raw) {
    pushToast('error', `清理失败:${asRelayError(raw).message}`);
    return false;
  } finally {
    cleaning.value = false;
  }
}

function setDraft<K extends keyof Draft>(key: K, value: Draft[K]) {
  const d = draft.value;
  if (d) draft.value = { ...d, [key]: value };
}

function numberInput(ev: JSX.TargetedEvent<HTMLInputElement>): number {
  const v = (ev.currentTarget as HTMLInputElement).value.trim();
  return v === '' ? Number.NaN : Number(v);
}

function Section({
  title,
  group,
  children,
  extra,
}: {
  title: string;
  group?: Group;
  children: ComponentChildren;
  extra?: ComponentChildren;
}) {
  return (
    <section class="settings-group" data-group={group ?? title}>
      <h2 class="settings-group-title">{title}</h2>
      <div class="settings-rows">{children}</div>
      {(group || extra) && (
        <div class="settings-actions">
          {extra}
          {group && (
            <button
              type="button"
              class="btn btn-primary"
              aria-label={`保存${title}`}
              disabled={saving.value !== null}
              onClick={() => void saveGroup(group)}
            >
              {saving.value === group ? '保存中…' : '保存'}
            </button>
          )}
        </div>
      )}
    </section>
  );
}

function Row({ label, hint, error, children }: { label: string; hint?: string; error?: string; children: ComponentChildren }) {
  return (
    <div class="settings-row" data-error={error ? true : undefined}>
      <div class="settings-label">{label}</div>
      <div class="settings-control">
        {children}
        {hint && <div class="settings-hint">{hint}</div>}
        {error && <div class="settings-error">{error}</div>}
      </div>
    </div>
  );
}

export function SettingsPage() {
  useEffect(() => {
    if (!draft.value) void loadSettings();
  }, []);
  const c = appConfig.value;
  const d = draft.value;
  if (!c || !d) {
    return (
      <section class="page page-settings">
        <header class="page-header">
          <h1 class="page-title">设置</h1>
        </header>
        <div class="skeleton" aria-busy="true" aria-label="加载中">
          <div class="skeleton-line w-40" />
          <div class="skeleton-line" />
          <div class="skeleton-line w-70" />
        </div>
      </section>
    );
  }
  const errs = validateSettings({
    port: d.port,
    globalMax: d.globalMax,
    perAccountMax: d.perAccountMax,
    cdpPort: d.cdpPort,
    resultRetentionHours: d.resultRetentionHours,
    metadataRetentionDays: d.metadataRetentionDays,
  });
  const token = c.server.bearerToken;
  const v = version.value;
  const localApi = apiStatus.value;

  const onVersionClick = () => {
    const result = registerVersionClick();
    if (result.unlocked) pushToast('success', '开发者模式已开启');
    else if (result.remaining <= 4) pushToast('info', `再点击 ${result.remaining} 次开启开发者模式`);
  };

  return (
    <section class="page page-settings">
      <header class="page-header">
        <h1 class="page-title">设置</h1>
      </header>

      <Section title="连接 V88">
        <Row label="连接状态">
          <div class={`connection-state ${c.v88Executor.executorId ? 'is-connected' : ''}`}>
            <span class="connection-dot" />
            {c.v88Executor.executorId ? '已配对，正在等待网页任务' : '尚未配对'}
          </div>
        </Row>
        <Row label="V88 地址" hint="填写你使用的一战晟铭 V88 网站地址">
          <input class="input" value={v88BaseUrl.value} onInput={(ev) => (v88BaseUrl.value = (ev.currentTarget as HTMLInputElement).value)} />
        </Row>
        <Row label="配对码" hint="在 V88 设置页生成，10 分钟内有效且只能使用一次">
          <div class="token-row">
            <input class="input" value={v88PairCode.value} onInput={(ev) => (v88PairCode.value = (ev.currentTarget as HTMLInputElement).value)} />
            <button type="button" class="btn btn-primary" disabled={v88Pairing.value} onClick={() => void pairV88()}>
              {v88Pairing.value ? '连接中…' : '连接'}
            </button>
          </div>
        </Row>
        {c.v88Executor.executorId && <Row label="设备编号"><code>{c.v88Executor.executorId}</code></Row>}
      </Section>

      {developerMode.value && <Section title="开发者 HTTP API" group="server">
        <Row label="启用本机接口" hint="仅供需要从本机程序直接调用时使用，修改后需重启执行器">
          <label class="checkbox">
            <input
              type="checkbox"
              checked={d.serverEnabled}
              onChange={(ev) => setDraft('serverEnabled', (ev.currentTarget as HTMLInputElement).checked)}
            />
            开启 HTTP API
          </label>
        </Row>
        <Row label="当前状态">
          <div class={`api-state ${localApi?.listening ? 'is-listening' : localApi?.error ? 'has-error' : ''}`}>
            {localApi?.listening
              ? `监听中 ${localApi.host}:${localApi.port}`
              : localApi?.error
                ? `启动失败：${localApi.error}`
                : d.serverEnabled
                  ? '等待重启后启动'
                  : '已关闭'}
          </div>
        </Row>
        <Row label="监听端口" error={errs.port} hint="改动后需要重启一战晟铭执行器才生效">
          <div class="token-row">
            <input
              class="input input-num"
              type="number"
              min={1}
              max={65535}
              value={Number.isNaN(d.port) ? '' : d.port}
              onInput={(ev) => setDraft('port', numberInput(ev))}
            />
            <button type="button" class="btn btn-small" onClick={() => void chooseAvailablePort()}>自动选择端口</button>
          </div>
        </Row>
        <Row label="Bearer Token" hint="复制后 60 秒会自动从剪贴板清除;重新生成后旧 token 立刻失效">
          <div class="token-row">
            <code class="token" title={tokenVisible.value ? token : undefined}>
              {tokenVisible.value ? token : maskToken(token)}
            </code>
            <button type="button" class="btn btn-small" onClick={() => (tokenVisible.value = !tokenVisible.value)}>
              {tokenVisible.value ? '隐藏' : '显示'}
            </button>
            <button type="button" class="btn btn-small" onClick={() => void copyToken(token)}>
              复制
            </button>
            <button
              type="button"
              class="btn btn-small"
              disabled={regenerating.value}
              onClick={() => void regenerateToken()}
            >
              {regenerating.value ? '生成中…' : '重新生成'}
            </button>
          </div>
        </Row>
        <Row label="调用示例">
          <div class="api-examples">
            {settingsApiExamples(c.server, tokenVisible.value).map((ex) => (
              <div class="api-example" key={ex.title}>
                <div class="api-example-title">
                  {ex.title}
                  {ex.note && <span class="api-example-note">{ex.note}</span>}
                </div>
                <pre class="curl-example">{ex.command}</pre>
              </div>
            ))}
          </div>
        </Row>
      </Section>}

      <Section title="并发" group="concurrency">
        <Row label="全局最大并发" error={errs.globalMax}>
          <input
            class="input input-num"
            type="number"
            min={1}
            max={100}
            value={Number.isNaN(d.globalMax) ? '' : d.globalMax}
            onInput={(ev) => setDraft('globalMax', numberInput(ev))}
          />
        </Row>
        <Row label="每账号并发" error={errs.perAccountMax} hint="固定为 1:同一账号同时只能有一个生成在豆包里进行">
          <input
            class="input input-num"
            type="number"
            readOnly
            disabled
            value={Number.isNaN(d.perAccountMax) ? '' : d.perAccountMax}
          />
        </Row>
      </Section>

      <Section
        title="结果保留"
        group="storage"
        extra={
          <button type="button" class="btn" disabled={cleaning.value} onClick={() => void runGc()}>
            {cleaning.value ? '清理中…' : '立即清理'}
          </button>
        }
      >
        <Row label="结果文件保留" error={errs.resultRetentionHours}>
          <select
            class="select"
            value={String(d.resultRetentionHours)}
            onChange={(ev) => setDraft('resultRetentionHours', Number((ev.currentTarget as HTMLSelectElement).value))}
          >
            {[...new Set([...RESULT_RETENTION_OPTIONS, d.resultRetentionHours])]
              .filter((n) => Number.isFinite(n))
              .sort((a, b) => a - b)
              .map((h) => (
                <option key={h} value={String(h)}>
                  {h >= 24 && h % 24 === 0 ? `${h / 24} 天` : `${h} 小时`}
                </option>
              ))}
          </select>
        </Row>
        <Row label="元数据保留" error={errs.metadataRetentionDays}>
          <select
            class="select"
            value={String(d.metadataRetentionDays)}
            onChange={(ev) => setDraft('metadataRetentionDays', Number((ev.currentTarget as HTMLSelectElement).value))}
          >
            {[...new Set([...METADATA_RETENTION_OPTIONS, d.metadataRetentionDays])]
              .filter((n) => Number.isFinite(n))
              .sort((a, b) => a - b)
              .map((days) => (
                <option key={days} value={String(days)}>
                  {days} 天
                </option>
              ))}
          </select>
        </Row>
      </Section>

      <Section title="日志" group="logging">
        <Row label="日志级别" hint="保存后立即生效">
          <select
            class="select"
            value={d.level}
            onChange={(ev) => setDraft('level', (ev.currentTarget as HTMLSelectElement).value)}
          >
            {LOG_LEVEL_OPTIONS.map((l) => (
              <option key={l} value={l}>
                {l}
              </option>
            ))}
          </select>
        </Row>
        <Row label="开启 CDP 日志" hint="调试豆包改版用,平时关闭;开启后日志页出现“CDP 调试”面板,并写入 cdp-debug.log">
          <label class="checkbox">
            <input
              type="checkbox"
              checked={d.cdpDebugLog}
              onChange={(ev) => setDraft('cdpDebugLog', (ev.currentTarget as HTMLInputElement).checked)}
            />
            记录每条 CDP 收发
          </label>
        </Row>
      </Section>

      <Section title="高级" group="advanced">
        <Row label="豆包 CDP 端口" error={errs.cdpPort} hint="改动后需要重启一战晟铭执行器才生效">
          <input
            class="input input-num"
            type="number"
            min={1}
            max={65535}
            value={Number.isNaN(d.cdpPort) ? '' : d.cdpPort}
            onInput={(ev) => setDraft('cdpPort', numberInput(ev))}
          />
        </Row>
        <Row label="启动时自动连豆包">
          <label class="checkbox">
            <input
              type="checkbox"
              checked={d.autoConnectOnStartup}
              onChange={(ev) => setDraft('autoConnectOnStartup', (ev.currentTarget as HTMLInputElement).checked)}
            />
            启动一战晟铭执行器时自动启动并连接豆包
          </label>
        </Row>
        <Row label="关闭时最小化到托盘">
          <label class="checkbox">
            <input
              type="checkbox"
              checked={d.minimizeToTrayOnClose}
              onChange={(ev) => setDraft('minimizeToTrayOnClose', (ev.currentTarget as HTMLInputElement).checked)}
            />
            点击关闭按钮时隐藏到托盘而不是退出
          </label>
        </Row>
      </Section>

      <Section title="关于">
        <Row label="版本">
          <button type="button" class="about-version" onClick={onVersionClick} title="版本信息">
            一战晟铭执行器 v{v?.version ?? '…'}
            {v && (
              <span class="muted">
                {' '}
                · {v.buildTarget} · Tauri {v.tauriVersion} · rustc {v.rustcVersion}
              </span>
            )}
          </button>
        </Row>
        <Row label="">
          <div class="about-links">
            <button type="button" class="btn btn-small" disabled title="后续版本提供">
              查看开源协议
            </button>
            <button type="button" class="btn btn-small" disabled title="后续版本提供(需用户主动开启)">
              检查更新
            </button>
            <button type="button" class="btn btn-small" disabled title="后续版本提供">
              帮助文档
            </button>
          </div>
        </Row>
      </Section>
    </section>
  );
}
