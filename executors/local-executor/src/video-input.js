function normalizeVideoInput(payload = {}) {
  const prompt = String(payload.prompt || '').trim();
  if (!prompt) throw new Error('video prompt is required');
  const images = Array.isArray(payload.images) ? payload.images.filter(Boolean) : [];
  return { prompt, images };
}

module.exports = { normalizeVideoInput };
