// src-tauri/src/doubao/scripts/doubao-generation-page.js 的契约标记测试。
// 脚本是沈浪的混淆产物,不能在 happy-dom 里跑完整的 prepare(需要豆包真实 DOM);这里只锁定
// docs/analysis/10-live-probe-macos.md "Task 2.1" 对 Task 2.3 / 3.2 的几条结论所依赖的字符串表条目,
// 脚本被替换或升级时由这里先报警,再去核对文档。只断言"存在":同一文案可能同时出现在字符串表和内联处。
import { describe, expect, it } from 'vitest';
import generationSource from '../../src-tauri/src/doubao/scripts/doubao-generation-page.js?raw';

const has = (needle: string | RegExp) =>
  typeof needle === 'string' ? generationSource.includes(needle) : needle.test(generationSource);

describe('doubao-generation-page.js 契约标记(对应 10-live-probe-macos.md Task 2.1)', () => {
  it('视频 prepare 有 needsDurationInput 与 {mode:"video"} 两种成功返回', () => {
    // 摘要与请求不一致 → needsDurationInput;一致 → 直接 mode:'video'(Task 2.3 两种都要接受)
    expect(has('needsDurationInput')).toBe(true);
    expect(has("'video'")).toBe(true);
    // 不一致分支里选比例 / 找滑块的失败文案仍在,说明条件分支没被删
    expect(has('未找到所选视频比例，任务未发送')).toBe(true);
    expect(has('未找到唯一视频时长滑块，任务未发送')).toBe(true);
  });

  it('durationState 有"摘要已匹配直接返回秒 + fromSummary"分支与"读滑块 aria-valuenow"分支', () => {
    expect(has('fromSummary')).toBe(true);
    expect(has('aria-valuenow')).toBe(true);
    expect(has('视频时长设置未展开，任务未发送')).toBe(true);
  });

  it('图像入口里有退出技能启发式,但进技能步先按文本找 `技能` 按钮并以"豆包控件未唯一匹配"失败', () => {
    expect(has(/\(关闭\|移除\|取消\)/)).toBe(true);
    expect(has('豆包控件未唯一匹配，已停止发送。')).toBe(true);
    expect(has('技能菜单未显示')).toBe(true);
  });

  it('checkSend/send 阶段的异常文案存在(Task 2.3 映射 generation_submit_failed)', () => {
    expect(has('豆包发送按钮尚未就绪，未发送。')).toBe(true);
    expect(has('视频提示词已变化，任务未发送')).toBe(true);
    expect(has('豆包模式、模型、比例或提示词已变化，未发送。')).toBe(true);
  });

  it('prepare/fill 阶段的异常文案存在(Task 2.3 映射 generation_preparing_failed)', () => {
    expect(has('豆包输入框已变化，未覆盖已有内容。')).toBe(true);
    expect(has('未找到所选生成比例。')).toBe(true);
    expect(has('对话模式仅支持Mini的11～15秒任务')).toBe(true);
    expect(has('新对话尚未准备好，任务未发送')).toBe(true);
  });
});
