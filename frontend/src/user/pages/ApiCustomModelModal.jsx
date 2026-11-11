// 「自定义模型」统一表单：三组里的新增与编辑都打开这一份。
// source='channel' → 写 apiIntegrations.customChannels；
// source='catalog' → 写模型目录里用户自建的那一条（PATCH /api/config/models/{id}）。
// 同一条数据只有这里一个可写入口，不再有行内零散输入框。
import { Button, Collapse, Input, InputNumber, Modal, Select, Switch, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { createHydrationGuard } from './hydrationGuard';

const API_FORMAT_OPTIONS = [
  { value: 'openai_compatible', label: 'OpenAI Chat Completions 格式' },
  { value: 'custom', label: '自定义 API 格式' }
];

const KIND_LABELS = { text: '文本', video: '视频', image: '图片' };

function emptyDraft(kind) {
  return {
    kind,
    apiFormat: 'openai_compatible',
    fullUrl: false,
    baseUrl: '',
    model: '',
    name: '',
    apiKey: '',
    concurrency: null
  };
}

function draftFromSource(initial, kind, source) {
  if (!initial) return emptyDraft(kind);
  if (source === 'catalog') {
    // 目录模型的字段名不同：displayName / modelId；目录本身没有格式、完整 URL 与并发。
    return {
      ...emptyDraft(initial.kind || kind),
      baseUrl: initial.baseUrl || '',
      model: initial.modelId || '',
      name: initial.displayName || initial.name || initial.id || ''
    };
  }
  return {
    ...emptyDraft(kind),
    kind: initial.group || kind,
    apiFormat: initial.apiFormat || 'openai_compatible',
    fullUrl: initial.fullUrl === true,
    baseUrl: initial.baseUrl || '',
    model: initial.model || '',
    name: initial.name || '',
    concurrency: initial.concurrency ?? null
  };
}

export default function ApiCustomModelModal({
  open,
  mode = 'create',
  kind = 'text',
  source = 'channel',
  initial = null,
  saving = false,
  onCancel,
  onSubmit
}) {
  const [draft, setDraft] = useState(() => emptyDraft(kind));
  const [keyTouched, setKeyTouched] = useState(false);
  const isCatalog = source === 'catalog';
  // 用户改过表单后，父组件的重渲染/异步回包不再整表重填。
  const guardRef = useRef(null);
  if (!guardRef.current) guardRef.current = createHydrationGuard();
  const guard = guardRef.current;

  // 每次打开都按当前条目重填：编辑态预填现有值，新增态是空表单。
  useEffect(() => {
    if (!open) {
      guard.reset();
      return;
    }
    if (!guard.canHydrate()) return;
    setDraft(draftFromSource(initial, kind, source));
    setKeyTouched(false);
  }, [open, kind, initial, source, guard]);

  function patch(next) {
    guard.touch();
    setDraft(current => ({ ...current, ...next }));
  }

  function reset() {
    guard.touch();
    setDraft(draftFromSource(initial, kind, source));
    setKeyTouched(false);
  }

  function submit() {
    const baseUrl = String(draft.baseUrl || '').trim();
    const model = String(draft.model || '').trim();
    const name = String(draft.name || '').trim();
    const apiKey = String(draft.apiKey || '').trim();
    if (!baseUrl) {
      message.warning('请填写自定义请求地址');
      return;
    }
    if (!/^https?:\/\//i.test(baseUrl)) {
      message.warning('请求地址必须以 http:// 或 https:// 开头');
      return;
    }
    if (!model) {
      message.warning('请填写模型 ID');
      return;
    }
    if (!name) {
      message.warning('请填写模型展示名称');
      return;
    }
    // 只有渠道的新增必须填 Key（没有 Key 不会进下拉）；编辑与目录模型都允许留空保持原样。
    if (!isCatalog && mode === 'create' && !apiKey) {
      message.warning('请填写 API 密钥；没有 Key 的模型不会出现在下拉里');
      return;
    }
    onSubmit({
      name,
      baseUrl,
      model,
      concurrency: draft.concurrency ?? null,
      apiFormat: draft.apiFormat,
      fullUrl: draft.fullUrl === true,
      apiKey,
      keyTouched
    });
  }

  return (
    <Modal
      open={open}
      title={mode === 'edit' ? `编辑自定义模型 · ${KIND_LABELS[kind] || kind}` : `添加自定义模型 · ${KIND_LABELS[kind] || kind}`}
      onCancel={onCancel}
      width={620}
      destroyOnClose
      footer={(
        <div className="api-custom-model-footer">
          <span className="api-custom-model-note">连通性测试会发起一次真实请求，会消耗少量模型 Token。</span>
          <div className="api-custom-model-actions">
            <Button onClick={reset}>重置</Button>
            <Button type="primary" loading={saving} onClick={submit}>{mode === 'edit' ? '保存修改' : '添加模型'}</Button>
          </div>
        </div>
      )}
    >
      <div className="api-custom-model-form">
        <label className="api-custom-model-field">
          <span>API 格式</span>
          <Select
            value={draft.apiFormat}
            options={isCatalog ? API_FORMAT_OPTIONS.slice(0, 1) : API_FORMAT_OPTIONS}
            disabled={isCatalog}
            onChange={next => patch({ apiFormat: next })}
          />
          {isCatalog ? <small className="api-custom-model-hint">目录模型固定按 OpenAI Chat Completions 协议调用。</small> : null}
        </label>

        <label className="api-custom-model-field">
          <span className="api-custom-model-label">
            自定义请求地址
            <span className="api-custom-model-inline-switch">
              <Switch size="small" checked={draft.fullUrl} disabled={isCatalog} onChange={checked => patch({ fullUrl: checked })} />
              <span>完整 URL</span>
            </span>
          </span>
          <Input
            value={draft.baseUrl}
            placeholder="https://api.example.com/v1"
            onChange={event => patch({ baseUrl: event.target.value })}
          />
          <small className="api-custom-model-hint">
            {isCatalog
              ? '目录模型地址不要以斜杠结尾，/chat/completions 由后台自动补充，不支持完整 URL 模式。'
              : draft.fullUrl
                ? '已按完整 URL 使用，不会再补充任何路径。'
                : '不要以斜杠结尾，/chat/completions 将会被补充到你填写的地址末尾。'}
          </small>
        </label>

        <label className="api-custom-model-field">
          <span>模型 ID</span>
          <Input value={draft.model} placeholder="例如 gpt-4o-mini" onChange={event => patch({ model: event.target.value })} />
        </label>

        <label className="api-custom-model-field">
          <span>模型展示名称</span>
          <Input
            value={draft.name}
            maxLength={64}
            showCount
            placeholder="下拉里显示的名字"
            onChange={event => patch({ name: event.target.value })}
          />
        </label>

        <label className="api-custom-model-field">
          <span>API 密钥</span>
          <Input.Password
            value={draft.apiKey}
            placeholder={mode === 'edit' ? '已保存；留空不修改，输入新 Key 覆盖' : '输入这个模型的 API Key'}
            onChange={event => { setKeyTouched(true); patch({ apiKey: event.target.value }); }}
          />
          {mode === 'edit' ? <small className="api-custom-model-hint">清空输入框并保存 = 删除这条模型的 Key，等同停用。</small> : null}
        </label>

        {isCatalog ? null : (
          <Collapse
            ghost
            className="api-custom-model-advanced"
            items={[{
              key: 'advanced',
              label: '高级配置',
              children: (
                <label className="api-custom-model-field">
                  <span>并发数</span>
                  <InputNumber min={1} max={100} value={draft.concurrency} placeholder="1-100，留空使用默认" onChange={next => patch({ concurrency: next ?? null })} />
                </label>
              )
            }]}
          />
        )}
      </div>
    </Modal>
  );
}
