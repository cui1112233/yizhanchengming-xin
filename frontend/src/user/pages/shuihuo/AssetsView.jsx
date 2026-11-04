import { useEffect, useMemo, useState } from 'react';
import { Button, Checkbox, Input, Modal, Popconfirm, Select, Tag, message } from 'antd';
import { analyzeAssetsAndBindings, createAsset, deleteAsset, downloadAssetImage, downloadGeneratedAssetImage, generateAssetImages, getAssetGenerationConfig, listAssetImages, listModels, listShuihuoPresetSlots, listTasks, replaceAssetImage, saveAssetGenerationConfig, setPrimaryAssetImage, updateAsset, uploadAssetImage } from '../../../shared/api/shuihuoProduction';
import { withDefaultImageModel } from './modelDefaults';
import { IMAGE_ASPECT_RATIO_OPTIONS } from '../../../shared/imageAspectRatios';

const categories = [{ value: 'all', label: '全部' }, { value: 'character', label: '人物' }, { value: 'scene', label: '场景' }, { value: 'prop', label: '道具' }, { value: 'voice', label: '音色' }];
const assetCategories = categories.filter(item => item.value !== 'all');
const presetTabs = [
  { value: 'character', label: 'AI角色' },
  { value: 'scene', label: 'AI场景' },
  { value: 'prop', label: 'AI道具' },
  { value: 'voice', label: 'AI音色' },
  { value: 'character-library', label: '角色库' },
  { value: 'scene-library', label: '场景库' },
  { value: 'voice-library', label: '音色库' }
];
const blankAsset = { category: 'character', name: '', prompt: '', referenceObjectKey: '' };
const defaultGenerationConfig = { textModelId: null, imageModelId: null, audioModelId: null, promptTemplateId: null, characterPresetId: null, scenePresetId: null, styleId: null, characterSheetPresetId: null, aspectRatio: '16:9', styleReferenceMediaId: null, threeView: false };
const characterSheetOptions = [
  { value: 'shuihuo-character-color-sheet', label: '配色图人物设计' },
  { value: 'shuihuo-character-accessory-sheet', label: '三视图配饰设计' },
  { value: 'shuihuo-character-three-view', label: '三视图' },
  { value: 'shuihuo-character-expression-sheet', label: '角色表情多状态' },
  { value: 'shuihuo-character-single-view', label: '单视图' }
];
const ttsVoiceOptions = [
  { value: 'zh-CN-XiaoxiaoNeural', label: '晓晓（女声·温柔）' },
  { value: 'zh-CN-XiaochenNeural', label: '晓辰（女声·知性）' },
  { value: 'zh-CN-YunxiNeural', label: '云希（男声·清朗）' },
  { value: 'zh-CN-YunyangNeural', label: '云扬（男声·阳光）' },
  { value: 'zh-CN-XiaoyiNeural', label: '晓伊（女声·甜美）' },
  { value: 'zh-CN-YunjianNeural', label: '云健（男声·稳重）' }
];
const defaultStyles = [
  { id: 'ancient', name: '古风1', prompt: '中国古风插画，工整构图，细腻人物刻画，柔和光影，统一服饰与场景质感', referenceName: '' },
  { id: 'comic', name: '赛璐璐风格', prompt: '高清赛璐璐二次元插画，清晰线稿，明快配色，干净背景层次', referenceName: '' },
  { id: 'suspense', name: '悬疑风格', prompt: '悬疑漫画视觉，强明暗对比，冷色环境光，紧张构图', referenceName: '' },
  { id: 'ink', name: '国风水墨', prompt: '中国水墨画风，留白构图，墨色层次，宣纸质感', referenceName: '' }
];
const assetGenerationMessageKey = 'shuihuo-asset-image-generation';
const assetTaskPollInterval = 2000;
const assetTaskPollAttempts = 60;

function categoryLabel(category) { return assetCategories.find(item => item.value === category)?.label || '人物'; }
function waitForAssetTaskPoll() { return new Promise(resolve => window.setTimeout(resolve, assetTaskPollInterval)); }
function hasUsableAssetImageModel(asset, fallbackModelId, imageModels) {
  const overrideModelId = asset?.imageModelId;
  const selectedModelId = overrideModelId !== undefined && overrideModelId !== null && overrideModelId !== '' ? overrideModelId : fallbackModelId;
  return selectedModelId !== undefined && selectedModelId !== null && selectedModelId !== '' && imageModels.some(model => String(model.id) === String(selectedModelId));
}

function GeneratedAssetPreview({ image }) {
  const [url, setUrl] = useState('');
  useEffect(() => {
    let active = true;
    let objectUrl = '';
    downloadGeneratedAssetImage(image.id).then(blob => {
      if (!active) return;
      objectUrl = URL.createObjectURL(blob);
      setUrl(objectUrl);
    }).catch(() => { if (active) setUrl(''); });
    return () => { active = false; if (objectUrl) URL.revokeObjectURL(objectUrl); };
  }, [image.id]);
  return url ? <img src={url} alt="资产生成结果" /> : <span>图片加载失败</span>;
}

function AssetEditorImagePreview({ asset, images, replacementFile }) {
  const [url, setUrl] = useState('');
  const primaryImage = images.find(image => image.isPrimary) || images[0];
  const useManualReference = Boolean(asset?.referenceObjectKey && asset?.manuallyEdited);

  useEffect(() => {
    let active = true;
    let objectUrl = '';
    setUrl('');
    const source = replacementFile
      ? Promise.resolve(replacementFile)
      : !asset?.id
        ? Promise.resolve(null)
        : useManualReference || !primaryImage
          ? downloadAssetImage(asset.id)
          : downloadGeneratedAssetImage(primaryImage.id);
    source.then(blob => {
      if (!blob || !active) return;
      objectUrl = URL.createObjectURL(blob);
      setUrl(objectUrl);
    }).catch(() => { if (active) setUrl(''); });
    return () => {
      active = false;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [asset?.id, primaryImage?.id, replacementFile, useManualReference]);

  if (url) return <><img src={url} alt={`${asset?.name || '资产'}图片`} /><span className="shuihuo-asset-image-replace">点击或拖放替换图片</span></>;
  return <div className="shuihuo-asset-image-placeholder"><span>☁</span><b>{replacementFile?.name || '点击或拖放上传图片'}</b></div>;
}

function readFileAsDataURL(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
}

export function AssetsView({ data, onRefresh, embedded = false }) {
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState(null);
  const [filter, setFilter] = useState('all');
  const [libraryTab, setLibraryTab] = useState('character');
  const [search, setSearch] = useState('');
  const [saving, setSaving] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [models, setModels] = useState([]);
  const [modelId, setModelId] = useState();
  const [generationConfig, setGenerationConfig] = useState(defaultGenerationConfig);
  const [assetPresetOptions, setAssetPresetOptions] = useState([]);
  const [assetImageFile, setAssetImageFile] = useState(null);
  const [selectedAssetIDs, setSelectedAssetIDs] = useState([]);
  const [assetImages, setAssetImages] = useState({});
  const [styles, setStyles] = useState(defaultStyles);
  const [selectedStyleId, setSelectedStyleId] = useState('');
  const [styleOpen, setStyleOpen] = useState(false);
  const [styleEditor, setStyleEditor] = useState(null);
  const [assetPreview, setAssetPreview] = useState(null);
  const textModels = useMemo(() => models.filter(model => model.kind === 'text'), [models]);
  const imageModels = useMemo(() => models.filter(model => model.kind === 'image' && model.enabled), [models]);
  const audioModels = useMemo(() => models.filter(model => model.kind === 'audio'), [models]);
  const voiceAssets = useMemo(() => (data.assets || []).filter(asset => asset.category === 'voice'), [data.assets]);
  const promptAssets = useMemo(() => (data.assets || []).filter(asset => {
    const categoryMatch = filter === 'all' || asset.category === filter;
    const text = `${asset.name || ''} ${asset.prompt || ''}`.toLowerCase();
    return categoryMatch && (!search.trim() || text.includes(search.trim().toLowerCase()));
  }), [data.assets, filter, search]);
  const selectedImageAssets = selectedAssetIDs.map(id => (data.assets || []).find(asset => asset.id === id)).filter(asset => asset && asset.category !== 'voice');
  const selectedImageAssetCount = selectedImageAssets.length;
  const imageAssets = useMemo(() => (data.assets || []).filter(asset => asset.category === libraryTab && (assetImages[asset.id] || []).length), [assetImages, data.assets, libraryTab]);
  const libraryImages = useMemo(() => imageAssets.flatMap(asset => (assetImages[asset.id] || []).map(image => ({ ...image, asset }))), [assetImages, imageAssets]);

  useEffect(() => () => {
    if (assetPreview?.url) URL.revokeObjectURL(assetPreview.url);
  }, [assetPreview?.url]);

  async function openAssetPreview(item) {
    setAssetPreview({ item, url: '' });
    try {
      const blob = await downloadGeneratedAssetImage(item.id);
      const url = URL.createObjectURL(blob);
      setAssetPreview(current => {
        if (current?.url) URL.revokeObjectURL(current.url);
        return { item, url };
      });
    } catch (error) {
      setAssetPreview(null);
      message.error(error.message || '读取预设图片失败');
    }
  }

  useEffect(() => {
    let active = true;
    Promise.all([listModels(), getAssetGenerationConfig(data.project.id), listShuihuoPresetSlots()]).then(([modelResult, config, presetResult]) => {
      if (!active) return;
      const nextModels = (modelResult.models || []).filter(model => typeof model?.id === 'number' && model?.name && model?.kind);
      const nextConfig = withDefaultImageModel({ ...defaultGenerationConfig, ...config }, nextModels);
      setModels(nextModels);
      setGenerationConfig(nextConfig);
      setAssetPresetOptions(presetResult.presets || []);
      if (nextConfig.textModelId && nextModels.some(model => model.id === nextConfig.textModelId && model.kind === 'text')) {
        setModelId(nextConfig.textModelId);
      }
    }).catch(error => message.error(error.message || '读取预设生成配置失败'));
    return () => { active = false; };
  }, [data.project.id]);

  useEffect(() => {
    try {
      const saved = JSON.parse(localStorage.getItem(`shuihuo.asset-styles.${data.project.id}`) || 'null');
      setStyles(Array.isArray(saved) && saved.length ? saved : defaultStyles);
    } catch { setStyles(defaultStyles); }
    setSelectedStyleId('');
  }, [data.project.id]);

  function saveStyles(next) {
    setStyles(next);
    localStorage.setItem(`shuihuo.asset-styles.${data.project.id}`, JSON.stringify(next));
  }

  useEffect(() => {
    let active = true;
    const imageAssets = (data.assets || []).filter(asset => asset.category !== 'voice');
    Promise.all(imageAssets.map(asset => listAssetImages(asset.id).then(result => [asset.id, result.images || []]).catch(() => [asset.id, []]))).then(entries => {
      if (active) setAssetImages(Object.fromEntries(entries));
    });
    return () => { active = false; };
  }, [data.assets]);

  async function updateGenerationConfig(patch) {
    const next = { ...generationConfig, ...patch };
    setGenerationConfig(next);
    if (patch.textModelId !== undefined) setModelId(patch.textModelId ?? undefined);
    try {
      const saved = await saveAssetGenerationConfig(data.project.id, next);
      setGenerationConfig({ ...defaultGenerationConfig, ...saved });
    } catch (error) {
      message.error(error.message || '保存预设生成配置失败');
    }
  }

  async function save() {
    if (!editing?.name?.trim()) { message.warning('请填写资产名称'); return; }
    if (editing.imageModelId && !imageModels.some(model => model.id === editing.imageModelId)) { message.warning('当前图片模型不可用，请使用项目默认或选择已启用模型'); return; }
    setSaving(true);
    try {
      const payload = { ...editing, name: editing.name.trim(), source: 'manual', manuallyEdited: true };
      if (editing.id) {
        await updateAsset(editing.id, payload);
        if (editing.category !== 'voice' && assetImageFile) {
          await replaceAssetImage(editing.id, { filename: assetImageFile.name, dataUrl: await readFileAsDataURL(assetImageFile) });
        }
      }
      else if (editing.category === 'voice') await createAsset(data.project.id, payload);
      else {
        if (!assetImageFile) { message.warning('请上传资产图片'); return; }
        const dataUrl = await readFileAsDataURL(assetImageFile);
        await uploadAssetImage(data.project.id, { ...payload, filename: assetImageFile.name, dataUrl });
      }
      setOpen(false); setEditing(null); setAssetImageFile(null); await onRefresh(); message.success('资产已保存');
    } catch (error) { message.error(error.message || '保存资产失败'); } finally { setSaving(false); }
  }

  async function remove(asset) { try { await deleteAsset(asset.id); await onRefresh(); message.success('资产已删除'); } catch (error) { message.error(error.message || '删除资产失败'); } }

  async function monitorAssetImageGeneration(taskIDs, total) {
    const pendingTaskIDs = new Set(taskIDs);
    for (let attempt = 0; attempt < assetTaskPollAttempts; attempt += 1) {
      try {
        const result = await listTasks(data.project.id);
        const tasks = (result.tasks || []).filter(task => pendingTaskIDs.has(task.id));
        const succeeded = tasks.filter(task => task.status === 'succeeded').length;
        const failed = tasks.filter(task => task.status === 'failed' || task.status === 'cancelled').length;
        if (succeeded + failed >= total) {
          await onRefresh();
          if (failed) {
            message.error({ key: assetGenerationMessageKey, content: `已生成 ${succeeded} 张图片，${failed} 张失败`, duration: 6 });
          } else {
            message.success({ key: assetGenerationMessageKey, content: `已生成 ${succeeded} 张图片` });
          }
          return;
        }
        const progress = succeeded ? `，已完成 ${succeeded} 张` : '';
        message.loading({ key: assetGenerationMessageKey, content: `生成 ${total} 张图片中...${progress}`, duration: 0 });
      } catch {
        message.warning({ key: assetGenerationMessageKey, content: '图片任务已提交，正在后台生成，可稍后在图片库查看', duration: 6 });
        return;
      }
      await waitForAssetTaskPoll();
    }
    message.info({ key: assetGenerationMessageKey, content: '图片仍在生成中，可稍后在图片库查看', duration: 6 });
  }

  async function generateSelectedAssets() {
    const assetIds = selectedImageAssets.map(asset => asset.id);
    if (!assetIds.length) { message.warning('请先勾选需要生成图片的人物、场景或道具'); return; }
    if (!selectedImageAssets.every(asset => hasUsableAssetImageModel(asset, generationConfig.imageModelId, imageModels))) { message.warning('请为每个已选资产设置可用的图片模型，或选择项目默认模型'); return; }
    setGenerating(true);
    try {
      const result = await generateAssetImages(data.project.id, {
        assetIds,
        modelId: generationConfig.imageModelId,
        aspectRatio: generationConfig.aspectRatio,
        stylePrompt: styles.find(style => style.id === selectedStyleId)?.prompt || '',
        characterSheetPresetId: generationConfig.characterSheetPresetId || ''
      });
      const taskIDs = (result.tasks || []).map(task => task.id).filter(Number.isFinite);
      const total = taskIDs.length || assetIds.length;
      message.loading({ key: assetGenerationMessageKey, content: `生成 ${total} 张图片中...`, duration: 0 });
      setSelectedAssetIDs([]);
      await onRefresh();
      if (taskIDs.length) void monitorAssetImageGeneration(taskIDs, total);
    } catch (error) { message.error(error.message || '资产图片任务提交失败'); } finally { setGenerating(false); }
  }

  async function makePrimary(image) {
    try { await setPrimaryAssetImage(image.id); await onRefresh(); message.success('已设为资产主图'); }
    catch (error) { message.error(error.message || '设置资产主图失败'); }
  }

  async function analyze() {
    const selectedModelId = generationConfig.textModelId ?? modelId ?? textModels[0]?.id;
    if (!selectedModelId) { message.warning('请选择文本模型'); return; }
    const assetPresetId = generationConfig.assetPresetId || 'shuihuo-extract-assets';
    if (!assetPresetOptions.some(item => item.id === assetPresetId && item.slot === 'shuihuo.asset.extraction')) {
      message.error('人物场景、道具提取提示词已失效，请在管理后台发布或重新选择。');
      return;
    }
    setSaving(true);
    try {
      await analyzeAssetsAndBindings(data.project.id, { modelId: selectedModelId, assetPresetId });
      await onRefresh();
      message.success('智能预设已更新资产并分配到分镜');
    } catch (error) { message.error(error.message || '资产分析失败'); } finally { setSaving(false); }
  }

  return <section className={`shuihuo-view${embedded ? ' is-modal' : ''}`}>
    <div className="shuihuo-preset-toolbar">
      <Select value={generationConfig.textModelId ?? undefined} onChange={textModelId => updateGenerationConfig({ textModelId: textModelId ?? null })} placeholder={textModels.length ? '选择文本模型' : '没有文本模型'} options={textModels.map(model => ({ value: model.id, label: model.name }))} />
      <Select value={generationConfig.assetPresetId || 'shuihuo-extract-assets'} onChange={assetPresetId => updateGenerationConfig({ assetPresetId })} placeholder="人物场景、道具提取提示词" options={assetPresetOptions.filter(item => item.slot === 'shuihuo.asset.extraction').map(item => ({ value: item.id, label: item.name }))} />
      <Button type="primary" loading={saving} onClick={analyze}>智能预设</Button>
      <Select allowClear value={generationConfig.imageModelId ?? undefined} onChange={imageModelId => updateGenerationConfig({ imageModelId: imageModelId ?? null })} placeholder={imageModels.length ? '选择图片模型' : '没有图片模型'} options={imageModels.map(model => ({ value: model.id, label: model.name }))} />
      <Select value={generationConfig.aspectRatio} onChange={aspectRatio => updateGenerationConfig({ aspectRatio })} options={IMAGE_ASPECT_RATIO_OPTIONS} />
      <Button onClick={() => setStyleOpen(true)}>{styles.find(style => style.id === selectedStyleId)?.name || '选择风格'}</Button>
      {(filter === 'all' || filter === 'character') ? <Select allowClear value={generationConfig.characterSheetPresetId ?? undefined} onChange={characterSheetPresetId => updateGenerationConfig({ characterSheetPresetId: characterSheetPresetId ?? null })} placeholder="人物设定（仅角色）" options={characterSheetOptions} /> : null}

      <Button type="primary" loading={generating} disabled={!selectedImageAssetCount} title={selectedImageAssetCount ? `提交 ${selectedImageAssetCount} 个资产图片任务` : '请先在左侧勾选人物、场景或道具'} onClick={generateSelectedAssets}>AI生成（已选 {selectedImageAssetCount}）</Button>
      <Button onClick={() => { setEditing({ ...blankAsset, category: filter === 'all' ? 'character' : filter }); setOpen(true); }}>手动添加</Button>
      <Button onClick={() => { setEditing({ ...blankAsset, category: 'voice' }); setOpen(true); }}>添加音色</Button>
      <Select allowClear value={generationConfig.audioModelId ?? undefined} onChange={audioModelId => updateGenerationConfig({ audioModelId: audioModelId ?? null })} placeholder={audioModels.length ? '选择音频模型' : '没有音频模型'} options={audioModels.map(model => ({ value: model.id, label: model.name }))} />
    </div>
    <div className="shuihuo-preset-layout">
      <aside className="shuihuo-preset-list">
        <div className="shuihuo-preset-list-head"><strong>预设列表 ({promptAssets.length})</strong><Select size="small" value={filter} onChange={setFilter} options={categories} /><Button type="text" danger size="small" onClick={() => setSearch('')}>清空</Button></div>
        <div className="shuihuo-prompt-list">
          {promptAssets.map(asset => {
            const selected = selectedAssetIDs.includes(asset.id);
            return <article className={`shuihuo-prompt-row${selected ? ' is-selected' : ''}`} key={asset.id}>
              <div className="shuihuo-prompt-row-head">
                {asset.category !== 'voice' ? <Checkbox checked={selected} onChange={event => setSelectedAssetIDs(ids => event.target.checked ? [...new Set([...ids, asset.id])] : ids.filter(id => id !== asset.id))}>生成</Checkbox> : null}
                <Tag>{categoryLabel(asset.category)}</Tag><strong>{asset.name}</strong>
                <Button type="text" size="small" onClick={() => { setEditing({ ...asset }); setOpen(true); }}>编辑</Button>
                <Popconfirm title="删除资产后会同时解除其分段绑定。" onConfirm={() => remove(asset)}><Button type="text" danger size="small">删除</Button></Popconfirm>
              </div>
              <p>{asset.prompt || '尚未填写视觉提示词'}</p>
            </article>;
          })}
          {!promptAssets.length ? <div className="shuihuo-prompt-empty"><b>＋</b><p>暂无提示词，请点击“智能预设”或手动添加</p><Button onClick={() => { setEditing({ ...blankAsset, category: filter === 'all' ? 'character' : filter }); setOpen(true); }}>添加提示词</Button></div> : null}
        </div>
      </aside>
      <main className="shuihuo-preset-library">
        <div className="shuihuo-asset-tabs">{presetTabs.map(item => <button type="button" className={libraryTab === item.value ? 'active' : ''} onClick={() => setLibraryTab(item.value)} key={item.value}>{item.label}</button>)}</div>
        <Input.Search value={search} onChange={event => setSearch(event.target.value)} placeholder={`搜索${categoryLabel(libraryTab)}图片...`} className="shuihuo-preset-search" allowClear />
        <div className="shuihuo-image-library-grid">{libraryImages.map(item => <article className="shuihuo-image-card" key={item.id} onClick={() => { setEditing({ ...item.asset }); setOpen(true); }} role="button" tabIndex={0}>
          <button type="button" className="shuihuo-image-card-visual" title="放大查看预设图" onClick={event => { event.stopPropagation(); openAssetPreview(item); }}><GeneratedAssetPreview image={item} /></button>
          <div className="shuihuo-image-card-title"><strong>{item.asset.name}</strong><Tag>{categoryLabel(item.asset.category)}</Tag></div>
          <p>{item.assetPromptSnapshot || item.asset.prompt || '生成时未保存提示词快照'}</p>
          <div className="shuihuo-image-card-actions"><Button type="text" size="small" onClick={event => { event.stopPropagation(); makePrimary(item); }}>设主图</Button><span>点击编辑</span></div>
        </article>)}{!libraryImages.length ? <div className="shuihuo-image-empty"><b>图片库</b><p>暂无已生成图片</p><span>在左侧勾选人物、场景或道具后，点击顶部“AI生成”</span></div> : null}</div>
      </main>
    </div>
    <Modal title={editing?.id ? '编辑资产' : '添加资产'} open={open === true} width={810} className="shuihuo-asset-editor-modal" onCancel={() => { setOpen(false); setEditing(null); setAssetImageFile(null); }} footer={<div className="shuihuo-asset-editor-footer"><div className="shuihuo-asset-binding-source"><span>绑定资产</span><Button className={editing?.category === 'character' ? 'active' : ''} onClick={() => setEditing(item => ({ ...item, category: 'character' }))}>角色库</Button><Button className={editing?.category === 'scene' ? 'active' : ''} onClick={() => setEditing(item => ({ ...item, category: 'scene' }))}>场景库</Button><Button className={editing?.category === 'prop' ? 'active' : ''} onClick={() => setEditing(item => ({ ...item, category: 'prop' }))}>道具库</Button><Button className={editing?.category === 'voice' ? 'active' : ''} onClick={() => setEditing(item => ({ ...item, category: 'voice' }))}>音色库</Button></div><div><Button onClick={() => { setOpen(false); setEditing(null); setAssetImageFile(null); }}>取消</Button><Button type="primary" loading={saving} onClick={save}>保存</Button></div></div>}><div className="shuihuo-asset-editor"><div className="shuihuo-asset-editor-fields"><label>资产分类<Select value={editing?.category} onChange={category => setEditing(item => ({ ...item, category, voiceAssetId: category === 'character' ? item.voiceAssetId : null }))} options={assetCategories} /></label><label>名称<Input value={editing?.name} onChange={event => setEditing(item => ({ ...item, name: event.target.value }))} /></label>{editing?.category === 'voice' ? <label>配音工具音色<Select value={editing?.prompt || undefined} placeholder="选择可直接调用的音色" options={ttsVoiceOptions} onChange={prompt => setEditing(item => ({ ...item, prompt }))} /></label> : <label>外观描述 / AI 生图提示词<Input.TextArea rows={11} value={editing?.prompt} onChange={event => setEditing(item => ({ ...item, prompt: event.target.value }))} placeholder="输入人物、场景或道具的可视化提示词" /></label>}{editing?.category !== 'voice' ? <><label>图片模型<Select allowClear value={editing?.imageModelId ?? undefined} placeholder="使用项目默认" options={imageModels.map(model => ({ value: model.id, label: model.name }))} onChange={imageModelId => setEditing(item => ({ ...item, imageModelId: imageModelId ?? null }))} /></label><label>图片画幅<Select allowClear value={editing?.imageAspectRatio ?? undefined} placeholder="使用项目默认" options={IMAGE_ASPECT_RATIO_OPTIONS} onChange={imageAspectRatio => setEditing(item => ({ ...item, imageAspectRatio: imageAspectRatio ?? null }))} /></label></> : null}{editing?.category === 'character' ? <label>角色声音<Select allowClear value={editing?.voiceAssetId ?? undefined} placeholder={voiceAssets.length ? '选择该角色的配音音色' : '请先创建音色预设'} options={voiceAssets.map(asset => ({ value: asset.id, label: asset.name }))} onChange={voiceAssetId => setEditing(item => ({ ...item, voiceAssetId: voiceAssetId ?? null }))} /></label> : null}</div>{editing?.category !== 'voice' ? <label className="shuihuo-asset-image-drop" onDragOver={event => event.preventDefault()} onDrop={event => { event.preventDefault(); setAssetImageFile(event.dataTransfer.files?.[0] || null); }}><Input type="file" accept="image/*" onChange={event => setAssetImageFile(event.target.files?.[0] || null)} /><AssetEditorImagePreview asset={editing} images={editing?.id ? assetImages[editing.id] || [] : []} replacementFile={assetImageFile} /></label> : <div className="shuihuo-asset-image-drop is-voice"><span>♪</span><b>可绑定给角色的音色</b></div>}</div></Modal>
    <Modal title={assetPreview?.item?.asset?.name ? `${assetPreview.item.asset.name}预设图` : '预设图预览'} open={Boolean(assetPreview)} footer={null} onCancel={() => setAssetPreview(null)} width={900} destroyOnClose><div className="shuihuo-media-modal-preview is-image">{assetPreview?.url ? <img src={assetPreview.url} alt={`${assetPreview.item.asset.name}预设图`} /> : <span className="shuihuo-muted">读取真实预设图片...</span>}</div></Modal>
    <Modal title="选择风格" open={styleOpen} width={620} onCancel={() => setStyleOpen(false)} footer={null}>
      <div className="shuihuo-style-grid">{styles.map(style => <button type="button" className={`shuihuo-style-card${selectedStyleId === style.id ? ' is-active' : ''}`} key={style.id} onClick={() => { setSelectedStyleId(style.id); setStyleOpen(false); }}><span className={`shuihuo-style-swatch is-${style.id}`}>{style.referenceName || style.name.slice(0, 2)}</span><b>{style.name}</b><i onClick={event => { event.stopPropagation(); setStyleEditor({ ...style }); }}>编辑</i></button>)}</div>
      <Button block onClick={() => setStyleEditor({ id: `style-${Date.now()}`, name: '', prompt: '', referenceName: '' })}>添加风格</Button>
    </Modal>
    <Modal title="编辑风格" open={!!styleEditor} onCancel={() => setStyleEditor(null)} onOk={() => { if (!styleEditor?.name?.trim() || !styleEditor?.prompt?.trim()) { message.warning('请填写风格名称和风格提示词'); return; } const exists = styles.some(item => item.id === styleEditor.id); saveStyles(exists ? styles.map(item => item.id === styleEditor.id ? styleEditor : item) : [...styles, styleEditor]); setStyleEditor(null); }} okText="保存修改" cancelText="取消">
      <label className="shuihuo-style-field">风格名称<Input value={styleEditor?.name || ''} onChange={event => setStyleEditor(item => ({ ...item, name: event.target.value }))} /></label><label className="shuihuo-style-field">风格提示词<Input.TextArea rows={5} value={styleEditor?.prompt || ''} onChange={event => setStyleEditor(item => ({ ...item, prompt: event.target.value }))} /></label><label className="shuihuo-style-field">参考图片（可选）<Input type="file" accept="image/*" onChange={event => setStyleEditor(item => ({ ...item, referenceName: event.target.files?.[0]?.name || '' }))} />{styleEditor?.referenceName ? <span>{styleEditor.referenceName}</span> : null}</label>
    </Modal>
  </section>;
}
