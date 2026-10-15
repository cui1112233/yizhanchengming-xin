const crypto = require('node:crypto');
const http = require('node:http');
const https = require('node:https');

function signature(secret, { username, isOwner, issuedAt, method, pathname }) {
  const payload = [username, issuedAt, String(isOwner), method, pathname].join('\n');
  return crypto.createHmac('sha256', secret).update(payload).digest('hex');
}

function requestProductionBridge({ username, isOwner, pathname, body, method = 'POST', targetBaseUrl, bridgeSecret, timeoutMs = 30000 }) {
  const target = new URL(targetBaseUrl || process.env.QIANTIE_GO_BASE_URL || 'http://127.0.0.1:4000');
  const secret = bridgeSecret || process.env.QIANTIE_BRIDGE_SECRET || '';
  const transport = target.protocol === 'https:' ? https : http;
  const requestMethod = String(method || 'POST').toUpperCase();
  const issuedAt = String(Math.floor(Date.now() / 1000));
  const hasBody = body !== undefined && body !== null && requestMethod !== 'GET' && requestMethod !== 'HEAD';
  const rawBody = hasBody ? Buffer.from(JSON.stringify(body)) : null;
  const requestPath = pathname.startsWith('/') ? pathname : `/${pathname}`;
  const headers = {
    'X-Qiantie-Username': username,
    'X-Qiantie-Is-Owner': String(isOwner === true),
    'X-Qiantie-Issued-At': issuedAt,
    'X-Qiantie-Signature': signature(secret, { username, isOwner: isOwner === true, issuedAt, method: requestMethod, pathname: requestPath }),
    Accept: 'application/json'
  };
  if (rawBody) {
    headers['Content-Type'] = 'application/json';
    headers['Content-Length'] = String(rawBody.length);
  }

  return new Promise((resolve, reject) => {
    const request = transport.request({
      protocol: target.protocol,
      hostname: target.hostname,
      port: target.port || undefined,
      method: requestMethod,
      path: requestPath,
      headers,
      timeout: timeoutMs
    }, response => {
      const chunks = [];
      response.on('data', chunk => chunks.push(chunk));
      response.on('end', () => {
        const text = Buffer.concat(chunks).toString('utf8');
        let payload = {};
        try {
          payload = text ? JSON.parse(text) : {};
        } catch (_) {
          payload = { error: text || '生产服务返回了无效响应' };
        }
        resolve({ statusCode: response.statusCode || 502, payload });
      });
    });
    request.on('timeout', () => request.destroy(new Error('生产服务响应超时')));
    request.on('error', reject);
    if (rawBody) request.write(rawBody);
    request.end();
  });
}

module.exports = { requestProductionBridge };
