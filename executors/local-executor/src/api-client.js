const fs = require('node:fs');

class ApiError extends Error {
  constructor(status, message, body) {
    super(message || `HTTP ${status}`);
    this.name = 'ApiError';
    this.status = status;
    this.body = body;
  }
}

class ExecutorApiClient {
  constructor({ baseUrl, fetchImpl = globalThis.fetch }) {
    this.baseUrl = String(baseUrl || '').replace(/\/+$/, '');
    if (!this.baseUrl) throw new Error('baseUrl is required');
    if (typeof fetchImpl !== 'function') throw new Error('fetch implementation is required');
    this.fetch = fetchImpl;
  }

  async request(path, { token, body } = {}) {
    const headers = { Accept: 'application/json' };
    if (token) headers.Authorization = `Bearer ${token}`;
    let encoded;
    if (body !== undefined) {
      headers['Content-Type'] = 'application/json';
      encoded = JSON.stringify(body);
    }
    return this.perform(path, { headers, body: encoded });
  }

  async perform(path, { headers, body }) {
    const init = { method: 'POST', headers, body };
    if (isNodeStream(body)) init.duplex = 'half';
    const response = await this.fetch(`${this.baseUrl}${path}`, init);
    if (response.status === 204) return null;
    const text = await response.text();
    let parsed = null;
    if (text) {
      try { parsed = JSON.parse(text); } catch { parsed = { raw: text }; }
    }
    if (!response.ok) throw new ApiError(response.status, parsed?.error || `HTTP ${response.status}`, parsed);
    return parsed;
  }

  pair(input) { return this.request('/api/local-executor/v1/pair', { body: input }); }
  heartbeat(token, input) { return this.request('/api/local-executor/v1/heartbeat', { token, body: input }); }
  claim(token) { return this.request('/api/local-executor/v1/jobs/claim', { token }); }
  renew(token, jobId, lease) { return this.request(`/api/local-executor/v1/jobs/${encodeURIComponent(jobId)}/renew`, { token, body: leaseBody(lease) }); }
  progress(token, jobId, lease, state) { return this.request(`/api/local-executor/v1/jobs/${encodeURIComponent(jobId)}/progress`, { token, body: { ...leaseBody(lease), state } }); }
  acceptance(token, jobId, lease, input) { return this.request(`/api/local-executor/v1/jobs/${encodeURIComponent(jobId)}/acceptance`, { token, body: { ...leaseBody(lease), ...input } }); }
  release(token, jobId, lease, reason) { return this.request(`/api/local-executor/v1/jobs/${encodeURIComponent(jobId)}/release`, { token, body: { ...leaseBody(lease), reason } }); }
  fail(token, jobId, lease, input) { return this.request(`/api/local-executor/v1/jobs/${encodeURIComponent(jobId)}/fail`, { token, body: { ...leaseBody(lease), ...input } }); }
  result(token, jobId, lease, artifactId) { return this.request(`/api/local-executor/v1/jobs/${encodeURIComponent(jobId)}/result`, { token, body: { ...leaseBody(lease), artifactId } }); }

  uploadArtifact(token, jobId, lease, source) {
    const credential = leaseBody(lease);
    const headers = {
      Accept: 'application/json',
      Authorization: `Bearer ${token}`,
      'Content-Type': 'video/mp4',
      'X-Lease-Token': credential.leaseToken,
      'X-Lease-Generation': String(credential.leaseGeneration)
    };
    const body = typeof source === 'string' ? fs.createReadStream(source) : source;
    return this.perform(`/api/local-executor/v1/jobs/${encodeURIComponent(jobId)}/artifact`, { headers, body });
  }
}

function isNodeStream(value) {
  return Boolean(value && typeof value.pipe === 'function' && typeof value.on === 'function');
}

function leaseBody(lease = {}) {
  return {
    leaseToken: lease.leaseToken || lease.token || '',
    leaseGeneration: lease.leaseGeneration ?? lease.generation ?? 0
  };
}

module.exports = { ExecutorApiClient, ApiError, leaseBody, isNodeStream };
