import test from 'node:test';
import assert from 'node:assert/strict';
import { getWorkshopLogs, getWorkshopJob, listWorkshopTasks, listWorkshopSchedules, saveWorkshopSchedule, updateWorkshopSchedule, deleteWorkshopSchedule, workshopPlatformsPath, workshopTaskDiagnosticPath, previewWorkshopProcess, getWorkshopCurrentBatch, getWorkshopBatches, rerunWorkshopBatch, stopWorkshopTasks, saveWorkshopBody } from './novelFetchWorkshop.js';
import * as workshopApi from './novelFetchWorkshop.js';
import { getWebSubmitHistory, getWebSubmitOrganizations, saveWebSubmitConfig, uploadLogin, webSubmitOrganizationsPath } from './novelFetch.js';

test('Novel Fetch platform reader uses the dedicated platform endpoint', () => {
  assert.equal(workshopPlatformsPath(), '/api/novel-fetch-workshop/platforms');
});

test('Novel Fetch exposes Go-backed schedule and audit-log endpoints for the V78 controls', () => {
  assert.equal(typeof listWorkshopSchedules, 'function');
  assert.equal(typeof saveWorkshopSchedule, 'function');
  assert.equal(typeof updateWorkshopSchedule, 'function');
  assert.equal(typeof deleteWorkshopSchedule, 'function');
  assert.equal(typeof getWorkshopLogs, 'function');
});

test('Novel Fetch web-submit API keeps the Go settings envelope and history endpoint', () => {
	assert.equal(typeof getWebSubmitHistory, 'function');
	assert.equal(typeof saveWebSubmitConfig, 'function');
});

test('Novel Fetch reads organization ownership from the dedicated Go catalog endpoint', () => {
	assert.equal(typeof getWebSubmitOrganizations, 'function');
	assert.equal(webSubmitOrganizationsPath(), '/api/batch-factory/v12/publish-121/organizations');
});

test('Novel Fetch V78 workflow exposes preview, batch history, rerun and stop endpoints', () => {
  assert.equal(typeof previewWorkshopProcess, 'function');
  assert.equal(typeof getWorkshopCurrentBatch, 'function');
  assert.equal(typeof getWorkshopBatches, 'function');
  assert.equal(typeof rerunWorkshopBatch, 'function');
  assert.equal(typeof stopWorkshopTasks, 'function');
  assert.equal(typeof getWorkshopJob, 'function');
});

test('Novel Fetch realtime queue status uses the canonical Go endpoint', () => {
  assert.equal(workshopApi.workshopRealtimeStatusPath?.(), '/api/novel-fetch-workshop/realtime/status');
});

test('Novel Fetch task polling accepts silent error options', async () => {
  const previousFetch = globalThis.fetch;
  const previousWindow = globalThis.window;
  const previousStorage = globalThis.localStorage;
  globalThis.localStorage = { getItem() { return ''; } };
  globalThis.window = { dispatchEvent() {} };
  let requestInit;
  globalThis.fetch = async (_path, init) => {
    requestInit = init;
    return new Response(JSON.stringify({ tasks: [] }), { status: 200, headers: { 'content-type': 'application/json' } });
  };
  try {
    await listWorkshopTasks({ silent: true, suppressGlobalError: true });
    assert.equal(requestInit.silent, true);
    assert.equal(requestInit.suppressGlobalError, true);
  } finally {
    globalThis.fetch = previousFetch;
    globalThis.window = previousWindow;
    globalThis.localStorage = previousStorage;
  }
});

test('Novel Fetch task detail can save edited source through the Go body endpoint', () => {
  assert.equal(typeof saveWorkshopBody, 'function');
});

test('Novel Fetch task diagnostics use the Go compatibility endpoints', () => {
  assert.equal(workshopTaskDiagnosticPath('book 1', 'sensitive-log'), '/api/batch-rewrite/tasks/book%201/sensitive-log');
  assert.equal(workshopTaskDiagnosticPath('book/2', 'rules-trace'), '/api/batch-rewrite/tasks/book%2F2/rules-trace');
  assert.equal(workshopTaskDiagnosticPath('book-3', 'site-submit-log'), '/api/batch-rewrite/tasks/book-3/site-submit-log');
});

test('video-management login leaves error presentation to its modal', async () => {
  const previousFetch = globalThis.fetch;
  const previousWindow = globalThis.window;
  const previousStorage = globalThis.localStorage;
  const previousCustomEvent = globalThis.CustomEvent;
  const events = [];
  globalThis.localStorage = { getItem() { return ''; }, removeItem() {}, setItem() {} };
  globalThis.CustomEvent = class CustomEvent { constructor(type, options) { this.type = type; this.detail = options?.detail; } };
  globalThis.window = { location: { pathname: '/novel-fetch' }, dispatchEvent(event) { events.push(event); } };
  globalThis.fetch = async path => path === '/api/client-errors'
    ? new Response('{}', { status: 200, headers: { 'content-type': 'application/json' } })
    : new Response(JSON.stringify({ error: '视频管理系统登录入口不可用' }), { status: 502, headers: { 'content-type': 'application/json' } });
  try {
    await assert.rejects(uploadLogin({ username: 'operator', password: 'secret' }), /视频管理系统登录入口不可用/);
    assert.equal(events.filter(event => event.type === 'qiantie:api-error').length, 0);
  } finally {
    globalThis.fetch = previousFetch;
    globalThis.window = previousWindow;
    globalThis.localStorage = previousStorage;
    globalThis.CustomEvent = previousCustomEvent;
  }
});
