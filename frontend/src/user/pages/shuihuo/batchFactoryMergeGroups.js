function mergeTime(job) {
  return String(job?.updatedAt || job?.createdAt || '');
}

function openingIndex(job) {
  const index = Number(job?.openingVariantIndex || 0);
  return Number.isInteger(index) && index >= 0 ? index : 0;
}

export function openingMergeLabel(index) {
  return index === 0 ? '原开头' : `换开头 ${index}`;
}

export function currentMergeJobsByOpening(mergeJobs = []) {
  const groups = new Map();
  for (const job of mergeJobs) {
    const index = openingIndex(job);
    const group = groups.get(index) || [];
    group.push(job);
    groups.set(index, group);
  }

  const current = [...groups.entries()]
    .sort(([left], [right]) => left - right)
    .map(([index, group]) => {
      const newestFirst = [...group].sort((left, right) => mergeTime(right).localeCompare(mergeTime(left)));
      const selected = newestFirst.find(job => job?.status === 'succeeded' && String(job?.outputUrl || '').trim()) || newestFirst[0];
      return { ...selected, openingVariantIndex: index, openingLabel: openingMergeLabel(index) };
    })
    .filter(Boolean);

  return {
    current,
    historyCount: Math.max(0, mergeJobs.length - current.length)
  };
}

export function publishMergeFiles(bookId, mergeJobs = [], settings = {}, storyboardCount = 0) {
  const configuredCount = Number(settings?.openingCount || 4);
  const requiredCount = settings?.openingEnabled === true && Number(storyboardCount) >= 2
    ? (Number.isInteger(configuredCount) && configuredCount >= 1 && configuredCount <= 8 ? configuredCount : 4)
    : 1;
  const current = currentMergeJobsByOpening(mergeJobs).current;
  const byOpening = new Map(current.map(job => [job.openingVariantIndex, job]));
  const missingOpeningIndexes = [];
  const selected = [];
  for (let index = 0; index < requiredCount; index += 1) {
    const job = byOpening.get(index);
    if (job?.status === 'succeeded' && String(job?.outputUrl || '').trim()) selected.push(job);
    else missingOpeningIndexes.push(index);
  }
  const files = selected.map(job => requiredCount > 1
    ? `${bookId}-opening-${job.openingVariantIndex}.mp4`
    : `${bookId}.mp4`);
  return { ready: missingOpeningIndexes.length === 0, requiredCount, selected, files, missingOpeningIndexes };
}
