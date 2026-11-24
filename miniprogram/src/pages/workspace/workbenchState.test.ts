import { describe, expect, test } from "vitest";
import type { AggregatedTask, TaskSource, TaskStatus } from "../../services/taskSummary";
import {
  selectCurrentTask,
  taskSourceForModule,
  workbenchProgress,
  workbenchStatus,
} from "./workbenchState";

function task(source: TaskSource, id: string, overrides: Partial<AggregatedTask> = {}): AggregatedTask {
  return {
    source,
    sourceLabel: source,
    id,
    title: `${source}-${id}`,
    createdAt: "2026-11-23T20:47:00+08:00",
    rawStatus: "running",
    status: "running" as TaskStatus,
    progress: undefined,
    statusDetail:"",visualHint:undefined,moduleKey:source,kind:undefined,stage:undefined,taskType:undefined,retryable:false,
    ...overrides,
  };
}

describe("workspace current task state", () => {
  test("maps only modules backed by a real task source", () => {
    expect(taskSourceForModule("script")).toBe("script");
    expect(taskSourceForModule("fetch")).toBe("fetch");
    expect(taskSourceForModule("batch")).toBe("batch");
    expect(taskSourceForModule("shuihuo")).toBe("shuihuo");
    expect(taskSourceForModule("agent")).toBe("agent");
    expect(taskSourceForModule("comic")).toBeUndefined();
  });

  test("selects the newest task only from the current module source", () => {
    const entries = [task("fetch", "other"), task("script", "current"), task("script", "older")];
    expect(selectCurrentTask("script", entries)?.id).toBe("current");
    expect(selectCurrentTask("comic", entries)).toBeUndefined();
  });

  test("water production resumes either real production source in newest-first order", () => {
    expect(selectCurrentTask("shuihuo", [task("script", "other"), task("batch", "batch-work"), task("shuihuo", "manual")])?.id).toBe("batch-work");
    expect(selectCurrentTask("shuihuo", [task("shuihuo", "manual"), task("batch", "batch-work")])?.id).toBe("manual");
  });

  test("does not invent progress when the server did not provide totals", () => {
    expect(workbenchProgress(task("script", "a"))).toBeUndefined();
    expect(workbenchProgress(task("script", "b", { progress: { completed: 3, total: 4 } })))
      .toEqual({ completed: 3, total: 4, percent: 75 });
    expect(workbenchProgress(task("script", "c", { progress: { completed: 2, total: 0 } }))).toBeUndefined();
  });

  test("derives concise UI status from the real task state", () => {
    expect(workbenchStatus()).toEqual({ label: "待开始", tone: "idle" });
    expect(workbenchStatus(task("script", "a"))).toEqual({ label: "执行中", tone: "running" });
    expect(workbenchStatus(task("script", "b", { status: "retry" }))).toEqual({ label: "需重试", tone: "failed" });
    expect(workbenchStatus(task("script", "e", { status: "attention",statusDetail:'待确认' }))).toEqual({label:'待确认',tone:'idle'});
    expect(workbenchStatus(task("script", "c", { status: "completed" }))).toEqual({ label: "已完成", tone: "completed" });
    expect(workbenchStatus(task("script", "d", { status: "unknown" }))).toEqual({ label: "状态同步中", tone: "unknown" });
  });
});
