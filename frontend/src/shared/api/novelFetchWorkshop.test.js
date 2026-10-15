import test from 'node:test';
import assert from 'node:assert/strict';
import { getWorkshopLogs, listWorkshopSchedules, saveWorkshopSchedule, deleteWorkshopSchedule, workshopPlatformsPath } from './novelFetchWorkshop.js';
import { getWebSubmitHistory, saveWebSubmitConfig } from './novelFetch.js';

test('Novel Fetch platform reader uses the dedicated platform endpoint', () => {
  assert.equal(workshopPlatformsPath(), '/api/novel-fetch-workshop/platforms');
});

test('Novel Fetch exposes Go-backed schedule and audit-log endpoints for the V78 controls', () => {
  assert.equal(typeof listWorkshopSchedules, 'function');
  assert.equal(typeof saveWorkshopSchedule, 'function');
  assert.equal(typeof deleteWorkshopSchedule, 'function');
  assert.equal(typeof getWorkshopLogs, 'function');
});

test('Novel Fetch web-submit API keeps the Go settings envelope and history endpoint', () => {
  assert.equal(typeof getWebSubmitHistory, 'function');
  assert.equal(typeof saveWebSubmitConfig, 'function');
});
