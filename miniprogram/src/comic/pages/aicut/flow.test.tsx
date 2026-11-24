import { beforeEach, expect, test, vi } from 'vitest';
const h = vi.hoisted(() => ({ slots: [] as any[], cursor: 0, effects: [] as any[], seen: new Set<number>(), api: vi.fn(), modal: vi.fn(), hide: () => {}, show: () => {} }));
vi.mock('react', () => ({
  useState: (v: any) => { const i = h.cursor++; if (!(i in h.slots)) h.slots[i] = typeof v === 'function' ? v() : v; return [h.slots[i], (next: any) => { h.slots[i] = typeof next === 'function' ? next(h.slots[i]) : next; }]; },
  useRef: (v: any) => { const i = h.cursor++; if (!(i in h.slots)) h.slots[i] = { current: v }; return h.slots[i]; },
  useEffect: (fn: any) => { const i = h.cursor++; if (!h.seen.has(i)) { h.seen.add(i); h.effects.push(fn); } },
}));
vi.mock('@tarojs/taro', () => ({ default: { showModal: h.modal, showToast: vi.fn(), enableAlertBeforeUnload: vi.fn(), disableAlertBeforeUnload: vi.fn() }, useLoad:()=>{},useDidHide: (fn: any) => { h.hide = fn; }, useDidShow: (fn: any) => { h.show = fn; } }));
vi.mock('@tarojs/components', () => Object.fromEntries(['Audio', 'Button', 'Image', 'Input', 'Picker', 'ScrollView', 'Text', 'Textarea', 'Video', 'View', 'Progress'].map(name => [name, name])));
vi.mock('../../../services/api', () => ({ apiRequest: h.api }));
vi.mock('../../../services/aicutMedia', () => ({ importCutMedia: vi.fn(), downloadCutExport: vi.fn() }));
vi.mock('../../../services/mediaCache', () => ({ loadAuthenticatedMedia: vi.fn() }));
vi.mock('../../../theme', () => ({ ThemedPage: 'Theme' }));
vi.mock('../../../components/AppIcon', () => ({ default: 'Icon' }));
vi.mock('../../../components/BrandHeader', () => ({ default: 'Header' }));
import CutEditor from './CutEditor';
import CutExports from './CutExports';
import AICutPage from './index';

let page: () => any;
const render = () => { h.cursor = 0; const tree = page(); while (h.effects.length) h.effects.shift()(); return tree; };
const nodes = (tree: any): any[] => tree && typeof tree === 'object' ? [tree, ...[tree.props?.children].flat(Infinity).flatMap(nodes)] : [];
const text = (tree: any): string => typeof tree === 'string' || typeof tree === 'number' ? String(tree) : tree?.props ? [tree.props.children].flat(Infinity).map(text).join('') : '';
const click = (label: string) => { const button = nodes(render()).find(node => node.type === 'Button' && text(node) === label); expect(button, label).toBeTruthy(); expect(button.props.disabled).not.toBe(true); return button.props.onClick(); };
const flush = async () => { for (let i = 0; i < 35; i++) await Promise.resolve(); };
const project = { id: 'p1', title: '已有工程', aspectRatio: '9:16' };
const value = { revision: 7, script: '保留剧本', markers: [{ id: 'm', timeMs: 100 }], tracks: [] };
beforeEach(() => {
  vi.useRealTimers(); h.slots = []; h.cursor = 0; h.effects = []; h.seen.clear(); h.api.mockReset(); h.modal.mockResolvedValue({ confirm: true });
  h.api.mockImplementation((path: string, options?: any) => Promise.resolve(path.endsWith('/snapshot') ? { timeline: { ...(options?.data?.timeline || value), revision: options?.method === 'PUT' ? 8 : 7 } } : path.endsWith('/timelines') ? { timelines: [{ id: 'default', title: '主时间线' }, { id: 'alt', title: '剪辑B' }] } : path.endsWith('/assets') ? { assets: [] } : path.endsWith('/export-jobs') ? { jobs: [] } : path.endsWith('/projects') ? { projects: [project] } : {}));
  page = () => CutEditor({ project, onBack: vi.fn() });
});
test('opens real AI-CUT endpoints and never invokes CM canvas', async () => {
  render(); await flush(); expect(text(render())).toContain('时间线还是空的');
  expect(h.api.mock.calls.some(([path]) => path === '/api/ai-cut/projects/p1/timelines/default/snapshot')).toBe(true);
  expect(h.api.mock.calls.every(([path]) => path.startsWith('/api/ai-cut/'))).toBe(true);
});
test('subtitle edits stay local until one explicit versioned save, including double taps', async () => {
  render(); await flush(); await click('字幕');
  nodes(render()).find(node => node.type === 'Textarea').props.onInput({ detail: { value: '第一幕' } });
  await click('添加到时间线'); expect(h.api.mock.calls.some(([, options]) => options?.method === 'PUT')).toBe(false);
  const save = nodes(render()).find(node => node.props?.id === 'cut-save').props.onClick;
  save(); save(); await flush();
  const writes = h.api.mock.calls.filter(([, options]) => options?.method === 'PUT'); expect(writes).toHaveLength(1);
  expect(writes[0][1].data).toMatchObject({ expectedRevision: 7, timeline: { script: '保留剧本', markers: [{ id: 'm' }], tracks: [{ clips: [{ text: '第一幕' }] }] } });
  expect(text(render())).toContain('已与工程同步');
});
test('409 leaves local edits intact with no automatic write retry', async () => {
  render(); await flush(); await click('字幕'); nodes(render()).find(node => node.type === 'Textarea').props.onInput({ detail: { value: '保留修改' } }); await click('添加到时间线');
  const normal = h.api.getMockImplementation()!; h.api.mockImplementation((path, options) => options?.method === 'PUT' ? Promise.reject(Object.assign(new Error('conflict'), { statusCode: 409 })) : normal(path, options));
  await click('保存工程'); await flush(); expect(text(render())).toContain('本地修改仍保留'); expect(text(render())).toContain('有未保存的修改');
  expect(h.api.mock.calls.filter(([, options]) => options?.method === 'PUT')).toHaveLength(1);
});
test('named timeline switches to its own snapshot and scope', async () => {
  render(); await flush(); const picker = nodes(render()).find(node => node.type === 'Picker'); await picker.props.onChange({ detail: { value: 1 } }); await flush();
  expect(h.api).toHaveBeenCalledWith('/api/ai-cut/projects/p1/timelines/alt/snapshot', { header: { 'X-AICut-Timeline-Id': 'alt' } });
  expect(text(render())).toContain('剪辑B');
});
test('an empty timeline can still open its existing export history', async () => {
  render(); await flush(); await click('导出');
  const panel = nodes(render()).find(node => node.type === CutExports); expect(panel).toBeTruthy(); expect(panel.props.canExport).toBe(false);
});
test('opening exports is read-only and shows real progress', async () => {
  page = () => CutExports({ projectId: 'p1', timelineId: 'alt', revision: 7 });
  h.api.mockResolvedValue({ jobs: [{ id: 'j1', state: 'failed', progress: 42, options: { timelineId: 'alt' }, error: '素材丢失' }] });
  render(); await flush(); expect(text(render())).toContain('42%'); expect(text(render())).toContain('素材丢失'); expect(h.api.mock.calls.some(([, options]) => options?.method === 'POST')).toBe(false);
});
test('export read polling stops when hidden and resumes reads without resubmitting', async () => {
  vi.useFakeTimers(); page = () => CutExports({ projectId: 'p1', timelineId: 'default', revision: 7 });
  h.api.mockResolvedValue({ jobs: [{ id: 'j1', state: 'running', progress: 3 }] });
  render(); await flush(); h.hide(); const count = h.api.mock.calls.length; await vi.advanceTimersByTimeAsync(9000); expect(h.api).toHaveBeenCalledTimes(count);
  h.show(); await flush(); expect(h.api.mock.calls.every(([, options]) => !options?.method)).toBe(true); h.hide(); vi.useRealTimers();
});
test('project library failure is displayed instead of claiming there are no projects', async () => {
  page = AICutPage; h.api.mockRejectedValue(new Error('网络不可用')); render(); await flush(); expect(text(render())).toContain('网络不可用'); expect(text(render())).not.toContain('你的下一部作品');
});
