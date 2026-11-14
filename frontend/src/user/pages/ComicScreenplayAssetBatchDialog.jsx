import {useComicConfirmation} from './comic-screenplay-confirm.js';
import { Alert, Button, Checkbox, Empty, Input, Modal, Select, Space } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { listAvailableModels } from '../../shared/api/modelCatalog.js';
import { getComicMedia } from '../../shared/api/comicScreenplayWorkflow.js';
import { actComicAssetMedia } from '../../shared/api/comicScreenplayAssetActions.js';
import { comicAssetMediaImage } from './comic-screenplay-stages.js';
import { comicImageModelSelection } from './comic-screenplay-creative.js';
import {comicMediaServerRequestId} from './comic-screenplay-asset-editor.js';
import AssetImage from './ComicScreenplayImage.jsx';
import OriginalIcon from './ComicScreenplayIcon.jsx';
import './comic-screenplay-asset-dialog.css';

export default function ComicScreenplayAssetBatchDialog({ projectId, workspace, assets, initialSelected = [], media, canGenerate, onClose, onMediaChanged }) {
  const confirmForProject=useComicConfirmation(`${projectId}:${canGenerate}`);
  const [selected, setSelected] = useState(initialSelected), [query, setQuery] = useState(''), [models, setModels] = useState([]), [modelOverride, setModelOverride] = useState(''), [loading, setLoading] = useState(true), [error, setError] = useState(''), [busy, setBusy] = useState(false), [progress, setProgress] = useState(''), [results, setResults] = useState({}), [currentMedia, setCurrentMedia] = useState(media);
  const alive = useRef(true), control = useRef(null), keys=useRef(new Map());
  useEffect(() => { alive.current = true; Promise.all([listAvailableModels('image'),getComicMedia(projectId)]).then(([value,latest]) => { if (alive.current) {setModels(value);setCurrentMedia(latest);onMediaChanged?.(latest);} }).catch(e => { if (alive.current) setError(e.message); }).finally(() => { if (alive.current) setLoading(false); }); return () => { alive.current = false; control.current?.abort(); }; }, []);
  const model = comicImageModelSelection(workspace?.project?.creative_config, { models, override: modelOverride, loading });
  const filtered = assets.filter(asset => asset.name.toLowerCase().includes(query.toLowerCase()));
  async function generate() {
    if (busy || !canGenerate || !currentMedia?.capabilities?.image?.available || !selected.length || model.error || loading) return;
    setBusy(true); setError(''); let active=null; const controller = new AbortController(); control.current = controller;
    try {
      for (let i = 0; i < selected.length; i++) {
        if (controller.signal.aborted || !alive.current) break;
        const asset = assets.find(item => item.id === selected[i]); if (!asset || results[asset.id] === 'succeeded') continue;
        setProgress(`${asset.name}（${i + 1}/${selected.length}）`);
        const current = await getComicMedia(projectId);
        if (current.items?.some(item => item.kind === 'image' && item.targetId === asset.id && item.status === 'running')) throw new Error(`${asset.name}已有生成请求正在执行，请完成后刷新核对。`);
        const fingerprint=JSON.stringify([asset.id,model.modelId,asset.revision]);let key=keys.current.get(fingerprint);if(!key){key=crypto.randomUUID();keys.current.set(fingerprint,key);}active={fingerprint,id:await comicMediaServerRequestId('image',key)};
        const result = await actComicAssetMedia(projectId, { action: 'image', targetId: asset.id, modelId: model.modelId, expectedRevision: Number(current.settings?.revision || 0), idempotencyKey: key }, { signal: controller.signal });
        if (!result?.id || result.status !== 'succeeded') throw new Error(result?.error || `${asset.name}没有返回成功图片`);
        await actComicAssetMedia(projectId, { action: 'select', targetId: asset.id, itemId: result.id }, { signal: controller.signal });
        keys.current.delete(fingerprint);active=null;
        if (alive.current) setResults(state => ({ ...state, [asset.id]: 'succeeded' }));
        const latest = await getComicMedia(projectId); if (alive.current) { setCurrentMedia(latest); onMediaChanged?.(latest); }
      }
    } catch (e) { if (alive.current) setError(controller.signal.aborted ? '已取消当前请求；后续图片不再提交。已提交的结果请刷新核对。' : e.message || '批量出图失败，后续请求已停止'); }
    finally { try { const latest = await getComicMedia(projectId); if (active && latest.items?.some(item=>item.id===active.id && ['failed','cancelled'].includes(item.status)))keys.current.delete(active.fingerprint); if (alive.current) { setCurrentMedia(latest); onMediaChanged?.(latest); } } catch {} if (alive.current) { setBusy(false); setProgress(''); } }
  }
  function requestGenerate(){const existing=selected.filter(id=>results[id]!=='succeeded'&&comicAssetMediaImage(currentMedia,id));if(existing.length){confirmForProject({title:'重新生成已有资产图片？',content:`所选有 ${existing.length} 个资产已有图片，将再次调用图片模型。旧图会保留，新图成功后设为当前图片。`,okText:'确认继续',cancelText:'取消',onOk:generate});}else void generate();}
  return <Modal open width={936} style={{ maxWidth: '92vw' }} className="comic-screenplay-asset-dialog" title="批量生成资产图片" onCancel={() => { if (busy) control.current?.abort(); else onClose(); }} maskClosable={!busy} keyboard={!busy} footer={<Space><span>{selected.length} 个已选 · {Object.values(results).filter(value => value === 'succeeded').length} 个本次完成</span><Button onClick={() => { if (busy) control.current?.abort(); else onClose(); }}>{busy ? '停止后续出图' : '关闭'}</Button><Button type="primary" loading={busy} disabled={!canGenerate || !currentMedia?.capabilities?.image?.available || busy || loading || !!model.error || !selected.some(id => results[id] !== 'succeeded')} onClick={requestGenerate}>开始生成</Button></Space>}>
    <p className="comic-asset-dialog-note">按选定顺序逐个生成，使用资产已保存的描述与作品出图配置。失败时停止，已成功图片会保留。</p>{error ? <Alert type="error" showIcon message={error} /> : null}{model.error ? <Alert type="warning" showIcon message={model.error} /> : null}
    <Space wrap style={{ marginBottom: 14 }}><Input disabled={busy} prefix={<OriginalIcon name="SearchOutline" size={14} />} placeholder="搜索资产" value={query} onChange={e => setQuery(e.target.value)} /><Select aria-label="批量资产图片模型" style={{ width: 230 }} loading={loading} disabled={busy || loading} value={model.modelId || undefined} options={model.options} onChange={id=>{setModelOverride(id);setResults({});}} /><Button disabled={busy} onClick={() => setSelected(assets.map(asset => asset.id))}>全选</Button><Button disabled={busy} onClick={() => setSelected(assets.filter(asset => !comicAssetMediaImage(currentMedia, asset.id)).map(asset => asset.id))}>只选缺图</Button><Button disabled={busy} onClick={() => setSelected([])}>取消选择</Button></Space>
    {busy ? <Alert type="info" showIcon message={`正在生成 ${progress}`} /> : null}<div className="comic-asset-batch-grid">{filtered.map(asset => <button key={asset.id} type="button" disabled={busy} className={selected.includes(asset.id) ? 'is-selected' : ''} aria-pressed={selected.includes(asset.id)} onClick={() => setSelected(rows => rows.includes(asset.id) ? rows.filter(id => id !== asset.id) : [...rows, asset.id])}><AssetImage src={comicAssetMediaImage(currentMedia, asset.id)?.url} name={asset.name} /><Checkbox checked={selected.includes(asset.id)} tabIndex={-1} /><b>{asset.name}</b><small>{results[asset.id] === 'succeeded' ? '本次已完成' : comicAssetMediaImage(currentMedia, asset.id) ? '已有图片' : '暂无图片'}</small></button>)}</div>{!filtered.length ? <Empty description="没有匹配的资产" /> : null}
  </Modal>;
}
