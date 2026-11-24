import { describe, expect, it } from "vitest";
import { MODULES, moduleRoute } from "./modules";

describe("小程序功能目录", () => {
  it("与网页功能目录保持一致", () => {
    expect(MODULES.map((item) => item.key)).toEqual([
      "script",
      "novel",
      "booktest",
      "shuihuo",
      "comic",
      "aicut",
      "agent",
      "voice",
      "plugins",
    ]);
    expect(MODULES.map((item) => item.label)).toEqual([
      "私人定制",
      "小说创作",
      "批量测书",
      "水货生产",
      "漫画生成",
      "AI-CUT",
      "CM 画布",
      "配音",
      "插件",
    ]);
  });

  it("三个分组各保留三个一级功能", () => {
    expect(MODULES.filter((item) => item.group === "创作")).toHaveLength(3);
    expect(MODULES.filter((item) => item.group === "生产")).toHaveLength(3);
    expect(MODULES.filter((item) => item.group === "助手")).toHaveLength(3);
  });

  it("使用网页对应的语义图标，而不是页面内的字符图标", () => {
    expect(MODULES.map((item) => item.icon)).toEqual([
      "file-pen-line",
      "book-open",
      "factory",
      "clapperboard",
      "sparkles",
      "sliders",
      "bot",
      "volume",
      "settings",
    ]);
  });

  it("功能页只展示线上真实入口，不暴露内部小说面板", () => {
    expect(MODULES.some((item) => String(item.key) === "panel")).toBe(false);
    expect(MODULES.some((item) => item.label === "小说面板")).toBe(false);
  });

  it("每个一级功能都有唯一的小程序入口路由", () => {
    expect(moduleRoute("script")).toBe("/pages/module-detail/index?key=script");
    expect(moduleRoute("agent")).toBe("/pages/module-detail/index?key=agent");
    expect(moduleRoute("comic")).toBe("/comic/pages/comic-new/index");
    expect(moduleRoute("batch")).toBe("/pages/batch/index");
    expect(moduleRoute("voice")).toBe("/comic/pages/tts/index");
    expect(moduleRoute("plugins")).toBe("/comic/pages/plugins/index");
  });
});
