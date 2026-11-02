// 新建任务弹窗(spec 06 页面 2"新建"):填 account / type / model / prompt / ratio / duration / 参考图 / 确认方式,
// `task.submit` 成功后跳详情页;字段错误内联红字(spec 06"字段错用 inline 红字"),其它错误 toast。
import { useEffect, useLayoutEffect, useRef, useState } from 'preact/hooks';
import * as api from '../api';
import { asRelayError } from '../api';
import type { ConfirmationMode, GenerateRequest, TaskType, UploadResult } from '../api';
import { openTask } from '../pages/tasks';
import { accountList, newTaskOpen, pushToast, route } from '../state';
import { durationRange, fieldErrorOf, formatBytes, modelsFor, ratiosFor } from '../view-model';
import type { FormField } from '../view-model';

export type NewTaskForm = {
  account: string;
  type: TaskType;
  modelName: string;
  prompt: string;
  ratio: string;
  duration: number;
  confirmationMode: ConfirmationMode;
  autoUseMembership: boolean;
  references: (UploadResult & { name: string })[];
};

export function defaultForm(account = ''): NewTaskForm {
  return {
    account,
    type: 'image',
    modelName: modelsFor('image')[1],
    prompt: '',
    ratio: '1:1',
    duration: 5,
    confirmationMode: 'auto',
    autoUseMembership: true,
    references: [],
  };
}

/** 切换类型时模型 / 比例 / 时长回到该类型的合法值。 */
export function withType(form: NewTaskForm, type: TaskType): NewTaskForm {
  if (form.type === type) return form;
  const models = modelsFor(type);
  const ratios = ratiosFor(type);
  return {
    ...form,
    type,
    modelName: models.includes(form.modelName) ? form.modelName : models[type === 'image' ? 1 : 0],
    ratio: ratios.includes(form.ratio) ? form.ratio : ratios[1],
  };
}

/** 表单 → `task.submit` 请求(prompt 不 trim:原样发给豆包;空白由后端判 missing)。 */
export function buildRequest(form: NewTaskForm): GenerateRequest {
  const req: GenerateRequest = {
    account: form.account.trim(),
    type: form.type,
    modelName: form.modelName,
    prompt: form.prompt,
    ratio: form.ratio,
    confirmationMode: form.confirmationMode,
    autoUseMembership: form.autoUseMembership,
  };
  if (form.type === 'video') req.duration = form.duration;
  if (form.references.length > 0) req.referenceFiles = form.references.map((r) => r.uploadId);
  return req;
}

const basename = (p: string) => p.split(/[\\/]/).pop() ?? p;

function FieldError({ field, errors }: { field: FormField; errors: Partial<Record<FormField, string>> }) {
  const m = errors[field];
  return m ? (
    <p class="field-error" role="alert" data-field={field}>
      {m}
    </p>
  ) : null;
}

export function NewTaskDialog() {
  const open = newTaskOpen.value;
  const current = accountList.value?.current ?? accountList.value?.accounts.find((a) => a.isCurrent)?.name ?? '';
  const [form, setForm] = useState<NewTaskForm>(() => defaultForm(current));
  const [errors, setErrors] = useState<Partial<Record<FormField, string>>>({});
  const [submitting, setSubmitting] = useState(false);
  const [uploading, setUploading] = useState(false);
  const promptRef = useRef<HTMLTextAreaElement>(null);

  // 每次打开:重置表单,默认当前账号
  useEffect(() => {
    if (open) {
      setForm(defaultForm(current));
      setErrors({});
      setSubmitting(false);
    }
  }, [open]);

  useLayoutEffect(() => {
    if (open) promptRef.current?.focus();
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !submitting) newTaskOpen.value = false;
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open, submitting]);

  if (!open) return null;

  const set = <K extends keyof NewTaskForm>(k: K, v: NewTaskForm[K]) => {
    setForm((f) => ({ ...f, [k]: v }));
    setErrors((e) => ({ ...e, [k]: undefined }));
  };
  const accounts = accountList.value?.accounts.filter((a) => !a.isAmbiguous).map((a) => a.name) ?? [];
  const [dmin, dmax] = durationRange(form.modelName);

  async function addReferences() {
    if (uploading) return;
    setUploading(true);
    try {
      const paths = await api.pickImageFiles();
      const added: NewTaskForm['references'] = [];
      for (const p of paths) {
        try {
          const r = await api.uploadAdd(p, 'image');
          added.push({ ...r, name: basename(p) });
        } catch (raw) {
          pushToast('error', `添加参考图“${basename(p)}”失败:${asRelayError(raw).message}`);
        }
      }
      if (added.length > 0) {
        setForm((f) => ({ ...f, references: [...f.references, ...added] }));
        setErrors((e) => ({ ...e, referenceFiles: undefined }));
      }
    } catch (raw) {
      pushToast('error', `打开文件选择框失败:${asRelayError(raw).message}`);
    } finally {
      setUploading(false);
    }
  }

  async function submit(e: Event) {
    e.preventDefault();
    if (submitting) return;
    setSubmitting(true);
    try {
      const r = await api.taskSubmit(buildRequest(form));
      newTaskOpen.value = false;
      pushToast('success', `任务已提交,队列第 ${r.queuePosition} 位`);
      route.value = 'tasks';
      openTask(r.taskId);
    } catch (raw) {
      const err = asRelayError(raw);
      const fe = fieldErrorOf(err);
      if (fe) setErrors({ [fe.field]: fe.message });
      else pushToast('error', `提交失败:${err.message}`);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div
      class="dialog-backdrop"
      onClick={(e) => {
        if (e.target === e.currentTarget && !submitting) newTaskOpen.value = false;
      }}
    >
      <form
        class="dialog dialog-wide new-task-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby="new-task-title"
        onSubmit={submit}
      >
        <h2 id="new-task-title" class="dialog-title">
          新建任务
        </h2>

        <div class="form-grid">
          <label for="nt-account">账号</label>
          <div>
            <input
              id="nt-account"
              class="input"
              list="nt-accounts"
              value={form.account}
              placeholder="豆包账号名"
              onInput={(e) => set('account', (e.currentTarget as HTMLInputElement).value)}
            />
            <datalist id="nt-accounts">
              {accounts.map((a) => (
                <option key={a} value={a} />
              ))}
            </datalist>
            <FieldError field="account" errors={errors} />
          </div>

          <span>类型</span>
          <div class="radio-row" role="radiogroup" aria-label="类型">
            {(['image', 'video'] as const).map((t) => (
              <label key={t} class="radio">
                <input
                  type="radio"
                  name="nt-type"
                  checked={form.type === t}
                  onChange={() => {
                    setForm((f) => withType(f, t));
                    setErrors({});
                  }}
                />
                {t === 'image' ? '图像' : '视频'}
              </label>
            ))}
          </div>

          <label for="nt-model">模型</label>
          <div>
            <select
              id="nt-model"
              class="select"
              value={form.modelName}
              onChange={(e) => set('modelName', (e.currentTarget as HTMLSelectElement).value)}
            >
              {modelsFor(form.type).map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
            <FieldError field="modelName" errors={errors} />
          </div>

          <label for="nt-ratio">比例</label>
          <div>
            <select
              id="nt-ratio"
              class="select"
              value={form.ratio}
              onChange={(e) => set('ratio', (e.currentTarget as HTMLSelectElement).value)}
            >
              {ratiosFor(form.type).map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
            <FieldError field="ratio" errors={errors} />
          </div>

          {form.type === 'video' && (
            <>
              <label for="nt-duration">时长</label>
              <div>
                <input
                  id="nt-duration"
                  class="input input-narrow"
                  type="number"
                  min={dmin}
                  max={dmax}
                  step={1}
                  value={form.duration}
                  onInput={(e) => set('duration', Number((e.currentTarget as HTMLInputElement).value))}
                />{' '}
                <span class="muted">
                  秒({dmin}–{dmax})
                </span>
                <FieldError field="duration" errors={errors} />
              </div>
            </>
          )}

          <label for="nt-prompt">Prompt</label>
          <div>
            <textarea
              id="nt-prompt"
              ref={promptRef}
              class="input textarea"
              rows={4}
              value={form.prompt}
              placeholder="描述想要生成的画面"
              onInput={(e) => set('prompt', (e.currentTarget as HTMLTextAreaElement).value)}
            />
            <FieldError field="prompt" errors={errors} />
          </div>

          {form.type === 'image' && (
            <>
              <span>参考图</span>
              <div>
                <div class="ref-list">
                  {form.references.map((r, i) => (
                    <span key={r.uploadId} class="chip" title={`${r.mime} · ${formatBytes(r.bytes)}`}>
                      {r.name}
                      <button
                        type="button"
                        class="chip-remove"
                        aria-label={`移除 ${r.name}`}
                        onClick={() =>
                          setForm((f) => ({ ...f, references: f.references.filter((_, j) => j !== i) }))
                        }
                      >
                        ×
                      </button>
                    </span>
                  ))}
                  <button type="button" class="btn btn-small" disabled={uploading} onClick={addReferences}>
                    {uploading ? '添加中…' : '添加参考图'}
                  </button>
                </div>
                <FieldError field="referenceFiles" errors={errors} />
              </div>
            </>
          )}

          <span>确认方式</span>
          <div class="radio-row">
            <label class="radio">
              <input
                type="radio"
                name="nt-confirm"
                checked={form.confirmationMode === 'auto'}
                onChange={() => set('confirmationMode', 'auto')}
              />
              自动
            </label>
            <label class="radio">
              <input
                type="radio"
                name="nt-confirm"
                checked={form.confirmationMode === 'manual'}
                onChange={() => set('confirmationMode', 'manual')}
              />
              手动
            </label>
            <label class="radio">
              <input
                type="checkbox"
                checked={form.autoUseMembership}
                onChange={(e) => set('autoUseMembership', (e.currentTarget as HTMLInputElement).checked)}
              />
              自动使用会员配额
            </label>
          </div>
        </div>

        <div class="dialog-actions">
          <button
            type="button"
            class="btn"
            disabled={submitting}
            onClick={() => (newTaskOpen.value = false)}
          >
            取消
          </button>
          <button type="submit" class="btn btn-primary" disabled={submitting || uploading}>
            {submitting ? (
              <>
                <span class="spinner" aria-hidden="true" />
                &nbsp;提交中…
              </>
            ) : (
              '提交'
            )}
          </button>
        </div>
      </form>
    </div>
  );
}
