class Mp4ValidationError extends Error {
  constructor(code, message) {
    super(message);
    this.name = 'Mp4ValidationError';
    this.code = code;
  }
}

function validateMp4Buffer(buffer, { minBytes = 1024 } = {}) {
  if (!Buffer.isBuffer(buffer)) {
    throw new Mp4ValidationError('MP4_BUFFER_REQUIRED', 'downloaded video must be a Buffer');
  }
  if (buffer.length < minBytes) {
    throw new Mp4ValidationError('MP4_TOO_SMALL', `downloaded video is too small: ${buffer.length} bytes`);
  }
  if (buffer.length < 12 || buffer.toString('ascii', 4, 8) !== 'ftyp') {
    throw new Mp4ValidationError('INVALID_MP4_HEADER', 'downloaded content does not have an MP4 ftyp header');
  }
  const boxSize = buffer.readUInt32BE(0);
  if (boxSize < 8 || boxSize > buffer.length) {
    throw new Mp4ValidationError('INVALID_MP4_HEADER', 'downloaded content has an invalid MP4 ftyp box size');
  }
  const brand = buffer.toString('ascii', 8, 12).replace(/[^\x20-\x7e]/g, '');
  return { ok: true, size: buffer.length, brand };
}

module.exports = { Mp4ValidationError, validateMp4Buffer };
