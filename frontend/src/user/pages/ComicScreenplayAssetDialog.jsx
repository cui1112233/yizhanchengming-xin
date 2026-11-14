import {useComicConfirmation} from './comic-screenplay-confirm.js';
import { Alert, Button, Empty, Image, Input, Modal, Select, Space, Spin, Upload } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { createComicScreenplayAsset, updateComicScreenplayAsset } from '../../shared/api/comicScreenplay.js';
import { getComicMedia, uploadComicMedia } from '../../shared/api/comicScreenplayWorkflow.js';
import { actComicAssetMedia } from '../../shared/api/comicScreenplayAssetActions.js';
import { listAvailableModels } from '../../shared/api/modelCatalog.js';
import { comicAssetMediaImage } from './comic-screenplay-stages.js';
import { comicImageModelSelection } from './comic-screenplay-creative.js';
import { assetEditorDraft, assetEditorPayload, comicMediaServerRequestId } from './comic-screenplay-asset-editor.js';
import { safeComicMediaUrl } from './comic-screenplay-workflow.js';
import AssetImage from './ComicScreenplayImage.jsx';
import OriginalIcon from './ComicScreenplayIcon.jsx';
import ComicCharacterDesignDialog from './ComicCharacterDesignDialog.jsx';
import { applyCharacterDesign, characterDesignSavedSelection } from './comic-character-design.js';
import './comic-screenplay-asset-dialog.css';

const labels = { character: '角色', scene: '场景', prop: '道具' };
export default function ComicScreenplayAssetDialog({ projectId, workspace, asset, imageOnly = false, type = 'character', degree, media, readOnly, canGenerate, onSaved, onClose, onMediaChanged }) {
  const confirmForProject=useComicConfirmation(`${projectId}:${readOnly}:${canGenerate}`);
  const [draft, setDraft] = useState(() => assetEditorDraft(asset, type, degree));
  const [designOpen,setDesignOpen]=useState(false);
  const [mode, setMode] = useState('ai'), [file, setFile] = useState(null), [fileUrl, setFileUrl] = useState(''), [selectedImage, setSelectedImage] = useState(''), [references, setReferences] = useState([]), [query, setQuery] = useState('');
  const [models, setModels] = useState([]), [modelLoading, setModelLoading] = useState(true), [modelError, setModelError] = useState(''), [modelOverride, setModelOverride] = useState(''), [attempt, setAttempt] = useState(0);
  const [busy, setBusy] = useState(''), [error, setError] = useState(''), [notice, setNotice] = useState(''), [currentMedia, setCurrentMedia] = useState(media);
  const alive = useRef(true), controller = useRef(null), persisted = useRef(asset || null), retryKeys = useRef(new Map()), uncertainCreate = useRef(false), uploadedFile=useRef(null);
  useEffect(() => { alive.current = true; return () => { alive.current = false; controller.current?.abort(); }; }, []);
  useEffect(() => { if (!file) { setFileUrl(''); return; } const url = URL.createObjectURL(file); setFileUrl(url); return () => URL.revokeObjectURL(url); }, [file]);
  useEffect(() => { let active = true; setModelLoading(true); setModelError(''); listAvailableModels('image').then(rows => { if (active) setModels(rows); }).catch(e => { if (active) setModelError(e.message || '图片模型读取失败'); }).finally(() => { if (active) setModelLoading(false); }); return () => { active = false; }; }, [attempt]);
  const label = labels[draft.type] || '资产', editing = !!asset;
  const images = (currentMedia?.items || []).filter(item => item.kind === 'image' && ['succeeded', 'ready', 'completed'].includes(item.status));
  const preview = fileUrl || safeComicMediaUrl(images.find(item => item.id === selectedImage)?.url) || safeComicMediaUrl(comicAssetMediaImage(currentMedia, persisted.current?.id)?.url);
  const selection = comicImageModelSelection(workspace?.project?.creative_config, { models, references, override: modelOverride, loading: modelLoading, loadError: modelError });
  const patch = values => setDraft(current => ({ ...current, ...values }));
  async function refresh() { const data = await getComicMedia(projectId); if (alive.current) { setCurrentMedia(data); onMediaChanged?.(data); } return data; }
  async function persist() {
    const input = assetEditorPayload({ ...draft, ...(persisted.current ? { id: persisted.current.id, revision: persisted.current.revision } : {}) });
    if(uncertainCreate.current)throw new Error('新增资产保存结果尚未确认，请关闭后刷新资产列表核对，避免重复新增。');
    let saved;try{saved = persisted.current ? await updateComicScreenplayAsset(projectId, persisted.current.id, input) : await createComicScreenplayAsset(projectId, input);}catch(e){if(!persisted.current&&(!e.status||e.status>=500))uncertainCreate.current=true;throw e;}
    if (!saved?.id) throw new Error('资产保存结果缺少编号，请刷新核对');
    persisted.current = saved; if (alive.current) { setDraft(assetEditorDraft(saved)); onSaved?.(saved); setNotice('设定已保存。图片操作失败时可重试，不会重复新增资产。'); } return saved;
  }
  async function run(kind) {
    if (busy || readOnly) return;
    if (kind !== 'save' && !canGenerate) { setError('当前没有图片生成权限'); return; }
    if (kind === 'image' && (selection.error || selection.loading)) { setError(selection.error || '正在读取图片模型'); return; }
    if (kind === 'image' && !draft.prompt.trim()) { setError('请先填写外观描述或提示词'); return; }
    setError(''); setBusy(kind); let activeFingerprint = '', activeServerId = ''; const abort = new AbortController(); controller.current = abort;
    try {
      const saved = await persist(); if (!alive.current || abort.signal.aborted) return;
      if (kind === 'save') {
        if (file) { if(uploadedFile.current?.file!==file){const uploaded = await uploadComicMedia(projectId, file, { kind: 'image', targetId: saved.id }); if (!uploaded.item?.id) throw new Error('图片上传未返回素材编号');uploadedFile.current={file,id:uploaded.item.id};} await actComicAssetMedia(projectId, { action: 'select', targetId: saved.id, itemId: uploadedFile.current.id }, { signal: abort.signal }); }
        else if (selectedImage) await actComicAssetMedia(projectId, { action: 'select', targetId: saved.id, itemId: selectedImage }, { signal: abort.signal });
        await refresh(); if (alive.current) onClose(); return;
      }
      const modelId = kind === 'asset_process' ? undefined : selection.modelId;
      const fingerprint = JSON.stringify([kind, saved.id, modelId, draft.prompt, references]); activeFingerprint = fingerprint;
      let key = retryKeys.current.get(fingerprint); if (!key) { key = crypto.randomUUID(); retryKeys.current.set(fingerprint, key); }
      activeServerId = await comicMediaServerRequestId(kind, key);
      const data = await refresh();
      if(data.items?.some(item=>item.targetId===saved.id && ['image','asset_process'].includes(item.kind) && item.status==='running' && item.id!==activeServerId)) throw new Error('这个资产还有图片请求正在执行，请等待完成后刷新核对，不会重复提交。');
      const result = await actComicAssetMedia(projectId, { action: kind, targetId: saved.id, modelId, ...(kind === 'image' ? { referenceItemIds: references } : {}), expectedRevision: Number(data.settings?.revision || 0), idempotencyKey: key }, { signal: abort.signal });
      if (result?.status === 'failed' || result?.error) throw new Error(result.error || '图片生成失败');
      if (result?.id) await actComicAssetMedia(projectId, { action: 'select', targetId: saved.id, itemId: result.id }, { signal: abort.signal }); await refresh(); retryKeys.current.delete(fingerprint); if (alive.current) setNotice('图片已生成并保存，可在右侧预览。');
    } catch (e) { if (alive.current) { setError(abort.signal.aborted ? '已取消当前请求。已保存的设定会保留，请刷新核对图片结果。' : e.message || '操作失败'); try { const latest = await refresh(); if (activeServerId && latest.items?.some(item=>item.id===activeServerId && ['failed','cancelled'].includes(item.status))) retryKeys.current.delete(activeFingerprint); } catch {} } }
    finally { if (alive.current) setBusy(''); }
  }
  function cancel() { if (busy) { controller.current?.abort(); return; } onClose(); }
  useEffect(() => { if (!currentMedia) void refresh().catch(e => { if (alive.current) setError(e.message); }); }, []);
  const uploadControl = <Upload accept="image/png,image/jpeg,image/webp" showUploadList={false} disabled={readOnly || !!busy} beforeUpload={value => { setFile(value); setSelectedImage(''); return false; }}><Button disabled={readOnly || !!busy} icon={<OriginalIcon name="CloudUploadOutline" size={15} />}>上传图片</Button></Upload>;
  const fields = <div className="comic-asset-dialog-fields">
    <div className={draft.type === 'character' ? 'comic-asset-dialog-pair' : ''}><label>{editing ? `${label}名称` : '名称'}<Input aria-label={`${label}名称`} disabled={readOnly || !!busy} value={draft.name} onChange={e => patch({ name: e.target.value })} placeholder={`请输入${label}名称`} /></label>{draft.type === 'character' ? <label>{editing ? '别名' : '角色描述'}<Input disabled={readOnly || !!busy} value={editing ? draft.alias : draft.identity} onChange={e => patch(editing ? { alias: e.target.value } : { identity: e.target.value })} placeholder={editing ? '请输入别名（可选）' : '请输入角色身份'} /></label> : null}</div>
    {editing && draft.type === 'character' ? <label>身份<Input value={draft.identity} disabled={readOnly || !!busy} onChange={e => patch({ identity: e.target.value })} placeholder="请输入身份" /></label> : null}
    {editing && draft.type !== 'character' ? <label>{label}描述<Input.TextArea rows={4} value={draft.description} disabled={readOnly || !!busy} onChange={e => patch({ description: e.target.value })} /></label> : null}
    {!editing ? <label>{draft.type === 'character' ? '形象生成方式' : '图片生成方式'}<div className="comic-asset-mode-switch" role="tablist" aria-label="图片生成方式">{[['ai', 'AI生成', 'SparklesOutline'], ['upload', '上传图片', 'CloudUploadOutline'], ['library', '素材库', 'ImagesOutline']].map(([key, text, icon]) => <button key={key} type="button" role="tab" aria-selected={mode === key} className={mode === key ? 'is-active' : ''} disabled={!!busy} onClick={() => { setMode(key); setFile(null); setSelectedImage(''); }}><OriginalIcon name={icon} size={14} />{text}</button>)}</div></label> : null}
    {editing || mode === 'ai' ? <><label><span className="comic-asset-design-field-heading"><span>{editing ? '外观描述' : '提示词'}</span>{draft.type==='character'?<Button size="small" disabled={readOnly||!!busy} icon={<OriginalIcon name="PersonOutline" size={14}/>} onClick={event=>{event.preventDefault();if(!editing)setMode('ai');setDesignOpen(true);}}>角色设计</Button>:null}</span><Input.TextArea aria-label="资产外观描述" rows={editing ? 6 : 4} disabled={readOnly || !!busy} value={draft.prompt} onChange={e => patch({ prompt: e.target.value })} placeholder="描述外貌、服装、环境或材质等画面内容" /></label><label>参考图片（可选）<div className="comic-asset-reference-grid">{images.map(item => <button type="button" disabled={!!busy} key={item.id} aria-pressed={references.includes(item.id)} className={references.includes(item.id) ? 'is-selected' : ''} onClick={() => setReferences(rows => rows.includes(item.id) ? rows.filter(id => id !== item.id) : [...rows, item.id])}><AssetImage src={item.url} name={item.name || '参考图'} /><span>{item.name || '参考图'}</span></button>)}</div>{!images.length ? <small>当前作品还没有可用参考图，可先上传图片。</small> : null}</label></> : null}
    {!editing && mode === 'upload' ? <div className="comic-asset-upload-choice">{fileUrl ? <Image src={fileUrl} width={100} height={100} alt="待保存图片" /> : <OriginalIcon name="ImageOutline" size={40} />}{uploadControl}<small>{file?.name || '选择图片后，点击保存才会上传。也可只保存手动设定。'}</small></div> : null}
    {!editing && mode === 'library' ? <section><Input prefix={<OriginalIcon name="SearchOutline" size={14} />} placeholder="搜索当前作品素材" value={query} onChange={e => setQuery(e.target.value)} /><p className="comic-asset-dialog-note">当前作品已保存图片</p><div className="comic-asset-reference-grid">{images.filter(item => (item.name || '').includes(query)).map(item => <button key={item.id} type="button" aria-pressed={selectedImage === item.id} className={selectedImage === item.id ? 'is-selected' : ''} onClick={() => { setSelectedImage(item.id); setFile(null); }}><AssetImage src={item.url} name={item.name || '素材图'} /><span>{item.name || '素材图'}</span></button>)}</div>{!images.length ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="当前作品素材库为空" /> : null}</section> : null}
    {draft.type === 'character' ? <label>音色描述<Input.TextArea rows={editing ? 4 : 2} disabled={readOnly || !!busy} value={draft.voice} onChange={e => patch({ voice: e.target.value })} placeholder="角色音色描述（可选）" /></label> : !editing ? <label>{label}描述<Input.TextArea rows={3} disabled={readOnly || !!busy} value={draft.description} onChange={e => patch({ description: e.target.value })} /></label> : null}
  </div>;
  return <><Modal open width={imageOnly ? 520 : editing ? 960 : 680} style={{ maxWidth: '92vw' }} className="comic-screenplay-asset-dialog" title={<Space><OriginalIcon name={draft.type === 'character' ? 'PeopleOutline' : draft.type === 'scene' ? 'LocationOutline' : 'CubeOutline'} size={20} />{imageOnly ? 'AI生图 · ' : editing ? '编辑' : '新增'}{label}</Space>} onCancel={cancel} maskClosable={!busy} keyboard={!busy} destroyOnHidden footer={<div className="comic-asset-dialog-footer"><Select aria-label="资产图片模型" showSearch optionFilterProp="label" disabled={!!busy || modelLoading} loading={modelLoading} options={selection.options} value={selection.modelId || undefined} onChange={setModelOverride} placeholder="选择图片模型" /><Space><Button onClick={cancel}>{busy ? '取消当前请求' : '取消'}</Button>{!readOnly && !imageOnly ? <Button disabled={!!busy} loading={busy === 'save'} onClick={() => run('save')}>{!editing && mode === 'ai' ? '仅保存设定' : '保存'}</Button> : null}{(imageOnly || !editing && mode === 'ai') ? <Button type="primary" icon={<OriginalIcon name="SparklesOutline" size={15} />} disabled={readOnly || !canGenerate || !!busy || !!selection.error || modelLoading || !currentMedia?.capabilities?.image?.available} loading={busy === 'image'} onClick={() => run('image')}>生成</Button> : null}</Space></div>}>
    {error ? <Alert type="error" showIcon message={error} /> : null}{notice ? <Alert type="info" showIcon message={notice} /> : null}{selection.error && (editing || mode === 'ai') ? <Alert type="warning" message={selection.error} action={modelError ? <Button size="small" onClick={() => setAttempt(value => value + 1)}>重新读取模型</Button> : undefined} /> : null}
    <Spin spinning={busy === 'save'}><div className={!imageOnly && editing ? 'comic-asset-editor-columns' : 'comic-asset-add-body'}>{imageOnly ? <div className="comic-asset-dialog-fields"><b>{draft.name}</b>{preview ? <Image src={preview} alt={draft.name} style={{maxHeight:160,objectFit:'contain'}} /> : null}<label>画面提示词<Input.TextArea rows={5} value={draft.prompt} onChange={e=>patch({prompt:e.target.value})} disabled={!!busy}/></label><label>参考图片（可选）<Select mode="multiple" value={references} options={images.map(item=>({value:item.id,label:item.name||item.id}))} onChange={setReferences} disabled={!!busy}/></label></div> : fields}{editing && !imageOnly ? <aside className="comic-asset-editor-image"><h4>{label}图片</h4>{preview ? <Image src={preview} alt={draft.name || label} width="100%" /> : <div className="comic-asset-no-image"><OriginalIcon name="ImageOutline" size={40} /><span>点击下方上传图片</span></div>}<p>上传或替换图片，也可以根据当前画面提示词生成图片。</p><Space wrap>{uploadControl}<Button icon={<OriginalIcon name="SparklesOutline" size={14} />} type="primary" loading={busy === 'image'} disabled={readOnly || !canGenerate || !!busy || !!selection.error || modelLoading || !currentMedia?.capabilities?.image?.available} onClick={() => run('image')}>AI生成图片</Button>{draft.type === 'character' && currentMedia?.capabilities?.asset_process?.available ? <Button disabled={readOnly || !canGenerate || !!busy || !comicAssetMediaImage(currentMedia, persisted.current?.id)} loading={busy === 'asset_process'} onClick={() => confirmForProject({ title: '处理当前角色图片？', content: '使用模型配置中已保存的资产处理要求和生图模型；不会叠加生图核心指令。', okText: '处理图片', cancelText: '取消', onOk: () => run('asset_process') })}>资产处理</Button> : null}</Space>{file ? <small>待保存：{file.name}</small> : null}</aside> : null}</div></Spin>
  </Modal>{designOpen?<ComicCharacterDesignDialog open initialSelection={characterDesignSavedSelection(draft)} onCancel={()=>setDesignOpen(false)} onSave={result=>{const updated=applyCharacterDesign(draft,result);const manualTail=!!draft.extra?.character_design_text&&updated.prompt===draft.prompt&&!!draft.prompt.trim();setDraft(updated);setDesignOpen(false);setNotice(result.text?'角色设计已应用到当前描述，点击资产窗口保存后生效。':manualTail?'设计选择已清空；提示词含手动修改，已保留原文，请核对后保存。':'角色设计已重置，已清除可确认的设计尾段，手写人物描述已保留。');}}/>:null}</>;
}
