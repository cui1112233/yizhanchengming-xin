import { describe, expect, it } from "vitest";
import {
  buildFetchProcessPayload,
  fetchTaskOverallStatus,
  fetchTaskStages,
  fetchWorkspaceState,
  normalizeFetchTask,
  taskDateKey,
} from "./fetchWorkspace";

describe("小说获取工作台状态", () => {
  it("沿用可用网页版的规则覆盖顺序，显式关闭优先于全局开启且合并各 AI 槽位", () => {
    const state = fetchWorkspaceState({ app_config: { sensitive_ai: { enabled: true }, rewrite: { ai_slot_methods: { ai1: "rewrite", ai2: "polish" } } }, work_form: { sensitive_ai_enabled: false, ai_slot_methods: { ai2: "rewrite" } } }, []);
    expect(state.sensitiveAiEnabled).toBe(false);
    expect(state.aiSlotMethods).toEqual({ ai1: "rewrite", ai2: "rewrite" });
  });
  it("服务端没有平台时不臆造平台或允许提交", () => {
    expect(fetchWorkspaceState({}, []).platformId).toBe("");
    expect(() => buildFetchProcessPayload({ input: "1001", platformId: "", maxTxt: 4000, selectedVersions: ["original"], aiSlotMethods: {}, sensitiveAiEnabled: false })).toThrow(/平台/);
  });
  it("把网页端的真实配置收敛成可选择的平台、模型和版本设置", () => {
    expect(fetchWorkspaceState({
      platforms: [
        { id: 2, name: "黑岩付费", visible: true },
        { id: 3, name: "隐藏平台", visible: false },
      ],
      app_config: { text_model_id: "text-2", fetch: { default_max_txt: 6000 } },
      work_form: {
        selected_versions: ["original", "ai2"],
        ai_slot_methods: { ai2: "rewrite" },
        sensitive_ai_enabled: true,
      },
    }, [
      { id: "text-1", displayName: "基础模型" },
      { id: "text-2", name: "创作模型" },
    ])).toEqual({
      platforms: [{ id: "2", name: "黑岩付费" }],
      platformId: "2",
      models: [
        { id: "text-1", name: "基础模型" },
        { id: "text-2", name: "创作模型" },
      ],
      textModelId: "text-2",
      maxTxt: 6000,
      selectedVersions: ["original", "ai2"],
      aiSlotMethods: { ai2: "rewrite" },
      sensitiveAiEnabled: true,
    });
  });

  it("处理请求携带当前工作台的真实配置而不是写死默认值", () => {
    expect(buildFetchProcessPayload({
      input: "1001\n1002",
      platformId: "2",
      maxTxt: 6000,
      selectedVersions: ["original", "ai2"],
      aiSlotMethods: { ai2: "rewrite" },
      sensitiveAiEnabled: true,
    })).toEqual({
      input_text: "1001\n1002",
      platform_id: "2",
      max_txt: 6000,
      selected_versions: ["original", "ai2"],
      ai_slot_methods: { ai2: "rewrite" },
      sensitive_ai_enabled: true,
    });
  });

  it("把网页端任务字段规范化为可筛选的阶段状态", () => {
    const task = normalizeFetchTask({
      book_id: "1001",
      book_name: "测试小说",
      original_status: "done",
      ai_generated_count: 1,
      selected_versions: ["original", "ai1"],
      site_submit_status: "pending",
      created_at: "2026-10-13T12:00:00",
    });

    expect(task).toMatchObject({
      bookId: "1001",
      bookName: "测试小说",
      originalStatus: "done",
      aiGeneratedCount: 1,
    });
    expect(taskDateKey(task)).toBe("2026-10-13");
    expect(fetchTaskStages(task)).toEqual([
      { key: "classify", label: "分类", state: "pending" },
      { key: "original", label: "原文", state: "done" },
      { key: "ai", label: "AI", state: "done" },
      { key: "submit", label: "提交", state: "active" },
    ]);
    expect(fetchTaskOverallStatus(task)).toBe("AI完成");
  });

  it("保留服务端已有的分类和 AI 文件事实，不把已完成任务误判为等待", () => {
    const task = normalizeFetchTask({
      id: "book-2",
      gender: "男频",
      style: "都市",
      ai_files: [{ version: "ai1" }, { version: "ai2" }],
      selected_versions: ["original", "ai1", "ai2"],
    });

    expect(task.gender).toBe("男频");
    expect(task.style).toBe("都市");
    expect(task.aiGeneratedCount).toBe(2);
    expect(fetchTaskStages(task)).toEqual(expect.arrayContaining([
      { key: "classify", label: "分类", state: "done" },
      { key: "ai", label: "AI", state: "done" },
    ]));
  });
});
