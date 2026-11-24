import { describe, expect, it } from "vitest";
import type { TaskSummary } from "./taskSummary";
import { deriveTaskPageState } from "./taskPageState";

const readySummary: TaskSummary = {
  entries: [{
    source: "script",
    sourceLabel: "剧本",
    id: "script-1",
    title: "第一集",
    createdAt: "2026-10-13T08:00:00.000Z",
    rawStatus: "completed",
    status: "completed",
    progress: undefined,
    statusDetail:"已完成",visualHint:undefined,moduleKey:"script",kind:undefined,stage:undefined,taskType:undefined,retryable:false,
  }],
  sources: [
    { source: "script", sourceLabel: "剧本", state: "ready", error: undefined },
    { source: "agent", sourceLabel: "Agent Studio", state: "ready", error: undefined },
    { source: "fetch", sourceLabel: "小说获取 / 批量生产", state: "ready", error: undefined },
  ],
};

describe("任务页状态", () => {
  it("把未配置服务端与空任务区分开", () => {
    expect(deriveTaskPageState({ loading: false, summary: null, configurationError: "服务端地址未配置" }))
      .toEqual({ kind: "configuration-error", message: "服务端地址未配置" });
  });

  it("在一个来源失败时保留其他真实任务", () => {
    const partial: TaskSummary = {
      ...readySummary,
      sources: readySummary.sources.map((item) => item.source === "agent"
        ? { ...item, state: "error" as const, error: "Agent 服务暂不可用" }
        : item),
    };

    expect(deriveTaskPageState({ loading: false, summary: partial, configurationError: "" }))
      .toMatchObject({ kind: "content", failedSourceCount: 1, entryCount: 1 });
  });

  it("只在所有来源成功且无条目时显示真实空状态", () => {
    expect(deriveTaskPageState({
      loading: false,
      summary: { ...readySummary, entries: [] },
      configurationError: "",
    })).toEqual({ kind: "empty", failedSourceCount: 0, entryCount: 0 });
  });
});
