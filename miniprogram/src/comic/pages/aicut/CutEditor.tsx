import Taro, { useDidHide } from '@tarojs/taro';
import { Audio, Button, Image, Input, Picker, ScrollView, Text, Textarea, Video, View } from '@tarojs/components';
import { useEffect, useRef, useState } from 'react';
import AppIcon from '../../../components/AppIcon';
import BrandHeader from '../../../components/BrandHeader';
import { apiRequest } from '../../../services/api';
import { createAICutClient, type CutRecord, type CutTimeline } from '../../../services/aicut';
import { addAsset, basicEditReason, editClip, removeClip } from '../../../services/aicutEditing';
import { importCutMedia } from '../../../services/aicutMedia';
import { loadAuthenticatedMedia } from '../../../services/mediaCache';
import CutExports from './CutExports';

const freshID = () => `mini-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 9)}`;
const seconds = (ms: number) => `${(ms / 1000).toFixed(1)}s`;
export default function CutEditor({ project, onBack,initialTimelineId='default',initialPanel='' }: { project: CutRecord;initialTimelineId?:string;initialPanel?:string; onBack: () => void }) {
  const [timeline, setTimeline] = useState<CutTimeline | null>(null), [timelineId, setTimelineId] = useState(initialTimelineId);
  const [timelines, setTimelines] = useState<CutRecord[]>([]), [assets, setAssets] = useState<CutRecord[]>([]);
  const [dirty, setDirty] = useState(false), [busy, setBusy] = useState(''), [error, setError] = useState('');
  const [panel, setPanel] = useState(''), [selection, setSelection] = useState<{ track: string; clip: string } | null>(null);
  const [draft, setDraft] = useState<CutRecord>({}), [subtitle, setSubtitle] = useState('');
  const [preview, setPreview] = useState<{ kind: string; path: string } | null>(null);
  const alive = useRef(true), lock = useRef(false), epoch = useRef(0);
  const state = useRef({ timeline, timelineId, dirty }); state.current = { timeline, timelineId, dirty };
  const client = createAICutClient(apiRequest);
  const run = async (label: string, work: () => Promise<void>) => {
    if (lock.current) return; lock.current = true; const ticket = epoch.current; setBusy(label); setError('');
    try { await work(); } catch (cause: any) { if (alive.current && ticket === epoch.current) setError(cause?.statusCode === 409 ? '网页或其他设备已修改此工程。本地修改仍保留，请核对后重新读取；没有自动覆盖。' : cause?.message || cause?.errMsg || '操作失败，请重试'); }
    finally { lock.current = false; if (alive.current && ticket === epoch.current) setBusy(''); }
  };
  const load = async (id: string) => {
    const ticket = ++epoch.current;
    const [value, media, metadata] = await Promise.all([client.timeline(project.id, id), client.assets(project.id), client.timelines(project.id)]);
    if (!alive.current || ticket !== epoch.current) return;
    setTimeline(value); setAssets(media); setTimelines(metadata); setTimelineId(id); setDirty(false); setSelection(null); setPanel(initialPanel==='exports'?'exports':''); setPreview(null);
  };
  // Loading owns its busy cleanup because it increments the project read version.
  const reload = async (id = state.current.timelineId) => {
    if (lock.current) return; lock.current = true; setBusy('读取工程'); setError('');
    try { await load(id); } catch (cause: any) { if (alive.current) setError(cause.message || '工程读取失败'); }
    finally { lock.current = false; if (alive.current) setBusy(''); }
  };
  useEffect(() => { void reload(); return () => { alive.current = false; epoch.current++; void Taro.disableAlertBeforeUnload?.(); }; }, []);
  useEffect(() => { if (dirty) void Taro.enableAlertBeforeUnload?.({ message: '此工程还有未保存的剪辑修改' }); else void Taro.disableAlertBeforeUnload?.(); }, [dirty]);
  useDidHide(() => setPreview(null));
  const discard = async () => !state.current.dirty || (await Taro.showModal({ title: '还有未保存的修改', content: '继续将放弃本地修改，已保存的工程不受影响。', confirmText: '放弃修改', cancelText: '继续编辑' })).confirm;
  const change = (value: CutTimeline) => { setTimeline(value); setDirty(true); setPreview(null); };
  const track = timeline?.tracks.find(item => item.id === selection?.track);
  const clip = track?.clips.find((item: CutRecord) => item.id === selection?.clip);
  const asset = assets.find(item => item.id === clip?.assetId);
  const selectedReason = track?.locked ? '该轨道已锁定，请在网页解锁' : clip ? basicEditReason(clip) : '';
  const save = () => run('保存工程', async () => {
    if (!state.current.timeline) return;
    const ticket = epoch.current;
    const value = await client.save(project.id, state.current.timelineId, state.current.timeline);
    if (alive.current && ticket === epoch.current) { setTimeline(value); setDirty(false); await Taro.showToast({ title: '工程已保存', icon: 'success' }); }
  });
  const apply = () => {
    if (!timeline || !selection || lock.current) return;
    try { change(editClip(timeline, selection.track, selection.clip, { name: draft.name, ...(clip.kind === 'subtitle' ? { text: draft.text } : {}), startMs: Math.round(Number(draft.start) * 1000), sourceStartMs: Math.round(Number(draft.source) * 1000), durationMs: Math.round(Number(draft.duration) * 1000) })); setPanel(''); }
    catch (cause: any) { setError(cause.message); }
  };
  const select = (trackId: string, item: CutRecord) => { setSelection({ track: trackId, clip: item.id }); setPreview(null); setDraft({ name: item.name || '', text: item.text || '', start: String(item.startMs / 1000), source: String(item.sourceStartMs / 1000), duration: String(item.durationMs / 1000) }); };
  const previewAsset = () => run('读取素材预览', async () => {
    if (!asset?.url) throw new Error('此片段没有可预览的素材，请通过导出查看合成结果');
    const ticket = epoch.current; const selected = clip.id;
    const path = await loadAuthenticatedMedia(asset.proxyUrl || asset.url, asset.kind === 'video' ? 'mp4' : asset.kind === 'audio' ? 'mp3' : /png/i.test(asset.contentType || '') ? 'png' : 'jpg');
    if (alive.current && ticket === epoch.current && selected === clip.id) setPreview({ kind: asset.kind, path: path.path });
  });
  const duration = timeline?.tracks.reduce((max, item) => Math.max(max, ...item.clips.map((entry: CutRecord) => entry.startMs + entry.durationMs)), 0) || 0;
  return <>
    <BrandHeader title="AI-CUT" showBack onBack={() => { if (!lock.current) void discard().then(confirmed => { if (confirmed && alive.current) onBack(); }); }} />
    <View className="cut-editor-header"><View><Text className="cut-editor-name">{project.title}</Text><Text className="cut-caption">{dirty ? '有未保存的修改' : timeline ? '已与工程同步' : '正在读取工程'}</Text></View><Button className="cut-primary" disabled={Boolean(busy) || !timeline || dirty} onClick={() => setPanel('exports')}>导出</Button></View>
    <View className="cut-content cut-editor-content">
      <View className="cut-preview">
        {preview?.kind === 'video' ? <Video id="cut-preview-video" src={preview.path} controls /> : preview?.kind === 'image' ? <Image src={preview.path} mode="aspectFit" /> : preview?.kind === 'audio' ? <Audio src={preview.path} controls /> : <View className="cut-empty"><AppIcon name="clapperboard" size={32} /><Text>{clip?.kind === 'subtitle' ? clip.text : '选中片段，预览原始素材'}</Text><Text className="cut-caption">完整合成效果可在导出后查看</Text></View>}
      </View>
      <View className="cut-heading cut-small"><Text>{project.aspectRatio} · {seconds(duration)}</Text>{asset ? <Button className="cut-link" disabled={Boolean(busy)} onClick={() => void previewAsset()}>预览此素材</Button> : null}</View>
      {error ? <View className="cut-error"><Text>{error}</Text><Button disabled={Boolean(busy)} onClick={async () => { if (await discard()) void reload(); }}>重新读取工程</Button></View> : null}
      {busy ? <Text className="cut-caption">正在{busy}…</Text> : null}
      <View className="cut-heading cut-small"><Text>时间线</Text><Picker disabled={Boolean(busy)} range={timelines.map(item => item.title)} value={Math.max(0, timelines.findIndex(item => item.id === timelineId))} onChange={async e => { const item = timelines[Number(e.detail.value)]; if (item && item.id !== timelineId && await discard()) void reload(item.id); }}><View className="cut-timeline-choice">{timelines.find(item => item.id === timelineId)?.title || '主时间线'} ⌄</View></Picker></View>
      <ScrollView scrollX className="cut-timeline"><View style={{ minWidth: `${Math.max(320, Math.min(2600, duration / 1000 * 24 + 90))}px` }}>
        {timeline?.tracks.map(item => <View className="cut-track" key={item.id}><Text className="cut-track-label">{item.name || item.kind}{item.locked ? ' · 锁定' : ''}</Text><View className="cut-track-lane">{item.clips.map((entry: CutRecord) => <Button key={entry.id} disabled={Boolean(busy)} className={`cut-clip kind-${item.kind}${selection?.clip === entry.id ? ' is-selected' : ''}`} style={{ width: `${Math.max(.3, entry.durationMs / Math.max(duration, 1) * 100)}%`, left: `${entry.startMs / Math.max(duration, 1) * 100}%` }} onClick={() => select(item.id, entry)}><Text>{entry.name || entry.text || '片段'}</Text><Text>{seconds(entry.startMs)} · {seconds(entry.durationMs)}</Text></Button>)}</View></View>)}
        {timeline && !timeline.tracks.some(item => item.clips.length) ? <View className="cut-timeline-empty"><Text>时间线还是空的</Text><Text className="cut-caption">从素材库添加第一个片段</Text></View> : null}
      </View></ScrollView>
      {clip ? <View className="cut-selected"><View className="cut-heading cut-small"><Text className="cut-clip-title">{clip.name || '已选片段'}</Text><Text className="cut-caption">{seconds(clip.durationMs)}</Text></View>{selectedReason ? <Text className="cut-caption">{selectedReason}</Text> : <View className="cut-actions"><Button disabled={Boolean(busy)} onClick={() => setPanel('edit')}>裁剪与编辑</Button><Button disabled={Boolean(busy)} onClick={() => void run('删除片段', async () => { if ((await Taro.showModal({ title: '移除这个片段？', content: '仅从时间线移除，素材文件会保留。', confirmText: '移除' })).confirm && timeline && selection) { change(removeClip(timeline, selection.track, selection.clip)); setSelection(null); } })}>移除片段</Button></View>}</View> : null}
      <Text className="cut-caption cut-editor-tip">手机提供基础剪辑。原工程的复杂效果会保留，需在网页调整。</Text>
    </View>
    <View className="cut-dock"><Button disabled={Boolean(busy) || !timeline} onClick={() => setPanel('assets')}><AppIcon name="plus" size={20} /><Text>素材</Text></Button><Button disabled={Boolean(busy) || !timeline} onClick={() => { setSubtitle(''); setPanel('subtitle'); }}><AppIcon name="file-pen-line" size={20} /><Text>字幕</Text></Button><Button id="cut-save" className="cut-primary" disabled={Boolean(busy) || !dirty} onClick={() => void save()}>{busy === '保存工程' ? '保存中' : '保存工程'}</Button></View>
    {panel ? <View className="cut-overlay"><View className="cut-mask" onClick={() => { if (!busy) setPanel(''); }} /><View className="cut-sheet"><View className="cut-handle" /><View className="cut-heading"><Text className="cut-title">{({ assets: '添加素材', subtitle: '添加字幕', edit: '片段编辑', exports: '导出视频' } as CutRecord)[panel]}</Text><Button className="cut-link" disabled={Boolean(busy)} onClick={() => setPanel('')}>关闭</Button></View><ScrollView scrollY className="cut-sheet-scroll">
      {panel === 'assets' ? <>
        <View className="cut-actions"><Button disabled={Boolean(busy)} onClick={() => void run('从相册导入', async () => { await importCutMedia(project.id); const media = await client.assets(project.id); if (alive.current) setAssets(media); })}>相册或拍摄</Button><Button disabled={Boolean(busy)} onClick={() => void run('刷新素材', async () => { const media = await client.assets(project.id); if (alive.current) setAssets(media); })}>刷新素材库</Button></View>
        {!assets.length ? <Text className="cut-caption">还没有素材，可从相册选择图片或视频。</Text> : assets.map(item => <Button className="cut-asset-row" key={item.id} disabled={Boolean(busy) || !['video', 'image', 'audio'].includes(item.kind)} onClick={() => { if (!timeline) return; try { change(addAsset(timeline, item, freshID())); setPanel(''); } catch (cause: any) { setError(cause.message); } }}><AppIcon name={item.kind === 'audio' ? 'volume' : 'clapperboard'} size={22} /><View><Text>{item.name}</Text><Text className="cut-caption">{item.kind} · {item.durationMs ? seconds(item.durationMs) : '图片 / 时长未返回'}</Text></View><Text>＋</Text></Button>)}
      </> : null}
      {panel === 'subtitle' ? <><Textarea className="cut-input cut-textarea" maxlength={2000} value={subtitle} onInput={e => setSubtitle(e.detail.value)} placeholder="写下这一段的字幕" /><Text className="cut-caption">默认添加到字幕轨道末尾，时长3秒，之后可修改。</Text><Button className="cut-primary" disabled={!subtitle.trim() || Boolean(busy)} onClick={() => { if (timeline) { change(addAsset(timeline, null, freshID(), subtitle.trim())); setPanel(''); } }}>添加到时间线</Button></> : null}
      {panel === 'edit' && clip ? <><Text className="cut-caption">片段名称</Text><Input className="cut-input" value={draft.name} onInput={e => setDraft({ ...draft, name: e.detail.value })} />{clip.kind === 'subtitle' ? <Textarea className="cut-input cut-textarea" value={draft.text} maxlength={2000} onInput={e => setDraft({ ...draft, text: e.detail.value })} /> : null}{[['start', '在时间线开始于（秒）'], ...(clip.kind === 'video' || clip.kind === 'audio' ? [['source', '从原素材第几秒开始']] : []), ['duration', '片段时长（秒）']].map(([key, label]) => <View key={key}><Text className="cut-caption">{label}</Text><Input type="digit" className="cut-input" value={draft[key]} onInput={e => setDraft({ ...draft, [key]: e.detail.value })} /></View>)}<Button className="cut-primary" onClick={apply}>应用修改</Button><Text className="cut-caption">应用后还需保存工程，才能同步到网页。</Text></> : null}
      {panel === 'exports' && timeline ? <CutExports projectId={project.id} timelineId={timelineId} revision={timeline.revision} canExport={timeline.tracks.some(item => item.clips.length)} /> : null}
      {error ? <Text className="cut-error">{error}</Text> : null}
    </ScrollView></View></View> : null}
  </>;
}
