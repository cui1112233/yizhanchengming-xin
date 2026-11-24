import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { Children, type ReactElement } from "react";
import { describe, expect, test, vi } from "vitest";

const taro = vi.hoisted(() => ({
  getCurrentPages: vi.fn(),
  getMenuButtonBoundingClientRect: vi.fn(),
  getSystemInfoSync: vi.fn(),
  navigateBack: vi.fn(),
  switchTab: vi.fn(),
}));

vi.mock("@tarojs/components", () => ({
  Button: "button",
  Image: "image",
  Input: "input",
  Text: "text",
  View: "view",
}));

vi.mock("@tarojs/taro", () => ({
  default: taro,
}));

import BrandHeader from "./BrandHeader";

const sourceRoot = resolve(__dirname, "..");

describe("小程序品牌工作台视觉", () => {
  test("从历史栈打开的详情页使用系统返回", () => {
    taro.getCurrentPages.mockReturnValue([{ route: "pages/workspace/index" }, { route: "pages/batch/index" }]);
    taro.navigateBack.mockClear();
    taro.switchTab.mockClear();
    const header = BrandHeader({ title: "批量工厂", showBack: true });
    const children = Children.toArray(header.props.children) as ReactElement[];

    children[0].props.onClick();

    expect(taro.navigateBack).toHaveBeenCalledOnce();
    expect(taro.switchTab).not.toHaveBeenCalled();
  });

  test("冷启动详情页时返回工作区而不是停留在空历史栈", () => {
    taro.getCurrentPages.mockReturnValue([{ route: "pages/module-detail/index" }]);
    taro.navigateBack.mockClear();
    taro.switchTab.mockClear();
    const header = BrandHeader({ title: "小说获取", showBack: true });
    const children = Children.toArray(header.props.children) as ReactElement[];

    children[0].props.onClick();

    expect(taro.navigateBack).not.toHaveBeenCalled();
    expect(taro.switchTab).toHaveBeenCalledWith({ url: "/pages/workspace/index" });
  });

  test("页面提供自定义返回时优先交给页面状态机", () => {
    const onBack = vi.fn();
    const header = BrandHeader({ title: "当前小说", showBack: true, onBack });
    const children = Children.toArray(header.props.children) as ReactElement[];
    children[0].props.onClick();
    expect(onBack).toHaveBeenCalledOnce();
    expect(taro.navigateBack).not.toHaveBeenCalled();
  });

  test("账号绑定页展示一个真实返回入口", () => {
    const bindPage = readFileSync(resolve(sourceRoot, "pages/bind/index.tsx"), "utf8");

    expect(bindPage).toMatch(/<BrandHeader[^>]*showBack/);
  });

  test("所有核心页面使用统一品牌顶栏", () => {
    const pages = [
      "pages/workspace/index.tsx",
      "pages/tasks/index.tsx",
      "pages/profile/index.tsx",
      "pages/batch/index.tsx",
      "pages/bind/index.tsx",
    ];

    for (const page of pages) {
      const source = readFileSync(resolve(sourceRoot, page), "utf8");
      expect(source).toContain("BrandHeader");
    }
    const functions = readFileSync(resolve(sourceRoot, "pages/functions/index.tsx"), "utf8");
    expect(functions).toContain("gallery-hero-art");
    expect(functions).toContain("functions-hero-reference-hd.jpg");
  });

  test("品牌顶栏不重复绘制微信系统胶囊", () => {
    const header = readFileSync(resolve(__dirname, "BrandHeader.tsx"), "utf8");
    const styles = readFileSync(resolve(sourceRoot, "app.scss"), "utf8");

    expect(header).not.toContain("mp-capsule");
    expect(styles).not.toContain(".mp-capsule");
  });

  test("详情页返回入口使用统一图标组件", () => {
    const header = readFileSync(resolve(__dirname, "BrandHeader.tsx"), "utf8");

    expect(header).toContain('<AppIcon name="arrow-left"');
    expect(header).not.toContain("‹");
  });

  test("网页参考的顶栏只显示品牌标识和当前栏目", () => {
    taro.getMenuButtonBoundingClientRect.mockReturnValue({ bottom: 64 });
    taro.getSystemInfoSync.mockReturnValue({ statusBarHeight: 24 });
    const header = BrandHeader({ eyebrow: "创作中枢", title: "工作区" });
    const children = Children.toArray(header.props.children) as ReactElement[];

    expect(header.props.style).toEqual({ minHeight: "72px", paddingTop: "24px" });
    expect(children).toHaveLength(1);
    expect(children[0].props.className).toBe("mp-brand-lockup");

    const lockupChildren = Children.toArray(children[0].props.children) as ReactElement[];
    expect(lockupChildren).toHaveLength(3);
    expect(lockupChildren[2].props.children).toBe("工作区");
  });

  test("浅色主题使用深色品牌标识，而不是把白色标识放在浅色背景上", () => {
    const header = readFileSync(resolve(__dirname, "BrandHeader.tsx"), "utf8");
    const styles = readFileSync(resolve(sourceRoot, "app.scss"), "utf8");

    expect(header).toContain("brand-logo-black.png");
    expect(header).toContain("mp-brand-logo-light");
    expect(styles).toContain(".mp-theme-light .mp-brand-logo-dark { display: none; }");
    expect(styles).toContain(".mp-theme-light .mp-brand-logo-light { display: block; }");
  });

  test("保留网页端的深色表面与绿色强调", () => {
    const styles = readFileSync(resolve(sourceRoot, "app.scss"), "utf8");

    expect(styles).toContain(".mp-brand-header");
    expect(styles).toContain("#38c77a");
    expect(styles).toContain("#191c20");
  });

  test("背景样式不使用开发者工具无法编译的遮罩属性", () => {
    const styles = readFileSync(resolve(sourceRoot, "app.scss"), "utf8");

    expect(styles).not.toContain("mask-image");
  });

  test("背景层不依赖受限的径向渐变语法", () => {
    const styles = readFileSync(resolve(sourceRoot, "app.scss"), "utf8");

    expect(styles).not.toContain("radial-gradient");
  });

  test("页面背景不依赖伪元素装饰层", () => {
    const styles = readFileSync(resolve(sourceRoot, "app.scss"), "utf8");

    expect(styles).not.toContain(".mp-page::before");
  });

  test("视觉结构对齐网页的左对齐品牌与星空层", () => {
    const styles = readFileSync(resolve(sourceRoot, "app.scss"), "utf8");

    expect(styles).toContain("#090b0d");
    expect(styles).toContain(".mp-ambient-starfield");
    expect(styles).toContain("@keyframes mp-twinkle");
    expect(styles).toMatch(/\.mp-brand-header[\s\S]*justify-content:\s*flex-start/);
    expect(styles).not.toContain("filter: blur");
    expect(styles).not.toContain(".mp-page >");
  });

  test("个人中心滚动区随自定义顶栏剩余空间伸缩", () => {
    const styles = readFileSync(resolve(sourceRoot, "app.scss"), "utf8");

    expect(styles).toMatch(/\.mp-profile-page \{[^}]*display:\s*flex/);
    expect(styles).toMatch(/\.mp-profile-page \{[^}]*height:\s*100vh/);
    expect(styles).toMatch(/\.mp-profile-page \{[^}]*flex-direction:\s*column/);
    expect(styles).toMatch(/\.mp-profile-scroll \{[^}]*flex:\s*1/);
    expect(styles).toMatch(/\.mp-profile-scroll \{[^}]*min-height:\s*0/);
    expect(styles).not.toContain("height: calc(100vh - 150rpx)");
  });

  test("任务页使用独立紧凑布局，避免旧工作台样式覆盖", () => {
    const source = readFileSync(resolve(sourceRoot, "pages/tasks/index.tsx"), "utf8");

    expect(source).not.toContain("studio-task-record");
    expect(source).not.toContain("studio-task-action");
    expect(source).toContain("task-studio-row-state");
    expect(source).toContain("mp-status-pill");
  });
});
