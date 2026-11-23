import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";
import { expect, test } from "vitest";

test("提供统一的本地图标组件供页面和工作区复用", () => {
  expect(existsSync(resolve(__dirname, "AppIcon.tsx"))).toBe(true);
});

test("任务中心只使用已经注册的本地图标", () => {
  const component = readFileSync(resolve(__dirname, "AppIcon.tsx"), "utf8");
  const page = readFileSync(resolve(__dirname, "../pages/tasks/index.tsx"), "utf8");
  const registered = new Set(Array.from(component.matchAll(/\"([a-z0-9-]+)\"/g), match => match[1]));
  const used = Array.from(page.matchAll(/<AppIcon[^>]+name=\"([a-z0-9-]+)\"/g), match => match[1]);
  expect(used.filter(name => !registered.has(name))).toEqual([]);
});
