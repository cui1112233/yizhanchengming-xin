import OriginalIcon, { ComicScreenplayModelIcon } from './ComicScreenplayIcon.jsx';
import { formatWorkDateTime } from '../../shared/workDateTime.js';
import { Alert, Button, Checkbox, Drawer, Dropdown, Empty, Input, InputNumber, message, Modal, Popconfirm, Select, Spin, Tabs, Upload } from 'antd';
import { ArrowDown, ArrowLeft, ArrowUp, ChevronDown, Clapperboard, Clock3, Download, FileText, Grid2X2, History, LayoutGrid, PauseCircle, Pencil, Play, Plus, Search, Sparkles, Trash2, UploadCloud, Users, Map, Package, WandSparkles, Settings2, Palette, Bot, Check, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { cancelComicScreenplayBatch, createComicScreenplay, createComicScreenplayAsset, createComicScreenplayShot, createComicStoryboardShot, deleteComicScreenplay, deleteComicScreenplayAsset, deleteComicScreenplayShot, deleteComicStoryboardShot, exportComicScreenplay, extractComicAssets, generateComicScreenplay, generateComicStoryboard, getComicScreenplayBatch, loadComicCreativeLibrary, loadComicCreativeConfig, saveComicCreativeConfig, listComicScreenplays, loadComicScreenplayWorkspace, mergeComicStoryboardShot, moveComicScreenplayShot, parseComicScreenplaySource, restoreComicScreenplayDocument, runComicScreenplayBatchNext, saveComicScreenplayDocument, saveComicScreenplaySource, splitComicStoryboardShot, startComicScreenplayBatch, updateComicScreenplay, updateComicScreenplayAsset, updateComicScreenplayShot, updateComicStoryboardShot } from '../../shared/api/comicScreenplay.js';
import { listAvailableModels } from '../../shared/api/modelCatalog.js';
import './comic-screenplay.css';
import { getComicDisplay, getComicMedia, listArchivedComics, listSharedComics, archiveComic, restoreComic, batchEditComicShots } from '../../shared/api/comicScreenplayWorkflow.js';
import { ComicChaptersPanel, ComicAssistantPanel, ComicTeamDialog, ComicDisplayDialog } from './ComicScreenplayAuthoringPanels.jsx';
import ComicScreenplayProductionPanel from './ComicScreenplayProductionPanel.jsx';
import ComicScreenplayMediaPanel from './ComicScreenplayMediaPanel.jsx';
import ComicScreenplayStoryStage from './ComicScreenplayStoryStage.jsx';
import ComicScreenplayAssetStage from './ComicScreenplayAssetStage.jsx';
import ComicScreenplayStoryboardStage from './ComicScreenplayStoryboardStage.jsx';
import { comicStages, comicStageLocation, comicNextState } from './comic-screenplay-stages.js';
import ComicScreenplayFileTools from './ComicScreenplayFileTools.jsx';
import { comicDraftPayload, createSourceImportGate } from './comic-screenplay-creation.js';
import { ComicDegreeCards, ComicSourceInput, ComicStylePicker, ComicCreativeFields } from './ComicScreenplayCreationFields.jsx';
import { applyCreativeConfigToProject, comicProjectStyleName, comicStoryboardEngine, createCreativeCreationFlow, creativeConfigDraft, creativeConfigPayload, emptyCreativeConfig } from './comic-screenplay-creative.js';
import { parseComicScreenplayDraftSource } from '../../shared/api/comicScreenplay.js';

const ratioOptions = ['9:16', '16:9', '1:1', '4:3'].map(value => ({ value, label: value }));
const assetTypeOptions = [{ value: 'character', label: '角色' }, { value: 'scene', label: '场景' }, { value: 'prop', label: '道具' }];

function prettyTime(value) {
  return formatWorkDateTime(value);
}

const episodeHeadingPattern = /^#[ \t]*第[ \t]*([0-9]+)[ \t]*集[^\n]*$/gm;

// 按「# 第N集」标题切分剧本正文，与后端 splitEpisodes 的集号识别保持一致。
function splitDocumentEpisodes(content) {
  const text = String(content || '');
  const matches = [...text.matchAll(episodeHeadingPattern)];
  if (!matches.length) return [];
  return matches
    .map((match, index) => {
      const start = match.index;
      const end = index + 1 < matches.length ? matches[index + 1].index : text.length;
      return { no: Number(match[1]), content: text.slice(start, end).trim() };
    })
    .filter(item => item.no > 0 && item.content);
}

function collectWarnings(result) {
  const warnings = Array.isArray(result?.warnings) ? result.warnings.filter(Boolean) : [];
  const failedChunks = Number(result?.failed_chunks || 0);
  if (failedChunks > 0 && !warnings.length) return [`有 ${failedChunks} 个批次失败，请检查后重试`];
  return warnings;
}

// 部分失败必须单独提示：不能再把「第 N 集 / 第 N 批失败」混成一句「处理完成」。
function describePartialFailure(result) {
  if (!result?.partial_failure) return '';
  const episodes = Array.isArray(result.failed_episodes) ? result.failed_episodes.filter(no => Number(no) > 0) : [];
  if (episodes.length) return `第 ${episodes.join('、')} 集生成失败`;
  const batches = Array.isArray(result.failed_batches)
    ? result.failed_batches.map(item => Number(item?.batch_index)).filter(no => no > 0)
    : [];
  if (batches.length) return `第 ${batches.join('、')} 批提取失败`;
  const failedChunks = Number(result?.failed_chunks || 0);
  if (failedChunks > 0) return `${failedChunks} 个批次失败`;
  return '部分内容失败';
}

const batchStatusLabels = {
  running: '执行中', waiting_user: '等待人工确认', awaiting_user: '等待人工确认',
  paused: '已暂停', completed: '已完成', failed: '失败', pending: '待执行',
  in_progress: '执行中', done: '已完成', stale: '需要重做', skipped: '已跳过', blocked: '审核未通过',
};

function batchStatusLabel(status) {
  return batchStatusLabels[status] || status || '待执行';
}

function artifactView(artifact) {
  const value = String(artifact || '');
  if (/^https?:\/\//i.test(value) || value.startsWith('/api/')) {
    return { href: value, label: value.split('/').filter(Boolean).at(-1) || '打开产物' };
  }
  const [kind, id] = value.split('://');
  const labels = { source: '原文', document: '剧本', asset: '资产', storyboard: '分镜', project: '项目' };
  return { label: `${labels[kind] || kind || '产物'}：${id || value}` };
}

export default function ComicScreenplayPage() {
  const [projects, setProjects] = useState([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');
  const [createOpen, setCreateOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [selectedId, setSelectedId] = useState('');
  const [workspace, setWorkspace] = useState(null);
  const [working, setWorking] = useState('');
  const [access, setAccess] = useState(null);
  const [display, setDisplay] = useState({editor_version:'studio',font_size:14,show_descriptions:true});
  const [activeTab,setActiveTab] = useState('source');
  const [settingsOpen,setSettingsOpen] = useState(false), [agentOpen,setAgentOpen] = useState(false), [tool,setTool] = useState(''), [openedTools,setOpenedTools] = useState({}), [mediaContext,setMediaContext] = useState({}), [editingEpisode,setEditingEpisode] = useState(0), [modelSearchOpen,setModelSearchOpen] = useState(false), [modelQuery,setModelQuery] = useState('');
  const settingsSnapshot = useRef(null), settingsRequest = useRef(0);
  function navigateStage(key) { const location = comicStageLocation(key); setActiveTab(location.stage); if (location.tool) { setTool(location.tool); setOpenedTools(current=>({...current,[location.tool]:true})); } if (location.agent) setAgentOpen(true); }
  function openSettings() { settingsSnapshot.current = { form: { ...form }, creative: { ...creative } }; setSettingsOpen(true); void listAvailableModels('text').then(setModels).catch(error=>message.error(error.message)); }
  function closeSettings() { if(settingsSaving)return; if(settingsSnapshot.current){setForm(settingsSnapshot.current.form);setCreative(settingsSnapshot.current.creative);} setSettingsOpen(false); }
  function openTool(key, context = {}) { if(key==='settings'){openSettings();return;} if(key==='history'){setVersionHistoryOpen(true);return;} if(key==='media')setMediaContext(context); setOpenedTools(current=>({...current,[key]:true})); setTool(key); }

  const [teamOpen,setTeamOpen] = useState(false), [displayOpen,setDisplayOpen] = useState(false);
  const [storyboardEngine,setStoryboardEngine] = useState('xiaoluo_v7_15s');
  const [selectedShots,setSelectedShots] = useState({}), [bulkDuration,setBulkDuration] = useState(3);
  const readOnly = !access?.edit, canGenerate = !!access?.generate;
  const [source, setSource] = useState('');
  const [createContent, setCreateContent] = useState('');
  const [sourceMode, setSourceMode] = useState('text');
  const [sourceFilename, setSourceFilename] = useState('');
  const [sourceParsing, setSourceParsing] = useState(false);
  const sourceImportGate = useRef(null);
  if (!sourceImportGate.current) sourceImportGate.current = createSourceImportGate(parseComicScreenplayDraftSource);
  const [models, setModels] = useState([]);
  const [creative, setCreative] = useState(emptyCreativeConfig);
  const [creativeLibrary, setCreativeLibrary] = useState({ styles: [], prompts: [] });
  const [creativeLoading, setCreativeLoading] = useState(false);
  const [creativeError, setCreativeError] = useState('');
  const [creativeReady, setCreativeReady] = useState(false);
  const [creativeRevision, setCreativeRevision] = useState(0);
  const [createdDraftId, setCreatedDraftId] = useState('');
  const createTicket = useRef(0), creativeTicket = useRef(0), createPending = useRef(false), settingsPending = useRef(false), createdDraft = useRef(null), creationFlow = useRef(null);
	const [shotEditor, setShotEditor] = useState(null);
	const [assetEditor, setAssetEditor] = useState(null);
  const [storyboardShotEditor, setStoryboardShotEditor] = useState(null);
  const [scope, setScope] = useState('personal');
  const [sortOrder, setSortOrder] = useState('desc');
  const [viewMode, setViewMode] = useState('grid');
  const [form, setForm] = useState({ title: '', model: '', promptId: 'b_director_screenplay', artStyle: '', selectedStyle: '', aspectRatio: '9:16', degree: 'master', scriptMode: 'b' });
  const [taskWarnings, setTaskWarnings] = useState([]);
  const [storyboardEpisode, setStoryboardEpisode] = useState(1);
  const [sourceSaving, setSourceSaving] = useState(false);
  const [settingsSaving, setSettingsSaving] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [versionHistoryOpen, setVersionHistoryOpen] = useState(false);
  const [documentDraft, setDocumentDraft] = useState('');
  const abortRef = useRef(null),sourceSavePromise=useRef(null);
  const libraryRequest=useRef(0),projectRequest=useRef(0),selectedProject=useRef(''),libraryScope=useRef(scope);libraryScope.current=scope;

  async function refreshProjects() {
    if(scope!==libraryScope.current)return;
    const ticket=++libraryRequest.current;setLoading(true);
    try { const rows=scope==='team'?await listSharedComics():scope==='archived'?await listArchivedComics():await listComicScreenplays(); if(ticket===libraryRequest.current)setProjects(Array.isArray(rows)?rows:rows.projects||[]); }
    catch (error) { if(ticket===libraryRequest.current)message.error(error.message || '漫剧项目读取失败'); }
    finally { if(ticket===libraryRequest.current)setLoading(false); }
  }

  useEffect(() => {setProjects([]);void refreshProjects(); return () => {libraryRequest.current+=1; abortRef.current?.abort(); sourceImportGate.current.cancel(); }; }, [scope]);
  useEffect(() => { setDocumentDraft(workspace?.document?.content || ''); }, [workspace?.document?.id]);
  useEffect(()=>{if(!selectedId)return;window.dispatchEvent(new CustomEvent('qiantie:comic-workspace',{detail:{open:true}}));return()=>window.dispatchEvent(new CustomEvent('qiantie:comic-workspace',{detail:{open:false}}));},[selectedId]);

  async function readCreativeLibrary() {
    const ticket = ++creativeTicket.current;
    setCreativeLoading(true); setCreativeError('');
    try {
      const library = await loadComicCreativeLibrary();
      if (ticket !== creativeTicket.current) return;
      setCreativeLibrary(library);
    } catch (error) { if (ticket === creativeTicket.current) setCreativeError(error.message || '创作目录读取失败'); }
    finally { if (ticket === creativeTicket.current) setCreativeLoading(false); }
  }

  async function openCreateDialog() {
    const ticket = ++createTicket.current;
    sourceImportGate.current.cancel();
    setCreateContent(''); setSourceMode('text'); setSourceFilename(''); setSourceParsing(false);
    setForm(current => ({ ...current, title: '', artStyle: '', selectedStyle: '', degree: 'master' }));
    setCreative(emptyCreativeConfig()); setCreativeReady(true);
    createdDraft.current = null; setCreatedDraftId('');
    creationFlow.current = createCreativeCreationFlow({ create: createComicScreenplay, load: loadComicCreativeConfig, save: saveComicCreativeConfig });
    setCreateOpen(true);
    void readCreativeLibrary();
    try {
      const modelRows = await listAvailableModels('text');
      if (ticket !== createTicket.current) return;
      setModels(modelRows);
      setForm(current => ({ ...current, model: current.model || modelRows[0]?.id || '' }));
    } catch (error) { if (ticket === createTicket.current) message.error(error.message || '模型读取失败'); }
  }

  async function readProjectCreative(id) {
    const ticket = ++creativeTicket.current;
    setCreativeLoading(true); setCreativeReady(false); setCreativeError('');
    try {
      const [library, response] = await Promise.all([loadComicCreativeLibrary(), loadComicCreativeConfig(id)]);
      if (ticket !== creativeTicket.current || selectedProject.current !== id) return;
      setCreativeLibrary(library); setCreative(creativeConfigDraft(response.creative_config));
      setCreativeRevision(response.revision); setCreativeReady(true);
      setStoryboardEngine(current => comicStoryboardEngine(response.creative_config?.storyboard_prompt_id, current));
      setWorkspace(current => current ? { ...current, project: applyCreativeConfigToProject(current.project, response.creative_config, response.revision) } : current);
      setForm(current => ({ ...current, artStyle: response.creative_config?.style?.name || '', selectedStyle: response.creative_config?.style?.id || '' }));
    } catch (error) { if (ticket === creativeTicket.current) setCreativeError(error.message || '创作配置读取失败，暂不能保存'); }
    finally { if (ticket === creativeTicket.current) setCreativeLoading(false); }
  }

  async function openProject(id) {
    const ticket=++projectRequest.current;
    const switching=id!==selectedProject.current;
    selectedProject.current=id;setSelectedId(id);setLoading(true);
    if(switching){settingsRequest.current+=1;settingsPending.current=false;setSettingsSaving(false);setSettingsOpen(false);setAgentOpen(false);setTool('');setOpenedTools({});setCreative(emptyCreativeConfig());setCreativeReady(false);setWorkspace(null);setAccess(null);setSelectedShots({});setShotEditor(null);setAssetEditor(null);setStoryboardShotEditor(null);setTeamOpen(false);setDisplayOpen(false);}
    try {
      const [data,authoring]=await Promise.all([loadComicScreenplayWorkspace(id),getComicDisplay(id)]);
      if(ticket!==projectRequest.current||selectedProject.current!==id)return;
      void readProjectCreative(id);
      setWorkspace(data);setAccess(authoring.access||data.access);setSource(data?.project?.source_text||'');
      if(switching){setTaskWarnings([]);const prefs=authoring.state?.display||{};setDisplay(prefs);navigateStage(prefs.active_tab||'source');}
      const episodes=splitDocumentEpisodes(data?.document?.content);
      setStoryboardEpisode(current=>episodes.some(item=>item.no===current)?current:episodes[0]?.no||1);
      setForm(current=>({...current,model:data?.project?.model_id||'',promptId:data?.project?.prompt_id||'b_director_screenplay',synopsis:data?.project?.synopsis||'',degree:data?.project?.asset_degree||'master',artStyle:data?.project?.art_style||'',selectedStyle:data?.project?.selected_style||'',aspectRatio:data?.project?.aspect_ratio||'9:16',scriptMode:data?.project?.script_mode||'b'}));
    }catch(error){if(ticket===projectRequest.current){message.error(error.message||'工作台读取失败');selectedProject.current='';setSelectedId('');setWorkspace(null);setAccess(null);}}
    finally{if(ticket===projectRequest.current)setLoading(false);}
  }

  async function refreshCurrentProject(id){if(selectedProject.current===id)return openProject(id);}

  async function persistSource() {
    if(readOnly)return workspace?.project;
    if([...source].length>600000)throw new Error('原文超过 60 万字，请按章拆分后分批提交');
    if(source===workspace?.project?.source_text)return workspace.project;
    const key=JSON.stringify([selectedId,source]);
    if(sourceSavePromise.current?.key===key)return sourceSavePromise.current.promise;
    setSourceSaving(true);
    const id=selectedId;
    const pending={key,promise:saveComicScreenplaySource(id,source).then(project=>{if(project&&selectedProject.current===id)setWorkspace(current=>current?{...current,project}:current);return project;}).finally(()=>{if(sourceSavePromise.current===pending){sourceSavePromise.current=null;setSourceSaving(false);}})};
    sourceSavePromise.current=pending;return pending.promise;
  }

  async function saveSource() {
    if(readOnly)return;
    if (!selectedId || sourceSaving) return;
    try { await persistSource(); message.success('原文已保存'); }
    catch (error) { message.error(error.message || '原文保存失败'); }
  }

  async function saveDocumentDraft(content = documentDraft) {
    if(readOnly){message.info('当前作品只有查看权限');return;}
    if (!selectedId || working) return;
    try {
      setWorking('document-save');
      await saveComicScreenplayDocument(selectedId, content);
      await refreshCurrentProject(selectedId);
      message.success('剧本已保存为新版本'); return true;
    } catch (error) { message.error(error.message || '剧本保存失败'); return false; }
    finally { setWorking(''); }
  }

  async function restoreDocumentVersion(documentId) {
    if(readOnly)return;
    if (!selectedId || working) return;
    try {
      setWorking('document-restore');
      await restoreComicScreenplayDocument(selectedId, documentId);
      await refreshCurrentProject(selectedId);
      setVersionHistoryOpen(false);
      message.success('历史版本已恢复，并保存为新的当前版本');
    } catch (error) { message.error(error.message || '版本恢复失败'); }
    finally { setWorking(''); }
  }

  async function saveSettings() {
    if(readOnly){message.info('当前作品只有查看权限');return;}
    if (!selectedId || settingsPending.current || !creativeReady || creativeLoading || creativeError) return;
    const id = selectedId;
    const saveTicket = ++settingsRequest.current;
    const project = workspace?.project || {};
    let projectSaved = false;
    try {
      const patch = creativeConfigPayload(creative, creativeRevision);
      settingsPending.current = true; setSettingsSaving(true);
      const updated = await updateComicScreenplay(id, {
        title: project.title || '', scriptMode: form.scriptMode, aspectRatio: form.aspectRatio,
        synopsis: form.synopsis ?? project.synopsis ?? '', modelId: form.model,
        promptId: creative.script_prompt_id, assetDegree: form.degree,
        expectedRevision: project.revision || 0,
      });
      projectSaved = true;
      const revision = updated?.revision ?? creativeRevision;
      if (selectedProject.current === id && saveTicket === settingsRequest.current) {
        setCreativeRevision(revision);
        if (updated) setWorkspace(current => current ? { ...current, project: updated } : current);
      }
      const saved = await saveComicCreativeConfig(id, { ...patch, expected_revision: revision });
      if (selectedProject.current !== id || saveTicket !== settingsRequest.current) return;
      setCreative(creativeConfigDraft(saved.creative_config)); setCreativeRevision(saved.revision);
      setStoryboardEngine(current => comicStoryboardEngine(saved.creative_config?.storyboard_prompt_id, current));
      setWorkspace(current => current ? { ...current, project: applyCreativeConfigToProject(current.project, saved.creative_config, saved.revision) } : current);
      setForm(current => ({ ...current, artStyle: saved.creative_config?.style?.name || '', selectedStyle: saved.creative_config?.style?.id || '' }));
      await refreshProjects();
      if(selectedProject.current === id && saveTicket === settingsRequest.current){message.success('设置已保存');setSettingsOpen(false);}
    } catch (error) {
      if (selectedProject.current === id && saveTicket === settingsRequest.current) message.error(`${projectSaved ? '基础设置已保存，创作配置未保存；请保留当前选择重试。' : ''}${error.message || '设置保存失败'}`);
    } finally { if(saveTicket===settingsRequest.current){settingsPending.current = false; setSettingsSaving(false);} }
  }

  async function submitCreate() {
    if (createPending.current || sourceParsing || creativeLoading || creativeError) return;
    const ticket = createTicket.current;
    try {
      const patch = creativeConfigPayload(creative, 0);
      const payload = comicDraftPayload({ ...form, promptId: creative.script_prompt_id }, createContent);
      createPending.current = true; setCreating(true);
      const project = await creationFlow.current.run(payload, patch, created => {
        createdDraft.current = created; setCreatedDraftId(created.id);
      });
      if (ticket !== createTicket.current) return;
      createdDraft.current = null; setCreatedDraftId(''); setCreateOpen(false);
      await refreshProjects(); await openProject(project.id);
    } catch (error) {
      message.error(`${createdDraft.current ? '作品已建立，创作配置未保存。再次点击只会补存当前作品；请勿重复新建。' : ''}${error.message || '创建失败'}`);
    } finally { createPending.current = false; setCreating(false); }
  }

  function cancelSourceImport() {
    sourceImportGate.current.cancel(); setSourceParsing(false);
  }

  function closeCreateDialog() {
    if (creating) return;
    createTicket.current += 1; creativeTicket.current += 1;
    cancelSourceImport(); setCreateOpen(false);
  }

  async function importDraftFile(file) {
    setSourceParsing(true);
    let current = false;
    try {
      const result = await sourceImportGate.current.load(file);
      current = result.current;
      if (!current) return;
      if (!result.content.trim()) throw new Error('文件中没有可用正文，请检查后重新导入');
      setCreateContent(result.content); setSourceFilename(file.name);
      setForm(previous => ({ ...previous, title: previous.title || file.name.replace(/\.(txt|docx)$/i, '') }));
    } catch (error) { current = true; message.error(error.message || '文件读取失败'); }
    finally { if (current) setSourceParsing(false); }
  }

  async function runTask(key, task) {
    if(!canGenerate){message.info('尚未获准使用作品拥有者的生成服务');return;}
    if (working) return;
    const controller = new AbortController(); abortRef.current = controller; setWorking(key); setTaskWarnings([]);
    try {
      const result = await task(controller.signal);
      if(selectedProject.current!==selectedId)return;await refreshCurrentProject(selectedId);
      const warnings = collectWarnings(result);
      const partialFailure = describePartialFailure(result);
      if (partialFailure) {
        setTaskWarnings(warnings);
        message.error(`${partialFailure}，已成功生成的部分已保存；请检查失败内容后重试`);
      } else if (warnings.length) {
        setTaskWarnings(warnings);
        message.warning(`处理完成，但有 ${warnings.length} 条警告`);
      } else {
        message.success('处理完成');
      }
    }
    catch (error) {
      if (error?.name === 'AbortError') message.info('已停止等待，请刷新核对已保存结果后再决定是否重试');
      else message.error(error.message || '处理失败');
    }
    finally { if (abortRef.current === controller) abortRef.current = null; setWorking(''); }
  }

  function startStoryboard(episodeNo = storyboardEpisode) {
    const episodes = splitDocumentEpisodes(workspace?.document?.content);
    const fallback = String(workspace?.document?.content || source || '').trim();
    if (!episodes.length && !fallback) { message.error('请先生成或输入剧本'); return; }
    const project = workspace?.project || {};
    if(!String(project.model_id||'').trim()){message.error('请先在页头选择并保存文本模型');return;}
    const payload = { engine: comicStoryboardEngine(project.creative_config?.storyboard_prompt_id, storyboardEngine), style: project.art_style || '', synopsis: project.synopsis || '', model: project.model_id };
    if (!episodes.length) {
      payload.episodeNo = 1; payload.content = fallback;
    } else if (episodeNo === 0) {
      payload.episodes = Object.fromEntries(episodes.map(item => [item.no, item.content]));
    } else {
      const selected = episodes.find(item => item.no === episodeNo) || episodes[0];
      payload.episodeNo = selected.no; payload.content = selected.content;
    }
    return runTask('storyboard', signal => generateComicStoryboard(selectedId, { ...payload, signal }));
  }

  async function downloadExport() {
    try {
      setWorking('export');
      const data = await exportComicScreenplay(selectedId);
      const url = URL.createObjectURL(new Blob([JSON.stringify(data, null, 2)], { type: 'application/json;charset=utf-8' }));
      const link = document.createElement('a'); link.href = url; link.download = `${workspace?.project?.title || '漫剧项目'}.json`; link.click(); URL.revokeObjectURL(url);
      message.success('项目数据已导出');
    } catch (error) { message.error(error.message || '导出失败'); }
    finally { setWorking(''); }
  }

  async function saveShot() {
    if(readOnly){message.info('当前作品只有查看权限');return;}
    if (!shotEditor) return;
    try {
      setWorking('shot-save');
      const input = { episodeNo: Number(shotEditor.episodeNo) || 1, heading: shotEditor.heading, scriptText: shotEditor.scriptText, characters: String(shotEditor.characters || '').split(/[、,，]/).map(item => item.trim()).filter(Boolean), scene: shotEditor.scene, props: String(shotEditor.props || '').split(/[、,，]/).map(item => item.trim()).filter(Boolean) };
      if (shotEditor.id) await updateComicScreenplayShot(selectedId, shotEditor.id, input);
      else await createComicScreenplayShot(selectedId, input);
      setShotEditor(null);
      await refreshCurrentProject(selectedId);
      message.success(shotEditor.id ? '场次已保存' : '场次已新增');
    } catch (error) { message.error(error.message || '场次保存失败'); }
    finally { setWorking(''); }
  }

  async function moveShot(shotId, direction) {
    if(readOnly){message.info('当前作品只有查看权限');return;}
    try {
      setWorking(`shot-${direction}`);
      await moveComicScreenplayShot(selectedId, shotId, direction);
      await refreshCurrentProject(selectedId);
      message.success(direction === 'up' ? '已上移' : '已下移');
    } catch (error) { message.error(error.message || '移动失败'); }
    finally { setWorking(''); }
  }

  async function removeShot(shotId) {
    if(readOnly){message.info('当前作品只有查看权限');return;}
    try {
      setWorking('shot-delete');
      await deleteComicScreenplayShot(selectedId, shotId);
      await refreshCurrentProject(selectedId);
      message.success('场次已删除');
    } catch (error) { message.error(error.message || '删除失败'); }
    finally { setWorking(''); }
  }

  function replaceStoryboard(updated) {
    if (!updated) return;
    setWorkspace(current => current ? {
      ...current,
      storyboards: (current.storyboards || []).map(item => item.id === updated.id ? updated : item),
    } : current);
  }

  async function saveStoryboardShot() {
    if(readOnly){message.info('当前作品只有查看权限');return;}
    if (!storyboardShotEditor || working) return;
    const editor = storyboardShotEditor;
    const input = {
      afterShotId: editor.afterShotId || '',
      sceneId: Number(editor.sceneId) || 0,
      durationSec: Number(editor.durationSec),
      characters: editor.characters || '',
      scene: editor.scene || '',
      props: editor.props || '',
      videoDesc: editor.videoDesc || '',
    };
    try {
      setWorking('storyboard-shot-save');
      const updated = editor.id
        ? await updateComicStoryboardShot(selectedId, editor.engine, editor.episodeNo, editor.id, input)
        : await createComicStoryboardShot(selectedId, editor.engine, editor.episodeNo, input);
      replaceStoryboard(updated);
      setStoryboardShotEditor(null);
      message.success(editor.id ? '分镜已保存' : '分镜已新增');
    } catch (error) { message.error(error.message || '分镜保存失败'); }
    finally { setWorking(''); }
  }

  async function runStoryboardShotAction(key, action, successText) {
    if(readOnly){message.info('当前作品只有查看权限');return;}
    if (working) return;
    try {
      setWorking(key);
      const updated = await action();
      replaceStoryboard(updated);
      message.success(successText);
    } catch (error) { message.error(error.message || '分镜操作失败'); }
    finally { setWorking(''); }
  }

  async function saveAsset() {
    if(readOnly){message.info('当前作品只有查看权限');return;}
    if (!assetEditor) return;
    try {
      setWorking('asset-save');
      const input = { type: assetEditor.type, degree: assetEditor.degree || form.degree, name: assetEditor.name, identity: assetEditor.identity, description: assetEditor.description, prompt: assetEditor.prompt, voice: assetEditor.voice, episodes: assetEditor.episodes, tag: assetEditor.tag, expectedRevision: assetEditor.revision || 0 };
      if (assetEditor.id) await updateComicScreenplayAsset(selectedId, assetEditor.id, input);
      else await createComicScreenplayAsset(selectedId, input);
      setAssetEditor(null);
      await refreshCurrentProject(selectedId);
      message.success(assetEditor.id ? '资产已保存' : '资产已新增');
    } catch (error) { message.error(error.message || '资产保存失败'); }
    finally { setWorking(''); }
  }

  async function removeAsset(assetId) {
    if(readOnly){message.info('当前作品只有查看权限');return;}
    try {
      setWorking('asset-delete');
      await deleteComicScreenplayAsset(selectedId, assetId);
      await refreshCurrentProject(selectedId);
      message.success('资产已删除');
    } catch (error) { message.error(error.message || '资产删除失败'); }
    finally { setWorking(''); }
  }

  async function bulkEdit(board,operation,order){
    if(readOnly||working)return;
    const ids=selectedShots[board.id]||[];
    if(!ids.length&&operation!=='reorder'){message.info('请先选择镜头');return;}
    try{setWorking('shots-bulk');const result=await batchEditComicShots(selectedId,board.engine,board.episode_no,{expected_revision:board.revision,operation,shot_ids:ids,duration:bulkDuration,order});replaceStoryboard(result.storyboard||result);setSelectedShots(current=>({...current,[board.id]:[]}));message.success('分镜修改已保存');}catch(error){message.error(error.message||'分镜修改失败');}finally{setWorking('');}
  }

  const visibleProjects = useMemo(() => {
    return projects
      .filter(item => item.title?.toLowerCase().includes(query.trim().toLowerCase()))
      .sort((a, b) => (sortOrder === 'desc' ? -1 : 1) * (new Date(a.updated_at) - new Date(b.updated_at)));
  }, [projects, query, scope, sortOrder]);

  const documentEpisodes = useMemo(() => splitDocumentEpisodes(workspace?.document?.content), [workspace?.document?.content]);

  if (selectedId) {
    const project = workspace?.project || projects.find(item => item.id === selectedId) || {};
    const assets = workspace?.assets || [];
    const sourceTools = (<div className="screenplay-editor-grid"><div className="editor-pane"><div className="pane-title">原文</div><Input.TextArea readOnly={readOnly} value={source} onChange={event => setSource(event.target.value)} onBlur={() => { if (!readOnly && selectedId && source && source!==workspace?.project?.source_text) void persistSource().catch(error => message.error(error.message || '原文保存失败')); }} placeholder="粘贴小说正文，或上传 TXT / DOCX 文件" autoSize={{ minRows: 18 }} /><div className="asset-actions"><Upload disabled={readOnly||!!working} accept=".txt,.docx" beforeUpload={file => {if(readOnly)return false; void parseComicScreenplaySource(selectedId, file).then(async content => { const id=selectedId;const project=await saveComicScreenplaySource(id,content);if(selectedProject.current===id){setSource(content);setWorkspace(current=>current?{...current,project}:current);message.success('原文已读取并保存');} }).catch(error => message.error(error.message || '文件读取失败')); return false; }} showUploadList={false}><Button disabled={readOnly||!!working} icon={<UploadCloud size={16} />}>上传 TXT / DOCX</Button></Upload><Button disabled={readOnly} loading={sourceSaving} onClick={saveSource}>保存原文</Button><Button disabled={!canGenerate} type="primary" icon={<Sparkles size={16} />} loading={working === 'script'} onClick={() => runTask('script', async signal => { if(!readOnly)await persistSource(); return generateComicScreenplay(selectedId, { mode: form.scriptMode, content: source, model: form.model, promptId: workspace?.project?.prompt_id, signal }); })}>生成分集剧本</Button></div></div><div className="editor-pane"><div className="pane-title"><span>当前剧本 · v{workspace?.document?.version || 0}</span><Button size="small" icon={<History size={14} />} onClick={() => setVersionHistoryOpen(true)}>版本历史</Button></div><Input.TextArea readOnly={readOnly} className="screenplay-output screenplay-document-editor" value={documentDraft} onChange={event => setDocumentDraft(event.target.value)} placeholder="生成后可在这里继续修改剧本" autoSize={{ minRows: 18 }} /><div className="asset-actions"><Button type="primary" loading={working === 'document-save'} disabled={readOnly||!documentDraft.trim()} onClick={() => saveDocumentDraft()}>保存新版本</Button></div></div></div>);
    const storyboardEditor = (<div>
              <div className="asset-actions">
                <Select style={{ minWidth: 150 }} value={storyboardEpisode} onChange={setStoryboardEpisode} options={documentEpisodes.length ? [...documentEpisodes.map(item => ({ value: item.no, label: `第 ${item.no} 集` })), { value: 0, label: '全部集' }] : [{ value: 1, label: '第 1 集（全文）' }]} />
                <Select aria-label="分镜引擎" title="引擎由所选分镜提示词确定；修改提示词后请先保存设置" disabled={!!comicStoryboardEngine(creative.storyboard_prompt_id)} value={comicStoryboardEngine(creative.storyboard_prompt_id, storyboardEngine)} onChange={setStoryboardEngine} options={[{value:'xiaoluo_v9_30s',label:'小罗 V9 · 30秒镜头组'},{value:'xiaoluo_v7_15s',label:'小罗 · 15秒镜头组'},{value:'xiaoyunque_v3',label:'小云雀 V3 分镜'}]}/><Button title={!String(project.model_id||'').trim()?'请先在页头选择并保存文本模型':creative.storyboard_prompt_id !== workspace?.project?.creative_config?.storyboard_prompt_id ? '请先保存分镜提示词选择' : undefined} disabled={!canGenerate || !String(project.model_id||'').trim() || creative.storyboard_prompt_id !== workspace?.project?.creative_config?.storyboard_prompt_id} type="primary" loading={working === 'storyboard'} onClick={() => startStoryboard()}>生成 CM 分镜</Button>
                <Button disabled={readOnly||!!working} icon={<Plus size={16} />} onClick={() => setShotEditor({ episodeNo: storyboardEpisode || 1, heading: '', scriptText: '', characters: '', scene: '', props: '' })}>新增场次</Button>
              </div>
              <p style={{ margin: '0 0 12px', color: '#8a8f99' }}>{documentEpisodes.length ? '按集生成：只提交选中集的内容；选「全部集」会逐集提交。' : '未检测到「# 第N集」标题，将按第 1 集提交全文。'}</p>
              <div className="shot-editor-list">{(workspace?.shots || []).length ? workspace.shots.filter(shot => !editingEpisode || shot.episode_no === editingEpisode).map((shot, index) => <article key={shot.id} className="shot-editor-card">
                <div><span>第 {shot.episode_no} 集 · 场次 {shot.sort_order}</span><b>{shot.heading || '未命名场次'}</b><p>{shot.script_text || '暂无内容'}</p></div>
                <div className="shot-editor-actions">
                  <Button aria-label="上移场次" icon={<ArrowUp size={15} />} disabled={readOnly || working || index === 0 || workspace.shots[index - 1]?.episode_no !== shot.episode_no} onClick={() => moveShot(shot.id, 'up')} />
                  <Button aria-label="下移场次" icon={<ArrowDown size={15} />} disabled={readOnly || working || index === workspace.shots.length - 1 || workspace.shots[index + 1]?.episode_no !== shot.episode_no} onClick={() => moveShot(shot.id, 'down')} />
                  <Button disabled={readOnly||!!working} aria-label="编辑场次" icon={<Pencil size={15} />} onClick={() => setShotEditor({ id: shot.id, episodeNo: shot.episode_no, heading: shot.heading, scriptText: shot.script_text, characters: (shot.characters || []).join('、'), scene: shot.scene || '', props: (shot.props || []).join('、') })} />
                  <Popconfirm title="删除这个场次？" description="删除后不可恢复" onConfirm={() => removeShot(shot.id)}><Button disabled={readOnly||!!working} danger aria-label="删除场次" icon={<Trash2 size={15} />} /></Popconfirm>
                </div>
              </article>) : <Empty description="尚未生成或新增场次" />}</div>
              <div className="storyboard-list">{(workspace?.storyboards || []).length ? (workspace.storyboards || []).filter(board => (!editingEpisode || board.episode_no === editingEpisode) && board.engine === comicStoryboardEngine(project.creative_config?.storyboard_prompt_id, storyboardEngine)).map(item => <section className="storyboard-card" key={item.id}>
                <header><div><b>第 {item.episode_no} 集分镜</b><span>{item.shot_count} 个镜头组 · 版本 {item.revision} · {item.status}</span></div><Button size="small" icon={<Plus size={14} />} disabled={readOnly || Boolean(working)} onClick={() => setStoryboardShotEditor({ engine: item.engine, episodeNo: item.episode_no, afterShotId: item.shots?.at(-1)?.id || '', durationSec: 3, sceneId: 0, characters: '', scene: '', props: '', videoDesc: '' })}>末尾新增</Button></header>
                <div className="comic-bulk-tools"><Checkbox disabled={readOnly} checked={!!item.shots?.length&&(selectedShots[item.id]||[]).length===item.shots.length} onChange={event=>setSelectedShots(current=>({...current,[item.id]:event.target.checked?item.shots.map(shot=>shot.id):[]}))}>全选镜头</Checkbox><span>已选 {(selectedShots[item.id]||[]).length}</span><InputNumber disabled={readOnly} aria-label="批量镜头时长" min={.1} max={120} step={.1} value={bulkDuration} onChange={value=>setBulkDuration(value||3)}/><Button disabled={readOnly||!!working} onClick={()=>bulkEdit(item,'duration')}>统一时长</Button><Button disabled={readOnly||!!working} onClick={()=>bulkEdit(item,'duplicate')}>复制选中</Button><Popconfirm title="删除选中的镜头？" onConfirm={()=>bulkEdit(item,'delete')}><Button danger disabled={readOnly||!!working}>删除选中</Button></Popconfirm></div><div className="storyboard-shot-list">{(item.shots || []).length ? item.shots.map((shot, index) => <article className="storyboard-shot-card" key={shot.id}>
                  <div className="storyboard-shot-copy"><Checkbox disabled={readOnly} checked={(selectedShots[item.id]||[]).includes(shot.id)} onChange={event=>setSelectedShots(current=>({...current,[item.id]:event.target.checked?[...(current[item.id]||[]),shot.id]:(current[item.id]||[]).filter(id=>id!==shot.id)}))}>选择</Checkbox><span>镜头 {shot.seq} · {Number(shot.duration_sec || 0).toFixed(1)} 秒</span><b>{shot.scene || '未命名场景'}</b><p>{shot.video_desc || '暂无画面描述'}</p><small>角色：{shot.characters || '无'} · 道具：{shot.props || '无'}</small></div>
                  <div className="storyboard-shot-actions"><Button disabled={readOnly||!!working||index===0} onClick={()=>{const order=item.shots.map(row=>row.id);[order[index],order[index-1]]=[order[index-1],order[index]];void bulkEdit(item,'reorder',order);}}>上移</Button>
                    <Button disabled={readOnly||!!working} size="small" onClick={() => setStoryboardShotEditor({ id: shot.id, engine: item.engine, episodeNo: item.episode_no, durationSec: shot.duration_sec, sceneId: shot.scene_id, characters: shot.characters || '', scene: shot.scene || '', props: shot.props || '', videoDesc: shot.video_desc || '' })}>编辑</Button>
                    <Button size="small" disabled={readOnly || Boolean(working)} onClick={() => setStoryboardShotEditor({ engine: item.engine, episodeNo: item.episode_no, afterShotId: shot.id, durationSec: 3, sceneId: shot.scene_id, characters: '', scene: shot.scene || '', props: '', videoDesc: '' })}>在后面新增</Button>
                    <Button size="small" disabled={readOnly || Boolean(working) || Number(shot.duration_sec) <= 0.4} onClick={() => runStoryboardShotAction('storyboard-shot-split', () => splitComicStoryboardShot(selectedId, item.engine, item.episode_no, shot.id), '分镜已等分拆开')}>拆分</Button>
                    <Button size="small" disabled={readOnly || Boolean(working) || index === item.shots.length - 1} onClick={() => runStoryboardShotAction('storyboard-shot-merge', () => mergeComicStoryboardShot(selectedId, item.engine, item.episode_no, shot.id), '已与下一条分镜合并')}>与下一条合并</Button>
                    <Popconfirm title="删除这条分镜？" description="删除后会重新排列后续镜头序号" onConfirm={() => runStoryboardShotAction('storyboard-shot-delete', () => deleteComicStoryboardShot(selectedId, item.engine, item.episode_no, shot.id), '分镜已删除')}><Button size="small" danger disabled={readOnly || Boolean(working)}>删除</Button></Popconfirm>
                  </div>
                </article>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="这份分镜还没有镜头，可从末尾新增" />}</div>
              </section>) : <Empty description="尚未生成分镜" />}</div>
            </div>);
    const settingsFields = (<fieldset disabled={readOnly} style={{border:0,padding:0,margin:0}}><h3>模型与创作配置</h3><label>文本模型<Select showSearch optionFilterProp="label" options={models.map(model=>({value:model.id,label:model.displayName||model.name||model.id}))} value={form.model} placeholder="使用账号默认模型" onChange={model => setForm(current => ({ ...current, model }))} /></label><label>画面比例<Select value={form.aspectRatio} options={ratioOptions} onChange={aspectRatio => setForm(current => ({ ...current, aspectRatio }))} /></label><ComicCreativeFields value={creative} onChange={patch => { setCreative(current => ({ ...current, ...patch })); if (patch.storyboard_prompt_id) setStoryboardEngine(current => comicStoryboardEngine(patch.storyboard_prompt_id, current)); }} library={creativeLibrary} loading={creativeLoading} error={creativeError} onRetry={() => readProjectCreative(selectedId)} onLibraryChange={readCreativeLibrary} disabled={readOnly || settingsSaving || !creativeReady} /><label>故事概要<Input.TextArea rows={3} value={form.synopsis || ''} onChange={event=>setForm(current=>({...current,synopsis:event.target.value}))}/></label><div className="comic-inspector-field"><span>资产解析程度</span><ComicDegreeCards compact value={form.degree} onChange={degree => setForm(current => ({ ...current, degree }))} /></div><label>创作模式<Select value={form.scriptMode} options={[{ value: 'a1', label: '小说转剧本' }, { value: 'b', label: '导演级剧本' }]} onChange={scriptMode => setForm(current => ({ ...current, scriptMode }))} /></label><Button type="primary" block loading={settingsSaving} disabled={readOnly || !creativeReady || creativeLoading || !!creativeError} onClick={saveSettings}>保存设置</Button><div className="workspace-stats"><span>版本 <b>{workspace?.documents?.length || 0}</b></span><span>场次 <b>{workspace?.shots?.length || 0}</b></span><span>角色 <b>{workspace?.characters?.length || 0}</b></span><span>场景 <b>{workspace?.scenes?.length || 0}</b></span></div></fieldset>);
    const next = comicNextState(activeTab, workspace, documentDraft !== (workspace?.document?.content || ''));
    const savedStyle = { style_id: project.creative_config?.style?.id || null, style_name: comicProjectStyleName(project) };
    async function savePillStyle(patch) {
      if(readOnly||settingsPending.current)return;
      const id=selectedId,ticket=++settingsRequest.current,current=()=>selectedProject.current===id&&ticket===settingsRequest.current;
      settingsPending.current=true;setSettingsSaving(true);
      try { const saved = await saveComicCreativeConfig(id,{style_id:patch.style_id,expected_revision:workspace?.project?.revision || creativeRevision});if(!current())return;setCreative(creativeConfigDraft(saved.creative_config));setCreativeRevision(saved.revision);setWorkspace(value=>value?{...value,project:applyCreativeConfigToProject(value.project,saved.creative_config,saved.revision)}:value);await refreshProjects();if(current())message.success('画风已保存'); }catch(error){if(current())message.error(error.message||'画风保存失败');}finally{if(current()){settingsPending.current=false;setSettingsSaving(false);}}
    }
    async function savePillModel(model) {
      if(readOnly||settingsPending.current)return;
      const id=selectedId,ticket=++settingsRequest.current,current=()=>selectedProject.current===id&&ticket===settingsRequest.current;
      settingsPending.current=true;setSettingsSaving(true);
      try { const updated=await updateComicScreenplay(id,{title:project.title,scriptMode:project.script_mode,aspectRatio:project.aspect_ratio,synopsis:project.synopsis||'',modelId:model,promptId:project.prompt_id,assetDegree:project.asset_degree,expectedRevision:project.revision});if(!current())return;setWorkspace(value=>value?{...value,project:updated}:value);setForm(value=>({...value,model}));setCreativeRevision(updated.revision);setModelSearchOpen(false);message.success('文本模型已保存'); }catch(error){if(current())message.error(error.message||'模型保存失败');}finally{if(current()){settingsPending.current=false;setSettingsSaving(false);}}
    }
    async function nextStage() {
      if(!next)return;
      const id=selectedId,ticket=projectRequest.current,current=()=>selectedProject.current===id&&ticket===projectRequest.current;
      if(activeTab==='assets'){
        try{const media=await getComicMedia(id);if(!current())return;const missing=assets.filter(asset=>!media.settings?.selected?.[asset.id]&&!media.items?.some(item=>item.kind==='image'&&(item.targetId===asset.id||item.target_id===asset.id)&&['succeeded','ready','completed'].includes(item.status)));if(missing.length){Modal.confirm({title:'部分资产尚未生成',content:<div><p>这些资产缺少图片，后续仍可继续补充：</p><ul>{missing.map(asset=><li key={asset.id}>{asset.name}</li>)}</ul></div>,okText:'忽略，下一步',cancelText:'返回生成',onOk:()=>{if(current())setActiveTab('storyboard');}});return;}}catch(error){if(current())message.error(error.message||'资产图片核对失败，请重试');return;}
      }
      if(current())setActiveTab(next.next);
    }
    return <div className={`comic-screenplay-page comic-screenplay-workbench comic-three-stage-workbench layout-${display.editor_version||'studio'} ${display.sidebar_collapsed?'sidebar-collapsed':''} ${display.show_descriptions===false?'hide-descriptions':''}`} style={{'--comic-body-size':`${display.font_size||14}px`}}>
      <header className="comic-original-titlebar"><button type="button" aria-label="返回项目库" title="返回项目库" onClick={() => {abortRef.current?.abort();projectRequest.current+=1;creativeTicket.current+=1;selectedProject.current='';setSelectedId('');setWorkspace(null);setAccess(null);void refreshProjects();}}><OriginalIcon name="ArrowBackOutline" size={18}/></button><h1>AI 推理分析</h1><span>{project.title || '新项目'}</span><div className="comic-title-tools">{working && abortRef.current ? <Button size="small" danger onClick={()=>abortRef.current?.abort()}>取消任务</Button>:null}<Dropdown menu={{items:[{key:'refresh',label:'刷新作品'},{key:'display',label:'显示设置'},...(access?.owner?[{key:'team',label:'团队权限'}]:[])],onClick:({key})=>{if(key==='refresh')void openProject(selectedId);if(key==='display')setDisplayOpen(true);if(key==='team')setTeamOpen(true);}}}><Button size="small">更多</Button></Dropdown></div></header>
      <div className="comic-top-main-tab-row"><div className="comic-top-config-bar"><button type="button" className="comic-model-trigger" title={project.model_id || '选择文本模型'} disabled={readOnly||settingsSaving} onClick={()=>{setModelQuery('');setModelSearchOpen(true);void listAvailableModels('text').then(setModels).catch(error=>message.error(error.message));}}><ComicScreenplayModelIcon modelId={models.find(model=>model.id===project.model_id)?.modelId||project.model_id} size={18}/><span>{models.find(model=>model.id===project.model_id)?.displayName || project.model_id || '选择模型'}</span><OriginalIcon name="ChevronDownOutline" size={11}/></button><button type="button" className="comic-top-pill" title="模型配置" onClick={openSettings}><OriginalIcon name="SettingsOutline" size={13}/><span>模型配置</span><OriginalIcon name="ChevronDownOutline" size={12}/></button><ComicStylePicker compact value={savedStyle} onChange={savePillStyle} styles={creativeLibrary.styles} disabled={readOnly||settingsSaving||!creativeReady} onLibraryChange={readCreativeLibrary}/><button type="button" className={`comic-top-pill${agentOpen?' is-active':''}`} aria-pressed={agentOpen} title="Agent" onClick={()=>setAgentOpen(value=>!value)}><span>Agent</span></button></div><nav className="comic-top-step-tabs" aria-label="创作阶段">{comicStages.map((stage,index)=><span className="comic-top-step-slot" key={stage.key}>{index>0?<span className="comic-top-step-connector"/>:null}<button type="button" aria-current={activeTab===stage.key?'step':undefined} className={`comic-top-step-tab${activeTab===stage.key?' is-active':''}`} onClick={()=>setActiveTab(stage.key)}><i>{index+1}</i><span>{stage.label}</span></button></span>)}</nav><div aria-hidden="true" className="comic-top-row-spacer"/></div>
      <div className="comic-three-stage-body"><main className="comic-stage-main"><div className="comic-stage-scroll">
      {taskWarnings.length?<Alert type="warning" showIcon closable onClose={()=>setTaskWarnings([])} message={`本次处理有 ${taskWarnings.length} 条警告`} description={<ul>{taskWarnings.map((item,index)=><li key={index}>{item}</li>)}</ul>}/>:null}
      <div hidden={activeTab!=='source'}><ComicScreenplayStoryStage key={selectedId} project={project} document={workspace?.document} draft={documentDraft} onDraft={setDocumentDraft} onSave={saveDocumentDraft} onTools={openTool} readOnly={readOnly} working={working}/></div>
      <div hidden={activeTab!=='assets'}><ComicScreenplayAssetStage projectId={selectedId} workspace={workspace} readOnly={readOnly} canGenerate={canGenerate} working={working} degree={form.degree} onDegree={degree=>setForm(current=>({...current,degree}))} onAdd={type=>setAssetEditor({type,degree:form.degree,name:'',identity:'',description:'',prompt:'',voice:'',episodes:'',tag:''})} onEdit={asset=>setAssetEditor({id:asset.id,revision:asset.revision,type:asset.type,degree:asset.degree,name:asset.name,identity:asset.identity||'',description:asset.description||'',prompt:asset.prompt||'',voice:asset.voice||'',episodes:asset.episodes||'',tag:asset.tag||''})} onDelete={id=>Modal.confirm({title:'删除这个资产？',content:'删除后不可恢复。',okText:'删除',cancelText:'取消',onOk:()=>removeAsset(id)})} onExtract={type=>runTask(type,signal=>extractComicAssets(selectedId,{type,degree:form.degree,content:workspace?.document?.content,model:project.model_id,signal}))} onMedia={context=>openTool('media',context)}/></div>
      {activeTab==='storyboard'?<ComicScreenplayStoryboardStage engine={comicStoryboardEngine(project.creative_config?.storyboard_prompt_id,storyboardEngine)} modelId={project.model_id} workspace={workspace} canGenerate={canGenerate} working={working} promptName={creativeLibrary.prompts.find(item=>item.id===project.creative_config?.storyboard_prompt_id)?.name} onGenerate={startStoryboard} onEditor={episode=>{setEditingEpisode(episode);openTool('storyboard-editor');}} onTools={openTool} onView={episode=>Modal.info({title:`第${episode.no}集 · ${episode.title}`,width:800,content:<pre className="comic-numbered-source">{episode.content}</pre>})}/>:null}
      </div>{next?<div className="comic-flow-next-bar"><span className="comic-flow-next-mark"><OriginalIcon name="SparklesOutline" size={16}/></span><div><b>{next.title}</b><small>{next.subtitle}</small></div><button type="button" onClick={nextStage}>下一步 <span aria-hidden="true">→</span></button></div>:null}</main>
      {agentOpen?<aside className="comic-stage-agent"><header><b>创作 Agent</b><Button size="small" type="text" aria-label="关闭Agent" icon={<X size={15}/>} onClick={()=>setAgentOpen(false)}/></header><ComicAssistantPanel key={selectedId} projectId={selectedId} readOnly={readOnly} canGenerate={canGenerate} onSaved={()=>refreshCurrentProject(selectedId)}/></aside>:null}</div>
      <Modal closeIcon={<OriginalIcon name="CloseOutline" size={16}/>} className="comic-model-config-modal" width={1176} style={{maxWidth:'96vw'}} title="模型配置" open={settingsOpen} onCancel={closeSettings} footer={<><Button disabled={settingsSaving} onClick={closeSettings}>取消</Button><Button type="primary" loading={settingsSaving} disabled={readOnly||!creativeReady||creativeLoading||!!creativeError} onClick={saveSettings}>保存配置</Button></>}>{settingsFields}</Modal>
      <Modal className="comic-original-model-dialog" width={420} style={{maxWidth:'92vw'}} closable={false} title={<div className="comic-original-model-search"><span className="comic-icon-sr-only">选择文本模型</span><OriginalIcon name="SearchOutline" size={16} style={{opacity:.4}}/><Input variant="borderless" value={modelQuery} onChange={event=>setModelQuery(event.target.value)} placeholder="搜索模型..." allowClear/><button type="button" aria-label="关闭模型选择" onClick={()=>setModelSearchOpen(false)}><OriginalIcon name="InlineModelDialogClose" size={14}/></button></div>} open={modelSearchOpen} onCancel={()=>setModelSearchOpen(false)} footer={null}><div className="comic-model-options">{models.filter(model=>`${model.displayName||model.name||''} ${model.id}`.toLowerCase().includes(modelQuery.toLowerCase())).map(model=><button type="button" key={model.id} disabled={settingsSaving||readOnly} onClick={()=>savePillModel(model.id)}><ComicScreenplayModelIcon modelId={model.modelId||model.id} size={24}/><span>{model.displayName||model.name||model.id}</span>{project.model_id===model.id?<OriginalIcon name="CheckmarkOutline" size={16} style={{color:'#3b82f6'}}/>:null}</button>)}</div>{!models.length?<Empty description="暂无可用文本模型"/>:null}</Modal>
      <Drawer className="comic-stage-tool-drawer" width="min(1180px,96vw)" title={({source:'原文与剧本工具',chapters:'章节规划',media:'媒体与剪辑',production:'自动制作任务',files:'分镜文件工具','storyboard-editor':`第${editingEpisode||1}集分镜编辑器`})[tool]} open={!!tool} onClose={()=>setTool('')} destroyOnClose={false}>
        <div hidden={tool!=='source'}>{sourceTools}</div>
        {openedTools.chapters?<div hidden={tool!=='chapters'}><ComicChaptersPanel key={selectedId} projectId={selectedId} readOnly={readOnly} canGenerate={canGenerate} onSaved={()=>refreshCurrentProject(selectedId)}/></div>:null}
        {tool==='files'?<ComicScreenplayFileTools key={selectedId} projectId={selectedId} workspace={workspace} readOnly={readOnly} onSaved={()=>refreshCurrentProject(selectedId)}/>:null}
        {tool==='storyboard-editor'?storyboardEditor:null}
        {openedTools.media?<div hidden={tool!=='media'}><ComicScreenplayMediaPanel key={selectedId} projectId={selectedId} workspace={workspace} readOnly={readOnly} canGenerate={canGenerate} canOpenCut={!!access?.owner} initialTab={mediaContext.tab||'images'} initialTargetId={mediaContext.targetId||''} initialBatchTargets={mediaContext.targetIds||[]}/></div>:null}
        {openedTools.production?<div hidden={tool!=='production'}><ComicScreenplayProductionPanel projectId={selectedId} canGenerate={canGenerate} onExportData={downloadExport} onSaved={()=>refreshCurrentProject(selectedId)} onNavigate={navigateStage}/></div>:null}
      </Drawer>
      {teamOpen?<ComicTeamDialog projectId={selectedId} open onClose={()=>setTeamOpen(false)}/>:null}{displayOpen?<ComicDisplayDialog readOnly={readOnly} projectId={selectedId} open onClose={()=>setDisplayOpen(false)} onSaved={prefs=>{setDisplay(prefs);navigateStage(prefs.active_tab||activeTab);}}/>:null}
      <Modal width={960} style={{maxWidth:'92vw'}} title={assetEditor?.id ? '编辑资产' : '新增资产'} open={Boolean(assetEditor)} onCancel={() => setAssetEditor(null)} confirmLoading={working === 'asset-save'} onOk={saveAsset} okText="保存" cancelText="取消" destroyOnHidden>
        <div className="shot-form">
          <label><span>资产类型</span><Select value={assetEditor?.type || 'character'} options={assetTypeOptions} onChange={type => setAssetEditor(current => ({ ...current, type }))} /></label>
          <label><span>名称</span><Input value={assetEditor?.name || ''} placeholder="请输入角色、场景或道具名称" onChange={event => setAssetEditor(current => ({ ...current, name: event.target.value }))} /></label>
          <label><span>身份定位</span><Input value={assetEditor?.identity || ''} placeholder="例如：核心女主角" onChange={event => setAssetEditor(current => ({ ...current, identity: event.target.value }))} /></label>
          <label><span>描述</span><Input.TextArea rows={3} value={assetEditor?.description || ''} onChange={event => setAssetEditor(current => ({ ...current, description: event.target.value }))} /></label>
          <label><span>生成提示词</span><Input.TextArea rows={4} value={assetEditor?.prompt || ''} placeholder="该资产的画面描述，不填写系统提示词" onChange={event => setAssetEditor(current => ({ ...current, prompt: event.target.value }))} /></label>
          {assetEditor?.type === 'character' ? <label><span>音色</span><Input value={assetEditor?.voice || ''} onChange={event => setAssetEditor(current => ({ ...current, voice: event.target.value }))} /></label> : null}
          <label><span>出现集数</span><Input value={assetEditor?.episodes || ''} placeholder="例如：1-5、8" onChange={event => setAssetEditor(current => ({ ...current, episodes: event.target.value }))} /></label>
          <label><span>标签</span><Input value={assetEditor?.tag || ''} onChange={event => setAssetEditor(current => ({ ...current, tag: event.target.value }))} /></label>
        </div>
      </Modal>
      <Modal title={shotEditor?.id ? '编辑场次' : '新增场次'} open={Boolean(shotEditor)} onCancel={() => setShotEditor(null)} confirmLoading={working === 'shot-save'} onOk={saveShot} okText="保存" cancelText="取消" destroyOnHidden>
        <div className="shot-form">
          <label><span>集数</span><Input type="number" min={1} value={shotEditor?.episodeNo || 1} onChange={event => setShotEditor(current => ({ ...current, episodeNo: event.target.value }))} /></label>
          <label><span>场次标题</span><Input value={shotEditor?.heading || ''} placeholder="例如：1-2 夜 外 天台" onChange={event => setShotEditor(current => ({ ...current, heading: event.target.value }))} /></label>
          <label><span>剧本内容</span><Input.TextArea rows={7} value={shotEditor?.scriptText || ''} placeholder="动作、对白和画面说明" onChange={event => setShotEditor(current => ({ ...current, scriptText: event.target.value }))} /></label>
          <label><span>角色</span><Input value={shotEditor?.characters || ''} placeholder="多个角色用顿号分隔" onChange={event => setShotEditor(current => ({ ...current, characters: event.target.value }))} /></label>
          <label><span>场景</span><Input value={shotEditor?.scene || ''} onChange={event => setShotEditor(current => ({ ...current, scene: event.target.value }))} /></label>
          <label><span>道具</span><Input value={shotEditor?.props || ''} placeholder="多个道具用顿号分隔" onChange={event => setShotEditor(current => ({ ...current, props: event.target.value }))} /></label>
        </div>
      </Modal>
      <Modal width={800} title={storyboardShotEditor?.id ? '编辑分镜' : '新增分镜'} open={Boolean(storyboardShotEditor)} onCancel={() => setStoryboardShotEditor(null)} confirmLoading={working === 'storyboard-shot-save'} onOk={saveStoryboardShot} okText="保存" cancelText="取消" destroyOnHidden>
        <div className="shot-form storyboard-shot-form">
          <label><span>时长（秒）</span><Input type="number" min={0.1} step={0.1} value={storyboardShotEditor?.durationSec ?? 3} onChange={event => setStoryboardShotEditor(current => ({ ...current, durationSec: event.target.value }))} /></label>
          <label><span>角色</span><Input value={storyboardShotEditor?.characters || ''} placeholder="例如：小Y、CM" onChange={event => setStoryboardShotEditor(current => ({ ...current, characters: event.target.value }))} /></label>
          <label><span>场景</span><Input value={storyboardShotEditor?.scene || ''} placeholder="例如：夜晚的客厅" onChange={event => setStoryboardShotEditor(current => ({ ...current, scene: event.target.value }))} /></label>
          <label><span>道具</span><Input value={storyboardShotEditor?.props || ''} placeholder="例如：手机、雨伞" onChange={event => setStoryboardShotEditor(current => ({ ...current, props: event.target.value }))} /></label>
          <label className="is-wide"><span>画面描述</span><Input.TextArea rows={5} value={storyboardShotEditor?.videoDesc || ''} placeholder="描述这个镜头里真正发生的画面和动作" onChange={event => setStoryboardShotEditor(current => ({ ...current, videoDesc: event.target.value }))} /></label>
        </div>
      </Modal>
      <Modal title="剧本版本历史" open={versionHistoryOpen} onCancel={() => setVersionHistoryOpen(false)} footer={<Button onClick={() => setVersionHistoryOpen(false)}>关闭</Button>}>
        <div className="screenplay-version-list">
          {(workspace?.documents || []).map(document => <article key={document.id}>
            <div><b>v{document.version}</b><span>{document.status === 'active' ? '当前版本' : `历史版本 · ${prettyTime(document.updated_at || document.created_at)}`}</span></div>
            {document.status === 'active' ? <span>正在使用</span> : <Popconfirm title={`恢复 v${document.version}？`} description="系统会复制成一个新版本，现有版本不会被覆盖。" onConfirm={() => restoreDocumentVersion(document.id)}><Button disabled={readOnly||!!working} size="small" loading={working === 'document-restore'}>恢复此版本</Button></Popconfirm>}
          </article>)}
        </div>
      </Modal>
    </div>;
  }

  return <div className="comic-screenplay-page comic-project-library">
    <div className="comic-screenplay-heading">
      <div><h1>漫剧创作</h1><p>管理和创建您的漫剧作品</p></div>
      <div className="library-heading-actions"><Button type="text" icon={<Clock3 size={18} />} onClick={() => setHistoryOpen(true)}>历史</Button><Button className="create-screenplay-button" type="primary" icon={<Plus size={16} />} onClick={openCreateDialog}>创建漫剧</Button></div>
    </div>
    <div className="library-tabs">{[['personal','个人作品'],['team','团队作品'],['archived','已归档']].map(([key,label])=><button key={key} className={scope===key?'is-active':''} onClick={()=>setScope(key)}>{label}</button>)}</div>
    <div className="comic-screenplay-filters">
      <Input prefix={<Search size={16} />} value={query} onChange={event => setQuery(event.target.value)} placeholder="搜索作品..." allowClear />
      <div className="library-filter-actions"><Select value={sortOrder} suffixIcon={<ChevronDown size={15} />} onChange={setSortOrder} options={[{ value: 'desc', label: '按时间降序' }, { value: 'asc', label: '按时间升序' }]} /><Button type="text" icon={viewMode === 'grid' ? <Grid2X2 size={18} /> : <LayoutGrid size={18} />} onClick={() => setViewMode(value => value === 'grid' ? 'list' : 'grid')} /></div>
    </div>
    <Spin spinning={loading}><div className={`comic-project-grid ${viewMode === 'list' ? 'is-list' : ''}`}>{visibleProjects.length ? visibleProjects.map(project => <article key={project.id} className="comic-project-card" onClick={() => openProject(project.id)}><div className="comic-project-cover"><Clapperboard size={38} /><span>{project.aspect_ratio || '9:16'}</span></div><div className="comic-project-meta"><div><h3>{project.title}</h3><p>{comicProjectStyleName(project) || '未指定画风'} · {prettyTime(project.updated_at)}</p></div>{scope!=='team'?<Button type="text" onClick={async event=>{event.stopPropagation();try{const result=await getComicDisplay(project.id);await (scope==='archived'?restoreComic:archiveComic)(project.id,{expected_revision:result.state?.revision||0});await refreshProjects();}catch(err){message.error(err.message);}}}>{scope==='archived'?'恢复':'归档'}</Button>:null}{scope!=='team'?<Popconfirm title="删除这个漫剧项目？" onConfirm={async event => { event?.stopPropagation?.(); await deleteComicScreenplay(project.id); await refreshProjects(); }}><Button type="text" danger icon={<Trash2 size={16} />} onClick={event => event.stopPropagation()} /></Popconfirm>:null}</div></article>) : <Empty description={scope === 'team' ? '暂无团队作品' : '暂无项目，在上方输入框输入内容开始创作'} />}</div></Spin>
    <Modal title="最近创作历史" open={historyOpen} onCancel={() => setHistoryOpen(false)} footer={<Button onClick={() => setHistoryOpen(false)}>关闭</Button>}>
      <div className="screenplay-version-list">{projects.length ? [...projects].sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at)).map(project => <button type="button" key={project.id} onClick={() => { setHistoryOpen(false); void openProject(project.id); }}><b>{project.title}</b><span>{prettyTime(project.updated_at)}</span></button>) : <Empty description="还没有创作记录" />}</div>
    </Modal>
    <Modal className="screenplay-create-modal" width={760} title="新建漫剧" open={createOpen} closeIcon={<X size={20} />} onCancel={closeCreateDialog} maskClosable={!creating} closable={!creating} footer={<><Button disabled={creating} onClick={closeCreateDialog}>取消</Button><Button type="primary" icon={<FileText size={16} />} loading={creating} disabled={sourceParsing || creativeLoading || !!creativeError} onClick={submitCreate}>{createdDraftId ? '重试保存配置' : '创建作品'}</Button></>}>
      {createdDraftId ? <Alert type="warning" showIcon message="作品已建立，创作配置尚未保存；重试会继续保存这一个作品。" /> : null}
      <div className="comic-create-form">
        <label><span>作品名称 <i>*</i></span><Input disabled={creating || !!createdDraftId} autoFocus value={form.title} maxLength={80} placeholder="请输入剧本名称" onChange={event => setForm(current => ({ ...current, title: event.target.value }))} /></label>
        <label><span>选择模型</span><Select disabled={creating || !!createdDraftId} value={form.model || undefined} placeholder="请选择已配置的文本模型" onChange={model => setForm(current => ({ ...current, model }))} options={models.map(model => ({ value: model.id, label: model.displayName || model.name || model.id }))} /></label>
        <ComicCreativeFields value={creative} onChange={patch => { setCreative(current => ({ ...current, ...patch })); if (patch.storyboard_prompt_id) setStoryboardEngine(current => comicStoryboardEngine(patch.storyboard_prompt_id, current)); }} library={creativeLibrary} loading={creativeLoading} error={creativeError} onRetry={readCreativeLibrary} onLibraryChange={readCreativeLibrary} disabled={creating} />
        <p className="comic-style-caption">建议小说原文生成前往 Agent / Chat 功能，<a href="/agent/chat">查看链接</a></p>
        <label><span>剧本比例</span><Select disabled={creating || !!createdDraftId} value={form.aspectRatio} options={ratioOptions} onChange={aspectRatio => setForm(current => ({ ...current, aspectRatio }))} /></label>
        <fieldset disabled={creating || !!createdDraftId} style={{ border: 0, margin: 0, padding: 0 }}><div className="degree-field"><span className="comic-field-title">资产解析程度</span><ComicDegreeCards value={form.degree} onChange={degree => setForm(current => ({ ...current, degree }))} /></div>
        <ComicSourceInput mode={sourceMode} onModeChange={mode => { cancelSourceImport(); setSourceMode(mode); }} content={createContent} onContentChange={text => { cancelSourceImport(); setCreateContent(text); }} filename={sourceFilename} parsing={sourceParsing} disabled={creating || !!createdDraftId} onFile={importDraftFile} onRemove={() => { cancelSourceImport(); setSourceFilename(''); }} /></fieldset>
      </div>
    </Modal>
  </div>;
}
