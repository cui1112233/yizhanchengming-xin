import { Text, View } from "@tarojs/components";
import { Button } from "@tarojs/components";
import AppIcon from "../../../components/AppIcon";
import type { ComicStage } from "./stageState";
type DisplayStage = { target: ComicStage; label: string; matches: ComicStage[] };

const STAGES: DisplayStage[] = [
  { target: "source", label: "输入正文", matches: ["source"] },
  { target: "assets", label: "角色场景", matches: ["assets"] },
  { target: "storyboard", label: "分镜分页", matches: ["storyboard"] },
  { target: "result", label: "生成结果", matches: ["render", "result"] },
];

export type ComicStageHeaderProps = {
  compact?: boolean;
  stage: ComicStage;
  title: string;
  subtitle: string;
  hasResult: boolean;
  canSelectStage: (stage: ComicStage) => boolean;
  onSelectStage: (stage: ComicStage) => void;
};

export default function ComicStageHeader({ compact, stage, title, subtitle, hasResult, canSelectStage, onSelectStage }: ComicStageHeaderProps) {
  if (compact) return <View className="ce-content-nav">{([
    ["source", "正文", "file-pen-line"], ["assets", "人物场景", "users-round"], ["render", "漫画页图", "book-open"],
  ] as const).map(([target,label,icon]) => <Button key={target} className={(target === "source" ? stage === "source" : target === "assets" ? stage === "assets" || stage === "storyboard" : stage === "render" || stage === "result") ? "is-active" : ""} disabled={!canSelectStage(target)} onClick={() => { if(canSelectStage(target)) onSelectStage(target); }}><AppIcon name={icon} size={21}/><Text>{label}</Text></Button>)}</View>;
  const current = Math.max(0, STAGES.findIndex((item) => item.matches.includes(stage)));
  return <>
    <View className="comic-stage-header">
      <View className="comic-stage-header-copy">
        <Text className="comic-eyebrow">PRIVATE COMIC STUDIO</Text>
        <Text className="comic-stage-title">{title}</Text>
        <Text className="comic-stage-subtitle">{subtitle}</Text>
      </View>
      {hasResult && canSelectStage("result") ? <Text className="comic-stage-result-link" onClick={() => onSelectStage("result")}>查看结果</Text> : null}
    </View>
    <View className="comic-stage-rail">
      {STAGES.map((item, index) => {
        const done = index < current || (item.target === "result" && hasResult);
        const active = item.matches.includes(stage);
        const enabled = canSelectStage(item.target);
        return <View
          key={item.target}
          className={`comic-stage-node ${active ? "is-active" : ""} ${done ? "is-done" : ""} ${enabled ? "" : "is-disabled"} ${index === STAGES.length - 1 ? "is-last" : ""}`}
          onClick={enabled ? () => onSelectStage(item.target) : undefined}
        >
          <Text className="comic-stage-node-dot">{done ? "✓" : index + 1}</Text>
          <Text className="comic-stage-node-label">{item.label}</Text>
        </View>;
      })}
    </View>
  </>;
}
