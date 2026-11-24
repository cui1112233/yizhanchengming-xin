import { describe, expect, test } from "vitest";
import appConfig from "./app.config";

describe("小程序页面配置", () => {
  test("使用自定义导航区承载网页同款品牌锁定", () => {
    expect(appConfig.window?.navigationStyle).toBe("custom");
  });

  test("声明四个原生 Tab 和批量工厂、绑定页面", () => {
    expect(appConfig.entryPagePath).toBe("pages/functions/index");
    expect(appConfig.pages).toEqual(expect.arrayContaining([
      "pages/workspace/index",
      "pages/tasks/index",
      "pages/functions/index",
      "pages/profile/index",
      "pages/batch/index",
      "pages/bind/index",
    ]));
    expect(appConfig.tabBar?.list.map((item) => item.pagePath)).toEqual([
      "pages/functions/index",
      "pages/workspace/index",
      "pages/tasks/index",
      "pages/profile/index",
    ]);
    expect(appConfig.tabBar?.list.map((item) => item.text)).toEqual(["功能", "工作台", "任务", "我的"]);
  });

  test("四个一级 Tab 均使用本地普通态和选中态图标", () => {
    for (const item of appConfig.tabBar?.list || []) {
      expect(item.iconPath).toMatch(/^assets\/icons\/tab-/);
      expect(item.selectedIconPath).toMatch(/^assets\/icons\/tab-/);
    }
  });

  test("新增功能页全部位于 comic 分包，主包页面列表不扩大", () => {
    const subPackage = appConfig.subPackages?.find((item) => item.root === "comic");
    expect(subPackage?.pages).toEqual([
      "pages/module-detail/index",
      "pages/index/index",
      "pages/aicut/index",
      "pages/novel/index",
      "pages/plugins/index",
      "pages/member-center/index",
      "pages/usage-stats/index",
      "pages/tts/index",
      "pages/comic-new/index",
      "pages/comic-config/index",
      "pages/comic-task/index",
      "pages/comic-result/index",
    ]);
    expect(appConfig.pages.some((page) => page.startsWith("comic/"))).toBe(false);
  });
});
