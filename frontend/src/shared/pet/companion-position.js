const PET_WIDTH = 120;
const PET_HEIGHT = 130;

export function responsiveCompanionPosition(position, viewportWidth, viewportHeight) {
  if (viewportWidth <= 640) return { left: null, top: null };
  const left = Number(position?.left);
  const top = Number(position?.top);
  if (!Number.isFinite(left) || !Number.isFinite(top)) return { left: null, top: null };
  return {
    left: Math.max(8, Math.min(viewportWidth - PET_WIDTH - 8, left)),
    top: Math.max(58, Math.min(viewportHeight - PET_HEIGHT - 8, top))
  };
}
