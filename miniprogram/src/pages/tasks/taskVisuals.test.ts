import { describe, expect, it } from "vitest";
import type { AggregatedTask, TaskSource } from "../../services/taskSummary";
import { taskVisualAsset } from "./taskVisuals";

function task(source: TaskSource, title: string, visualHint?: string): AggregatedTask {
  return { source, sourceLabel: source, id: title, title, createdAt: undefined, rawStatus: undefined, status: "running", progress: undefined, visualHint };
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
});
