function shotValue(value) {
  return {
    sceneIndex: Number(value?.scene_index ?? value?.sceneIndex ?? 1),
    shotIndex: Number(value?.shot_index ?? value?.shotIndex ?? 1),
    prompt: typeof value?.prompt === 'string' ? value.prompt : typeof value?.video_prompt === 'string' ? value.video_prompt : '',
    sourceText: typeof value?.source_text === 'string' ? value.source_text : ''
  };
}

export function buildOutlineViewModel(value, previous = { scenes: [], dirty: false }) {
  const rawShots = Array.isArray(value?.outline_shots) ? value.outline_shots : null;
  if (!rawShots) return previous;
  const groups = new Map();
  rawShots.map(shotValue).forEach(shot => {
    if (!groups.has(shot.sceneIndex)) groups.set(shot.sceneIndex, []);
    groups.get(shot.sceneIndex).push(shot);
  });
  const scenes = [...groups.entries()].sort((a, b) => a[0] - b[0]).map(([index, shots]) => ({
    index,
    shots: shots.sort((a, b) => a.shotIndex - b.shotIndex)
  }));
  return { scenes, dirty: false };
}
