const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');
const zlib = require('node:zlib');

const BOOK_ID_PATTERN = /^[A-Za-z0-9_.-]+$/;
const VERSION_ID_PATTERN = /^[A-Za-z0-9_.-]{1,64}$/;

class LocalNovelBodyStore {
  constructor({ rootDir, fsImpl = fs, now = () => new Date() } = {}) {
    const normalizedRoot = String(rootDir || '').trim();
    if (!normalizedRoot) throw new Error('rootDir is required');
    this.rootDir = path.resolve(normalizedRoot);
    this.fs = fsImpl;
    this.now = now;
  }

  put({ owner, bookId, versionId, content, sourceRevision = 0 } = {}) {
    const identity = normalizeIdentity({ owner, bookId, versionId });
    if (typeof content !== 'string') throw new Error('body content must be a string');
    const upstreamRevision = normalizeSourceRevision(sourceRevision);
    const paths = this.paths(identity);
    const encoded = Buffer.from(content, 'utf8');
    const compressed = zlib.gzipSync(encoded);
    const contentHash = sha256(encoded);
    const existing = this.readMeta(paths.metaPath, { allowMissing: true });

    if (existing && existing.contentHash === contentHash && existing.sourceRevision === upstreamRevision) {
      const current = this.get(identity);
      if (current) return current.ref;
    }

    const revision = Math.max(0, Number(existing?.revision) || 0) + 1;
    const meta = {
      version: 1,
      bookId: identity.bookId,
      versionId: identity.versionId,
      revision,
      sourceRevision: upstreamRevision,
      contentHash,
      charCount: [...content].length,
      storageBytes: compressed.length,
      updatedAt: this.now().toISOString()
    };

    this.fs.mkdirSync(paths.directory, { recursive: true, mode: 0o700 });
    atomicWrite(this.fs, paths.bodyPath, compressed);
    atomicWrite(this.fs, paths.metaPath, Buffer.from(JSON.stringify(meta, null, 2), 'utf8'));
    return { ...meta };
  }

  get({ owner, bookId, versionId } = {}) {
    const identity = normalizeIdentity({ owner, bookId, versionId });
    const paths = this.paths(identity);
    const bodyExists = exists(this.fs, paths.bodyPath);
    const metaExists = exists(this.fs, paths.metaPath);
    if (!bodyExists && !metaExists) return null;
    if (!bodyExists || !metaExists) throw new Error('local novel body data is incomplete');

    const meta = this.readMeta(paths.metaPath);
    const compressed = this.fs.readFileSync(paths.bodyPath);
    let decoded;
    try {
      decoded = zlib.gunzipSync(compressed);
    } catch (error) {
      throw new Error(`local novel body gzip is invalid: ${error.message}`);
    }
    const content = decoded.toString('utf8');
    const actualHash = sha256(decoded);
    if (actualHash !== meta.contentHash) throw new Error('local novel body hash mismatch');
    if ([...content].length !== meta.charCount) throw new Error('local novel body character count mismatch');
    return { ref: { ...meta }, content };
  }

  paths({ owner, bookId, versionId } = {}) {
    const identity = normalizeIdentity({ owner, bookId, versionId });
    const ownerKey = sha256(Buffer.from(identity.owner, 'utf8'));
    const directory = path.join(this.rootDir, 'novel-bodies', ownerKey, identity.bookId);
    const base = path.join(directory, identity.versionId);
    return {
      ownerKey,
      directory,
      bodyPath: `${base}.body.gz`,
      metaPath: `${base}.meta.json`
    };
  }

  readMeta(metaPath, { allowMissing = false } = {}) {
    let raw;
    try {
      raw = this.fs.readFileSync(metaPath, 'utf8');
    } catch (error) {
      if (allowMissing && error?.code === 'ENOENT') return null;
      throw error;
    }
    let meta;
    try {
      meta = JSON.parse(raw);
    } catch {
      throw new Error('local novel body metadata is invalid');
    }
    if (
      meta?.version !== 1 ||
      !BOOK_ID_PATTERN.test(String(meta.bookId || '')) ||
      !VERSION_ID_PATTERN.test(String(meta.versionId || '')) ||
      !Number.isInteger(meta.revision) || meta.revision < 1 ||
      !Number.isInteger(meta.sourceRevision) || meta.sourceRevision < 0 ||
      !/^[a-f0-9]{64}$/.test(String(meta.contentHash || '')) ||
      !Number.isInteger(meta.charCount) || meta.charCount < 0 ||
      !Number.isInteger(meta.storageBytes) || meta.storageBytes < 0
    ) {
      throw new Error('local novel body metadata is invalid');
    }
    return meta;
  }
}

function normalizeIdentity({ owner, bookId, versionId } = {}) {
  const normalizedOwner = String(owner || '').trim();
  const normalizedBookId = String(bookId || '').trim();
  const normalizedVersionId = String(versionId || '').trim();
  if (!normalizedOwner) throw new Error('owner is required');
  if (normalizedOwner.length > 191) throw new Error('owner is too long');
  if (!BOOK_ID_PATTERN.test(normalizedBookId)) throw new Error('invalid book id');
  if (!VERSION_ID_PATTERN.test(normalizedVersionId)) throw new Error('invalid version id');
  return { owner: normalizedOwner, bookId: normalizedBookId, versionId: normalizedVersionId };
}

function normalizeSourceRevision(value) {
  const revision = Number(value);
  if (!Number.isInteger(revision) || revision < 0 || revision > Number.MAX_SAFE_INTEGER) {
    throw new Error('source revision must be a non-negative integer');
  }
  return revision;
}

function sha256(buffer) {
  return crypto.createHash('sha256').update(buffer).digest('hex');
}

function exists(fsImpl, filePath) {
  try {
    fsImpl.accessSync(filePath);
    return true;
  } catch (error) {
    if (error?.code === 'ENOENT') return false;
    throw error;
  }
}

function atomicWrite(fsImpl, filePath, content) {
  const tempPath = `${filePath}.${process.pid}.${crypto.randomBytes(6).toString('hex')}.tmp`;
  try {
    fsImpl.writeFileSync(tempPath, content, { mode: 0o600 });
    fsImpl.renameSync(tempPath, filePath);
  } catch (error) {
    try { fsImpl.rmSync?.(tempPath, { force: true }); } catch {}
    throw error;
  }
}

module.exports = {
  LocalNovelBodyStore,
  normalizeIdentity,
  normalizeSourceRevision
};
