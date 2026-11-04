import { ConfigProvider, Form, Input, InputNumber, Select, Slider, Space, Spin, message } from 'antd';
import { AudioLines, Download, FileUp, Plus, Save, Trash2, WandSparkles } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { textToSpeech } from '../../shared/api/tts';
import { getConfig, saveConfig } from '../../shared/api/config';
import { V88Button } from '../../shared/components/V88Button';
import { V88EmptyState } from '../../shared/components/V88EmptyState';
import { createV88AntTheme } from '../../shared/styles/theme';
import { dispatchPetContext, dispatchPetState } from '../../shared/pet/stacky';
import './tts-workbench.css';

const voices = [
  { label: '晓晓（女声·温柔）', value: 'zh-CN-XiaoxiaoNeural' },
  { label: '晓辰（女声·知性）', value: 'zh-CN-XiaochenNeural' },
  { label: '云希（男声·清朗）', value: 'zh-CN-YunxiNeural' },
  { label: '云扬（男声·阳光）', value: 'zh-CN-YunyangNeural' },
  { label: '晓伊（女声·甜美）', value: 'zh-CN-XiaoyiNeural' },
  { label: '云健（男声·稳重）', value: 'zh-CN-YunjianNeural' }
];

const styles = [
  { label: '通用', value: 'general' },
  { label: '开心', value: 'cheerful' },
  { label: '悲伤', value: 'sad' },
  { label: '友好', value: 'friendly' },
  { label: '聊天', value: 'chat' }
];

function createCard(text = '') {
  return {
    id: 'tts-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 6),
    input: text,
    voice: 'zh-CN-XiaoxiaoNeural',
    style: 'general',
    speed: 1.8,
    pitch: 10,
    loading: false,
    startedAt: 0,
    audioUrl: '',
    audioBlob: null
  };
}

function NumberSlider({ ariaLabel, min, max, step, value, onChange }) {
  const safeValue = Number.isFinite(Number(value)) ? Number(value) : min;
  return (
    <div className="tts-number-slider">
      <Slider aria-label={ariaLabel} min={min} max={max} step={step} value={safeValue} onChange={onChange} />
      <InputNumber aria-label={`${ariaLabel}数值`} min={min} max={max} step={step} value={safeValue} onChange={next => onChange(next ?? safeValue)} />
    </div>
  );
}

export function TtsPage({ theme = 'dark' }) {
  const [cards, setCards] = useState([]);
  const [defaults, setDefaults] = useState({ voice: 'zh-CN-XiaoxiaoNeural', style: 'general', speed: 1.8, pitch: 10 });
  const [configLoading, setConfigLoading] = useState(true);
  const [savingDefaults, setSavingDefaults] = useState(false);
  const [readingFile, setReadingFile] = useState(false);
  const [bulkGenerating, setBulkGenerating] = useState(false);
  const [bulkProgress, setBulkProgress] = useState({ done: 0, total: 0 });
  // 只用于让「已等待 Ns」随真实时间刷新，不参与业务判断。
  const [nowMs, setNowMs] = useState(0);
  const audioUrlsRef = useRef(new Set());
  const cardRequestRef = useRef(new Map());
  const cardsRef = useRef([]);
  const mountedRef = useRef(true);

  useEffect(() => {
    cardsRef.current = cards;
  }, [cards]);

  const antTheme = useMemo(() => createV88AntTheme(theme), [theme]);

  function nextCardRequest(id) {
    const requestId = (cardRequestRef.current.get(id) || 0) + 1;
    cardRequestRef.current.set(id, requestId);
    return requestId;
  }

  function isCurrentCardRequest(id, requestId) {
    return mountedRef.current && cardRequestRef.current.get(id) === requestId;
  }

  function cardStillExists(id) {
    return cardsRef.current.some(card => card.id === id);
  }

  useEffect(() => () => {
    mountedRef.current = false;
    cardRequestRef.current.clear();
    audioUrlsRef.current.forEach(url => URL.revokeObjectURL(url));
    audioUrlsRef.current.clear();
  }, []);

  useEffect(() => {
    let active = true;
    setConfigLoading(true);
    getConfig()
      .then(config => {
        if (active && config.tts) setDefaults(config.tts);
      })
      .catch(error => message.error(error.message || '读取默认配音失败'))
      .finally(() => {
        if (active) setConfigLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  const hasCards = cards.length > 0;
  const hasLoadingCard = cards.some(card => card.loading);
  const anyWaiting = hasLoadingCard || bulkGenerating || configLoading || savingDefaults || readingFile;

  useEffect(() => {
    if (!anyWaiting) {
      setNowMs(0);
      return undefined;
    }
    setNowMs(Date.now());
    const timer = window.setInterval(() => setNowMs(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, [anyWaiting]);

  useEffect(() => {
    const generatedCount = cards.filter(card => Boolean(card.audioUrl || card.audioBlob)).length;
    dispatchPetContext({
      page: '配音',
      pagePath: '/tts',
      summary: `配音卡片 ${cards.length} 张；已生成音频 ${generatedCount} 张；默认音色：${defaults.voice || '未设置'}`,
      entities: {
        cardCount: cards.length,
        generatedCount,
        defaultVoice: defaults.voice || '',
        speed: defaults.speed,
        pitch: defaults.pitch
      },
      actions: ['检查配音卡片', '生成全部配音']
    });
  }, [cards, defaults]);

  function updateCard(id, patch) {
    setCards(current => current.map(card => card.id === id ? { ...card, ...patch } : card));
  }

  function addCard(text) {
    setCards(current => [...current, { ...createCard(text), ...defaults, loading: false, startedAt: 0 }]);
  }

  async function saveDefaults() {
    setSavingDefaults(true);
    try {
      const config = await getConfig();
      await saveConfig({ ...config, apiKey: '', tts: defaults });
      message.success('默认配音已保存');
    } catch (error) {
      message.error(error.message || '保存默认配音失败');
    } finally {
      setSavingDefaults(false);
    }
  }

  function removeCard(id) {
    nextCardRequest(id);
    setCards(current => {
      const card = current.find(item => item.id === id);
      if (card?.audioUrl) {
        URL.revokeObjectURL(card.audioUrl);
        audioUrlsRef.current.delete(card.audioUrl);
      }
      cardRequestRef.current.delete(id);
      return current.filter(item => item.id !== id);
    });
  }

  async function generateCard(id) {
    const card = cards.find(item => item.id === id);
    if (!card?.input?.trim()) {
      message.warning('请先输入要转换的文本');
      return;
    }

    const requestId = nextCardRequest(id);
    updateCard(id, { loading: true, startedAt: Date.now() });
    dispatchPetState('working', { title: '配音正在生成' });
    try {
      const blob = await textToSpeech(card);
      if (!blob || blob.size === 0) throw new Error('TTS 服务返回空音频');
      if (!isCurrentCardRequest(id, requestId) || !cardStillExists(id)) return;
      if (card.audioUrl) {
        URL.revokeObjectURL(card.audioUrl);
        audioUrlsRef.current.delete(card.audioUrl);
      }
      const nextAudioUrl = URL.createObjectURL(blob);
      audioUrlsRef.current.add(nextAudioUrl);
      if (!isCurrentCardRequest(id, requestId) || !cardStillExists(id)) {
        audioUrlsRef.current.delete(nextAudioUrl);
        URL.revokeObjectURL(nextAudioUrl);
        return;
      }
      updateCard(id, {
        audioUrl: nextAudioUrl,
        audioBlob: blob,
        loading: false,
        startedAt: 0
      });
      message.success('语音生成成功');
      dispatchPetState('success', { title: '配音生成完成', detail: '音频已经可以播放和下载。' });
    } catch (error) {
      if (!isCurrentCardRequest(id, requestId) || !cardStillExists(id)) return;
      updateCard(id, { loading: false, startedAt: 0, audioUrl: '', audioBlob: null });
      message.error(error.message || '生成失败');
      dispatchPetState('error', { title: '配音生成失败', detail: error.message || '请检查配音模型配置后重试。' });
    }
  }

  async function generateAll() {
    if (!hasCards) {
      message.warning('先添加卡片再生成');
      return;
    }
    const targets = cards.filter(card => card.input.trim());
    if (!targets.length) {
      message.warning('没有可生成的文本');
      return;
    }
    setBulkGenerating(true);
    setBulkProgress({ done: 0, total: targets.length });
    try {
      for (const card of targets) {
        await generateCard(card.id);
        setBulkProgress(progress => ({ ...progress, done: progress.done + 1 }));
      }
    } finally {
      setBulkGenerating(false);
      setBulkProgress({ done: 0, total: 0 });
    }
  }

  function downloadCard(card) {
    if (!card.audioUrl || !card.audioBlob) return;
    const link = document.createElement('a');
    link.href = card.audioUrl;
    link.download = `${card.id}.mp3`;
    link.click();
  }

  function uploadNovel(event) {
    const file = event.target.files?.[0];
    if (!file) return;
    setReadingFile(true);
    const reader = new FileReader();
    reader.onload = () => {
      const text = String(reader.result || '').trim().slice(0, 1800);
      addCard(text);
      event.target.value = '';
      setReadingFile(false);
    };
    reader.onerror = () => {
      message.error('读取文件失败');
      setReadingFile(false);
    };
    reader.readAsText(file);
  }

  return (
    <ConfigProvider theme={antTheme}>
      <div className="tts-workbench utility-workbench">
        <div className="tts-page-head">
          <div>
            <h1 className="tts-page-title">配音</h1>
            <p className="tts-page-subtitle">逐段输入文本，选好音色与风格后生成语音；全部生成会按卡片顺序依次排队。</p>
          </div>
          {bulkGenerating ? (
            <span className="tts-bulk-progress" role="status" aria-live="polite">
              <Spin size="small" />
              正在批量生成 {bulkProgress.done}/{bulkProgress.total} 张…
            </span>
          ) : null}
        </div>

        <div className="tts-toolbar">
          <div className="tts-toolbar-group">
            <V88Button variant="primary" icon={<Plus size={16} strokeWidth={1.8} aria-hidden="true" />} onClick={() => addCard()}>添加卡片</V88Button>
            <V88Button variant="secondary" icon={<FileUp size={16} strokeWidth={1.8} aria-hidden="true" />} loading={readingFile}>
              <label style={{ cursor: 'pointer' }}>
                上传小说
                <input type="file" accept=".txt" hidden onChange={uploadNovel} />
              </label>
            </V88Button>
          </div>

          <div className="tts-toolbar-group">
            <Select style={{ width: 190 }} value={defaults.voice} options={voices} onChange={voice => setDefaults(current => ({ ...current, voice }))} />
            <Select style={{ width: 110 }} value={defaults.style} options={styles} onChange={style => setDefaults(current => ({ ...current, style }))} />
          </div>

          <div className="tts-toolbar-group">
            <div className="tts-default-slider" title="默认语速">
              <span>语速</span>
              <NumberSlider ariaLabel="默认语速" min={0.5} max={2} step={0.1} value={defaults.speed} onChange={speed => setDefaults(current => ({ ...current, speed }))} />
            </div>
            <div className="tts-default-slider" title="默认音调">
              <span>音调</span>
              <NumberSlider ariaLabel="默认音调" min={-50} max={50} step={1} value={defaults.pitch} onChange={pitch => setDefaults(current => ({ ...current, pitch }))} />
            </div>
          </div>

          <div className="tts-toolbar-spacer" />

          <div className="tts-toolbar-group">
            <V88Button variant="secondary" icon={<Save size={16} strokeWidth={1.8} aria-hidden="true" />} loading={savingDefaults} onClick={saveDefaults}>保存为默认配音</V88Button>
            <V88Button variant="primary" icon={<WandSparkles size={16} strokeWidth={1.8} aria-hidden="true" />} loading={bulkGenerating} onClick={generateAll}>全部生成</V88Button>
          </div>

          {configLoading ? <span className="tts-config-hint"><Spin size="small" />正在读取默认配音…</span> : null}
        </div>

        <div className="tts-card-grid">
          {!hasCards && (
            <V88EmptyState
              icon={<AudioLines size={24} strokeWidth={1.6} aria-hidden="true" />}
              title="还没有配音卡片"
              description="点击「添加卡片」输入文本，或上传 .txt 小说自动填入开篇（最多 1800 字）。"
              action={<V88Button variant="primary" icon={<Plus size={16} strokeWidth={1.8} aria-hidden="true" />} onClick={() => addCard()}>添加卡片</V88Button>}
            />
          )}

          {cards.map((card, index) => (
            <TtsCard
              key={card.id}
              card={card}
              index={index}
              nowMs={nowMs}
              onChange={patch => updateCard(card.id, patch)}
              onGenerate={() => generateCard(card.id)}
              onDownload={() => downloadCard(card)}
              onRemove={() => removeCard(card.id)}
            />
          ))}
        </div>
      </div>
    </ConfigProvider>
  );
}

export default TtsPage;

function TtsCard({ card, index, nowMs, onChange, onGenerate, onDownload, onRemove }) {
  const stat = useMemo(() => {
    const text = card.input || '';
    return {
      total: text.length,
      chinese: (text.match(/[\u4e00-\u9fa5]/g) || []).length,
      letters: (text.match(/[a-zA-Z]/g) || []).length
    };
  }, [card.input]);

  // 超过 2s 才显示已等待秒数，避免一闪而过的等待文案干扰。
  const waitedSeconds = card.loading && card.startedAt && nowMs
    ? Math.max(0, Math.floor((nowMs - card.startedAt) / 1000))
    : 0;

  return (
    <div className="tts-card legacy-panel-card">
      <div className="tts-card-header">
        <strong>配音卡片 {index + 1}</strong>
        <V88Button
          variant="danger"
          size="small"
          icon={<Trash2 size={15} strokeWidth={1.8} aria-hidden="true" />}
          disabled={card.loading}
          onClick={onRemove}
        >
          删除
        </V88Button>
      </div>
      <div className="tts-card-body">
        <Input.TextArea
          className="legacy-input"
          rows={6}
          value={card.input}
          placeholder="在此输入要转换的文本..."
          onChange={event => onChange({ input: event.target.value })}
        />
        <div className="tts-stat">总计: {stat.total}　汉字: {stat.chinese}　字母: {stat.letters}</div>
        <Form layout="vertical">
          <Form.Item label="语音选择">
            <Select value={card.voice} options={voices} onChange={voice => onChange({ voice })} />
          </Form.Item>
          <Space wrap align="start">
            <Form.Item label="说话风格">
              <Select style={{ width: 140 }} value={card.style} options={styles} onChange={style => onChange({ style })} />
            </Form.Item>
            <Form.Item label="语速" style={{ width: 230 }}>
              <NumberSlider ariaLabel="卡片语速" min={0.5} max={2} step={0.1} value={card.speed} onChange={speed => onChange({ speed })} />
            </Form.Item>
            <Form.Item label="音调" style={{ width: 230 }}>
              <NumberSlider ariaLabel="卡片音调" min={-50} max={50} step={1} value={card.pitch} onChange={pitch => onChange({ pitch })} />
            </Form.Item>
          </Space>
        </Form>
        {card.audioUrl && <audio controls src={card.audioUrl} style={{ width: '100%' }} />}
      </div>
      <div className="tts-card-footer">
        <V88Button
          variant="primary"
          icon={<WandSparkles size={16} strokeWidth={1.8} aria-hidden="true" />}
          loading={card.loading}
          onClick={onGenerate}
        >
          生成语音
        </V88Button>
        {card.loading ? (
          <span className="tts-wait-note" role="status" aria-live="polite">
            <Spin size="small" />
            {waitedSeconds >= 2 ? `正在合成音频…已等待 ${waitedSeconds}s` : '正在合成音频…'}
          </span>
        ) : null}
        <V88Button
          variant="secondary"
          icon={<Download size={16} strokeWidth={1.8} aria-hidden="true" />}
          disabled={!card.audioBlob || card.loading}
          onClick={onDownload}
        >
          下载
        </V88Button>
      </div>
    </div>
  );
}
