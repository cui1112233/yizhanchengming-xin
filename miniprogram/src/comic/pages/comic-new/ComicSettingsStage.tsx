import { Button, Image, ScrollView, Text, Textarea, View } from "@tarojs/components";
import Taro from "@tarojs/taro";
import { useEffect, useRef, useState } from "react";
import { createSheetDismissal } from "./sheetMotion";
import AppIcon from "../../../components/AppIcon";
import {
  COMIC_ASPECT_RATIOS, COMIC_SOURCE_TRUNCATE_HINT_CHARS, MIN_COMIC_SOURCE_CHARS,
  comicSourceLength, type ComicCatalogPreset, type ComicModelOption, type ComicPresetOption,
  type ComicSettingCatalog, type ComicSettingSelections, type ComicStylePreset,
} from "../../services/comicFlow";
import { privateComicAction } from "./privateComicParity";
import { estimateComicRunText } from "../../services/runStage";

export type ComicSettingsStageProps = {
  source: string;
  styleId: string;
  aspectRatio: string;
  referencePath: string;
  styles: ComicStylePreset[];
  preset?: ComicPresetOption;
  presets: ComicPresetOption[];
  selectedPresetId: string;
  advancedOpen: boolean;
  settingPicker: "" | "prompt" | "derivation" | "model" | "constraints";
  settingCatalog: ComicSettingCatalog;
  models: ComicModelOption[];
  settingSelections: ComicSettingSelections;
  thumbs: Record<string, string>;
  thumbFailed: Record<string, boolean>;
  setupLoading: boolean;
  error: string;
  onSourceChange: (value: string) => void;
  onPasteSource: () => void;
  onChooseExisting: () => void;
  onStyleChange: (value: string) => void;
  onAspectChange: (value: string) => void;
  onPresetChange: (value: string) => void;
  onToggleAdvanced: () => void;
  onOpenSettingPicker: (value: "prompt" | "derivation" | "model" | "constraints") => void;
  onCloseSettingPicker: () => void;
  onSettingSelectionChange: (key: keyof ComicSettingSelections, value: string) => void;
  onConfirmSettingSelections: (values: ComicSettingSelections) => void;
  onOpenReference: () => void;
  onRetryThumb: (styleId: string) => void;
  onContinue: () => void;
  continueDisabled?: boolean;
  continueLoading?: boolean;
  continueHint?: string;
};

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : {};
}

function text(value: unknown) {
  return typeof value === "string" && value.trim() ? value.trim() : "";
}

function ratioOutline(value: string) {
  const [width, height] = value.split(":").map(Number);
  return {width: `${22 * Math.min(1, width / height)}px`, height: `${22 * Math.min(1, height / width)}px`};
}

function presetSummary(option?: ComicPresetOption) {
  const config = asRecord(option?.preset.config);
  const comic = asRecord(asRecord(config.aiPromptConfig).comic);
  const constraints = asRecord(comic.constraints);
  return {
    prompt: text(comic.presetName) || (text(comic.presetId) || text(comic.body) ? "已配置" : "未选择"),
    derivation: text(asRecord(comic.imageDerivation).presetName) || text(comic.imageDerivation) || (comic.imageDerivation ? "已配置" : "未选择"),
    model: text(config.imageModelName) || text(config.imageModelId) || "未选择",
    constraints: Object.keys(constraints).length || text(comic.quality) || text(comic.restriction) || text(comic.negative)
      ? "已随预设配置"
      : "使用默认约束",
  };
}

/** 设置只负责输入和继续意图；创建、上传、提取等服务动作由页面持有。 */
export default function ComicSettingsStage({
  source, styleId, aspectRatio, referencePath, styles, preset, presets, selectedPresetId, advancedOpen, settingPicker, settingCatalog, models, settingSelections, thumbs, thumbFailed, setupLoading, error,
  onSourceChange, onPasteSource, onChooseExisting, onStyleChange, onAspectChange, onPresetChange, onToggleAdvanced, onOpenReference, onRetryThumb, onContinue,
  onOpenSettingPicker, onCloseSettingPicker, onSettingSelectionChange, onConfirmSettingSelections,
  continueDisabled = false, continueLoading = false, continueHint = "",
}: ComicSettingsStageProps) {
  const [tool, setTool] = useState<"" | "style" | "ratio">("");
  const [draft, setDraft] = useState<ComicSettingSelections>({});
  const [draftStyle, setDraftStyle] = useState(styleId);
  const [draftRatio, setDraftRatio] = useState(aspectRatio);
  const [detail, setDetail] = useState("");
  const [closing, setClosing] = useState(false);
  const [reducedMotion, setReducedMotion] = useState(() => Taro.getStorageSync("yzcm.mobile.reduceMotion") === true);
  const dismissal = useRef<ReturnType<typeof createSheetDismissal>>();
  if (!dismissal.current) dismissal.current = createSheetDismissal(setClosing);
  useEffect(() => () => dismissal.current?.cancel(), []);
  const beginOpen = () => { dismissal.current!.open(); setReducedMotion(Taro.getStorageSync("yzcm.mobile.reduceMotion") === true); };
  const dismiss = (finish: () => void, commit?: () => void) => dismissal.current!.dismiss(Taro.getStorageSync("yzcm.mobile.reduceMotion") === true, finish, commit);
  const openPicker = (value: "prompt" | "derivation" | "model" | "constraints") => { beginOpen(); setDraft({}); setDetail(""); onOpenSettingPicker(value); };
  const openTool = (value: "style" | "ratio") => { beginOpen(); setDraftStyle(styleId); setDraftRatio(aspectRatio); setTool(value); };
  const closeTool = () => { dismiss(() => setTool("")); };
  const closePicker = () => { dismiss(onCloseSettingPicker); };
  const motionClass = closing ? " is-closing" : "";
  const choose = (key: keyof ComicSettingSelections, value: string) => setDraft(prev => ({...prev, [key]: value}));
  const selected = (key: keyof ComicSettingSelections, fallback: string) => draft[key] ?? settingSelections[key] ?? fallback;
  const length = comicSourceLength(source);
  const remaining = Math.max(0, MIN_COMIC_SOURCE_CHARS - length);
  const tooLong = length > COMIC_SOURCE_TRUNCATE_HINT_CHARS;
  const action = privateComicAction("settings", {
    sourceLength: length, minSourceLength: MIN_COMIC_SOURCE_CHARS, styleId,
    batchId: "", assetsCount: 0, pagesCount: 0, submitted: false,
    busy: setupLoading || continueLoading || continueDisabled,
  });
  const countText = length ? (remaining > 0 ? `已输入 ${length} 字，还需 ${remaining} 字` : `已输入 ${length} 字，可以开始生成了`) : "还没有输入正文";
  // 估算只在够提交长度后出现，措辞固定「预计 / 约 / 估算」，不写成承诺。
  const estimate = length >= MIN_COMIC_SOURCE_CHARS ? estimateComicRunText(length) : "";
  const selectedStyle = styles.find((style) => style.id === styleId)?.name || "未选择";
  const baseSummary = presetSummary(preset);
  const pickerTitle = settingPicker ? {
    prompt: "选择漫画提示词",
    derivation: "选择图片描述词推导",
    model: "选择模型配置",
    constraints: "选择画质与约束",
  }[settingPicker] : "";
  const config = asRecord(preset?.preset.config);
  const selectedTextModel = models.find((model) => model.kind === "text" && model.id === text(config.textModelId));
  const selectedImageModel = models.find((model) => model.kind === "image" && model.id === text(config.imageModelId));
  const selectedModelNames = [selectedTextModel?.name, selectedImageModel?.name].filter(Boolean).join(" · ");
  const summary = { ...baseSummary, model: selectedModelNames || baseSummary.model };
  const comic = asRecord(asRecord(config.aiPromptConfig).comic);
  const currentPromptId = text(comic.presetId);
  const currentDerivationId = text(asRecord(comic.imageDerivation).presetId);
  const currentConstraintId = (category: string) => (Array.isArray(comic.constraints) ? comic.constraints : [])
    .map(asRecord).find((item) => text(item.constraintCategory) === category)?.presetId as string || "";
  const choice = (id: string, name: string, caption: string, active: boolean, onClick: () => void) => <View key={id} className={`comic-preset-choice ce-option-row${active ? " is-active" : ""}`} hoverClass="ce-row-press" hoverStartTime={0} hoverStayTime={70} onClick={() => {if (!closing) onClick();}}>
    <View className="comic-preset-copy"><Text className="comic-preset-name">{name}</Text><Text className="comic-caption">{caption}</Text>{detail === id ? <Text className="ce-option-detail">完整名称：{name}。{caption}</Text> : null}</View>
    <View className="ce-option-side"><View className={`ce-option-radio${active ? " is-checked" : ""}`}>{active ? <Text className="ce-check-glyph">✓</Text> : null}</View><Button className="ce-option-detail-button" onClick={event => {event.stopPropagation();setDetail(detail === id ? "" : id);}}>{detail === id ? "收起" : "详情"}</Button></View>
  </View>;
  const presetChoice = (key: keyof ComicSettingSelections, option: ComicCatalogPreset, current: string, caption: string) => choice(`${key}:${option.id}`, option.name, `${caption} · v${option.version}`, selected(key, current) === option.id, () => choose(key, option.id));
  const hint = continueHint || (setupLoading ? "正在读取画风和预设…" : !styleId ? "还没有可用画风，稍后再试" : !length ? "先粘贴小说正文" : remaining > 0 ? `还需 ${remaining} 字` : "先创建项目并提取资产，下一步再确认分镜");
  return <View className={`comic-settings-stage${reducedMotion ? " is-motion-reduced" : ""}`}>
    {error ? <Text className="mp-error comic-error">{error}</Text> : null}


    <View className={`ce-source-card${tooLong || setupLoading ? " has-notice" : ""}`}>
    <View className="ce-source-head"><Text>故事正文</Text><View><Button onClick={onPasteSource}>粘贴内容</Button><Button onClick={onChooseExisting}>选择已有书籍</Button></View></View>
    <Textarea className="ce-source-input" value={source} maxlength={-1} placeholder="粘贴要生成漫画的正文…" onInput={event => onSourceChange(String(event.detail.value))} />
    <View className="ce-source-count"><Text>{length.toLocaleString()} 字</Text><Button onClick={() => onSourceChange("")}>清空</Button></View>
    {tooLong ? <Text className="comic-hint">内容较长，一次最多生成 35 页，请确认处理范围。</Text> : null}
    {setupLoading ? <Text className="comic-hint">正在读取画风和生成设置…</Text> : null}
    </View>
    <View className="ce-tools ce-source-tools">
      <Button id="comic-tool-style" className={tool === "style" ? "is-open" : ""} hoverClass="ce-tool-press" hoverStartTime={0} hoverStayTime={70} onClick={() => openTool("style")}><AppIcon name="sparkles" size={23} /><Text>画风</Text><Text className="ce-tool-value">{selectedStyle}</Text></Button>
      <Button id="comic-tool-ratio" className={tool === "ratio" ? "is-open" : ""} hoverClass="ce-tool-press" hoverStartTime={0} hoverStayTime={70} onClick={() => openTool("ratio")}><AppIcon name="camera" size={23} /><Text>画幅</Text><Text className="ce-tool-value">{aspectRatio}</Text></Button>
      <Button id="comic-tool-model" className={settingPicker === "model" ? "is-open" : ""} hoverClass="ce-tool-press" hoverStartTime={0} hoverStayTime={70} onClick={() => openPicker("model")}><AppIcon name="bot" size={23} /><Text>模型</Text><Text className="ce-tool-value">{summary.model}</Text></Button>
      <Button id="comic-tool-prompt" className={Boolean(settingPicker && settingPicker !== "model") ? "is-open" : ""} hoverClass="ce-tool-press" hoverStartTime={0} hoverStayTime={70} onClick={() => openPicker("prompt")}><AppIcon name="file-pen-line" size={23} /><Text>提示词</Text></Button>
      <Button id="comic-tool-reference" hoverClass="ce-tool-press" hoverStartTime={0} hoverStayTime={70} onClick={onOpenReference}><AppIcon name="book-open" size={23} /><Text>参考图</Text></Button>
    </View>
    {tool ? <><View className={`comic-sheet-mask ce-motion-mask${motionClass}`} onClick={closeTool} /><View className={`comic-sheet ce-choice-sheet ce-motion-sheet${tool === "ratio" ? " is-ratio" : ""}${motionClass}`}>
      <View className="comic-sheet-handle" /><View className="ce-sheet-title"><Text>{tool === "style" ? "漫画风格" : "图片画幅"}</Text><Button onClick={closeTool}>取消</Button></View>
      <Text className="ce-sheet-subtitle">{tool === "style" ? "选择这次漫画的画面风格" : "选择适合发布平台的图片比例"}</Text>
      <ScrollView scrollY className="ce-sheet-scroll">{tool === "style" ? <View className="ce-style-grid">{styles.map(style => <Button key={style.id} className={draftStyle === style.id ? "is-selected" : ""} onClick={() => setDraftStyle(style.id)}>
        {thumbs[style.id] ? <Image src={thumbs[style.id]} mode="aspectFill" /> : <View className={thumbFailed[style.id] ? "comic-style-thumb is-failed" : style.previewUrl ? "comic-style-thumb is-skeleton" : "ce-style-placeholder"} onClick={event=>{if(thumbFailed[style.id]){event.stopPropagation();onRetryThumb(style.id);}}}><Text>{thumbFailed[style.id] ? "图片读取失败，点按重试" : style.previewUrl ? "读取预览图…" : style.name}</Text></View>}
        <View className="ce-style-label"><Text>{style.name}</Text><View className={`ce-option-radio${draftStyle === style.id ? " is-checked" : ""}`}>{draftStyle === style.id ? <Text className="ce-check-glyph">✓</Text> : null}</View></View>
      </Button>)}</View> : <View className="comic-aspect-grid">{COMIC_ASPECT_RATIOS.map(value => <Button key={value} className={draftRatio === value ? "comic-aspect-choice is-active" : "comic-aspect-choice"} onClick={() => setDraftRatio(value)}><View className="ce-ratio-outline" style={ratioOutline(value)} /><Text>{value}</Text>{draftRatio === value ? <Text className="ce-ratio-check">✓</Text> : null}</Button>)}</View>}
      {tool === "style" && !styles.length ? <Text className="comic-hint">暂无可用画风，请检查账号配置后重试。</Text> : null}</ScrollView>
      <View className="ce-sheet-footer"><Button className="ce-sheet-confirm" hoverClass="ce-confirm-press" hoverStartTime={0} hoverStayTime={70} disabled={tool === "style" && !styles.some(style => style.id === draftStyle)} onClick={() => {dismiss(() => setTool(""), () => {if(tool === "style") onStyleChange(draftStyle);else onAspectChange(draftRatio);});}}>完成</Button></View>
    </View></> : null}
    {pickerTitle ? <>
      <View className={`comic-sheet-mask ce-motion-mask${motionClass}`} onClick={closePicker} />
      <View className={`comic-sheet comic-setting-sheet ce-catalog-sheet ce-motion-sheet${motionClass}`}>
        <View className="comic-sheet-handle" />
        <View className="comic-setting-sheet-head"><Text className="comic-head-title">{pickerTitle}</Text><Button className="comic-sheet-cancel" onClick={closePicker}>取消</Button></View>
        <Text className="ce-sheet-subtitle">{settingPicker === "model" ? "分别选择文本和图片使用的模型" : "选择适合这次创作的提示词配置"}</Text>
        {settingPicker !== "model" ? <View className="ce-prompt-categories"><View className="ce-category-slider" style={{transform: `translateX(${settingPicker === "constraints" ? 200 : settingPicker === "derivation" ? 100 : 0}%)`}} />{([["prompt", "漫画分页"], ["derivation", "图片描述词推导"], ["constraints", "画质与约束"]] as const).map(([value, label]) => <Button key={value} id={`comic-category-${value}`} disabled={closing} className={settingPicker === value ? "is-active" : ""} onClick={() => {beginOpen();setDetail("");onOpenSettingPicker(value);}}>{label}</Button>)}</View> : null}
        <ScrollView key={settingPicker} scrollY className="comic-setting-sheet-list ce-category-content">
          {settingPicker === "prompt" ? <View className="comic-setting-group">{settingCatalog.pagePrompts.map(option => presetChoice("pagePromptId",option,currentPromptId,"漫画分页规则"))}{!settingCatalog.pagePrompts.length ? <Text className="ce-sheet-empty">暂无可用漫画分页提示词</Text> : null}</View> : null}
          {settingPicker === "derivation" ? <View className="comic-setting-group">{settingCatalog.imageDerivations.map(option => presetChoice("imageDerivationId",option,currentDerivationId,"图片描述词推导"))}{!settingCatalog.imageDerivations.length ? <Text className="ce-sheet-empty">暂无可用图片描述词</Text> : null}</View> : null}
          {settingPicker === "model" ? <>{(["text", "image"] as const).map(kind => <View key={kind} className="comic-setting-group"><Text className="comic-setting-group-title">{kind === "text" ? "文本模型" : "图片模型"}</Text>{models.filter(model => model.kind === kind).map(model => {const key = kind === "text" ? "textModelId" : "imageModelId";return choice(`${key}:${model.id}`,model.name,kind === "text" ? "文本模型" : model.provider || "图片模型",selected(key,text(config[key])) === model.id,() => choose(key,model.id));})}{!models.some(model => model.kind === kind) ? <Text className="ce-sheet-empty">暂无可用{kind === "text" ? "文本" : "图片"}模型</Text> : null}</View>)}</> : null}
          {settingPicker === "constraints" ? <>{([
            ["画质约束", "qualityId", "comic_quality", settingCatalog.constraints.quality],
            ["画面限制", "restrictionId", "comic_restriction", settingCatalog.constraints.restriction],
            ["负面提示词", "negativeId", "comic_negative", settingCatalog.constraints.negative],
          ] as const).map(([label,key,category,options]) => <View key={category} className="comic-setting-group"><Text className="comic-setting-group-title">{label}</Text>{options.map(option => presetChoice(key,option,currentConstraintId(category),label))}{!options.length ? <Text className="ce-sheet-empty">暂无可用{label}</Text> : null}</View>)}</> : null}
        </ScrollView>
        <View className="ce-sheet-footer"><Text className="ce-sheet-note">仅用于下一次创建，已有项目保持原配置</Text><Button className="ce-sheet-confirm" hoverClass="ce-confirm-press" hoverStartTime={0} hoverStayTime={70} onClick={() => {dismiss(onCloseSettingPicker, () => onConfirmSettingSelections(draft));}}>{settingPicker === "model" ? "使用此配置" : "使用此提示词"}</Button></View>
      </View>
    </> : null}

    <View className="comic-dock">
      <View className="comic-dock-copy">
        <Text className="comic-dock-line">{countText}</Text>
        {estimate ? <Text className="comic-dock-caption">{estimate}</Text> : null}
        <Text className="comic-dock-caption">{hint}</Text>
      </View>
      <Button className="mp-primary" loading={continueLoading} disabled={!action.enabled} onClick={() => { if (action.enabled) onContinue(); }}>
        {continueLoading ? "正在创建项目…" : "提取人物场景"}
      </Button>
    </View>
  </View>;
}
