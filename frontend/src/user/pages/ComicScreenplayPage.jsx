import { formatWorkDateTime } from '../../shared/workDateTime.js';
import { Alert, Button, Empty, Input, message, Modal, Popconfirm, Select, Spin, Tabs, Upload } from 'antd';
import { ArrowDown, ArrowLeft, ArrowUp, Check, ChevronDown, Clapperboard, Clock3, Download, FileText, Grid2X2, HelpCircle, History, LayoutGrid, Palette, PauseCircle, Pencil, Play, Plus, Search, Sparkles, Trash2, UploadCloud, Users, Map, Package, WandSparkles, X } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { cancelComicScreenplayBatch, createComicScreenplay, createComicScreenplayAsset, createComicScreenplayShot, createComicStoryboardShot, deleteComicScreenplay, deleteComicScreenplayAsset, deleteComicScreenplayShot, deleteComicStoryboardShot, exportComicScreenplay, extractComicAssets, generateComicScreenplay, generateComicStoryboard, getComicScreenplayBatch, listComicScreenplayPrompts, listComicScreenplays, loadComicScreenplayWorkspace, mergeComicStoryboardShot, moveComicScreenplayShot, parseComicScreenplaySource, restoreComicScreenplayDocument, runComicScreenplayBatchNext, saveComicScreenplayDocument, saveComicScreenplaySource, splitComicStoryboardShot, startComicScreenplayBatch, updateComicScreenplay, updateComicScreenplayAsset, updateComicScreenplayShot, updateComicStoryboardShot } from '../../shared/api/comicScreenplay.js';
import { listAvailableModels } from '../../shared/api/modelCatalog.js';
import './comic-screenplay.css';

const ratioOptions = ['9:16', '16:9', '1:1', '4:3'].map(value => ({ value, label: value }));
const degreeOptions = [
  { value: 'master', label: '大师模式 · 3D' },
  { value: 'master_real', label: '大师模式 · 写实' },
  { value: 'more_diff', label: '差异化' },
  { value: 'standard', label: '标准' }
];
const styleOptions = ['通用动漫', '国风3D', '写实影视', '日系漫画', '水墨国风', '赛博朋克'];
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
  const [styleOpen, setStyleOpen] = useState(false);
  const [creating, setCreating] = useState(false);
  const [selectedId, setSelectedId] = useState('');
  const [workspace, setWorkspace] = useState(null);
  const [working, setWorking] = useState('');
  const [batchData, setBatchData] = useState(null);
  const [source, setSource] = useState('');
  const [sourceFile, setSourceFile] = useState(null);
  const [models, setModels] = useState([]);
  const [prompts, setPrompts] = useState([]);
	const [shotEditor, setShotEditor] = useState(null);
	const [assetEditor, setAssetEditor] = useState(null);
  const [storyboardShotEditor, setStoryboardShotEditor] = useState(null);
  const [scope, setScope] = useState('personal');
  const [sortOrder, setSortOrder] = useState('desc');
  const [viewMode, setViewMode] = useState('grid');
  const [form, setForm] = useState({ title: '', model: '', promptId: 'b_director_screenplay', artStyle: '', aspectRatio: '9:16', degree: 'master', scriptMode: 'b' });
  const [taskWarnings, setTaskWarnings] = useState([]);
  const [storyboardEpisode, setStoryboardEpisode] = useState(1);
  const [sourceSaving, setSourceSaving] = useState(false);
  const [settingsSaving, setSettingsSaving] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [versionHistoryOpen, setVersionHistoryOpen] = useState(false);
  const [documentDraft, setDocumentDraft] = useState('');
  const abortRef = useRef(null);

  async function refreshProjects() {
    setLoading(true);
    try { setProjects(await listComicScreenplays()); }
    catch (error) { message.error(error.message || '漫剧项目读取失败'); }
    finally { setLoading(false); }
  }

  useEffect(() => { void refreshProjects(); return () => abortRef.current?.abort(); }, []);
  useEffect(() => { setDocumentDraft(workspace?.document?.content || ''); }, [workspace?.document?.id]);

  async function openCreateDialog() {
    setCreateOpen(true);
    const [modelRows, promptRows] = await Promise.all([
      listAvailableModels('text').catch(() => []),
      listComicScreenplayPrompts().catch(() => []),
    ]);
    setModels(modelRows);
    setPrompts(promptRows);
    setForm(current => ({ ...current, model: current.model || modelRows[0]?.id || '', promptId: current.promptId || promptRows[0]?.id || 'b_director_screenplay' }));
  }

  async function openProject(id) {
    const switching = id !== selectedId;
    setSelectedId(id); setLoading(true);
    try {
      const data = await loadComicScreenplayWorkspace(id); setWorkspace(data);
      setSource(data?.project?.source_text || '');
      if (switching) setTaskWarnings([]);
      const episodes = splitDocumentEpisodes(data?.document?.content);
      setStoryboardEpisode(current => (episodes.some(item => item.no === current) ? current : episodes[0]?.no || 1));
      setForm(current => ({ ...current, model: data?.project?.model_id || current.model, promptId: data?.project?.prompt_id || current.promptId, degree: data?.project?.asset_degree || current.degree, artStyle: data?.project?.art_style || current.artStyle, aspectRatio: data?.project?.aspect_ratio || current.aspectRatio, scriptMode: data?.project?.script_mode || current.scriptMode }));
      try { setBatchData(await getComicScreenplayBatch(id)); } catch { setBatchData(null); }
    }
    catch (error) { message.error(error.message || '工作台读取失败'); setSelectedId(''); }
    finally { setLoading(false); }
  }

  async function persistSource() {
    if ([...source].length > 600000) throw new Error('原文超过 60 万字，请按章拆分后分批提交');
    setSourceSaving(true);
    try {
      const project = await saveComicScreenplaySource(selectedId, source);
      if (project) setWorkspace(current => (current ? { ...current, project } : current));
      return project;
    } finally { setSourceSaving(false); }
  }

  async function saveSource() {
    if (!selectedId || sourceSaving) return;
    try { await persistSource(); message.success('原文已保存'); }
    catch (error) { message.error(error.message || '原文保存失败'); }
  }

  async function saveDocumentDraft() {
    if (!selectedId || working) return;
    try {
      setWorking('document-save');
      await saveComicScreenplayDocument(selectedId, documentDraft);
      await openProject(selectedId);
      message.success('剧本已保存为新版本');
    } catch (error) { message.error(error.message || '剧本保存失败'); }
    finally { setWorking(''); }
  }

  async function restoreDocumentVersion(documentId) {
    if (!selectedId || working) return;
    try {
      setWorking('document-restore');
      await restoreComicScreenplayDocument(selectedId, documentId);
      await openProject(selectedId);
      setVersionHistoryOpen(false);
      message.success('历史版本已恢复，并保存为新的当前版本');
    } catch (error) { message.error(error.message || '版本恢复失败'); }
    finally { setWorking(''); }
  }

  async function saveSettings() {
    if (!selectedId || settingsSaving) return;
    const project = workspace?.project || {};
    try {
      setSettingsSaving(true);
      const updated = await updateComicScreenplay(selectedId, {
        title: project.title || '',
        scriptMode: form.scriptMode,
        aspectRatio: form.aspectRatio,
        artStyle: form.artStyle,
        styleDesc: project.style_desc || '',
        selectedStyle: project.selected_style || form.artStyle,
        synopsis: project.synopsis || '',
        modelId: form.model,
        promptId: form.promptId,
        assetDegree: form.degree,
        expectedRevision: project.revision || 0,
      });
      if (updated) setWorkspace(current => (current ? { ...current, project: updated } : current));
      await refreshProjects();
      message.success('设置已保存');
    } catch (error) { message.error(error.message || '设置保存失败'); }
    finally { setSettingsSaving(false); }
  }

  async function submitCreate() {
    try {
      setCreating(true);
      if (!sourceFile) throw new Error('请上传剧本文件');
      if (!form.model) throw new Error('请选择可用的文本模型');
      if (!form.artStyle) throw new Error('请选择画风');
      const selectedPrompt = prompts.find(item => item.id === form.promptId);
      const project = await createComicScreenplay({ title: form.title, scriptMode: selectedPrompt?.mode || form.scriptMode, aspectRatio: form.aspectRatio, artStyle: form.artStyle, selectedStyle: form.artStyle, modelId: form.model, promptId: form.promptId, assetDegree: form.degree, permissionMode: 'bypass' });
      let parsedSource = source;
      if (sourceFile) parsedSource = await parseComicScreenplaySource(project.id, sourceFile);
      if (parsedSource) await saveComicScreenplaySource(project.id, parsedSource);
      setCreateOpen(false); await refreshProjects(); await openProject(project.id);
    } catch (error) { message.error(error.message || '创建失败'); }
    finally { setCreating(false); }
  }

  async function runTask(key, task) {
    if (working) return;
    const controller = new AbortController(); abortRef.current = controller; setWorking(key); setTaskWarnings([]);
    try {
      const result = await task(controller.signal);
      setWorkspace(await loadComicScreenplayWorkspace(selectedId));
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
      if (error?.name === 'AbortError') message.info('已取消，本次生成的内容不会保存');
      else message.error(error.message || '处理失败');
    }
    finally { if (abortRef.current === controller) abortRef.current = null; setWorking(''); }
  }

  function startStoryboard() {
    const episodes = splitDocumentEpisodes(workspace?.document?.content);
    const fallback = String(workspace?.document?.content || source || '').trim();
    if (!episodes.length && !fallback) { message.error('请先生成或输入剧本'); return; }
    const project = workspace?.project || {};
    const payload = { engine: 'xiaoluo_v7_15s', style: project.art_style || '', synopsis: project.synopsis || '', model: form.model };
    if (!episodes.length) {
      payload.episodeNo = 1; payload.content = fallback;
    } else if (storyboardEpisode === 0) {
      payload.episodes = Object.fromEntries(episodes.map(item => [item.no, item.content]));
    } else {
      const selected = episodes.find(item => item.no === storyboardEpisode) || episodes[0];
      payload.episodeNo = selected.no; payload.content = selected.content;
    }
    return runTask('storyboard', signal => generateComicStoryboard(selectedId, { ...payload, signal }));
  }

  async function startBatch() {
    if (!selectedId || working) return;
    const controller = new AbortController();
    try {
      abortRef.current = controller;
      setWorking('batch');
      let data = await startComicScreenplayBatch(selectedId);
      setBatchData(data);
      message.info('生产任务已启动，正在执行整条链路');
      for (let attempt = 0; attempt < 64; attempt += 1) {
        const status = data?.batch?.status;
        if (['completed', 'failed', 'paused', 'waiting_user', 'awaiting_user'].includes(status)) break;
        data = await runComicScreenplayBatchNext(selectedId, controller.signal);
        setBatchData(current => ({ ...current, ...data }));
      }
      const latest = await getComicScreenplayBatch(selectedId);
      setBatchData(latest);
      setWorkspace(await loadComicScreenplayWorkspace(selectedId));
      if (latest?.batch?.status === 'completed') message.success('漫剧生产链路已全部完成');
      else if (latest?.batch?.status === 'waiting_user' || latest?.batch?.status === 'awaiting_user') message.warning('任务等待人工确认，请处理审核问题后继续');
      else if (latest?.batch?.status === 'failed') message.error(latest?.batch?.last_error || '生产链路执行失败，已保留完成的产物');
      else if (latest?.batch?.status === 'paused') message.info('生产任务已暂停，已完成的产物已保存');
    }
    catch (error) {
      try { setBatchData(await getComicScreenplayBatch(selectedId)); } catch { /* 保留当前状态 */ }
      if (error?.name === 'AbortError') message.info('已暂停，完成的产物已经保存');
      else message.error(error.message || '生产任务执行失败');
    }
    finally { if (abortRef.current === controller) abortRef.current = null; setWorking(''); }
  }

  async function pauseBatch() {
    try { abortRef.current?.abort(); setWorking('batch-cancel'); const data = await cancelComicScreenplayBatch(selectedId); setBatchData(current => ({ ...current, ...data })); message.success('生产任务已暂停'); }
    catch (error) { message.error(error.message || '暂停失败'); }
    finally { setWorking(''); }
  }

  async function downloadExport() {
    try {
      setWorking('export');
      const data = await exportComicScreenplay(selectedId);
      const url = URL.createObjectURL(new Blob([JSON.stringify(data, null, 2)], { type: 'application/json;charset=utf-8' }));
      const link = document.createElement('a'); link.href = url; link.download = `${workspace?.project?.title || '漫剧项目'}.json`; link.click(); URL.revokeObjectURL(url);
      message.success('项目包已导出');
    } catch (error) { message.error(error.message || '导出失败'); }
    finally { setWorking(''); }
  }

  async function saveShot() {
    if (!shotEditor) return;
    try {
      setWorking('shot-save');
      const input = { episodeNo: Number(shotEditor.episodeNo) || 1, heading: shotEditor.heading, scriptText: shotEditor.scriptText, characters: String(shotEditor.characters || '').split(/[、,，]/).map(item => item.trim()).filter(Boolean), scene: shotEditor.scene, props: String(shotEditor.props || '').split(/[、,，]/).map(item => item.trim()).filter(Boolean) };
      if (shotEditor.id) await updateComicScreenplayShot(selectedId, shotEditor.id, input);
      else await createComicScreenplayShot(selectedId, input);
      setShotEditor(null);
      await openProject(selectedId);
      message.success(shotEditor.id ? '场次已保存' : '场次已新增');
    } catch (error) { message.error(error.message || '场次保存失败'); }
    finally { setWorking(''); }
  }

  async function moveShot(shotId, direction) {
    try {
      setWorking(`shot-${direction}`);
      await moveComicScreenplayShot(selectedId, shotId, direction);
      await openProject(selectedId);
      message.success(direction === 'up' ? '已上移' : '已下移');
    } catch (error) { message.error(error.message || '移动失败'); }
    finally { setWorking(''); }
  }

  async function removeShot(shotId) {
    try {
      setWorking('shot-delete');
      await deleteComicScreenplayShot(selectedId, shotId);
      await openProject(selectedId);
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
    if (!assetEditor) return;
    try {
      setWorking('asset-save');
      const input = { type: assetEditor.type, degree: assetEditor.degree || form.degree, name: assetEditor.name, identity: assetEditor.identity, description: assetEditor.description, prompt: assetEditor.prompt, voice: assetEditor.voice, episodes: assetEditor.episodes, tag: assetEditor.tag, expectedRevision: assetEditor.revision || 0 };
      if (assetEditor.id) await updateComicScreenplayAsset(selectedId, assetEditor.id, input);
      else await createComicScreenplayAsset(selectedId, input);
      setAssetEditor(null);
      await openProject(selectedId);
      message.success(assetEditor.id ? '资产已保存' : '资产已新增');
    } catch (error) { message.error(error.message || '资产保存失败'); }
    finally { setWorking(''); }
  }

  async function removeAsset(assetId) {
    try {
      setWorking('asset-delete');
      await deleteComicScreenplayAsset(selectedId, assetId);
      await openProject(selectedId);
      message.success('资产已删除');
    } catch (error) { message.error(error.message || '资产删除失败'); }
    finally { setWorking(''); }
  }

  const visibleProjects = useMemo(() => {
    if (scope === 'team') return [];
    return projects
      .filter(item => item.title?.toLowerCase().includes(query.trim().toLowerCase()))
      .sort((a, b) => (sortOrder === 'desc' ? -1 : 1) * (new Date(a.updated_at) - new Date(b.updated_at)));
  }, [projects, query, scope, sortOrder]);

  const documentEpisodes = useMemo(() => splitDocumentEpisodes(workspace?.document?.content), [workspace?.document?.content]);

  if (selectedId) {
    const project = workspace?.project || projects.find(item => item.id === selectedId) || {};
    const assets = workspace?.assets || [];
    return <div className="comic-screenplay-page comic-screenplay-workbench">
      <header className="comic-screenplay-toolbar">
        <Button type="text" icon={<ArrowLeft size={16} />} onClick={() => { abortRef.current?.abort(); setSelectedId(''); setWorkspace(null); }}>返回项目库</Button>
        <div><strong>{project.title || '漫剧创作'}</strong><span>{project.aspect_ratio || '9:16'} · {project.art_style || '通用动漫'}</span></div>
        <div className="comic-screenplay-toolbar-actions">
          {working ? <Button danger onClick={() => abortRef.current?.abort()}>取消任务</Button> : null}
          <Button onClick={() => openProject(selectedId)}>刷新</Button>
        </div>
      </header>
      <main className="comic-screenplay-main">
        <aside className="comic-screenplay-steps">
          <div className="step-card is-active"><FileText size={18} /><div><b>1. 剧本</b><span>上传原文并生成分集剧本</span></div></div>
          <div className="step-card"><Users size={18} /><div><b>2. 资产</b><span>角色、场景、道具</span></div></div>
          <div className="step-card"><Clapperboard size={18} /><div><b>3. 分镜</b><span>逐集生成镜头组</span></div></div>
          <div className="step-card"><WandSparkles size={18} /><div><b>4. 生产</b><span>图片、视频、旁白与导出</span></div></div>
        </aside>
        <section className="comic-screenplay-content">
          {taskWarnings.length ? <Alert type="warning" showIcon closable onClose={() => setTaskWarnings([])} message={`本次处理有 ${taskWarnings.length} 条警告`} description={<ul style={{ margin: 0, paddingLeft: 18 }}>{taskWarnings.map((item, index) => <li key={index}>{item}</li>)}</ul>} style={{ marginBottom: 12 }} /> : null}
          <Tabs items={[
            { key: 'source', label: '原文与剧本', children: <div className="screenplay-editor-grid"><div className="editor-pane"><div className="pane-title">原文</div><Input.TextArea value={source} onChange={event => setSource(event.target.value)} onBlur={() => { if (selectedId && source) void persistSource().catch(error => message.error(error.message || '原文保存失败')); }} placeholder="粘贴小说正文，或上传 TXT / DOCX 文件" autoSize={{ minRows: 18 }} /><div className="asset-actions"><Upload accept=".txt,.docx" beforeUpload={file => { void parseComicScreenplaySource(selectedId, file).then(async content => { setSource(content); await saveComicScreenplaySource(selectedId, content); message.success('原文已读取并保存'); }).catch(error => message.error(error.message || '文件读取失败')); return false; }} showUploadList={false}><Button icon={<UploadCloud size={16} />}>上传 TXT / DOCX</Button></Upload><Button loading={sourceSaving} onClick={saveSource}>保存原文</Button><Button type="primary" icon={<Sparkles size={16} />} loading={working === 'script'} onClick={() => runTask('script', async signal => { await persistSource(); return generateComicScreenplay(selectedId, { mode: form.scriptMode, content: source, model: form.model, promptId: form.promptId, signal }); })}>生成分集剧本</Button></div></div><div className="editor-pane"><div className="pane-title"><span>当前剧本 · v{workspace?.document?.version || 0}</span><Button size="small" icon={<History size={14} />} onClick={() => setVersionHistoryOpen(true)}>版本历史</Button></div><Input.TextArea className="screenplay-output screenplay-document-editor" value={documentDraft} onChange={event => setDocumentDraft(event.target.value)} placeholder="生成后可在这里继续修改剧本" autoSize={{ minRows: 18 }} /><div className="asset-actions"><Button type="primary" loading={working === 'document-save'} disabled={!workspace?.document} onClick={saveDocumentDraft}>保存新版本</Button></div></div></div> },
            { key: 'assets', label: `资产库 ${assets.length}`, children: <div><div className="asset-actions"><Select value={form.degree} options={degreeOptions} onChange={degree => setForm(current => ({ ...current, degree }))} /><Button icon={<Plus size={16} />} onClick={() => setAssetEditor({ type: 'character', degree: form.degree, name: '', identity: '', description: '', prompt: '', voice: '', episodes: '', tag: '' })}>新增资产</Button><Button icon={<Users size={16} />} onClick={() => runTask('character', signal => extractComicAssets(selectedId, { type: 'character', degree: form.degree, content: workspace?.document?.content || source, model: form.model, signal }))}>提取角色</Button><Button icon={<Map size={16} />} onClick={() => runTask('scene', signal => extractComicAssets(selectedId, { type: 'scene', degree: form.degree, content: workspace?.document?.content || source, model: form.model, signal }))}>提取场景</Button><Button icon={<Package size={16} />} onClick={() => runTask('prop', signal => extractComicAssets(selectedId, { type: 'prop', degree: form.degree, content: workspace?.document?.content || source, model: form.model, signal }))}>提取道具</Button></div><div className="asset-grid">{assets.length ? assets.map(asset => <article key={asset.id} className="asset-card"><span>{assetTypeOptions.find(item => item.value === asset.type)?.label || asset.type}</span><b>{asset.name}</b><p>{asset.description || asset.identity || '暂无描述'}</p><div className="asset-card-actions"><Button size="small" icon={<Pencil size={14} />} onClick={() => setAssetEditor({ id: asset.id, revision: asset.revision, type: asset.type, degree: asset.degree, name: asset.name, identity: asset.identity || '', description: asset.description || '', prompt: asset.prompt || '', voice: asset.voice || '', episodes: asset.episodes || '', tag: asset.tag || '' })}>编辑</Button><Popconfirm title="删除这个资产？" description="删除后不可恢复" onConfirm={() => removeAsset(asset.id)}><Button size="small" danger icon={<Trash2 size={14} />}>删除</Button></Popconfirm></div></article>) : <Empty description="尚未提取或新增资产" />}</div></div> },
            { key: 'storyboard', label: '分镜工作台', children: <div>
              <div className="asset-actions">
                <Select style={{ minWidth: 150 }} value={storyboardEpisode} onChange={setStoryboardEpisode} options={documentEpisodes.length ? [...documentEpisodes.map(item => ({ value: item.no, label: `第 ${item.no} 集` })), { value: 0, label: '全部集' }] : [{ value: 1, label: '第 1 集（全文）' }]} />
                <Button type="primary" loading={working === 'storyboard'} onClick={startStoryboard}>生成 CM 分镜</Button>
                <Button icon={<Plus size={16} />} onClick={() => setShotEditor({ episodeNo: storyboardEpisode || 1, heading: '', scriptText: '', characters: '', scene: '', props: '' })}>新增场次</Button>
              </div>
              <p style={{ margin: '0 0 12px', color: '#8a8f99' }}>{documentEpisodes.length ? '按集生成：只提交选中集的内容；选「全部集」会逐集提交。' : '未检测到「# 第N集」标题，将按第 1 集提交全文。'}</p>
              <div className="shot-editor-list">{(workspace?.shots || []).length ? workspace.shots.map((shot, index) => <article key={shot.id} className="shot-editor-card">
                <div><span>第 {shot.episode_no} 集 · 场次 {shot.sort_order}</span><b>{shot.heading || '未命名场次'}</b><p>{shot.script_text || '暂无内容'}</p></div>
                <div className="shot-editor-actions">
                  <Button aria-label="上移场次" icon={<ArrowUp size={15} />} disabled={working || index === 0 || workspace.shots[index - 1]?.episode_no !== shot.episode_no} onClick={() => moveShot(shot.id, 'up')} />
                  <Button aria-label="下移场次" icon={<ArrowDown size={15} />} disabled={working || index === workspace.shots.length - 1 || workspace.shots[index + 1]?.episode_no !== shot.episode_no} onClick={() => moveShot(shot.id, 'down')} />
                  <Button aria-label="编辑场次" icon={<Pencil size={15} />} onClick={() => setShotEditor({ id: shot.id, episodeNo: shot.episode_no, heading: shot.heading, scriptText: shot.script_text, characters: (shot.characters || []).join('、'), scene: shot.scene || '', props: (shot.props || []).join('、') })} />
                  <Popconfirm title="删除这个场次？" description="删除后不可恢复" onConfirm={() => removeShot(shot.id)}><Button danger aria-label="删除场次" icon={<Trash2 size={15} />} /></Popconfirm>
                </div>
              </article>) : <Empty description="尚未生成或新增场次" />}</div>
              <div className="storyboard-list">{(workspace?.storyboards || []).length ? (workspace.storyboards || []).map(item => <section className="storyboard-card" key={item.id}>
                <header><div><b>第 {item.episode_no} 集分镜</b><span>{item.shot_count} 个镜头组 · 版本 {item.revision} · {item.status}</span></div><Button size="small" icon={<Plus size={14} />} disabled={Boolean(working)} onClick={() => setStoryboardShotEditor({ engine: item.engine, episodeNo: item.episode_no, afterShotId: item.shots?.at(-1)?.id || '', durationSec: 3, sceneId: 0, characters: '', scene: '', props: '', videoDesc: '' })}>末尾新增</Button></header>
                <div className="storyboard-shot-list">{(item.shots || []).length ? item.shots.map((shot, index) => <article className="storyboard-shot-card" key={shot.id}>
                  <div className="storyboard-shot-copy"><span>镜头 {shot.seq} · {Number(shot.duration_sec || 0).toFixed(1)} 秒</span><b>{shot.scene || '未命名场景'}</b><p>{shot.video_desc || '暂无画面描述'}</p><small>角色：{shot.characters || '无'} · 道具：{shot.props || '无'}</small></div>
                  <div className="storyboard-shot-actions">
                    <Button size="small" onClick={() => setStoryboardShotEditor({ id: shot.id, engine: item.engine, episodeNo: item.episode_no, durationSec: shot.duration_sec, sceneId: shot.scene_id, characters: shot.characters || '', scene: shot.scene || '', props: shot.props || '', videoDesc: shot.video_desc || '' })}>编辑</Button>
                    <Button size="small" disabled={Boolean(working)} onClick={() => setStoryboardShotEditor({ engine: item.engine, episodeNo: item.episode_no, afterShotId: shot.id, durationSec: 3, sceneId: shot.scene_id, characters: '', scene: shot.scene || '', props: '', videoDesc: '' })}>在后面新增</Button>
                    <Button size="small" disabled={Boolean(working) || Number(shot.duration_sec) <= 0.4} onClick={() => runStoryboardShotAction('storyboard-shot-split', () => splitComicStoryboardShot(selectedId, item.engine, item.episode_no, shot.id), '分镜已等分拆开')}>拆分</Button>
                    <Button size="small" disabled={Boolean(working) || index === item.shots.length - 1} onClick={() => runStoryboardShotAction('storyboard-shot-merge', () => mergeComicStoryboardShot(selectedId, item.engine, item.episode_no, shot.id), '已与下一条分镜合并')}>与下一条合并</Button>
                    <Popconfirm title="删除这条分镜？" description="删除后会重新排列后续镜头序号" onConfirm={() => runStoryboardShotAction('storyboard-shot-delete', () => deleteComicStoryboardShot(selectedId, item.engine, item.episode_no, shot.id), '分镜已删除')}><Button size="small" danger disabled={Boolean(working)}>删除</Button></Popconfirm>
                  </div>
                </article>) : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="这份分镜还没有镜头，可从末尾新增" />}</div>
              </section>) : <Empty description="尚未生成分镜" />}</div>
            </div> },
            { key: 'production', label: '生产与导出', children: <div className="production-panel">
              <div className="asset-actions"><Button type="primary" icon={<Play size={16} />} loading={working === 'batch'} onClick={startBatch}>{batchData?.batch?.status === 'paused' ? '继续任务' : '启动生产任务'}</Button><Button danger icon={<PauseCircle size={16} />} disabled={!batchData?.batch || batchData?.batch?.status === 'paused'} loading={working === 'batch-cancel'} onClick={pauseBatch}>暂停任务</Button><Button icon={<Download size={16} />} loading={working === 'export'} onClick={downloadExport}>导出项目包</Button></div>
              <div className="production-status">
                <h3>任务状态：{batchData?.batch ? batchStatusLabel(batchData.batch.status) : '尚未启动'}</h3>
                {batchData?.batch?.last_error ? <Alert type="error" showIcon message="最近一次失败" description={batchData.batch.last_error} /> : null}
                <div className="production-steps">{(batchData?.steps || []).map(step => {
                  const state = batchData?.batch?.steps?.[step.key];
                  return <article key={step.key} className={`status-${state?.status || 'pending'}`}><span>{step.label}</span><b>{batchStatusLabel(state?.status)}</b>{state?.total_episodes > 0 ? <small>{Object.values(state.episodes || {}).filter(item => item.status === 'done').length}/{state.total_episodes} 集</small> : null}{state?.error ? <em title={state.error}>{state.error}</em> : null}</article>;
                })}</div>
                <p>系统按“原文 → 剧本 → 审核 → 资产 → 审核 → 分镜 → 审核 → 视频 → 完成”逐步执行；暂停、刷新和继续都会读取已保存进度。</p>
              </div>
              {Object.values(batchData?.batch?.steps || {}).some(step => step?.artifacts?.length) ? <section className="production-artifacts"><h4>已保存产物</h4>{Object.values(batchData.batch.steps).flatMap(step => step?.artifacts || []).map((artifact, index) => { const view = artifactView(artifact); return view.href ? <a key={`${artifact}-${index}`} href={view.href} target="_blank" rel="noreferrer">{view.label}</a> : <span key={`${artifact}-${index}`}>{view.label}</span>; })}</section> : null}
              {batchData?.logs?.length ? <section className="production-logs"><h4>执行记录</h4>{batchData.logs.slice().reverse().map((log, index) => <div key={log.id || index} className={`log-${log.level || 'info'}`}><time>{prettyTime(log.created_at)}</time><span>{log.message}</span></div>)}</section> : null}
            </div> }
          ]} />
        </section>
        <aside className="comic-screenplay-inspector"><h3>项目设置</h3><label>文本模型<Input value={form.model} placeholder="使用账号默认模型" onChange={event => setForm(current => ({ ...current, model: event.target.value }))} /></label><label>画面比例<Select value={form.aspectRatio} options={ratioOptions} onChange={aspectRatio => setForm(current => ({ ...current, aspectRatio }))} /></label><label>创作模式<Select value={form.scriptMode} options={[{ value: 'a1', label: '小说转剧本' }, { value: 'b', label: '导演级剧本' }]} onChange={scriptMode => setForm(current => ({ ...current, scriptMode }))} /></label><Button type="primary" block loading={settingsSaving} onClick={saveSettings}>保存设置</Button><div className="workspace-stats"><span>版本 <b>{workspace?.documents?.length || 0}</b></span><span>场次 <b>{workspace?.shots?.length || 0}</b></span><span>角色 <b>{workspace?.characters?.length || 0}</b></span><span>场景 <b>{workspace?.scenes?.length || 0}</b></span></div></aside>
      </main>
      <Modal title={assetEditor?.id ? '编辑资产' : '新增资产'} open={Boolean(assetEditor)} onCancel={() => setAssetEditor(null)} confirmLoading={working === 'asset-save'} onOk={saveAsset} okText="保存" cancelText="取消" destroyOnHidden>
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
      <Modal title={storyboardShotEditor?.id ? '编辑分镜' : '新增分镜'} open={Boolean(storyboardShotEditor)} onCancel={() => setStoryboardShotEditor(null)} confirmLoading={working === 'storyboard-shot-save'} onOk={saveStoryboardShot} okText="保存" cancelText="取消" destroyOnHidden>
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
            {document.status === 'active' ? <span>正在使用</span> : <Popconfirm title={`恢复 v${document.version}？`} description="系统会复制成一个新版本，现有版本不会被覆盖。" onConfirm={() => restoreDocumentVersion(document.id)}><Button size="small" loading={working === 'document-restore'}>恢复此版本</Button></Popconfirm>}
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
    <div className="library-tabs"><button className={scope === 'personal' ? 'is-active' : ''} onClick={() => setScope('personal')}>个人作品</button><button title="团队协作的数据权限尚未接入，当前不会展示假项目" aria-disabled="true" onClick={() => message.info('团队协作尚未接入成员和权限数据，暂不展示假项目')}>团队作品</button></div>
    <div className="comic-screenplay-filters">
      <Input prefix={<Search size={16} />} value={query} onChange={event => setQuery(event.target.value)} placeholder="搜索作品..." allowClear />
      <div className="library-filter-actions"><Select value={sortOrder} suffixIcon={<ChevronDown size={15} />} onChange={setSortOrder} options={[{ value: 'desc', label: '按时间降序' }, { value: 'asc', label: '按时间升序' }]} /><Button type="text" icon={viewMode === 'grid' ? <Grid2X2 size={18} /> : <LayoutGrid size={18} />} onClick={() => setViewMode(value => value === 'grid' ? 'list' : 'grid')} /><Button type="text" icon={<Sparkles size={16} />} disabled title="新版工作台尚未开放">新版</Button><span className="version-active"><History size={16} />旧版</span></div>
    </div>
    <Spin spinning={loading}><div className={`comic-project-grid ${viewMode === 'list' ? 'is-list' : ''}`}>{visibleProjects.length ? visibleProjects.map(project => <article key={project.id} className="comic-project-card" onClick={() => openProject(project.id)}><div className="comic-project-cover"><Clapperboard size={38} /><span>{project.aspect_ratio || '9:16'}</span></div><div className="comic-project-meta"><div><h3>{project.title}</h3><p>{project.art_style || '未选择风格'} · {prettyTime(project.updated_at)}</p></div><Popconfirm title="删除这个漫剧项目？" onConfirm={async event => { event?.stopPropagation?.(); await deleteComicScreenplay(project.id); await refreshProjects(); }}><Button type="text" danger icon={<Trash2 size={16} />} onClick={event => event.stopPropagation()} /></Popconfirm></div></article>) : <Empty description={scope === 'team' ? '暂无团队作品' : '暂无项目，在上方输入框输入内容开始创作'} />}</div></Spin>
    <Modal title="最近创作历史" open={historyOpen} onCancel={() => setHistoryOpen(false)} footer={<Button onClick={() => setHistoryOpen(false)}>关闭</Button>}>
      <div className="screenplay-version-list">{projects.length ? [...projects].sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at)).map(project => <button type="button" key={project.id} onClick={() => { setHistoryOpen(false); void openProject(project.id); }}><b>{project.title}</b><span>{prettyTime(project.updated_at)}</span></button>) : <Empty description="还没有创作记录" />}</div>
    </Modal>
    <Modal className="screenplay-create-modal" width={650} title="新建剧本" open={createOpen} closeIcon={<X size={20} />} onCancel={() => setCreateOpen(false)} footer={<><Button onClick={() => setCreateOpen(false)}>取消</Button><Button type="primary" icon={<FileText size={16} />} loading={creating} onClick={submitCreate}>创建</Button></>}>
      <div className="comic-create-form">
        <label><span>作品名称 <i>*</i></span><Input autoFocus value={form.title} maxLength={80} placeholder="请输入剧本名称" onChange={event => setForm(current => ({ ...current, title: event.target.value }))} /></label>
        <label><span>选择模型</span><Select value={form.model || undefined} placeholder="请选择已配置的文本模型" onChange={model => setForm(current => ({ ...current, model }))} options={models.map(model => ({ value: model.id, label: model.displayName || model.name || model.id }))} /></label>
        <label><span>选择提示词</span><div className="prompt-select-row"><Select value={form.promptId} onChange={promptId => { const prompt = prompts.find(item => item.id === promptId); setForm(current => ({ ...current, promptId, scriptMode: prompt?.mode || current.scriptMode })); }} options={prompts.map(item => ({ value: item.id, label: item.name }))} /><p>建议小说原文生成前往 Agent / Chat 功能，不清楚？<a href="/agent/chat">查看链接</a></p></div></label>
        <div className="comic-create-row"><label><span>选择风格</span><button type="button" className="style-picker-trigger" onClick={() => setStyleOpen(true)}><Palette size={17} />{form.artStyle || '点击选择画风'}</button></label><label><span>剧本比例</span><Select value={form.aspectRatio} options={[{ value: '9:16', label: '9:16（竖屏）' }, { value: '16:9', label: '16:9（横屏）' }, { value: '1:1', label: '1:1（方形）' }, { value: '4:3', label: '4:3' }]} onChange={aspectRatio => setForm(current => ({ ...current, aspectRatio }))} /></label></div>
        <label className="degree-field"><span>资产解析程度</span><div className="degree-segments">{degreeOptions.map(option => <button type="button" className={form.degree === option.value ? 'is-active' : ''} key={option.value} onClick={() => setForm(current => ({ ...current, degree: option.value }))}>{option.label.replace(' · ', '-')}</button>)}</div><small>大师两档解析人物 / 场景 / 道具，并带「颜值定位」「出现集数」字段；差异化与标准档仅作用于人物。</small></label>
        <label className="upload-field"><span>上传剧本格式 <i>*</i> <HelpCircle size={13} /></span><Upload.Dragger accept=".txt,.docx" maxCount={1} beforeUpload={file => { setSourceFile(file); return false; }} onRemove={() => setSourceFile(null)}><FileText size={30} /><p>{sourceFile?.name || '点击或拖动文件 (txt / docx)'}</p></Upload.Dragger></label>
      </div>
    </Modal>
    <Modal className="screenplay-style-modal" title="选择画风" open={styleOpen} onCancel={() => setStyleOpen(false)} footer={null} width={620}><div className="style-grid">{styleOptions.map((style, index) => <button type="button" key={style} className={form.artStyle === style ? 'is-active' : ''} onClick={() => { setForm(current => ({ ...current, artStyle: style })); setStyleOpen(false); }}><span className={`style-preview style-${index + 1}`}>{form.artStyle === style ? <Check size={20} /> : null}</span><b>{style}</b></button>)}</div></Modal>
  </div>;
}
