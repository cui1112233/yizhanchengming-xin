import { Button, Text, View } from "@tarojs/components";
import { COMIC_ASSET_KIND_LABELS, comicAssetKindCounts, type ComicAssetItem } from "../../services/comicAssets";
import { estimateComicPages, estimateComicRunMinutes } from "../../services/runStage";

export type ComicConfigConfirmStageProps = {
  sourceLength: number;
  styleName: string;
  aspectRatio: string;
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
      <Text className="mp-row-description">核对正文、画风与资产。确认后将按这些配置自动分页并生成漫画图片。</Text>
      <View className="comic-config-summary">
        <View className="comic-config-summary-row"><Text>故事来源</Text><Text>正文 {props.sourceLength} 字</Text></View>
        <View className="comic-config-summary-row"><Text>画风</Text><Text>{props.styleName || "未选择画风"}</Text></View>
        <View className="comic-config-summary-row"><Text>图片画幅</Text><Text>{props.aspectRatio}</Text></View>
        <View className="comic-config-summary-row"><Text>参考图</Text><Text>{props.referenceReady ? "参考图已上传" : "未使用参考图"}</Text></View>
        <View className="comic-config-summary-row"><Text>预计页数</Text><Text>{pages ? `预计 ${pages} 页` : "待正文确定"}</Text></View>
        <View className="comic-config-summary-row"><Text>预计时长</Text><Text>{estimate}</Text></View>
      </View>
      <Text className="comic-config-assets-title">资产摘要 · 共 {props.assets.length} 项</Text>
      <View className="comic-config-counts">
        {(["character", "scene", "prop"] as const).map((kind) => <Text key={kind}>{COMIC_ASSET_KIND_LABELS[kind]} {counts[kind]}</Text>)}
      </View>
      <View className="comic-config-asset-list">
        {props.assets.map((asset) => <Text key={asset.id} className="comic-config-asset-item">{COMIC_ASSET_KIND_LABELS[asset.kind] || "资产"} · {asset.name || asset.formalName || "未命名资产"}</Text>)}
      </View>
      {!props.assets.length ? <Text className="comic-hint">资产还在提取中，不影响继续生成；需要调整时再回资产工作台。</Text> : null}
      <Text className="comic-hint">确认后自动分页出图。生成后可在“漫画页图”中逐页编辑提示词；如需调整资产，请先返回人物场景。</Text>
      {props.error ? <Text className="mp-error">{props.error}</Text> : null}
    </View>
    <View className="comic-dock">
      <View className="comic-dock-copy"><Text className="comic-dock-line">{props.submitting ? "正在提交生成配置…" : props.submitted ? "当前任务已提交，可查看生成进度" : "确认后开始分页出图"}</Text></View>
      <View className="comic-sticky-row">
        <Button className="mp-secondary" disabled={props.submitting} onClick={() => { if (!props.submitting) props.onBack(); }}>返回资产工作台</Button>
        <Button className="mp-primary" loading={props.submitting} disabled={props.submitting}
          onClick={() => { if (!props.submitting) props.onConfirm(); }}>{props.submitted ? "查看生成进度" : "确认配置，开始分页出图"}</Button>
      </View>
    </View>
  </View>;
}
