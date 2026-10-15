const fs = require('node:fs');
const fsp = require('node:fs/promises');
const path = require('node:path');
const { randomUUID } = require('node:crypto');
const { Readable, Transform } = require('node:stream');
const { pipeline } = require('node:stream/promises');

class DoubaoDownloadError extends Error {
  constructor(code, message) {
    super(message);
    this.name = 'DoubaoDownloadError';
    this.code = code;
  }
}

async function downloadMp4WithSession({ webContents, url, downloadDir, jobId, mediaId, maxBytes = 1024 * 1024 * 1024, minBytes = 1024 }) {
  const targetUrl = String(url || '').trim();
  if (!/^https?:\/\//i.test(targetUrl)) {
    throw new DoubaoDownloadError('DOWNLOAD_URL_INVALID', 'Doubao download URL must be http(s)');
  }
  const fetchImpl = webContents?.session?.fetch;
  if (typeof fetchImpl !== 'function') {
    throw new DoubaoDownloadError('SESSION_FETCH_UNAVAILABLE', 'Doubao account session fetch is unavailable');
  }
  const dir = path.resolve(String(downloadDir || '').trim() || path.join(process.cwd(), 'downloads'));
  await fsp.mkdir(dir, { recursive: true });

  const response = await fetchImpl.call(webContents.session, targetUrl, { method: 'GET', redirect: 'follow' });
  if (!response?.ok || !response.body) {
    throw new DoubaoDownloadError('DOWNLOAD_FAILED', `Doubao MP4 download failed with HTTP ${response?.status || 0}`);
  }

  const safeJob = safeName(jobId || 'job');
  const safeMedia = safeName(mediaId || 'media');
  const finalPath = path.join(dir, `${safeJob}-${safeMedia}-${randomUUID()}.mp4`);
  const tempPath = `${finalPath}.part`;
  let byteSize = 0;
  const limiter = new Transform({
    transform(chunk, _encoding, callback) {
      byteSize += chunk.length;
      if (byteSize > maxBytes) {
        callback(new DoubaoDownloadError('MP4_TOO_LARGE', `downloaded MP4 exceeds ${maxBytes} bytes`));
        return;
      }
      callback(null, chunk);
    }
  });

  try {
    const body = typeof Readable.fromWeb === 'function' && typeof response.body?.getReader === 'function'
      ? Readable.fromWeb(response.body)
      : Readable.from(response.body);
    await pipeline(body, limiter, fs.createWriteStream(tempPath, { flags: 'wx' }));
    await validateMp4File(tempPath, { byteSize, minBytes });
    await fsp.rename(tempPath, finalPath);
    return { filePath: finalPath, byteSize };
  } catch (error) {
    await fsp.rm(tempPath, { force: true }).catch(() => {});
    await fsp.rm(finalPath, { force: true }).catch(() => {});
    throw error;
  }
}

async function validateMp4File(filePath, { byteSize, minBytes = 1024 } = {}) {
  const stat = await fsp.stat(filePath);
  const size = Number(byteSize || stat.size);
  if (size < minBytes) {
    throw new DoubaoDownloadError('MP4_TOO_SMALL', `downloaded video is too small: ${size} bytes`);
  }
  const handle = await fsp.open(filePath, 'r');
  try {
    const header = Buffer.alloc(12);
    const { bytesRead } = await handle.read(header, 0, header.length, 0);
    if (bytesRead < 12 || header.toString('ascii', 4, 8) !== 'ftyp') {
      throw new DoubaoDownloadError('INVALID_MP4_HEADER', 'downloaded content is not a valid MP4');
    }
    const boxSize = header.readUInt32BE(0);
    if (boxSize < 8 || boxSize > size) {
      throw new DoubaoDownloadError('INVALID_MP4_HEADER', 'downloaded MP4 has an invalid ftyp box');
    }
  } finally {
    await handle.close();
  }
  return true;
}

function safeName(value) {
  return String(value || '').replace(/[^a-zA-Z0-9_-]/g, '_').slice(0, 64) || 'item';
}

module.exports = { DoubaoDownloadError, downloadMp4WithSession, validateMp4File, safeName };
