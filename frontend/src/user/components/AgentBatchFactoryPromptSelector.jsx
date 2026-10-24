import { Button, Popover, Select } from 'antd';
import { WandSparkles } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { getBatch, listPrompts, saveBatchSettings } from '../../shared/api/batchFactoryV11';
import { batchFactoryPromptSelectorTypes, saveBatchFactoryPromptSelection } from '../pages/agent-canvas-prompt-selection';

const PROMPT_LABELS = {
  hook: 'Hook 提示词',
  director: 'Director 提示词',
  visual: '画面提示词',
  video: '视频提示词',
  audio_match: '音频匹配提示词',
  book_merge: '成片合成提示词'
};

function batchFrom(result) {
  return result?.batch || result || null;
}

function promptSelections(batch) {
  const settings = batch?.settingsState || batch?.settings || {};
  const selections = settings?.patch?.promptSelections;
  return selections && typeof selections === 'object' ? selections : {};
}

export function AgentBatchFactoryPromptSelector({ config = {}, disabled = false, onSaved, display = 'inline' }) {
  const batchId = String(config.batchId || '').trim();
  const bookId = String(config.bookId || '').trim();
  const types = useMemo(() => batchFactoryPromptSelectorTypes(config), [config.source, config.resourceType]);
  const [batch, setBatch] = useState(null);
  const [catalog, setCatalog] = useState([]);
  const [loading, setLoading] = useState(false);
  const [savingType, setSavingType] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    if (!batchId || !types.length) return undefined;
    let active = true;
    setLoading(true);
    setError('');
    Promise.all([getBatch(batchId), listPrompts({ scope: 'batch_factory', limit: 200 })])
      .then(([batchResult, promptResult]) => {
        if (!active) return;
        setBatch(batchFrom(batchResult));
        setCatalog(Array.isArray(promptResult?.prompts) ? promptResult.prompts : []);
      }).catch(loadError => {
        if (active) {
          setBatch(null);
          setCatalog([]);
          setError(loadError?.message || '提示词模板读取失败');
        }
      }).finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [batchId, types.join('|')]);

  if (!batchId || !types.length) return null;
  const selected = promptSelections(batch);

  async function change(type, promptId) {
    if (!batch || savingType) return;
    setSavingType(type);
    setError('');
    try {
      const saved = await saveBatchFactoryPromptSelection({ batch, type, promptId, save: saveBatchSettings });
      const nextSelections = { ...selected, [type]: String(promptId || '').trim() };
      setBatch(current => ({
        ...(current || {}),
        settingsState: {
          ...(current?.settingsState || current?.settings || {}),
          patch: { ...((current?.settingsState || current?.settings || {}).patch || {}), promptSelections: nextSelections },
          revision: Number(saved?.revision || (current?.settingsState || current?.settings || {}).revision || 0)
        }
      }));
      await onSaved?.({ batchId, bookId, type });
    } catch (saveError) {
      setError(saveError?.message || '提示词模板保存失败');
    } finally {
      setSavingType('');
    }
  }

  const selector = <section className="agent-node-composer__batch-prompt-selector" aria-label="批量工厂提示词模板">
    <strong>提示词模板</strong>
    {types.map(type => {
      const options = catalog.filter(item => item?.type === type && item?.enabled !== false)
        .map(item => ({ value: item.id, label: item.version ? `${item.name} · v${item.version}` : item.name }));
      return <label key={type}>
        <span>{PROMPT_LABELS[type] || type}</span>
        <Select
          allowClear
          size="small"
          aria-label={`选择${PROMPT_LABELS[type] || type}`}
          placeholder={loading ? '正在读取提示词模板' : '继承系统默认提示词'}
          value={selected[type] || undefined}
          options={options}
          loading={loading}
          disabled={disabled || loading || Boolean(savingType)}
          onChange={value => { void change(type, value); }}
        />
      </label>;
    })}
    {error ? <small className="is-error" role="alert">{error}</small> : null}
  </section>;

  if (!display.startsWith('popover')) return selector;
  const compact = display === 'popover-compact';
  return <Popover content={selector} trigger="click" placement="bottomLeft" getPopupContainer={trigger => trigger.parentElement}>
    <Button size="small" type="text" disabled={disabled} aria-label="选择提示词" title="选择提示词">
      {compact ? <WandSparkles size={14} /> : '选择提示词'}
    </Button>
  </Popover>;
}
