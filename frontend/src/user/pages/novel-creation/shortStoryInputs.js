// Recovered short-form count rules: defaults, rounding and upper bounds.
export function shortStoryCounts(targetWordCount, targetChapters) {
  const words = Number(targetWordCount), chapters = Number(targetChapters);
  return {
    targetWordCount: !Number.isFinite(words) || words < 1000 ? 13000 : Math.min(500000, Math.round(words)),
    targetChapters: !Number.isFinite(chapters) || chapters < 1 ? 1 : Math.min(50, Math.round(chapters))
  };
}
