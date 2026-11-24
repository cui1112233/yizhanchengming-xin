import { Button, Image, ScrollView, Text, Textarea, View } from "@tarojs/components";
import { useState } from "react";
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
  onOpenSettingPicker, onCloseSettingPicker, onSettingSelectionChange,
  continueDisabled = false, continueLoading = false, continueHint = "",
}: ComicSettingsStageProps) {
  const [tool, setTool] = useState<"" | "style" | "ratio">("");
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
  const chooseCatalogPreset = (key: keyof ComicSettingSelections, option: ComicCatalogPreset) => {
    onSettingSelectionChange(key, option.id);
    onCloseSettingPicker();
  };
  const hint = continueHint || (setupLoading ? "正在读取画风和预设…" : !styleId ? "还没有可用画风，稍后再试" : !length ? "先粘贴小说正文" : remaining > 0 ? `还需 ${remaining} 字` : "先创建项目并提取资产，下一步再确认分镜");
  return <View className="comic-settings-stage">
    {error ? <Text className="mp-error comic-error">{error}</Text> : null}


    <View className="ce-source-head"><Text>故事正文</Text><View><Button onClick={onPasteSource}>粘贴内容</Button><Button onClick={onChooseExisting}>选择已有书籍</Button></View></View>
    <Textarea className="ce-source-input" value={source} maxlength={-1} placeholder="粘贴要生成漫画的正文…" onInput={event => onSourceChange(String(event.detail.value))} />
    <View className="ce-source-count"><Text>{length.toLocaleString()} 字</Text><Button onClick={() => onSourceChange("")}>清空</Button></View>
    {tooLong ? <Text className="comic-hint">内容较长，一次最多生成 35 页，请确认处理范围。</Text> : null}
    {setupLoading ? <Text className="comic-hint">正在读取画风和生成设置…</Text> : null}
    <View className="ce-tools ce-source-tools">
      <Button id="comic-tool-style" onClick={() => setTool("style")}><AppIcon name="sparkles" size={23} /><Text>画风</Text><Text className="ce-tool-value">{selectedStyle}</Text></Button>
      <Button id="comic-tool-ratio" onClick={() => setTool("ratio")}><AppIcon name="camera" size={23} /><Text>画幅</Text><Text className="ce-tool-value">{aspectRatio}</Text></Button>
      <Button id="comic-tool-model" onClick={() => onOpenSettingPicker("model")}><AppIcon name="bot" size={23} /><Text>模型</Text><Text className="ce-tool-value">{summary.model}</Text></Button>
      <Button id="comic-tool-prompt" onClick={() => onOpenSettingPicker("prompt")}><AppIcon name="file-pen-line" size={23} /><Text>提示词</Text></Button>
      <Button id="comic-tool-reference" onClick={onOpenReference}><AppIcon name="book-open" size={23} /><Text>参考图</Text></Button>
    </View>
    {tool ? <><View className="comic-sheet-mask" onClick={() => setTool("")} /><View className={`comic-sheet ce-choice-sheet${tool === "ratio" ? " is-ratio" : ""}`}><View className="ce-sheet-title"><Text>{tool === "style" ? "漫画风格" : "图片画幅"}</Text><Button onClick={() => setTool("")}>完成</Button></View>
      <ScrollView scrollY className="ce-sheet-scroll">{tool === "style" ? <View className="ce-style-grid">{styles.map(style => <Button key={style.id} className={styleId === style.id ? "is-selected" : ""} onClick={() => onStyleChange(style.id)}>
        {thumbs[style.id] ? <Image src={thumbs[style.id]} mode="aspectFill" /> : <View className={thumbFailed[style.id] ? "comic-style-thumb is-failed" : style.previewUrl ? "comic-style-thumb is-skeleton" : "ce-style-placeholder"} onClick={event=>{if(thumbFailed[style.id]){event.stopPropagation();onRetryThumb(style.id);}}}><Text>{thumbFailed[style.id] ? "图片读取失败，点按重试" : style.previewUrl ? "读取预览图…" : style.name}</Text></View>}
        <Text>{style.name}</Text>{styleId === style.id ? <Text>已选</Text> : null}
      </Button>)}</View> : <View className="comic-aspect-grid">{COMIC_ASPECT_RATIOS.map(value => <Button key={value} className={aspectRatio === value ? "comic-aspect-choice is-active" : "comic-aspect-choice"} onClick={() => onAspectChange(value)}>{value}</Button>)}</View>}
      {tool === "style" && !styles.length ? <Text className="comic-hint">暂无可用画风，请检查账号配置后重试。</Text> : null}</ScrollView>
    </View></> : null}
    {pickerTitle ? <>
      <View className="comic-sheet-mask" onClick={onCloseSettingPicker} />
      <View className="comic-sheet comic-setting-sheet">
        <View className="comic-setting-sheet-head"><Text className="comic-head-title">{pickerTitle}</Text><Text className="comic-sheet-cancel" onClick={onCloseSettingPicker}>取消</Text></View>
        {settingPicker !== "model" ? <View className="ce-prompt-categories"><Button onClick={() => onOpenSettingPicker("prompt")}>漫画分页</Button><Button onClick={() => onOpenSettingPicker("derivation")}>图片描述词推导</Button><Button onClick={() => onOpenSettingPicker("constraints")}>画质与约束</Button></View> : null}
        <Text className="comic-hint">各项单独选择，关闭面板后保留本次选择。</Text>
        <ScrollView scrollY className="comic-setting-sheet-list">
          {settingPicker === "prompt" ? <View className="comic-setting-group">
            {settingCatalog.pagePrompts.map((option) => <View key={option.id} className={`comic-preset-choice ${currentPromptId === option.id ? "is-active" : ""}`} onClick={() => chooseCatalogPreset("pagePromptId", option)}>
              <View className="comic-preset-copy"><Text className="comic-preset-name">{option.name}</Text><Text className="comic-caption">漫画分页规则 · v{option.version}</Text></View><Text className="comic-choice-mark">{currentPromptId === option.id ? "已选" : "选择"}</Text>
            </View>)}
          </View> : null}
          {settingPicker === "derivation" ? <View className="comic-setting-group">
            {settingCatalog.imageDerivations.map((option) => <View key={option.id} className={`comic-preset-choice ${currentDerivationId === option.id ? "is-active" : ""}`} onClick={() => chooseCatalogPreset("imageDerivationId", option)}>
              <View className="comic-preset-copy"><Text className="comic-preset-name">{option.name}</Text><Text className="comic-caption">图片描述词推导 · v{option.version}</Text></View><Text className="comic-choice-mark">{currentDerivationId === option.id ? "已选" : "选择"}</Text>
            </View>)}
          </View> : null}
          {settingPicker === "model" ? <>
            <View className="comic-setting-group"><Text className="comic-setting-group-title">文本模型</Text>{models.filter((model) => model.kind === "text").map((model) => <View key={model.id} className={`comic-preset-choice ${text(config.textModelId) === model.id ? "is-active" : ""}`} onClick={() => onSettingSelectionChange("textModelId", model.id)}><View className="comic-preset-copy"><Text className="comic-preset-name">{model.name}</Text><Text className="comic-caption">文本模型</Text></View><Text className="comic-choice-mark">{text(config.textModelId) === model.id ? "已选" : "选择"}</Text></View>)}</View>
            <View className="comic-setting-group"><Text className="comic-setting-group-title">图片模型</Text>{models.filter((model) => model.kind === "image").map((model) => <View key={model.id} className={`comic-preset-choice ${text(config.imageModelId) === model.id ? "is-active" : ""}`} onClick={() => onSettingSelectionChange("imageModelId", model.id)}><View className="comic-preset-copy"><Text className="comic-preset-name">{model.name}</Text><Text className="comic-caption">{model.provider || "图片生成通道随模型配置"}</Text></View><Text className="comic-choice-mark">{text(config.imageModelId) === model.id ? "已选" : "选择"}</Text></View>)}</View>
          </> : null}
          {settingPicker === "constraints" ? <>
            {([
              ["画质约束", "qualityId", "comic_quality", settingCatalog.constraints.quality],
              ["画面限制", "restrictionId", "comic_restriction", settingCatalog.constraints.restriction],
              ["负面提示词", "negativeId", "comic_negative", settingCatalog.constraints.negative],
            ] as const).map(([label, key, category, options]) => <View key={category} className="comic-setting-group"><Text className="comic-setting-group-title">{label}</Text>{options.map((option) => <View key={option.id} className={`comic-preset-choice ${currentConstraintId(category) === option.id ? "is-active" : ""}`} onClick={() => onSettingSelectionChange(key, option.id)}><View className="comic-preset-copy"><Text className="comic-preset-name">{option.name}</Text><Text className="comic-caption">v{option.version}</Text></View><Text className="comic-choice-mark">{currentConstraintId(category) === option.id ? "已选" : "选择"}</Text></View>)}</View>)}
          </> : null}
        </ScrollView>
        <Text className="comic-hint">用于下一次创建的漫画；已有项目沿用创建时的配置。</Text>
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
