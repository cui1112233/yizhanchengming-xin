import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { expect, test } from "vitest";

const pageRoot = resolve(__dirname, "..");

test("账号绑定页为账号凭证输入显示本地语义图标", () => {
  const source = readFileSync(resolve(pageRoot, "../components/LoginScene.tsx"), "utf8");

  expect(source).toContain('<AppIcon name="user-round"');
  expect(source).toContain('<AppIcon name="key-round"');
});
