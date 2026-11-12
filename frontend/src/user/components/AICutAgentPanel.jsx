import { Button, Input, Segmented, Select, Spin, Tag, message } from 'antd';
import { Bot, ChevronLeft, ChevronRight, CircleStop, Clapperboard, Plus, Scissors, Send, Sparkles } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { cancelAICutAgentRun, createAICutAgentRun, getAICutAgentRun, listAICutAgentEvents, listAICutAgentMessages } from '../../shared/api/aicutAgent.js';
import AICutAutoEditDrawer from './AICutAutoEditDrawer.jsx';
import { agentFollowupFromMessages, agentStepRows, agentToolNames, formatAgentToolMessage } from './aicutAgentPresentation.js';
import './aicut-agent-unified.css';

const COLLAPSED_KEY = 'aicut.agentCollapsed';
const MODEL_KEY = 'aicut.agentModel';
const terminalStates = new Set(['succeeded', 'failed', 'canceled']);
const suggestions = [
  { title: '粗剪成片', text: '根据现有字幕，保留与指定主题有关的完整句子，先问我想保留的主题', skill: '字幕选片', icon: Scissors },
  { title: '识别字幕', text: '读取工程，确认要识别字幕的素材和可用的 ASR 模型，再提交转写', icon: Clapperboard },
  { title: '去口误停顿', text: '读取字幕，保留自然停顿，把超过 1 秒的句间间隔缩短到 1 秒', skill: '口播精剪', icon: Scissors },
  { title: '配背景音乐', text: '从本工程已上传的音频中选择背景音乐，先确认素材和目标位置', icon: Plus }
];

function requestID() {
  return `agent-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 9)}`;
}

function modelLabel(model) {
  return model?.displayName || model?.name || model?.modelId || model?.id;
}

export default function AICutAgentPanel({ projectId, timelineRevision = 0, models = [], assets = [], activeSkill = '', draft, embedded = false, onTimelineChanged, onCollapsedChange, autoEditOpen = false, onSectionChange, onAutoApplied }) {
  const [collapsed, setCollapsed] = useState(() => window.localStorage.getItem(COLLAPSED_KEY) === '1');
  const [mode, setMode] = useState('agent');
  const [modelId, setModelId] = useState(() => window.localStorage.getItem(MODEL_KEY) || '');
  const [prompt, setPrompt] = useState('');
  const [messages, setMessages] = useState([]);
  const [events, setEvents] = useState([]);
  const [run, setRun] = useState(null);
  const [sending, setSending] = useState(false);
  const [skill, setSkill] = useState(activeSkill);
  const [answers, setAnswers] = useState({});
  const [autoBusy, setAutoBusy] = useState(false);
  const eventSequence = useRef(0);
  const composerRef = useRef(null);
  const refreshedSequences = useRef(new Set());

  const modelOptions = useMemo(() => (Array.isArray(models) ? models : []).filter(item => item?.id).map(item => ({ value: item.id, label: modelLabel(item) })), [models]);
  const running = Boolean(run && !terminalStates.has(run.state));
  const followup = agentFollowupFromMessages(messages);
  const steps = agentStepRows(events);
  useEffect(() => { setSkill(activeSkill); }, [activeSkill]);
  useEffect(() => { setAnswers({}); }, [followup?.question]);

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
    eventSequence.current = 0;
    refreshedSequences.current = new Set();
    if (!projectId) {
      setMessages([]);
      return;
    }
    let active = true;
    listAICutAgentMessages(projectId).then(async items => {
      if (!active) return;
      setMessages(items);
      const lastRunId = items.at(-1)?.runId;
      if (!lastRunId) return;
      const latestRun = await getAICutAgentRun(projectId, lastRunId);
      const latestEvents = await listAICutAgentEvents(projectId, lastRunId);
      if (!active) return;
      setRun(latestRun);
      setSkill(latestRun.activeSkillId || activeSkill);
      setEvents(latestEvents);
      eventSequence.current = Math.max(0, ...latestEvents.map(item => Number(item.sequence) || 0));
      const completedRevision = Math.max(0, ...latestEvents.filter(item => item.type === 'tool_done').map(item => Number(item.revision) || 0));
      if (completedRevision > Number(timelineRevision)) await onTimelineChanged?.(completedRevision);
    }).catch(error => { if (active) message.error(error?.message || 'Agent 对话读取失败'); });
    return () => { active = false; };
  }, [projectId]);

  useEffect(() => {
    if (!projectId || !run?.id || terminalStates.has(run.state)) return undefined;
    let active = true;
    const refresh = async () => {
      try {
        const after = eventSequence.current;
        const nextRun = await getAICutAgentRun(projectId, run.id);
        const [nextEvents, nextMessages] = await Promise.all([
          listAICutAgentEvents(projectId, run.id, after),
          listAICutAgentMessages(projectId)
        ]);
        if (!active) return;
        eventSequence.current = Math.max(after, ...nextEvents.map(item => Number(item.sequence) || 0));
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
    let timer;
    const poll = async () => { await refresh(); if (active) timer = window.setTimeout(poll, 900); };
    poll();
    return () => { active = false; window.clearTimeout(timer); };
  }, [projectId, run?.id, run?.state, onTimelineChanged]);

  function chooseAgentSuggestion(item) {
    setPrompt(item.text);
    setSkill(item.skill || '');
    window.setTimeout(() => composerRef.current?.focus?.(), 0);
  }

  async function sendAgentPrompt(override) {
    const instruction = (typeof override === 'string' ? override : prompt).trim();
    if (!instruction || !projectId || sending || running) return;
    setSending(true);
    try {
      const created = await createAICutAgentRun(projectId, {
        instruction,
        mode,
        modelId,
        activeSkillId: skill,
        expectedRevision: Number(timelineRevision) || 0,
        clientRequestId: requestID()
      });
      setRun(created);
      setEvents([]);
      eventSequence.current = 0;
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
    <div className="aicut-assistant-tabs"><Segmented block value={autoEditOpen ? 'auto' : 'chat'} disabled={running || autoBusy} onChange={value => onSectionChange?.(value === 'auto')} options={[{ label: '对话剪辑', value: 'chat' }, { label: '自动成片', value: 'auto' }]} /><p>{autoEditOpen ? '选择素材 → 生成草案 → 勾选应用' : '描述目标 → 分步执行 → 随时撤销'}</p></div>
    <div className="aicut-assistant-workflow" style={{ display: autoEditOpen ? 'flex' : 'none' }}><AICutAutoEditDrawer embedded open projectId={projectId} revision={timelineRevision} assets={assets} models={models} selectedModelId={modelId} onModelChange={setModelId} onApplied={onAutoApplied} onBusyChange={setAutoBusy} /></div>
    <div className="aicut-assistant-chat" style={{ display: autoEditOpen ? 'none' : 'contents' }}>
    <div className="aicut-agent-controls"><Segmented value={mode} disabled={running} onChange={setMode} options={[{ label: '执行', value: 'agent' }, { label: '问答', value: 'ask' }]} /><Select value={modelId || undefined} disabled={running} onChange={setModelId} options={modelOptions} placeholder="选择文本模型" /></div>
    <Select className="aicut-agent-skill-select" allowClear value={skill || undefined} disabled={running} onChange={value => setSkill(value || '')} placeholder="选择剪辑技能（可选）" options={['短剧AI剪辑', '字幕选片', '口播精剪', '画面润色', '技能创作器'].map(value => ({ value, label: value === '技能创作器' ? '整理剪辑流程' : value }))} />
    <div className="aicut-agent-conversation">
      {!messages.length ? <div className="aicut-agent-intro"><Sparkles size={24} /><h2>今天想怎么剪？</h2><p>一句话选片、整理对白、调整画面。自动成片也在这里，先看草案再应用。</p><small>字幕选片依据已有文字；转写和智能服务需使用已配置模型。</small></div> : messages.map(item => <article key={item.id} className={`aicut-agent-message ${item.role}`}><Tag>{item.role === 'assistant' ? 'Agent' : item.role === 'tool' ? '工具' : '我'}</Tag><span>{item.role === 'tool' ? formatAgentToolMessage(item.content) : item.content.replace(/\n\{"question":[\s\S]*$/, '')}</span></article>)}
      {steps.length ? <details className="aicut-agent-run-details" open><summary>执行过程 · {steps.filter(item => item.state === 'done').length}/{steps.length} 步{running ? ' · 进行中' : ''}</summary>{steps.map(item => <details key={item.toolCallId || item.sequence} className={`aicut-agent-step ${item.state}`}><summary><i />{item.title}<small>{item.state === 'done' ? '已完成' : item.state === 'failed' ? '未完成' : running ? '执行中' : '已停止'}</small></summary><p>{item.type === 'tool_start' ? '正在执行这一步' : item.message}</p>{item.arguments ? <pre>{JSON.stringify(item.arguments, null, 2)}</pre> : null}{item.revision ? <small>工程版本 {item.revision}</small> : null}</details>)}</details> : null}
      {running ? <div className="aicut-agent-thinking"><Spin size="small" /><span>{events.at(-1)?.type === 'thinking' ? events.at(-1).message : '正在处理当前步骤…'}</span></div> : null}
      {followup && !running ? <form className="aicut-agent-followup" onSubmit={event => { event.preventDefault(); sendAgentPrompt(`针对上一步问题“${followup.question}”的回答：\n${followup.fields.map(field => `${field.label}：${answers[field.id] || ''}`).join('\n')}\n请结合之前要求继续执行。`); }}><strong>{followup.question}</strong>{followup.fields.map(field => <label key={field.id}><span>{field.label}{field.required ? ' *' : ''}</span>{field.options?.length ? <Select allowClear value={answers[field.id] || undefined} onChange={value => setAnswers(current => ({ ...current, [field.id]: value }))} options={field.options.map(value => ({ value, label: value }))} /> : <Input value={answers[field.id] || ''} onChange={event => setAnswers(current => ({ ...current, [field.id]: event.target.value }))} />}</label>)}<Button htmlType="submit" type="primary" loading={sending} disabled={followup.fields.some(field => field.required && !String(answers[field.id] || '').trim())}>补充并继续</Button></form> : null}
    {!messages.length ? <div className="aicut-agent-skills">{suggestions.map(item => { const Icon = item.icon; return <button type="button" key={item.title} onClick={() => chooseAgentSuggestion(item)}><Icon size={15} /><strong>{item.title}</strong><span>{item.text}</span></button>; })}</div> : null}
    </div>
    <div className="aicut-agent-notice"><i />{mode === 'ask' ? '问答模式不会改动时间线' : 'Agent 模式执行工具前会校验工程版本'}</div>
    <div className="aicut-agent-compose"><Input.TextArea ref={composerRef} value={prompt} onChange={event => setPrompt(event.target.value)} onPressEnter={event => { if (!event.shiftKey) { event.preventDefault(); sendAgentPrompt(); } }} disabled={!projectId || running} autoSize={{ minRows: 2, maxRows: 5 }} placeholder="描述这一集要怎么剪… (Enter 发送，Shift+Enter 换行)" /><div><Tag>{mode === 'ask' ? '问答' : 'Agent'}</Tag>{running ? <Button danger icon={<CircleStop size={14} />} onClick={stopRun}>停止</Button> : <Button type="primary" icon={<Send size={14} />} loading={sending} disabled={!prompt.trim() || !projectId} onClick={sendAgentPrompt}>发送</Button>}</div></div>
    <details className="aicut-agent-capabilities"><summary>已接入能力 · {Object.keys(agentToolNames).length} 项</summary><p>{Object.values(agentToolNames).join(' · ')}</p><small>按字幕压缩间隔不能代替真实静音检测；自动成片共用现有剪辑规则。自定义流程目前可整理复制。</small></details>
    </div>
  </aside>;
}
