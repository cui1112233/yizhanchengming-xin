import { apiRequest } from './client.js';

export function prioritizeBatchBook(batchId, bookId) {
  if (!batchId || !bookId) return Promise.reject(new Error('请选择当前批次的一本书'));
  return apiRequest(`/api/batch-factory/v12/batches/${encodeURIComponent(batchId)}/automation/priority`, {
    method:'POST', body:JSON.stringify({bookId})
  });
}
