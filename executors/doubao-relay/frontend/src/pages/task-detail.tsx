// 任务详情(docs/spec/06-webview-ui.md 页面 3):基本信息、live 时间轴、结果缩略图 / 视频、
// 取消 / 手动确认 / 打开目录、原始请求折叠。live 更新:app.tsx 收到本任务的 task_updated
// 就 `refreshTaskDetail`,收到 task_completed 直接用 payload。
import { useEffect } from 'preact/hooks';
import * as api from '../api';
import { asRelayError } from '../api';
import type { ResultFileView, TaskDetail } from '../api';
import { confirm } from '../components/confirm-dialog';
import {
  pushToast,
  selectedTaskId,
  taskAction,
  taskDetail,
  taskDetailError,
} from '../state';
import {
  canCancel,
  canConfirm,
  formatBytes,
  statusIcon,
  taskErrorHint,
  taskTypeText,
  timelineRows,
} from '../view-model';

let detailSeq = 0;

/** 读详情;切到别的任务后旧请求结果丢弃。 */
export async function refreshTaskDetail(id: string) {
  const seq = ++detailSeq;
  try {
    const t = await api.taskGet(id);
    if (seq !== detailSeq || selectedTaskId.value !== id) return;
    taskDetail.value = t;
    taskDetailError.value = null;
  } catch (raw) {
    if (seq !== detailSeq || selectedTaskId.value !== id) return;
    const e = asRelayError(raw);
    taskDetailError.value = e.message;
    if (e.code !== 'task_not_found') pushToast('error', `读取任务详情失败:${e.message}`);
  }
}

/** task_completed payload(完整任务)→ 直接替换详情。 */
export function applyCompletedDetail(t: TaskDetail) {
  if (selectedTaskId.value === t.taskId) {
    taskDetail.value = t;
    taskDetailError.value = null;
  }
}

export function closeTaskDetail() {
  selectedTaskId.value = null;
  taskDetail.value = null;
  taskDetailError.value = null;
}

export async function cancelTask(id: string) {
  if (taskAction.value) return;
  const ok = await confirm({
    title: '取消任务',
    message: '确定取消这个任务吗?已发给豆包的生成不会被撤回,但 relay 不再等待结果。',
    confirmLabel: '取消任务',
    danger: true,
  });
  if (!ok || taskAction.value) return;
  taskAction.value = 'cancel';
  try {
    await api.taskCancel(id);
    pushToast('success', '任务已取消');
  } catch (raw) {
    pushToast('error', `取消失败:${asRelayError(raw).message}`);
  } finally {
    taskAction.value = null;
    void refreshTaskDetail(id);
  }
}

export async function confirmTask(id: string) {
  if (taskAction.value) return;
  taskAction.value = 'confirm';
  try {
    await api.taskConfirm(id);
    pushToast('success', '已确认,豆包继续生成');
  } catch (raw) {
    pushToast('error', `确认失败:${asRelayError(raw).message}`);
  } finally {
    taskAction.value = null;
    void refreshTaskDetail(id);
  }
}

export async function openResultFolder(id: string) {
  if (taskAction.value) return;
  taskAction.value = 'open';
  try {
    await api.taskOpenResultFolder(id);
  } catch (raw) {
    pushToast('error', `打开目录失败:${asRelayError(raw).message}`);
  } finally {
    taskAction.value = null;
  }
}

function ResultFile({ f }: { f: ResultFileView }) {
  const src = f.localPath ? api.fileSrc(f.localPath) : null;
  const isVideo = f.mime.startsWith('video/');
  const dims = f.width && f.height ? `${f.width}x${f.height}` : null;
  return (
    <figure class="result-file">
      {src ? (
        isVideo ? (
          <video class="result-media" src={src} controls preload="metadata" />
        ) : (
          <a href={src} target="_blank" rel="noreferrer" title="查看原图">
            <img class="result-media" src={src} alt={f.filename} loading="lazy" />
          </a>
        )
      ) : (
        <div class="result-media result-missing">无法预览</div>
      )}
      <figcaption class="result-caption">
        <span class="mono">{f.filename}</span>
        {dims && <span class="muted">{dims}</span>}
        <span class="muted">{formatBytes(f.bytes)}</span>
      </figcaption>
    </figure>
  );
}

function Info({ t }: { t: TaskDetail }) {
  const refs = t.request.referenceFiles?.length ?? 0;
  const s = statusIcon(t.status);
  const errorHint = taskErrorHint(t.error);
  return (
    <dl class="info-grid">
      <dt>状态</dt>
      <dd>
        <span class={`task-status tone-${s.tone}`}>
          {s.icon}
          {s.label}
        </span>
        {t.queuePosition != null && t.status === 'pending' && (
          <span class="muted"> (队列第 {t.queuePosition})</span>
        )}
        {t.message && <span class="muted"> · {t.message}</span>}
      </dd>
      <dt>账号</dt>
      <dd>{t.account}</dd>
      <dt>类型</dt>
      <dd>{taskTypeText(t)}</dd>
      <dt>Prompt</dt>
      <dd class="prompt selectable">{t.request.prompt}</dd>
      {refs > 0 && (
        <>
          <dt>参考图</dt>
          <dd>{refs} 张</dd>
        </>
      )}
      <dt>确认方式</dt>
      <dd>
        {t.request.confirmationMode === 'manual' ? '手动确认' : '自动确认'}
        {t.request.autoUseMembership === false ? ' · 不自动使用会员配额' : ''}
      </dd>
      {t.error && (
        <>
          <dt>错误</dt>
          <dd class="error-text selectable">
            {t.error.message} <span class="mono muted">({t.error.code})</span>
            {errorHint && <span class="error-detail">具体原因：{errorHint}</span>}
          </dd>
        </>
      )}
    </dl>
  );
}

export function TaskDetailPage({ id }: { id: string }) {
  const t = taskDetail.value?.taskId === id ? taskDetail.value : null;
  const err = taskDetailError.value;
  const busy = taskAction.value;

  useEffect(() => {
    void refreshTaskDetail(id);
  }, [id]);

  const rows = timelineRows(t?.stageHistory);
  const files = t?.result?.files ?? [];

  return (
    <section class="page page-task-detail">
      <header class="page-header">
        <div class="detail-title">
          <button type="button" class="btn btn-small" onClick={closeTaskDetail}>
            ← 返回
          </button>
          <h1 class="page-title">
            任务 <span class="mono selectable">{id}</span>
          </h1>
        </div>
        <div class="page-actions">
          {t && canConfirm(t.status) && (
            <button
              type="button"
              class="btn btn-primary"
              disabled={busy !== null}
              onClick={() => confirmTask(id)}
            >
              {busy === 'confirm' ? '确认中…' : '手动确认'}
            </button>
          )}
          {t && canCancel(t.status) && (
            <button type="button" class="btn" disabled={busy !== null} onClick={() => cancelTask(id)}>
              {busy === 'cancel' ? '取消中…' : '取消'}
            </button>
          )}
        </div>
      </header>

      {!t ? (
        err ? (
          <div class="empty-state" data-kind="error" data-error>
            <p class="empty-title">读取任务失败</p>
            <p class="empty-hint">{err}</p>
            <button type="button" class="btn" onClick={closeTaskDetail}>
              返回列表
            </button>
          </div>
        ) : (
          <div class="skeleton" aria-busy="true" aria-label="加载中">
            <div class="skeleton-line w-40" />
            <div class="skeleton-line" />
            <div class="skeleton-line w-70" />
          </div>
        )
      ) : (
        <>
          <Info t={t} />

          <h2 class="section-title">状态时间轴</h2>
          <ol class="timeline">
            {rows.map((r, i) => (
              <li key={`${r.ts}-${i}`} class={`timeline-row tone-${r.tone}`}>
                <span class="timeline-time mono">{r.time}</span>
                <span class="timeline-icon" aria-hidden="true">
                  {r.icon}
                </span>
                <span class="timeline-label">{r.label}</span>
                {r.message && r.message !== r.label && <span class="timeline-msg muted">{r.message}</span>}
              </li>
            ))}
          </ol>

          {t.status === 'completed' && (
            <>
              <h2 class="section-title">结果</h2>
              {files.length > 0 ? (
                <>
                  <div class="result-grid">
                    {files.map((f) => (
                      <ResultFile key={f.index} f={f} />
                    ))}
                  </div>
                  <p class="result-meta muted">
                    结果文件保留到 {new Date(t.expiresAt).toLocaleString()},之后自动清理。
                    <button
                      type="button"
                      class="btn btn-small"
                      disabled={busy !== null}
                      onClick={() => openResultFolder(id)}
                    >
                      打开目录
                    </button>
                  </p>
                </>
              ) : (
                <p class="muted">结果文件已过期清理。</p>
              )}
            </>
          )}

          <details class="raw-request">
            <summary>原始请求</summary>
            <pre class="selectable">{JSON.stringify(t.request, null, 2)}</pre>
          </details>
        </>
      )}
    </section>
  );
}
