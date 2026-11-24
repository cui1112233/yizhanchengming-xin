import Taro, { useDidShow } from "@tarojs/taro";
import { Button, Input, Picker, ScrollView, Text, Textarea, View } from "@tarojs/components";
import { useEffect, useRef, useState } from "react";
import BrandHeader from "../../components/BrandHeader";
import AppIcon from "../../components/AppIcon";
import ScriptVideoProduction from "./ScriptVideoProduction";
import ScriptPersonalPrompts from "./ScriptPersonalPrompts";
import ScriptEntityEditor from "./ScriptAssetsEditor";
import ScriptHistoryPanel from "./ScriptHistoryPanel";
import { apiRequest } from "../../services/api";
import { readTtsOptions, synthesizeSpeech } from "../../services/tts";
import { buildExtractRequest, buildGenerateRequest, buildHistoryEntry, normalizeScriptEntities, parseScriptOutput, restoreScriptHistory, scriptConstraintChoices, updateScriptConstraint, scriptEntityKey, ScriptEntity, ScriptWorkspaceState, ScriptConstraints } from "../../services/scriptWorkspace";
import { withThemedPage } from "../../theme";
import { retainScriptReferences, validScriptAssets } from '../../services/scriptAssets';
import { onSessionInvalidated } from '../../services/sessionBoundary';
import './script-studio.scss';

const OPENINGS = ["连续开头", "爆款开头", "分段开头"];
const FORMATS = ["画布", "剧本", "剧情", "分镜", "Q版"];
const DURATIONS = ["10 秒", "15 秒"];
const LAYERS = [["baseSetup", "基本设定"], ["prefix", "视觉前缀"], ["quality", "质量要求"], ["restriction", "限制规则"], ["negative", "负面约束"]] as const;
type Option = { id: string; name: string; constraintCategory?: string; kind?: string; extractionPreset?: boolean; enabled?: boolean };
const emptyState: ScriptWorkspaceState = { novelText: "", characters: [], scenes: [], opening: "分段开头", format: "分镜", duration: "15 秒", matchAudio: false, constraints: {}, protagonists: [] };

function entityName(entity: any, fallback: string) {
  const data = entity?.data || entity;
  return String(data?.name || data?.角色名称 || data?.场景名称 || data?.名称 || entity?.name || fallback);
}
function options(payload: any): Option[] {
  const source = Array.isArray(payload) ? payload : payload?.models || payload?.entries || payload?.catalog;
  if (!Array.isArray(source)) throw new Error("服务端未返回有效选项目录");
  return source.filter((item) => item?.id || item?.key).map((item) => ({ ...item, id: String(item.id || item.key), name: String(item.name || item.label || item.id || item.key) }));
}

function ScriptWorkspace({ historyId = "" }: { historyId?: string }) {
  const [state, setState] = useState<ScriptWorkspaceState>(emptyState);
  const stateRef = useRef(state); stateRef.current = state;
  const [tab, setTab] = useState("编辑");
  const [models, setModels] = useState<Option[]>([]);
  const [imageModels, setImageModels] = useState<Option[]>([]);
  const [savedHistoryId, setSavedHistoryId] = useState("");
  const [personalCategory, setPersonalCategory] = useState("");
  const [catalog, setCatalog] = useState<Option[]>([]);
  const [catalogError, setCatalogError] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState("");
  const [result, setResult] = useState("");
  const [showConstraints, setShowConstraints] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [showHistory, setShowHistory] = useState(false);
  const [editorStage, setEditorStage] = useState<'source' | 'entities'>('source');
  const [entityDraft, setEntityDraft] = useState<{ characters: ScriptEntity[]; scenes: ScriptEntity[] }>({ characters: [], scenes: [] });
  const [editingEntities, setEditingEntities] = useState(false);
  const [sourceExpanded, setSourceExpanded] = useState(true);
  const [showAudio, setShowAudio] = useState(false);
  const [audioPlaying, setAudioPlaying] = useState(false);
  const audio = useRef<Taro.InnerAudioContext | null>(null);
  const audioFile = useRef("");
  const version = useRef(0);
  const catalogVersion = useRef(0);
  const restored = useRef("");
  const actionLock = useRef(false);
  const historyChoiceLock = useRef(false);
  const mounted = useRef(true);
  const patch = (next: Partial<ScriptWorkspaceState>) => { const creative = Object.keys(next).some(key => key !== 'videoTasks'); setState((current) => ({ ...current, ...next, ...(creative ? { videoTasks: {} } : {}) })); if (creative) { setResult(''); setSavedHistoryId(''); } };
  const clearAudio = () => {
    audio.current?.destroy(); audio.current = null;
    if (audioFile.current) Taro.getFileSystemManager().unlink({ filePath: audioFile.current, fail: () => undefined });
    audioFile.current = ""; setAudioPlaying(false);
  };
  useEffect(() => {
    const off = onSessionInvalidated(() => { version.current++; catalogVersion.current++; clearAudio(); setState(emptyState); setResult(''); setSavedHistoryId(''); setEditingEntities(false); setShowHistory(false); setShowSettings(false); setShowConstraints(false); setShowAudio(false); setBusy(''); setCatalog([]); setModels([]); setImageModels([]); restored.current=''; setTab('编辑'); setEditorStage('source'); setError('账号已切换，请重新打开私人定制'); });
    return () => { mounted.current=false; version.current += 1; catalogVersion.current += 1; clearAudio(); off(); };
  }, []);
  const loadCatalog = async () => {
    const current = ++catalogVersion.current;
    setCatalogError("");
    try {
      const [modelPayload, presetPayload] = await Promise.all([apiRequest("/api/models?kind=text"), apiRequest("/api/presets?module=script")]);
      if (current !== catalogVersion.current) return;
      const nextModels = options(modelPayload).filter(item => item.enabled !== false); const nextCatalog = options(presetPayload).filter(item => item.enabled !== false);
      setModels(nextModels); setCatalog(nextCatalog);
      void apiRequest("/api/models?kind=image").then(payload => { if (current === catalogVersion.current) setImageModels(options(payload).filter((item: any) => item.enabled !== false)); }).catch(cause => { if (current === catalogVersion.current) setCatalogError(`图片模型读取失败：${cause.message}`); });
      setState((value) => ({ ...value, textModelId: value.textModelId || nextModels[0]?.id || "", extractionPreset: value.extractionPreset || nextCatalog.find((item) => item.kind === "base" && item.extractionPreset)?.id || "" }));
    } catch (cause: any) { if (current === catalogVersion.current) setCatalogError(cause.message || "生成选项读取失败"); }
  };
  useDidShow(() => { void loadCatalog(); });
  const restoreHistory = async (id: string) => {
    if (!id || actionLock.current) return;
    actionLock.current = true;
    const current = ++version.current;
    setBusy('恢复历史'); setError('');
    try {
      const payload = await apiRequest(`/api/history/${encodeURIComponent(id)}`);
      if (current !== version.current || !mounted.current) return;
      const returned = (payload as any)?.entry || payload;
      if ((returned as any)?.id !== id) throw new Error('历史编号不匹配，未替换当前内容，请刷新历史列表');
      const next = restoreScriptHistory(payload); clearAudio(); setState(next); setResult(next.output); setSavedHistoryId(id); setTab(next.output ? '结果' : '编辑'); restored.current=id; setShowHistory(false); setSourceExpanded(false); setEditorStage('entities');
    } catch(cause:any) { if (current === version.current && mounted.current) setError(cause.message || '历史恢复失败'); }
    finally { actionLock.current=false; if (current === version.current && mounted.current) setBusy(''); }
  };
  useEffect(() => {
    if (!historyId || restored.current === historyId) return;
    void restoreHistory(historyId);
  }, [historyId]);

  const extract = async () => {
    if (actionLock.current) return;
    if (!state.novelText.trim() || !catalog.some(item => item.id === state.extractionPreset && item.extractionPreset) || !models.some(item => item.id === state.textModelId)) { setError("请填写原文并选择可用的文本模型和提取模板"); return; }
    actionLock.current=true;
    const current = ++version.current; setBusy("提取人物场景"); setError("");
    try {
      const payload = await apiRequest("/api/chat", { method: "POST", timeout: 180000, data: buildExtractRequest(state) });
      if (current !== version.current) return;
      const entities = normalizeScriptEntities(payload);
      if (!entities.characters.length && !entities.scenes.length) throw new Error("模型未返回人物或场景");
      patch({ ...entities, protagonists: [], videoTasks: {} }); setResult(""); setSavedHistoryId(""); setSourceExpanded(false); setEditorStage('entities');
    } catch (cause: any) { if (current === version.current) setError(cause.message || "提取失败"); }
    finally { actionLock.current=false; if (current === version.current) setBusy(""); }
  };
  const generate = async () => {
    if (actionLock.current) return;
    if (!state.novelText.trim() || (!state.characters.length && !state.scenes.length) || !models.some(item => item.id === state.textModelId)) { setError("请先填写原文、确认人物场景并选择可用模型"); return; }
    if (state.matchAudio && !state.audioTotalSeconds) { setError("请先为当前原文生成配音"); return; }
    actionLock.current=true;
    const current = ++version.current; setBusy("生成剧本"); setError("");
    try {
      let input = { ...state, constraints: JSON.parse(JSON.stringify(state.constraints || {})) as ScriptConstraints };
      if (input.constraints.prefix?.enabled && input.constraints.prefix.source === "system" && input.constraints.prefix.presetId === "script-constraint-prefix-smart-unified") {
        const style = await apiRequest<any>("/api/script/smart-unified-style", { method: "POST", timeout: 180000, data: { novelText: input.novelText, characters: input.characters, scenes: input.scenes } });
        if (current !== version.current) return;
        const prompt = style?.smartUnifiedStyle?.prompt;
        if (typeof prompt !== "string" || !prompt.trim()) throw new Error("智能统一视觉未返回有效风格");
        input = { ...input, visualStyle: prompt, constraints: { ...input.constraints, prefix: { ...input.constraints.prefix, smartUnifiedStyle: prompt } } };
      }
      const payload = await apiRequest("/api/chat", { method: "POST", timeout: 180000, data: buildGenerateRequest(input) });
      if (current !== version.current) return;
      const output = parseScriptOutput(payload); setState(value => ({...value,videoTasks:{}})); setResult(output); setSavedHistoryId(""); setTab("结果");
      try {
        const entry = buildHistoryEntry(input, output);
        await apiRequest("/api/history", { method: "POST", data: entry });
        const readback = await apiRequest<any>(`/api/history/${encodeURIComponent(entry.id)}`);
        if ((readback?.entry || readback)?.id !== entry.id) throw new Error("历史保存后的编号读回不一致");
        if (current === version.current) { setSavedHistoryId(entry.id); Taro.showToast({ title: "剧本已生成并保存", icon: "success" }); }
      } catch (cause: any) { if (current === version.current) setError(`剧本已生成，历史保存失败：${cause.message}`); }
    } catch (cause: any) { if (current === version.current) setError(cause.message || "生成失败"); }
    finally { actionLock.current=false; if (current === version.current) setBusy(""); }
  };
  const generateVoice = async () => {
    if (actionLock.current || !state.novelText.trim()) return;
    actionLock.current=true;
    const current = ++version.current; setBusy("生成原文配音"); setError(""); clearAudio(); patch({ audioTotalSeconds: null, matchAudio: false });
    try {
      const options = await readTtsOptions();
      if (current !== version.current) return;
      const filePath = await synthesizeSpeech(state.novelText.trim(), options);
      if (current !== version.current) return;
      audioFile.current = filePath;
      if (current !== version.current) { clearAudio(); return; }
      const context = Taro.createInnerAudioContext(); audio.current = context;
      const seconds = await new Promise<number>((resolve, reject) => {
        let interval: ReturnType<typeof setInterval> | undefined;
        let settled = false;
        const finish = (failure?: Error, duration?: number) => { if (settled) return; settled = true; clearTimeout(deadline); clearInterval(interval); failure ? reject(failure) : resolve(duration!); };
        const deadline = setTimeout(() => finish(new Error("配音时长读取超时，请重试")), 12000);
        context.onCanplay(() => { if (settled || interval) return; interval = setInterval(() => { if (context.duration > 0) finish(undefined, Math.ceil(context.duration)); }, 100); });
        context.onError(() => finish(new Error("配音文件无法播放"))); context.src = filePath;
      });
      if (current !== version.current) return;
      context.onEnded(() => setAudioPlaying(false)); context.onPause(() => setAudioPlaying(false));
      patch({ audioTotalSeconds: seconds });
    } catch (cause: any) { if (current === version.current) { clearAudio(); setError(cause.message || "原文配音失败"); } }
    finally { actionLock.current=false; if (current === version.current) setBusy(""); }
  };

  const pick = (label: string, list: string[], value: string, select: (value: string) => void) => <Picker disabled={Boolean(busy) || !list.length} range={list} value={Math.max(0, list.indexOf(value))} onChange={event => { const selected = list[Number(event.detail.value)]; if (!actionLock.current && selected !== undefined) select(selected); }}><View className="studio-setting"><Text className="studio-setting-label">{label}</Text><View className="studio-setting-value"><Text>{value || '暂无可用选项'}</Text><AppIcon name="chevron-right" size={12} /></View></View></Picker>;
  const pickOption = (label: string, list: Option[], id: string | undefined, select: (id: string) => void) => <Picker disabled={Boolean(busy) || !list.length} range={list.map(item => item.name)} value={Math.max(0,list.findIndex(item => item.id === id))} onChange={event => { const selected = list[Number(event.detail.value)]; if (!actionLock.current && selected) select(selected.id); }}><View className="studio-setting"><Text className="studio-setting-label">{label}</Text><View className="studio-setting-value"><Text>{list.find(item => item.id === id)?.name || (list.length ? '请选择' : '暂无可用选项')}</Text><AppIcon name="chevron-right" size={12} /></View></View></Picker>;
  const presets = catalog.filter(item => item.kind === 'base' && item.extractionPreset);
  const constraintCount = LAYERS.filter(([key]) => state.constraints?.[key]?.enabled).length;
  const hasEntities = Boolean(state.characters.length || state.scenes.length);
  const validModel = models.some(item => item.id === state.textModelId);
  const validPreset = presets.some(item => item.id === state.extractionPreset);
  const closePanel = () => { setEditingEntities(false); setShowConstraints(false); setShowAudio(false); setShowSettings(false); setShowHistory(false); setPersonalCategory(''); };
  const openEntities = () => { setError(''); setEntityDraft(JSON.parse(JSON.stringify({ characters: state.characters, scenes: state.scenes }))); setEditingEntities(true); };
  const readiness = !state.novelText.trim() ? '先放入你的小说原文' : !validModel ? '请选择可用文本模型' : !hasEntities && !validPreset ? '请选择人物场景提取模板' : !hasEntities ? '先提取人物场景，也可以手动添加' : state.matchAudio && !state.audioTotalSeconds ? '请先生成当前原文的配音' : '人物场景已就绪，可以开始创作';
  const actionDisabled = Boolean(busy) || !state.novelText.trim() || !validModel || (!hasEntities && !validPreset) || (hasEntities && state.matchAudio && !state.audioTotalSeconds);
  const applyAssets = () => { if (!validScriptAssets(entityDraft)) { setError('请为每个人物和场景填写名称'); return; } patch({ ...entityDraft, ...retainScriptReferences(entityDraft, state.protagonists || [], state.constraints), videoTasks: {} }); setEditingEntities(false); setEditorStage('entities'); setSourceExpanded(false); };
  const changeSource = (novelText: string) => { if (actionLock.current) return; version.current++; clearAudio(); patch({ novelText, characters: [], scenes: [], protagonists: [], constraints: { ...state.constraints, entityReferences: [] }, audioTotalSeconds: null, matchAudio: false, videoTasks: {} }); setEditorStage('source'); };
  const chooseHistory = async (id: string) => { if (actionLock.current || historyChoiceLock.current) return; historyChoiceLock.current=true; const current=version.current; try { if (state.novelText.trim() || result) { const confirmed = await Taro.showModal({ title: '打开历史剧本？', content: '当前页面内容将被历史版本替换，请先保留需要的文字。', confirmText: '打开历史', cancelText: '继续编辑' }); if (!confirmed.confirm || !mounted.current || current!==version.current) return; } await restoreHistory(id); } finally {historyChoiceLock.current=false;} };
  const settings = <View className="studio-settings-panel">{pickOption('文本模型',models,state.textModelId,id=>patch({textModelId:id}))}{pickOption('提取模板',presets,state.extractionPreset,id=>patch({extractionPreset:id}))}{pickOption('图片模型',imageModels,state.imageModelId,id=>patch({imageModelId:id}))}{pick('开头模式',OPENINGS,state.opening,opening=>patch({opening}))}{pick('输出模式',FORMATS,state.format,format=>patch({format}))}{pick('时长上限',DURATIONS,state.duration,duration=>patch({duration}))}</View>;
  return <View className="mp-page mp-script-page script-studio"><BrandHeader eyebrow="私人定制" title="私人定制" showBack />
    <View className="mp-script-console">
      <View className="studio-titlebar"><View><View className="studio-title"><AppIcon name="file-pen-line" size={18} /><Text>剧本创作</Text></View><Text className="studio-caption">从原文到角色，再到你的下一幕</Text></View><Button className="studio-history" aria-label="历史记录" disabled={Boolean(busy)} onClick={() => { setError(''); setShowHistory(true); }}><AppIcon name="folder-clock" size={19} /></Button></View>
      <View className="studio-tabs"><Button className={tab === '编辑' && editorStage === 'source' ? 'studio-tab is-active' : 'studio-tab'} disabled={Boolean(busy)} onClick={() => { setTab('编辑'); setEditorStage('source'); setSourceExpanded(true); }}>原文</Button><Button className={tab === '编辑' && editorStage === 'entities' ? 'studio-tab is-active' : 'studio-tab'} disabled={Boolean(busy)} onClick={() => { setTab('编辑'); setEditorStage('entities'); setSourceExpanded(false); }}>人物场景</Button><Button className={tab === '结果' ? 'studio-tab is-active' : 'studio-tab'} disabled={Boolean(busy) || !result} onClick={() => setTab('结果')}>创作结果{result ? <Text className="studio-result-dot" /> : null}</Button></View>
      {error ? <Text className="mp-error">{error}</Text> : null}{busy ? <Text className="studio-caption">正在{busy}…</Text> : null}
      {tab === '编辑' ? <View className="mp-script-editor">
        <View className="studio-paper"><View className="studio-paper-head"><Text>小说原文</Text>{editorStage === 'source' ? <Button className="studio-link" disabled={Boolean(busy)} onClick={() => { if (actionLock.current) return; const current=version.current; void Taro.getClipboardData().then(value=>{if(mounted.current && current===version.current && !actionLock.current)changeSource(value.data || '');}).catch(()=>setError('剪贴板读取失败，请直接长按粘贴')); }}>粘贴内容</Button> : null}</View>
          {sourceExpanded ? <Textarea maxlength={-1} className="studio-source-input" disabled={Boolean(busy)} value={state.novelText} onInput={event => changeSource(event.detail.value)} placeholder="粘贴小说原文" /> : <Text className="studio-source-excerpt">{state.novelText.trim() || '还没有原文，展开后开始输入。'}</Text>}
          <View className="studio-paper-foot"><Text>{state.novelText.length.toLocaleString()} 字</Text><Button className="studio-link" disabled={Boolean(busy)} onClick={() => setSourceExpanded(!sourceExpanded)}>{sourceExpanded ? '收起原文' : '展开编辑'}</Button></View></View>
        {catalogError ? <View className="studio-option-error"><Text className="mp-error">{catalogError}</Text><Button disabled={Boolean(busy)} onClick={()=>void loadCatalog()}>重试选项</Button></View> : null}
        {editorStage === 'source' ? <><View className="studio-section-head"><Text>提取设置</Text></View><View className="studio-source-options">{pickOption('文本模型',models,state.textModelId,id=>patch({textModelId:id}))}{pickOption('提取模板',presets,state.extractionPreset,id=>patch({extractionPreset:id}))}</View><Button className="studio-manual-entry" disabled={Boolean(busy)} onClick={openEntities}>手动添加人物场景</Button><Text className="studio-caption studio-source-hint">先提取或添加设定，确认后再生成剧本。</Text></> : <>
        <View className="studio-section-head"><Text>人物与场景 · {state.characters.length + state.scenes.length}</Text><Button className="studio-link" disabled={Boolean(busy)} onClick={openEntities}>{hasEntities ? '查看与编辑' : '手动添加人物场景'}</Button></View>
        {hasEntities ? <><ScrollView scrollX className="studio-entities"><View className="studio-entity-track">{[...state.characters.map(entity => ({entity,type:'character'})),...state.scenes.map(entity=>({entity,type:'scene'}))].map(({entity,type},index)=><Button key={entity.id || `${type}-${index}`} className="studio-entity" disabled={Boolean(busy)} onClick={openEntities}><View className={`studio-entity-icon ${type}`}><AppIcon name={type==='character'?'user-round':'camera'} size={24}/></View><Text>{entityName(entity,`${type==='character'?'人物':'场景'} ${index+1}`)}</Text></Button>)}</View></ScrollView><View className="studio-focus-list">{state.characters.map((entity,index)=>{const id=scriptEntityKey(entity,index);const focused=state.protagonists?.includes(id);return <Button key={id} className={focused?'studio-focus is-focused':'studio-focus'} disabled={Boolean(busy)} onClick={()=>patch({protagonists:focused?state.protagonists?.filter(item=>item!==id):[...(state.protagonists||[]),id]})}>{focused?'已聚焦 · ':'聚焦 · '}{entityName(entity,`人物 ${index+1}`)}</Button>;})}</View><Button className="studio-link studio-reextract" disabled={Boolean(busy)} onClick={extract}>重新提取</Button></> : <View className="studio-entity-empty"><AppIcon name="users-round" size={25}/><View><Text>还没有人物或场景</Text><Text className="studio-caption">可手动添加，或回到原文进行提取。</Text></View></View>}
        </>}
        <View className="studio-entry-list"><Button id="script-settings" className="studio-entry" disabled={Boolean(busy)} onClick={()=>setShowSettings(true)}><View><AppIcon name="settings" size={17}/><Text>创作设置</Text></View><View><Text className="studio-caption">{state.format} · {state.duration}</Text><AppIcon name="chevron-right" size={14}/></View></Button><Button className="studio-entry" disabled={Boolean(busy)} onClick={()=>setShowConstraints(true)}><View><AppIcon name="sliders" size={17}/><Text>约束与提示词</Text></View><View><Text className="studio-caption">已开启 {constraintCount} 项</Text><AppIcon name="chevron-right" size={14}/></View></Button><Button className="studio-entry" disabled={Boolean(busy)} onClick={()=>setShowAudio(true)}><View><AppIcon name="volume" size={17}/><Text>原文配音</Text></View><View><Text className="studio-caption">{state.audioTotalSeconds?`${state.audioTotalSeconds} 秒`:'未生成'} · 匹配{state.matchAudio?'开':'关'}</Text><AppIcon name="chevron-right" size={14}/></View></Button></View>
      </View> : <View className="mp-script-results"><View className="studio-result-header"><Text className="studio-result-title">创作结果</Text><Text className="studio-caption">{result?savedHistoryId?'已保存到历史记录':'历史尚未保存，请保留当前结果':'确认原文与人物场景后开始生成'}</Text></View><Text selectable className="studio-result-text">{result || '还没有生成结果'}</Text><Button className="studio-copy mp-secondary" disabled={!result} onClick={()=>Taro.setClipboardData({data:result})}>复制完整剧本</Button>{result?<ScriptVideoProduction key={`${savedHistoryId}:${result}`} output={result} historyId={savedHistoryId} tasks={state.videoTasks||{}} onTasksChange={tasks=>patch({videoTasks:tasks})} duration={state.duration} entities={[...state.characters,...state.scenes]}/>:null}</View>}
    </View>
    <View className="studio-dock"><View className="studio-dock-copy"><Text>{busy?`正在${busy}…`:tab==='结果'?'继续编辑你的作品':hasEntities?'人物场景已就绪':'从原文开始创作'}</Text><Text className="studio-caption">{tab==='结果'?'已有结果保留在当前页':readiness}</Text></View><Button id="script-main-action" className="studio-primary" loading={Boolean(busy)} disabled={tab==='结果'?Boolean(busy):actionDisabled} onClick={()=>{if(tab==='结果'){setTab('编辑');setEditorStage('entities');}else if(hasEntities && editorStage==='source'){setEditorStage('entities');setSourceExpanded(false);}else if(hasEntities)return generate();else return extract();}}>{tab==='结果'?'返回编辑':hasEntities && editorStage==='source'?'下一步':hasEntities?'生成剧本':'提取人物场景'}</Button></View>
    {editingEntities || showSettings || showHistory || showConstraints || showAudio ? <View className="studio-overlay"><View className="studio-backdrop" onClick={()=>{if(!busy)closePanel();}} onTouchMove={event=>event.stopPropagation()}/><View className={editingEntities ? "studio-sheet is-assets" : "studio-sheet"}><View className="studio-sheet-handle"/><View className="studio-sheet-header"><Text>{editingEntities?'人物与场景':showHistory?'剧本历史':showSettings?'创作设置':showConstraints?'约束与提示词':'原文配音'}</Text><Button className="studio-link" disabled={Boolean(busy)} onClick={closePanel}>{editingEntities?'取消修改':'完成'}</Button></View><ScrollView scrollY className="studio-sheet-scroll"><View className="studio-sheet-body">
      {editingEntities?<><ScriptEntityEditor value={entityDraft} onChange={setEntityDraft} model={state.imageModelId||''} models={imageModels} onModelChange={imageModelId=>patch({imageModelId})} novel={state.novelText} preset={state.extractionPreset||''}/></>:null}
      {showSettings?settings:null}
      {showHistory?<ScriptHistoryPanel disabled={Boolean(busy)} onSelect={id=>void chooseHistory(id)}/>:null}
      {showConstraints?<><Text className="mp-field-label">统一视觉风格</Text><Input className="mp-input" value={state.visualStyle||''} onInput={event=>patch({visualStyle:event.detail.value})}/>{LAYERS.map(([key,label])=>{const layer=state.constraints?.[key]||{};const update=(next:object)=>patch({constraints:updateScriptConstraint(stateRef.current.constraints||{},key,next)});const choices=scriptConstraintChoices(catalog,key);return <View className="studio-constraint-layer" key={key}><Button className="studio-layer-toggle" onClick={()=>update({enabled:!layer.enabled})}><Text>{label}</Text><Text className={layer.enabled?'studio-toggle is-on':'studio-toggle'}>{layer.enabled?'已开启':'已关闭'}</Text></Button>{pickOption('系统模板',choices,layer.presetId,id=>update({source:'system',presetId:id,enabled:true}))}<Textarea maxlength={-1} className="mp-textarea" value={layer.body||''} placeholder="填写本次约束；填写后以自定义内容生成" onInput={event=>update({body:event.detail.value,source:'draft',presetId:'',enabled:Boolean(event.detail.value.trim())})}/>{key!=='baseSetup'?<><Button className="studio-link" onClick={()=>setPersonalCategory(personalCategory===key?'':key)}>我的{label}提示词</Button>{personalCategory===key?<ScriptPersonalPrompts key={key} category={key} onSelect={prompt=>{update({source:'personal',personalPromptId:prompt.id,presetId:'',body:prompt.body,enabled:true});setPersonalCategory('');}}/>:null}</>:null}</View>;})}</>:null}
      {showAudio?<View className="studio-audio"><AppIcon name="volume" size={30}/><Text className="studio-result-title">原文配音</Text><Text className="studio-caption">{state.audioTotalSeconds?`${state.audioTotalSeconds} 秒${audio.current?'':' · 历史时长，需重新生成才能播放'}`:'生成后可播放原文配音，并用真实时长匹配分镜。'}</Text><Button className="studio-primary studio-apply" disabled={Boolean(busy)||!state.novelText.trim()} loading={busy==='生成原文配音'} onClick={generateVoice}>生成原文配音</Button>{audio.current?<Button className="mp-secondary" onClick={()=>{if(audioPlaying)audio.current?.pause();else audio.current?.play();setAudioPlaying(!audioPlaying);}}>{audioPlaying?'暂停':'播放配音'}</Button>:null}<Button className="studio-layer-toggle" disabled={Boolean(busy)||!state.audioTotalSeconds} onClick={()=>state.audioTotalSeconds&&patch({matchAudio:!state.matchAudio})}><Text>匹配音频</Text><Text className={state.matchAudio?'studio-toggle is-on':'studio-toggle'}>{state.matchAudio?'已开启':'已关闭'}</Text></Button></View>:null}
      {error?<Text className="mp-error">{error}</Text>:null}
    </View></ScrollView>{editingEntities?<View className="studio-sheet-footer"><Button className="studio-primary" onClick={applyAssets}>应用修改</Button></View>:null}</View></View>:null}
  </View>;
}
export default withThemedPage(ScriptWorkspace);
