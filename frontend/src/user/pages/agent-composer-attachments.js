const textExtensions = /\.(txt|md|csv|json)$/i;
const mediaTypes = /^(image|video|audio)\//i;

export function classifyComposerAttachment(file, { maxMediaBytes = 128 * 1024 * 1024, maxTextBytes = 200 * 1024 } = {}) {
  const name = String(file?.name || '');
  const type = String(file?.type || '');
  const size = Number(file?.size) || 0;
  if (textExtensions.test(name) || /^text\/(plain|markdown|csv)$/i.test(type) || /^application\/json$/i.test(type)) {
    return size > maxTextBytes ? { kind: 'too-large', limit: maxTextBytes } : { kind: 'text' };
  }
  if (mediaTypes.test(type)) return size > maxMediaBytes ? { kind: 'too-large', limit: maxMediaBytes } : { kind: 'media' };
  return { kind: 'unsupported' };
}

export function appendUniqueAssetID(assetIDs, assetID) {
  const id = String(assetID || '').trim();
  const current = Array.isArray(assetIDs) ? assetIDs.filter(Boolean) : [];
  return id && !current.includes(id) ? [...current, id] : current;
}
