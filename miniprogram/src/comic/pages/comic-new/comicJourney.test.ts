import { describe, expect, test } from "vitest";
import { COMIC_EXAMPLE_SOURCE, comicPageReadiness } from "./comicJourney";

describe("comic first creation", () => {
  test("example is a complete story long enough for the existing input gate", () => {
    expect(COMIC_EXAMPLE_SOURCE.length).toBeGreaterThanOrEqual(100);
    expect(COMIC_EXAMPLE_SOURCE.endsWith("。")).toBe(true);
  });
  test("counts actual main images and distinguishes failed pages from waiting", () => {
    expect(comicPageReadiness([
      {primaryImageId:"main",images:[{id:"main"}],status:"completed"},
      {primaryImageId:"",images:[],status:"failed"},
      {primaryImageId:"",images:[],status:"running"},
      {primaryImageId:"missing",images:[{id:"other"}],status:"completed"},
      {primaryImageId:"old",images:[{id:"old"}],status:"failed"},
    ])).toEqual({total:5,ready:2,failed:1,waiting:2,firstFailedIndex:1});
  });
  test("empty or pending result never reports completed pages", () => {
    expect(comicPageReadiness([])).toEqual({total:0,ready:0,failed:0,waiting:0,firstFailedIndex:-1});
  });
});
