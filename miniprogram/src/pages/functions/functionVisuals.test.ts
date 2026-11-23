import { expect, test } from "vitest";
import { FUNCTION_VISUALS } from "./functionVisuals";

test("provides one dedicated illustration for every function entry", () => {
  expect(Object.keys(FUNCTION_VISUALS)).toEqual([
    "script", "fetch", "batch", "shuihuo", "comic", "aicut", "agent", "voice", "plugins",
  ]);
  expect(new Set(Object.values(FUNCTION_VISUALS).map(item => item.asset)).size).toBe(9);
  expect(FUNCTION_VISUALS.script.asset).toContain("private-custom-reference-3d.webp");
  expect(FUNCTION_VISUALS.comic.asset).toContain("comic-production-reference-3d.webp");
});
