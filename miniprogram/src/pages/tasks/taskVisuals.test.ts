import { describe, expect, it } from "vitest";
import type { AggregatedTask, TaskSource } from "../../services/taskSummary";
import { taskVisualAsset } from "./taskVisuals";
import privateCustom from "../../assets/functions/private-custom-reference-3d.png";
import storyboardFilm from "../../assets/functions/comic-production-reference-3d.png";

function task(source: TaskSource, title: string, visualHint?: string, patch: Partial<AggregatedTask> = {}): AggregatedTask {
  return { source, sourceLabel: source, id: title, title, createdAt: undefined, rawStatus: "running", status: "running", statusDetail: "进行中", progress: undefined, visualHint, moduleKey: source, kind: visualHint, stage: undefined, taskType: undefined, retryable: false, ...patch };
}

describe("任务图标对应", () => {
  it("优先使用真实类型和明确标题词选择不同的功能图标", () => {
    const assets = [
      taskVisualAsset(task("batch", "人物场景提取", "image")),
      taskVisualAsset(task("fetch", "批量续写 · 仙侠篇")),
      taskVisualAsset(task("shuihuo", "漫画自动生产", "comic")),
      taskVisualAsset(task("shuihuo", "分镜生成 · 第 8 集", "storyboard")),
      taskVisualAsset(task("shuihuo", "成片导出 · 竖屏版", "export")),
    ];
    expect(new Set(assets).size).toBe(5);
  });

  it("没有类型提示时按真实来源使用不同的回退图标", () => {
    const assets = (["script", "agent", "fetch", "batch", "shuihuo"] as TaskSource[]).map((source) => taskVisualAsset(task(source, "普通任务")));
    expect(new Set(assets).size).toBe(5);
  });

  it("所属功能优先于容易误导的标题词，功能内部再按真实阶段细分", () => {
    expect(taskVisualAsset(task("script", "批量生成剧本"))).toBe(privateCustom);
    expect(taskVisualAsset(task("shuihuo", "第八集", undefined, { kind: "storyboard", stage: "storyboard" }))).toBe(storyboardFilm);
  });
});
