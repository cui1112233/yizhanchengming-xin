import { Button, Input, Segmented, Tag, message } from 'antd';
import { useCallback, useEffect, useState } from 'react';
import { appendAgentProxyMessage, createAgentProxySession, createAgentProxyWorkspace, listAgentProxyMessages, listAgentProxyScheduledJobs, listAgentProxySkills, listAgentProxySessions, listAgentProxyWorkspaces, updateAgentProxySession } from '../../shared/api/agentProxy';
import './agent-proxy-workspace.css';

const PERMISSION_OPTIONS = [
  { label: 'plan 计划', value: 'plan' },
  { label: 'auto 自动', value: 'auto' }
];

const ROLE_LABELS = { user: '你', assistant: 'CM', system: '系统', tool: '工具' };

function permissionLabel(mode) {
  return mode === 'auto' ? 'auto 自动' : 'plan 计划';
}

function formatTime(value) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
}

// Agent 代理工作区：工作区 → 会话 → 消息，独立于右侧画布对话。
export function AgentProxyWorkspace({ open, onClose }) {
  const [workspaces, setWorkspaces] = useState([]);
  const [workspacesLoading, setWorkspacesLoading] = useState(false);
  const [workspaceId, setWorkspaceId] = useState('');
  const [workspaceDraftName, setWorkspaceDraftName] = useState('');
  const [workspaceDraftMode, setWorkspaceDraftMode] = useState('plan');
  const [creatingWorkspace, setCreatingWorkspace] = useState(false);

  const [sessions, setSessions] = useState([]);
  const [sessionsLoading, setSessionsLoading] = useState(false);
  const [sessionId, setSessionId] = useState('');
  const [sessionDraftTitle, setSessionDraftTitle] = useState('');
  const [creatingSession, setCreatingSession] = useState(false);

  const [messages, setMessages] = useState([]);
  const [messagesLoading, setMessagesLoading] = useState(false);
  const [draft, setDraft] = useState('');
  const [sending, setSending] = useState(false);
  const [permissionBusy, setPermissionBusy] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [placeholderCounts, setPlaceholderCounts] = useState({ skills: 0, jobs: 0 });

  const activeWorkspace = workspaces.find(item => item.id === workspaceId) || null;
  const activeSession = sessions.find(item => item.id === sessionId) || null;

  const loadWorkspaces = useCallback(async (preferredId = '') => {
    setWorkspacesLoading(true);
    try {
      const data = await listAgentProxyWorkspaces();
      const items = Array.isArray(data?.workspaces) ? data.workspaces : [];
      setWorkspaces(items);
      setWorkspaceId(current => {
        if (preferredId && items.some(item => item.id === preferredId)) return preferredId;
        if (current && items.some(item => item.id === current)) return current;
        return items[0]?.id || '';
      });
    } catch (error) {
      message.error(error.message || '读取工作区失败');
    } finally {
      setWorkspacesLoading(false);
    }
  }, []);

  const loadMessages = useCallback(async targetId => {
    if (!targetId) {
      setMessages([]);
      return;
    }
    setMessagesLoading(true);
    try {
      const data = await listAgentProxyMessages(targetId);
      setMessages(Array.isArray(data?.messages) ? data.messages : []);
    } catch (error) {
      message.error(error.message || '读取消息失败');
    } finally {
      setMessagesLoading(false);
    }
  }, []);

  useEffect(() => {
    if (open) void loadWorkspaces();
  }, [open, loadWorkspaces]);

  useEffect(() => {
    if (!open) return undefined;
    const onKeyDown = event => {
      if (event.key === 'Escape') onClose?.();
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [open, onClose]);

  useEffect(() => {
    if (!open) return undefined;
    if (!workspaceId) {
      setSessions([]);
      setSessionId('');
      setMessages([]);
      return undefined;
    }
    let cancelled = false;
    const loadWorkspaceDetail = async () => {
      setSessionsLoading(true);
      try {
        const data = await listAgentProxySessions(workspaceId);
        if (cancelled) return;
        const items = Array.isArray(data?.sessions) ? data.sessions : [];
        setSessions(items);
        setSessionId(current => (current && items.some(item => item.id === current) ? current : items[0]?.id || ''));
      } catch (error) {
        if (!cancelled) message.error(error.message || '读取会话失败');
      } finally {
        if (!cancelled) setSessionsLoading(false);
      }
      try {
        const [skillData, jobData] = await Promise.all([listAgentProxySkills(workspaceId), listAgentProxyScheduledJobs(workspaceId)]);
        if (!cancelled) {
          setPlaceholderCounts({
            skills: Array.isArray(skillData?.skills) ? skillData.skills.length : 0,
            jobs: Array.isArray(jobData?.jobs) ? jobData.jobs.length : 0
          });
        }
      } catch {
        // 占位面板的数据读取失败不影响工作区和会话主流程。
        if (!cancelled) setPlaceholderCounts({ skills: 0, jobs: 0 });
      }
    };
    void loadWorkspaceDetail();
    return () => { cancelled = true; };
  }, [open, workspaceId]);

  useEffect(() => {
    if (!open) return;
    if (!sessionId) {
      setMessages([]);
      return;
    }
    void loadMessages(sessionId);
  }, [open, sessionId, loadMessages]);

  async function refreshAll() {
    setRefreshing(true);
    try {
      await loadWorkspaces(workspaceId);
      if (sessionId) await loadMessages(sessionId);
    } finally {
      setRefreshing(false);
    }
  }

  async function createWorkspace() {
    const name = workspaceDraftName.trim();
    if (!name) {
      message.warning('请先填写工作区名称');
      return;
    }
    setCreatingWorkspace(true);
    try {
      const data = await createAgentProxyWorkspace({ name, permissionMode: workspaceDraftMode });
      setWorkspaceDraftName('');
      const createdId = data?.workspace?.id || '';
      message.success('工作区已创建');
      await loadWorkspaces(createdId);
    } catch (error) {
      message.error(error.message || '新建工作区失败');
    } finally {
      setCreatingWorkspace(false);
    }
  }

  async function createSession() {
    if (!workspaceId) {
      message.warning('请先选择或新建工作区');
      return;
    }
    const title = sessionDraftTitle.trim() || `新会话 ${sessions.length + 1}`;
    setCreatingSession(true);
    try {
      const data = await createAgentProxySession(workspaceId, { title });
      setSessionDraftTitle('');
      const createdId = data?.session?.id || '';
      const list = await listAgentProxySessions(workspaceId);
      const items = Array.isArray(list?.sessions) ? list.sessions : [];
      setSessions(items);
      setSessionId(items.some(item => item.id === createdId) ? createdId : items[0]?.id || '');
      message.success('会话已创建');
    } catch (error) {
      message.error(error.message || '新建会话失败');
    } finally {
      setCreatingSession(false);
    }
  }

  async function switchPermission(nextMode) {
    if (!activeSession || activeSession.permissionMode === nextMode) return;
    setPermissionBusy(true);
    try {
      const data = await updateAgentProxySession(activeSession.id, { permissionMode: nextMode });
      const updated = data?.session;
      if (updated) {
        setSessions(current => current.map(item => (item.id === updated.id ? { ...item, ...updated } : item)));
      }
      message.success(`权限档位已切换为 ${permissionLabel(nextMode)}`);
    } catch (error) {
      message.error(error.message || '切换权限档位失败');
    } finally {
      setPermissionBusy(false);
    }
  }

  async function sendMessage() {
    const content = draft.trim();
    if (!activeSession || !content) return;
    setSending(true);
    try {
      await appendAgentProxyMessage(activeSession.id, { role: 'user', content });
      setDraft('');
      await loadMessages(activeSession.id);
      const list = await listAgentProxySessions(activeSession.workspaceId);
      setSessions(Array.isArray(list?.sessions) ? list.sessions : []);
    } catch (error) {
      message.error(error.message || '发送消息失败');
    } finally {
      setSending(false);
    }
  }

  if (!open) return null;

  return (
    <div className="agent-proxy-overlay" role="dialog" aria-modal="true" aria-label="Agent 代理">
      <header className="agent-proxy-header">
        <div className="agent-proxy-header-title">
          <strong>Agent 代理</strong>
          <span>工作区 · 会话 · 消息（最小骨架，消息只写库，暂不触发 AI 回复）</span>
        </div>
        <div className="agent-proxy-header-actions">
          <Button size="small" loading={refreshing} onClick={refreshAll}>刷新</Button>
          <Button size="small" type="primary" onClick={onClose}>返回创作画布</Button>
        </div>
      </header>

      <div className="agent-proxy-body">
        <section className="agent-proxy-column agent-proxy-column--workspaces" aria-label="工作区">
          <div className="agent-proxy-column-head">
            <strong>工作区</strong>
            <span>{workspaces.length} 个</span>
          </div>
          <div className="agent-proxy-create">
            <Input size="small" value={workspaceDraftName} maxLength={60} placeholder="工作区名称" onChange={event => setWorkspaceDraftName(event.target.value)} onPressEnter={createWorkspace} />
            <Segmented size="small" block value={workspaceDraftMode} onChange={setWorkspaceDraftMode} options={PERMISSION_OPTIONS} />
            <Button size="small" type="primary" block loading={creatingWorkspace} onClick={createWorkspace}>新建工作区</Button>
            <p className="agent-proxy-hint">plan 只读建议，auto 可直接执行；bypass 档位未开放。</p>
          </div>
          <div className="agent-proxy-list">
            {workspacesLoading ? <p className="agent-proxy-empty">正在读取工作区…</p> : null}
            {!workspacesLoading && workspaces.length === 0 ? <p className="agent-proxy-empty">还没有工作区，先在上方新建一个。</p> : null}
            {workspaces.map(item => (
              <button key={item.id} type="button" className={`agent-proxy-item${item.id === workspaceId ? ' is-active' : ''}`} onClick={() => setWorkspaceId(item.id)}>
                <span className="agent-proxy-item-title">{item.name}</span>
                <span className="agent-proxy-item-meta">
                  <Tag bordered={false}>{permissionLabel(item.permissionMode)}</Tag>
                  <span>{item.sessionCount} 个会话</span>
                </span>
              </button>
            ))}
          </div>
        </section>

        <section className="agent-proxy-column agent-proxy-column--sessions" aria-label="会话">
          <div className="agent-proxy-column-head">
            <strong>会话</strong>
            <span>{activeWorkspace ? activeWorkspace.name : '未选择工作区'}</span>
          </div>
          <div className="agent-proxy-create">
            <Input size="small" value={sessionDraftTitle} maxLength={60} placeholder="会话标题（留空自动生成）" disabled={!workspaceId} onChange={event => setSessionDraftTitle(event.target.value)} onPressEnter={createSession} />
            <Button size="small" block loading={creatingSession} disabled={!workspaceId} onClick={createSession}>新建会话</Button>
            <p className="agent-proxy-hint">新会话继承工作区档位，可在右侧随时切换。</p>
          </div>
          <div className="agent-proxy-list">
            {sessionsLoading ? <p className="agent-proxy-empty">正在读取会话…</p> : null}
            {!sessionsLoading && workspaceId && sessions.length === 0 ? <p className="agent-proxy-empty">这个工作区还没有会话。</p> : null}
            {!workspaceId ? <p className="agent-proxy-empty">先在左侧选择或新建工作区。</p> : null}
            {sessions.map(item => (
              <button key={item.id} type="button" className={`agent-proxy-item${item.id === sessionId ? ' is-active' : ''}`} onClick={() => setSessionId(item.id)}>
                <span className="agent-proxy-item-title">{item.title}</span>
                <span className="agent-proxy-item-meta">
                  <Tag bordered={false}>{permissionLabel(item.permissionMode)}</Tag>
                  <span>{item.messageCount} 条消息</span>
                  {item.lastMessageAt ? <span>{formatTime(item.lastMessageAt)}</span> : null}
                </span>
              </button>
            ))}
          </div>
          <div className="agent-proxy-placeholders">
            <article className="agent-proxy-placeholder">
              <header><strong>技能</strong><Tag bordered={false}>尚未接入</Tag></header>
              <p>技能包解析与技能加载未实现，当前只保留数据结构与接口位置。</p>
              <small>已记录 {placeholderCounts.skills} 项</small>
            </article>
            <article className="agent-proxy-placeholder">
              <header><strong>定时任务</strong><Tag bordered={false}>尚未接入</Tag></header>
              <p>第一版只提供列表查询，不创建、不调度、不执行。</p>
              <small>已记录 {placeholderCounts.jobs} 项</small>
            </article>
          </div>
        </section>

        <section className="agent-proxy-column agent-proxy-column--conversation" aria-label="消息">
          <div className="agent-proxy-column-head agent-proxy-column-head--conversation">
            <div className="agent-proxy-conversation-title">
              <strong>{activeSession ? activeSession.title : '未选择会话'}</strong>
              <span>{activeSession ? `${activeSession.messageCount} 条消息` : '选择会话后查看消息'}</span>
            </div>
            {activeSession ? (
              <div className="agent-proxy-permission">
                <span>权限档位</span>
                <Segmented size="small" value={activeSession.permissionMode} disabled={permissionBusy} onChange={switchPermission} options={PERMISSION_OPTIONS} />
              </div>
            ) : null}
          </div>
          <div className="agent-proxy-messages">
            {!activeSession ? <p className="agent-proxy-empty">先在中间选择或新建一个会话。</p> : null}
            {activeSession && messagesLoading ? <p className="agent-proxy-empty">正在读取消息…</p> : null}
            {activeSession && !messagesLoading && messages.length === 0 ? <p className="agent-proxy-empty">这个会话还没有消息。</p> : null}
            {messages.map(item => (
              <article key={item.id} className={`agent-proxy-message is-${item.role}`}>
                <header>
                  <strong>{ROLE_LABELS[item.role] || item.role}</strong>
                  <time>{formatTime(item.createdAt)}</time>
                </header>
                <p>{item.content}</p>
              </article>
            ))}
          </div>
          <div className="agent-proxy-composer">
            <Input.TextArea
              value={draft}
              autoSize={{ minRows: 2, maxRows: 6 }}
              maxLength={4000}
              disabled={!activeSession || sending}
              placeholder="输入一条消息，写入当前会话…（Enter 发送，Shift+Enter 换行）"
              onChange={event => setDraft(event.target.value)}
              onPressEnter={event => {
                if (!event.shiftKey) {
                  event.preventDefault();
                  void sendMessage();
                }
              }}
            />
            <div className="agent-proxy-composer-actions">
              <span>{activeSession?.permissionMode === 'plan' ? '当前 plan 档位：消息仅作为建议记录，不会执行。' : '当前 auto 档位：后续可接入直接执行。'}</span>
              <Button type="primary" loading={sending} disabled={!activeSession || !draft.trim()} onClick={sendMessage}>发送</Button>
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}
