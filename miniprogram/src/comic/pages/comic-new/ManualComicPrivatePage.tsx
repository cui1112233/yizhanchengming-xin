import Taro, { useDidShow, useDidHide } from '@tarojs/taro';
import { Button, Image, Text, View } from '@tarojs/components';
import { useEffect, useRef, useState } from 'react';
import BrandHeader from '../../../components/BrandHeader';
import { ThemedPage } from '../../../theme';
import ComicSettingsStage from './ComicSettingsStage';
import ComicAssetsStage from './ComicAssetsStage';
import ComicConfigConfirmStage from './ComicConfigConfirmStage';
import ComicPageWorkspace from './ComicPageWorkspace';
import ComicStageHeader from './ComicStageHeader';
import { applyComicSettingSelections, buildComicInitialBatchSettings, comicPresetAvailability, comicPresetComicReason, comicSourceLength, createComicBatch, DEFAULT_COMIC_ASPECT_RATIO, loadComicPrivateSetup, MIN_COMIC_SOURCE_CHARS, resolveComicStyleThumb, uploadComicReference, type ComicSettingSelections, type ComicReferenceInfo } from '../../services/comicFlow';
import { extractComicAssets, generateComicAssetImages, loadComicAssets, loadComicAssetImages, loadComicAssetImagePath, readComicAssetImageTask, retryComicBookStage, type ComicAssetItem, type ComicAssetImageItem } from '../../services/comicAssets';
import { createComicCollection, downloadComicCollectionZip, downloadComicPagePackage, loadComicImagePaths, loadComicPages, regenerateComicPage, saveComicImageToAlbum, setPrimaryComicPageImage, type ComicPageItem } from '../../services/comicResult';
import { comicPromptMode, generateComicPrompts, saveComicManualSettings, watchComicPageImages } from '../../services/comicManual';
import { clearComicActiveRun, readComicActiveRun, saveComicActiveRun, type ComicActiveRun } from '../../services/comicRecovery';
import type { ComicStage } from './stageState';
import '../../styles/comic.scss';
import './editor.scss';
type Setup = Awaited<ReturnType<typeof loadComicPrivateSetup>>;
type Project = Omit<ComicActiveRun, 'version'> & {
    mode: 'manual';
    phase: 'assets' | 'pages';
    pendingImages: Record<string, string[]>;
};
type State = {
    source: string;
    stage: ComicStage;
    project: Project | null;
    setup: Setup | null;
    setupLoading: boolean;
    setupError: string;
    styleId: string;
    aspectRatio: string;
    referencePath: string;
    referenceOpen: boolean;
    settingPicker: '' | 'prompt' | 'derivation' | 'model' | 'constraints';
    selections: ComicSettingSelections;
    thumbs: Record<string, string>;
    thumbFailed: Record<string, boolean>;
    assets: ComicAssetItem[];
    assetImages: Record<string, ComicAssetImageItem[]>;
    assetThumbs: Record<string, string>;
    pages: ComicPageItem[];
    paths: Record<string, string>;
    pageIndex: number;
    busy: string;
    error: string;
};
function initial(initialSource = ""): State {
    const saved = readComicActiveRun();
    const project = saved && (!saved.submitted || saved.mode === 'manual') ? { ...saved, mode: 'manual' as const, phase: saved.phase || 'assets' as const, pendingImages: saved.pendingImages || {} } : null;
    return { source: project?.source || initialSource, stage: project?.phase === 'pages' ? 'render' : project ? 'assets' : 'source', project, setup: null, setupLoading: true, setupError: '', styleId: project?.styleId || '', aspectRatio: project?.aspectRatio || DEFAULT_COMIC_ASPECT_RATIO, referencePath: '', referenceOpen: false, settingPicker: '', selections: {}, thumbs: {}, thumbFailed: {}, assets: [], assetImages: {}, assetThumbs: {}, pages: [], paths: {}, pageIndex: 0, busy: '', error: '' };
}
export default function ManualComicPrivatePage({ initialSource = "" }: {
    initialSource?: string;
} = {}) {
    const [s, set] = useState<State>(() => initial(initialSource));
    const latest = useRef(s);
    latest.current = s;
    const project = useRef(s.project);
    const alive = useRef(true);
    const visible = useRef(true);
    const epoch = useRef(0);
    const lock = useRef(false);
    const watcher = useRef<(() => void)>();
    const assetsRead = useRef(0);
    const pagesRead = useRef(0);
    const imageReads = useRef(0);
    const patch = (value: Partial<State>) => { if (alive.current)
        set(prev => ({ ...prev, ...value })); };
    const current = (version: number, p: Project) => alive.current && epoch.current === version && project.current?.batchId === p.batchId && project.current?.bookId === p.bookId;
    const remember = (p: Project) => { project.current = p; saveComicActiveRun(p); patch({ project: p }); };
    const acceptPages = async (p: Project, pages: ComicPageItem[], version = epoch.current) => {
        if (!current(version, p))
            return;
        const read = ++imageReads.current;
        patch({ pages, pageIndex: Math.min(latest.current.pageIndex, Math.max(0, pages.length - 1)) });
        const paths: Record<string, string> = {};
        for (let offset = 0; offset < pages.length; offset += 3) {
            if (!current(version, p))
                return;
            const loaded = await Promise.all(pages.slice(offset, offset + 3).map(page => loadComicImagePaths(page).catch(() => ({}))));
            Object.assign(paths, ...loaded);
        }
        if (current(version, p) && read === imageReads.current)
            patch({ paths });
    };
    const refreshPages = async (p = project.current) => { if (!p)
        return; const version = epoch.current; const read = ++pagesRead.current; const pages = await loadComicPages(p.batchId, p.bookId); if (current(version, p) && read === pagesRead.current)
        await acceptPages(p, pages, version); return pages; };
    const refreshAssets = async (p = project.current) => {
        if (!p)
            return;
        const version = epoch.current;
        const read = ++assetsRead.current;
        const assets = await loadComicAssets(p.batchId, p.bookId);
        const entries = await Promise.all(assets.map(async (asset) => [asset.id, await loadComicAssetImages(p.batchId, p.bookId, asset.id)] as const));
        const images = Object.fromEntries(entries);
        const thumbs: Record<string, string> = {};
        await Promise.all(assets.map(async (asset) => { const main = images[asset.id]?.find(image => image.isPrimary) || images[asset.id]?.[0]; if (main)
            try {
                thumbs[asset.id] = await loadComicAssetImagePath(main);
            }
            catch { /* the version chooser exposes failed media */ } }));
        if (current(version, p) && read === assetsRead.current)
            patch({ assets, assetImages: images, assetThumbs: thumbs });
    };
    const beginWatch = (p = project.current) => {
        if (!p || !Object.keys(p.pendingImages).length || !visible.current)
            return;
        watcher.current?.();
        const version = epoch.current;
        patch({ busy: 'images' });
        let stopped = false;
        const stop = watchComicPageImages(p.batchId, p.bookId, p.pendingImages, {
            onPages: pages => { if (current(version, p))
                void acceptPages(p, pages, version); },
            onDone: () => { stopped = true; watcher.current = undefined; if (current(version, p)) {
                remember({ ...project.current!, pendingImages: {} });
                patch({ busy: '' });
            } },
            onError: error => { stopped = true; watcher.current = undefined; if (current(version, p))
                patch({ busy: '', error: error.message }); },
        });
        watcher.current = stopped ? undefined : stop;
    };
    const perform = async (label: string, action: () => Promise<void>) => {
        if (lock.current || watcher.current)
            return;
        lock.current = true;
        const version = epoch.current;
        patch({ busy: label, error: '' });
        try {
            await action();
        }
        catch (cause: any) {
            if (alive.current && epoch.current === version)
                patch({ error: cause?.message || cause?.errMsg || '操作失败，请稍后重试' });
        }
        finally {
            lock.current = false;
            if (alive.current && epoch.current === version && !watcher.current)
                patch({ busy: '' });
        }
    };
    useEffect(() => {
        let active = true;
        void loadComicPrivateSetup().then(setup => {
            if (!active)
                return;
            set(prev => ({ ...prev, setup, setupLoading: false, styleId: prev.styleId || setup.styles[0]?.id || '' }));
            setup.styles.forEach(style => { if (style.previewUrl)
                void resolveComicStyleThumb(style.previewUrl).then(path => { if (active)
                    set(prev => ({ ...prev, thumbs: { ...prev.thumbs, ...(path ? { [style.id]: path } : {}) }, thumbFailed: { ...prev.thumbFailed, [style.id]: !path } })); }).catch(() => { if (active)
                    set(prev => ({ ...prev, thumbFailed: { ...prev.thumbFailed, [style.id]: true } })); }); });
        }).catch(cause => { if (active)
            patch({ setupLoading: false, setupError: cause?.message || '创作配置读取失败' }); });
        return () => { active = false; alive.current = false; epoch.current++; watcher.current?.(); };
    }, []);
    useEffect(() => { if (s.project) {
        const restoredProject=s.project; const restoredVersion=epoch.current;
        void refreshAssets().catch(cause => {if(current(restoredVersion,restoredProject))patch({error:cause.message});});
        if (s.project.phase === 'pages')
            void refreshPages().catch(cause => {if(current(restoredVersion,restoredProject))patch({error:cause.message});});
        if (Object.keys(s.project.pendingImages).length)
            beginWatch();
    } }, []);
    useDidHide(() => { visible.current = false; watcher.current?.(); watcher.current = undefined; if (!lock.current)
        patch({ busy: '' }); });
    useDidShow(() => {
        visible.current = true;
        const chosen = Taro.getStorageSync('comic:selected-existing-book');
        if (chosen && typeof chosen === 'object' && typeof chosen.sourceText === 'string') {
            Taro.removeStorageSync('comic:selected-existing-book');
            patch({ source: chosen.sourceText, stage: 'source' });
            return;
        }
        if (lock.current || !project.current)
            return;
        const shownProject=project.current; const shownVersion=epoch.current;
        void refreshAssets().catch(cause => {if(current(shownVersion,shownProject))patch({error:cause.message});});
        if (project.current.phase === 'pages')
            void refreshPages().catch(cause => {if(current(shownVersion,shownProject))patch({error:cause.message});});
        if (Object.keys(project.current.pendingImages).length)
            beginWatch();
    });
    const setup = s.setup;
    const preset = setup?.presets.find(option => option.usable);
    const effective = preset ? { ...preset, preset: { ...preset.preset, config: applyComicSettingSelections(preset.preset, setup!.settingCatalog, setup!.models, s.selections).config } } : undefined;
    const configHint = s.setupLoading ? '正在读取配置…' : s.setupError || (!setup?.styles.length ? '还没有可用画风' : comicPresetAvailability(setup.presets) || comicPresetComicReason(effective?.preset));
    const blocked = Boolean(s.busy);
    const sourceReason = configHint || (!s.styleId ? '请选择画风' : comicSourceLength(s.source) < MIN_COMIC_SOURCE_CHARS ? `正文至少需要 ${MIN_COMIC_SOURCE_CHARS} 字` : '');
    const extract = () => perform('extract', async () => {
        const input = latest.current;
        if (!effective || sourceReason)
            throw new Error(sourceReason || '请选择创作配置');
        let p = project.current;
        const version = epoch.current;
        if (!p || p.source.trim() !== input.source.trim()) {
            const created = await createComicBatch({ batchTitle: '', bookTitle: '', sourceText: input.source, initialBatchSettings: { ...buildComicInitialBatchSettings({ aspectRatio: input.aspectRatio, stylePresetId: input.styleId, reference: null }, effective.preset), assetImagesEnabled: true } });
            if (!alive.current || epoch.current !== version)
                return;
            p = { batchId: created.batchId, bookId: created.bookId, source: input.source, sourceLength: comicSourceLength(input.source), styleId: input.styleId, aspectRatio: input.aspectRatio, presetId: effective.preset.id, presetConfig: JSON.parse(JSON.stringify(effective.preset.config || {})), reference: null, submitted: false, startedAt: Date.now(), mode: 'manual', phase: 'assets', pendingImages: {} };
            remember(p);
            patch({ assets: [], assetImages: {}, assetThumbs: {}, pages: [], paths: {}, pageIndex: 0, stage: 'assets' });
        }
        else
            patch({ stage: 'assets' });
        if (input.referencePath && !p.reference) {
            const reference = await uploadComicReference(p.batchId, input.referencePath);
            if (!current(version, p))
                return;
            const settings = buildComicInitialBatchSettings({ aspectRatio: p.aspectRatio, stylePresetId: p.styleId, reference }, { id: p.presetId, config: p.presetConfig } as any);
            await saveComicManualSettings(p.batchId, { ...settings, assetImagesEnabled: true });
            if (!current(version, p))
                return;
            p = { ...p, reference };
            remember(p);
        }
        await extractComicAssets({ batchId: p.batchId, bookId: p.bookId, mode: input.assets.length && input.project?.batchId === p.batchId ? 'force' : 'missing', textModelId: String(p.presetConfig.textModelId || '') || undefined });
        if (current(version, p))
            await refreshAssets(p);
    });
    const prompts = () => perform('prompts', async () => {
        const p = project.current;
        if (!p)
            throw new Error('请先提取人物场景');
        const version = epoch.current;
        const pages = await loadComicPages(p.batchId, p.bookId);
        if (!current(version, p))
            return;
        const mode = comicPromptMode(pages);
        if (mode === 'force') {
            const answer = await Taro.showModal({ title: '重新生成漫画提示词？', content: '将重新生成提示词，可能改变分页；此步骤调用文本模型，不会自动生成图片。', confirmText: '重新生成', cancelText: '先不生成' });
            if (!answer.confirm || !current(version, p))
                return;
        }
        await generateComicPrompts(p.batchId, p.bookId, mode);
        if (!current(version, p))
            return;
        remember({ ...p, phase: 'pages', submitted: true });
        patch({ stage: 'render' });
        const result = await refreshPages(p);
        if (current(version, p) && !result?.length)
            patch({ error: '暂未读到漫画提示词，请刷新检查；没有自动生成图片' });
    });
    const hasAssetImage = Object.values(s.assetImages).some(images => images.length > 0);
    const generateImages = (onlyVideoId?: string) => perform('images', async () => {
        const p = project.current;
        if (!p)
            return;
        const version = epoch.current;
        if (!Object.values(latest.current.assetImages).some(images => images.length))
            throw new Error('请先生成或上传至少一张资产图，再生成漫画图片');
        const pages = await loadComicPages(p.batchId, p.bookId);
        if (!current(version, p))
            return;
        const targets = pages.filter(page => onlyVideoId ? page.videoId === onlyVideoId : !page.images.length).filter(page => Boolean(page.imagePrompt?.trim()));
        if (!targets.length)
            throw new Error('没有可生成的页，请先生成或补齐图片提示词');
        if (onlyVideoId) {
            const answer = await Taro.showModal({ title: '生成这一页图片？', content: '只处理当前页，会使用出图额度；重新生成会新增候选图，原图保留。', confirmText: '生成本页', cancelText: '先不生成' });
            if (!answer.confirm || !current(version, p))
                return;
        }
        const modelId = String(p.presetConfig.imageModelId || '');
        if (!modelId)
            throw new Error('当前项目没有图片模型');
        const baseline: Record<string, string[]> = {};
        const failures: string[] = [];
        for (const page of targets) {
            if (!current(version, p))
                return;
            baseline[page.videoId] = page.images.map(image => image.id);
            remember({ ...project.current!, phase: 'pages', pendingImages: { ...baseline } });
            try {
                await regenerateComicPage(p.batchId, p.bookId, page.videoId, { comicPrompt: page.comicPrompt, imagePrompt: page.imagePrompt, modelId, imageProvider: String(p.presetConfig.imageProvider || '') || undefined, aspectRatio: page.aspectRatio || p.aspectRatio });
            }
            catch (cause: any) {
                if (Number(cause?.statusCode) !== 409) {
                    failures.push(`第${page.ordinal}页：${cause?.message || '提交失败'}`);
                    delete baseline[page.videoId];
                }
                if ([401, 403].includes(Number(cause?.statusCode)))
                    break;
            }
        }
        if (!current(version, p))
            return;
        remember({ ...project.current!, pendingImages: { ...baseline }, phase: 'pages' });
        if (failures.length)
            patch({ error: failures.join('；') });
        if (Object.keys(baseline).length)
            beginWatch();
        else
            await refreshPages(p);
    });
    const assetImages = (onlyId?: string) => perform('assetImages', async () => {
        const p = project.current;
        if (!p)
            return;
        const version = epoch.current;
        const input = latest.current;
        const ids = input.assets.filter(asset => onlyId ? asset.id === onlyId : !input.assetImages[asset.id]?.length).map(asset => asset.id);
        if (!ids.length)
            return;
        const result = await generateComicAssetImages({ batchId: p.batchId, bookId: p.bookId, assetIds: ids, modelId: String(p.presetConfig.imageModelId || '') || undefined, imageAspectRatio: '9:16' });
        if (result.taskId && !result.images.length) {
            for (let attempt = 0; attempt < 40; attempt++) {
                if (!current(version, p) || !visible.current)
                    return;
                await new Promise(resolve => setTimeout(resolve, 3000));
                const status: any = await readComicAssetImageTask(p.batchId, p.bookId, result.taskId);
                if (status.status === 'failed')
                    throw new Error(status.error || '资产图生成失败');
                if (!['pending', 'processing'].includes(status.status))
                    break;
            }
        }
        if (current(version, p))
            await refreshAssets(p);
    });
    const exportPages = () => perform('export', async () => {
        const p = project.current;
        if (!p || !latest.current.pages.length)
            return;
        let tapIndex: number;
        try {
            ({ tapIndex } = await Taro.showActionSheet({ itemList: ['导出提示词与图片 ZIP', '导出全部漫画图片', '导出抖音图文包'] }));
        }
        catch {
            return;
        }
        let result: {
            filePath?: string;
            url?: string;
        };
        if (tapIndex === 0)
            result = await downloadComicPagePackage(p.batchId, p.bookId, 'prompts');
        else if (tapIndex === 2)
            result = await downloadComicPagePackage(p.batchId, p.bookId, 'douyin');
        else {
            const ids = latest.current.pages.filter(page => page.primaryImageId && page.images.some(image => image.id === page.primaryImageId)).map(page => page.videoId);
            if (!ids.length)
                throw new Error('还没有漫画图片，可以先导出提示词 ZIP');
            const collection = await createComicCollection(p.batchId, p.bookId, ids);
            result = await downloadComicCollectionZip(p.batchId, p.bookId, collection.id);
        }
        await Taro.showModal({ title: '导出文件已准备好', content: '文件已保存到小程序文件中。也可以复制下载链接，在浏览器中下载。', confirmText: '复制链接', cancelText: '完成' }).then(answer => { if (answer.confirm && result.url)
            void Taro.setClipboardData({ data: result.url }); });
    });
    const selectPage = (index: number) => patch({ pageIndex: index });
    const chooseReference = async () => { try {
        const result = await Taro.chooseMedia({ count: 1, mediaType: ['image'], sourceType: ['album', 'camera'] });
        if (result.tempFiles[0]?.tempFilePath)
            patch({ referencePath: result.tempFiles[0].tempFilePath });
    }
    catch { /* user cancelled */ } };
    const toSource = () => patch({ stage: 'source' });
    const newProject = async () => {
        if (lock.current || watcher.current) return;
        if (Object.keys(project.current?.pendingImages || {}).length) {
            const choice = await Taro.showModal({title:'离开当前漫画？',content:'当前仍有图片未确认完成。开始新作品不会取消服务端生成，请先确认已记录当前作品。',confirmText:'开始新作品',cancelText:'留在这里'});
            if (!choice.confirm || lock.current || watcher.current) return;
        }
        epoch.current++; watcher.current=undefined; project.current=null; clearComicActiveRun();
        set({...initial(),setup:s.setup,setupLoading:s.setupLoading,setupError:s.setupError,styleId:s.styleId,thumbs:s.thumbs,thumbFailed:s.thumbFailed});
    };
    const p = s.project;
    const content = s.stage === 'source' ? 'source' : s.stage === 'assets' || s.stage === 'storyboard' ? 'assets' : 'pages';
    return <ThemedPage className={`mp-page comic-page comic-new comic-editor is-${content}${s.stage === 'storyboard' ? ' is-confirm' : ''}`} dynamicGlass={false}>
  <BrandHeader eyebrow="私人定制" title="漫画生成" showBack/>
  <View className="ce-project-bar"><View><Text className="ce-project-name">{(s.stage === 'source' ? s.source : p?.source || s.source).trim().split(/\r?\n/)[0]?.slice(0, 22) || '未命名漫画'}</Text><Text className="ce-muted">一战晟铭 · 先提示词，后出图</Text></View>{p ? <Button disabled={blocked} onClick={()=>void newProject()}>新作品</Button> : null}<Button disabled={blocked || !s.pages.length} onClick={() => void exportPages()}>导出</Button></View>
  <ComicStageHeader compact stage={s.stage} title="漫画生成" subtitle="" hasResult={Boolean(s.pages.length)} canSelectStage={stage => !blocked && (stage === 'source' || Boolean(p) && (stage === 'assets' || stage === 'storyboard' || s.pages.length > 0))} onSelectStage={stage => { const selectedProject=project.current; const selectedVersion=epoch.current; patch({ stage }); if (stage === 'render' && selectedProject)
        void refreshPages().catch(cause => {if(current(selectedVersion,selectedProject))patch({error:cause.message});}); }}/>
  <View className="ce-workspace-body">
   {s.stage === 'source' ? <ComicSettingsStage promptOnly source={s.source} styleId={s.styleId} aspectRatio={s.aspectRatio} referencePath={s.referencePath} styles={setup?.styles || []} preset={effective} presets={setup?.presets || []} selectedPresetId={effective?.preset.id || ''} advancedOpen={false} settingPicker={s.settingPicker} settingCatalog={setup?.settingCatalog || { pagePrompts: [], imageDerivations: [], constraints: { quality: [], restriction: [], negative: [] } }} models={setup?.models || []} settingSelections={s.selections} thumbs={s.thumbs} thumbFailed={s.thumbFailed} setupLoading={s.setupLoading} error={s.error} onSourceChange={source => patch({ source })} onPasteSource={() => { void Taro.getClipboardData().then(value => patch({ source: value.data || '' })).catch(() => patch({ error: '剪贴板读取失败，请直接粘贴' })); }} onChooseExisting={() => void Taro.navigateTo({ url: '/comic/pages/comic-config/index?mode=pick' })} onStyleChange={styleId => patch({ styleId })} onAspectChange={aspectRatio => patch({ aspectRatio })} onPresetChange={() => { }} onToggleAdvanced={() => { }} onOpenSettingPicker={settingPicker => patch({ settingPicker })} onCloseSettingPicker={() => patch({ settingPicker: '' })} onSettingSelectionChange={(key, value) => set(prev => ({ ...prev, selections: { ...prev.selections, [key]: value } }))} onConfirmSettingSelections={values => set(prev => ({ ...prev, selections: { ...prev.selections, ...values } }))} onOpenReference={() => patch({ referenceOpen: true })} onRetryThumb={id => { const style = setup?.styles.find(item => item.id === id); if (style?.previewUrl)
        void resolveComicStyleThumb(style.previewUrl).then(path => { set(prev => ({ ...prev, thumbs: { ...prev.thumbs, [id]: path }, thumbFailed: { ...prev.thumbFailed, [id]: !path } })); }); }} onContinue={() => void extract()} continueDisabled={Boolean(sourceReason) || blocked} continueLoading={s.busy === 'extract'} continueHint={sourceReason} configurationHint={configHint}/> : null}
   {s.stage === 'assets' && p ? <ComicAssetsStage key={p.batchId + ":" + p.bookId} batchId={p.batchId} bookId={p.bookId} assets={s.assets} assetImages={s.assetImages} assetThumbs={s.assetThumbs} loading={s.busy === 'extract'} busy={s.busy} busyAt={0} error={s.error} onGenerateImage={id => void assetImages(id)} onGenerateMissingImages={() => void assetImages()} onExtract={() => void extract()} onRefresh={() => void perform('refresh', async () => { await refreshAssets(); })} onRetry={() => void perform('retry', async () => { await retryComicBookStage({ batchId: p.batchId, bookId: p.bookId }); await refreshAssets(); })} onContinue={() => patch({ stage: 'storyboard' })} onBack={toSource}/> : null}
   {s.stage === 'storyboard' && p ? <ComicConfigConfirmStage promptOnly sourceLength={p.sourceLength} styleName={setup?.styles.find(style => style.id === p.styleId)?.name || ''} aspectRatio={p.aspectRatio} referenceReady={Boolean(p.reference)} assets={s.assets} submitting={blocked} error={s.error} textModelName={setup?.models.find(model => model.id === p.presetConfig.textModelId)?.name || '当前项目文本模型'} imageModelName={setup?.models.find(model => model.id === p.presetConfig.imageModelId)?.name || '当前项目图片模型'} onBack={() => patch({ stage: 'assets' })} onConfirm={() => void prompts()}/> : null}
   {(s.stage === 'render' || s.stage === 'result') && p ? <>
    <View className="ce-manual-status"><Text>{s.busy === 'prompts' ? '正在生成漫画提示词…' : s.busy === 'images' ? '正在读取页图生成结果…' : '先检查提示词，确认后再生成图片'}</Text><Button disabled={blocked} onClick={() => void perform('refresh', async () => { await refreshPages(); await refreshAssets(); if (Object.keys(project.current?.pendingImages || {}).length)
            beginWatch(); })}>刷新提示词与页图</Button></View>
    {s.error ? <Text className="mp-error">{s.error}</Text> : null}
    {!s.pages.length ? <View className="ce-empty"><Text>暂未读到漫画提示词</Text><Button disabled={blocked} onClick={() => void prompts()}>重试生成提示词</Button><Button disabled={blocked} onClick={() => patch({ stage: 'assets' })}>返回人物场景</Button></View> : <ComicPageWorkspace key={p.batchId + ":" + p.bookId} batchId={p.batchId} bookId={p.bookId} pages={s.pages} pageIndex={s.pageIndex} paths={s.paths} busy={s.busy} error="" generating={s.busy === 'images'} onSelect={selectPage} onPreview={id => { const paths = s.pages.flatMap(page => page.images).map(image => s.paths[image.id]).filter(Boolean); const url = s.paths[id]; if (url)
            void Taro.previewImage({ urls: paths, current: url }); }} onRegenerate={() => void generateImages(s.pages[s.pageIndex]?.videoId)} onPrimary={id => void perform('primary', async () => { await setPrimaryComicPageImage(p.batchId, p.bookId, s.pages[s.pageIndex].videoId, id); await refreshPages(); })} onSave={() => void perform('save', async () => { const page = s.pages[s.pageIndex]; const path = s.paths[page.primaryImageId]; if (!path)
            throw new Error('图片还未读取，请刷新'); await saveComicImageToAlbum(path); await Taro.showToast({ title: '已保存到相册', icon: 'success' }); })} onExport={() => void exportPages()} onReload={() => refreshPages()} onGenerateImages={() => void generateImages()} onGeneratePrompts={() => void prompts()} imageGenerationBlockedReason={!hasAssetImage ? '请先在人物场景中生成或上传至少一张资产图' : ''}/>}</> : null}
  </View>
  {s.referenceOpen ? <><View className="comic-sheet-mask" onClick={() => patch({ referenceOpen: false })}/><View className="comic-sheet"><View className="comic-sheet-handle"/><View className="comic-sheet-header"><Text>参考图</Text><Button onClick={() => patch({ referenceOpen: false })}>完成</Button></View><View className="comic-sheet-body"><Text className="comic-hint">参考图会随项目保存，并用于资产及漫画页的画风参考。</Text>{s.referencePath ? <Image className="comic-reference" mode="aspectFit" src={s.referencePath}/> : null}<View className="comic-sheet-actions"><Button className="mp-secondary" onClick={() => void chooseReference()}>选择图片</Button>{s.referencePath ? <Button onClick={() => patch({ referencePath: '' })}>移除参考图</Button> : null}</View></View></View></> : null}
 </ThemedPage>;
}
