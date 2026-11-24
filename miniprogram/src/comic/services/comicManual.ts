import { apiRequest } from '../../services/api';
import { BATCH_FACTORY_BASE } from '../../services/batchFactoryV11';
import { loadComicPages, type ComicPageItem } from './comicResult';
let sequence = 0;
export function comicPromptMode(pages: Array<{
    comicPrompt?: string;
    imagePrompt?: string;
}>) {
    const ready = pages.filter(page => Boolean(page.comicPrompt?.trim() || page.imagePrompt?.trim())).length;
    return pages.length && ready === pages.length ? 'force' : ready ? 'missing' : 'bootstrap';
}
export function generateComicPrompts(batch: string, book: string, mode: 'bootstrap' | 'missing' | 'force') {
    if (!batch || !book)
        throw new Error('请先提取人物场景');
    return apiRequest(`${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batch)}/books/${encodeURIComponent(book)}/stages/comic`, { method: 'POST', timeout: 180000, data: { mode, requestId: `mini-comic-prompts-${Date.now()}-${++sequence}` } });
}
export async function saveComicManualSettings(batch: string, patch: Record<string, unknown>) {
    const value = await apiRequest<any>(`${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batch)}`, { timeout: 30000 });
    const revision = Number((value?.batch || value)?.revision);
    if (!Number.isFinite(revision))
        throw new Error('未读到项目版本，请刷新后重试');
    return apiRequest(`${BATCH_FACTORY_BASE}/batches/${encodeURIComponent(batch)}/settings`, { method: 'PUT', data: { expectedRevision: revision, patch } });
}
/** Read-only polling. A new candidate ID is required for regeneration, not an old primary image. */
export function watchComicPageImages(batch: string, book: string, baseline: Record<string, string[]>, callbacks: {
    onPages: (pages: ComicPageItem[]) => void;
    onDone: () => void;
    onError: (error: Error) => void;
}, options: {
    timeoutMs?: number;
} = {}) {
    let active = true;
    let timer: ReturnType<typeof setTimeout> | undefined;
    let reads = 0;
    const deadline = Date.now() + (options.timeoutMs ?? 300000);
    const stop = () => { active = false; if (timer !== undefined)
        clearTimeout(timer); };
    const tick = async () => {
        try {
            const pages = await loadComicPages(batch, book);
            if (!active)
                return;
            reads++;
            callbacks.onPages(pages);
            const targets = Object.keys(baseline);
            const hasNew = (page: ComicPageItem) => page.images.some(image => !baseline[page.videoId].includes(image.id));
            const matched = targets.map(id => pages.find(page => page.videoId === id));
            if (targets.length && matched.every(page => page && hasNew(page))) {
                stop();
                callbacks.onDone();
                return;
            }
            if (reads > 1 && matched.some(page => page && !hasNew(page) && /fail|error/i.test(page.status))) {
                stop();
                callbacks.onError(new Error('有漫画页生成失败，请检查失败页后手动重试'));
                return;
            }
            if (Date.now() >= deadline) {
                stop();
                callbacks.onError(new Error('图片仍未全部返回，可稍后刷新；没有自动重复提交生成'));
                return;
            }
            timer = setTimeout(tick, 3000);
        }
        catch (cause: any) {
            if (active) {
                stop();
                callbacks.onError(new Error(cause?.message || '页图状态读取失败，请刷新'));
            }
        }
    };
    void tick();
    return stop;
}
