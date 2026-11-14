import { Alert, Button, Empty, Input, Modal, Space, Spin, Tag, message } from 'antd';
import { useCallback, useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { getComicMedia } from '../../shared/api/comicScreenplayWorkflow.js';
import { actComicAssetMedia } from '../../shared/api/comicScreenplayAssetActions.js';
import { safeComicMediaUrl } from './comic-screenplay-workflow.js';
import { comicMediaServerRequestId, validateVoiceDesign, voiceDesignAudio, voiceDraftTarget } from './comic-screenplay-asset-editor.js';
import OriginalIcon from './ComicScreenplayIcon.jsx';
import './comic-screenplay-asset-dialog.css';

export default function ComicScreenplayVoiceDesignStage({ projectId, workspace, readOnly = false, canGenerate = false, toolbarHost }) {
  const [data, setData] = useState(null), [loading, setLoading] = useState(true), [error, setError] = useState(''), [busy, setBusy] = useState(''), [progress, setProgress] = useState(''), [editor, setEditor] = useState(null), [editError, setEditError] = useState(''), [editConflict,setEditConflict]=useState(false);
  const alive = useRef(true), current = useRef(null), controller = useRef(null), busyRef = useRef(false), batchStop = useRef(false), keys = useRef(new Map()), ticket = useRef(0), editorBaseline=useRef(null);
  const refresh = useCallback(async () => {
    const sequence = ++ticket.current;
    const result = await getComicMedia(projectId);
    if (alive.current && sequence === ticket.current) { current.current = result; setData(result); }
    return result;
  }, [projectId]);
  useEffect(() => { alive.current = true; setLoading(true); void refresh().catch(e => { if (alive.current) setError(e.message || '音色读取失败'); }).finally(() => { if (alive.current) setLoading(false); }); const timer = setInterval(() => { if (!document.hidden && !busyRef.current) void refresh().catch(() => {}); }, 10000); return () => { alive.current = false; ticket.current++; batchStop.current = true; controller.current?.abort(); clearInterval(timer); }; }, [refresh]);
  const rows = data?.settings?.voiceDesign?.items || [], items = data?.items || [], caps = data?.capabilities || {};
  const engine = workspace?.project?.creative_config?.voice_design_model || 'minimax';
  const engineName = engine === 'qwen-tts' ? 'Qwen' : engine === 'minimax' ? 'MiniMax' : engine;
  const textModel = workspace?.project?.model_id || '';
  async function perform(label, callback) {
    if (busyRef.current) return;
    busyRef.current = true; batchStop.current = false; setBusy(label); setError('');
    const abort = new AbortController(); controller.current = abort;
    try { return await callback(abort.signal); }
    catch (e) { if (alive.current) { const text = abort.signal.aborted ? '已取消当前请求。请刷新核对已保存的音色结果。' : e.message || '音色操作失败'; setError(text); setEditError(e.status===409?'音色描述已有新版本，本地修改仍保留。请重新载入后核对。':text); setEditConflict(e.status===409); } }
    finally { try { await refresh(); } catch {} if (alive.current) { busyRef.current = false; setBusy(''); setProgress(''); } }
  }
  async function saveRows(next, expectedRevision, signal) {
    const result = await actComicAssetMedia(projectId, { action: 'voice_design_drafts', items: next, expectedRevision }, { signal });
    if (!result || !Array.isArray(result.items)) throw new Error('音色描述保存未返回有效结果');
    await refresh(); return result;
  }
  async function analyze(signal) {
    if (readOnly || !canGenerate || !caps.voice_analyze?.available || !textModel) return;
    const fingerprint = `analyze:${textModel}`; let key = keys.current.get(fingerprint); if (!key) { key = crypto.randomUUID(); keys.current.set(fingerprint, key); }
    const serverId = await comicMediaServerRequestId('voice_analyze', key);
    const latest=await refresh();if(latest.settings?.voiceDesign?.status==='running' && latest.settings.voiceDesign.requestKey!==serverId)throw new Error('音色分析仍在执行，请等待完成后刷新核对。');
    let result;
    try { result = await actComicAssetMedia(projectId, { action: 'voice_analyze', modelId: textModel, expectedRevision: Number(latest.settings?.revision||0), idempotencyKey: key }, { signal }); } catch (e) { const state = await refresh().catch(()=>null); if (state?.settings?.voiceDesign?.requestKey === serverId && ['failed','cancelled'].includes(state.settings.voiceDesign.status)) keys.current.delete(fingerprint); throw e; }
    if (result?.status !== 'succeeded' || !Array.isArray(result.items)) throw new Error(result?.error || '分析没有返回有效音色描述');
    if(alive.current)setData(previous=>({...previous,settings:{...previous?.settings,voiceDesign:result}}));
    try{await refresh();}catch{throw new Error('音色描述分析已成功，状态读取失败，请刷新核对。');}
    keys.current.delete(fingerprint); if (alive.current) message.success(`已识别 ${result.items.length} 个音色描述，尚未生成音频`);
  }
  async function design(row, signal) {
    const invalid = validateVoiceDesign(row, engine); if (invalid) throw new Error(`${row.name}：${invalid}`);
    const fingerprint = JSON.stringify(['design', voiceDraftTarget(row), engine, row.prompt, row.preview_text]); let key = keys.current.get(fingerprint); if (!key) { key = crypto.randomUUID(); keys.current.set(fingerprint, key); }
    const serverId = await comicMediaServerRequestId('voice_design', key, {engine,targetId:voiceDraftTarget(row)});
    const latest=await refresh();if(latest.items?.some(item=>item.kind==='voice_design'&&item.targetId===voiceDraftTarget(row)&&item.status==='running'&&item.id!==serverId))throw new Error('这个音色仍有设计请求正在执行，请等待完成后刷新核对。');
    let result;
    try { result = await actComicAssetMedia(projectId, { action: 'voice_design', targetId: voiceDraftTarget(row), prompt: row.prompt, text: row.preview_text, expectedRevision: Number(latest.settings?.revision||0), idempotencyKey: key }, { signal }); } catch (e) { const state = await refresh().catch(()=>null); const attempt = state?.items?.find(item=>item.id===serverId); if (['failed','cancelled'].includes(attempt?.status)) keys.current.delete(fingerprint); throw e; }
    if (result?.kind !== 'audio' || !result.voiceId || !result.voiceEngine) throw new Error(result?.error || '音色设计未返回真实音频和 voice_id');
    if(alive.current)setData(previous=>({...previous,items:[...(previous?.items||[]).filter(item=>item.id!==result.id),result]}));
    try{await refresh();}catch{throw new Error('音色已设计成功，状态读取失败，请刷新核对。');}
    keys.current.delete(fingerprint);return result;
  }
  function requestAnalyze() {
    Modal.confirm({ title: rows.length ? '重新分析角色音色？' : 'AI音色分析', content: rows.length ? '将读取作品原文重新分析，并替换当前描述列表。已生成的音频会保留。此操作调用文字模型。' : '读取作品原文分析人物音色，保存描述列表。此步骤调用文字模型，不设计音频。', okText: '开始分析', cancelText: '取消', onOk: () => perform('analyze', analyze) });
  }
  function requestDesign(row) {
    Modal.confirm({ title: `设计「${row.name}」音色？`, content: `使用已保存的 ${engineName} 音色设计引擎，将按描述和试听文本生成新的音色及试听音频。`, okText: '设计音色', cancelText: '取消', onOk: () => perform(`design:${row.id}`, signal => design(row, signal)) });
  }
  function requestBatch() {
    const pending = rows.filter(row => !voiceDesignAudio(items, row));
    Modal.confirm({ title: '批量设计音色', content: `使用 ${engineName} 依次设计 ${pending.length} 个尚无音频的音色。失败时停止，已完成的结果会保留。`, okText: '开始设计', cancelText: '取消', onOk: () => perform('batch', async signal => { for (let i = 0; i < pending.length; i++) { if (batchStop.current || signal.aborted || !alive.current) break; setProgress(`${pending[i].name}（${i + 1}/${pending.length}）`); await design(pending[i], signal); } }) });
  }
  async function saveEditor() {
    if (!editor?.name?.trim()) { setEditError('请输入音色名称'); return; }
    const baseline=editorBaseline.current;
    if(!baseline){setEditError('请重新打开音色编辑窗口');return;}
    return perform('save', async signal => { const previous = baseline.rows; const exists = previous.some(row => row.id === editor.id); const next = exists ? previous.map(row => row.id === editor.id ? { ...editor, name: editor.name.trim() } : row) : [...previous, { ...editor, name: editor.name.trim() }]; await saveRows(next, baseline.revision, signal); if (alive.current) { setEditor(null); message.success('音色描述已保存，尚未调用音频设计'); } });
  }
  function removeRow(row) { const baseline={rows:[...rows],revision:Number(data?.settings?.revision||0)};Modal.confirm({ title: `删除「${row.name}」描述？`, content: '只删除描述条目；已经生成的音频文件仍保留在作品媒体中。', okText: '删除描述', cancelText: '取消', onOk: () => perform('save', signal => saveRows(baseline.rows.filter(item => item.id !== row.id), baseline.revision, signal)) }); }
  function edit(row) { setEditError('');setEditConflict(false);editorBaseline.current={rows:rows.map(item=>({...item})),revision:Number(data?.settings?.revision||0)}; setEditor(row ? { ...row } : { id: Math.max(0, ...rows.map(item => Number(item.id) || 0)) + 1, name: '', role: '', prompt: '', preview_text: '' }); }
  function reloadEditor(){Modal.confirm({title:'放弃本地音色修改并重新载入？',content:'将读取服务器最新描述。',okText:'重新载入',cancelText:'保留本地修改',onOk:async()=>{const latest=await refresh();const next=latest.settings?.voiceDesign?.items||[];editorBaseline.current={rows:next.map(item=>({...item})),revision:Number(latest.settings?.revision||0)};setEditor(previous=>next.find(item=>item.id===previous?.id)?{...next.find(item=>item.id===previous.id)}:null);setEditError('');setEditConflict(false);}});}
  const ungenerated = rows.filter(row => !voiceDesignAudio(items, row));
  const toolbar = <Space size={8} wrap><Button size="small" icon={<OriginalIcon name="AddOutline" size={15} />} disabled={readOnly || !!busy || loading} onClick={() => edit(null)}>添加音色</Button><Button size="small" icon={<OriginalIcon name="SparklesOutline" size={15} />} loading={busy === 'analyze'} disabled={readOnly || !canGenerate || !!busy || !caps.voice_analyze?.available || !textModel} title={!textModel ? '请先选择并保存文字模型' : caps.voice_analyze?.reason} onClick={requestAnalyze}>AI音色分析</Button><Button size="small" icon={<OriginalIcon name="MusicalNotesOutline" size={15} />} loading={busy === 'batch'} disabled={readOnly || !canGenerate || !!busy || !caps.voice_design?.available || !ungenerated.length} title={caps.voice_design?.reason} onClick={requestBatch}>批量设计音色 {ungenerated.length || ''}</Button><Button size="small" disabled={!!busy} onClick={() => { setError(''); void refresh().catch(e => setError(e.message)); }}>刷新</Button>{busy ? <Button size="small" onClick={() => { batchStop.current = true; controller.current?.abort(); }}>取消当前请求</Button> : null}</Space>;
  return <section className="comic-voice-design-stage">{toolbarHost ? createPortal(toolbar, toolbarHost) : toolbar}<div className="comic-voice-design-summary"><span>音色设计引擎 <Tag>{engineName}</Tag><small>在模型配置中修改</small></span>{data?.settings?.voiceDesign?.source ? <small>分析来源：{data.settings.voiceDesign.source === 'authoring_chapters' ? '作品章节原文' : data.settings.voiceDesign.source === 'project_original' ? '作品原文' : '手动保存'}</small> : null}</div>
    {error ? <Alert type="error" showIcon message={error} /> : null}{data?.settings?.voiceDesign?.error ? <Alert type="warning" showIcon message={data.settings.voiceDesign.error} /> : null}{!loading && !caps.voice_analyze?.available ? <Alert type="info" showIcon message={caps.voice_analyze?.reason || '音色分析服务尚未配置'} /> : null}{!loading && !caps.voice_design?.available ? <Alert type="info" showIcon message={caps.voice_design?.reason || '音色设计服务尚未配置；可以先整理描述'} /> : null}
    {busy ? <div className="comic-voice-design-progress" role="status"><Spin size="small" />{busy === 'analyze' ? '正在读取原文并分析音色描述…' : busy === 'batch' ? `正在设计 ${progress}` : busy === 'save' ? '正在保存音色描述…' : '正在设计试听音色…'}</div> : null}
    <Spin spinning={loading}><div className="comic-voice-description-list">{rows.map(row => { const audio = voiceDesignAudio(items, row), attempt = items.filter(item => item.kind === 'voice_design' && item.targetId === voiceDraftTarget(row)).at(-1); return <article key={row.id} className="comic-voice-description-card"><header><OriginalIcon name="MusicalNotesOutline" size={20} /><div><b>{row.name}</b><small>{row.role || '未填写角色身份'}</small></div><Space><Button size="small" disabled={readOnly || !!busy} onClick={() => edit(row)}>编辑</Button><Button size="small" danger type="text" disabled={readOnly || !!busy} aria-label={`删除${row.name}音色描述`} onClick={() => removeRow(row)} icon={<OriginalIcon name="TrashOutline" size={14} />} /></Space></header><div className="comic-voice-description-copy"><label>音色描述</label><p>{row.prompt || '尚未填写音色描述'}</p><label>试听文本</label><p>{row.preview_text || '尚未填写试听文本'}</p></div><footer><Button size="small" icon={<OriginalIcon name="SparklesOutline" size={14} />} loading={busy === `design:${row.id}`} disabled={readOnly || !canGenerate || !!busy || !caps.voice_design?.available} title={validateVoiceDesign(row, engine) || caps.voice_design?.reason} onClick={() => requestDesign(row)}>{audio ? '重新设计' : '设计音色'}</Button>{audio ? <Tag color="cyan">已生成 · {audio.voiceEngine === 'qwen-tts' ? 'Qwen' : audio.voiceEngine === 'minimax' ? 'MiniMax' : audio.voiceEngine}</Tag> : <small>仅描述 · 尚未生成音频</small>}</footer>{audio ? <div className="comic-voice-designed-audio">{safeComicMediaUrl(audio.url) ? <audio controls preload="none" src={safeComicMediaUrl(audio.url)} aria-label={`${row.name}音色试听`} /> : <small>音频试听地址暂不可用，请刷新</small>}<span>voice_id <code>{audio.voiceId}</code></span>{audio.voiceTargetModel ? <small>模型：{audio.voiceTargetModel}</small> : null}<details><summary>本条音频生成时的描述与试听文本</summary><p>{audio.voiceDescription}</p><p>{audio.voicePreviewText}</p></details></div> : null}{attempt?.status === 'failed' || attempt?.status === 'cancelled' ? <Alert type="warning" showIcon message={attempt.error || (attempt.status === 'cancelled' ? '上次设计已取消' : '上次设计失败')} /> : null}</article>; })}</div>{!rows.length && !loading ? <Empty image={<OriginalIcon name="MusicalNotesOutline" size={48} style={{ opacity: .4 }} />} description={<><b>暂无AI生成音色</b><br /><small>点击「AI音色分析」识别角色描述，再单独设计试听音频。</small></>} /> : null}</Spin>
    <Modal open={!!editor} width={460} className="comic-screenplay-asset-dialog" title={rows.some(row => row.id === editor?.id) ? '编辑音色' : '添加音色'} onCancel={() => { if (!busy) setEditor(null); }} maskClosable={!busy} keyboard={!busy} onOk={saveEditor} okText="保存" cancelText="取消" confirmLoading={busy === 'save'} okButtonProps={{ disabled: editConflict || (!!busy && busy !== 'save') }} cancelButtonProps={{ disabled: !!busy }} destroyOnHidden>{editError ? <Alert type="error" message={editError} action={editConflict?<Button size="small" onClick={reloadEditor}>重新载入描述</Button>:undefined}/> : null}<div className="comic-asset-dialog-fields">{[['name', '音色名称', '如：旁白、角色名'], ['role', '角色身份', '如：旁白/叙述、主角'], ['prompt', '音色描述', '描述音色特征'], ['preview_text', '试听文本', '用于生成试听音频的文本']].map(([field, label, placeholder]) => <label key={field}>{label}{['name', 'role'].includes(field) ? <Input disabled={!!busy} value={editor?.[field] || ''} placeholder={placeholder} onChange={e => setEditor(previous => ({ ...previous, [field]: e.target.value }))} /> : <Input.TextArea rows={field === 'prompt' ? 3 : 2} disabled={!!busy} value={editor?.[field] || ''} placeholder={placeholder} onChange={e => setEditor(previous => ({ ...previous, [field]: e.target.value }))} />}</label>)}</div></Modal>
  </section>;
}
