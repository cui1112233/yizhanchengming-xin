import { test, expect } from '@playwright/test';

const baseURL = process.env.E2E_BASE_URL || 'http://localhost:4001';

function fixture() {
  const raw = String(process.env.BATCH_FACTORY_AGENT_SHARED_FIXTURE || '').trim();
  if (!raw) {
    throw new Error('BATCH_FACTORY_AGENT_SHARED_FIXTURE is required; it must identify an authenticated disposable two-book fixture.');
  }
  let value;
  try {
    value = JSON.parse(raw);
  } catch (_) {
    throw new Error('BATCH_FACTORY_AGENT_SHARED_FIXTURE must be valid JSON.');
  }
  for (const path of ['batchId', 'book.id', 'book.title', 'otherBook.id', 'asset.id', 'asset.primaryImageId', 'asset.alternateImageId']) {
    const actual = path.split('.').reduce((current, key) => current?.[key], value);
    if (!String(actual || '').trim()) throw new Error(`BATCH_FACTORY_AGENT_SHARED_FIXTURE is missing ${path}.`);
  }
  return value;
}

async function api(page, path, options = {}) {
  const result = await page.evaluate(async ({ path, options }) => {
    const token = localStorage.getItem('auth_token') || '';
    const response = await fetch(path, {
      ...options,
      headers: { Authorization: `Bearer ${token}`, ...(options.headers || {}) }
    });
    const text = await response.text();
    let body = text;
    try { body = text ? JSON.parse(text) : null; } catch (_) {}
    return { status: response.status, body };
  }, { path, options });
  expect(result.status, `${options.method || 'GET'} ${path}`).toBeGreaterThanOrEqual(200);
  expect(result.status, `${options.method || 'GET'} ${path}`).toBeLessThan(300);
  return result.body;
}

function canvasOf(response) {
  return response?.canvas || response || { nodes: [], edges: [] };
}

function assetsOf(response) {
  return Array.isArray(response?.assets) ? response.assets : Array.isArray(response) ? response : [];
}

function imagesOf(response) {
  return Array.isArray(response?.images) ? response.images : Array.isArray(response) ? response : [];
}

function primaryImage(images) {
  return images.find(image => image?.isPrimary === true) || {};
}

function mappedAssetNode(canvas, fixtureValue) {
  return (canvas.nodes || []).find(node => (
    node?.data?.groupId === `book:${fixtureValue.book.id}`
    && node?.data?.config?.resourceType === 'book_asset'
    && node?.data?.config?.resourceId === fixtureValue.asset.id
  ));
}

test('batch factory and Agent use the same book media and merge state', async ({ page }) => {
  const value = fixture();
  const expectedTypes = value.expectedNodeTypes || ['script', 'text', 'image', 'video', 'audio'];
  const groupLabel = value.groupLabel || `书籍分组 ${value.book.id}`;
  const storageState = await page.evaluate(() => Boolean(localStorage.getItem('auth_token')));
  expect(storageState, 'the E2E storage state must contain auth_token').toBe(true);

  await page.goto(`${baseURL}/batch-factory?batch=${encodeURIComponent(value.batchId)}`, { waitUntil: 'networkidle' });
  const requestedBook = page.getByRole('row').filter({ hasText: value.book.title });
  await expect(requestedBook).toContainText(value.book.title);
  await expect(page.getByRole('row').filter({ hasText: value.otherBook.id })).toHaveCount(1);
  await requestedBook.getByRole('button', { name: 'Agent 优化' }).click();
  await expect(page).toHaveURL(new RegExp(`/agent/canvas\\?.*batchId=${encodeURIComponent(value.batchId)}.*bookId=${encodeURIComponent(value.book.id)}`));
  await expect(page.getByRole('button', { name: groupLabel })).toBeVisible();

  const taskId = new URL(page.url()).searchParams.get('taskId');
  expect(taskId, 'Agent navigation must include its bound taskId').toBeTruthy();
  const batchSnapshot = await api(page, `/api/batch-factory/v12/batches/${encodeURIComponent(value.batchId)}`);
  const books = batchSnapshot?.books || batchSnapshot?.batch?.books || [];
  expect(books.map(book => book?.id)).toEqual(expect.arrayContaining([value.book.id, value.otherBook.id]));
  const canvasBefore = canvasOf(await api(page, `/api/agent/tasks/${encodeURIComponent(taskId)}/canvas`));
  const requestedNodes = canvasBefore.nodes.filter(node => node?.data?.groupId === `book:${value.book.id}`);
  const otherNodes = canvasBefore.nodes.filter(node => node?.data?.groupId === `book:${value.otherBook.id}`);
  expect(requestedNodes.length).toBeGreaterThan(0);
  expect(otherNodes).toHaveLength(0);
  expect(requestedNodes.every(node => expectedTypes.includes(node?.data?.type))).toBe(true);
  expect(canvasBefore.edges.every(edge => requestedNodes.some(node => node.id === edge.source) && requestedNodes.some(node => node.id === edge.target))).toBe(true);

  const assetsPath = `/api/batch-factory/v12/batches/${encodeURIComponent(value.batchId)}/books/${encodeURIComponent(value.book.id)}/assets`;
  const assetPath = `${assetsPath}/${encodeURIComponent(value.asset.id)}`;
  const imagesPath = `${assetPath}/images`;
  expect(assetsOf(await api(page, assetsPath))).toHaveLength(value.expectedAssetCount || 5);
  const imagesBefore = imagesOf(await api(page, imagesPath));
  expect(primaryImage(imagesBefore).id).toBe(value.asset.primaryImageId);
  expect(imagesBefore.map(image => image.id)).toContain(value.asset.alternateImageId);

  try {
    await api(page, `${imagesPath}/${encodeURIComponent(value.asset.alternateImageId)}/primary`, { method: 'PUT' });
    await api(page, `/api/agent/tasks/${encodeURIComponent(taskId)}/batch-factory/books/${encodeURIComponent(value.book.id)}/sync`, { method: 'POST' });
    const canvasAfterVersionChange = canvasOf(await api(page, `/api/agent/tasks/${encodeURIComponent(taskId)}/canvas`));
    expect(mappedAssetNode(canvasAfterVersionChange, value)?.data?.assetRef?.id).toBe(value.asset.alternateImageId);
    expect(primaryImage(imagesOf(await api(page, imagesPath))).id).toBe(value.asset.alternateImageId);

    const removableNodeID = String(value.removableNodeId || mappedAssetNode(canvasAfterVersionChange, value)?.id || '');
    expect(removableNodeID, 'fixture must expose a mapped node that can be removed from the canvas only').toBeTruthy();
    await api(page, `/api/agent/tasks/${encodeURIComponent(taskId)}/batch-factory/nodes/${encodeURIComponent(removableNodeID)}`, { method: 'DELETE' });
    expect(assetsOf(await api(page, assetsPath)).some(asset => asset?.id === value.asset.id)).toBe(true);
    await api(page, `/api/agent/tasks/${encodeURIComponent(taskId)}/batch-factory/books/${encodeURIComponent(value.book.id)}/sync`, { method: 'POST' });

    const refreshedCanvas = canvasOf(await api(page, `/api/agent/tasks/${encodeURIComponent(taskId)}/canvas`));
    expect(mappedAssetNode(refreshedCanvas, value)?.data?.assetRef?.id).toBe(value.asset.alternateImageId);
    expect(JSON.stringify(refreshedCanvas)).toContain(String(value.expectedMergeState || 'succeeded'));
    expect(JSON.stringify(batchSnapshot)).toContain(String(value.expectedPublicationState || 'published'));
    expect(JSON.stringify(refreshedCanvas)).toContain(String(value.expectedPublicationState || 'published'));
    if (value.expectedDurationWarning) await expect(page.getByText(value.expectedDurationWarning)).toBeVisible();
  } finally {
    await api(page, `${imagesPath}/${encodeURIComponent(value.asset.primaryImageId)}/primary`, { method: 'PUT' });
    await api(page, `/api/agent/tasks/${encodeURIComponent(taskId)}/batch-factory/books/${encodeURIComponent(value.book.id)}/sync`, { method: 'POST' });
  }

  await page.goto(`${baseURL}/batch-factory?batch=${encodeURIComponent(value.batchId)}`, { waitUntil: 'networkidle' });
  await expect(page.getByRole('row').filter({ hasText: value.book.title })).toContainText(value.book.title);
  expect(primaryImage(imagesOf(await api(page, imagesPath))).id).toBe(value.asset.primaryImageId);
});
