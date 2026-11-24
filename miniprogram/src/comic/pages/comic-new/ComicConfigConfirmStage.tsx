import { Button, Text, View } from "@tarojs/components";
import { COMIC_ASSET_KIND_LABELS, comicAssetKindCounts, type ComicAssetItem } from "../../services/comicAssets";
import { estimateComicPages, estimateComicRunMinutes } from "../../services/runStage";

export type ComicConfigConfirmStageProps = {
  promptOnly?: boolean;
  sourceLength: number;
  styleName: string;
  aspectRatio: string;
  textModelName?: string;
  imageModelName?: string;
  referenceReady: boolean;
  assets: ComicAssetItem[];
  submitting: boolean;
  submitted?: boolean;
  error: string;
  onBack: () => void;
  onConfirm: () => void;
};

export default function ComicConfigConfirmStage(props: ComicConfigConfirmStageProps) {
  const counts = comicAssetKindCounts(props.assets);
  const pages = estimateComicPages(props.sourceLength);
  const minutes = estimateComicRunMinutes(pages);
  const estimate = pages ? `约 ${minutes.low}~${minutes.high} 分钟（估算）` : "待正文确定";
  return <View className="comic-config-confirm-stage">
    <View className="mp-card comic-stage-gate">
      <Text className="mp-row-title">生成配置确认</Text>
      <View className="ce-generation-scope"><Text>本次会执行</Text><Text>{props.promptOnly ? "只生成漫画分页和提示词，不生成图片" : "自动划分漫画页，并为每页生成图片"}</Text><Text>{props.promptOnly ? "检查并编辑提示词后，再单独点击生成图片。" : "实际页数与耗时以生成结果为准，下面仅为估算。"}</Text></View>
      <Text className="mp-row-description">{props.promptOnly ? "核对当前项目的正文、画风和资产。本次只调用文本模型生成漫画提示词。" : "核对正文、画风与资产。确认后将按这些配置自动分页并生成漫画图片。"}</Text>
      <View className="comic-config-summary">
        <View className="comic-config-summary-row"><Text>故事来源</Text><Text>正文 {props.sourceLength} 字</Text></View>
        <View className="comic-config-summary-row"><Text>画风</Text><Text>{props.styleName || "未选择画风"}</Text></View>
        <View className="comic-config-summary-row"><Text>文本模型</Text><Text>{props.textModelName || "查看项目配置"}</Text></View>
        <View className="comic-config-summary-row"><Text>图片模型</Text><Text>{props.imageModelName || "查看项目配置"}</Text></View>
        <View className="comic-config-summary-row"><Text>图片画幅</Text><Text>{props.aspectRatio}</Text></View>
        <View className="comic-config-summary-row"><Text>参考图</Text><Text>{props.referenceReady ? "参考图已上传" : "未使用参考图"}</Text></View>
        <View className="comic-config-summary-row"><Text>预计页数</Text><Text>{pages ? `预计 ${pages} 页` : "待正文确定"}</Text></View>
        {!props.promptOnly ? <View className="comic-config-summary-row"><Text>预计时长</Text><Text>{estimate}</Text></View> : null}
      </View>
      <Text className="comic-config-assets-title">资产摘要 · 共 {props.assets.length} 项</Text>
      <View className="comic-config-counts">
        {(["character", "scene", "prop"] as const).map((kind) => <Text key={kind}>{COMIC_ASSET_KIND_LABELS[kind]} {counts[kind]}</Text>)}
      </View>
      <View className="comic-config-asset-list">
        {props.assets.map((asset) => <Text key={asset.id} className="comic-config-asset-item">{COMIC_ASSET_KIND_LABELS[asset.kind] || "资产"} · {asset.name || asset.formalName || "未命名资产"}</Text>)}
      </View>
      {!props.assets.length ? <Text className="comic-hint">当前未读取到人物或场景。可以返回刷新检查，也可继续提交，由服务端处理。</Text> : null}
      <Text className="comic-hint">{props.promptOnly ? "生成提示词后可逐页检查、编辑或导出；漫画图片需要另行生成，不会自动出图。" : "确认后自动分页出图。生成后可在“漫画页图”中逐页编辑提示词；如需调整资产，请先返回人物场景。"}</Text>
      {props.error ? <Text className="mp-error">{props.error}</Text> : null}
    </View>
    <View className="comic-dock">
      <View className="comic-dock-copy"><Text className="comic-dock-line">{props.submitting ? "正在提交生成配置…" : props.submitted ? "当前任务已提交，可查看生成进度" : props.promptOnly ? "确认后只生成提示词" : "确认后开始分页出图"}</Text></View>
      <View className="comic-sticky-row">
        <Button className="mp-secondary" disabled={props.submitting} onClick={() => { if (!props.submitting) props.onBack(); }}>返回资产工作台</Button>
        <Button className="mp-primary" loading={props.submitting} disabled={props.submitting}
          onClick={() => { if (!props.submitting) props.onConfirm(); }}>{props.submitted ? "查看生成进度" : props.promptOnly ? "生成漫画提示词" : "开始生成漫画"}</Button>
      </View>
    </View>
  </View>;
}
