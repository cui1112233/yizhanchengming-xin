import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, test } from "vitest";

const root = resolve(__dirname, "../..");

describe("剧本工作台网页视觉适配", () => {
  test("保留网页工作台的控制台层级与移动端安全底部留白", () => {
    const styles = readFileSync(resolve(root, "app.scss"), "utf8");

    expect(styles).toContain(".mp-script-console");
    expect(styles).toContain(".mp-script-control-strip");
    expect(styles).toContain(".mp-script-page { padding-bottom: calc(72rpx + env(safe-area-inset-bottom)); }");
  });

  test("工作区使用网页实际存在的本地图标，而不是字符图标", () => {
    const iconByPage = {
      "ScriptWorkspace.tsx": "file-pen-line",
      "PanelWorkspace.tsx": "notebook-tabs",
      "WaterLibrary.tsx": "clapperboard",
      "AgentWorkspace.tsx": "bot",
      "../batch/index.tsx": "factory",
    };

    for (const [page, icon] of Object.entries(iconByPage)) {
      const source = readFileSync(resolve(__dirname, page), "utf8");
      expect(source).toContain("AppIcon");
      expect(source).toContain(`name=\"${icon}\"`);
      expect(source).not.toMatch(/[✦⚡◎▤▥▦♧✣◷⌘◖▣›]/);
    }

    const fetchWorkspace = readFileSync(resolve(__dirname, "FetchWorkspace.tsx"), "utf8");
    expect(fetchWorkspace).toContain("mp-fetch-hero");
    expect(fetchWorkspace).toContain('name="sparkles"');
    expect(fetchWorkspace).not.toMatch(/[✦⚡◎▤▥▦♧✣◷⌘◖▣›]/);
  });
});
