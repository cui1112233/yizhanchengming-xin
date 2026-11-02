const fs = require('node:fs/promises');
const path = require('node:path');

const maxReferenceBytes = 20 * 1024 * 1024;

async function downloadReferenceImages({ imageUrls, jobId, baseUrl, token, tempDir, fetchImpl = globalThis.fetch }) {
  const urls = Array.isArray(imageUrls) ? imageUrls.filter(Boolean) : [];
  if (urls.length === 0) return [];
  if (typeof fetchImpl !== 'function') throw new Error('reference image fetch is unavailable');
  if (!String(baseUrl || '').trim() || !String(token || '').trim() || !String(tempDir || '').trim()) {
    throw new Error('reference image download is not configured');
  }

  const destination = path.join(tempDir, safePathPart(jobId || 'job'));
  await fs.mkdir(destination, { recursive: true });
  const files = [];
  for (let index = 0; index < urls.length; index++) {
    const response = await fetchImpl(resolveReferenceURL(baseUrl, urls[index]), {
      headers: { Authorization: `Bearer ${token}` },
      signal: undefined
    });
    if (!response.ok) throw new Error(`reference image download failed (${response.status})`);
    const contentType = String(response.headers.get('content-type') || '').toLowerCase().split(';')[0].trim();
    const extension = imageExtension(contentType);
    if (!extension) throw new Error('reference asset is not an image');
    const bytes = Buffer.from(await response.arrayBuffer());
    if (bytes.length === 0 || bytes.length > maxReferenceBytes) throw new Error('reference image is empty or too large');
    const filePath = path.join(destination, `reference-${index + 1}.${extension}`);
    await fs.writeFile(filePath, bytes, { flag: 'w' });
    files.push(filePath);
  }
  return files;
}

function resolveReferenceURL(baseUrl, rawURL) {
  const value = String(rawURL || '').trim();
  if (!value) throw new Error('reference image URL is empty');
  if (/^https?:\/\//i.test(value)) return value;
  if (!value.startsWith('/')) throw new Error('reference image URL must be an absolute path');
  return String(baseUrl).replace(/\/+$/, '') + value;
}

function imageExtension(contentType) {
  return ({ 'image/png': 'png', 'image/jpeg': 'jpg', 'image/webp': 'webp', 'image/gif': 'gif' })[contentType] || '';
}

function safePathPart(value) {
  return String(value).replace(/[^A-Za-z0-9_-]/g, '_').slice(0, 96) || 'job';
}

module.exports = { downloadReferenceImages, resolveReferenceURL, imageExtension };
