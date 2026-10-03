import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Alert,
  Avatar,
  Badge,
  Button,
  Divider,
  Drawer,
  Empty,
  Input,
  Modal,
  Progress,
  Space,
  Spin,
  Tag,
  Tooltip,
  Typography,
  message,
} from 'antd';
import {
  createAgentThread,
  getAgentThread,
  listAgentTasks,
  listAgentThreads,
  sendAgentMessage,
} from './agentApi.js';
import {
  buildNavigationTarget,
  groupAgentTasks,
  normalizeMediaAssetIds,
  taskStatusMeta,
  toolCardMeta,
} from './agentViewModel.js';
import './AgentWorkspace.css';

const { Text, Title, Paragraph } = Typography;
const { TextArea } = Input;

const QUICK_PROMPTS = [
  '帮我打开小说获取',
  '打开批量工厂',
  '图片模型在哪里设置？',
  '创建任务：检查今天的生产任务',
];

function timeLabel(value) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return new Intl.DateTimeFormat('zh-CN', { hour: '2-digit', minute: '2-digit' }).format(date);
}

function threadTime(value) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  const now = new Date();
  if (date.toDateString() === now.toDateString()) return timeLabel(value);
  return new Intl.DateTimeFormat('zh-CN', { month: 'numeric', day: 'numeric' }).format(date);
}

function TaskCard({ task, compact = false, onInspect }) {
  const meta = taskStatusMeta(task?.status);
  const current = Number(task?.progress_current || 0);
  const total = Number(task?.progress_total || 0);
  const percent = total > 0 ? Math.max(0, Math.min(100, Math.round((current / total) * 100))) : null;
  return (
    <button type="button" className={`agent-task-card${compact ? ' is-compact' : ''}`} onClick={() => onInspect?.({ type: 'task', item: task })}>
      <div className="agent-task-card-head">
        <strong>{task?.title || '未命名任务'}</strong>
        <Tag color={meta.tone === 'default' ? undefined : meta.tone}>{meta.label}</Tag>
      </div>
      {task?.detail ? <span className="agent-task-detail">{task.detail}</span> : null}
      {percent !== null ? <Progress percent={percent} size="small" showInfo={!compact} /> : null}
      {total > 0 ? <small>{current} / {total}</small> : null}
    </button>
  );
}

function MediaReferenceCard({ id, onInspect }) {
  return (
    <button type="button" className="agent-media-card" onClick={() => onInspect?.({ type: 'media', item: { id } })}>
      <div className="agent-media-preview">◫</div>
      <div className="agent-media-copy">
        <strong>媒体资产</strong>
        <span>{id}</span>
        <small>TOS 资产引用</small>
      </div>
    </button>
  );
}

function ToolCard({ toolCall, onInspect, onNavigate }) {
  const meta = toolCardMeta(toolCall);
  const target = buildNavigationTarget(toolCall);
  return (
    <div className={`agent-tool-card is-${meta.tone}`}>
      <button type="button" className="agent-tool-main" onClick={() => onInspect?.({ type: 'tool', item: toolCall })}>
        <span className="agent-tool-icon">↗</span>
        <span className="agent-tool-copy"><strong>{meta.title}</strong><small>{meta.detail}</small></span>
      </button>
      {target ? <Button size="small" type="primary" ghost onClick={() => onNavigate(target)}>{meta.actionLabel}</Button> : null}
    </div>
  );
}

function ChatMessage({ item, tools, onInspect, onNavigate }) {
  const isUser = item?.role === 'user';
  const attachedTools = (tools || []).filter(tool => tool?.message_id === item?.id);
  const mediaIDs = normalizeMediaAssetIds(item?.media_asset_ids || []);
  return (
    <article className={`agent-message ${isUser ? 'is-user' : 'is-agent'}`}>
      {!isUser ? <Avatar className="agent-avatar" size={34}>A</Avatar> : null}
      <div className="agent-message-stack">
        <div className="agent-message-meta">
          <strong>{isUser ? '你' : '一战 Agent'}</strong>
          <span>{timeLabel(item?.created_at)}</span>
        </div>
        <div className="agent-message-bubble">
          {String(item?.content || '').split('\n').map((line, index) => <React.Fragment key={`${item?.id}-${index}`}>{index ? <br /> : null}{line}</React.Fragment>)}
        </div>
        {mediaIDs.length ? <div className="agent-media-grid">{mediaIDs.map(id => <MediaReferenceCard key={id} id={id} onInspect={onInspect} />)}</div> : null}
        {attachedTools.length ? <div className="agent-message-tools">{attachedTools.map(tool => <ToolCard key={tool.id} toolCall={tool} onInspect={onInspect} onNavigate={onNavigate} />)}</div> : null}
      </div>
    </article>
  );
}

function InspectorContent({ selected }) {
  if (!selected) return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="选择任务、工具或媒体查看详情" />;
  if (selected.type === 'task') {
    const task = selected.item || {};
    const meta = taskStatusMeta(task.status);
    return <div className="agent-inspector-content"><Tag color={meta.tone === 'default' ? undefined : meta.tone}>{meta.label}</Tag><Title level={4}>{task.title || '任务'}</Title>{task.detail ? <Paragraph>{task.detail}</Paragraph> : null}<Divider /><Text type="secondary">任务 ID</Text><code>{task.id}</code><Text type="secondary">Thread</Text><code>{task.thread_id}</code></div>;
  }
  if (selected.type === 'tool') {
    const tool = selected.item || {};
    return <div className="agent-inspector-content"><Tag color="processing">工具执行</Tag><Title level={4}>{toolCardMeta(tool).title}</Title><Text type="secondary">工具</Text><code>{tool.tool_name}</code><Text type="secondary">状态</Text><code>{tool.status}</code><Text type="secondary">参数</Text><pre>{JSON.stringify(tool.arguments || {}, null, 2)}</pre></div>;
  }
  if (selected.type === 'media') {
    return <div className="agent-inspector-content"><Tag color="purple">TOS 媒体</Tag><Title level={4}>媒体资产</Title><div className="agent-inspector-media">◫</div><Text type="secondary">media_asset_id</Text><code>{selected.item?.id}</code><Paragraph type="secondary">正式图片/视频只通过媒体资产 ID 被 Agent 引用；文件字节属于统一 TOS 媒体层。</Paragraph></div>;
  }
  return null;
}

export default function AgentWorkspace() {
  const [threads, setThreads] = useState([]);
  const [activeThreadId, setActiveThreadId] = useState(() => sessionStorage.getItem('yizhan:agent:thread') || '');
  const [snapshot, setSnapshot] = useState(null);
  const [allTasks, setAllTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [threadLoading, setThreadLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [composer, setComposer] = useState('');
  const [contextMedia, setContextMedia] = useState([]);
  const [error, setError] = useState('');
  const [selected, setSelected] = useState(null);
  const [taskDrawerOpen, setTaskDrawerOpen] = useState(false);
  const [mobileThreadsOpen, setMobileThreadsOpen] = useState(false);
  const [mediaModalOpen, setMediaModalOpen] = useState(false);
  const [mediaDraft, setMediaDraft] = useState('');
  const [messageApi, contextHolder] = message.useMessage();
  const endRef = useRef(null);

  const taskGroups = useMemo(() => groupAgentTasks(allTasks), [allTasks]);
  const threadTasks = snapshot?.tasks || [];
  const messages = snapshot?.messages || [];
  const tools = snapshot?.tool_calls || [];

  const refreshTasks = useCallback(async () => {
    try { setAllTasks(await listAgentTasks()); } catch { /* Sidebar task failure must not blank chat. */ }
  }, []);

  const loadThread = useCallback(async threadId => {
    if (!threadId) { setSnapshot(null); return; }
    setThreadLoading(true);
    setError('');
    try {
      const data = await getAgentThread(threadId);
      setSnapshot(data);
      setActiveThreadId(threadId);
      sessionStorage.setItem('yizhan:agent:thread', threadId);
    } catch (requestError) {
      setError(requestError?.message || '读取 Agent 对话失败');
    } finally {
      setThreadLoading(false);
    }
  }, []);

  const refreshThreads = useCallback(async ({ autoCreate = false } = {}) => {
    const rows = await listAgentThreads();
    setThreads(rows);
    if (!rows.length && autoCreate) {
      const created = await createAgentThread({ title: '新对话' });
      setThreads([created]);
      await loadThread(created.id);
      return;
    }
    if (rows.length) {
      const preferred = rows.some(row => row.id === activeThreadId) ? activeThreadId : rows[0].id;
      if (!snapshot || snapshot.thread?.id !== preferred) await loadThread(preferred);
    }
  }, [activeThreadId, loadThread, snapshot]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError('');
      try {
        if (!cancelled) await refreshThreads({ autoCreate: true });
        if (!cancelled) await refreshTasks();
      } catch (requestError) {
        if (!cancelled) setError(requestError?.message || 'Agent 工作区暂时无法连接');
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
    // Initial boot only. refreshThreads is intentionally invoked once here.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => { endRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' }); }, [messages.length, tools.length]);

  async function handleNewThread() {
    try {
      const created = await createAgentThread({ title: '新对话' });
      setThreads(current => [created, ...current]);
      setMobileThreadsOpen(false);
      setContextMedia([]);
      setComposer('');
      await loadThread(created.id);
    } catch (requestError) { messageApi.error(requestError?.message || '创建对话失败'); }
  }

  function handleNavigation(target) {
    if (!target || !target.startsWith('/')) return;
    window.location.assign(target);
  }

  async function handleSend(value = composer) {
    const content = String(value || '').trim();
    if ((!content && !contextMedia.length) || sending) return;
    let threadID = activeThreadId;
    try {
      setSending(true);
      setError('');
      if (!threadID) {
        const created = await createAgentThread({ title: '新对话' });
        setThreads(current => [created, ...current]);
        threadID = created.id;
        setActiveThreadId(threadID);
      }
      const result = await sendAgentMessage(threadID, { content, mediaAssetIds: contextMedia });
      setComposer('');
      setContextMedia([]);
      await Promise.all([loadThread(threadID), refreshTasks(), listAgentThreads().then(setThreads)]);
      const target = buildNavigationTarget(result?.tool_call);
      if (target) window.setTimeout(() => handleNavigation(target), 650);
    } catch (requestError) {
      setError(requestError?.message || '发送失败');
    } finally {
      setSending(false);
    }
  }

  function addMediaContext() {
    const ids = normalizeMediaAssetIds(mediaDraft.split(/[\s,，]+/));
    if (!ids.length) {
      messageApi.warning('请输入有效的 media_asset_id');
      return;
    }
    setContextMedia(current => normalizeMediaAssetIds([...current, ...ids]));
    setMediaDraft('');
    setMediaModalOpen(false);
  }

  const sidebar = <div className="agent-sidebar-inner">
    <div className="agent-brand"><div className="agent-brand-mark">A</div><div><strong>一战 Agent</strong><span>AI 工作区</span></div></div>
    <Button className="agent-new-thread" type="primary" block onClick={handleNewThread}>＋ 新对话</Button>
    <div className="agent-side-section-title"><span>最近对话</span><small>{threads.length}</small></div>
    <div className="agent-thread-list">
      {threads.map(thread => <button type="button" key={thread.id} className={`agent-thread-item${thread.id === activeThreadId ? ' is-active' : ''}`} onClick={() => { setMobileThreadsOpen(false); loadThread(thread.id); }}><span className="agent-thread-dot" /><span className="agent-thread-copy"><strong>{thread.title || '新对话'}</strong><small>{threadTime(thread.updated_at)}</small></span></button>)}
    </div>
    <Divider />
    <div className="agent-side-section-title"><span>任务</span><Button type="text" size="small" onClick={() => setTaskDrawerOpen(true)}>查看全部</Button></div>
    <div className="agent-task-summary">
      <button type="button" onClick={() => setTaskDrawerOpen(true)}><Badge status="processing" /><span>进行 / 等待</span><strong>{taskGroups.active.length}</strong></button>
      <button type="button" onClick={() => setTaskDrawerOpen(true)}><Badge status="warning" /><span>待我决定</span><strong>{taskGroups.needsDecision.length}</strong></button>
      <button type="button" onClick={() => setTaskDrawerOpen(true)}><Badge status="success" /><span>已完成</span><strong>{taskGroups.completed.length}</strong></button>
    </div>
    <div className="agent-sidebar-footer"><button type="button" onClick={() => handleNavigation('/batch-factory')}>批量工厂 <span>↗</span></button><button type="button" onClick={() => setComposer('帮我打开小说获取')}>问 Agent 打开功能 <span>⌘</span></button></div>
  </div>;

  return (
    <div className="agent-workspace">
      {contextHolder}
      <aside className="agent-sidebar">{sidebar}</aside>
      <Drawer width={300} placement="left" open={mobileThreadsOpen} onClose={() => setMobileThreadsOpen(false)} styles={{ body: { padding: 0 } }}>{sidebar}</Drawer>

      <main className="agent-main">
        <header className="agent-header">
          <div className="agent-header-left">
            <Button className="agent-mobile-menu" type="text" onClick={() => setMobileThreadsOpen(true)}>☰</Button>
            <div className="agent-header-avatar"><span>✦</span></div>
            <div><div className="agent-title-row"><strong>一战 Agent</strong><Badge status="success" text="在线" /></div><span className="agent-header-thread">{snapshot?.thread?.title || '新对话'}</span></div>
          </div>
          <Space><Tooltip title="当前对话任务"><Button onClick={() => setTaskDrawerOpen(true)}>任务 {threadTasks.length ? `· ${threadTasks.length}` : ''}</Button></Tooltip><Button onClick={handleNewThread}>新对话</Button></Space>
        </header>

        {error ? <Alert className="agent-error" type="error" showIcon closable message={error} onClose={() => setError('')} action={<Button size="small" onClick={() => activeThreadId ? loadThread(activeThreadId) : refreshThreads({ autoCreate: true })}>重试</Button>} /> : null}

        <section className="agent-chat">
          {loading ? <div className="agent-center-state"><Spin size="large" /><span>正在打开 Agent 工作区…</span></div> : threadLoading ? <div className="agent-center-state"><Spin /><span>正在读取对话…</span></div> : !messages.length ? <div className="agent-welcome"><div className="agent-welcome-orb">✦</div><Title level={2}>今天想让我帮你做什么？</Title><Paragraph>不用记页面和按钮。你可以直接问设置在哪里、让我打开小说获取，或者把要做的事情记录成任务。</Paragraph><div className="agent-quick-grid">{QUICK_PROMPTS.map(prompt => <button type="button" key={prompt} onClick={() => handleSend(prompt)}><span>↗</span>{prompt}</button>)}</div></div> : <div className="agent-message-list">{messages.map(item => <ChatMessage key={item.id} item={item} tools={tools} onInspect={setSelected} onNavigate={handleNavigation} />)}<div ref={endRef} /></div>}
        </section>

        <footer className="agent-composer-wrap">
          <div className="agent-composer">
            {contextMedia.length ? <div className="agent-context-row"><span>正在引用</span>{contextMedia.map(id => <Tag key={id} closable onClose={() => setContextMedia(current => current.filter(item => item !== id))}>◫ {id}</Tag>)}</div> : null}
            <TextArea autoSize={{ minRows: 2, maxRows: 7 }} value={composer} onChange={event => setComposer(event.target.value)} placeholder="告诉 Agent 你想做什么… 例如：帮我打开小说获取" onKeyDown={event => { if (event.key === 'Enter' && !event.shiftKey) { event.preventDefault(); handleSend(); } }} disabled={sending} />
            <div className="agent-composer-actions"><Space><Button type="text" onClick={() => setMediaModalOpen(true)}>＋ 媒体</Button><Button type="text" onClick={() => setComposer(current => current ? `${current} @资产` : '@资产 ')}>＠资产</Button></Space><div className="agent-send-side"><span>Enter 发送 · Shift+Enter 换行</span><Button className="agent-send" type="primary" loading={sending} disabled={!composer.trim() && !contextMedia.length} onClick={() => handleSend()}>↑</Button></div></div>
          </div>
          <span className="agent-disclaimer">Agent 会调用系统真实能力；尚未接入的生成工具不会伪装成已完成。</span>
        </footer>
      </main>

      {selected ? <aside className="agent-inspector"><div className="agent-inspector-head"><strong>详情</strong><Button type="text" onClick={() => setSelected(null)}>×</Button></div><InspectorContent selected={selected} /></aside> : null}

      <Drawer title="Agent 任务" width={420} open={taskDrawerOpen} onClose={() => setTaskDrawerOpen(false)}>
        <div className="agent-task-drawer">
          <div className="agent-task-group"><h4>进行 / 等待 <Tag>{taskGroups.active.length}</Tag></h4>{taskGroups.active.length ? taskGroups.active.map(task => <TaskCard key={task.id} task={task} onInspect={item => { setSelected(item); setTaskDrawerOpen(false); }} />) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无进行中的任务" />}</div>
          <div className="agent-task-group"><h4>待我决定 <Tag color="warning">{taskGroups.needsDecision.length}</Tag></h4>{taskGroups.needsDecision.map(task => <TaskCard key={task.id} task={task} onInspect={item => { setSelected(item); setTaskDrawerOpen(false); }} />)}</div>
          <div className="agent-task-group"><h4>已完成 <Tag color="success">{taskGroups.completed.length}</Tag></h4>{taskGroups.completed.slice(0, 20).map(task => <TaskCard key={task.id} task={task} compact onInspect={item => { setSelected(item); setTaskDrawerOpen(false); }} />)}</div>
          {taskGroups.failed.length ? <div className="agent-task-group"><h4>失败 <Tag color="error">{taskGroups.failed.length}</Tag></h4>{taskGroups.failed.map(task => <TaskCard key={task.id} task={task} onInspect={item => { setSelected(item); setTaskDrawerOpen(false); }} />)}</div> : null}
        </div>
      </Drawer>

      <Modal title="引用媒体资产" open={mediaModalOpen} onCancel={() => setMediaModalOpen(false)} onOk={addMediaContext} okText="引用" cancelText="取消">
        <Paragraph type="secondary">输入一个或多个 `media_asset_id`。Agent 只保存媒体资产引用，正式图片/视频文件统一由 TOS 媒体层管理。</Paragraph>
        <Input value={mediaDraft} onChange={event => setMediaDraft(event.target.value)} placeholder="例如：asset_xxx" onPressEnter={addMediaContext} />
      </Modal>
    </div>
  );
}
