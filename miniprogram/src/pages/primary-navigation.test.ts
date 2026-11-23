import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, test } from "vitest";

const pagesRoot = resolve(__dirname);

describe("一级页面工作台界面", () => {
  test("工作区、任务页提供统一的移动工作台结构", () => {
    const expectations = [
      ["workspace/index.tsx", "workbench-current-card", "workbench-project-card", "workbench-detail-card"],
      ["tasks/index.tsx", "task-studio-hero", "task-studio-segments", "task-studio-source-error"],
    ];

    for (const [page, ...classNames] of expectations) {
      const source = readFileSync(resolve(pagesRoot, page), "utf8");
      for (const className of classNames) expect(source).toContain(className);
    }
  });

  test("功能页保持统一模块配置与路由", () => {
    const source = readFileSync(resolve(pagesRoot, "functions/index.tsx"), "utf8");

    expect(source).toContain("CONTENT_KEYS");
    expect(source).toContain("PRODUCTION_KEYS");
    expect(source).toContain("TOOL_KEYS");
    expect(source).toContain("moduleRoute(key)");
    expect(source).toContain("activeKey === key");
    expect(source).not.toContain("入口名称与当前真实模块保持一致");
  });

  test("一级页面不再以 Unicode 字符承担功能图标", () => {
    for (const page of ["workspace/index.tsx", "tasks/index.tsx", "functions/index.tsx"]) {
      const source = readFileSync(resolve(pagesRoot, page), "utf8");
      expect(source).not.toMatch(/[✦⚡◎▤▥▦◌›]/);
    }
  });

  test("公共页面不再挂载重复的静态折射贴图", () => {
    for (const page of ["workspace/index.tsx", "tasks/index.tsx", "functions/index.tsx"]) {
      const source = readFileSync(resolve(pagesRoot, page), "utf8");
      expect(source).not.toContain("TextureRefraction");
      expect(source).not.toContain("PageTextureRefraction");
    }
  });
});
