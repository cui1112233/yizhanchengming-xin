import { Button, Dropdown, Empty, Input, Select, Switch, Tooltip, Typography, message as antMessage } from 'antd';
import {
  Archive,
  Bot,
  ChevronDown,
  ChevronRight,
  Clock3,
  File as FileIcon,
  FolderOpen,
  FolderTree,
  MessageCircle,
  Plus,
  RefreshCw,
  Search,
  Send,
  Square,
  Trash2,
  X
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  clearAgentChatMessages,
  createAgentChatSession,
  createAgentChatWorkspace,
  deleteAgentChatSession,
  deleteAgentChatWorkspace,
  getAgentChatMessages,
  listAgentChatModels,
  listAgentChatPermissions,
  listAgentChatSessions,
  listAgentChatTree,
  listAgentChatWorkspaces,
  readAgentChatFile,
  stopAgentChatRun,
  streamAgentChatMessage,
  updateAgentChatSession,
  updateAgentChatWorkspace
} from '../../shared/api/agentChat';

const NEW_SESSION_TITLE = '新 Agent 会话';
const DEFAULT_PERMISSION = 'plan';
const normalizePermission = value => {
  const mode = String(value || '').trim().toLowerCase();
  return ['plan', 'auto', 'bypass'].includes(mode) ? mode : DEFAULT_PERMISSION;
};
const COMPOSER_PLACEHOLDER = '给 Agent 分配任务，输入 @ 引用文件，/ 调用技能，$ 调起命令…';

function formatSessionTime(value) {
  if (!value) return '';
  const date = new Date(String(value).replace(' ', 'T'));
  if (Number.isNaN(date.getTime())) return '';
  return new Intl.DateTimeFormat('zh-CN', { month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit' }).format(date);
}

function statusDotClass(status) {
  if (status === 'done') return 'is-done';
  if (status === 'failed') return 'is-failed';
  if (status === 'running') return 'is-running';
  return 'is-planned';
}

function ToolBlockList({ blocks }) {
  if (!blocks || !blocks.length) return null;
  return (
    <div className="agent-chat-tools">
      {blocks.map(block => (
        <div key={block.id || `${block.toolName}-${block.startedAt}`} className="agent-chat-tool">
          <span className={`agent-chat-tool-dot ${statusDotClass(block.status)}`} />
          <div className="agent-chat-tool-body">
            <div className="agent-chat-tool-head">
              <strong>{block.name || block.toolName}</strong>
              <span>{block.message || (block.status === 'planned' ? '已列入计划' : '')}</span>
            </div>
            {block.error ? <p className="agent-chat-tool-error">{block.error}</p> : null}
            {!block.error && block.preview ? <pre className="agent-chat-tool-preview">{block.preview}</pre> : null}
          </div>
        </div>
      ))}
    </div>
  );
}

function ChatMessage({ item }) {
  const isUser = item.role === 'user';
  return (
    <article className={`agent-chat-message ${isUser ? 'is-user' : 'is-assistant'}`}>
      <header>
        <span className="agent-chat-avatar">{isUser ? '你' : <Bot size={14} aria-hidden="true" />}</span>
        <strong>{isUser ? '你' : 'Agent'}</strong>
        <time>{item.timestamp || ''}</time>
      </header>
      <ToolBlockList blocks={item.toolBlocks} />
      {item.content ? <div className="agent-chat-content">{item.content}</div> : null}
      {item.status === 'thinking' ? <div className="agent-chat-thinking">正在执行中…</div> : null}
      {item.canContinue ? <div className="agent-chat-hint">已达到本轮工具调用上限，可发送「继续」接着执行。</div> : null}
    </article>
  );
}

export function AgentChatPage() {
  const [workspaces, setWorkspaces] = useState([]);
  const [sessions, setSessions] = useState([]);
  const [activeWorkspaceId, setActiveWorkspaceId] = useState('');
  const [activeSessionId, setActiveSessionId] = useState(null);
  const [activeSession, setActiveSession] = useState(null);
  const [items, setItems] = useState([]);
  const [draft, setDraft] = useState('');
  const [permissions, setPermissions] = useState([]);
  const [models, setModels] = useState([]);
  const [model, setModel] = useState('');
  const [permission, setPermission] = useState(DEFAULT_PERMISSION);
  const [streaming, setStreaming] = useState(false);
  const [error, setError] = useState('');
  const [search, setSearch] = useState('');
  const [autoContinue, setAutoContinue] = useState(false);
  const [filesOpen, setFilesOpen] = useState(true);
  const [entries, setEntries] = useState([]);
  const [openDirs, setOpenDirs] = useState({});
  const [fileView, setFileView] = useState(null);
  const [collapsedProjects, setCollapsedProjects] = useState({});
  const [loading, setLoading] = useState(true);
  const abortRef = useRef(null);
  const autoContinueRef = useRef(false);
  const scrollRef = useRef(null);

  autoContinueRef.current = autoContinue;

  const activeWorkspace = useMemo(
    () => workspaces.find(workspace => workspace.id === activeWorkspaceId) || null,
    [workspaces, activeWorkspaceId]
  );

  const loadWorkspaces = useCallback(async () => {
    const payload = await listAgentChatWorkspaces();
    const list = Array.isArray(payload?.workspaces) ? payload.workspaces : [];
    setWorkspaces(list);
    setActiveWorkspaceId(current => current || list.find(workspace => !workspace.archived)?.id || '');
    return list;
  }, []);

  const loadSessions = useCallback(async () => {
    const payload = await listAgentChatSessions('');
    setSessions(Array.isArray(payload?.sessions) ? payload.sessions : []);
  }, []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const [, permissionPayload, modelPayload] = await Promise.all([
          loadWorkspaces(),
          listAgentChatPermissions(),
          listAgentChatModels()
        ]);
        if (cancelled) return;
        setPermissions(Array.isArray(permissionPayload?.options) ? permissionPayload.options : []);
        const modelList = Array.isArray(modelPayload?.models) ? modelPayload.models : [];
        setModels(modelList);
        setModel(modelList[0]?.id || '');
        await loadSessions();
      } catch (failure) {
        if (!cancelled) setError(failure?.message || '请求失败');
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => { cancelled = true; };
  }, [loadWorkspaces, loadSessions]);

  useEffect(() => {
    if (!activeWorkspaceId) return;
    let cancelled = false;
    (async () => {
      try {
        const payload = await listAgentChatTree(activeWorkspaceId, '');
        if (!cancelled) setEntries(Array.isArray(payload?.entries) ? payload.entries : []);
      } catch (failure) {
        if (!cancelled) setError(failure?.message || '获取文件树失败');
      }
    })();
    return () => { cancelled = true; };
  }, [activeWorkspaceId]);

  useEffect(() => {
    if (!scrollRef.current) return;
    scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
  }, [items]);

  const openSession = useCallback(async sessionId => {
    setActiveSessionId(sessionId);
    setError('');
    try {
      const payload = await getAgentChatMessages(sessionId);
      setActiveSession(payload?.session || null);
      setItems(Array.isArray(payload?.messages) ? payload.messages : []);
      if (payload?.session?.model) setModel(payload.session.model);
      if (payload?.session?.permission_mode) setPermission(normalizePermission(payload.session.permission_mode));
      if (payload?.session?.workspace_id) setActiveWorkspaceId(payload.session.workspace_id);
    } catch (failure) {
      setError(failure?.message || '加载 Agent 会话失败:');
    }
  }, []);

  const createSession = useCallback(async () => {
    if (!activeWorkspaceId) {
      antMessage.warning('请先选择工作区');
      return;
    }
    try {
      const payload = await createAgentChatSession({
        title: NEW_SESSION_TITLE,
        workspaceId: activeWorkspaceId,
        model,
        permission
      });
      const session = payload?.session;
      if (!session) return;
      await loadSessions();
      setActiveSessionId(session.id);
      setActiveSession(session);
      setItems([]);
    } catch (failure) {
      antMessage.error(failure?.message || '创建会话失败');
    }
  }, [activeWorkspaceId, loadSessions, model, permission]);

  const changePermission = useCallback(async next => {
    setPermission(next);
    if (!activeSessionId) return;
    try {
      const payload = await updateAgentChatSession(activeSessionId, { permission: next });
      if (payload?.session) setActiveSession(payload.session);
    } catch (_) {
      // the choice still applies to the next send
    }
  }, [activeSessionId]);

  const patchAssistant = useCallback(updater => {
    setItems(current => {
      if (!current.length) return current;
      const next = current.slice();
      const last = next[next.length - 1];
      if (!last || last.role !== 'assistant') return current;
      next[next.length - 1] = updater(last);
      return next;
    });
  }, []);

  const upsertToolBlock = useCallback(block => {
    patchAssistant(assistant => {
      const blocks = Array.isArray(assistant.toolBlocks) ? assistant.toolBlocks.slice() : [];
      const index = blocks.findIndex(entry => entry.id === block.id);
      if (index >= 0) blocks[index] = { ...blocks[index], ...block };
      else blocks.push(block);
      return { ...assistant, toolBlocks: blocks };
    });
  }, [patchAssistant]);

  const send = useCallback(async (rawText, options = {}) => {
    const content = String(rawText ?? '').trim();
    if (!content || streaming || !activeSessionId) return;
    setDraft('');
    setError('');
    const now = Date.now();
    const userMessage = { id: now, role: 'user', content, timestamp: '', toolBlocks: [] };
    const assistantMessage = { id: now + 1, role: 'assistant', content: '', timestamp: '', toolBlocks: [], status: 'thinking' };
    setItems(current => [...current, userMessage, assistantMessage]);
    setStreaming(true);
    const controller = new AbortController();
    abortRef.current = controller;
    let finished = null;
    try {
      await streamAgentChatMessage(activeSessionId, {
        content,
        model,
        permission
      }, {
        agent_status: payload => {
          if (payload?.text) patchAssistant(assistant => ({ ...assistant, status: 'thinking' }));
        },
        tool_start: block => upsertToolBlock(block),
        tool_progress: payload => upsertToolBlock({ id: payload?.id, message: payload?.message }),
        tool_done: block => upsertToolBlock(block),
        token: payload => patchAssistant(assistant => ({ ...assistant, content: `${assistant.content || ''}${payload?.text || ''}` })),
        error: payload => setError(payload?.message || '请求失败'),
        done: payload => {
          finished = payload?.message || null;
          if (finished) {
            patchAssistant(() => ({ ...finished, status: 'idle' }));
          } else {
            patchAssistant(assistant => ({ ...assistant, status: 'idle' }));
          }
        }
      }, controller.signal);
    } catch (failure) {
      if (failure?.name !== 'AbortError') setError(failure?.message || '请求失败');
    } finally {
      setStreaming(false);
      abortRef.current = null;
      await loadSessions().catch(() => {});
    }
    if (!options.silent && autoContinueRef.current && finished?.canContinue) {
      send('继续', { silent: true });
    }
  }, [activeSessionId, loadSessions, model, patchAssistant, permission, streaming, upsertToolBlock]);

  const stop = useCallback(async () => {
    if (!activeSessionId) return;
    try {
      await stopAgentChatRun(activeSessionId);
    } finally {
      abortRef.current?.abort();
    }
  }, [activeSessionId]);

  const clearContext = useCallback(async () => {
    if (!activeSessionId) return;
    try {
      await clearAgentChatMessages(activeSessionId);
      setItems([]);
    } catch (failure) {
      antMessage.error(failure?.message || '请求失败');
    }
  }, [activeSessionId]);

  const removeSession = useCallback(async sessionId => {
    try {
      await deleteAgentChatSession(sessionId);
      if (activeSessionId === sessionId) {
        setActiveSessionId(null);
        setActiveSession(null);
        setItems([]);
      }
      await loadSessions();
    } catch (failure) {
      antMessage.error(failure?.message || '请求失败');
    }
  }, [activeSessionId, loadSessions]);

  const addWorkspace = useCallback(async () => {
    // eslint-disable-next-line no-alert
    const name = window.prompt('工作区名称');
    if (!name || !name.trim()) return;
    try {
      const payload = await createAgentChatWorkspace(name.trim());
      await loadWorkspaces();
      if (payload?.workspace?.id) setActiveWorkspaceId(payload.workspace.id);
    } catch (failure) {
      antMessage.error(failure?.message || '创建工作区失败');
    }
  }, [loadWorkspaces]);

  const toggleArchive = useCallback(async workspace => {
    try {
      await updateAgentChatWorkspace(workspace.id, { archived: !workspace.archived });
      await loadWorkspaces();
    } catch (failure) {
      antMessage.error(failure?.message || '更新工作区失败');
    }
  }, [loadWorkspaces]);

  const removeWorkspace = useCallback(async workspace => {
    try {
      await deleteAgentChatWorkspace(workspace.id);
      await loadWorkspaces();
    } catch (failure) {
      antMessage.error(failure?.message || '删除工作区失败');
    }
  }, [loadWorkspaces]);

  const toggleDir = useCallback(async entry => {
    setOpenDirs(current => ({ ...current, [entry.path]: !current[entry.path] }));
    if (openDirs[entry.path]) return;
    try {
      const payload = await listAgentChatTree(activeWorkspaceId, entry.path);
      setEntries(current => {
        const others = current.filter(item => !String(item.path).startsWith(`${entry.path}/`));
        return [...others, ...(payload?.entries || [])];
      });
    } catch (failure) {
      antMessage.error(failure?.message || '获取文件树失败');
    }
  }, [activeWorkspaceId, openDirs]);

  const openFile = useCallback(async entry => {
    try {
      const payload = await readAgentChatFile(activeWorkspaceId, entry.path);
      setFileView(payload);
    } catch (failure) {
      antMessage.error(failure?.message || '读取文件失败');
    }
  }, [activeWorkspaceId]);

  const filteredSessions = useMemo(() => {
    const keyword = search.trim().toLowerCase();
    if (!keyword) return sessions;
    return sessions.filter(session => String(session.title || '').toLowerCase().includes(keyword));
  }, [search, sessions]);

  const activeProjects = workspaces.filter(workspace => !workspace.archived);
  const archivedProjects = workspaces.filter(workspace => workspace.archived);
  const permissionOptions = permissions.length
    ? permissions
    : [
      { id: 'plan', label: '计划模式' },
      { id: 'auto', label: '自动审批' },
      { id: 'bypass', label: '完全运行' }
    ];
  const currentPermission = permissionOptions.find(option => option.id === permission) || permissionOptions[0];
  const fileEntries = entries.filter(entry => entry.path.split('/').length === 1 || openDirs[entry.path.split('/').slice(0, -1).join('/')]);

  const renderProject = workspace => {
    const workspaceSessions = filteredSessions.filter(session => session.workspace_id === workspace.id);
    const collapsed = collapsedProjects[workspace.id];
    return (
      <div key={workspace.id} className={`agent-chat-project ${activeWorkspaceId === workspace.id ? 'is-active' : ''}`}>
        <div className="agent-chat-project-head">
          <button
            type="button"
            className="agent-chat-project-toggle"
            onClick={() => setCollapsedProjects(current => ({ ...current, [workspace.id]: !current[workspace.id] }))}
          >
            {collapsed ? <ChevronRight size={14} aria-hidden="true" /> : <ChevronDown size={14} aria-hidden="true" />}
            <FolderOpen size={14} aria-hidden="true" />
            <span>{workspace.name}</span>
          </button>
          <Dropdown
            menu={{
              items: [
                { key: 'archive', label: workspace.archived ? '取消归档' : '归档' },
                { key: 'delete', label: '删除工作区', danger: true, disabled: workspace.is_default }
              ],
              onClick: ({ key }) => {
                if (key === 'archive') toggleArchive(workspace);
                if (key === 'delete') removeWorkspace(workspace);
              }
            }}
            trigger={['click']}
          >
            <button type="button" className="agent-chat-project-more" aria-label="工作区操作">⋯</button>
          </Dropdown>
        </div>
        {collapsed ? null : (
          <div className="agent-chat-project-sessions">
            {workspaceSessions.map(session => (
              <div key={session.id} className={`agent-chat-session ${activeSessionId === session.id ? 'is-active' : ''}`}>
                <button type="button" className="agent-chat-session-main" onClick={() => openSession(session.id)}>
                  <MessageCircle size={13} aria-hidden="true" />
                  <span className="agent-chat-session-title">{session.title}</span>
                  <time>{formatSessionTime(session.updated_at)}</time>
                </button>
                <Tooltip title="永久删除">
                  <button type="button" className="agent-chat-session-delete" onClick={() => removeSession(session.id)} aria-label="删除会话">
                    <Trash2 size={12} aria-hidden="true" />
                  </button>
                </Tooltip>
              </div>
            ))}
            {!workspaceSessions.length ? <p className="agent-chat-empty-line">还没有会话</p> : null}
            <button
              type="button"
              className="agent-chat-project-new"
              onClick={() => { setActiveWorkspaceId(workspace.id); createSession(); }}
            >
              <Plus size={13} aria-hidden="true" /> 新 Agent 会话
            </button>
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="agent-chat-shell">
      <aside className="agent-chat-rail" aria-label="Agent 会话">
        <div className="agent-chat-rail-top">
          <Button type="primary" block icon={<Plus size={14} />} onClick={createSession}>新 Agent 会话</Button>
          <Input
            allowClear
            prefix={<Search size={13} />}
            placeholder="搜索会话"
            value={search}
            onChange={event => setSearch(event.target.value)}
          />
        </div>
        <div className="agent-chat-rail-scroll">
          <section className="agent-chat-rail-section">
            <header><Clock3 size={13} aria-hidden="true" /><span>自动任务</span></header>
            <p className="agent-chat-empty-line">还没有自动任务</p>
          </section>
          <section className="agent-chat-rail-section">
            <header><Bot size={13} aria-hidden="true" /><span>Agent 技能</span></header>
            <a className="agent-chat-rail-link" href="/agent">管理 Agent 技能</a>
          </section>
          <section className="agent-chat-rail-section">
            <header>
              <FolderTree size={13} aria-hidden="true" /><span>项目</span>
              <button type="button" className="agent-chat-rail-action" onClick={addWorkspace}>添加工作区</button>
            </header>
            {loading ? <p className="agent-chat-empty-line">正在读取工作区…</p> : null}
            {!loading && !activeProjects.length ? <p className="agent-chat-empty-line">还没有工作区</p> : null}
            {activeProjects.map(renderProject)}
          </section>
          <section className="agent-chat-rail-section">
            <header><Archive size={13} aria-hidden="true" /><span>已归档</span><span className="agent-chat-count">{archivedProjects.length}</span></header>
            {archivedProjects.map(renderProject)}
          </section>
        </div>
      </aside>

      <main className="agent-chat-main">
        <header className="agent-chat-topbar">
          <div>
            <strong>{activeSession?.title || NEW_SESSION_TITLE}</strong>
            <span>{activeWorkspace?.name || '未选择工作区'}</span>
          </div>
          <div className="agent-chat-topbar-actions">
            <span className="agent-chat-permission-tag">{currentPermission.label}</span>
            <Tooltip title={filesOpen ? '收起文件面板' : '展开文件面板'}>
              <Button size="small" icon={<FolderTree size={14} />} onClick={() => setFilesOpen(value => !value)} />
            </Tooltip>
          </div>
        </header>

        {error ? <div className="agent-chat-error">{error}</div> : null}

        <div className="agent-chat-stream" ref={scrollRef}>
          {!activeSessionId ? (
            <Empty description="请选择左侧工作区，或新建一个 Agent 会话" image={Empty.PRESENTED_IMAGE_SIMPLE} />
          ) : null}
          {activeSessionId && !items.length ? (
            <div className="agent-chat-empty-state">
              <Bot size={22} aria-hidden="true" />
              <p>{COMPOSER_PLACEHOLDER}</p>
            </div>
          ) : null}
          {items.map(item => <ChatMessage key={item.id} item={item} />)}
        </div>

        <footer className="agent-chat-composer">
          <textarea
            value={draft}
            disabled={!activeSessionId}
            placeholder={COMPOSER_PLACEHOLDER}
            onChange={event => setDraft(event.target.value)}
            onKeyDown={event => {
              if (event.key === 'Enter' && !event.shiftKey) {
                event.preventDefault();
                send(draft);
              }
            }}
          />
          <div className="agent-chat-composer-bar">
            <div className="agent-chat-composer-left">
              <Button size="small" type="text" icon={<Plus size={14} />} disabled={!activeSessionId}>添加附件</Button>
              <button type="button" className="agent-chat-text-action" onClick={clearContext} disabled={!activeSessionId}>清除上下文</button>
              <label className="agent-chat-switch">
                <span>自动续接</span>
                <Switch size="small" checked={autoContinue} onChange={setAutoContinue} />
              </label>
              <Select
                size="small"
                value={permission}
                onChange={changePermission}
                className="agent-chat-permission-select"
                options={permissionOptions.map(option => ({ value: option.id, label: option.label }))}
              />
            </div>
            <div className="agent-chat-composer-right">
              <Select
                size="small"
                value={model || undefined}
                placeholder="选择模型"
                className="agent-chat-model-select"
                options={models.map(entry => ({ value: entry.id, label: entry.name || entry.id }))}
                onChange={setModel}
              />
              {streaming ? (
                <Button danger type="primary" icon={<Square size={13} />} onClick={stop}>停止</Button>
              ) : (
                <Button type="primary" icon={<Send size={13} />} disabled={!activeSessionId || !draft.trim()} onClick={() => send(draft)}>发送</Button>
              )}
            </div>
          </div>
        </footer>
      </main>

      {filesOpen ? (
        <aside className="agent-chat-files" aria-label="工作区文件">
          <header>
            <div><strong>工作区文件</strong><span>{activeWorkspace?.name || '未选择工作区'}</span></div>
            <div>
              <Tooltip title="刷新">
                <button
                  type="button"
                  aria-label="刷新"
                  onClick={async () => {
                    if (!activeWorkspaceId) return;
                    const payload = await listAgentChatTree(activeWorkspaceId, '');
                    setEntries(Array.isArray(payload?.entries) ? payload.entries : []);
                  }}
                >
                  <RefreshCw size={13} aria-hidden="true" />
                </button>
              </Tooltip>
              <button type="button" aria-label="关闭" onClick={() => setFilesOpen(false)}><X size={13} aria-hidden="true" /></button>
            </div>
          </header>
          {fileView ? (
            <div className="agent-chat-file-view">
              <div className="agent-chat-file-view-head">
                <Typography.Text ellipsis>{fileView.path}</Typography.Text>
                <button type="button" onClick={() => setFileView(null)} aria-label="返回文件列表"><X size={12} aria-hidden="true" /></button>
              </div>
              <pre>{fileView.content}</pre>
            </div>
          ) : (
            <div className="agent-chat-file-tree">
              {!fileEntries.length ? <p className="agent-chat-empty-line">当前工作区还没有文件</p> : null}
              {fileEntries.map(entry => {
                const depth = entry.path.split('/').length - 1;
                return (
                  <button
                    key={entry.path}
                    type="button"
                    className={`agent-chat-file-row ${entry.type === 'directory' ? 'is-dir' : ''}`}
                    style={{ paddingLeft: 8 + depth * 12 }}
                    onClick={() => (entry.type === 'directory' ? toggleDir(entry) : openFile(entry))}
                  >
                    {entry.type === 'directory'
                      ? (openDirs[entry.path] ? <ChevronDown size={12} /> : <ChevronRight size={12} />)
                      : <FileIcon size={12} />}
                    <span>{entry.name}</span>
                  </button>
                );
              })}
            </div>
          )}
        </aside>
      ) : null}
    </div>
  );
}

export default AgentChatPage;
