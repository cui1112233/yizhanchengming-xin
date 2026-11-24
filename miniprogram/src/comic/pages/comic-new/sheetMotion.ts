/** Keep the closing animation separate from committing a draft. Reopening/unmounting cancels old callbacks. */
export const SHEET_EXIT_MS = 220;
export function createSheetDismissal(onClosing: (value: boolean) => void) {
  let timer: ReturnType<typeof setTimeout> | undefined;
  let locked = false;
  let generation = 0;
  const cancel = () => { if (timer !== undefined) clearTimeout(timer); timer = undefined; generation++; locked = false; };
  return {
    cancel,
    open() { cancel(); onClosing(false); },
    dismiss(reduced: boolean, finish: () => void, commit?: () => void) {
      if (locked) return false;
      locked = true;
      const version = ++generation;
      try { commit?.(); } catch (error) { locked = false; throw error; }
      if (reduced) { finish(); return true; }
      onClosing(true);
      timer = setTimeout(() => { if (version !== generation) return; timer = undefined; onClosing(false); finish(); }, SHEET_EXIT_MS);
      return true;
    },
  };
}
