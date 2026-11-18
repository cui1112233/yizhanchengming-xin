import LongImitationContext from './novel-creation/LongImitationContext.jsx';
import LongImitationDialog from './novel-creation/LongImitationDialog.jsx';
import ComposerToolbar from './novel-creation/ComposerToolbar.jsx';
import ShortStoryDialog from './novel-creation/ShortStoryDialog.jsx';
import ShortStoryPanel from './novel-creation/ShortStoryPanel.jsx';
import { Button, Drawer, Empty, Form, Input, InputNumber, Modal, Select, Spin, Tabs, message } from 'antd';
import { Archive, BookOpen, Bot, CalendarClock, ChartColumn, FileText, FilePlus2, CornerDownRight, Pencil, NotebookPen, Plus } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { getCurrentAccount, getCurrentUsername } from '../../shared/api/auth.js';
import { canManageModelCatalog, getConfig } from '../../shared/api/config.js';
import { listAvailableModels, modelSelectOptions } from '../../shared/api/modelCatalog.js';
import { getNovelCreationSessionSnapshot, getNovelCreationRun, getNovelCreationProjectChapters, novelProjectRequest, novelRequest } from '../../shared/api/novelCreation.js';
import { getProjectRetention, indexRetention } from '../../shared/api/projectRetention.js';
import AgentPanel from './novel-creation/AgentPanel.jsx';
import AutomationPanel from './novel-creation/AutomationPanel.jsx';
import MaterialsPanel from './novel-creation/MaterialsPanel.jsx';
import NotesPanel from './novel-creation/NotesPanel.jsx';
import ChaptersPanel from './novel-creation/ChaptersPanel.jsx';
import RunCard from './novel-creation/RunCard.jsx';
import BlueprintEditor from './novel-creation/BlueprintEditor.jsx';
import StudioChat from './novel-creation/StudioChat.jsx';
import ContinuationImportDialog from './novel-creation/ContinuationImportDialog.jsx';
import useWorkspaceScope from './novel-creation/useWorkspaceScope.js';
import { MODE_FIELDS, MODES, STATUS, TERMINAL, availableModel, draftKey, mergeSnapshot, modelOptions, readSaved, validRun } from './novel-creation/contracts.js';
import './NovelCreationPage.css';

const TOOLS = [['写作风格', FileText], ['导入素材', BookOpen], ['Agent设置', Bot], ['定时写作', CalendarClock], ['数据看板', ChartColumn], ['备忘录', NotebookPen]];
const creationIcons = { long_original: FileText, short_original: FileText, long_imitation: Pencil, short_imitation: Pencil, fanfiction: CornerDownRight, continuation: FilePlus2 };
const width = value => Math.max(360, Math.min(800, Number(value) || Math.round(window.innerWidth * .38)));

export default function NovelCreationPage() {
  const [owner, setOwner] = useState(getCurrentUsername), [account, setAccount] = useState(null);
  const [projects, setProjects] = useState([]), [retention, setRetention] = useState({}), [models, setModels] = useState([]);
  const [projectId, setProjectId] = useState(''), [activeMode, setActiveMode] = useState(() => localStorage.getItem('novel:mode:' + getCurrentUsername()) || 'long_original');
  const [studioSessions,setStudioSessions]=useState([]),[studioSessionId,setStudioSessionId]=useState('');
  const [sessions, setSessions] = useState([]), [sessionId, setSessionId] = useState(''), [snapshot, setSnapshot] = useState(null);
  const [snapshotRead, setSnapshotRead] = useState(null);
  const [loading, setLoading] = useState(true), [connection, setConnection] = useState(''), [tool, setTool] = useState('');
  const [importOpen, setImportOpen] = useState(false), [memoOpen, setMemoOpen] = useState(false);
  const [longImitationOpen, setLongImitationOpen] = useState(false), [longImitationEntry, setLongImitationEntry] = useState(0);
  const [longImitationCreation,setLongImitationCreation]=useState(null);
  const [shortDialog, setShortDialog] = useState(null), [shortStoryData, setShortStoryData] = useState(null);
  const [continuationOpen, setContinuationOpen] = useState(false), [projectChapters, setProjectChapters] = useState(null), [sessionsLoadedKey, setSessionsLoadedKey] = useState('');
  const memoPopup = useRef(null);
  const [projectModal, setProjectModal] = useState(false), [manage, setManage] = useState(null), [archiveOpen, setArchiveOpen] = useState(false);
  const [blueprint, setBlueprint] = useState(null), [batchOpen, setBatchOpen] = useState(false), [dashboard, setDashboard] = useState(null);
  const [inspectorWidth, setInspectorWidth] = useState(() => width(localStorage.getItem('novel:inspector-width')));
  const [draftState, setDraftState] = useState(null), [sendKey, setSendKey] = useState(''), [actionBusy, setActionBusy] = useState(false);
  const [references, setReferences] = useState([]), [refresh, setRefresh] = useState(0), [, setIntentRevision] = useState(0); const [projectForm] = Form.useForm();
  const flight = useRef(null), resize = useRef(false), composition = useRef(false), listSequence = useRef(0), snapshotSequence = useRef(0);
  const ownerScope = useWorkspaceScope(owner), listScope = useWorkspaceScope(`${owner}:${projectId}:${activeMode}`);
  const scopeKey = `${owner}:${projectId}:${sessionId}:${activeMode}`, scope = useWorkspaceScope(scopeKey);
  const projectReadKey = `${owner}:${projectId}`, projectReadScope = useWorkspaceScope(projectReadKey);
  const shortStory = shortStoryData?.key === projectReadKey ? shortStoryData.story : null;
  const chapterRevisions = useRef(new Map()), currentView = useRef(), studioNavigation = useRef({key:'',epoch:0});
  const studioViewKey=JSON.stringify([owner,projectId,activeMode,studioSessionId,tool]);
  if(studioNavigation.current.key!==studioViewKey)studioNavigation.current={key:studioViewKey,epoch:studioNavigation.current.epoch+1};
  currentView.current = { owner, projectId, scopeKey, studioEpoch:studioNavigation.current.epoch };
  const chapterRevision = chapterRevisions.current.get(projectReadKey) || 0;
  const hasLongImitationCreation=longImitationCreation?.key===projectReadKey && !!longImitationCreation.creation;
  const key = draftKey(owner, projectId, sessionId, activeMode), project = projects.find(p => p.id === projectId);
  const defaultDraft = { content: '', modeInputs: {}, knowledgeIds: [], noteIds: [], styleProfileId: '', modelId: project?.modelId || models[0]?.value || '', thinkingLevel: 'standard' };
  const draft = draftState?.key === key ? draftState.value : { ...defaultDraft, ...readSaved(key) };
  const chapterRead = sessionId ? snapshotRead : projectChapters;
  const chaptersReady = chapterRead?.key === (sessionId ? scopeKey : projectReadKey) && chapterRead.revision === chapterRevision && chapterRead.status === 'ready';
  const run = validRun(snapshot?.run), chapters = chaptersReady ? (sessionId ? snapshot?.chapters || [] : projectChapters.chapters) : [], currentSession = sessions.find(s => s.id === sessionId);
  const archived = !!currentSession?.archivedAt || project?.status === 'archived' || project?.status === 'purging' || !!retention[`novel_creation:${projectId}`]?.archivedAt;
  const active = !!run && !TERMINAL.has(run.status), sending = sendKey === key;
  const intent = readSaved(key + ':intent', null);
  const maintenanceIntent = readSaved(key + ':maintenance', null);
  function changeDraft(patch) { const next = { ...draft, ...readSaved(key), ...patch }; localStorage.setItem(key, JSON.stringify(next)); setDraftState({ key, value: next }); }
  function clearDraftAt(target) { localStorage.removeItem(target); if (target === key) setDraftState({ key, value: { ...defaultDraft, content: '' } }); }
  function error(e) { if (scope.isCurrent() && !e?.aborted) message.error(e.message || '操作未完成，请稍后重试'); }
  function recordImportedBook(result) {
    if (!ownerScope.isCurrent()) return;
    const importedKey = `${owner}:${result.book.id}`, revision = (chapterRevisions.current.get(importedKey) || 0) + 1;
    chapterRevisions.current.set(importedKey, revision);
    setProjects(old => [result.book, ...old.filter(p => p.id !== result.book.id)]);
    setProjectChapters(old => old?.key === importedKey ? { key: importedKey, revision, status: 'loading', chapters: [] } : old);
    const current = currentView.current;
    if (current.owner === owner && current.projectId === result.book.id) {
      setSnapshot(null); setSnapshotRead({ key: current.scopeKey, revision, status: 'loading' }); setRefresh(v => v + 1);
    }
  }
  const request = (path, input, options = {}) => novelProjectRequest(projectId, path, input, { signal: scope.signal, ...options });
  const studioView = !projectId && activeMode === 'long_original';
  const modeName = MODES.find(([id]) => id === activeMode)?.[1] || '创作';
  useEffect(() => {
    if (!projectId) { setShortStoryData(null); return; }
    projectReadScope.call(novelProjectRequest(projectId, '/short-story', undefined, { signal: projectReadScope.signal }), result => {
      setShortStoryData({ key: projectReadKey, story: result.story });
      if (result.story?.input?.mode) setActiveMode(result.story.input.mode);
    }).catch(() => {});
  }, [projectReadKey]);

  function enterCreation(id) {
    setTool('');
    if (id === 'long_imitation') { setLongImitationEntry(n=>n+1); setLongImitationOpen(true); return; }
    if (id === 'continuation' || id === 'fanfiction') {
      setActiveMode(id);
      if (id === 'continuation') setContinuationOpen(true);
      return;
    }
    // A creation entry is not a mode switch inside whichever old book happened
    // to be selected. Keep that book accessible only through its own entry.
    setProjectId(''); setSessionId(''); setSnapshot(null); setShortStoryData(null);
    localStorage.setItem('novel:view:'+owner,'studio'); setActiveMode(id);
    if (id === 'short_imitation') setShortDialog(id);
    if (id === 'continuation') setContinuationOpen(true);
  }

  useEffect(() => {
    const check = () => { const next = getCurrentUsername(); setOwner(previous => previous === next ? previous : next); };
    window.addEventListener('storage', check); const timer = setInterval(check, 1000);
    return () => { window.removeEventListener('storage', check); clearInterval(timer); };
  }, []);
  useEffect(() => {
    setLoading(true); setShortDialog(null); setShortStoryData(null); setProjectId(''); setSessionId(''); setSnapshot(null); setSessions([]); setModels([]); setTool(''); setImportOpen(false); setMemoOpen(false); setContinuationOpen(false); setLongImitationOpen(false); setProjectChapters(null); setSessionsLoadedKey(''); setStudioSessions([]); setStudioSessionId(localStorage.getItem('novel:studio:selected:'+owner)||'');
    ownerScope.call(novelRequest('/studio-sessions',undefined,{signal:ownerScope.signal}),r=>setStudioSessions(r.sessions||[])).catch(()=>{});
    ownerScope.call(Promise.allSettled([novelRequest('/projects', undefined, { signal: ownerScope.signal }), listAvailableModels('text'), getProjectRetention(), getCurrentAccount(), getConfig()]), results => {
      if (results[0].status === 'fulfilled') { const items = results[0].value.projects || []; setProjects(items); const saved = localStorage.getItem('novel:project:' + owner); setProjectId(localStorage.getItem('novel:view:'+owner)==='studio'?'':items.some(p => p.id === saved) ? saved : items[0]?.id || ''); }
      else message.error('作品列表暂时无法读取，请刷新重试');
      if (results[1].status === 'fulfilled') setModels(modelSelectOptions(results[1].value, 'text'));
      if (results[2].status === 'fulfilled') setRetention(indexRetention(results[2].value));
      setAccount({ canManageApi: results[4].status === 'fulfilled' && canManageModelCatalog(results[4].value) });
      setLoading(false);
    });
  }, [owner]);
  useEffect(() => {
    const sequence = ++listSequence.current;
    setSessionsLoadedKey('');
    setSessions([]); setSessionId(''); setSnapshot(null); setTool(''); setImportOpen(false); setMemoOpen(false); setBlueprint(null); setBatchOpen(false); setConnection(''); setReferences([]);
    localStorage.setItem('novel:mode:' + owner, activeMode);
    if (!projectId) return;
    localStorage.setItem('novel:project:' + owner, projectId);
    listScope.call(novelProjectRequest(projectId, '/sessions', undefined, { signal: listScope.signal }), r => {
      if (sequence !== listSequence.current) return;
      const items = r.sessions || []; setSessions(items);
      const saved = localStorage.getItem(`novel:session:${owner}:${projectId}:${activeMode}`);
      const matching = items.filter(s => s.mode === activeMode && !s.archivedAt);
      const available=items.filter(s=>!s.archivedAt);
      if(!matching.length && available.length && available.every(s=>s.mode==='long_imitation')){localStorage.setItem(`novel:project-mode:${owner}:${projectId}`,'long_imitation');setActiveMode('long_imitation');return;}
      setSessionId(matching.some(s => s.id === saved) ? saved : matching[0]?.id || '');
      setSessionsLoadedKey(`${owner}:${projectId}:${activeMode}`);
    }).catch(e => { if (listScope.isCurrent()) message.error(e.message); });
  }, [owner, projectId, activeMode]);
  useEffect(() => {
    // Session snapshots already contain chapters. Only sessionless books need this read.
    if (!projectId || sessionId || sessionsLoadedKey !== `${owner}:${projectId}:${activeMode}`) return;
    let disposed = false;
    const revision = chapterRevisions.current.get(projectReadKey) || 0;
    setProjectChapters({ key: projectReadKey, revision, status: 'loading', chapters: [] });
    projectReadScope.call(getNovelCreationProjectChapters(projectId, projectReadScope.signal), result => {
      if (!disposed && revision === (chapterRevisions.current.get(projectReadKey) || 0)) { setProjectChapters({ key: projectReadKey, revision, status: 'ready', chapters: result.chapters || [] }); setConnection('已连接'); }
    }).catch(e => { if (!disposed && projectReadScope.isCurrent() && revision === (chapterRevisions.current.get(projectReadKey) || 0)) { setProjectChapters({ key: projectReadKey, revision, status: 'error', chapters: [] }); setConnection('正式章节暂时无法读回，保留草稿；点击重新连接'); message.error(e.message || '章节暂时无法读取'); } });
    return () => { disposed = true; };
  }, [projectReadKey, sessionId, sessionsLoadedKey, refresh]);
  useEffect(() => {
    const sequence = ++snapshotSequence.current;
    setSnapshot(null); setConnection(''); setBlueprint(null); setManage(null); setActionBusy(false);
    if (!sessionId) { setSnapshotRead(null); return; }
    const revision = chapterRevisions.current.get(projectReadKey) || 0;
    setSnapshotRead({ key: scopeKey, revision, status: 'loading' });
    localStorage.setItem(`novel:session:${owner}:${projectId}:${activeMode}`, sessionId);
    scope.call(getNovelCreationSessionSnapshot(sessionId, scope.signal), r => { if (sequence === snapshotSequence.current && revision === (chapterRevisions.current.get(projectReadKey) || 0)) { setSnapshot(r); setSnapshotRead({ key: scopeKey, revision, status: 'ready' }); setConnection('已连接'); } }).catch(() => { if (scope.isCurrent() && sequence === snapshotSequence.current && revision === (chapterRevisions.current.get(projectReadKey) || 0)) { setSnapshotRead({ key: scopeKey, revision, status: 'error' }); setConnection('暂时无法连接，保留草稿；点击重新连接'); } });
  }, [scopeKey, refresh]);
  useEffect(()=>{
    if(activeMode!=='long_imitation' || !hasLongImitationCreation || snapshot?.project?.id!==projectId)return;
    const incoming=snapshot.project;
    setProjects(old=>old.map(item=>{if(item.id!==projectId || item.title===incoming.title)return item;if(item.updatedAt && incoming.updatedAt && Date.parse(item.updatedAt)>Date.parse(incoming.updatedAt))return item;return {...item,...incoming};}));
    if(snapshot.session?.id)setSessions(old=>old.map(item=>item.id===snapshot.session.id?{...item,...snapshot.session}:item));
  },[activeMode,hasLongImitationCreation,projectId,snapshot?.project?.title,snapshot?.session?.title]);
  useEffect(() => {
    if (!run?.id || TERMINAL.has(run.status) || ['awaiting_confirmation', 'paused'].includes(run.status)) return;
    let timer, disposed = false, failures = 0, cursor = Math.max(0, ...(snapshot.events || []).map(e => Number(e.eventId) || 0));
    async function poll() {
      try {
        const next = await scope.call(getNovelCreationRun(run.id, cursor, scope.signal)); if (!next || disposed || !scope.isCurrent()) return;
        failures = 0; cursor = Math.max(cursor, ...(next.events || []).map(e => Number(e.eventId) || 0));
        setSnapshot(old => mergeSnapshot(old, next)); setConnection('已连接');
        if (TERMINAL.has(next.run?.status) || ['awaiting_confirmation', 'paused'].includes(next.run?.status)) return;
      } catch { if (!scope.isCurrent() || disposed) return; failures++; setConnection('连接中断，正在重新连接；后台任务继续保留'); }
      if (scope.isCurrent() && !disposed) timer = setTimeout(poll, Math.min(1500 * 2 ** failures, 15000));
    }
    timer = setTimeout(poll, 1500); return () => { disposed = true; clearTimeout(timer); };
  }, [scopeKey, run?.id, run?.status]);
  useEffect(() => {
    if (!memoOpen) return;
    const previous = document.activeElement, popup = memoPopup.current;
    popup?.querySelector('button')?.focus();
    const keyboard = e => {
      if (e.key === 'Escape') { e.preventDefault(); setMemoOpen(false); }
      if (e.key !== 'Tab') return;
      const controls = [...popup.querySelectorAll('button,input,textarea,[tabindex="0"]')].filter(el => !el.disabled && el.getClientRects().length);
      const first = controls[0], last = controls.at(-1);
      if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last?.focus(); }
      else if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first?.focus(); }
    };
    popup?.addEventListener('keydown', keyboard);
    return () => { popup?.removeEventListener('keydown', keyboard); if (previous?.isConnected) previous.focus(); };
  }, [memoOpen]);
  useEffect(() => {
    const move = e => { if (resize.current) { const next = width(window.innerWidth - e.clientX); setInspectorWidth(next); localStorage.setItem('novel:inspector-width', String(next)); } };
    const end = () => { resize.current = false; document.body.classList.remove('novel-creation-resizing'); };
    window.addEventListener('pointermove', move); window.addEventListener('pointerup', end);
    return () => { window.removeEventListener('pointermove', move); window.removeEventListener('pointerup', end); end(); };
  }, []);

  async function createUnboundSession() {
    if (flight.current || !ownerScope.isCurrent()) return;
    const token={};flight.current=token;const startedEpoch=currentView.current.studioEpoch;const cache='novel:studio:create-intent:'+owner;
    try {
      let creation=readSaved(cache,null);
      if(!creation){creation={idempotencyKey:crypto.randomUUID(),mode:'long_original'};localStorage.setItem(cache,JSON.stringify(creation));}
      const result=await ownerScope.call(novelRequest('/studio-sessions',creation,{signal:ownerScope.signal}));
      if(!result?.session?.id)return;
      localStorage.removeItem(cache);setStudioSessions(old=>[result.session,...old.filter(x=>x.id!==result.session.id)]);
      if(currentView.current.owner===owner&&currentView.current.projectId===''&&currentView.current.studioEpoch===startedEpoch){
        setStudioSessionId(result.session.id);localStorage.setItem('novel:studio:selected:'+owner,result.session.id);localStorage.setItem('novel:view:'+owner,'studio');setTool('');
      }
      return result.session;
    } catch(e){if(ownerScope.isCurrent())message.error(e.message||'创作会话暂时无法创建，原编号已保留');}
    finally{if(flight.current===token)flight.current=null;}
  }
  async function createSession(explicit = true) {
    if (!scope.isCurrent()) return;
    if (!projectId && activeMode==='long_original') return createUnboundSession();
    if (!projectId) { setProjectModal(true); return; }
    if (archived || flight.current === key + ':session') return;
    flight.current = key + ':session';
    try {
      const result = await scope.call(request('/sessions', { mode: activeMode, title: modeName + '会话', contextBudget: 120000 }));
      if (!result) return; setSessions(old => [result.session, ...old]);
      if (explicit) { setSessionId(result.session.id); setSnapshot(null); setTool(''); }
      return result.session;
    } catch (e) { error(e); } finally { if (flight.current === key + ':session') flight.current = null; }
  }
  async function send(recover = false) {
    if (!scope.isCurrent()) return;
    if (activeMode === 'short_imitation') { setShortDialog(activeMode); return; }
    if (flight.current === key || sending || archived) return;
    if (!projectId) { if (activeMode === 'continuation') message.warning('请先导入正式章节'); else setProjectModal(true); return; }
    const prior = readSaved(key + ':intent', null);
    if (prior && !recover) { message.info('上次提交尚未核对，请先恢复原提交'); return; }
    const sendDraft = { ...draft, modeInputs: { ...draft.modeInputs } };
    if (!prior) {
      if (active) return;
      if (activeMode === 'continuation') {
        if (!chaptersReady) { message.warning('正式章节尚未读回，请稍后重试或重新连接'); return; }
        const lastChapter = chapters.reduce((last, chapter) => !last || chapter.number > last.number ? chapter : last, null);
        if (!lastChapter?.content?.trim()) { message.warning('请先导入正式章节'); return; }
        sendDraft.modeInputs.previousText = lastChapter.content;
      }
      if (!draft.content.trim()) { message.warning('请输入创作要求'); return; }
      if (!availableModel(models, draft.modelId)) { message.warning('所选文本模型当前不可用，请重新选择'); return; }
      const field = MODE_FIELDS[activeMode];
      if (field && !(activeMode==='long_imitation' && hasLongImitationCreation) && !String(sendDraft.modeInputs?.[field[0]] || '').trim() && !(activeMode.includes('imitation') && (draft.styleProfileId || draft.knowledgeIds?.length))) { message.warning(activeMode.includes('imitation') ? '请导入参考文本或选择文风' : '请填写' + field[1] + '或引用已保存内容'); return; }
    }
    flight.current = key; setSendKey(key); let actualKey = key;
    try {
      let submission = prior;
      if (!submission) {
        const session = currentSession || (await scope.call(request('/sessions', { mode: activeMode, title: modeName + '会话', contextBudget: 120000 })))?.session;
        if (!session || !scope.isCurrent()) return;
        actualKey = draftKey(owner, projectId, session.id, activeMode);
        submission = { sessionId: session.id, payload: { projectId, ...sendDraft, content: draft.content.trim(), idempotencyKey: crypto.randomUUID() } };
        localStorage.setItem(actualKey, JSON.stringify(draft)); localStorage.setItem(actualKey + ':intent', JSON.stringify(submission)); localStorage.setItem(key + ':intent', JSON.stringify(submission));
        if (!currentSession) setSessions(old => [session, ...old]);
      } else actualKey = draftKey(owner, projectId, submission.sessionId, activeMode);
      const result = await scope.call(novelRequest(`/sessions/${encodeURIComponent(submission.sessionId)}/messages`, submission.payload, { signal: scope.signal }));
      if (!result || !scope.isCurrent()) return;
      localStorage.removeItem(actualKey + ':intent'); localStorage.removeItem(key + ':intent'); clearDraftAt(actualKey); clearDraftAt(key); setIntentRevision(v => v + 1);
      setSessionId(submission.sessionId); setSnapshot(old => ({ ...old, run: result.run, events: [], messages: [...(old?.messages || []), { id: result.run.id + ':user', role: 'user', content: submission.payload.content }] }));
      if (submission.sessionId === sessionId) scope.call(getNovelCreationSessionSnapshot(sessionId, scope.signal), setSnapshot).catch(() => { if (scope.isCurrent()) setConnection('提交成功，正在恢复完整会话'); });
    } catch (e) {
      if (!scope.isCurrent()) return;
      // Definite input rejection unlocks editing; timeout/5xx retain the exact payload and key.
      if (e.status >= 400 && e.status < 500 && ![408, 409, 429].includes(e.status)) { localStorage.removeItem(actualKey + ':intent'); localStorage.removeItem(key + ':intent'); }
      setIntentRevision(v => v + 1); error(e);
    } finally { if (flight.current === key) flight.current = null; if (scope.isCurrent()) setSendKey(''); }
  }
  async function runAction(action, input = {}) {
    if (!scope.isCurrent()) return;
    if (!run?.id || actionBusy || archived) return; setActionBusy(true);
    try { await scope.call(novelRequest(`/runs/${encodeURIComponent(run.id)}/${action}`, input, { signal: scope.signal }), r => { setSnapshot(old => ({ ...old, run: r.run })); setBlueprint(null); }); }
    catch (e) { error(e); } finally { if (scope.isCurrent()) setActionBusy(false); }
  }
  async function maintenance(kind, styleId) {
    if (!scope.isCurrent()) return;
    if (!sessionId) { message.info('请先新建会话'); return; }
    const pending = readSaved(key + ':maintenance', null);
    if ((active && !pending) || archived || flight.current === key + ':maintenance') { message.warning('请先完成或停止当前任务'); return; }
    const payload = pending || { kind, sessionId, ...(styleId ? { styleId } : {}), idempotencyKey: crypto.randomUUID() };
    localStorage.setItem(key + ':maintenance', JSON.stringify(payload)); setIntentRevision(v => v + 1); flight.current = key + ':maintenance';
    try { await scope.call(request('/maintenance', payload), r => { localStorage.removeItem(key + ':maintenance'); setIntentRevision(v => v + 1); setSnapshot(old => ({ ...old, run: r.run, events: [] })); setTool(''); }); }
    catch (e) { if (scope.isCurrent() && e.status >= 400 && e.status < 500 && ![408, 409, 429].includes(e.status)) { localStorage.removeItem(key + ':maintenance'); setIntentRevision(v => v + 1); } error(e); }
    finally { if (flight.current === key + ':maintenance') flight.current = null; }
  }
  async function projectAction() {
    if (!scope.isCurrent() || !manage) return;
    try {
      const target = manage;
      const path = target.type === 'session' ? `/sessions/${target.id}` : target.action === 'rename' ? '' : '/' + target.action;
      const input = target.type === 'session' ? { action: target.action, ...(target.action === 'rename' ? { title: target.title } : {}) } : target.action === 'rename' ? { title: target.title } : {};
      await scope.call(request(path, input, { method: target.type === 'session' || target.action === 'rename' ? 'PATCH' : 'POST' }), r => {
        if (target.type === 'session') setSessions(old => old.map(s => s.id === target.id ? r.session : s));
        else setProjects(old => old.map(p => p.id === projectId ? r.project || { ...p, status: target.action === 'archive' ? 'archived' : 'active' } : p));
        setManage(null); if (target.action === 'restore') message.success('已恢复，自动任务保持暂停；需要手动继续');
      });
    } catch (e) { error(e); }
  }
  async function createProject(values) {
    if (!scope.isCurrent()) return;
    try { await scope.call(novelRequest('/projects', values, { signal: scope.signal }), r => { localStorage.setItem('novel:view:'+owner,'book');setProjects(old => [r.project, ...old]); setProjectId(r.project.id); setProjectModal(false); projectForm.resetFields(); }); } catch (e) { error(e); }
  }
  async function openTool(label) {
    if (label === '导入素材') { setImportOpen(true); return; }
    if (label === '备忘录') { setMemoOpen(true); return; }
    setTool(label); if (label === '数据看板' && projectId) { setDashboard(null); try { await scope.call(request('/dashboard'), r => setDashboard(r.dashboard)); } catch (e) { error(e); } }
  }
  const input = { ...draft, modeInputs: draft.modeInputs || {}, knowledgeIds: draft.knowledgeIds || [], noteIds: draft.noteIds || [], styleProfileId: draft.styleProfileId || '' };
  const materialsProps = { projectId, selectedIds: draft.knowledgeIds || [], onSelect: knowledgeIds => changeDraft({ knowledgeIds }), onSelection: selection => { if (activeMode === 'fanfiction') changeDraft({ modeInputs: { ...draft.modeInputs, [MODE_FIELDS[activeMode][0]]: selection.content } }); }, onReferencesChanged: items => setReferences(old => [...new Map([...old, ...items].map(i => [i.id, i])).values()]), styleId: draft.styleProfileId, onStyle: styleProfileId => changeDraft({ styleProfileId }), onMaintenance: maintenance };
  const archivedProjects = projects.filter(p => ['archived', 'purging'].includes(p.status) || retention[`novel_creation:${p.id}`]?.archivedAt);
  const visibleProjects = projects.filter(p => !archivedProjects.includes(p));

  if (loading) return <div className="novel-creation-loading"><Spin /><span>正在打开小说创作</span></div>;
  return <div className="novel-creation-page" style={{ '--novel-inspector-width': inspectorWidth + 'px' }}>
    <main className="novel-creation-workbench">
      <aside className="novel-creation-sidebar">
        <div className="novel-sidebar-content">
        <h3>开始创作</h3><div className="novel-creation-mode-tabs">{MODES.map(([id, name]) => { const Icon=creationIcons[id]; return <button type="button" className={(longImitationOpen ? id==='long_imitation' : activeMode === id) ? 'active' : ''} key={id} onClick={() => enterCreation(id)}><Icon size={15} aria-hidden="true"/><span>{name}</span></button>; })}</div>
        <div className="novel-sidebar-heading"><b>我的作品 {visibleProjects.length}</b><Button aria-label="新建作品" type="text" icon={<Plus size={16} />} onClick={() => setProjectModal(true)} /><Button aria-label="作品归档库" type="text" icon={<Archive size={16} />} onClick={() => setArchiveOpen(true)} /></div>
        <div className="novel-projects">{visibleProjects.map(p => <div key={p.id} className="novel-project"><button type="button" aria-label={`选择作品${p.title}`} className={projectId === p.id ? 'active' : ''} onClick={() => { localStorage.setItem('novel:view:'+owner,'book');setProjectId(p.id);const savedMode=localStorage.getItem(`novel:project-mode:${owner}:${p.id}`);if(savedMode)setActiveMode(savedMode);setTool(''); }}>{p.title}</button>{projectId === p.id && <div className="novel-actions"><Button size="small" onClick={() => setManage({ type: 'project', action: 'rename', title: p.title })}>重命名作品</Button><Button size="small" onClick={() => setManage({ type: 'project', action: 'archive' })}>归档作品</Button></div>}</div>)}</div>
        {studioSessions.length>0&&<><div className="novel-session-heading">创作会话记录</div><div className="novel-sessions">{studioSessions.map(item=><div className="novel-session" key={item.id}><button type="button" className={studioView&&studioSessionId===item.id?'active':''} onClick={()=>{setActiveMode('long_original');setProjectId('');setStudioSessionId(item.id);setTool('');localStorage.setItem('novel:studio:selected:'+owner,item.id);localStorage.setItem('novel:view:'+owner,'studio');}}>{item.title}{item.projectId?' · 已建作品':''}</button></div>)}</div></>}
        <div className="novel-session-heading">{studioView?'未关联作品的长篇原创会话':modeName+'会话'}</div><div className="novel-sessions">{sessions.filter(s => s.mode === activeMode).map(s => <div className="novel-session" key={s.id}><button type="button" disabled={!!s.archivedAt} className={sessionId === s.id ? 'active' : ''} onClick={() => { setSessionId(s.id); setTool(''); }}>{s.title}{s.archivedAt ? '（已归档）' : s.currentRunId ? ' · 有任务' : ''}</button><div className="novel-actions"><Button size="small" onClick={() => setManage({ type: 'session', id: s.id, action: 'rename', title: s.title })}>重命名会话</Button><Button size="small" onClick={() => setManage({ type: 'session', id: s.id, action: s.archivedAt ? 'restore' : 'archive' })}>{s.archivedAt ? '恢复会话' : '归档会话'}</Button></div></div>)}</div><Button onClick={() => createSession()} icon={<Plus size={14} />} disabled={archived}>新建会话</Button>
        </div><nav className="novel-creation-toolbar" aria-label="创作工具">{TOOLS.map(([label, Icon]) => <Button key={label} type="text" className={tool === label ? 'active' : ''} icon={<Icon size={16} />} onClick={() => openTool(label)}>{label}</Button>)}</nav>
      </aside>
      {tool ? <section className="novel-creation-tool-workspace" aria-label={tool}>
        <header className="novel-chat-heading"><Button onClick={() => setTool('')}>返回对话</Button><b>{tool}</b></header>
        <div className="novel-tool-content" key={scopeKey + ':' + tool}>{!projectId ? <Empty description="请先选择作品后使用此工具" /> : tool === 'Agent设置' ? <AgentPanel projectId={projectId} models={models} /> : tool === '写作风格' ? <MaterialsPanel {...materialsProps} styleMode /> : tool === '定时写作' ? <AutomationPanel projectId={projectId} sessionId={sessionId} mode={activeMode} input={input} models={models} ensureSession={() => createSession(false)} /> : <div className="novel-panel"><p>统计已保存正文，失败次数不计为章节。</p>{dashboard ? <dl>{[['章节数', dashboard.chapterCount], ['总字数', dashboard.totalWords], ['任务总数', dashboard.runTotal], ['成功率', `${Number(dashboard.successRate ?? 0).toFixed(1)}%`]].map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl> : <Empty description="暂无统计" />}</div>}</div>
      </section> : studioView ? <StudioChat key={`${owner}:${studioSessionId}`} owner={owner} sessionId={studioSessionId} models={models} onSessionSelected={id=>setStudioSessionId(id)} onSessionCreated={item=>setStudioSessions(old=>[item,...old.filter(x=>x.id!==item.id)])} onTransition={result=>{setProjects(old=>[result.project,...old.filter(x=>x.id!==result.project.id)]);localStorage.setItem('novel:view:'+owner,'book');localStorage.setItem('novel:project:'+owner,result.project.id);localStorage.setItem(`novel:session:${owner}:${result.project.id}:long_original`,result.productionSession.id);setProjectId(result.project.id);setActiveMode('long_original');setSessionId(result.productionSession.id);setTool('');setRefresh(v=>v+1);}}/> : <><section className="novel-creation-conversation">
        <header className="novel-chat-heading"><b>{currentSession?.title || modeName}</b>{connection && <button type="button" className="novel-connection" onClick={() => setRefresh(v => v + 1)}>{connection}</button>}</header>
        <div className="novel-creation-chat-scroll">{!snapshot?.messages?.length && !run && <div className="novel-creation-welcome"><p>告诉我你想写什么——题材、世界观、主角、核心冲突。</p></div>}
          {(snapshot?.messages || []).map((m, i) => <article key={m.id || i} className={`novel-creation-message ${m.role}`}><b>{m.role === 'user' ? '我' : '创作助手'}</b><p>{m.content}</p></article>)}
          {activeMode==='long_imitation' && projectId && <LongImitationContext key={projectReadKey} owner={owner} projectId={projectId} refreshKey={`${refresh}:${snapshot?.run?.status||''}`} onLoaded={(id,creation)=>setLongImitationCreation({key:`${owner}:${id}`,creation})}/>}
          <RunCard snapshot={snapshot} onAction={runAction} onBlueprint={setBlueprint} busy={actionBusy} />
        </div>
        <div className="novel-creation-composer novel-creation-composer--compact">
          {archived && <p role="status">此作品或会话已归档，恢复后才可以创作。</p>}

          <Input.TextArea aria-label="创作要求" value={draft.content} disabled={!!intent || sending} rows={3} placeholder="输入创作要求（Enter发送，Shift+Enter换行）" onChange={e => changeDraft({ content: e.target.value })} onCompositionStart={() => { composition.current = true; }} onCompositionEnd={() => { composition.current = false; }} onPressEnter={e => { if (!e.shiftKey && !composition.current && !e.nativeEvent?.isComposing && e.nativeEvent?.keyCode !== 229) { e.preventDefault(); send(); } }} />
          {(!!draft.knowledgeIds?.length || !!draft.noteIds?.length || !!draft.styleProfileId) && <div className="novel-reference-summary">参考资料 {draft.knowledgeIds?.length || 0} 份 · 备忘录 {draft.noteIds?.length || 0} 份 · {draft.styleProfileId ? '已选择文风' : '未选择文风'}</div>}
          <ComposerToolbar referenceControl={activeMode !== 'continuation' && !(activeMode==='long_imitation' && hasLongImitationCreation) && MODE_FIELDS[activeMode] && <details key={activeMode} className="novel-reference-editor"><summary><BookOpen size={14}/>{String(draft.modeInputs?.[MODE_FIELDS[activeMode]?.[0]] || '').trim() ? '参考文本 · 已添加' : '添加参考文本'}</summary><div className="novel-reference-editor-panel"><label>{MODE_FIELDS[activeMode][1]}<Input.TextArea aria-label={MODE_FIELDS[activeMode][1]} rows={2} value={draft.modeInputs?.[MODE_FIELDS[activeMode][0]] || ''} onChange={e => changeDraft({ modeInputs: { ...draft.modeInputs, [MODE_FIELDS[activeMode][0]]: e.target.value } })} placeholder={activeMode.includes('imitation') ? '参考文本与已保存文风可以二选一' : '也可在导入素材中选择正文或设定'} />{activeMode === 'fanfiction' && <Button size="small" disabled={!project?.world && !project?.protagonist} onClick={() => changeDraft({ modeInputs: { ...draft.modeInputs, existingSetting: [project.world, project.protagonist, project.coreConflict].filter(Boolean).join('\n') } })}>使用本作品人物与世界设定</Button>}</label></div></details>} models={modelOptions(models, draft.modelId)} modelId={draft.modelId} onModelChange={modelId=>changeDraft({modelId})} thinkingLevel={draft.thinkingLevel} onThinkingChange={thinkingLevel=>changeDraft({thinkingLevel})} locked={sending||!!intent} sending={sending} disabled={archived||active||!!intent} hasContent={!!draft.content.trim()} sendLabel="开始创作" onSend={()=>send()}/>
          {draft.modelId && !availableModel(models, draft.modelId) && <p role="alert">原来选择的模型已不可用，请重新选择授权模型。</p>}
          {!models.length && <p>{account?.canManageApi === true ? '暂无可用文本模型，请在模型管理中检查配置。' : '暂无授权文本模型，请联系管理员。'}</p>}
          {intent && <div className="novel-notice" role="status">上次提交结果尚未核对。草稿已保留，恢复时只重复原来的提交编号。<Button loading={sending} onClick={() => send(true)}>恢复原提交</Button></div>}
          {maintenanceIntent && <div className="novel-notice" role="status">资料分析提交结果尚未核对，将使用原来的提交编号恢复。<Button onClick={() => maintenance(maintenanceIntent.kind, maintenanceIntent.styleId)}>恢复分析提交</Button></div>}
        </div>
      </section>
      {projectId && (shortStory || chapters.length > 0 || run || draft.knowledgeIds?.length) && <><div className="novel-creation-resizer" role="separator" aria-label="调整资料栏宽度" onPointerDown={() => { resize.current = true; document.body.classList.add('novel-creation-resizing'); }} />
      <aside className="novel-creation-inspector">{shortStory ? <ShortStoryPanel key={projectReadKey} owner={owner} projectId={projectId} refreshKey={`${refresh}:${snapshot?.run?.status || ""}`} onRun={nextRun=>{setSnapshot(old=>({...old,run:nextRun,events:[]}));setRefresh(v=>v+1);}} onTitle={title=>setProjects(old=>old.map(p=>p.id===projectId && title ? {...p,title} : p))}/> : <><div className="novel-current-reference"><h3>当前资料</h3><p>{project?.title} · {project?.genre}</p>{(draft.knowledgeIds || []).map(id => <small key={id}>{references.find(r => r.id === id)?.title || '已保存参考范围'}</small>)}</div>
        <Tabs defaultActiveKey="chapters" items={[
          { key: 'chapters', label: `正文${chapters.length ? '（' + chapters.length + '）' : ''}`, children: projectId ? <ChaptersPanel key={scopeKey} projectId={projectId} sessionId={sessionId} chapters={chapters} run={run} canExportShort={snapshot?.canExportShort === true} onChapter={chapter => sessionId ? setSnapshot(old => ({ ...old, chapters: (old?.chapters || []).map(c => c.id === chapter.id ? chapter : c) })) : setProjectChapters(old => old?.key === projectReadKey ? { ...old, chapters: old.chapters.map(c => c.id === chapter.id ? chapter : c) } : old)} onMaintenance={maintenance} onBatch={() => setBatchOpen(true)} /> : <Empty description="选择作品后查看正文" /> },
          ...(run ? [{ key: 'progress', label: '创作进度', children: <ol className="novel-timeline">{(snapshot.events || []).filter(e => e.type !== 'assistant.delta').map(e => <li key={e.eventId}><b>{STATUS[e.status] || '处理进度'}</b><p>{e.message}</p></li>)}</ol> }] : [])
        ]} />
      </>}</aside></>}</>}
    </main>
    <LongImitationDialog key={'long-imitation:'+owner} owner={owner} open={longImitationOpen} viewKey={`${scopeKey}:${tool}:${longImitationEntry}`} models={models} modelId={draft.modelId} thinkingLevel={draft.thinkingLevel} onCancel={()=>setLongImitationOpen(false)} onPersisted={result=>setProjects(old=>{const previous=old.find(p=>p.id===result.project.id);const newer=previous?.updatedAt && result.project.updatedAt && Date.parse(previous.updatedAt)>Date.parse(result.project.updatedAt);return [newer?previous:result.project,...old.filter(p=>p.id!==result.project.id)];})} onCreated={result=>{
      localStorage.setItem('novel:view:'+owner,'book');localStorage.setItem('novel:project:'+owner,result.project.id);localStorage.setItem(`novel:session:${owner}:${result.project.id}:long_imitation`,result.session.id);localStorage.setItem(`novel:project-mode:${owner}:${result.project.id}`,'long_imitation');setProjectId(result.project.id);setActiveMode('long_imitation');setSessionId(result.session.id);setLongImitationOpen(false);setTool('');setRefresh(n=>n+1);
    }}/>
    <ShortStoryDialog key={`${owner}:${shortDialog || 'short_imitation'}`} owner={owner} mode={shortDialog || 'short_imitation'} open={!!shortDialog} viewKey={scopeKey + ':' + tool} models={models} modelId={draft.modelId} thinkingLevel={draft.thinkingLevel} projects={visibleProjects} onCancel={()=>setShortDialog(null)} onPersisted={result=>setProjects(old=>[result.project,...old.filter(p=>p.id!==result.project.id)])} onCreated={result=>{
      localStorage.setItem('novel:view:'+owner,'book');localStorage.setItem('novel:project:'+owner,result.project.id);localStorage.setItem(`novel:session:${owner}:${result.project.id}:${result.story.input.mode}`,result.session.id);localStorage.setItem(`novel:project-mode:${owner}:${result.project.id}`,result.story.input.mode);
      setProjectId(result.project.id);setSessionId(result.session.id);setActiveMode(result.story.input.mode);setShortStoryData({key:`${owner}:${result.project.id}`,story:result.story});setShortDialog(null);setSnapshot({run:result.run,messages:[],events:[]});setRefresh(n=>n+1);
    }}/>
    <ContinuationImportDialog key={owner} owner={owner} open={continuationOpen} viewKey={scopeKey + ':' + tool} projects={projects} onCancel={() => setContinuationOpen(false)} onPersisted={recordImportedBook} onImported={result => {
      if (!ownerScope.isCurrent()) return;
      setProjectId(result.book.id); setTool(''); setContinuationOpen(false);
    }} />
    <Modal rootClassName="novel-modal" open={projectModal} title="新建小说作品" okText="创建作品" cancelText="取消" onCancel={() => setProjectModal(false)} onOk={() => projectForm.submit()} destroyOnHidden>
      <Form form={projectForm} layout="vertical" onFinish={createProject} initialValues={{ genre: '都市', modelId: models[0]?.value, targetChapterWords: 2000 }}><Form.Item name="title" label="作品名称" rules={[{ required: true, whitespace: true, message: '请输入作品名称' }]}><Input /></Form.Item><Form.Item name="genre" label="题材" rules={[{ required: true, whitespace: true, message: '请输入题材' }]}><Input /></Form.Item><Form.Item name="modelId" label="文本模型" rules={[{ required: true, message: '请选择文本模型' }]}><Select options={models} /></Form.Item><Form.Item name="targetChapterWords" label="每章目标字数"><InputNumber min={800} max={10000} /></Form.Item>{[['world', '世界设定'], ['protagonist', '主角设定'], ['coreConflict', '核心冲突']].map(([name, label]) => <Form.Item key={name} name={name} label={label}><Input.TextArea rows={2} /></Form.Item>)}</Form>
    </Modal>
    <Modal rootClassName="novel-modal" open={!!manage} title={manage?.action === 'rename' ? '重命名' : manage?.action === 'restore' ? '恢复' : '归档'} okText={manage?.action === 'rename' ? '保存名称' : manage?.action === 'restore' ? '确认恢复' : '确认归档'} cancelText="取消" onCancel={() => setManage(null)} onOk={projectAction}>{manage?.action === 'rename' ? <Input aria-label="新名称" value={manage.title} onChange={e => setManage(old => ({ ...old, title: e.target.value }))} /> : <p>{manage?.action === 'restore' ? '恢复会重新计算保留时间；自动写作保持暂停。' : '进入可恢复的归档；正在创作时请先停止任务。'}</p>}</Modal>
    <BlueprintEditor value={blueprint} onChange={setBlueprint} onCancel={() => setBlueprint(null)} onConfirm={value => runAction('confirm', { blueprint: value })} />
    <Modal rootClassName="novel-modal" open={batchOpen} title="批量续写" width={720} footer={null} onCancel={() => setBatchOpen(false)} destroyOnHidden>{batchOpen && projectId && <AutomationPanel key={scopeKey} projectId={projectId} sessionId={sessionId} mode={activeMode} input={input} models={models} kind="batches" ensureSession={() => createSession(false)} />}</Modal>
    <Modal rootClassName="novel-modal novel-import-modal" open={importOpen} title="导入素材" width={720} footer={null} onCancel={() => setImportOpen(false)} destroyOnHidden>{importOpen && (projectId ? <MaterialsPanel key={scopeKey} {...materialsProps} /> : <Empty description="请先选择作品后导入素材" />)}</Modal>
    {memoOpen && createPortal(<div className="novel-memo-overlay" onClick={e => { if (e.target === e.currentTarget) setMemoOpen(false); }}><section ref={memoPopup} className="novel-memo-popup" role="dialog" aria-modal="true" aria-labelledby="novel-memo-title"><header className="novel-chat-heading"><b id="novel-memo-title">备忘录</b><Button aria-label="关闭备忘录" onClick={() => setMemoOpen(false)}>关闭</Button></header>{projectId ? <NotesPanel key={scopeKey} projectId={projectId} selectedIds={draft.noteIds || []} onSelect={noteIds => changeDraft({ noteIds })} /> : <Empty description="请先选择作品后使用备忘录" />}</section></div>, document.body)}
    <Drawer rootClassName="novel-drawer" open={archiveOpen} title="作品归档库" width={520} onClose={() => setArchiveOpen(false)}>{archivedProjects.map(p => { const purging = p.status === 'purging' || retention[`novel_creation:${p.id}`]?.purging; return <div className="novel-item" key={p.id}><b>{p.title}</b><p>{purging ? '作品正在清理，暂不可恢复' : '归档后可恢复；恢复后重新计算保留时间'}</p><Button disabled={purging} onClick={async () => { try { await ownerScope.call(novelProjectRequest(p.id, '/restore', {}, { signal: ownerScope.signal }), () => { setProjects(old => old.map(item => item.id === p.id ? { ...item, status: 'active' } : item)); setRetention(old => ({ ...old, [`novel_creation:${p.id}`]: {} })); message.success('作品已恢复，自动写作保持暂停'); }); } catch (e) { if (ownerScope.isCurrent()) message.error(e.message); } }}>恢复作品</Button></div>; })}</Drawer>
  </div>;
}
