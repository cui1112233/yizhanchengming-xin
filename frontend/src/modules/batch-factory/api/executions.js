import { apiRequest } from '../../../shared/api/client.js';

const BASE = '/api/batch-factory/v12';
const id = value => encodeURIComponent(String(value ?? ''));

export function ensureBatchExecution(batchId, payload, request = apiRequest) {
  return request(`${BASE}/batches/${id(batchId)}/executions/ensure`, { method: 'POST', body: JSON.stringify(payload ?? {}) });
}

export function getBatchExecution(executionId, request = apiRequest) {
  return request(`${BASE}/executions/${id(executionId)}`);
}

export function getBatchExecutionTimeline(executionId, bookId, request = apiRequest) {
  return request(`${BASE}/executions/${id(executionId)}/books/${id(bookId)}/timeline`);
}

export function reconcileBatchExecution(executionId, request = apiRequest) {
  return request(`${BASE}/executions/${id(executionId)}/reconcile`, { method: 'POST', body: '{}' });
}

export function cancelBatchExecution(executionId, request = apiRequest) {
  return request(`${BASE}/executions/${id(executionId)}/cancel`, { method: 'POST', body: '{}' });
}
