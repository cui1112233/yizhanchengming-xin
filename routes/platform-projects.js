const http = require('node:http');
const https = require('node:https');
const express = require('express');

const { apiAuth } = require('../middleware/auth');
const { USERS_DIR, readHistoryIndex } = require('../lib/shared');
const { createNovelPanelStore } = require('../lib/novel-panel/project-store');
const { signBridgeRequest, resolveShuihuoBaseUrl } = require('./shuihuo-production');

const DEFAULT_LIMIT = 100;

function toTime(value) {
  const time = Date.parse(value || '');
  return Number.isFinite(time) && new Date(time).getUTCFullYear() >= 2000 ? time : 0;
}

function usableTimestamp(value) {
  return toTime(value) ? value : '';
}

function normalizeLimit(value) {
  const parsed = Number.parseInt(value, 10);
  if (!Number.isFinite(parsed)) return DEFAULT_LIMIT;
  return Math.max(1, Math.min(DEFAULT_LIMIT, parsed));
}

function listScriptHistory(username) {
  const entries = readHistoryIndex(username).entries || [];
  return entries.map(entry => ({
    id: `script-history:${entry.id}`,
    sourceId: entry.id,
    kind: 'script-history',
    typeLabel: '剧本生成',
    name: entry.formatName || entry.format || '剧本生成记录',
    summary: entry.preview || '',
    updatedAt: entry.createdAt || '',
    route: `/history?entry=${encodeURIComponent(`script-history:${entry.id}`)}`
  }));
}

function listNovelPanelProjects(store, username) {
  store.migrateLegacyIfNeeded(username);
  return store.listProjects(username).map(project => ({
    id: `novel-panel:${project.id}`,
    sourceId: project.id,
    kind: 'novel-panel',
    typeLabel: '小说面板',
    name: project.name || '未命名小说面板项目',
    summary: '小说分析、人物场景与分镜项目',
    createdAt: project.created_at || '',
    updatedAt: project.updated_at || project.created_at || '',
    route: `/novel-panel?project=${encodeURIComponent(project.id)}`
  }));
}

function listShuihuoProjects({ targetBaseUrl, bridgeSecret }, account) {
  const target = new URL(resolveShuihuoBaseUrl(targetBaseUrl));
  const transport = target.protocol === 'https:' ? https : http;
  const pathname = '/api/shuihuo-production/projects';
  const issuedAt = String(Math.floor(Date.now() / 1000));
  const signature = signBridgeRequest(bridgeSecret || process.env.QIANTIE_BRIDGE_SECRET || '', {
    username: account.username,
    isOwner: account.isOwner === true,
    issuedAt,
    method: 'GET',
    pathname
  });

  return new Promise((resolve, reject) => {
    const request = transport.request({
      protocol: target.protocol,
      hostname: target.hostname,
      port: target.port || undefined,
      method: 'GET',
      path: pathname,
      timeout: 5000,
      headers: {
        Accept: 'application/json',
        'X-Qiantie-Username': account.username,
        'X-Qiantie-Is-Owner': String(account.isOwner === true),
        'X-Qiantie-Issued-At': issuedAt,
        'X-Qiantie-Signature': signature
      }
    }, response => {
      const chunks = [];
      response.on('data', chunk => chunks.push(chunk));
      response.on('end', () => {
        if ((response.statusCode || 500) >= 400) return reject(new Error(`水货生产服务返回 ${response.statusCode}`));
        try {
          const body = JSON.parse(Buffer.concat(chunks).toString('utf8'));
          const projects = Array.isArray(body.projects) ? body.projects : [];
          resolve(projects.map(project => ({
            id: `shuihuo-production:${project.id}`,
            sourceId: String(project.id),
            kind: 'shuihuo-production',
            typeLabel: '水货生产',
            name: project.name || '未命名水货生产作品',
            summary: project.segmentationStatus === 'confirmed' ? '已确认分段，可继续生成素材' : '已保存原文，等待分段确认',
            createdAt: usableTimestamp(project.createdAt),
            updatedAt: usableTimestamp(project.updatedAt) || usableTimestamp(project.createdAt),
            route: `/shuihuo-production?project=${encodeURIComponent(project.id)}`
          })));
        } catch (error) {
          reject(error);
        }
      });
    });
    request.on('timeout', () => request.destroy(new Error('水货生产服务响应超时')));
    request.on('error', reject);
    request.end();
  });
}

function createPlatformProjectsRouter({ shuihuoGateway = {} } = {}) {
  const router = express.Router();
  const novelPanelStore = createNovelPanelStore({ usersDir: USERS_DIR });
  router.use(apiAuth);

  router.get('/', async (req, res) => {
    const limit = normalizeLimit(req.query.limit);
    const entries = [
      ...listScriptHistory(req.username),
      ...listNovelPanelProjects(novelPanelStore, req.username)
    ];
    const unavailableSources = [];
    try {
      entries.push(...await listShuihuoProjects(shuihuoGateway, req.auth.account));
    } catch (_) {
      unavailableSources.push('shuihuo-production');
    }
    entries.sort((left, right) => toTime(right.updatedAt) - toTime(left.updatedAt) || left.id.localeCompare(right.id));
    res.json({ entries: entries.slice(0, limit), unavailableSources });
  });

  return router;
}

module.exports = { createPlatformProjectsRouter, listShuihuoProjects };
