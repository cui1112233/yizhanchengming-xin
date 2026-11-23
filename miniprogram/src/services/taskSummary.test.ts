import { describe, expect, it } from "vitest";
import { filterTaskEntries, loadTaskSummary, taskCounts, taskStatusFromServer } from "./taskSummary";

type ResponseMap = Record<string, unknown | Error>;

function requestFrom(responses: ResponseMap) {
  return async (path: string) => {
    const response = responses[path];
    if (response instanceof Error) throw response;
    return response;
  };
}

function source(summary: Awaited<ReturnType<typeof loadTaskSummary>>, name: string) {
  return summary.sources.find((item) => item.source === name);
}

describe("真实任务汇总", () => {
  it("只把明确需要人工处理的状态归为待确认", () => {
    expect(["waiting-review", "pending_confirmation", "accepted_pending", "待确认", "待审核"].map(taskStatusFromServer)).toEqual([
      "attention", "attention", "attention", "attention", "attention",
    ]);
    expect(["queued", "pending", "waiting"].map(taskStatusFromServer)).toEqual(["running", "running", "running"]);
  });

  it("待确认筛选和顶部统计使用同一批真实任务", () => {
    const entries = [
      { status: "running", id: "run" },
      { status: "attention", id: "review" },
      { status: "completed", id: "done" },
      { status: "failed", id: "failed" },
      { status: "unknown", id: "unknown" },
    ] as Parameters<typeof filterTaskEntries>[0];
    expect(filterTaskEntries(entries, "待确认").map((entry) => entry.id)).toEqual(["review"]);
    expect(taskCounts(entries)).toEqual({ running: 1, attention: 1, completed: 1 });
  });

  it("小说获取阶段状态进入正确分类，同时失败记录保留在全部任务", async () => {
    const summary = await loadTaskSummary(requestFrom({
      "/api/batch-rewrite/tasks": { tasks: [
        { id: "1", status: "ai_processing" },
        { id: "2", status: "original_failed" },
        { id: "3", status: "ai_failed" },
      ] },
    }));
    expect(filterTaskEntries(summary.entries, "进行中").map(row => row.id)).toEqual(["1"]);
    expect(filterTaskEntries(summary.entries, "全部").filter(row => row.status === "failed").map(row => row.id)).toEqual(["2", "3"]);
    expect(summary.entries.map(row => row.rawStatus)).toEqual(["ai_processing", "original_failed", "ai_failed"]);
  });
  it("汇总服务端来源并保留真实状态、时间和视频进度", async () => {
    const summary = await loadTaskSummary(requestFrom({
      "/api/history": {
        entries: [{
          id: "script-2",
          preview: "第 2 集",
          updatedAt: "2026-10-13T08:00:00.000Z",
          videoTasks: {
            "0": { status: "completed" },
            "1": { status: "running" },
          },
        }],
      },
      "/api/studio/runs": {
        runs: [{ id: "agent-1", title: "修订对白", status: "completed", created_at: "2026-10-12T08:00:00.000Z" }],
      },
      "/api/batch-rewrite/tasks": {
        tasks: [
          { id: "fetch-1", bookName: "原著 A", status: "failed", createdAt: "2026-10-11T08:00:00.000Z" },
          { id: "batch-1", name: "漫剧批次", status: "waiting-review", createdAt: "2026-10-10T08:00:00.000Z" },
        ],
      },
    }));

    expect(summary.entries.map((entry) => [entry.source, entry.title, entry.status])).toEqual([
      ["script", "第 2 集", "running"],
      ["agent", "修订对白", "completed"],
      ["fetch", "原著 A", "failed"],
      ["fetch", "漫剧批次", "attention"],
    ]);
    expect(summary.entries[0]).toMatchObject({
      createdAt: "2026-10-13T08:00:00.000Z",
      progress: { completed: 1, total: 2 },
    });
    expect(source(summary, "fetch")).toMatchObject({ state: "ready" });
  });

  it("保留成功来源，并将单个请求失败限定在所属来源", async () => {
    const summary = await loadTaskSummary(requestFrom({
      "/api/history": { entries: [{ id: "script-1", preview: "第一集", status: "completed" }] },
      "/api/studio/runs": new Error("Agent 服务暂不可用"),
      "/api/batch-rewrite/tasks": { tasks: [{ id: "fetch-1", bookName: "原著 A", status: "running" }] },
    }));

    expect(summary.entries.map((entry) => entry.source)).toEqual(["script", "fetch"]);
    expect(source(summary, "agent")).toMatchObject({ state: "error", error: "Agent 服务暂不可用" });
    expect(source(summary, "script")).toMatchObject({ state: "ready" });
    expect(source(summary, "fetch")).toMatchObject({ state: "ready" });
  });

  it("不把缺失字段伪造成完成任务、标题、时间或进度", async () => {
    const summary = await loadTaskSummary(requestFrom({
      "/api/history": { entries: [{ id: "script-unknown" }] },
      "/api/studio/runs": { runs: [] },
      "/api/batch-rewrite/tasks": { tasks: [] },
    }));

    expect(summary.entries).toEqual([{
      source: "script",
      sourceLabel: "剧本",
      id: "script-unknown",
      title: undefined,
      createdAt: undefined,
      rawStatus: undefined,
      status: "unknown",
      progress: undefined,
    }]);
    expect(filterTaskEntries(summary.entries, "已完成")).toEqual([]);
    expect(filterTaskEntries(summary.entries, "全部")).toHaveLength(1);
  });
});
