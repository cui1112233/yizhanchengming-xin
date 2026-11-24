import { Button, Image, Input, ScrollView, Text, Textarea, View } from "@tarojs/components";
import { useEffect, useRef, useState } from "react";
import AppIcon from "../../../components/AppIcon";
import { COMIC_ASSET_KIND_LABELS, comicAssetKindCounts, loadComicAssetImagePath, type ComicAssetImageItem, type ComicAssetItem } from "../../services/comicAssets";
import { saveComicAsset, setComicAssetPrimary, replaceComicImage } from "../../services/comicEditing";
export type ComicAssetsStageProps = {
  batchId?: string; bookId?: string; onGenerateImage?: (id: string) => void;
  assets: ComicAssetItem[];
  assetImages: Record<string, ComicAssetImageItem[]>;
  assetThumbs: Record<string, string>;
  loading: boolean;
  /** 服务端正在 comic_assets 阶段提取：此时列表为空是正常的，不是失败。 */
  serverExtracting?: boolean;
  /** 出图开关关闭时服务端会跳过资产图，空图不能显示成失败。 */
  imagesSkipped?: boolean;
  busy: string;
  busyAt: number;
  error: string;
  onExtract: () => void;
  onGenerateMissingImages: () => void;
  onRetry: () => void;
  onRefresh: () => void;
  onContinue: () => void;
  onBack: () => void;
};


export default function ComicAssetsStage(props: ComicAssetsStageProps) {
  const [kind, setKind] = useState("character");
  const [query, setQuery] = useState("");
  const [selectedId, setSelectedId] = useState("");
  const [tool, setTool] = useState<"" | "edit" | "versions" | "actions">("");
  const [failedThumbs, setFailedThumbs] = useState<Record<string,string>>({});
  const [draft, setDraft] = useState<ComicAssetItem | null>(null);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");
  const [paths, setPaths] = useState<Record<string,string>>({});
  const [versionsLoading, setVersionsLoading] = useState(false);
  const lock = useRef(false);
  const alive = useRef(true);
  useEffect(() => { alive.current = true; return () => { alive.current = false; }; }, []);
  const counts = comicAssetKindCounts(props.assets);
  const visible = props.assets.filter(a => a.kind === kind && (!query.trim() || [a.name,a.appearance,a.facts].join(" ").includes(query.trim())));
  const selected = visible.find(a => a.id === selectedId) || visible[0];
  const images = selected ? props.assetImages[selected.id] || [] : [];
  useEffect(() => {
    let active = true;
    setPaths({});
    setVersionsLoading(tool === "versions");
    if (tool === "versions") void Promise.all(images.map(async image => [image.id, await loadComicAssetImagePath(image)] as const)).then(entries => { if(active) {setPaths(Object.fromEntries(entries));setVersionsLoading(false);} });
    return () => { active = false; };
  }, [tool, selected?.id, images.map(i => i.id).join(",")]);
  const blocked = pending || props.loading || Boolean(props.busy) || Boolean(props.serverExtracting);
  const perform = async (action: () => Promise<unknown>) => {
    if (lock.current || blocked) return;
    lock.current = true; setPending(true); setError("");
    try { await action(); if(alive.current) { setTool(""); props.onRefresh(); } }
    catch (cause:any) { if(alive.current && !/cancel/i.test(cause?.errMsg || "")) setError(cause?.message || cause?.errMsg || "操作失败，请重试"); }
    finally { lock.current = false; if(alive.current) setPending(false); }
  };
  const batch = props.batchId || ""; const book = props.bookId || "";
  const missing = props.assets.filter(a => !(props.assetImages[a.id] || []).length).length;
  return <View className="comic-assets-stage ce-assets">
    <View className="ce-section-title"><Text>人物与场景</Text><View><Button disabled={blocked} onClick={() => {if(!blocked){setFailedThumbs({});props.onRefresh();}}}>刷新</Button><Button disabled={blocked} onClick={()=>setTool("actions")}>更多</Button></View></View>
    <View className="ce-asset-tabs">{(["character","scene","prop"] as const).map(value => <Button key={value} className={kind === value ? "is-active" : ""} onClick={() => { setKind(value); setSelectedId(""); }}>{COMIC_ASSET_KIND_LABELS[value]} {counts[value]}</Button>)}</View>
    <Input className="ce-search" value={query} placeholder="搜索名称或描述" onInput={e => setQuery(e.detail.value)} />
    {props.loading ? <Text className="comic-hint">正在读取资产…</Text> : null}
    {props.serverExtracting ? <Text className="comic-hint">正在提取角色与场景…</Text> : null}
    {props.busy ? <Text className="comic-hint">当前操作进行中，请稍候…</Text> : null}
    {props.error || error ? <Text className="mp-error">{error || props.error}</Text> : null}
    <View className="ce-asset-list">{visible.map(asset => <Button key={asset.id} className={selected?.id === asset.id ? "ce-asset-row is-selected" : "ce-asset-row"} onClick={() => setSelectedId(asset.id)}>
      {props.assetThumbs[asset.id] && failedThumbs[asset.id] !== props.assetThumbs[asset.id] ? <Image src={props.assetThumbs[asset.id]} mode="aspectFill" onError={()=>setFailedThumbs({...failedThumbs,[asset.id]:props.assetThumbs[asset.id]})} /> : <View className="ce-asset-placeholder"><AppIcon name={asset.kind === "character" ? "user-round" : "book-open"} size={27} /></View>}
      <View className="ce-asset-copy"><Text>{asset.name || asset.formalName || "未命名资产"}</Text><Text className="ce-muted ce-asset-excerpt">{asset.appearance || asset.facts || "暂无描述"}</Text><Text className="ce-muted">{props.assetImages[asset.id]?.length ? !props.assetThumbs[asset.id] || failedThumbs[asset.id] === props.assetThumbs[asset.id] ? "图片暂不可用" : "已有 " + props.assetImages[asset.id].length + " 个图片版本" : "尚未生成图片"}</Text></View>
      {selected?.id === asset.id ? <Text className="ce-check">✓</Text> : null}
    </Button>)}</View>
    {!visible.length && !props.loading && !props.serverExtracting ? <View className="ce-empty"><Text>{props.assets.length ? "当前分类没有匹配资产" : "还没有提取到资产"}</Text><Button disabled={blocked} onClick={props.onExtract}>重新提取资产</Button></View> : null}
    <Text className="ce-selection-hint">选择一项后，可编辑描述或生成图片。资产图可选。</Text>
    <View className="ce-tools ce-asset-tools">
      <Button disabled={!selected || blocked} onClick={() => { setDraft({...selected}); setTool("edit"); }}><AppIcon name="file-pen-line" size={23}/><Text>编辑描述</Text></Button>
      <Button disabled={!selected || blocked} onClick={() => selected && props.onGenerateImage?.(selected.id)}><AppIcon name="sparkles" size={23}/><Text>生成图片</Text></Button>
      <Button disabled={!selected || blocked} onClick={() => setTool("versions")}><AppIcon name="notebook-tabs" size={23}/><Text>图片版本</Text></Button>
      <Button disabled={!selected || !images.length || blocked} onClick={() => setTool("versions")}><AppIcon name="shield-check" size={23}/><Text>设为主图</Text></Button>
    </View>
    <View className="comic-dock"><View className="comic-dock-copy"><Text className="comic-dock-line">{props.assets.length} 项资产 · 缺 {missing} 张图（可选）</Text></View><View className="comic-sticky-row"><Button className="mp-secondary" disabled={blocked || !missing} onClick={() => { if(!blocked && missing) props.onGenerateMissingImages(); }}>补齐资产图（缺 {missing}）</Button><Button className="mp-primary" disabled={false} onClick={props.onContinue}>继续：生成配置确认</Button></View></View>
    {tool ? <><View className="comic-sheet-mask" onClick={() => { if(!pending) setTool(""); }}/><View className="comic-sheet ce-choice-sheet"><View className="ce-sheet-title"><Text>{tool === "edit" ? "编辑资产" : tool === "actions" ? "资产操作" : "图片版本"}</Text><Button disabled={pending} onClick={() => setTool("")}>关闭</Button></View><ScrollView scrollY className="ce-sheet-scroll">
      {tool === "edit" && draft ? <View className="ce-edit-fields">{(["name","formalName","appearance","facts","prompt"] as const).map(key => <View key={key}><Text>{{name:"名称",formalName:"正式名称",appearance:"外形描述",facts:"设定",prompt:"图片提示词"}[key]}</Text><Textarea maxlength={-1} value={draft[key]} disabled={pending} onInput={e => setDraft({...draft,[key]:e.detail.value})}/></View>)}<Button disabled={pending} onClick={() => perform(() => replaceComicImage(batch, book, {assetId:draft.id}))}>从相册选择图片</Button><Button className="mp-primary" loading={pending} disabled={!draft.name.trim() || pending} onClick={() => perform(() => saveComicAsset(batch,book,draft,{name:draft.name,formalName:draft.formalName,facts:draft.facts,appearance:draft.appearance,prompt:draft.prompt}))}>保存资产</Button></View> : null}
      {tool === "versions" ? <View className="ce-version-list">{images.map(image => <View key={image.id}>{paths[image.id] ? <Image src={paths[image.id]} mode="aspectFit"/> : <Text>{versionsLoading ? "正在读取图片…" : "图片暂不可用，请关闭后重试"}</Text>}<Button disabled={image.isPrimary || pending} onClick={() => perform(() => setComicAssetPrimary(batch,book,selected!.id,image.id))}>{image.isPrimary ? "当前主图" : "设为主图"}</Button></View>)}{!images.length ? <Text>暂无图片版本，请先生成或选择图片。</Text> : null}<Button disabled={pending || !selected} onClick={() => perform(() => replaceComicImage(batch,book,{assetId:selected!.id}))}>从相册选择图片</Button></View> : null}
      {tool === "actions" ? <View className="ce-edit-fields"><Button disabled={blocked} onClick={()=>{if(!blocked){setTool("");props.onExtract();}}}>重新提取资产</Button><Button disabled={blocked} onClick={()=>{if(!blocked){setTool("");props.onRetry();}}}>重试失败阶段</Button><Button disabled={blocked} onClick={()=>{if(!blocked)props.onBack();}}>返回设置</Button></View> : null}
      {error ? <Text className="mp-error">{error}</Text> : null}
    </ScrollView></View></> : null}
  </View>;
}
