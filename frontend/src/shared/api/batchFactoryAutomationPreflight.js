import { apiRequest } from './client.js';
import { bf11Path } from './batchFactoryV12.js';

// 执行前只读预检：与 automation/start 同入参，但不建任务、不写任何配置。
// 单独成文件是为了让「立即执行」链路只依赖一个显式命名的只读接口。
export function preflightBatchAutomation(batchId, payload = {}) {
  const id = encodeURIComponent(String(batchId ?? ''));
  return apiRequest(bf11Path(`batches/${id}/automation/preflight`), {
    method: 'POST',
    body: JSON.stringify(payload ?? {})
  });
}
