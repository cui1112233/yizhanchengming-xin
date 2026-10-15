const path = require('node:path');

class DeviceStore {
  constructor({ safeStorage, fs = require('node:fs'), filePath }) {
    if (!safeStorage || !filePath) throw new Error('safeStorage and filePath are required');
    this.safeStorage = safeStorage;
    this.fs = fs;
    this.filePath = filePath;
  }

  save({ baseUrl, executorId, token }) {
    if (!this.safeStorage.isEncryptionAvailable?.()) throw new Error('OS encryption is unavailable');
    if (!baseUrl || !executorId || !token) throw new Error('paired device data is incomplete');
    const encrypted = this.safeStorage.encryptString(String(token));
    const data = {
      version: 1,
      baseUrl: String(baseUrl).replace(/\/+$/, ''),
      executorId: String(executorId),
      tokenEncrypted: Buffer.from(encrypted).toString('base64')
    };
    this.fs.mkdirSync(path.dirname(this.filePath), { recursive: true });
    const tempPath = `${this.filePath}.tmp`;
    this.fs.writeFileSync(tempPath, JSON.stringify(data, null, 2), { encoding: 'utf8', mode: 0o600 });
    this.fs.renameSync(tempPath, this.filePath);
    return { baseUrl: data.baseUrl, executorId: data.executorId };
  }

  load() {
    let raw;
    try { raw = this.fs.readFileSync(this.filePath, 'utf8'); }
    catch (error) { if (error?.code === 'ENOENT') return null; throw error; }
    const data = JSON.parse(raw);
    if (data.version !== 1 || !data.tokenEncrypted) throw new Error('invalid device store');
    if (!this.safeStorage.isEncryptionAvailable?.()) throw new Error('OS encryption is unavailable');
    const token = this.safeStorage.decryptString(Buffer.from(data.tokenEncrypted, 'base64'));
    return { baseUrl: data.baseUrl, executorId: data.executorId, token };
  }
}

module.exports = { DeviceStore };
