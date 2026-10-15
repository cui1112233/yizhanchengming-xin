const fs = require('node:fs');
const path = require('node:path');

const VALID_CHANNELS = new Set(['beta', 'stable']);

class UpdatePreferencesStore {
  constructor({ filePath, fsImpl = fs } = {}) {
    if (!filePath) throw new Error('update preferences filePath is required');
    this.filePath = filePath;
    this.fs = fsImpl;
  }

  load() {
    try {
      const parsed = JSON.parse(this.fs.readFileSync(this.filePath, 'utf8'));
      return { channel: normalizeChannel(parsed?.channel) };
    } catch (error) {
      if (error?.code !== 'ENOENT' && !(error instanceof SyntaxError)) throw error;
      return { channel: 'beta' };
    }
  }

  save(value = {}) {
    const next = { channel: normalizeChannel(value.channel) };
    this.fs.mkdirSync(path.dirname(this.filePath), { recursive: true });
    const temp = `${this.filePath}.tmp`;
    this.fs.writeFileSync(temp, `${JSON.stringify(next, null, 2)}\n`, 'utf8');
    this.fs.renameSync(temp, this.filePath);
    return next;
  }
}

function normalizeChannel(value) {
  const channel = String(value || 'beta').trim().toLowerCase();
  if (!VALID_CHANNELS.has(channel)) throw new Error('update channel must be beta or stable');
  return channel;
}

module.exports = { UpdatePreferencesStore, normalizeChannel, VALID_CHANNELS };
