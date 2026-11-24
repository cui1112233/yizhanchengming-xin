import { describe, expect, test } from "vitest";
import { profileQuota, profileTaskMetrics } from "./profileOverview";

describe("profileTaskMetrics", () => {
  test("uses the real task statuses instead of decorative profile numbers", () => {
    const summary = {
      entries: [
        { id: "a", status: "running" },
        { id: "b", status: "running" },
        { id: "c", status: "retry" },
        { id: "d", status: "completed" },
        { id: "e", status: "completed" },
        { id: "f", status: "completed" },
      ],
    } as any;

    expect(profileTaskMetrics(summary)).toEqual({ running: 2, completed: 3 });
  });

  test("keeps unavailable task data distinct from a real zero", () => {
    expect(profileTaskMetrics(undefined)).toEqual({ running: undefined, completed: undefined });
    expect(profileTaskMetrics({ entries: [] } as any)).toEqual({ running: 0, completed: 0 });
  });
});

describe("profileQuota", () => {
  test("shows real monthly usage with an unlimited allowance", () => {
    expect(profileQuota({ member: { monthlyTokenLimit: null }, usage: { month: { totalTokens: 1250 } } })).toEqual({
      value: "1,250",
      caption: "本月已用 · 不限额",
    });
  });

  test("does not turn a real zero or missing data into a fake quota", () => {
    expect(profileQuota({ member: { monthlyTokenLimit: 5000 }, usage: { month: { totalTokens: 0 } } })).toEqual({
      value: "0 / 5,000",
      caption: "本月创作额度",
    });
    expect(profileQuota(undefined)).toEqual({ value: "—", caption: "创作额度暂不可用" });
  });
});
