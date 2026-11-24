import { afterEach, describe, expect, test, vi } from "vitest";
import { createSheetDismissal, SHEET_EXIT_MS } from "./sheetMotion";

afterEach(() => vi.useRealTimers());
describe("sheet dismissal", () => {
  test("commits once immediately, closes after the exit, ignores repeated taps", () => {
    vi.useFakeTimers(); const closing = vi.fn(); const finish = vi.fn(); const commit = vi.fn();
    const sheet = createSheetDismissal(closing);
    expect(sheet.dismiss(false, finish, commit)).toBe(true);
    expect(sheet.dismiss(false, finish, commit)).toBe(false);
    expect(commit).toHaveBeenCalledTimes(1); expect(finish).not.toHaveBeenCalled();
    vi.advanceTimersByTime(SHEET_EXIT_MS); expect(finish).toHaveBeenCalledTimes(1);
  });
  test("reopening invalidates the old close callback", () => {
    vi.useFakeTimers(); const finish = vi.fn(); const sheet = createSheetDismissal(vi.fn());
    sheet.dismiss(false, finish); sheet.open(); vi.runAllTimers(); expect(finish).not.toHaveBeenCalled();
    expect(sheet.dismiss(true, finish)).toBe(true); expect(finish).toHaveBeenCalledTimes(1);
  });
  test("unmount cancellation cannot call a stale parent", () => {
    vi.useFakeTimers(); const finish = vi.fn(); const sheet = createSheetDismissal(vi.fn());
    sheet.dismiss(false, finish); sheet.cancel(); vi.runAllTimers(); expect(finish).not.toHaveBeenCalled();
  });
  test("reduced motion closes immediately without an exit phase", () => {
    vi.useFakeTimers(); const closing = vi.fn(); const finish = vi.fn(); const sheet = createSheetDismissal(closing);
    sheet.dismiss(true, finish); expect(finish).toHaveBeenCalledTimes(1); expect(closing).not.toHaveBeenCalledWith(true); expect(vi.getTimerCount()).toBe(0);
  });
});
