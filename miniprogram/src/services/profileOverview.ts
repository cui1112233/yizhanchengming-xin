import { formatTokenAmount, profileFacts, type PublicMemberCenter } from "../config/profileData";
import { taskCounts, type TaskSummary } from "./taskSummary";

export type ProfileTaskMetrics = {
  running: number | undefined;
  completed: number | undefined;
};

export function profileTaskMetrics(summary: TaskSummary | undefined): ProfileTaskMetrics {
  if (!summary) return { running: undefined, completed: undefined };
  const counts = taskCounts(summary.entries);
  return { running: counts.running, completed: counts.completed };
}

export function profileQuota(center: PublicMemberCenter | null | undefined) {
  if (!center) return { value: "—", caption: "创作额度暂不可用" };

  const facts = profileFacts(center);
  const usage = facts.monthlyTokenUsage;
  const limit = facts.monthlyTokenLimit;
  if (usage === undefined && limit === undefined) {
    return { value: "—", caption: "创作额度暂不可用" };
  }
  if (limit === null) {
    return {
      value: formatTokenAmount(usage),
      caption: "本月已用 · 不限额",
    };
  }
  if (usage !== undefined && limit !== undefined) {
    return {
      value: `${formatTokenAmount(usage)} / ${formatTokenAmount(limit)}`,
      caption: "本月创作额度",
    };
  }
  return {
    value: formatTokenAmount(usage ?? limit),
    caption: usage === undefined ? "本月创作额度" : "本月已用",
  };
}
