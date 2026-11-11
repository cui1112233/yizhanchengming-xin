import { Button, Input, InputNumber, Select, Skeleton, Switch, message } from 'antd';
import { ChevronDown, Eye, EyeOff, HelpCircle, KeyRound, Plus, ShieldCheck, Trash2 } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import ApiCustomModelModal from './ApiCustomModelModal';
import ApiPresetSelectionModal from './ApiPresetSelectionModal';
import { API_CHANNEL_PRESETS, API_INTEGRATION_GROUPS } from './apiIntegrationSchema';
import { buildPresetSelectionPayload, filterPresetsBySelection, selectedPresetIds } from './apiPresetSelection';
import { createRevisionTracker, sameText } from './hydrationGuard';
import { isCatalogModelAvailable, isPlatformManagedModel, isUserManagedCatalogModel, mergeGroupEntries, modelIdentity } from './apiConfigCatalogMerge';
import { getApiIntegrations, revealApiIntegration, saveApiIntegrations } from '../../shared/api/apiIntegrations';
import { invalidateAvailableModelCatalog } from '../../shared/api/modelCatalog';
import '../../shared/styles/api-integrations.css';

// 输入类字段停止输入后自动提交；失焦会立即提交。
const AUTOSAVE_DELAY = 800;

const SAVE_STATUS_TEXT = { idle: '修改后自动保存', saving: '保存中…', saved: '已保存', error: '保存失败' };

function channelPayload(list) {
  return list.map(channel => ({
    ...(channel.id ? { id: channel.id } : {}),
    name: channel.name,
    group: channel.group || '',
    baseUrl: channel.baseUrl,
    model: channel.model,
    concurrency: channel.concurrency ?? null,
    apiFormat: channel.apiFormat || 'openai_compatible',
    fullUrl: channel.fullUrl === true,
    ...(channel.apiKey ? { apiKey: channel.apiKey } : {}),
    // 显式清空 Key 才发这个标记，避免部分保存（例如只改名字）误删已存 Key。
    ...(channel.clearApiKey ? { clearApiKey: true } : {})
  }));
}

function isSecret(field) {
  return field.control === 'password';
}

function SecretControl({ field, value, configured, revealed, onReveal, onHide, onChange, onCommit, onFocusChange }) {
  const [typingVisible, setTypingVisible] = useState(false);
  const name = field.label || field.placeholder;

  if (configured && !revealed) {
    return (
      <div className="api-secret-masked">
        <span className="api-secret-dots" aria-hidden="true">••••••••••••</span>
        <span className="api-secret-tag">已配置</span>
        <button type="button" className="api-secret-eye" aria-label={`查看${name}明文`} title="点击查看明文（单字段取回）" onClick={onReveal}>
          <Eye size={15} aria-hidden="true" />
        </button>
      </div>
    );
  }

  // 密钥类字段不做停输防抖保存：只在失焦或显式保存时提交，避免半截 Key 落库。
  return (
    <Input
      type={typingVisible || revealed ? 'text' : 'password'}
      value={value || ''}
      placeholder={field.placeholder}
      onChange={event => onChange(event.target.value, { autoSave: false })}
      onFocus={() => onFocusChange?.(true)}
      onBlur={() => {
        onFocusChange?.(false);
        onCommit?.();
      }}
      suffix={(
        <button
          type="button"
          className="api-secret-eye"
          aria-label={typingVisible || revealed ? `隐藏${name}` : `显示${name}`}
          title={typingVisible || revealed ? '隐藏' : '显示'}
          onClick={() => {
            if (revealed) {
              onHide();
              return;
            }
            setTypingVisible(current => !current);
          }}
        >
          {typingVisible || revealed ? <EyeOff size={15} aria-hidden="true" /> : <Eye size={15} aria-hidden="true" />}
        </button>
      )}
    />
  );
}

function FieldLabel({ field }) {
  return (
    <div className="api-field-label">
      {field.label ? <span>{field.label}</span> : null}
      {field.help ? <HelpCircle size={13} className="api-field-help" aria-label="有字段说明" /> : null}
      {field.applyUrl ? (
        <a className="api-field-apply" href={field.applyUrl} target="_blank" rel="noopener noreferrer">申请地址</a>
      ) : null}
    </div>
  );
}

function FieldControl({ field, value, configured, revealed, onReveal, onHide, onChange, onCommit, onFocusChange }) {
  if (isSecret(field)) {
    return (
      <SecretControl
        field={field}
        value={value}
        configured={configured}
        revealed={revealed}
        onReveal={onReveal}
        onHide={onHide}
        onChange={onChange}
        onCommit={onCommit}
        onFocusChange={onFocusChange}
      />
    );
  }
  if (field.control === 'number') {
    return (
      <InputNumber
        className="api-number-input"
        min={field.min}
        max={field.max}
        step={1}
        value={value === '' ? null : value}
        placeholder={field.placeholder}
        aria-label={field.label}
        onBlur={onCommit}
        onChange={next => onChange(next === null ? '' : next)}
      />
    );
  }
  if (field.control === 'select') {
    return (
      <Select
        allowClear
        value={value}
        options={field.options || []}
        placeholder={field.placeholder}
        aria-label={field.label}
        onChange={next => { onChange(next); onCommit?.(); }}
      />
    );
  }
  if (field.control === 'switch') {
    return <Switch checked={value !== false} disabled={field.locked === true} aria-label={field.label} onChange={checked => { onChange(checked); onCommit?.(); }} />;
  }
  return (
    <Input
      value={value || ''}
      placeholder={field.placeholder}
      aria-label={field.label}
      onChange={event => onChange(event.target.value)}
      onBlur={onCommit}
    />
  );
}

function ViduAccounts() {
  return (
    <div className="api-vidu">
      <p>最多可添加 1 个账号，勾选后新任务会实时使用当前账号。</p>
      <div className="api-vidu-card">
        <strong>当前没有可用的 Vidu 账号</strong>
        <span>账号验证码添加已禁用，请使用上方「开启浏览器」模式。</span>
      </div>
      <Button disabled>Vidu 账号添加已禁用</Button>
    </div>
  );
}

// channelSelection 是旧版「勾选启用」留下的数据。现在只用来修老数据：
// group 为空的渠道，按它出现在哪一组的勾选列表里补回所属分组。
function legacyChannelGroup(channelId, channelSelection) {
  if (!channelId) return '';
  for (const [groupId, raw] of Object.entries(channelSelection || {})) {
    const list = Array.isArray(raw?.enabled)
      ? raw.enabled
      : (typeof raw === 'string' && raw ? [raw] : []);
    if (list.includes(channelId)) return groupId;
  }
  return '';
}

export default function ApiIntegrationSettings({
  visibleGroupIds = null,
  models = [],
  quotas = [],
  platformPresetIds = [],
  onSaveCatalogCredential = null,
  onUpdateCatalogModel = null,
  catalogModelSavingId = ''
}) {
  const [loading, setLoading] = useState(true);
  const [settings, setSettings] = useState({ values: {}, configured: {}, pinned: [], initialized: false, customChannels: [] });
  const [edits, setEdits] = useState({});
  const [secrets, setSecrets] = useState({});
  const [channels, setChannels] = useState([]);
  const [collapsed, setCollapsed] = useState({});
  const [saveStatus, setSaveStatus] = useState('idle');
  const [customModal, setCustomModal] = useState(null);
  const [customModalSaving, setCustomModalSaving] = useState(false);
  const [presetModal, setPresetModal] = useState(null);
  const [presetSaving, setPresetSaving] = useState(false);
  const [catalogKeyDrafts, setCatalogKeyDrafts] = useState({});
  const [expandedCatalog, setExpandedCatalog] = useState({});

  // 自动保存：串行队列 + 防抖定时器 + 最新值引用（定时器里不能用旧闭包）。
  const queueRef = useRef(Promise.resolve());
  const timersRef = useRef({});
  const editsRef = useRef({});
  const channelsRef = useRef([]);
  const channelTimerRef = useRef(null);
  // 渠道整表保存的版本号：请求飞行期间用户又改过渠道，旧回包直接丢弃，避免把刚删的渠道复活。
  const channelsTrackerRef = useRef(null);
  if (!channelsTrackerRef.current) channelsTrackerRef.current = createRevisionTracker();
  const channelsTracker = channelsTrackerRef.current;
  // 只有“输入过又清空”才算清除 Key，避免把未改动的已存 Key 误删。
  const catalogKeyTouchedRef = useRef(new Set());
  // 正在聚焦输入的密钥字段：保存回包先不清掩码，避免把用户还在输的内容冲掉。
  const focusedSecretRef = useRef(new Set());

  useEffect(() => { channelsRef.current = channels; }, [channels]);

  // 所有本地渠道变更都走这里：递增版本号，异步回包才知道自己是否过期。
  function applyLocalChannels(next) {
    channelsTracker.bump();
    channelsRef.current = next;
    setChannels(next);
  }

  const groups = useMemo(
    () => (visibleGroupIds ? API_INTEGRATION_GROUPS.filter(group => visibleGroupIds.includes(group.id)) : API_INTEGRATION_GROUPS),
    [visibleGroupIds]
  );

  const platformPresetIdSet = useMemo(() => new Set(platformPresetIds), [platformPresetIds]);

  // 模型目录条目按 kind 归到文本/视频/图片三组，只有「可用」的才进入页面。
  // 可用性由服务端 available 字段给出，与 /api/models 下拉完全同一口径。
  // 平台托管的模型（预设面板里的条目、本地执行器）在上方面板或引擎设置里维护，
  // 这里按归属过滤，不写死 id 清单；以后预设面板新增条目会自动去重。
  const catalogByGroup = useMemo(() => {
    const map = { text: [], video: [], image: [] };
    (Array.isArray(models) ? models : []).forEach(model => {
      const kind = model?.kind;
      if (!map[kind]) return;
      if (isPlatformManagedModel(model, platformPresetIdSet)) return;
      if (isCatalogModelAvailable(model)) map[kind].push(model);
    });
    return map;
  }, [models, platformPresetIdSet]);

  const quotaByModelId = useMemo(
    () => new Map((Array.isArray(quotas) ? quotas : []).map(quota => [quota?.modelId, quota])),
    [quotas]
  );

  useEffect(() => {
    let alive = true;
    getApiIntegrations()
      .then(next => {
        if (!alive) return;
        setSettings(next || {});
        // 老数据平滑迁移：旧版把未分组渠道放在 channelSelection 里，
        // 这里按勾选列表补回 group，并落库一次，之后不再依赖 channelSelection。
        const repaired = (Array.isArray(next?.customChannels) ? next.customChannels : []).map(channel => {
          if (channel?.group) return channel;
          const group = legacyChannelGroup(channel?.id, next?.channelSelection);
          return group ? { ...channel, group } : channel;
        });
        setChannels(repaired);
        if (repaired.some((channel, index) => channel.group !== (next?.customChannels || [])[index]?.group)) {
          channelsRef.current = repaired;
          channelTimerRef.current = setTimeout(() => {
            channelTimerRef.current = null;
            saveChannels();
          }, AUTOSAVE_DELAY);
        }
      })
      .catch(error => message.error(error.message || 'API 接口配置加载失败'))
      .finally(() => { if (alive) setLoading(false); });
    return () => {
      alive = false;
      if (channelTimerRef.current) clearTimeout(channelTimerRef.current);
    };
  }, []);

  function hasValue(key) {
    if (edits[key] !== undefined) return String(edits[key]).length > 0;
    if (secrets[key] !== undefined) return true;
    return settings.configured?.[key] === true;
  }

  function setValue(key, next, { autoSave = true } = {}) {
    setEdits(current => {
      const updated = { ...current, [key]: next };
      editsRef.current = updated;
      return updated;
    });
    if (autoSave) {
      scheduleFieldSave(key);
      return;
    }
    // 密钥字段只等失焦提交；这里清掉可能残留的防抖定时器，避免输入中途被保存。
    if (timersRef.current[key]) {
      clearTimeout(timersRef.current[key]);
      delete timersRef.current[key];
    }
  }

  function handleSecretFocus(key, focused) {
    if (focused) {
      focusedSecretRef.current.add(key);
      return;
    }
    focusedSecretRef.current.delete(key);
    if (editsRef.current[key] !== undefined) return;
    setSecrets(current => {
      if (current[key] === undefined) return current;
      const next = { ...current };
      delete next[key];
      return next;
    });
  }

  // 所有保存走同一条串行队列：同一时刻只有一个请求在飞，
  // 后提交的一定后落库，避免“先发的旧值覆盖后发的新值”。
  function enqueue(task) {
    queueRef.current = queueRef.current.then(task).catch(error => {
      setSaveStatus('error');
      message.error(error.message || '保存失败');
    });
    return queueRef.current;
  }

  function saveFields(keys) {
    const values = {};
    keys.forEach(key => { if (editsRef.current[key] !== undefined) values[key] = editsRef.current[key]; });
    if (Object.keys(values).length === 0) return;
    const snapshot = { ...values };
    enqueue(async () => {
      setSaveStatus('saving');
      // 只提交本次改动的字段；服务端按字段合并，不会冲掉同组其他值。
      const next = await saveApiIntegrations({ values: snapshot });
      setSettings(current => ({
        ...current,
        values: next?.values || current.values,
        configured: next?.configured || current.configured
      }));
      setEdits(current => {
        const copy = { ...current };
        Object.entries(snapshot).forEach(([key, value]) => {
          if (String(copy[key] ?? '') === String(value ?? '')) delete copy[key];
        });
        editsRef.current = copy;
        return copy;
      });
      // 保存成功后回到掩码态，避免明文一直留在输入框里；仍在聚焦的字段先保留，失焦后再收。
      setSecrets(current => {
        const copy = { ...current };
        Object.keys(snapshot).forEach(key => {
          if (!focusedSecretRef.current.has(key)) delete copy[key];
        });
        return copy;
      });
      invalidateAvailableModelCatalog();
      setSaveStatus('saved');
    });
  }

  function scheduleFieldSave(key, immediate = false) {
    if (timersRef.current[key]) {
      clearTimeout(timersRef.current[key]);
      delete timersRef.current[key];
    }
    if (immediate) { saveFields([key]); return; }
    timersRef.current[key] = setTimeout(() => {
      delete timersRef.current[key];
      saveFields([key]);
    }, AUTOSAVE_DELAY);
  }

  function flushField(key) {
    if (timersRef.current[key]) {
      clearTimeout(timersRef.current[key]);
      delete timersRef.current[key];
    }
    if (editsRef.current[key] === undefined) return;
    saveFields([key]);
  }

  function saveChannels() {
    if (channelTimerRef.current) {
      clearTimeout(channelTimerRef.current);
      channelTimerRef.current = null;
    }
    const list = channelsRef.current;
    const revision = channelsTracker.current();
    enqueue(async () => {
      setSaveStatus('saving');
      // customChannels 是整表提交；服务端只在显式传该键时才替换，不影响 values。
      const next = await saveApiIntegrations({ customChannels: channelPayload(list) });
      // 请求飞行期间用户又改过渠道：这份回包已经过期，不能拿它覆盖本地新状态。
      if (channelsTracker.isCurrent(revision)) {
        const nextChannels = Array.isArray(next?.customChannels) ? next.customChannels : list;
        channelsRef.current = nextChannels;
        setChannels(nextChannels);
      }
      invalidateAvailableModelCatalog();
      setSaveStatus('saved');
    });
  }

  function scheduleChannelSave() {
    if (channelTimerRef.current) clearTimeout(channelTimerRef.current);
    channelTimerRef.current = setTimeout(() => {
      channelTimerRef.current = null;
      saveChannels();
    }, AUTOSAVE_DELAY);
  }

  function flushChannelSave() {
    if (channelTimerRef.current) {
      clearTimeout(channelTimerRef.current);
      channelTimerRef.current = null;
    }
    saveChannels();
  }

  function groupChannels(groupId) {
    return channels.filter(channel => channel.group === groupId);
  }

  async function reveal(field) {
    try {
      const result = await revealApiIntegration(field.key);
      setSecrets(current => ({ ...current, [field.key]: result?.value || '' }));
    } catch (error) {
      message.error(error.message || '读取密钥失败');
    }
  }

  function fieldValue(field) {
    if (edits[field.key] !== undefined) return edits[field.key];
    if (secrets[field.key] !== undefined) return secrets[field.key];
    return settings.values?.[field.key];
  }

  // 自定义模型的新增与编辑都打开同一份表单（ApiCustomModelModal）。
  // 目录模型同样走这份表单，只是写回它自己的真实位置。
  function openCustomModel(groupId, channel = null) {
    setCustomModal({ groupId, channel });
  }

  function openCatalogModel(groupId, catalogModel) {
    setCustomModal({ groupId, catalogModel });
  }

  // 「自定义预设」只保存“展示哪些预置模型”，写入服务端 channelSelection；
  // 不写 Key、不改地址，也不影响填了 Key 的模型进入下拉。
  function savePresetSelection(groupId, ids) {
    const nextSelection = buildPresetSelectionPayload(settings.channelSelection, groupId, ids);
    setPresetSaving(true);
    enqueue(async () => {
      setSaveStatus('saving');
      try {
        const next = await saveApiIntegrations({ channelSelection: nextSelection });
        setSettings(current => ({
          ...current,
          channelSelection: next?.channelSelection || nextSelection
        }));
        setSaveStatus('saved');
        setPresetModal(null);
      } catch (error) {
        setSaveStatus('error');
        message.error(error?.message || '预置模型选择保存失败');
      } finally {
        setPresetSaving(false);
      }
    });
  }

  function submitCustomModel(values) {
    if (!customModal) return;
    const groupId = customModal.groupId;
    const existing = customModal.channel;
    const catalogModel = customModal.catalogModel;
    if (catalogModel) {
      if (!onUpdateCatalogModel) {
        message.warning('当前账号无权修改目录模型');
        return;
      }
      setCustomModalSaving(true);
      Promise.resolve(onUpdateCatalogModel(catalogModel, values))
        .then(() => setCustomModal(null))
        .catch(() => {})
        .finally(() => setCustomModalSaving(false));
      return;
    }
    if (existing) {
      // 编辑的是渠道那条真实数据；留空 Key 表示保留原 Key。
      const patch = {
        name: values.name,
        baseUrl: values.baseUrl,
        model: values.model,
        concurrency: values.concurrency,
        apiFormat: values.apiFormat,
        fullUrl: values.fullUrl
      };
      if (values.keyTouched && values.apiKey) {
        patch.apiKey = values.apiKey;
        patch.hasApiKey = true;
        patch.clearApiKey = false;
      } else if (values.keyTouched && !values.apiKey && (existing.hasApiKey || existing.apiKey)) {
        patch.apiKey = '';
        patch.hasApiKey = false;
        patch.clearApiKey = true;
      }
      const next = channelsRef.current.map(channel => (channel.id === existing.id ? { ...channel, ...patch } : channel));
      applyLocalChannels(next);
    } else {
      const id = `ch_local_${Math.random().toString(36).slice(2, 10)}`;
      const entry = {
        id,
        group: groupId,
        name: values.name,
        baseUrl: values.baseUrl,
        model: values.model,
        concurrency: values.concurrency,
        apiFormat: values.apiFormat,
        fullUrl: values.fullUrl,
        apiKey: values.apiKey,
        hasApiKey: Boolean(values.apiKey)
      };
      const next = [...channelsRef.current, entry];
      applyLocalChannels(next);
    }
    setCustomModal(null);
    flushChannelSave();
    message.success(existing ? '自定义模型已保存' : '自定义模型已添加');
  }

  function deleteCustomChannel(channel) {
    applyLocalChannels(channelsRef.current.filter(item => item.id !== channel.id));
    scheduleChannelSave();
    message.success(`已删除渠道「${channel.name}」`);
  }

  // 目录模型的凭证在这里编辑；保存后服务端重新计算 available，页面与下拉一起变化。
  async function commitCatalogKey(model) {
    if (!catalogKeyTouchedRef.current.has(model.id)) return;
    const draft = String(catalogKeyDrafts[model.id] || '');
    const value = draft.trim();
    if (!value && model.hasCredential !== true) {
      catalogKeyTouchedRef.current.delete(model.id);
      return;
    }
    if (!onSaveCatalogCredential) {
      message.warning('当前账号无权修改目录模型凭证');
      return;
    }
    try {
      await onSaveCatalogCredential(model, value, !value);
      // 只清掉刚提交的那份草稿；保存期间用户又输入的新内容保留。
      setCatalogKeyDrafts(current => (sameText(current[model.id], draft) ? { ...current, [model.id]: '' } : current));
      catalogKeyTouchedRef.current.delete(model.id);
    } catch {
      // 错误提示已由上层给出，保留输入方便重试。
    }
  }

  function renderPresetChannelCard(entry, group, merged = false) {
    const field = {
      key: entry.fieldKey,
      label: `${entry.label} API Key`,
      control: 'password',
      placeholder: `输入 ${entry.label} API Key`
    };
    return (
      <div className={`api-channel-card${merged ? ' is-merged-channel' : ''}`} key={`${merged ? 'merged-' : ''}${entry.id}`}>
        <div className="api-channel-card-head">
          <strong>{entry.label}</strong>
          <span className="api-channel-tag">预置</span>
          {merged ? <span className="api-channel-tag is-merged">已合并进目录模型</span> : null}
          {entry.pending ? <span className="api-channel-tag is-warn">待核</span> : null}
          {hasValue(entry.fieldKey) ? <span className="api-channel-tag is-on">已配置 Key</span> : null}
        </div>
        {entry.addressField ? (
          <label className="api-channel-preset-line">
            <span>地址</span>
            <Input
              value={fieldValue({ key: entry.addressField }) || ''}
              placeholder={entry.baseUrl || '填写 baseURL'}
              aria-label={`${entry.label}地址`}
              onChange={event => setValue(entry.addressField, event.target.value)}
              onBlur={() => flushField(entry.addressField)}
            />
          </label>
        ) : (
          <p className="api-channel-preset-meta">地址 <code>{entry.baseUrl || '待核实'}</code></p>
        )}
        <p className="api-channel-preset-meta">模型 <code>{entry.model}</code>{entry.pending ? '（地址待核实）' : ''}</p>
        <FieldControl
          field={field}
          value={fieldValue(field)}
          configured={hasValue(entry.fieldKey)}
          revealed={secrets[entry.fieldKey] !== undefined}
          onReveal={() => reveal(field)}
          onHide={() => setSecrets(current => {
            const next = { ...current };
            delete next[entry.fieldKey];
            return next;
          })}
          onChange={(next, options) => setValue(entry.fieldKey, next, options)}
          onCommit={() => flushField(entry.fieldKey)}
          onFocusChange={focused => handleSecretFocus(entry.fieldKey, focused)}
        />
        {entry.extraFields?.length ? (
          <div className="api-channel-custom-row">
            {entry.extraFields.map(extraKey => {
              const extraField = group.fields.find(item => item.key === extraKey);
              if (!extraField) return null;
              return (
                <label className="api-channel-custom-concurrency" key={extraKey}>
                  <span>{extraField.label}</span>
                  <FieldControl
                    field={extraField}
                    value={fieldValue(extraField)}
                    configured={hasValue(extraKey)}
                    revealed={false}
                    onReveal={() => {}}
                    onHide={() => {}}
                    onChange={(next, options) => setValue(extraKey, next, options)}
                    onCommit={() => flushField(extraKey)}
                    onFocusChange={focused => handleSecretFocus(extraKey, focused)}
                  />
                </label>
              );
            })}
          </div>
        ) : null}
      </div>
    );
  }

  // 自定义模型卡只展示摘要；新增/编辑都走同一个表单，卡上不再放零散输入框。
  // 被合并进目录卡时用同一个渲染函数，编辑的就是这条渠道的真实数据。
  function renderCustomChannelCard(entry, merged = false) {
    return (
      <div className={`api-channel-card${merged ? ' is-merged-channel' : ''}`} key={`${merged ? 'merged-' : ''}${entry.id}`}>
        <div className="api-channel-card-head">
          <strong>{entry.name}</strong>
          <span className="api-channel-tag">自定义</span>
          {merged ? <span className="api-channel-tag is-merged">已合并进目录模型</span> : null}
          {entry.hasApiKey || entry.apiKey ? <span className="api-channel-tag is-on">已配置 Key</span> : <span className="api-channel-tag">未填 Key</span>}
          <Button className="api-channel-edit-btn" size="small" type="text" onClick={() => openCustomModel(entry.group, entry)}>编辑</Button>
          <button type="button" className="api-channel-option-delete" aria-label={`删除自定义模型${entry.name}`} title="删除这个自定义模型" onClick={() => deleteCustomChannel(entry)}>
            <Trash2 size={14} aria-hidden="true" />
          </button>
        </div>
        <p className="api-channel-preset-meta">
          {entry.baseUrl || '未填写地址'} · 模型 <code>{entry.model || '未填写'}</code>
          {entry.concurrency ? ` · 并发 ${entry.concurrency}` : ''}
          {entry.fullUrl ? ' · 完整 URL' : ''}
        </p>
      </div>
    );
  }

  // 目录模型卡片：凭证可直接编辑；同源渠道收在卡片里，展开后编辑的是渠道自己的数据。
  function renderCatalogCard(model, mergedChannels, group) {
    const quota = quotaByModelId.get(model.id);
    const hasQuota = quota && quota.status !== 'not_applicable'
      && Number.isFinite(Number(quota.available)) && Number.isFinite(Number(quota.total));
    const quotaText = hasQuota
      ? `${Number(quota.available).toLocaleString('zh-CN', { maximumFractionDigits: 2 })} / ${Number(quota.total).toLocaleString('zh-CN', { maximumFractionDigits: 2 })} ${quota.unit || ''}`.trim()
      : '';
    const expanded = expandedCatalog[model.id] === true;
    const saving = catalogModelSavingId === model.id;
    // 用户自建的目录模型：编辑入口走统一表单，Key 也只在表单里改，
    // 避免同一条数据出现两个可写位置。带适配器的目录模型仍保留行内 Key。
    const userManaged = isUserManagedCatalogModel(model, platformPresetIdSet);
    return (
      <div className="api-channel-card is-catalog" key={`catalog-${model.id}`}>
        <div className="api-channel-card-head">
          <strong>{model.displayName || model.modelId || model.id}</strong>
          <span className="api-channel-tag is-catalog">目录模型</span>
          {quotaText ? <span className="api-channel-tag">额度 {quotaText}</span> : null}
          {model.hasCredential ? <span className="api-channel-tag is-on">已配置 Key</span> : null}
          {userManaged ? (
            <Button className="api-channel-edit-btn" size="small" type="text" onClick={() => openCatalogModel(model.kind, model)}>编辑</Button>
          ) : null}
        </div>
        <p className="api-channel-preset-meta">
          {model.baseUrl ? <>地址 <code>{model.baseUrl}</code></> : null}
          {model.baseUrl && model.modelId ? ' · ' : null}
          {model.modelId ? <>模型 <code>{model.modelId}</code></> : null}
          {!model.baseUrl && !model.modelId ? '该模型使用平台适配器，不需要地址与模型名' : ''}
        </p>
        {model.hasCredential && !userManaged ? (
          <label className="api-channel-preset-line api-catalog-credential">
            <span>目录 Key</span>
            <Input.Password
              value={catalogKeyDrafts[model.id] || ''}
              prefix={<KeyRound size={15} />}
              disabled={saving}
              placeholder={saving ? '保存中…' : '已保存；输入新 Key 覆盖，输入后清空并离开即停用'}
              aria-label={`${model.displayName || model.id} API Key`}
              onChange={event => {
                catalogKeyTouchedRef.current.add(model.id);
                setCatalogKeyDrafts(current => ({ ...current, [model.id]: event.target.value }));
              }}
              onBlur={() => commitCatalogKey(model)}
            />
          </label>
        ) : null}
        {mergedChannels.length ? (
          <div className="api-catalog-merged">
            <button
              type="button"
              className="api-catalog-merged-toggle"
              aria-expanded={expanded}
              onClick={() => setExpandedCatalog(current => ({ ...current, [model.id]: !expanded }))}
            >
              <span>已合并渠道 {mergedChannels.map(channel => channel.name).join('、')}</span>
              <small>展开后编辑的是这些渠道自己的地址 / 模型名 / Key</small>
              <ChevronDown size={14} className={`api-group-chevron${expanded ? ' is-open' : ''}`} aria-hidden="true" />
            </button>
            {expanded ? mergedChannels.map(channel => (
              channel.entry.kind === 'preset'
                ? renderPresetChannelCard(channel.entry, group, true)
                : renderCustomChannelCard(channel.entry, true)
            )) : null}
          </div>
        ) : null}
      </div>
    );
  }

  function buildGroupMerge(group, def, presets = def.presets) {
    const entries = [
      ...presets.map(preset => ({ kind: 'preset', ...preset })),
      ...groupChannels(group.id).map(channel => ({ kind: 'custom', ...channel }))
    ];
    // 目录模型与渠道按「地址 + 模型名」合并；同源渠道不再重复出卡。
    const merged = mergeGroupEntries({
      catalogModels: catalogByGroup[group.id] || [],
      channelEntries: entries.map(entry => (entry.kind === 'preset'
        ? {
          id: entry.id,
          name: entry.label,
          baseUrl: entry.addressField
            ? (String(fieldValue({ key: entry.addressField }) || '').trim() || entry.baseUrl)
            : entry.baseUrl,
          model: entry.model,
          entry
        }
        : {
          id: entry.id,
          name: entry.name,
          baseUrl: entry.baseUrl,
          model: entry.model,
          entry
        }))
    });
    return { entries, merged };
  }

  function renderChannelGroup(group, def) {
    // 预置渠道按「自定义预设」的勾选结果展示；自定义渠道始终展示。
    // 填了 Key 就是可用，清空 Key 就等于停用，不再有启用开关。
    const presets = filterPresetsBySelection(def.presets, settings.channelSelection, group.id);
    const { merged } = buildGroupMerge(group, def, presets);
    const mergedByIdentity = new Map();
    merged.mergedChannels.forEach(channel => {
      const list = mergedByIdentity.get(channel.identity) || [];
      list.push(channel);
      mergedByIdentity.set(channel.identity, list);
    });

    return (
      <div className="api-channel-cards">
        <div className="api-enabled-bar">
          <span className="api-enabled-count">目录模型 {merged.catalog.length} 条 · 渠道 {merged.channels.length} 条</span>
          <span className="api-enabled-actions">
            <Button className="api-channel-open api-preset-open" size="small" onClick={() => setPresetModal(group.id)}>自定义预设</Button>
            <Button className="api-channel-open api-custom-open" size="small" icon={<Plus size={14} aria-hidden="true" />} onClick={() => openCustomModel(group.id)}>添加自定义模型</Button>
          </span>
        </div>
        <p className="api-channel-merge-note">
          填写 Key 后立即生效并进入模型下拉；清空 Key 即从下拉移除，不需要额外开关。
          {merged.mergedChannels.length ? ` 有 ${merged.mergedChannels.length} 条渠道与目录模型同源，已合并进目录卡，展开即可编辑，下拉里只会出现一条。` : ''}
        </p>
        {merged.catalog.map(model => (
          renderCatalogCard(model, mergedByIdentity.get(modelIdentity(model.baseUrl, model.modelId)) || [], group)
        ))}
        {merged.channels.map(({ entry }) => (
          entry.kind === 'preset' ? renderPresetChannelCard(entry, group) : renderCustomChannelCard(entry)
        ))}
      </div>
    );
  }

  const totalFields = useMemo(
    () => groups.flatMap(group => group.fields.filter(field => !field.subheading)).length,
    [groups]
  );
  const configuredTotal = useMemo(
    () => groups.flatMap(group => group.fields.filter(field => !field.subheading && hasValue(field.key))).length,
    [groups, edits, secrets, settings] // eslint-disable-line react-hooks/exhaustive-deps
  );

  if (loading) {
    return (
      <section className="settings-section settings-api-section" id="api-integrations">
        <Skeleton active paragraph={{ rows: 8 }} />
      </section>
    );
  }

  return (
    <section className="settings-section settings-api-section" aria-labelledby="settings-api-title" id="api-integrations">
      <div>
        <h2 id="settings-api-title">API 接口</h2>
        <p>所有模型密钥只保存在服务端；列表只显示“已配置”，点击眼睛单字段取回明文。输入框失焦或停止输入约 0.8 秒自动保存并立即生效，不需要改代码或重启服务。</p>
      </div>
      <div className="api-integration-toolbar">
        <span className="api-integration-count"><ShieldCheck size={14} aria-hidden="true" />已配置 {configuredTotal} / {totalFields}</span>
        <span className="api-integration-hint" role="status" aria-live="polite">{SAVE_STATUS_TEXT[saveStatus]}</span>
        <span className="api-integration-hint">文本 / 视频 / 图片组里同时列出「目录模型」与「渠道」；渠道填了 Key 就可用，清空 Key 就不出现</span>
      </div>
      <div className="api-integration-groups">
        {groups.map(group => {
          const channelDef = API_CHANNEL_PRESETS[group.id];
          // 渠道组默认展开；非渠道组是纯配置项，默认收起。
          const open = collapsed[group.id] !== undefined ? collapsed[group.id] !== true : Boolean(channelDef);
          const groupFields = group.fields.filter(field => !field.subheading);
          const groupConfigured = groupFields.filter(field => hasValue(field.key)).length;
          // 顶部计数与组内计数用同一份合并结果，避免出现“渠道 9 / 实际 8 张卡”。
          const mergedForHeader = channelDef
            ? buildGroupMerge(group, channelDef, filterPresetsBySelection(channelDef.presets, settings.channelSelection, group.id)).merged
            : null;
          const channelCount = mergedForHeader ? mergedForHeader.channels.length : 0;
          const catalogCount = mergedForHeader ? mergedForHeader.catalog.length : 0;
          return (
            <section className="api-group" id={`api-group-${group.id}`} key={group.id}>
              <button
                type="button"
                className="api-group-head"
                aria-expanded={open}
                onClick={() => setCollapsed(current => ({ ...current, [group.id]: open }))}
              >
                <span className={`api-group-badge is-${group.tone}`}>{group.title}</span>
                <span className="api-group-meta">
                  {channelDef ? `目录 ${catalogCount} · 渠道 ${channelCount}` : `已配置 ${groupConfigured} / 共 ${groupFields.length}`}
                </span>
                <ChevronDown size={15} className={`api-group-chevron${open ? ' is-open' : ''}`} aria-hidden="true" />
              </button>
              {open ? (
                <div className="api-group-body">
                  <p className="api-group-description">{group.description}</p>
                  {group.warning ? <p className="api-group-warning">{group.warning}</p> : null}
                  {group.note ? <p className="api-group-note">{group.note}</p> : null}
                  {group.channelNote ? <p className="api-group-note">{group.channelNote}</p> : null}
                  {group.kind === 'vidu-accounts' ? (
                    <>
                      <div className="api-fields api-fields-single">
                        {group.fields.map(field => (
                          <div className="api-field" key={field.key}>
                            <FieldLabel field={field} />
                            <FieldControl
                              field={field}
                              value={fieldValue(field)}
                              configured={hasValue(field.key)}
                              revealed={secrets[field.key] !== undefined}
                              onReveal={() => reveal(field)}
                              onHide={() => setSecrets(current => {
                                const next = { ...current };
                                delete next[field.key];
                                return next;
                              })}
                              onChange={(next, options) => setValue(field.key, next, options)}
                              onCommit={() => flushField(field.key)}
                              onFocusChange={focused => handleSecretFocus(field.key, focused)}
                            />
                          </div>
                        ))}
                      </div>
                      <ViduAccounts />
                    </>
                  ) : channelDef ? (
                    renderChannelGroup(group, channelDef)
                  ) : (
                    <div className="api-fields">
                      {group.fields.map(field => {
                        if (field.subheading) {
                          return <div className="api-subheading" key={field.key}>{field.subheading}</div>;
                        }
                        return (
                          <div className="api-field" key={field.key}>
                            <FieldLabel field={field} />
                            <FieldControl
                              field={field}
                              value={fieldValue(field)}
                              configured={hasValue(field.key)}
                              revealed={secrets[field.key] !== undefined}
                              onReveal={() => reveal(field)}
                              onHide={() => setSecrets(current => {
                                const next = { ...current };
                                delete next[field.key];
                                return next;
                              })}
                              onChange={(next, options) => setValue(field.key, next, options)}
                              onCommit={() => flushField(field.key)}
                              onFocusChange={focused => handleSecretFocus(field.key, focused)}
                            />
                            {field.rangeHint ? <small className="api-field-hint">{field.rangeHint}</small> : null}
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              ) : null}
            </section>
          );
        })}
      </div>
      <div className="settings-section-footer">
        <span className="api-integration-hint">输入类字段自动保存（失焦立即保存，停止输入约 0.8 秒保存）；渠道与目录模型都只看 Key 是否填写：填了就出现在模型下拉，清空就不出现。</span>
      </div>

      <ApiCustomModelModal
        open={Boolean(customModal)}
        mode={customModal?.channel || customModal?.catalogModel ? 'edit' : 'create'}
        source={customModal?.catalogModel ? 'catalog' : 'channel'}
        kind={customModal?.groupId || 'text'}
        initial={customModal?.channel || customModal?.catalogModel || null}
        saving={customModalSaving}
        onCancel={() => setCustomModal(null)}
        onSubmit={submitCustomModel}
      />

      <ApiPresetSelectionModal
        open={Boolean(presetModal)}
        groupTitle={API_INTEGRATION_GROUPS.find(group => group.id === presetModal)?.title || ''}
        presets={API_CHANNEL_PRESETS[presetModal]?.presets || []}
        selectedIds={selectedPresetIds(
          settings.channelSelection,
          presetModal,
          (API_CHANNEL_PRESETS[presetModal]?.presets || []).map(preset => preset.id)
        )}
        saving={presetSaving}
        onCancel={() => setPresetModal(null)}
        onConfirm={ids => savePresetSelection(presetModal, ids)}
      />
    </section>
  );
}
