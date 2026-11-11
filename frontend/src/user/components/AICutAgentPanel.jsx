import { Button, Input, Segmented, Select, Spin, Tag, message } from 'antd';
import { Bot, ChevronLeft, ChevronRight, CircleStop, Clapperboard, Plus, Scissors, Send, Sparkles } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { cancelAICutAgentRun, createAICutAgentRun, getAICutAgentRun, listAICutAgentEvents, listAICutAgentMessages } from '../../shared/api/aicutAgent.js';

const COLLAPSED_KEY = 'aicut.agentCollapsed';
const MODEL_KEY = 'aicut.agentModel';
const terminalStates = new Set(['succeeded', 'failed', 'canceled']);
const suggestions = [
  { title: '粗剪成片', text: '按剧本把这一集粗剪出来，删掉重复和废镜头', icon: Scissors },
  { title: '识别字幕', text: '识别全片字幕，并对照剧本纠正错字', icon: Clapperboard },
  { title: '去口误停顿', text: '删掉口误、语气词和超过 1 秒的停顿', icon: Scissors },
  { title: '配背景音乐', text: '根据情绪给高潮段落配一段合适的背景音乐', icon: Plus }
];

function requestID() {
  return `agent-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 9)}`;
}

function modelLabel(model) {
  return model?.displayName || model?.name || model?.modelId || model?.id;
}

export default function AICutAgentPanel({ projectId, timelineRevision = 0, models = [], activeSkill = '', draft, embedded = false, onTimelineChanged, onCollapsedChange }) {
  const [collapsed, setCollapsed] = useState(() => window.localStorage.getItem(COLLAPSED_KEY) === '1');
  const [mode, setMode] = useState('agent');
  const [modelId, setModelId] = useState(() => window.localStorage.getItem(MODEL_KEY) || '');
  const [prompt, setPrompt] = useState('');
  const [messages, setMessages] = useState([]);
  const [events, setEvents] = useState([]);
  const [run, setRun] = useState(null);
  const [sending, setSending] = useState(false);
  const composerRef = useRef(null);
  const refreshedSequences = useRef(new Set());

  const modelOptions = useMemo(() => (Array.isArray(models) ? models : []).filter(item => item?.id).map(item => ({ value: item.id, label: modelLabel(item) })), [models]);
  const running = Boolean(run && !terminalStates.has(run.state));

  useEffect(() => {
    if (!modelId && modelOptions[0]?.value) setModelId(modelOptions[0].value);
  }, [modelId, modelOptions]);

  useEffect(() => {
    window.localStorage.setItem(COLLAPSED_KEY, collapsed ? '1' : '0');
    onCollapsedChange?.(collapsed);
  }, [collapsed, onCollapsedChange]);

  useEffect(() => {
    if (modelId) window.localStorage.setItem(MODEL_KEY, modelId);
  }, [modelId]);

  useEffect(() => {
    if (!draft?.text) return;
    setPrompt(draft.text);
    window.setTimeout(() => composerRef.current?.focus?.(), 0);
  }, [draft]);

  useEffect(() => {
    setRun(null);
    setEvents([]);
    refreshedSequences.current = new Set();
    if (!projectId) {
      setMessages([]);
      return;
    }
    let active = true;
    listAICutAgentMessages(projectId).then(items => { if (active) setMessages(items); }).catch(error => { if (active) message.error(error?.message || 'Agent 对话读取失败'); });
    return () => { active = false; };
  }, [projectId]);

  useEffect(() => {
    if (!projectId || !run?.id || terminalStates.has(run.state)) return undefined;
    let active = true;
    const refresh = async () => {
      try {
        const after = events.reduce((max, item) => Math.max(max, Number(item.sequence) || 0), 0);
        const [nextRun, nextEvents, nextMessages] = await Promise.all([
          getAICutAgentRun(projectId, run.id),
          listAICutAgentEvents(projectId, run.id, after),
          listAICutAgentMessages(projectId)
        ]);
        if (!active) return;
        setRun(nextRun);
        if (nextEvents.length) setEvents(current => [...current, ...nextEvents.filter(item => !current.some(old => old.sequence === item.sequence))]);
        setMessages(nextMessages);
        for (const event of nextEvents) {
          if (event.type === 'tool_done' && event.revision > 0 && !refreshedSequences.current.has(event.sequence)) {
            refreshedSequences.current.add(event.sequence);
            await onTimelineChanged?.(event.revision);
          }
        }
        if (nextRun.state === 'failed') message.error(nextRun.error || 'Agent 执行失败');
        if (nextRun.state === 'canceled') message.info('Agent 已停止');
      } catch (error) {
        if (active) message.error(error?.message || 'Agent 状态读取失败');
      }
    };
    refresh();
    const timer = window.setInterval(refresh, 900);
    return () => { active = false; window.clearInterval(timer); };
  }, [projectId, run?.id, run?.state, events, onTimelineChanged]);

  function chooseAgentSuggestion(text) {
    setPrompt(text);
    window.setTimeout(() => composerRef.current?.focus?.(), 0);
  }

  async function sendAgentPrompt() {
    const instruction = prompt.trim();
    if (!instruction || !projectId || sending || running) return;
    setSending(true);
    try {
      const created = await createAICutAgentRun(projectId, {
        instruction,
        mode,
        modelId,
        activeSkillId: activeSkill,
        expectedRevision: Number(timelineRevision) || 0,
        clientRequestId: requestID()
      });
      setRun(created);
      setEvents([]);
      setPrompt('');
      setMessages(await listAICutAgentMessages(projectId));
    } catch (error) {
      message.error(error?.message || 'Agent 任务提交失败');
    } finally {
      setSending(false);
    }
  }

  async function stopRun() {
    if (!projectId || !run?.id || !running) return;
    try {
      setRun(await cancelAICutAgentRun(projectId, run.id));
      message.info('正在停止 Agent');
    } catch (error) {
      message.error(error?.message || '停止 Agent 失败');
    }
  }

  if (collapsed && !embedded) return <aside className="aicut-agent-panel collapsed"><Button type="text" aria-label="展开 Agent" title="展开 Agent" icon={<ChevronRight size={17} />} onClick={() => setCollapsed(false)} /><Bot size={20} /><span>Agent</span>{running ? <Spin size="small" /> : null}</aside>;

  return <aside className={`aicut-agent-panel${embedded ? ' embedded' : ''}`}>
    <header className="aicut-agent-title"><span><Sparkles size={18} /><strong>AI-CUT Agent</strong><small>● 就绪</small></span>{embedded ? null : <Button type="text" aria-label="收起 Agent" title="收起 Agent" icon={<ChevronLeft size={17} />} onClick={() => setCollapsed(true)} />}</header>
    <div className="aicut-agent-controls"><Segmented value={mode} onChange={setMode} options={[{ label: 'Agent', value: 'agent' }, { label: '问答', value: 'ask' }]} /><Select value={modelId || undefined} onChange={setModelId} options={modelOptions} placeholder="选择文本模型" /></div>
    <div className="aicut-agent-conversation">
      {!messages.length ? <div className="aicut-agent-intro"><Sparkles size={24} /><h2>今天想怎么剪？</h2><p>直接描述剪辑目标，Agent 会调用时间线工具；切到“问答”时只给建议，不会修改工程。</p></div> : messages.map(item => <article key={item.id} className={`aicut-agent-message ${item.role}`}><Tag>{item.role === 'assistant' ? 'Agent' : item.role === 'tool' ? '工具' : '我'}</Tag><span>{item.content}</span></article>)}
      {events.length ? <div className="aicut-agent-events">{events.slice(-8).map(item => <div key={item.sequence} className={item.type}><span>{item.toolName || item.type}</span><small>{item.message}</small></div>)}</div> : null}
    </div>
    {!messages.length ? <div className="aicut-agent-skills">{suggestions.map(item => { const Icon = item.icon; return <button type="button" key={item.title} onClick={() => chooseAgentSuggestion(item.text)}><Icon size={15} /><strong>{item.title}</strong><span>{item.text}</span></button>; })}</div> : null}
    <div className="aicut-agent-notice"><i />{mode === 'ask' ? '问答模式不会改动时间线' : 'Agent 模式执行工具前会校验工程版本'}</div>
    <div className="aicut-agent-compose"><Input.TextArea ref={composerRef} value={prompt} onChange={event => setPrompt(event.target.value)} onPressEnter={event => { if (!event.shiftKey) { event.preventDefault(); sendAgentPrompt(); } }} disabled={!projectId || running} autoSize={{ minRows: 2, maxRows: 5 }} placeholder="描述这一集要怎么剪… (Enter 发送，Shift+Enter 换行)" /><div><Tag>{mode === 'ask' ? '问答' : 'Agent'}</Tag>{running ? <Button danger icon={<CircleStop size={14} />} onClick={stopRun}>停止</Button> : <Button type="primary" icon={<Send size={14} />} loading={sending} disabled={!prompt.trim() || !projectId} onClick={sendAgentPrompt}>发送</Button>}</div></div>
  </aside>;
}
