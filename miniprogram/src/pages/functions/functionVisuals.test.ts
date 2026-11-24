import { expect, test } from "vitest";
import { FUNCTION_VISUALS } from "./functionVisuals";
import {readFileSync} from 'node:fs';
import {resolve,basename} from 'node:path';

test("provides one dedicated illustration for every function entry", () => {
  expect(Object.keys(FUNCTION_VISUALS)).toEqual([
    "script", "fetch", "novel", "batch", "booktest", "shuihuo", "comic", "aicut", "agent", "voice", "plugins",
  ]);
  expect(new Set(Object.values(FUNCTION_VISUALS).map(item => item.asset)).size).toBe(9);
  expect(FUNCTION_VISUALS.script.asset).toContain("private-custom-reference-3d.png");
  expect(FUNCTION_VISUALS.comic.asset).toContain("comic-production-reference-3d.png");
});

test('bundled phone entry illustrations use actual PNG data rather than local WebP',()=>{
  for(const asset of new Set(Object.values(FUNCTION_VISUALS).map(item=>item.asset))){
    expect(asset).toMatch(/\.png$/);
    const bytes=readFileSync(resolve(__dirname,'../../assets/functions',basename(asset)));
    expect(bytes.subarray(0,8).toString('hex')).toBe('89504e470d0a1a0a');
  }
});
