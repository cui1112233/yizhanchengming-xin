import type { ModuleKey } from "../../config/modules";
import type { AggregatedTask, TaskSource } from "../../services/taskSummary";

export type WorkbenchStatusTone = "idle" | "running" | "failed" | "completed" | "unknown";

const MODULE_SOURCES: Partial<Record<ModuleKey, TaskSource>> = {
  script: "script",
  fetch: "fetch",
  booktest: "fetch",
  batch: "batch",
  shuihuo: "shuihuo",
  agent: "agent",
  novel:"novel",
  aicut:"aicut",
};

export function taskSourceForModule(key: ModuleKey) {
  return MODULE_SOURCES[key];
}

export function selectCurrentTask(key: ModuleKey, entries: AggregatedTask[]) {
  const source=taskSourceForModule(key);
  const matches=entries.filter(entry=>key==='shuihuo'?(entry.source==='shuihuo'||entry.source==='batch'):entry.source===source);
  return matches.find(entry=>entry.status==='running')||matches.find(entry=>entry.status==='attention')||matches.find(entry=>entry.status==='retry')||matches[0];
}

export function workbenchProgress(task?: AggregatedTask) {
  const progress = task?.progress;
  if (!progress || progress.total <= 0) return undefined;
  return {
    completed: progress.completed,
    total: progress.total,
    percent: Math.round((progress.completed / progress.total) * 100),
  };
}

export function workbenchStatus(task?: AggregatedTask): { label: string; tone: WorkbenchStatusTone } {
  if (!task) return { label: "待开始", tone: "idle" };
  if (task.status === "running") return { label: "执行中", tone: "running" };
  if (task.status === "retry") return { label: "需重试", tone: "failed" };
  if (task.status === "attention") return {label:task.statusDetail||"待确认",tone:"idle"};
  if (task.status === "completed") return { label: "已完成", tone: "completed" };
  return { label: "状态同步中", tone: "unknown" };
}
