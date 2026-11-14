import { apiRequest } from './client.js';
import { comicWorkflowPath } from './comicScreenplayWorkflow.js';

// Used only after an explicit user action. Abort cancels the active request.
export function actComicAssetMedia(projectId, input, { signal } = {}, request = apiRequest) {
  return request(comicWorkflowPath(projectId, 'media/actions'), { method: 'POST', body: JSON.stringify(input), signal, timeoutMs: 0, suppressGlobalError: true });
}
