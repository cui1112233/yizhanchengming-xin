let database;
function openDatabase() {
  if (!database) database = new Promise((resolve, reject) => {
    const request = indexedDB.open('cm-novel-creation-drafts', 1);
    request.onupgradeneeded = () => request.result.createObjectStore('drafts');
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error || Error('创作草稿存储暂时不可用'));
    request.onblocked = () => reject(Error('草稿存储正在被旧页面占用，请关闭旧页面后重试'));
  }).catch(error => { database = null; throw error; });
  return database;
}
const key = owner => `long-imitation:${owner}`;
export async function readLongImitationDraft(owner) {
  const db = await openDatabase();
  return new Promise((resolve, reject) => {
    const request = db.transaction('drafts', 'readonly').objectStore('drafts').get(key(owner));
    request.onsuccess = () => resolve(request.result || null);
    request.onerror = () => reject(request.error);
  });
}
export async function writeLongImitationDraft(owner, value) {
  const db = await openDatabase();
  return new Promise((resolve, reject) => {
    const transaction = db.transaction('drafts', 'readwrite');
    const store = transaction.objectStore('drafts');
    if (value === null) store.delete(key(owner)); else store.put(value, key(owner));
    transaction.oncomplete = () => resolve();
    transaction.onerror = () => reject(transaction.error || Error('创作草稿保存失败'));
    transaction.onabort = () => reject(transaction.error || Error('创作草稿保存未完成'));
  });
}
