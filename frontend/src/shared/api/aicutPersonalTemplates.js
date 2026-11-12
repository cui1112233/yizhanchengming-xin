import { apiRequest } from './client.js';
const path = '/api/ai-cut/personal-templates';
export async function listPersonalTemplates(request = apiRequest) { return (await request(path)).templates || []; }
export async function savePersonalTemplate(title, layout, request = apiRequest) { return (await request(path, { method: 'POST', body: JSON.stringify({ title, layout }) })).template; }
export async function deletePersonalTemplate(id, request = apiRequest) {
  if (typeof id !== 'string' || !/^[A-Za-z0-9_-]{1,64}$/.test(id)) throw new Error('模板编号无效');
  return request(`${path}/${id}`, { method: 'DELETE' });
}
