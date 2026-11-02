function finiteNumber(value) {
  const number = Number(value);
  return Number.isFinite(number) ? number : null;
}

function nodeSize(node, nodeSizes) {
  const measured = node?.measured || {};
  const fallback = nodeSizes?.[node?.id] || {};
  return {
    width: finiteNumber(measured.width) || finiteNumber(node?.width) || finiteNumber(fallback.width) || 280,
    height: finiteNumber(measured.height) || finiteNumber(node?.height) || finiteNumber(fallback.height) || 180
  };
}

export function shouldRecoverCanvasViewport(nodes = [], viewport, canvasSize, nodeSizes = {}) {
  if (!Array.isArray(nodes) || nodes.length === 0 || !viewport || !canvasSize) return false;
  const viewportX = finiteNumber(viewport.x);
  const viewportY = finiteNumber(viewport.y);
  const zoom = finiteNumber(viewport.zoom);
  const canvasWidth = finiteNumber(canvasSize.width);
  const canvasHeight = finiteNumber(canvasSize.height);
  if (viewportX === null || viewportY === null || !zoom || !canvasWidth || !canvasHeight) return false;

  const hasVisibleNode = nodes.some(node => {
    const nodeX = finiteNumber(node?.position?.x);
    const nodeY = finiteNumber(node?.position?.y);
    if (nodeX === null || nodeY === null || node?.hidden) return false;
    const size = nodeSize(node, nodeSizes);
    const left = viewportX + nodeX * zoom;
    const top = viewportY + nodeY * zoom;
    const right = left + size.width * zoom;
    const bottom = top + size.height * zoom;
    return right > 0 && bottom > 0 && left < canvasWidth && top < canvasHeight;
  });

  return !hasVisibleNode;
}

