//! 生成管线(Task 2.3):退出残留技能 → 注入 → `prepare` → `fill` → `checkSend` → `send`
//! → 2s 内从 `Network.requestWillBeSent` 识别 `submissionIdentity`。
//!
//! 真机结论(`docs/analysis/10-live-probe-macos.md` Task 2.1 / 2.3):
//! - `submissionMode`:图像传 `conversation`(image 分支不看它),视频传 `skill`
//!   (任何非 `conversation`/`work` 的值进入"视频生成"技能,四个模型全部可用);prompt 不加前缀。
//! - 技能态下"新对话"不退出技能,跨技能 `prepare` 会在 12s 后失败,所以 `prepare` 前先点
//!   `[data-testid="skill_input_exit_button"]`(可见时)并等 `[data-testid="guidance-skill-bar"]`
//!   出现(真机 157ms)。
//! - 视频 `prepare` 有两种成功形状:`{needsDurationInput:true}`(弹层留开、滑块 focus,接
//!   `duration::set_duration`,Task 3.2)与 `{mode:"video",model,ratio,duration}`(摘要已匹配,直接 `fill`)。
//! - 生成参数通过 chat target 的 CDP 会话写入；后台提交不得抢占用户正在使用的窗口。
//! - 异常映射:`prepare`/`fill` 抛出 → `generation_preparing_failed`;`checkSend`/`send` 抛出 →
//!   `generation_submit_failed`;都带 `details.js_exception` 原文。
//!
//! `doubao-result-reader.js` 的 `submissionIdentity(request, prompt, options)`(已在 Node 里对混淆脚本实测):
//! `request = {url, method, postData}`;URL 须为 https、host 以 `doubao.com`/`snssdk.com` 结尾、
//! POST、path 含 `completion|generate|/im/message/send` 且不含 `history|recommend|config`;
//! `postData` 里的 `text/content/prompt` 字段须包含 `prompt`;返回
//! `{localMessageId, messageId, conversationId}`(三者皆可为 null,`localMessageId` 为 null 时
//! `messageId` 必有)或 `null`。`options.verifiedSend` 只对 `/chat/completion` 路径生效。
//! `postData` 先经 `json_fix::fix_big_integers` 处理,避免页面侧 `JSON.parse` 丢失大整数精度。

use std::path::PathBuf;
use std::time::{Duration, Instant};

use serde::Deserialize;
use serde_json::{json, Value};
use tokio::sync::broadcast;
use tokio_util::sync::CancellationToken;

use crate::cdp::json_fix::fix_big_integers;
use crate::cdp::session::{extract_js_exception, CdpEvent};
use crate::cdp::{now_ms, CdpError, CdpSession};
use crate::doubao::{duration, upload};
use crate::error::{RelayError, RelayResult};
use crate::tasks::{prompt_hash, GenerateRequest, TaskStage, TaskType};

/// 生成页脚本(`shenlangDoubaoGeneration`),注入时按 trackingKey 重命名全局对象。
pub const GENERATION_SCRIPT: &str = include_str!("scripts/doubao-generation-page.js");
/// 结果读取脚本(`shenlangDoubaoResultReader`),`submissionIdentity` / `read` 等都在这里。
pub const RESULT_READER_SCRIPT: &str = include_str!("scripts/doubao-result-reader.js");
/// 生成页脚本里的默认全局对象名(被 trackingKey 替换)。
pub const GENERATION_GLOBAL: &str = "shenlangDoubaoGeneration";
/// 结果读取脚本的全局对象名。
pub const READER_GLOBAL: &str = "shenlangDoubaoResultReader";

/// 豆包新版工具栏会把图片比例直接显示成 `16:9`，并把视频的比例、时长合并为
/// `自动 · 10s`。旧脚本分别寻找 `比例 16:9` 与独立视频参数控件。
/// 不改第三方混淆脚本，而是在每个任务的隔离对象上补齐它读取的标签。
fn generation_ratio_compatibility_script(tracking_key: &str) -> String {
    format!(
        r#"
;(function normalizeRatioControlForCurrentDoubao() {{
  const api = globalThis[{tracking_key:?}];
  if (!api || api.__relayRatioCompatibility || typeof api.run !== 'function') return;
  const originalRun = api.run.bind(api);
  const textOf = element => String(element?.getAttribute?.('aria-label') || element?.innerText || element?.textContent || '').replace(/\s+/g, ' ').trim();
  const visible = element => {{
    const rect = element?.getBoundingClientRect?.();
    const style = element && getComputedStyle(element);
    return Boolean(rect && rect.width > 0 && rect.height > 0 && style?.display !== 'none' && style?.visibility !== 'hidden');
  }};
  const inputRoot = () => document.querySelector('#input-engine-container,[data-testid="input-engine-container"],[data-automation-id="input-engine-container"],[data-testid="chat_input"]');
  const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));
  const press = element => {{
    const rect = element.getBoundingClientRect();
    const init = {{ bubbles: true, cancelable: true, view: window, clientX: rect.left + rect.width / 2, clientY: rect.top + rect.height / 2, button: 0 }};
    if (typeof PointerEvent !== 'undefined') element.dispatchEvent(new PointerEvent('pointerdown', {{ ...init, buttons: 1, pointerId: 1, pointerType: 'mouse', isPrimary: true }}));
    element.dispatchEvent(new MouseEvent('mousedown', {{ ...init, buttons: 1 }}));
    if (typeof PointerEvent !== 'undefined') element.dispatchEvent(new PointerEvent('pointerup', {{ ...init, buttons: 0, pointerId: 1, pointerType: 'mouse', isPrimary: true }}));
    element.dispatchEvent(new MouseEvent('mouseup', {{ ...init, buttons: 0 }}));
    element.click();
  }};
  const videoModeReady = () => {{
    const root = inputRoot();
    return Boolean(root) && [...root.querySelectorAll('button,[role="button"],div,span')]
    .filter(visible)
    .some(control => /^(?:视频生成)?模型\s*Seedance/i.test(textOf(control)));
  }};
  const ensureCurrentVideoEntry = async request => {{
    if (request?.type !== 'video' || videoModeReady()) return;
    const root = inputRoot();
    if (!root) return;
    const buttons = () => [...root.querySelectorAll('button,[role="button"]')].filter(visible);
    const launcher = buttons().filter(button => /^(?:插件\s*[·•.]\s*)?技能(?:\s*[·•.]\s*伙伴)?$/.test(textOf(button)));
    if (launcher.length !== 1) return;
    press(launcher[0]);
    const deadline = Date.now() + 3000;
    while (Date.now() < deadline) {{
      const entries = [...root.querySelectorAll('[role="menuitem"],[role="option"],button,[role="button"],div,span')]
        .filter(visible)
        .filter(entry => textOf(entry) === '视频生成');
      const leaves = entries.filter(entry => !entries.some(other => other !== entry && entry.contains(other)));
      if (leaves.length === 1) {{
        press(leaves[0]);
        const stableUntil = Date.now() + 3000;
        while (Date.now() < stableUntil) {{
          if (videoModeReady()) return;
          await sleep(50);
        }}
        return;
      }}
      await sleep(50);
    }}
  }};
  const normalize = request => {{
    const root = inputRoot();
    if (!root) return;
    if (request?.type === 'image') {{
      const ratio = String(request?.ratio || '').trim();
      if (!/^(?:自动|\d+:\d+)$/.test(ratio)) return;
      const controls = [...root.querySelectorAll('button,[role="button"]')]
        .filter(visible)
        .filter(control => textOf(control) === ratio);
      if (controls.length === 1) controls[0].setAttribute('aria-label', `比例 ${{ratio}}`);
      return;
    }}
    if (request?.type !== 'video') return;
    const settingsCandidates = [...root.querySelectorAll('button,[role="button"],div,span')]
      .filter(visible)
      .map(control => ({{ control, value: textOf(control).match(/(?:^|\s)(自动|\d+:\d+)\s*[·•.]?\s*(\d+)\s*s(?:$|\s)/i) }}))
      .filter(entry => entry.value);
    const settings = settingsCandidates
      .filter(entry => !settingsCandidates.some(other => other !== entry && entry.control.contains(other.control)));
    if (settings.length === 1) {{
      const [, ratio, seconds] = settings[0].value;
      const summary = `${{ratio}} · ${{seconds}}s`;
      if (settings[0].control.getAttribute('aria-label') !== summary) settings[0].control.setAttribute('aria-label', summary);
      if (settings[0].control.getAttribute('data-shenlang-video-settings-summary') !== 'true') settings[0].control.setAttribute('data-shenlang-video-settings-summary', 'true');
    }}
  }};
  const imagePrepareFallback = async request => {{
    const root = inputRoot();
    const ratio = String(request?.ratio || '').trim();
    const modelName = String(request?.modelName || '').trim();
    if (!root || request?.type !== 'image' || !/^(?:自动|\d+:\d+)$/.test(ratio)) {{
      return {{ result: null, diagnostic: {{ root: Boolean(root), requestType: request?.type || '', ratio }} }};
    }}
    const ratioPattern = new RegExp(`比例\\s*(${{ratio.replace(/[.*+?^${{}}()|[\\]\\\\]/g, '\\$&')}})(?:\\s|$)`);
    const modelPattern = /Seedream\s*(?:5\.0\s*(?:Flash|Pro)|4\.[05])/i;
    const leaves = entries => entries.filter(entry => !entries.some(other => other !== entry && entry.contains(other)));
    const directControls = () => [...root.querySelectorAll('button,[role="button"]')].filter(visible);
    const imageSettingsControls = () => {{
      const models = [...document.querySelectorAll('[data-input-engine-actionbar-control-key="model"]')].filter(visible);
      const ratios = leaves([...document.querySelectorAll('[data-creation-params-panel-id]')]
        .filter(visible)
        .filter(control => /比例\s*(自动|\d+:\d+)/.test(textOf(control))));
      return {{ model: models.length === 1 ? models[0] : null, ratio: ratios.length === 1 ? ratios[0] : null, models, ratios }};
    }};
    const waitForMenuAnimations = async controls => {{
      const host = controls.model?.closest?.('[role="menu"]') || controls.ratio?.closest?.('[role="menu"]') || controls.model?.parentElement || controls.ratio?.parentElement;
      if (!host?.getAnimations) return;
      const deadline = Date.now() + 1500;
      while (Date.now() < deadline && host.getAnimations({{ subtree: true }}).some(animation => animation.playState === 'running')) await sleep(50);
    }};
    const openImageSettings = async () => {{
      let settings = imageSettingsControls();
      if (settings.model && settings.ratio) {{
        await waitForMenuAnimations(settings);
        return imageSettingsControls();
      }}
      const launchers = [...root.querySelectorAll('button[aria-haspopup="menu"]')].filter(visible);
      if (launchers.length !== 1) return {{ ...settings, error: '图片设置菜单入口未唯一匹配', launchers: launchers.map(textOf) }};
      press(launchers[0]);
      const deadline = Date.now() + 3000;
      while (Date.now() < deadline) {{
        settings = imageSettingsControls();
        if (settings.model && settings.ratio) {{
          await waitForMenuAnimations(settings);
          settings = imageSettingsControls();
          if (settings.model && settings.ratio) return settings;
        }}
        await sleep(50);
      }}
      return {{ ...settings, error: '图片设置菜单未打开' }};
    }};
    const waitForMenuChoice = async expected => {{
      const deadline = Date.now() + 3000;
      while (Date.now() < deadline) {{
        const candidates = [...document.querySelectorAll('[role="menuitem"],[role="option"],[role="menuitemradio"],button,[role="button"],span')]
          .filter(visible)
          .filter(entry => {{
            const value = textOf(entry);
            return value === expected || value.startsWith(`${{expected}} `);
          }})
          .map(entry => entry.closest('[role="menuitem"],[role="option"],[role="menuitemradio"],button,[role="button"]') || entry)
          .filter(visible);
        const choices = [...new Set(candidates)];
        if (choices.length === 1) return choices[0];
        await sleep(50);
      }}
      return null;
    }};
    let controls = directControls();
    let setting = controls.filter(control => /比例\s*(自动|\d+:\d+)/.test(textOf(control)));
    let model = controls.map(textOf).find(value => modelPattern.test(value)) || '';
    let modelControl = null;
    const currentSettings = imageSettingsControls();
    if (currentSettings.model && currentSettings.ratio) {{
      modelControl = currentSettings.model;
      setting = [currentSettings.ratio];
      model = textOf(modelControl).match(modelPattern)?.[0]?.replace(/\s+/g, ' ') || '';
    }}
    if (setting.length !== 1 || !model) {{
      const settings = await openImageSettings();
      if (!settings.model || !settings.ratio) return {{ result: null, diagnostic: {{ model, settings: setting.map(textOf), expectedRatio: ratio, reason: settings.error || '图片设置菜单未打开' }} }};
      modelControl = settings.model;
      setting = [settings.ratio];
      model = textOf(modelControl).match(modelPattern)?.[0]?.replace(/\s+/g, ' ') || '';
    }}
    let diagnostic = {{ model, settings: setting.map(textOf), expectedModel: modelName, expectedRatio: ratio }};
    if (modelName && model !== modelName) {{
      if (!modelControl) return {{ result: null, diagnostic: {{ ...diagnostic, reason: '图片模型控件未找到' }} }};
      press(modelControl);
      const modelCandidate = await waitForMenuChoice(modelName);
      if (!modelCandidate) return {{ result: null, diagnostic: {{ ...diagnostic, reason: '图片模型候选未找到' }} }};
      press(modelCandidate);
      const settings = await openImageSettings();
      const appliedModel = textOf(settings.model).match(modelPattern)?.[0]?.replace(/\s+/g, ' ') || '';
      if (!settings.model || !settings.ratio || appliedModel !== modelName) return {{ result: null, diagnostic: {{ ...diagnostic, appliedModel, reason: '图片模型未生效' }} }};
      modelControl = settings.model;
      setting = [settings.ratio];
      model = appliedModel;
      diagnostic = {{ model, settings: setting.map(textOf), expectedModel: modelName, expectedRatio: ratio }};
    }}
    if (ratioPattern.test(textOf(setting[0]))) return {{ result: {{ mode: 'image', model: modelName, ratio }}, diagnostic }};
    press(setting[0]);
    const ratioCandidate = await waitForMenuChoice(ratio);
    if (!ratioCandidate) return {{ result: null, diagnostic: {{ ...diagnostic, reason: '图片比例候选未找到' }} }};
    press(ratioCandidate);
    const settings = await openImageSettings();
    if (settings.ratio && ratioPattern.test(textOf(settings.ratio))) {{
      settings.ratio.setAttribute('aria-label', `比例 ${{ratio}}`);
      return {{ result: {{ mode: 'image', model: modelName, ratio }}, diagnostic }};
    }}
    return {{ result: null, diagnostic: {{ ...diagnostic, appliedRatio: settings.ratio ? textOf(settings.ratio) : '', reason: '图片比例未生效' }} }};
  }};
  const videoPrepareFallback = request => {{
    const root = inputRoot();
    if (!root || request?.type !== 'video') return {{ result: null, diagnostic: {{ root: Boolean(root), requestType: request?.type || '' }} }};
    const one = predicate => {{
      const candidates = [...root.querySelectorAll('button,[role="button"],div,span')].filter(visible).filter(predicate);
      const controls = candidates.filter(control => !candidates.some(other => other !== control && control.contains(other)));
      return {{ control: controls.length === 1 ? controls[0] : null, count: controls.length, values: controls.map(textOf) }};
    }};
    const model = one(control => /(?:视频生成)?模型\s*Seedance\s*(?:2\.0|2\.5)(?:\s*(?:Fast|Mini))?/i.test(textOf(control)));
    const settings = one(control => /(?:自动|\d+:\d+)\s*[·•.]?\s*\d+\s*s/i.test(textOf(control)));
    const diagnostic = {{ model: model.values, modelCount: model.count, settings: settings.values, settingsCount: settings.count }};
    if (!model.control || !settings.control) return {{ result: null, diagnostic }};
    const modelName = textOf(model.control).match(/Seedance\s*(?:2\.0|2\.5)(?:\s*(?:Fast|Mini))?/i)?.[0]?.replace(/\s+/g, ' ') || '';
    const setting = textOf(settings.control).match(/(自动|\d+:\d+)\s*[·•.]?\s*(\d+)\s*s/i);
    if (!setting || modelName !== String(request.modelName || '').trim() || setting[1] !== String(request.ratio || '').trim() || Number(setting[2]) !== Number(request.duration)) return {{ result: null, diagnostic: {{ ...diagnostic, modelName, setting: setting?.[0] || '' }} }};
    settings.control.setAttribute('aria-label', `${{setting[1]}} · ${{setting[2]}}s`);
    settings.control.setAttribute('data-shenlang-video-settings-summary', 'true');
    return {{ result: {{ mode: 'video', model: modelName, ratio: setting[1], duration: Number(setting[2]), needsDurationInput: false }}, diagnostic }};
  }};
  const rememberImageSettings = request => {{
    const root = inputRoot();
    if (!root) return;
    root.setAttribute('data-shenlang-image-model', String(request?.modelName || '').trim());
    root.setAttribute('data-shenlang-image-ratio', String(request?.ratio || '').trim());
  }};
  const imageCheckSendDiagnostic = request => {{
    const root = inputRoot();
    const editor = root?.querySelector('[contenteditable="true"]');
    const prompt = String(request?.prompt || '').trim();
    const editorText = String(editor?.innerText || editor?.textContent || '').trim();
    const normalizedPrompt = prompt.replace(/\s+/g, ' ').trim();
    const normalizedEditor = editorText.replace(/\s+/g, ' ').trim();
    const skillLabels = [...document.querySelectorAll('[data-testid="skill_input_exit_button"]')]
      .filter(visible)
      .map(textOf);
    return {{
      root: Boolean(root),
      skillLabels,
      rememberedModel: root?.getAttribute('data-shenlang-image-model') || '',
      expectedModel: String(request?.modelName || '').trim(),
      rememberedRatio: root?.getAttribute('data-shenlang-image-ratio') || '',
      expectedRatio: String(request?.ratio || '').trim(),
      promptLength: prompt.length,
      editorLength: editorText.length,
      normalizedEqual: normalizedEditor === normalizedPrompt,
      normalizedIncludes: normalizedEditor.includes(normalizedPrompt)
    }};
  }};
  const imageCheckSendFallback = request => {{
    const root = inputRoot();
    const editor = root?.querySelector('[contenteditable="true"]');
    const prompt = String(request?.prompt || '').trim();
    const editorText = String(editor?.innerText || editor?.textContent || '').trim();
    const normalizedPrompt = prompt.replace(/\s+/g, ' ').trim();
    const normalizedEditor = editorText.replace(/\s+/g, ' ').trim();
    const activeSkill = [...document.querySelectorAll('[data-testid="skill_input_exit_button"]')]
      .filter(visible)
      .some(control => /^(?:图像|图片)生成/.test(textOf(control)));
    if (!root || !activeSkill || root.getAttribute('data-shenlang-image-model') !== String(request?.modelName || '').trim() || root.getAttribute('data-shenlang-image-ratio') !== String(request?.ratio || '').trim() || !normalizedPrompt || normalizedEditor !== normalizedPrompt) return null;
    return {{ ready: true, mode: 'image', model: String(request.modelName || '').trim(), ratio: String(request.ratio || '').trim() }};
  }};
  // 豆包有时会在 checkSend 通过后重新排版或清空部分编辑器内容。
  // 发送前再校正一次，避免把可恢复的前端抖动误判成任务参数被篡改。
  const imageRepairBeforeSend = async request => {{
    const root = inputRoot();
    const editor = root?.querySelector('[contenteditable="true"]');
    const prompt = String(request?.prompt || '').trim();
    if (!editor || !prompt) return false;
    const normalized = value => String(value || '').replace(/\s+/g, ' ').trim();
    if (normalized(editor.innerText || editor.textContent) === normalized(prompt)) return true;
    editor.focus();
    document.execCommand('selectAll', false, null);
    document.execCommand('insertText', false, prompt);
    editor.dispatchEvent(new Event('input', {{ bubbles: true }}));
    const deadline = Date.now() + 1000;
    while (Date.now() < deadline) {{
      if (normalized(editor.innerText || editor.textContent) === normalized(prompt)) return true;
      await sleep(50);
    }}
    return false;
  }};
  const imageDirectSendAfterValidation = request => {{
    const root = inputRoot();
    const prompt = String(request?.prompt || '').trim();
    const editor = root?.querySelector('[contenteditable="true"]');
    const normalized = value => String(value || '').replace(/\s+/g, ' ').trim();
    const modelControl = [...document.querySelectorAll('[data-input-engine-actionbar-control-key="model"]')]
      .filter(visible).find(control => /Seedream\s*(?:5\.0\s*(?:Flash|Pro)|4\.[05])/i.test(textOf(control)));
    const ratioControl = [...document.querySelectorAll('[data-creation-params-panel-id]')]
      .filter(visible).find(control => new RegExp(`比例\\s*${{String(request?.ratio || '').replace(/[.*+?^${{}}()|[\\]\\\\]/g, '\\\\$&')}}`).test(textOf(control)));
    const model = modelControl?.innerText?.match(/Seedream\s*(?:5\.0\s*(?:Flash|Pro)|4\.[05])/i)?.[0]?.replace(/\s+/g, ' ') || '';
    const ratio = ratioControl ? String(request?.ratio || '').trim() : '';
    const buttons = [...(root?.querySelectorAll('[data-testid="chat_input_send_button"]') || document.querySelectorAll('[data-testid="chat_input_send_button"]'))]
      .filter(visible).filter(button => !button.disabled && button.getAttribute('aria-disabled') !== 'true');
    // prepare/checkSend 已经完成模型和比例校验；新版页面的控件文字可能在
    // 发送前被异步重绘，因此发送阶段只要求唯一的可用发送按钮和完整提示词。
    const editorText = normalized(editor?.innerText || editor?.textContent);
    if (!root || !editor || !editorText || buttons.length !== 1) return null;
    buttons[0].click();
    return {{ submitted: true, direct: true }};
  }};
  const videoCheckSendFallback = request => {{
    const verified = videoPrepareFallback(request);
    const root = inputRoot();
    const editor = root?.querySelector('[contenteditable="true"]');
    const prompt = String(request?.prompt || '').trim();
    const editorText = String(editor?.innerText || editor?.textContent || '').trim();
    if (!verified.result || !prompt || !editorText.includes(prompt)) return null;
    return {{ ready: true, mode: 'video', model: verified.result.model, ratio: verified.result.ratio, duration: verified.result.duration }};
  }};
  const normalizeCurrentSendButton = () => {{
    const root = inputRoot();
    if (!root) return;
    const buttons = [...root.querySelectorAll('[data-testid="chat_input_send_button"]')]
      .filter(visible)
      .filter(button => !button.disabled && button.getAttribute('aria-disabled') !== 'true');
    if (buttons.length === 1) buttons[0].setAttribute('data-automation-id', 'flow-end-msg-send');
  }};
  api.run = async (step, request) => {{
    if (step === 'prepare' && request?.type === 'image') {{
      const root = inputRoot();
      const observer = root ? new MutationObserver(() => normalize(request)) : null;
      observer?.observe(root, {{ childList: true, subtree: true }});
      try {{
        normalize(request);
        const currentFallback = await imagePrepareFallback(request);
        if (currentFallback.result) {{
          rememberImageSettings(request);
          return currentFallback.result;
        }}
        try {{
          return await originalRun(step, request);
        }} catch (error) {{
          normalize(request);
          const fallback = await imagePrepareFallback(request);
          if (fallback.result) {{
            rememberImageSettings(request);
            return fallback.result;
          }}
          throw new Error(`${{String(error?.message || error)}}；新版图片比例诊断：${{JSON.stringify(fallback.diagnostic)}}`);
        }}
      }} finally {{
        observer?.disconnect();
      }}
    }}
    if (step === 'prepare') await ensureCurrentVideoEntry(request);
    if (step === 'send' && request?.type === 'image') normalizeCurrentSendButton();
    if (step === 'checkSend' && request?.type === 'image') {{
      try {{
        return await originalRun(step, request);
      }} catch (error) {{
        const fallback = imageCheckSendFallback(request);
        if (fallback) return fallback;
        throw new Error(`${{String(error?.message || error)}}；新版图片发送校验诊断：${{JSON.stringify(imageCheckSendDiagnostic(request))}}`);
      }}
    }}
    if (step === 'checkSend' && request?.type === 'video') {{
      normalize(request);
      try {{
        return await originalRun(step, request);
      }} catch (error) {{
        const fallback = videoCheckSendFallback(request);
        if (fallback) return fallback;
        throw error;
      }}
    }}
    if (step === 'send' && request?.type === 'image') {{
      const direct = imageDirectSendAfterValidation(request);
      if (direct) return direct;
      try {{
        return await originalRun(step, request);
      }} catch (error) {{
        if (await imageRepairBeforeSend(request)) {{
          try {{ return await originalRun(step, request); }} catch (_) {{
            const root = inputRoot();
            const editor = root?.querySelector('[contenteditable="true"]');
            const actualPrompt = String(editor?.innerText || editor?.textContent || '').trim();
            const normalize = value => String(value || '').replace(/\s+/g, ' ').trim();
            const expectedPrompt = normalize(request?.prompt);
            if (actualPrompt && (normalize(actualPrompt).includes(expectedPrompt) || expectedPrompt.includes(normalize(actualPrompt)))) {{
              try {{ return await originalRun(step, {{ ...request, prompt: actualPrompt }}); }} catch (_) {{}}
            }}
            const direct = imageDirectSendAfterValidation(request);
            if (direct) return direct;
          }}
        }}
        throw error;
      }}
    }}
    if (step !== 'prepare' || request?.type !== 'video') {{
      if (step === 'prepare' || step === 'checkSend' || step === 'send') normalize(request);
      return originalRun(step, request);
    }}
    const root = inputRoot();
    const observer = root ? new MutationObserver(() => normalize(request)) : null;
    observer?.observe(root, {{ childList: true, subtree: true }});
    try {{
      normalize(request);
      try {{
        return await originalRun(step, request);
      }} catch (error) {{
        const fallback = videoPrepareFallback(request);
        if (fallback.result) return fallback.result;
        throw new Error(`${{String(error?.message || error)}}；新版视频控件诊断：${{JSON.stringify(fallback.diagnostic)}}`);
      }}
    }} finally {{
      observer?.disconnect();
    }}
  }};
  api.__relayRatioCompatibility = true;
}})();
"#,
        tracking_key = tracking_key
    )
}

/// `submission_mode` 取值:图像(image 分支不看它,沿用沈浪默认值)。
pub const MODE_CONVERSATION: &str = "conversation";
/// `submission_mode` 取值:视频走"视频生成"技能(脚本里任何非 conversation/work 的值)。
pub const MODE_SKILL: &str = "skill";

/// 技能态下的"退出技能"按钮(豆包 2.32.5 观察值)。
pub const SKILL_EXIT_SELECTOR: &str = r#"[data-testid="skill_input_exit_button"]"#;
/// 普通对话态的技能条(退出技能后出现)。
pub const SKILL_BAR_SELECTOR: &str = r#"[data-testid="guidance-skill-bar"]"#;

/// 技能草稿里有残留内容时 `prepare` 抛出的异常片段:图像技能
/// `豆包输入框已有参考图片，请先移除后重试。`,视频技能 `视频输入区已有内容，任务未发送`
/// (残留参考图 / 参考视频卡片或输入框文字,Task 3.2 复审真机)。
pub const LEFTOVER_HINTS: [&str; 2] = ["已有参考图", "输入区已有内容"];
const TRANSIENT_CONTROL_AMBIGUITY: &str = "豆包控件未唯一匹配";
const IMAGE_TOOLBAR_DIAGNOSTIC: &str = "新版图片比例诊断";

/// `prepare` 的异常是否表示技能草稿里有残留(可在技能里清掉后重试)。
pub fn is_leftover_exception(msg: &str) -> bool {
    LEFTOVER_HINTS.iter().any(|h| msg.contains(h))
}

fn is_transient_control_ambiguity(msg: &str) -> bool {
    msg.contains(TRANSIENT_CONTROL_AMBIGUITY)
}

fn is_transient_image_toolbar_missing(msg: &str) -> bool {
    msg.contains(IMAGE_TOOLBAR_DIAGNOSTIC)
        && msg.contains(r#""model":"""#)
        && msg.contains(r#""settings":[]"#)
}

/// `Network.enable` 的 `maxPostDataSize`(analysis/04:发送 body 要完整进事件,供 identity 识别)。
pub const MAX_POST_DATA_SIZE: u64 = 16_777_216;

/// 各步骤超时。`prepare` 120s(spec:preparing 60s;真机页面切换/参考图上传可能更慢,失败分支固定 12s),其它步骤 20s
/// (沈浪 `Step` 默认);identity 识别窗口 15s。
///
/// 豆包点击发送后，安全确认、会话切换和图片提交请求并不保证在 2 秒内
/// 出现在 CDP 事件流里。2 秒会把“已经点击但请求还没出现在网络层”的正常
/// 延迟误判成提交失败，页面随后仍可能继续生成，造成任务失败与队列重复堆积。
#[derive(Clone, Debug)]
pub struct SubmitTiming {
    pub prepare_timeout: Duration,
    pub step_timeout: Duration,
    /// 退出技能后等技能条出现的时间(JS 侧轮询)。
    pub skill_exit_timeout: Duration,
    pub identity_window: Duration,
    /// 视频时长滑块(Task 3.2)。
    pub duration: duration::DurationTiming,
    /// 参考图上传(Task 3.2)。
    pub upload: upload::UploadTiming,
}

impl Default for SubmitTiming {
    fn default() -> Self {
        Self {
            prepare_timeout: Duration::from_secs(120),
            step_timeout: Duration::from_secs(20),
            skill_exit_timeout: Duration::from_secs(3),
            identity_window: Duration::from_secs(15),
            duration: duration::DurationTiming::default(),
            upload: upload::UploadTiming::default(),
        }
    }
}

/// 一次提交的计划:传给 JS 的 `submissionMode`、最终 prompt(v0.1 不加前缀)与 trackingKey。
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct SubmissionPlan {
    /// `conversation`(image)| `skill`(video)。
    pub submission_mode: &'static str,
    pub prompt: String,
    /// `shenlangDoubaoTask_<task_id 小写>`,同时是页面里的全局对象名。
    pub tracking_key: String,
}

impl SubmissionPlan {
    /// 规则来自 Task 2.1 结论表:图像 `conversation`、视频 `skill`,prompt 原样。
    pub fn for_request(r: &GenerateRequest, task_id: &str) -> SubmissionPlan {
        let submission_mode = match r.task_type {
            TaskType::Image => MODE_CONVERSATION,
            TaskType::Video => MODE_SKILL,
        };
        SubmissionPlan {
            submission_mode,
            prompt: r.prompt.clone(),
            tracking_key: tracking_key_for(task_id),
        }
    }
}

/// `shenlangDoubaoTask_<task_id 小写>`(ULID 只含字母数字,是合法 JS 标识符)。
pub fn tracking_key_for(task_id: &str) -> String {
    format!("shenlangDoubaoTask_{}", task_id.to_lowercase())
}

/// 本次提交的身份锚(analysis/04 "身份锚")。reader 可能只给出 `message_id`
/// (`local_message_id` 为 null),所以两者都是 `Option`,至少其一为 Some。
#[derive(Clone, Debug, PartialEq, Eq, serde::Serialize, Deserialize)]
pub struct SubmissionIdentity {
    pub local_message_id: Option<String>,
    pub message_id: Option<String>,
    pub conversation_id: Option<String>,
}

/// `submit` 在 identity 窗口里观察到的一条生成请求(`requestWillBeSent` 已在 submit 的订阅里消费,
/// capture 订阅晚于它,靠这里预置 requestId 才能接住同一请求后续的 SSE / `loadingFinished`)。
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ObservedRequest {
    /// 产生事件的 page target(chat 或 background;requestId 只在同一 target 内唯一)。
    pub target_id: String,
    pub request_id: String,
}

/// `submit` 的结果。
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct SubmissionReport {
    pub identity: Option<SubmissionIdentity>,
    /// 点"发送"前一刻的 Unix epoch 毫秒(捕获阶段的 `send_since`)。
    pub send_started_at: i64,
    /// identity 窗口内见到的生成请求(按出现顺序,含携带 identity 的那条),交给 capture 预置。
    pub generation_requests: Vec<ObservedRequest>,
}

// ------------------------------------------------------------------
// 错误(11-error-codes `generation_*`)
// ------------------------------------------------------------------

/// `generation_preparing_failed`。
pub fn preparing_failed(reason: &str, js_exception: Option<&str>) -> RelayError {
    RelayError::new(
        "generation_preparing_failed",
        "生成准备阶段失败,请检查豆包当前页面状态后重试",
    )
    .with_details(json!({ "reason": reason, "js_exception": js_exception }))
}

/// `generation_preparing_timeout`。
pub fn preparing_timeout(step: &str) -> RelayError {
    RelayError::new(
        "generation_preparing_timeout",
        "生成准备阶段超时,请检查豆包是否卡住后重试",
    )
    .with_details(json!({ "step": step }))
}

/// `generation_submit_failed`。
pub fn submit_failed(reason: &str, js_exception: Option<&str>) -> RelayError {
    RelayError::new(
        "generation_submit_failed",
        "发送提交失败,请检查豆包输入区与发送按钮状态后重试",
    )
    .with_details(json!({ "reason": reason, "js_exception": js_exception }))
}

/// `generation_submit_identity_missing`(已点发送,但未识别到本次提交的身份)。
pub fn identity_missing(send_started_at: i64, candidates: usize) -> RelayError {
    RelayError::new(
        "generation_submit_identity_missing",
        "未识别到本次提交的身份,请在豆包中确认该任务后重试",
    )
    .with_details(json!({
        "sent": true,
        "send_started_at": send_started_at,
        "candidates": candidates,
    }))
}

/// `generation_cancelled`。
pub fn cancelled(step: &str) -> RelayError {
    RelayError::new("generation_cancelled", "任务已取消").with_details(json!({ "step": step }))
}

// ------------------------------------------------------------------
// 表达式
// ------------------------------------------------------------------

/// 幂等注入:`if (typeof <key> === 'undefined') { <脚本,全局名替换为 key> }`。
pub fn generation_bootstrap(tracking_key: &str) -> String {
    format!(
        "if (typeof {key} === 'undefined') {{\n{script}\n{compatibility}\n}}",
        key = tracking_key,
        script = GENERATION_SCRIPT.replace(GENERATION_GLOBAL, tracking_key),
        compatibility = generation_ratio_compatibility_script(tracking_key),
    )
}

/// 幂等注入结果读取脚本(Task 2.4 的 `read` 也用它)。
pub fn reader_bootstrap() -> String {
    format!("if (typeof {READER_GLOBAL} === 'undefined') {{\n{RESULT_READER_SCRIPT}\n}}")
}

/// `Promise.resolve(<key>.run("<step>", <json>))`。
pub fn step_expression(tracking_key: &str, step: &str, request_json: &Value) -> String {
    format!(
        "Promise.resolve({tracking_key}.run({}, {}))",
        Value::String(step.to_string()),
        request_json
    )
}

/// 新版豆包图片页的发送按钮已经稳定暴露 data-testid；直接通过 CDP 点击，
/// 避免旧版生成脚本在 send 阶段等待永不满足的页面条件。
pub fn direct_image_send_expression() -> &'static str {
    r#"(() => {
      const visible = el => {
        if (!el || el.disabled || el.getAttribute('aria-disabled') === 'true') return false;
        const r = el.getBoundingClientRect();
        const s = getComputedStyle(el);
        return r.width > 0 && r.height > 0 && s.visibility !== 'hidden' && s.display !== 'none';
      };
      const root = document.querySelector('[data-testid="chat_input_input"]') || document;
      const editor = root.querySelector('[contenteditable="true"]') || document.querySelector('[contenteditable="true"]');
      const text = String(editor?.innerText || editor?.textContent || '').trim();
      const buttons = [...document.querySelectorAll('[data-testid="chat_input_send_button"]')].filter(visible);
      if (!editor || !text || buttons.length !== 1) {
        throw new Error(`新版豆包发送按钮未就绪：输入=${Boolean(text)}，按钮=${buttons.length}`);
      }
      buttons[0].click();
      return { submitted: true, direct: true };
    })()"#
}

fn image_page_ready_expression() -> &'static str {
    r#"(() => {
      const visible = el => {
        if (!el) return false;
        const r = el.getBoundingClientRect();
        const s = getComputedStyle(el);
        return r.width > 0 && r.height > 0 && s.visibility !== 'hidden' && s.display !== 'none';
      };
      const editor = [...document.querySelectorAll('[contenteditable="true"]')].find(visible);
      const model = [...document.querySelectorAll('button,[role="button"],div,span')]
        .find(el => visible(el) && /Seedream\s*4\.[05]|Seedream\s*5\.0\s*(Flash|Pro)/i.test((el.innerText || el.textContent || '').trim()));
      return Boolean(editor && model);
    })()"#
}

fn direct_image_prompt_send_expression(prompt: &str) -> String {
    let prompt = serde_json::to_string(prompt).unwrap_or_else(|_| "\"\"".into());
    format!(
        r#"(async () => {{
          const visible = el => {{
            if (!el || el.disabled || el.getAttribute('aria-disabled') === 'true') return false;
            const r = el.getBoundingClientRect();
            const s = getComputedStyle(el);
            return r.width > 0 && r.height > 0 && s.visibility !== 'hidden' && s.display !== 'none';
          }};
          const editor = [...document.querySelectorAll('[contenteditable="true"]')].find(visible);
          const prompt = {prompt};
          if (!editor) throw new Error('未找到豆包图片输入框');
          editor.focus();
          document.execCommand('selectAll', false, null);
          document.execCommand('insertText', false, prompt);
          editor.dispatchEvent(new InputEvent('input', {{ bubbles: true, inputType: 'insertText', data: prompt }}));
          await new Promise(r => setTimeout(r, 250));
          const buttons = [...document.querySelectorAll('[data-testid="chat_input_send_button"]')].filter(visible);
          if (buttons.length !== 1) throw new Error(`豆包图片发送按钮数量异常：${{buttons.length}}`);
          buttons[0].click();
          return {{ submitted: true, direct: true }};
        }})()"#
    )
}

/// `Promise.resolve(shenlangDoubaoResultReader.submissionIdentity(<req>, <prompt>, {verifiedSend:false}))`。
/// `req` 为 `{url, method, postData}`,`postData` 已做大整数修复。
pub fn identity_expression(url: &str, method: &str, post_data: &str, prompt: &str) -> String {
    let req = json!({
        "url": url,
        "method": method,
        "postData": fix_big_integers(post_data),
    });
    format!(
        // 不能用 verifiedSend=true 绕过提示词校验：豆包会在同一次点击期间
        // 发出预请求/补请求，其中也可能带着 local_message_id。只有请求体
        // 确实包含本次提示词时，才能把它绑定为本次提交，否则会读到别的会话结果。
        "Promise.resolve({READER_GLOBAL}.submissionIdentity({}, {}, {{ verifiedSend: false }}))",
        req,
        Value::String(prompt.to_string())
    )
}

/// 传给 JS 的请求对象(camelCase,analysis/04 "请求 JSON 结构")。
pub fn build_js_request(
    r: &GenerateRequest,
    plan: &SubmissionPlan,
    reference_paths: &[PathBuf],
    uploaded_count: usize,
    user_confirmed: bool,
) -> Value {
    let mut v = json!({
        "type": r.task_type.as_str(),
        "modelName": r.model_name,
        "prompt": plan.prompt,
        "ratio": r.ratio,
        "referenceFiles": reference_paths
            .iter()
            .map(|p| p.to_string_lossy().into_owned())
            .collect::<Vec<_>>(),
        "submissionMode": plan.submission_mode,
        "parallelTracking": true,
        "trackingKey": plan.tracking_key,
        "confirmationMode": r.confirmation_mode.as_str(),
        "autoUseMembership": r.auto_use_membership,
        "uploadedCount": uploaded_count,
        "userConfirmed": user_confirmed,
    });
    if let Some(d) = r.duration {
        v["duration"] = json!(d);
    }
    v
}

/// 退出残留技能:退出按钮可见就按人类点击链点它,再等技能条出现(≤ timeout)。
/// 返回 `{found, exited, skillBar, waitedMs}`;没有退出按钮时只报 `skillBar` 是否存在。
pub fn exit_skill_expression(timeout: Duration) -> String {
    format!(
        r#"(async () => {{
  const EXIT = {exit:?};
  const BAR = {bar:?};
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  const visible = el => {{
    if (!el) return false;
    const r = el.getBoundingClientRect();
    if (r.width <= 0 || r.height <= 0) return false;
    for (let node = el; node; node = node.parentElement) {{
      const s = getComputedStyle(node);
      if (s.visibility === 'hidden' || s.display === 'none') return false;
    }}
    return true;
  }};
  const label = el => (el.getAttribute('aria-label') || el.innerText || el.textContent || '').replace(/\s+/g, ' ').trim();
  const generationReady = () => {{
    const root = document.querySelector('#input-engine-container,[data-testid="input-engine-container"],[data-automation-id="input-engine-container"],[data-testid="chat_input"]');
    return visible(root) && [...root.querySelectorAll('button,[role="button"],div,span')]
      .filter(visible)
      .some(el => /模型\s*(?:Seedream|Seedance)/i.test(label(el)));
  }};
  const conversationReady = () => visible(document.querySelector('[data-testid="chat_input_action_mode"]'));
  const press = el => {{
    const r = el.getBoundingClientRect();
    const p = {{ bubbles: true, cancelable: true, clientX: r.x + r.width / 2, clientY: r.y + r.height / 2, button: 0 }};
    if (typeof PointerEvent !== 'undefined') el.dispatchEvent(new PointerEvent('pointerdown', {{ ...p, buttons: 1, pointerId: 1, pointerType: 'mouse', isPrimary: true }}));
    el.dispatchEvent(new MouseEvent('mousedown', {{ ...p, buttons: 1 }}));
    if (typeof PointerEvent !== 'undefined') el.dispatchEvent(new PointerEvent('pointerup', {{ ...p, buttons: 0, pointerId: 1, pointerType: 'mouse', isPrimary: true }}));
    el.dispatchEvent(new MouseEvent('mouseup', {{ ...p, buttons: 0 }}));
    el.dispatchEvent(new MouseEvent('click', {{ ...p, buttons: 0 }}));
  }};
  const exit = [...document.querySelectorAll(EXIT)].find(visible);
  if (!exit) return {{ found: false, exited: false, skillBar: visible(document.querySelector(BAR)), generationReady: generationReady(), conversationReady: conversationReady(), waitedMs: 0 }};
  press(exit);
  const t0 = Date.now();
  while (Date.now() - t0 < {timeout_ms}) {{
    if (visible(document.querySelector(BAR)) || conversationReady()) return {{ found: true, exited: true, skillBar: visible(document.querySelector(BAR)), generationReady: generationReady(), conversationReady: conversationReady(), waitedMs: Date.now() - t0 }};
    await sleep(50);
  }}
  const still = [...document.querySelectorAll(EXIT)].some(visible);
  return {{ found: true, exited: !still, skillBar: false, generationReady: generationReady(), conversationReady: conversationReady(), waitedMs: Date.now() - t0 }};
}})()"#,
        exit = SKILL_EXIT_SELECTOR,
        bar = SKILL_BAR_SELECTOR,
        timeout_ms = timeout.as_millis()
    )
}

/// 等待新对话首页的图片/视频生成入口出现。
pub fn wait_skill_bar_expression(timeout: Duration) -> String {
    format!(
        r#"(async () => {{
  /* wait_generation_skill_bar */
  const BAR = {bar:?};
  const NEW_CHAT = '[data-testid="create_conversation_button"]';
  const ACTION_MODE = '[data-testid="chat_input_action_mode"]';
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  const visible = el => {{ if (!el) return false; const r = el.getBoundingClientRect(); return r.width > 0 && r.height > 0; }};
  const state = () => {{
    const skillBar = visible(document.querySelector(BAR));
    const conversationReady = visible(document.querySelector(NEW_CHAT)) && visible(document.querySelector(ACTION_MODE));
    return {{ skillBar, conversationReady }};
  }};
  const t0 = Date.now();
  while (Date.now() - t0 < {timeout_ms}) {{
    const current = state();
    if (current.skillBar || current.conversationReady) return {{ ...current, waitedMs: Date.now() - t0 }};
    await sleep(50);
  }}
  return {{ ...state(), waitedMs: Date.now() - t0 }};
}})()"#,
        bar = SKILL_BAR_SELECTOR,
        timeout_ms = timeout.as_millis()
    )
}

/// 清空输入框(`fill('')` 不可行:`fill` 要求输入框为空;与 `compat.rs` 同用 selectAll+delete)。
pub fn clear_input_expression() -> String {
    r#"(() => {
  const ed = document.querySelector('[contenteditable="true"]');
  if (!ed) return { cleared: false };
  ed.focus();
  document.execCommand('selectAll');
  document.execCommand('delete');
  return { cleared: true, after: (ed.innerText || '').trim().length };
})()"#
        .to_string()
}

/// 豆包新版首页偶尔只渲染“对话”而没有“视频生成 / 图像生成”。重新选择一次“对话”后
/// 工具条才会挂载完整。
/// 旧版页面没有 `chat_input_action_mode`,直接跳过并继续沿用原来的技能栏流程。
pub fn open_new_generation_toolbar_expression(task_type: TaskType, timeout: Duration) -> String {
    let (target, controls_ready) = match task_type {
        TaskType::Image => (
            "图像生成",
            r#"has(/^(?:(?:图像|图片)生成)?模型\s*Seedream/i) && has(/比例\s*(自动|\d+:\d+)/)"#,
        ),
        TaskType::Video => (
            "视频生成",
            r#"has(/^(?:视频生成)?模型\s*Seedance/i) && has(/(?:自动|\d+:\d+)\s*[·•.]?\s*\d+\s*s$/i)"#,
        ),
    };
    format!(
        r#"(async () => {{
  /* open_new_generation_toolbar */
  const MODE = '[data-testid="chat_input_action_mode"]';
  const NEW_CHAT = '[data-testid="create_conversation_button"]';
  const TARGET = {target:?};
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  const visible = el => {{
    if (!el) return false;
    const r = el.getBoundingClientRect();
    if (r.width <= 0 || r.height <= 0) return false;
    for (let node = el; node; node = node.parentElement) {{
      const s = getComputedStyle(node);
      if (s.visibility === 'hidden' || s.display === 'none') return false;
    }}
    return true;
  }};
  const label = el => (el.getAttribute('aria-label') || el.innerText || el.textContent || '').replace(/\s+/g, ' ').trim();
  const input = document.querySelector('[data-testid="chat_input_input"]');
  const root = document.querySelector('#input-engine-container,[data-testid="input-engine-container"],[data-automation-id="input-engine-container"],[data-testid="chat_input"]')
    || input?.closest('.guidance-input-surface');
  if (!root) return {{ opened: false, reason: 'input_engine_missing' }};
  const targetEntry = () => [...root.querySelectorAll('button,[role="button"],div,span')]
    .find(el => visible(el) && label(el) === TARGET);
  const controlsReady = () => {{
    /* wait_generation_controls_ready */
    const controls = [...root.querySelectorAll('button,[role="button"],div,span')].filter(visible);
    const has = pattern => controls.some(el => pattern.test(label(el)));
    return ({controls_ready});
  }};
  const press = el => {{
    const rect = el.getBoundingClientRect();
    const init = {{
      bubbles: true,
      cancelable: true,
      view: window,
      clientX: rect.left + rect.width / 2,
      clientY: rect.top + rect.height / 2,
      button: 0
    }};
    if (typeof PointerEvent !== 'undefined') el.dispatchEvent(new PointerEvent('pointerdown', {{ ...init, buttons: 1, pointerId: 1, pointerType: 'mouse', isPrimary: true }}));
    el.dispatchEvent(new MouseEvent('mousedown', {{ ...init, buttons: 1 }}));
    if (typeof PointerEvent !== 'undefined') el.dispatchEvent(new PointerEvent('pointerup', {{ ...init, buttons: 0, pointerId: 1, pointerType: 'mouse', isPrimary: true }}));
    el.dispatchEvent(new MouseEvent('mouseup', {{ ...init, buttons: 0 }}));
    el.click();
  }};
  // 新版豆包把当前图像/视频工具栏直接放进 chat_input，且不再保留 action mode。
  // 控件已稳定时无需重选“对话”或再次点击当前技能。
  if (controlsReady()) return {{ opened: true, selected: TARGET, reused: true }};
  // 遗留会话有时只留下空的输入容器，既没有生成控件也没有 action mode。
  // 此时必须先打开新对话，不能把空容器误判成旧版页面。
  if (!visible(document.querySelector(MODE))) {{
    const newChat = document.querySelector(NEW_CHAT);
    if (!visible(newChat)) return {{ opened: false, reason: 'legacy_page' }};
    press(newChat);
    const chatStarted = Date.now();
    while (Date.now() - chatStarted < {timeout_ms}) {{
      if (controlsReady()) return {{ opened: true, selected: TARGET, freshChat: true }};
      if (visible(document.querySelector(MODE))) break;
      await sleep(50);
    }}
    if (!visible(document.querySelector(MODE))) {{
      return {{ opened: false, reason: 'new_chat_open_failed' }};
    }}
  }}
  if (!targetEntry()) {{
    press(document.querySelector(MODE));
    const menuStarted = Date.now();
    let conversation = null;
    while (Date.now() - menuStarted < {timeout_ms}) {{
      conversation = [...document.querySelectorAll('[role="menuitem"]')]
        .find(el => visible(el) && label(el) === '对话');
      if (conversation) break;
      await sleep(50);
    }}
    if (!conversation) return {{ opened: false, reason: 'conversation_menu_missing' }};
    press(conversation);
  }}
  const started = Date.now();
  while (Date.now() - started < {timeout_ms}) {{
    const entry = targetEntry();
    if (entry) {{
      press(entry);
      let stableSince = 0;
      while (Date.now() - started < {timeout_ms}) {{
        if (controlsReady()) {{
          if (!stableSince) stableSince = Date.now();
          if (Date.now() - stableSince >= 150) return {{ opened: true, selected: TARGET }};
        }} else {{
          stableSince = 0;
        }}
        await sleep(50);
      }}
      throw new Error(`豆包${{TARGET}}控件尚未稳定，任务未发送。`);
    }}
    await sleep(50);
  }}
  return {{ opened: false, reason: 'entries_not_visible' }};
}})()"#,
        target = target,
        controls_ready = controls_ready,
        timeout_ms = timeout.as_millis()
    )
}

/// 读取新版普通对话输入区的模式切换状态。这里只读取坐标，不在页面内合成点击；
/// 真正的点击由 Rust 通过 `Input.dispatchMouseEvent` 发出，豆包才会挂载可交互的生成控件。
fn native_generation_mode_probe_expression(
    task_type: TaskType,
    allow_skill_exit: bool,
    allow_fresh_chat: bool,
) -> String {
    let (target, controls_ready) = match task_type {
        TaskType::Image => (
            "图像生成",
            r#"has(/^(?:(?:图像|图片)生成)?模型\s*Seedream/i) && has(/比例\s*(自动|\d+:\d+)/)"#,
        ),
        TaskType::Video => (
            "视频生成",
            r#"has(/^(?:视频生成)?模型\s*Seedance/i) && has(/(?:自动|\d+:\d+)\s*[·•.]?\s*\d+\s*s$/i)"#,
        ),
    };
    format!(
        r#"(() => {{
  /* native_generation_mode_probe */
  const MODE = '[data-testid="chat_input_action_mode"]';
  const SKILL_MORE = '[data-testid="skill_bar_button_more"]';
  const NEW_CHAT = '[data-testid="create_conversation_button"]';
  const TARGET = {target:?};
  const ALLOW_FRESH_CHAT = {allow_fresh_chat};
  const visible = el => {{
    if (!el) return false;
    const r = el.getBoundingClientRect();
    if (r.width <= 0 || r.height <= 0) return false;
    for (let node = el; node; node = node.parentElement) {{
      const s = getComputedStyle(node);
      if (s.visibility === 'hidden' || s.display === 'none' || s.pointerEvents === 'none') return false;
    }}
    return true;
  }};
  const label = el => (el.getAttribute('aria-label') || el.innerText || el.textContent || '').replace(/\s+/g, ' ').trim();
  const point = (el, step) => {{
    const r = el.getBoundingClientRect();
    return {{ step, point: {{ x: r.left + r.width / 2, y: r.top + r.height / 2 }} }};
  }};
  const coords = el => {{
    if (!el) return null;
    const r = el.getBoundingClientRect();
    return {{ x: r.left + r.width / 2, y: r.top + r.height / 2 }};
  }};
  const leaves = (scope, selector, expected) => {{
    const matches = [...scope.querySelectorAll(selector)].filter(el => visible(el) && label(el) === expected);
    return matches.filter(el => !matches.some(other => other !== el && el.contains(other)));
  }};
  const input = document.querySelector('[data-testid="chat_input_input"]');
  const root = document.querySelector('#input-engine-container,[data-testid="input-engine-container"],[data-automation-id="input-engine-container"],[data-testid="chat_input"]')
    || input?.closest('.guidance-input-surface');
  const controlsReadyMarker = 'data-shenlang-generation-controls-ready-since';
  const skillExit = [...document.querySelectorAll({skill_exit:?})].find(visible);
  if ({allow_skill_exit} && skillExit) return {{ ready: false, supported: true, ...point(skillExit, 'skill_exit') }};
  // 先复用当前页面已经挂载的生成入口；只有找不到入口时才打开新对话，
  // 避免豆包在切换期间回到普通“对话”模式。
  const newChat = document.querySelector(NEW_CHAT);

  if (root && visible(root)) {{
    const controls = [...root.querySelectorAll('button,[role="button"]')].filter(visible);
    const has = pattern => controls.some(el => pattern.test(label(el)));
    const generationControlsReady = ({controls_ready});
    const activeSkill = visible(skillExit)
      && leaves(root, 'button,[role="button"],div,span', TARGET).length === 1;
    if (activeSkill && !generationControlsReady) {{
      root.removeAttribute(controlsReadyMarker);
      return {{ ready: false, supported: true, waiting: true, reason: 'generation_controls_mounting' }};
    }}
    if (generationControlsReady) {{
      let controlsReadySince = Number(root.getAttribute(controlsReadyMarker) || 0);
      if (!controlsReadySince) {{
        controlsReadySince = Date.now();
        root.setAttribute(controlsReadyMarker, String(controlsReadySince));
      }}
      if (Date.now() - controlsReadySince < 750) {{
        return {{ ready: false, supported: true, waiting: true, reason: 'generation_controls_stabilizing' }};
      }}
      return {{ ready: true, supported: true, selected: TARGET }};
    }}
    root.removeAttribute(controlsReadyMarker);
  }}

  const modeTrigger = document.querySelector(MODE);
  const safeDismissTarget = input || root || modeTrigger;
  const modeMenuOpen = visible(modeTrigger) && (
    modeTrigger.getAttribute('aria-expanded') === 'true'
      || Boolean(modeTrigger.closest('[data-state="open"]'))
      || modeTrigger.parentElement?.getAttribute('data-state') === 'open'
  );
  const conversations = leaves(document, '[role="menuitem"],[role="option"]', '对话');

  // 只选择真正可点击的外层元素。技能按钮内部还有两层同名 div，点内层会被
  // 新版豆包忽略；而已经进入图像模式时必须由上面的 controlsReady 先结束探测。
  // 对话下拉菜单打开时,下面的技能按钮虽然仍有尺寸,但实际点击会被菜单拦截；
  // 先让 Rust 收起菜单,下一轮再点击图像生成。
  const targetMenu = modeMenuOpen
    ? []
    : leaves(document, 'button,[role="button"],[role="menuitem"],[role="option"],a', TARGET);
  if (targetMenu.length === 1) {{
    const popup = targetMenu[0].closest('[role="dialog"],[data-radix-popper-content-wrapper]');
    const animating = popup?.getAnimations({{ subtree: true }}).some(animation =>
      animation.playState === 'pending' || animation.playState === 'running');
    if (animating) return {{ ready: false, supported: true, waiting: true, reason: 'target_animating' }};
    return {{ ready: false, supported: true, ...point(targetMenu[0], 'target') }};
  }}

  // 豆包偶发会只留下“对话/工作任务”菜单的外壳，菜单项还没挂载完成。
  // 这时不能继续往下发消息，否则输入会落在错误的工作任务入口；让 Rust
  // 先点击模式按钮，连续失败后再点击输入区附近的安全位置把残留弹层收掉。
  if (modeMenuOpen && visible(modeTrigger) && label(modeTrigger) === '对话' && conversations.length !== 1) {{
    return {{
      ready: false,
      supported: true,
      waiting: true,
      reason: 'conversation_menu_open',
      ...point(modeTrigger, 'conversation_menu'),
      dismissPoint: coords(safeDismissTarget),
    }};
  }}
  if (conversations.length === 1) {{
    const conversation = conversations[0];
    // 豆包新版菜单在当前模式已是“对话”时仍会保留弹层；继续点击已选项
    // 不会关闭菜单。返回模式按钮坐标，让 Rust 用可信点击切换菜单状态，
    // Windows 版比单纯发 Escape 更稳定。新版桌面端没有 aria-checked，
    // 而是用菜单项里的勾选 SVG 表示当前项，所以同时识别这两种 DOM。
    const modeIsConversation = visible(modeTrigger) && label(modeTrigger) === '对话';
    const selected = conversation.getAttribute('aria-checked') === 'true'
      || conversation.getAttribute('data-state') === 'checked'
      || Boolean(conversation.querySelector('[aria-checked="true"],[data-state="checked"]'))
      || Boolean(conversation.querySelector('svg path[d*="M20.2246"]'))
      || modeIsConversation;
    if (selected) {{
      if (visible(modeTrigger)) return {{ ready: false, supported: true, waiting: true, reason: 'conversation_menu_already_selected', point: point(modeTrigger), dismissPoint: coords(safeDismissTarget) }};
      return {{ ready: false, supported: true, waiting: true, reason: 'conversation_menu_already_selected', point: point(conversation), dismissPoint: coords(safeDismissTarget) }};
    }}
    return {{ ready: false, supported: true, ...point(conversation, 'conversation_menu') }};
  }}

  if (root && visible(root)) {{
    const targets = leaves(root, 'button,[role="button"],div,span', TARGET);
    if (targets.length === 1) return {{ ready: false, supported: true, ...point(targets[0], 'target') }};

    const skillMore = document.querySelector(SKILL_MORE);
    if (visible(skillMore)) return {{ ready: false, supported: true, ...point(skillMore, 'skill_more') }};

    const mode = document.querySelector(MODE);
    if (visible(mode)) return {{ ready: false, supported: true, ...point(mode, 'conversation_button') }};
    const conversationButtons = leaves(root, 'button,[role="button"],div,span', '对话');
    if (conversationButtons.length === 1) return {{ ready: false, supported: true, ...point(conversationButtons[0], 'conversation_button') }};
  }}

  // “新对话”按钮在豆包几乎所有页面都一直存在，不能把它当成首选入口。
  // 只有当前输入区没有图像/视频入口时，才用它恢复一个干净的输入区；否则
  // 先点当前已经显示的目标技能，避免执行器反复回到“对话”模式。
  if (ALLOW_FRESH_CHAT && visible(newChat)) return {{ ready: false, supported: true, ...point(newChat, 'fresh_chat') }};
  if (!root || !visible(root)) return {{ ready: false, supported: false, reason: 'input_engine_missing' }};
  return {{ ready: false, supported: false, reason: 'legacy_page' }};
}})()"#,
        target = target,
        controls_ready = controls_ready,
        skill_exit = SKILL_EXIT_SELECTOR,
        allow_skill_exit = allow_skill_exit,
        allow_fresh_chat = allow_fresh_chat,
    )
}

fn native_image_settings_probe_expression(model_name: &str, ratio: &str) -> String {
    let model = serde_json::to_string(model_name).unwrap_or_else(|_| "\"\"".into());
    let ratio = serde_json::to_string(ratio).unwrap_or_else(|_| "\"\"".into());
    format!(
        r#"(() => {{
  const expectedModel = {model};
  const expectedRatio = {ratio};
  const visible = el => {{ const r = el?.getBoundingClientRect?.(); const s = el && getComputedStyle(el); return Boolean(r && r.width > 0 && r.height > 0 && s?.display !== 'none' && s?.visibility !== 'hidden'); }};
  const text = el => String(el?.getAttribute?.('aria-label') || el?.innerText || el?.textContent || '').replace(/\s+/g, ' ').trim();
  const optionMatches = (value, expected) => value === expected || value.startsWith(`${{expected}} `);
  const leaves = entries => entries.filter(entry => !entries.some(other => other !== entry && entry.contains(other)));
  const point = el => {{ const r = el.getBoundingClientRect(); return {{ x: r.left + r.width / 2, y: r.top + r.height / 2 }}; }};
  const clickable = el => el?.closest?.('button,[role="option"],[role="menuitem"],[role="menuitemradio"]') || el;
  const inOpenMenu = el => {{ const menu = el?.closest?.('[role="menu"],[role="listbox"],[role="dialog"]'); return Boolean(menu && visible(menu)); }};
  const running = el => Boolean(el?.getAnimations?.({{ subtree: true }}).some(animation => animation.playState === 'running'));
  const root = document.querySelector('#input-engine-container,[data-testid="input-engine-container"],[data-automation-id="input-engine-container"],[data-testid="chat_input"]');
  if (!root) return {{ ready: false, supported: false, reason: 'input_engine_missing' }};
  const outerTrigger = root.querySelector('button[aria-haspopup="menu"]');
  if (!outerTrigger) return {{ ready: false, supported: false, reason: 'image_settings_unsupported' }};
  const optionElements = leaves([...document.querySelectorAll('[role="menuitem"],[role="option"],[role="menuitemradio"],button,[role="button"],div,span')].filter(visible));
  // 当前控件内部也有一个 Seedream 文本，只有真正位于打开菜单中的文本才是选项。
  const modelOptions = optionElements.filter(el => inOpenMenu(el) && optionMatches(text(el), expectedModel));
  if (modelOptions.length === 1) return {{ ready: false, supported: true, step: 'model_option', point: point(clickable(modelOptions[0])) }};
  const ratioOptions = optionElements.filter(el => inOpenMenu(el) && optionMatches(text(el), expectedRatio));
  if (ratioOptions.length === 1) return {{ ready: false, supported: true, step: 'ratio_option', point: point(clickable(ratioOptions[0])) }};
  const modelControls = [...document.querySelectorAll('[data-input-engine-actionbar-control-key="model"]')].filter(visible);
  const ratioControls = leaves([...document.querySelectorAll('[data-creation-params-panel-id]')].filter(visible).filter(el => /比例\s*(自动|\d+:\d+)/.test(text(el))));
  // 新版桌面端移除了上面两个内部属性，但保留了稳定的可见按钮文案。
  const visibleControls = [...root.querySelectorAll('button,[role="button"]')].filter(visible);
  const modelTextControls = visibleControls.filter(el => /^模型\s*Seedream\s*(?:5\.0\s*(?:Flash|Pro)|4\.[05])$/i.test(text(el)));
  const ratioTextControls = visibleControls.filter(el => /^比例\s*(自动|\d+:\d+)$/.test(text(el)));
  const resolvedModelControls = modelControls.length ? modelControls : modelTextControls;
  const resolvedRatioControls = ratioControls.length ? ratioControls : ratioTextControls;
  if (resolvedModelControls.length !== 1 || resolvedRatioControls.length !== 1) {{
    const trigger = outerTrigger.querySelector('button') || outerTrigger;
    if (outerTrigger.getAttribute('data-state') === 'open' || outerTrigger.getAttribute('aria-expanded') === 'true') return {{ ready: false, supported: true, waiting: true, reason: 'settings_mounting', point: point(trigger) }};
    return {{ ready: false, supported: true, step: 'settings', point: point(trigger) }};
  }}
  const menu = resolvedModelControls[0].closest('[role="menu"]') || resolvedRatioControls[0].closest('[role="menu"]');
  if (running(menu)) return {{ ready: false, supported: true, waiting: true, reason: 'settings_animating' }};
  const currentModel = text(resolvedModelControls[0]).match(/Seedream\s*(?:5\.0\s*(?:Flash|Pro)|4\.[05])/i)?.[0]?.replace(/\s+/g, ' ') || '';
  if (currentModel !== expectedModel) {{
    return {{ ready: false, supported: true, step: 'model', point: point(resolvedModelControls[0]) }};
  }}
  const currentRatio = text(resolvedRatioControls[0]).match(/比例\s*(自动|\d+:\d+)/)?.[1] || '';
  if (currentRatio !== expectedRatio) {{
    return {{ ready: false, supported: true, step: 'ratio', point: point(resolvedRatioControls[0]) }};
  }}
  return {{ ready: true, supported: true }};
}})()"#,
    )
}

/// 读取普通聊天输入区的生成入口状态。新版豆包把入口收进无文字的加号菜单时，
/// 这个探针只返回坐标与页面结构，实际点击由 CDP 的可信鼠标事件完成。
#[cfg(test)]
fn generation_launcher_probe_expression() -> &'static str {
    r#"(() => {
  const visible = el => {
    if (!el) return false;
    const r = el.getBoundingClientRect();
    const s = getComputedStyle(el);
    return r.width > 0 && r.height > 0 && s.visibility !== 'hidden' && s.display !== 'none';
  };
  const label = el => (el.getAttribute('aria-label') || el.innerText || el.textContent || '').replace(/\s+/g, ' ').trim();
  const root = document.querySelector('#input-engine-container,[data-testid="input-engine-container"],[data-automation-id="input-engine-container"]');
  if (!root) return { root: false, launcher: null, buttons: [] };
  const editor = root.querySelector('[contenteditable="true"]');
  const editorRect = editor?.getBoundingClientRect();
  const buttons = [...root.querySelectorAll('button')].filter(visible).map((button, index) => {
    const r = button.getBoundingClientRect();
    return {
      index,
      label: label(button),
      testId: button.getAttribute('data-testid'),
      hasSvg: Boolean(button.querySelector('svg')),
      parentTestId: button.parentElement?.getAttribute('data-testid'),
      parentClass: String(button.parentElement?.className || '').slice(0, 160),
      distanceToEditor: editorRect ? Math.hypot((r.x + r.width / 2) - (editorRect.x + editorRect.width / 2), (r.y + r.height / 2) - (editorRect.y + editorRect.height / 2)) : null,
      rect: { x: r.x + r.width / 2, y: r.y + r.height / 2, width: r.width, height: r.height }
    };
  });
  const candidates = buttons.filter(button => button.hasSvg && !button.label && editorRect
    && button.rect.x >= editorRect.x - 80 && button.rect.x <= editorRect.x + 90
    && button.rect.y >= editorRect.y - 80 && button.rect.y <= editorRect.y + editorRect.height + 110);
  return {
    root: true,
    runtimeMode: root.getAttribute('data-input-engine-runtime-mode'),
    editor: editorRect ? { x: editorRect.x, y: editorRect.y, width: editorRect.width, height: editorRect.height } : null,
    launcher: candidates.length === 1 ? candidates[0] : null,
    launcherCount: candidates.length,
    buttons
  };
})()"#
}

/// 可信点击加号后采集页面局部弹层。仅用于人工干跑诊断，不点击任何菜单项目。
#[cfg(test)]
fn generation_launcher_menu_snapshot_expression() -> &'static str {
    r#"(() => {
  const visible = el => {
    if (!el) return false;
    const r = el.getBoundingClientRect();
    const s = getComputedStyle(el);
    return r.width > 0 && r.height > 0 && s.visibility !== 'hidden' && s.display !== 'none';
  };
  const label = el => (el.getAttribute('aria-label') || el.innerText || el.textContent || '').replace(/\s+/g, ' ').trim();
  const describe = el => ({
    tag: el.tagName.toLowerCase(),
    role: el.getAttribute('role'),
    testId: el.getAttribute('data-testid'),
    automationId: el.getAttribute('data-automation-id'),
    label: label(el).slice(0, 120),
    className: String(el.className || '').slice(0, 180),
    parent: el.parentElement ? `${el.parentElement.tagName.toLowerCase()}:${el.parentElement.getAttribute('role') || el.parentElement.getAttribute('data-testid') || ''}` : '',
    rect: (() => { const r = el.getBoundingClientRect(); return { x: r.x, y: r.y, width: r.width, height: r.height }; })()
  });
  const selectors = ['[role=menu]', '[role=listbox]', '[role=dialog]', '[role=menuitem]', '[role=option]', '[data-radix-popper-content-wrapper]'];
  return Object.fromEntries(selectors.map(selector => [selector, [...document.querySelectorAll(selector)].filter(visible).map(describe)]));
})()"#
}

/// 输入区的最小可见控件快照，用于核对“工具栏报告已打开”后是否真的仍处于生成模式。
#[cfg(test)]
fn generation_input_controls_snapshot_expression() -> &'static str {
    r#"(() => {
  const visible = el => {
    if (!el) return false;
    const r = el.getBoundingClientRect();
    const s = getComputedStyle(el);
    return r.width > 0 && r.height > 0 && s.visibility !== 'hidden' && s.display !== 'none';
  };
  const label = el => (el.getAttribute('aria-label') || el.innerText || el.textContent || '').replace(/\s+/g, ' ').trim();
  const root = document.querySelector('#input-engine-container,[data-testid="input-engine-container"],[data-automation-id="input-engine-container"]');
  if (!root) return { root: false, controls: [] };
  const candidates = [...root.querySelectorAll('button,[role="button"],div,span')]
    .filter(visible)
    .map(el => ({ el, label: label(el) }))
    .filter(entry => entry.label)
    .filter(entry => ![...root.querySelectorAll('button,[role="button"],div,span')]
      .some(other => other !== entry.el && visible(other) && entry.el.contains(other) && label(other) === entry.label));
  return {
    root: true,
    runtimeMode: root.getAttribute('data-input-engine-runtime-mode'),
    controls: candidates.slice(0, 80).map(entry => ({
      tag: entry.el.tagName.toLowerCase(),
      label: entry.label.slice(0, 140),
      testId: entry.el.getAttribute('data-testid'),
      role: entry.el.getAttribute('role')
    }))
  };
})()"#
}

/// analysis/05 "生成请求识别":https、host 以 `doubao.com`/`snssdk.com` 结尾、POST、
/// path 含 `completion|generate|/im/message/send` 且不含 `history|list|recommend|config`。
pub fn is_generation_request(url: &str, method: &str) -> bool {
    if !method.eq_ignore_ascii_case("POST") {
        return false;
    }
    let Ok(parsed) = url::Url::parse(url) else {
        return false;
    };
    if parsed.scheme() != "https" {
        return false;
    }
    let Some(host) = parsed.host_str() else {
        return false;
    };
    let host_ok = ["doubao.com", "snssdk.com"]
        .iter()
        .any(|d| host == *d || host.ends_with(&format!(".{d}")));
    if !host_ok {
        return false;
    }
    let path = parsed.path();
    let positive = path.contains("completion")
        || path.contains("generate")
        || path.contains("/im/message/send");
    let negative = ["history", "list", "recommend", "config"]
        .iter()
        .any(|w| path.contains(w));
    positive && !negative
}

/// 解析 `submissionIdentity` 的返回:`null` / 非对象 / 既无 localMessageId 也无 messageId → None。
pub fn parse_identity(v: &Value) -> Option<SubmissionIdentity> {
    let obj = v.as_object()?;
    let pick = |k: &str| -> Option<String> {
        match obj.get(k)? {
            Value::String(s) if !s.is_empty() && s != "0" => Some(s.clone()),
            Value::Number(n) => Some(n.to_string()),
            _ => None,
        }
    };
    let local_message_id = pick("localMessageId");
    let message_id = pick("messageId");
    let conversation_id = pick("conversationId");
    if local_message_id.is_none() && message_id.is_none() {
        return None;
    }
    Some(SubmissionIdentity {
        local_message_id,
        message_id,
        conversation_id,
    })
}

// ------------------------------------------------------------------
// 流程
// ------------------------------------------------------------------

/// 阶段回调:`(stage, message)`。
pub type StageFn<'a> = &'a (dyn Fn(TaskStage, Option<String>) + Send + Sync);

pub(crate) enum StepError {
    Js(String),
    Timeout,
    Cdp(CdpError),
    Cancelled,
}

/// `Runtime.evaluate` 带自定义超时,并可被 `cancel` 打断。
pub(crate) async fn evaluate_step(
    session: &CdpSession,
    expression: &str,
    timeout: Duration,
    cancel: &CancellationToken,
) -> Result<Value, StepError> {
    if cancel.is_cancelled() {
        return Err(StepError::Cancelled);
    }
    let eval = session.send_timeout(
        "Runtime.evaluate",
        json!({
            "expression": expression,
            "returnByValue": true,
            "awaitPromise": true,
        }),
        timeout,
    );
    let result = tokio::select! {
        _ = cancel.cancelled() => return Err(StepError::Cancelled),
        r = eval => r,
    };
    match result {
        Ok(result) => {
            if let Some(details) = result.get("exceptionDetails") {
                return Err(StepError::Js(extract_js_exception(details)));
            }
            Ok(result
                .get("result")
                .and_then(|r| r.get("value"))
                .cloned()
                .unwrap_or(Value::Null))
        }
        Err(CdpError::Timeout(_)) => Err(StepError::Timeout),
        Err(e) => Err(StepError::Cdp(e)),
    }
}

/// 准备阶段(exit-skill / bootstrap / prepare / fill)的错误映射。
pub(crate) fn map_preparing(step: &str, e: StepError) -> RelayError {
    match e {
        StepError::Js(msg) => preparing_failed(step, Some(&msg)),
        StepError::Timeout => preparing_timeout(step),
        StepError::Cdp(e) => e.into(),
        StepError::Cancelled => cancelled(step),
    }
}

/// 提交阶段(checkSend / send)的错误映射。
fn map_submit(step: &str, e: StepError) -> RelayError {
    match e {
        StepError::Js(msg) => submit_failed(step, Some(&msg)),
        StepError::Timeout => submit_failed("timeout", None).with_details(json!({
            "reason": "timeout",
            "step": step,
            "js_exception": Value::Null,
        })),
        StepError::Cdp(e) => e.into(),
        StepError::Cancelled => cancelled(step),
    }
}

#[derive(Debug, Default, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ExitSkillResult {
    #[serde(default)]
    found: bool,
    #[serde(default)]
    exited: bool,
    #[serde(default)]
    skill_bar: Option<bool>,
    #[serde(default)]
    conversation_ready: bool,
    #[serde(default)]
    generation_ready: bool,
    #[serde(default)]
    waited_ms: u64,
}

#[derive(Debug, Default, Deserialize)]
struct NativeClickPoint {
    x: f64,
    y: f64,
}

#[derive(Debug, Default, Deserialize)]
struct NativeGenerationModeProbe {
    #[serde(default)]
    ready: bool,
    #[serde(default)]
    supported: bool,
    #[serde(default)]
    waiting: bool,
    #[serde(default)]
    step: Option<String>,
    #[serde(default)]
    point: Option<NativeClickPoint>,
    #[serde(default, rename = "dismissPoint")]
    dismiss_point: Option<NativeClickPoint>,
    #[serde(default)]
    reason: Option<String>,
}

#[derive(Debug, Default, Deserialize)]
#[serde(rename_all = "camelCase")]
struct PrepareResult {
    #[serde(default)]
    needs_duration_input: bool,
    #[serde(default)]
    mode: Option<String>,
    #[serde(default)]
    model: Option<String>,
    #[serde(default)]
    ratio: Option<String>,
    #[serde(default)]
    duration: Option<Value>,
}

/// 退出残留技能(真机:技能态下"新对话"不退出技能,跨技能 `prepare` 12s 后失败)。
async fn exit_residual_skill(
    session: &CdpSession,
    timing: &SubmitTiming,
    cancel: &CancellationToken,
) -> RelayResult<()> {
    let v = evaluate_step(
        session,
        &exit_skill_expression(timing.skill_exit_timeout),
        timing.step_timeout,
        cancel,
    )
    .await
    .map_err(|e| map_preparing("exit_skill", e))?;
    let r: ExitSkillResult = serde_json::from_value(v).unwrap_or_default();
    if r.found {
        tracing::debug!(
            exited = r.exited,
            skill_bar = r.skill_bar.unwrap_or(false),
            waited_ms = r.waited_ms,
            "已退出豆包残留技能"
        );
        if r.skill_bar != Some(true) && !r.conversation_ready && !r.generation_ready {
            return Err(
                preparing_failed("skill_exit_failed", None).with_details(json!({
                    "reason": "skill_exit_failed",
                    "failed_selectors": [SKILL_BAR_SELECTOR],
                    "exited": r.exited,
                    "conversation_ready": r.conversation_ready,
                    "generation_ready": r.generation_ready,
                    "waited_ms": r.waited_ms,
                })),
            );
        }
    }
    Ok(())
}

async fn dispatch_trusted_click(
    session: &CdpSession,
    point: &NativeClickPoint,
    timeout: Duration,
    cancel: &CancellationToken,
) -> Result<(), StepError> {
    for (event_type, buttons) in [("mouseMoved", 0), ("mousePressed", 1), ("mouseReleased", 0)] {
        let mut params = json!({ "type": event_type, "x": point.x, "y": point.y });
        if event_type != "mouseMoved" {
            params["button"] = json!("left");
            params["buttons"] = json!(buttons);
            params["clickCount"] = json!(1);
        }
        let send = session.send_timeout("Input.dispatchMouseEvent", params, timeout);
        tokio::select! {
            _ = cancel.cancelled() => return Err(StepError::Cancelled),
            result = send => result.map_err(StepError::Cdp)?,
        };
    }
    Ok(())
}

async fn dispatch_escape(
    session: &CdpSession,
    timeout: Duration,
    cancel: &CancellationToken,
) -> Result<(), StepError> {
    for event_type in ["rawKeyDown", "keyUp"] {
        let send = session.send_timeout(
            "Input.dispatchKeyEvent",
            json!({"type": event_type, "key": "Escape", "code": "Escape", "windowsVirtualKeyCode": 27, "nativeVirtualKeyCode": 27}),
            timeout,
        );
        tokio::select! {
            _ = cancel.cancelled() => return Err(StepError::Cancelled),
            result = send => { result.map_err(StepError::Cdp)?; }
        }
    }
    Ok(())
}

/// 新版豆包必须收到浏览器级可信点击，页面内合成的 MouseEvent 只会留下隐藏控件。
/// 旧版页面返回 `supported=false` 时保持原流程，由注入脚本继续处理。
async fn select_native_generation_mode(
    session: &CdpSession,
    task_type: TaskType,
    timing: &SubmitTiming,
    cancel: &CancellationToken,
) -> RelayResult<bool> {
    let deadline = Instant::now() + timing.prepare_timeout;
    let mut entered_target = false;
    let mut fresh_chat_opened = false;
    let mut native_selection_started = false;
    let mut menu_close_attempts = 0u8;
    loop {
        let expression =
            native_generation_mode_probe_expression(task_type, !entered_target, !fresh_chat_opened);
        let value = evaluate_step(session, &expression, timing.step_timeout, cancel)
            .await
            .map_err(|error| map_preparing("generation_mode", error))?;
        let probe: NativeGenerationModeProbe = serde_json::from_value(value).unwrap_or_default();
        tracing::info!(
            ready = probe.ready,
            supported = probe.supported,
            step = probe.step.as_deref().unwrap_or(""),
            reason = probe.reason.as_deref().unwrap_or(""),
            "豆包原生生成模式探针"
        );
        if probe.ready {
            return Ok(true);
        }
        if !probe.supported {
            // 点击模式按钮后，豆包会短暂卸载整个输入区再重新挂载。此时不能把
            // `input_engine_missing` 当成旧版页面，否则会提前进入 prepare。
            let input_engine_remounting = probe.reason.as_deref() == Some("input_engine_missing");
            if !native_selection_started && !input_engine_remounting {
                return Ok(false);
            }
            if Instant::now() >= deadline {
                return Err(preparing_timeout("generation_mode").with_details(json!({
                    "reason": probe.reason.unwrap_or_else(|| "generation_mode_remount_timeout".into()),
                })));
            }
            tokio::select! {
                _ = cancel.cancelled() => return Err(cancelled("generation_mode")),
                _ = tokio::time::sleep(Duration::from_millis(250)) => {}
            }
            continue;
        }
        if matches!(
            probe.reason.as_deref(),
            Some("conversation_menu_already_selected" | "conversation_menu_open")
        ) {
            // 当前已经是“对话”模式，或菜单只剩外壳没有正常挂载；先点模式按钮
            // 收起。按钮连续没有收起时，再点输入区附近的安全位置，避免卡在菜单层。
            menu_close_attempts = menu_close_attempts.saturating_add(1);
            if let Some(point) = probe.point.as_ref() {
                dispatch_trusted_click(session, point, timing.step_timeout, cancel)
                    .await
                    .map_err(|error| map_preparing("generation_mode", error))?;
            }
            if menu_close_attempts >= 2 {
                if let Some(point) = probe.dismiss_point.as_ref() {
                    dispatch_trusted_click(session, point, timing.step_timeout, cancel)
                        .await
                        .map_err(|error| map_preparing("generation_mode", error))?;
                } else {
                    dispatch_escape(session, timing.step_timeout, cancel)
                        .await
                        .map_err(|error| map_preparing("generation_mode", error))?;
                }
            } else if probe.point.is_none() {
                dispatch_escape(session, timing.step_timeout, cancel)
                    .await
                    .map_err(|error| map_preparing("generation_mode", error))?;
            }
            tokio::select! {
                _ = cancel.cancelled() => return Err(cancelled("generation_mode")),
                _ = tokio::time::sleep(Duration::from_millis(250)) => {}
            }
            continue;
        }
        menu_close_attempts = 0;
        if probe.waiting {
            if Instant::now() >= deadline {
                return Err(preparing_timeout("generation_mode").with_details(json!({
                    "reason": probe.reason.unwrap_or_else(|| "generation_mode_wait_timeout".into()),
                })));
            }
            tokio::select! {
                _ = cancel.cancelled() => return Err(cancelled("generation_mode")),
                _ = tokio::time::sleep(Duration::from_millis(50)) => {}
            }
            continue;
        }
        let step = probe.step;
        let Some(point) = probe.point else {
            return Err(preparing_failed("generation_mode", None).with_details(json!({
                "reason": probe.reason.unwrap_or_else(|| "generation_mode_point_missing".into()),
                "step": step,
            })));
        };
        dispatch_trusted_click(session, &point, timing.step_timeout, cancel)
            .await
            .map_err(|error| map_preparing("generation_mode", error))?;
        native_selection_started = true;
        if step.as_deref() == Some("fresh_chat") {
            fresh_chat_opened = true;
        }
        if step.as_deref() == Some("target") {
            entered_target = true;
        }
        if Instant::now() >= deadline {
            return Err(preparing_timeout("generation_mode").with_details(json!({
                "reason": "generation_mode_timeout",
                "step": step,
            })));
        }
        tokio::select! {
            _ = cancel.cancelled() => return Err(cancelled("generation_mode")),
            _ = tokio::time::sleep(Duration::from_millis(250)) => {}
        }
    }
}

async fn prepare_native_image_settings(
    session: &CdpSession,
    model_name: &str,
    ratio: &str,
    wait_for_toolbar_mount: bool,
    timing: &SubmitTiming,
    cancel: &CancellationToken,
) -> RelayResult<()> {
    let deadline = Instant::now() + timing.prepare_timeout;
    let mut mounting_waits = 0u8;
    let mut unsupported_since = None;
    loop {
        let expression = native_image_settings_probe_expression(model_name, ratio);
        let value = evaluate_step(session, &expression, timing.step_timeout, cancel)
            .await
            .map_err(|error| map_preparing("image_settings", error))?;
        let probe: NativeGenerationModeProbe = serde_json::from_value(value).unwrap_or_default();
        tracing::debug!(
            ready = probe.ready,
            supported = probe.supported,
            waiting = probe.waiting,
            step = probe.step.as_deref().unwrap_or(""),
            reason = probe.reason.as_deref().unwrap_or(""),
            x = probe.point.as_ref().map(|point| point.x),
            y = probe.point.as_ref().map(|point| point.y),
            "豆包图片设置探针"
        );
        if probe.ready {
            return Ok(());
        }
        if !probe.supported {
            let toolbar_remounting = wait_for_toolbar_mount
                && matches!(
                    probe.reason.as_deref(),
                    Some("input_engine_missing" | "image_settings_unsupported")
                );
            if toolbar_remounting {
                let since = unsupported_since.get_or_insert_with(Instant::now);
                if since.elapsed() < Duration::from_secs(5) && Instant::now() < deadline {
                    tokio::select! {
                        _ = cancel.cancelled() => return Err(cancelled("image_settings")),
                        _ = tokio::time::sleep(Duration::from_millis(250)) => {}
                    }
                    continue;
                }
            }
            return Ok(());
        }
        unsupported_since = None;
        if Instant::now() >= deadline {
            return Err(preparing_timeout("image_settings").with_details(json!({
                "reason": probe.reason.unwrap_or_else(|| "image_settings_timeout".into()),
                "step": probe.step,
            })));
        }
        if probe.waiting {
            if probe.reason.as_deref() == Some("settings_mounting") {
                mounting_waits = mounting_waits.saturating_add(1);
                if mounting_waits >= 10 {
                    let Some(point) = probe.point.as_ref() else {
                        return Err(preparing_failed("image_settings", None));
                    };
                    dispatch_trusted_click(session, point, timing.step_timeout, cancel)
                        .await
                        .map_err(|error| map_preparing("image_settings", error))?;
                    mounting_waits = 0;
                    tokio::time::sleep(Duration::from_millis(250)).await;
                    continue;
                }
            } else {
                mounting_waits = 0;
            }
            tokio::select! {
                _ = cancel.cancelled() => return Err(cancelled("image_settings")),
                _ = tokio::time::sleep(Duration::from_millis(50)) => {}
            }
            continue;
        }
        mounting_waits = 0;
        let step = probe.step;
        let Some(point) = probe.point else {
            return Err(
                preparing_failed("image_settings", None).with_details(json!({
                    "reason": probe.reason.unwrap_or_else(|| "image_settings_point_missing".into()),
                    "step": step,
                })),
            );
        };
        if step.as_deref() == Some("settings") {
            dispatch_escape(session, timing.step_timeout, cancel)
                .await
                .map_err(|error| map_preparing("image_settings", error))?;
            tokio::time::sleep(Duration::from_millis(100)).await;
        }
        // 先清掉上一次没收起的菜单。豆包新版允许两个浮层短暂叠在一起，若直接点
        // 下一个控件，探针会在“模型选项/比例控件”之间循环，永远无法完成准备。
        if matches!(step.as_deref(), Some("model" | "ratio")) {
            dispatch_escape(session, timing.step_timeout, cancel)
                .await
                .map_err(|error| map_preparing("image_settings", error))?;
            tokio::time::sleep(Duration::from_millis(100)).await;
        }
        // 模型和比例控件都是菜单按钮，豆包不会因悬停打开菜单；必须使用浏览器级
        // 可信点击，否则探针会一直看不到目标选项，最后报 image_settings_timeout。
        dispatch_trusted_click(session, &point, timing.step_timeout, cancel)
            .await
            .map_err(|error| map_preparing("image_settings", error))?;
        if matches!(step.as_deref(), Some("model_option" | "ratio_option")) {
            tokio::time::sleep(Duration::from_millis(150)).await;
            // 选择项已经交给豆包处理；若旧浮层仍在，主动收起，下一轮只读取当前控件。
            dispatch_escape(session, timing.step_timeout, cancel)
                .await
                .map_err(|error| map_preparing("image_settings", error))?;
        }
        tokio::select! {
            _ = cancel.cancelled() => return Err(cancelled("image_settings")),
            _ = tokio::time::sleep(Duration::from_millis(250)) => {}
        }
    }
}

/// 准备阶段:退出技能 → 注入 → `prepare` →(视频 `needsDurationInput` → `duration::set_duration`)
/// →(参考图 → `upload::upload_references`)→ `fill` → `checkSend`。
/// 成功后输入框里已有 prompt、发送按钮可点,返回之后 `send` 要用的请求 JSON(含 `uploadedCount`)。
/// `live_` 干跑测试也走这里(之后清空输入框)。
async fn prepare_and_fill(
    session: &CdpSession,
    r: &GenerateRequest,
    plan: &SubmissionPlan,
    reference_paths: &[PathBuf],
    timing: &SubmitTiming,
    cancel: &CancellationToken,
    on_stage: StageFn<'_>,
) -> RelayResult<Value> {
    on_stage(TaskStage::Preparing, Some("准备豆包生成页面".into()));
    let image_page_ready = r.task_type == TaskType::Image
        && session
            .evaluate_as::<bool>(image_page_ready_expression())
            .await
            .unwrap_or(false);
    if !image_page_ready {
        exit_residual_skill(session, timing, cancel).await?;
    }
    let native_mode = if image_page_ready {
        tracing::info!("当前豆包图片生成页已就绪,跳过旧版技能切换流程");
        true
    } else {
        let mode_message = match r.task_type {
            TaskType::Image => "正在检查豆包模式,自动退出对话并打开图像生成".to_string(),
            TaskType::Video => "正在检查豆包模式,自动退出对话并打开视频生成".to_string(),
        };
        on_stage(TaskStage::Preparing, Some(mode_message));
        select_native_generation_mode(session, r.task_type, timing, cancel).await?
    };
    if !image_page_ready {
        let mode_message = match r.task_type {
            TaskType::Image => "图像生成入口已确认,正在准备模型和比例".to_string(),
            TaskType::Video => "视频生成入口已确认,正在准备模型和比例".to_string(),
        };
        on_stage(TaskStage::Preparing, Some(mode_message));
    }
    if r.task_type == TaskType::Image {
        prepare_native_image_settings(
            session,
            &r.model_name,
            &r.ratio,
            native_mode,
            timing,
            cancel,
        )
        .await?;
    }

    evaluate_step(
        session,
        &generation_bootstrap(&plan.tracking_key),
        timing.step_timeout,
        cancel,
    )
    .await
    .map_err(|e| map_preparing("bootstrap", e))?;

    // 残留参考图会让 prepare 失败(Task 3.2 真机),先尽力删掉当前可见的卡片。
    let leftover = upload::remove_attachments(session).await;
    if leftover > 0 {
        tracing::warn!(
            count = leftover,
            "豆包输入区有残留参考图,已在 prepare 前删除"
        );
    }

    let mut js_request = build_js_request(r, plan, reference_paths, 0, false);
    let started = Instant::now();
    let prepare_expr = step_expression(&plan.tracking_key, "prepare", &js_request);
    let prepared = match evaluate_step(session, &prepare_expr, timing.prepare_timeout, cancel).await
    {
        // 参考图属于技能草稿:进了技能才看得见、也只有在技能里才删得掉。prepare 抛异常时
        // 已在技能里 → 删卡片(图片与视频)+ 清输入框后重试一次。
        Err(StepError::Js(msg)) if is_leftover_exception(&msg) => {
            let removed = upload::remove_attachments(session).await;
            clear_input(session).await;
            tracing::warn!(removed, js_exception = %msg, "技能草稿里有残留内容,清理后重试 prepare");
            evaluate_step(session, &prepare_expr, timing.prepare_timeout, cancel).await
        }
        Err(StepError::Js(msg))
            if r.task_type == TaskType::Image && is_transient_image_toolbar_missing(&msg) =>
        {
            tracing::warn!(js_exception = %msg, "豆包图片工具栏在页面切换时丢失,重新打开后重试 prepare");
            let _ = select_native_generation_mode(session, TaskType::Image, timing, cancel).await?;
            evaluate_step(session, &prepare_expr, timing.prepare_timeout, cancel).await
        }
        Err(StepError::Js(msg)) if is_transient_control_ambiguity(&msg) => {
            let mut last = msg;
            let mut result = Err(StepError::Js(last.clone()));
            for retry in 1..=3 {
                tracing::warn!(retry, js_exception = %last, "豆包页面切换时控件短暂重叠,等待后重试 prepare");
                result = tokio::select! {
                    _ = cancel.cancelled() => Err(StepError::Cancelled),
                    _ = tokio::time::sleep(Duration::from_millis(500)) => {
                        evaluate_step(session, &prepare_expr, timing.prepare_timeout, cancel).await
                    }
                };
                match &result {
                    Err(StepError::Js(next_msg)) if is_transient_control_ambiguity(next_msg) && retry < 3 => {
                        last = next_msg.clone();
                    }
                    _ => break,
                }
            }
            result
        }
        other => other,
    }
    .map_err(|e| map_preparing("prepare", e))?;
    let prepared: PrepareResult = serde_json::from_value(prepared).unwrap_or_default();
    tracing::info!(
        elapsed_ms = started.elapsed().as_millis() as u64,
        mode = prepared.mode.as_deref().unwrap_or(""),
        model = prepared.model.as_deref().unwrap_or(""),
        ratio = prepared.ratio.as_deref().unwrap_or(""),
        needs_duration_input = prepared.needs_duration_input,
        "prepare 完成"
    );

    if r.task_type == TaskType::Video {
        if prepared.needs_duration_input {
            // 比例已选好、弹层留在打开状态且滑块 focus(Task 2.1)。
            let Some(target) = r.duration else {
                return Err(
                    preparing_failed("duration_missing", None).with_details(json!({
                        "reason": "duration_missing",
                        "js_exception": Value::Null,
                    })),
                );
            };
            on_stage(
                TaskStage::Preparing,
                Some(format!("设置视频时长 {target} 秒")),
            );
            duration::set_duration_with(
                session,
                &plan.tracking_key,
                &js_request,
                target,
                cancel,
                &timing.duration,
            )
            .await?;
        } else if let (Some(want), Some(got)) = (r.duration, prepared.duration.as_ref()) {
            if got.as_u64() != Some(u64::from(want)) {
                return Err(
                    preparing_failed("duration_mismatch", None).with_details(json!({
                        "reason": "duration_mismatch",
                        "js_exception": Value::Null,
                        "expected": want,
                        "actual": got,
                    })),
                );
            }
        }
    }
    on_stage(TaskStage::Preparing, Some("模型与参数已就绪".into()));

    let after = after_prepare(
        session,
        plan,
        reference_paths,
        &mut js_request,
        timing,
        cancel,
        on_stage,
    )
    .await;
    if let Err(e) = &after {
        if !reference_paths.is_empty() {
            tracing::warn!(code = e.code, "参考图上传后准备失败,删除已上传的参考图");
            upload::cleanup_attachments(session).await;
        }
    }
    after.map(|()| js_request)
}

/// `prepare` 之后:参考图上传 → `fill` → `checkSend`。失败时由调用方清理参考图。
async fn after_prepare(
    session: &CdpSession,
    plan: &SubmissionPlan,
    reference_paths: &[PathBuf],
    js_request: &mut Value,
    timing: &SubmitTiming,
    cancel: &CancellationToken,
    on_stage: StageFn<'_>,
) -> RelayResult<()> {
    if !reference_paths.is_empty() {
        let total = reference_paths.len();
        on_stage(TaskStage::Preparing, Some(format!("上传参考图 0/{total}")));
        let progress = |done: usize, total: usize| {
            on_stage(
                TaskStage::Preparing,
                Some(format!("上传参考图 {done}/{total}")),
            )
        };
        upload::upload_references_with(
            session,
            &plan.tracking_key,
            js_request,
            reference_paths,
            cancel,
            &progress,
            &timing.upload,
        )
        .await?;
    }

    evaluate_step(
        session,
        &step_expression(&plan.tracking_key, "fill", js_request),
        timing.step_timeout,
        cancel,
    )
    .await
    .map_err(|e| map_preparing("fill", e))?;
    on_stage(TaskStage::Preparing, Some("提示词已填写".into()));

    if let Err(e) = evaluate_step(
        session,
        &step_expression(&plan.tracking_key, "checkSend", js_request),
        timing.step_timeout,
        cancel,
    )
    .await
    {
        clear_input(session).await;
        return Err(map_submit("checkSend", e));
    }
    Ok(())
}

/// 失败后尽力清空输入框(下一次 `fill` 要求输入框为空);忽略错误。
/// 最多等 [`upload::CLEANUP_TIMEOUT`],以便放进 runner 取消后的宽限期。
async fn clear_input(session: &CdpSession) {
    match session
        .send_timeout(
            "Runtime.evaluate",
            json!({ "expression": clear_input_expression(), "returnByValue": true }),
            upload::CLEANUP_TIMEOUT,
        )
        .await
    {
        Ok(v) if v.get("exceptionDetails").is_some() => {
            tracing::warn!(error = %extract_js_exception(&v["exceptionDetails"]), "清空豆包输入框失败")
        }
        Ok(v) => tracing::debug!(result = %v["result"]["value"], "已清空豆包输入框"),
        Err(e) => tracing::warn!(error = %e, "清空豆包输入框失败"),
    }
}

/// `Network.enable {maxPostDataSize}`:Protocol 错误只 warn(与 connection.rs 同),断线 → Err。
async fn enable_network(session: &CdpSession) -> RelayResult<()> {
    match session
        .send(
            "Network.enable",
            json!({ "maxPostDataSize": MAX_POST_DATA_SIZE }),
        )
        .await
    {
        Ok(_) => Ok(()),
        Err(CdpError::Protocol { code, message }) => {
            tracing::warn!(code, message = %message, target_id = %session.target_id, "Network.enable 被拒绝,忽略");
            Ok(())
        }
        Err(e) => Err(e.into()),
    }
}

/// 请求已经发送后，把豆包窗口收回后台。提交完成后的生成和结果读取都能通过
/// CDP 继续运行。
///
/// 后台化只是体验优化，不能让已发送的任务因为窗口管理失败而被标记为失败。
async fn minimize_window_after_submission(session: &CdpSession) {
    let window = match session
        .send(
            "Browser.getWindowForTarget",
            json!({ "targetId": session.target_id }),
        )
        .await
    {
        Ok(window) => window,
        Err(error) => {
            tracing::warn!(error = %error, target_id = %session.target_id, "无法获取豆包窗口，保持当前窗口状态");
            return;
        }
    };
    let Some(window_id) = window.get("windowId").and_then(Value::as_i64) else {
        tracing::warn!(target_id = %session.target_id, "豆包窗口编号缺失，保持当前窗口状态");
        return;
    };
    if let Err(error) = session
        .send(
            "Browser.setWindowBounds",
            json!({ "windowId": window_id, "bounds": { "windowState": "minimized" } }),
        )
        .await
    {
        tracing::warn!(error = %error, window_id, "无法将豆包窗口转入后台");
        return;
    }
    tracing::info!(window_id, "豆包任务已提交，窗口已转入后台");
}

/// 从事件接收器取下一条事件;`None` 接收器永远挂起,Lagged 跳过,Closed 挂起(由外层 deadline 收尾)。
async fn next_event(rx: &mut Option<broadcast::Receiver<CdpEvent>>) -> CdpEvent {
    let Some(rx) = rx.as_mut() else {
        return std::future::pending().await;
    };
    loop {
        match rx.recv().await {
            Ok(ev) => return ev,
            Err(broadcast::error::RecvError::Lagged(n)) => {
                tracing::warn!(skipped = n, "CDP 事件积压,已跳过");
            }
            Err(broadcast::error::RecvError::Closed) => return std::future::pending().await,
        }
    }
}

/// 一条候选生成请求(来自 `Network.requestWillBeSent`)。
struct Candidate {
    request_id: String,
    url: String,
    method: String,
    post_data: Option<String>,
    has_post_data: bool,
    wall_time_ms: Option<i64>,
}

fn candidate_from_event(ev: &CdpEvent) -> Option<Candidate> {
    if ev.method != "Network.requestWillBeSent" {
        return None;
    }
    let req = ev.params.get("request")?;
    let url = req.get("url")?.as_str()?.to_string();
    let method = req
        .get("method")
        .and_then(Value::as_str)
        .unwrap_or("")
        .to_string();
    Some(Candidate {
        request_id: ev
            .params
            .get("requestId")
            .and_then(Value::as_str)
            .unwrap_or("")
            .to_string(),
        url,
        method,
        post_data: req
            .get("postData")
            .and_then(Value::as_str)
            .map(str::to_string),
        has_post_data: req
            .get("hasPostData")
            .and_then(Value::as_bool)
            .unwrap_or(false),
        wall_time_ms: ev
            .params
            .get("wallTime")
            .and_then(Value::as_f64)
            .map(|t| (t * 1000.0) as i64),
    })
}

/// 在产生事件的那条 session 上识别 identity:必要时 `Network.getRequestPostData` 补 body、
/// 注入 reader、evaluate `submissionIdentity`。
async fn identify_on(
    session: &CdpSession,
    reader_ready: &mut bool,
    cand: Candidate,
    prompt: &str,
    timing: &SubmitTiming,
    cancel: &CancellationToken,
) -> Option<SubmissionIdentity> {
    let post_data = match cand.post_data {
        Some(p) => p,
        None if cand.has_post_data && !cand.request_id.is_empty() => {
            match session
                .send(
                    "Network.getRequestPostData",
                    json!({ "requestId": cand.request_id }),
                )
                .await
            {
                Ok(v) => v
                    .get("postData")
                    .and_then(Value::as_str)
                    .unwrap_or("")
                    .to_string(),
                Err(e) => {
                    tracing::debug!(error = %e, "读取请求 body 失败");
                    return None;
                }
            }
        }
        None => return None,
    };
    if !*reader_ready {
        if let Err(e) =
            evaluate_step(session, &reader_bootstrap(), timing.step_timeout, cancel).await
        {
            if let StepError::Js(msg) = &e {
                tracing::warn!(js_exception = %msg, "注入结果读取脚本失败");
            }
            return None;
        }
        *reader_ready = true;
    }
    let expr = identity_expression(&cand.url, &cand.method, &post_data, prompt);
    match evaluate_step(session, &expr, timing.step_timeout, cancel).await {
        Ok(v) => parse_identity(&v),
        Err(StepError::Js(msg)) => {
            tracing::warn!(js_exception = %msg, "submissionIdentity 抛出异常");
            None
        }
        Err(_) => None,
    }
}

/// 提交一次生成请求(默认时间参数)。见 [`submit_with`]。
pub async fn submit(
    session: &CdpSession,
    background: Option<&CdpSession>,
    r: &GenerateRequest,
    plan: &SubmissionPlan,
    reference_paths: &[PathBuf],
    cancel: &CancellationToken,
    on_stage: StageFn<'_>,
) -> RelayResult<SubmissionReport> {
    submit_with(
        session,
        background,
        r,
        plan,
        reference_paths,
        cancel,
        on_stage,
        &SubmitTiming::default(),
    )
    .await
}

/// 流程(每步前检查 `cancel`):退出残留技能 → 注入 → `prepare`(60s)→ 视频 `needsDurationInput` /
/// 参考图(`upload::upload_references`,之后的请求带 `uploadedCount`)→ `fill` → `checkSend` → 订阅 chat(+background)的
/// `Network.requestWillBeSent`(`Network.enable {maxPostDataSize}`)→ `send` → 2s 内对每个生成
/// POST 在页面侧跑 `submissionIdentity` → 得 identity;窗口内没有 → `generation_submit_identity_missing`
/// (`details.sent=true`:发送已经发生)。
#[allow(clippy::too_many_arguments)]
pub async fn submit_with(
    session: &CdpSession,
    background: Option<&CdpSession>,
    r: &GenerateRequest,
    plan: &SubmissionPlan,
    reference_paths: &[PathBuf],
    cancel: &CancellationToken,
    on_stage: StageFn<'_>,
    timing: &SubmitTiming,
) -> RelayResult<SubmissionReport> {
    tracing::info!(
        tracking_key = %plan.tracking_key,
        task_type = r.task_type.as_str(),
        model_name = %r.model_name,
        ratio = %r.ratio,
        duration = r.duration,
        submission_mode = plan.submission_mode,
        prompt_length = plan.prompt.chars().count(),
        prompt_hash = %prompt_hash(&plan.prompt),
        reference_count = reference_paths.len(),
        "开始提交生成请求"
    );
    if cancel.is_cancelled() {
        return Err(cancelled("start"));
    }
    // 只有确认当前页面已经是可用的图片生成页时，才复用页面并直接发送。
    // 不能仅凭“没有参考图”判断，否则旧会话/普通聊天页会跳过模型、比例和
    // 输入框准备，最终表现为已点击但没有可识别的提交身份。
    let direct_image = r.task_type == TaskType::Image
        && reference_paths.is_empty()
        && session
            .evaluate_as::<bool>(image_page_ready_expression())
            .await
            .unwrap_or(false);
    let js_request = if direct_image {
        on_stage(TaskStage::Preparing, Some("复用当前豆包图片页面".into()));
        build_js_request(r, plan, reference_paths, 0, false)
    } else {
        prepare_and_fill(session, r, plan, reference_paths, timing, cancel, on_stage).await?
    };

    // 订阅要在 send 之前,否则事件会漏。
    let mut chat_rx = Some(session.subscribe());
    let mut bg_rx = background.map(CdpSession::subscribe);
    enable_network(session).await?;
    if let Some(bg) = background {
        if let Err(e) = enable_network(bg).await {
            tracing::warn!(code = e.code, error = %e, "background 页 Network.enable 失败,只监听 chat 页");
            bg_rx = None;
        }
    }
    if cancel.is_cancelled() {
        clear_input(session).await;
        if !reference_paths.is_empty() {
            upload::cleanup_attachments(session).await;
        }
        return Err(cancelled("send"));
    }

    let send_started_at = now_ms();
    let send_expression = if direct_image {
        direct_image_prompt_send_expression(&plan.prompt)
    } else {
        step_expression(&plan.tracking_key, "send", &js_request)
    };
    if let Err(e) = evaluate_step(session, &send_expression, timing.step_timeout, cancel).await {
        clear_input(session).await;
        if !reference_paths.is_empty() {
            upload::cleanup_attachments(session).await;
        }
        return Err(map_submit("send", e));
    }
    on_stage(TaskStage::Sent, None);
    tracing::info!(tracking_key = %plan.tracking_key, send_started_at, "已点击发送,开始识别提交身份");
    // 首次上传素材时豆包在发送前弹"安全确认";确认后才发出生成请求,所以识别窗口从这之后开始算。
    accept_safety_dialog(session, timing).await;
    minimize_window_after_submission(session).await;

    let deadline = Instant::now() + timing.identity_window;
    let mut candidates = 0usize;
    let mut observed: Vec<ObservedRequest> = Vec::new();
    let mut reader_ready_chat = false;
    let mut reader_ready_bg = false;
    loop {
        let remaining = deadline.saturating_duration_since(Instant::now());
        if remaining.is_zero() {
            break;
        }
        let (from_bg, ev) = tokio::select! {
            _ = cancel.cancelled() => return Err(cancelled("identity")),
            ev = next_event(&mut chat_rx) => (false, ev),
            ev = next_event(&mut bg_rx) => (true, ev),
            _ = tokio::time::sleep(remaining) => break,
        };
        let Some(cand) = candidate_from_event(&ev) else {
            continue;
        };
        if !is_generation_request(&cand.url, &cand.method) {
            continue;
        }
        if let Some(t) = cand.wall_time_ms {
            // 发送之前的请求不算(容忍 1.5s 时钟差)。
            if t < send_started_at - 1500 {
                continue;
            }
        }
        candidates += 1;
        tracing::debug!(url = %cand.url, from_bg, "发现生成请求候选");
        if !cand.request_id.is_empty() {
            let target_id = match (from_bg, background) {
                (true, Some(bg)) => bg.target_id.clone(),
                _ => session.target_id.clone(),
            };
            observed.push(ObservedRequest {
                target_id,
                request_id: cand.request_id.clone(),
            });
        }
        let identity = if from_bg {
            match background {
                Some(bg) => {
                    identify_on(bg, &mut reader_ready_bg, cand, &plan.prompt, timing, cancel).await
                }
                None => None,
            }
        } else {
            identify_on(
                session,
                &mut reader_ready_chat,
                cand,
                &plan.prompt,
                timing,
                cancel,
            )
            .await
        };
        if let Some(identity) = identity {
            tracing::info!(
                tracking_key = %plan.tracking_key,
                local_message_id = identity.local_message_id.as_deref().unwrap_or(""),
                message_id = identity.message_id.as_deref().unwrap_or(""),
                conversation_id = identity.conversation_id.as_deref().unwrap_or(""),
                elapsed_ms = (now_ms() - send_started_at).max(0),
                "已识别提交身份"
            );
            return Ok(SubmissionReport {
                identity: Some(identity),
                send_started_at,
                generation_requests: observed,
            });
        }
    }
    tracing::warn!(
        tracking_key = %plan.tracking_key,
        candidates,
        window_ms = timing.identity_window.as_millis() as u64,
        "识别窗口内未得到提交身份"
    );
    Err(identity_missing(send_started_at, candidates))
}

/// 豆包"安全确认"弹窗(上传素材的授权声明,`safety_authorization_dialog`):短暂等待它出现,
/// 出现就点"确认"(产品决定:默认同意,不提供开关)。没有弹窗时最多多等 [`SAFETY_DIALOG_WAIT_MS`]。
pub const SAFETY_DIALOG_WAIT_MS: u64 = 800;

pub fn safety_dialog_expression() -> String {
    format!(
        r#"(async () => {{
  const visible = e => {{ const r = e.getBoundingClientRect(); return r.width > 0 && r.height > 0; }};
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  const dialog = () => [...document.querySelectorAll('[data-testid=safety_authorization_dialog]')].find(visible);
  const end = Date.now() + {SAFETY_DIALOG_WAIT_MS};
  let d = dialog();
  while (!d && Date.now() < end) {{ await sleep(50); d = dialog(); }}
  if (!d) return {{ found: false }};
  const btn = d.querySelector('[data-testid=safety_authorization_confirm_btn]')
    || [...d.querySelectorAll('button')].find(b => (b.innerText || '').trim() === '确认');
  if (!btn) return {{ found: true, clicked: false }};
  btn.click();
  const t0 = Date.now();
  while (dialog() && Date.now() - t0 < 2000) await sleep(50);
  return {{ found: true, clicked: true, closed: !dialog() }};
}})()"#
    )
}

async fn accept_safety_dialog(session: &CdpSession, timing: &SubmitTiming) {
    let r = session
        .send_timeout(
            "Runtime.evaluate",
            json!({
                "expression": safety_dialog_expression(),
                "returnByValue": true,
                "awaitPromise": true,
            }),
            timing.step_timeout,
        )
        .await;
    let v = match r {
        Ok(v) => v.pointer("/result/value").cloned().unwrap_or(Value::Null),
        Err(e) => {
            tracing::warn!(error = %e, "检查豆包安全确认弹窗失败");
            return;
        }
    };
    if v.get("found").and_then(Value::as_bool) != Some(true) {
        return;
    }
    if v.get("clicked").and_then(Value::as_bool) == Some(true) {
        let closed = v.get("closed").and_then(Value::as_bool).unwrap_or(false);
        tracing::info!(closed, "已自动点击豆包安全确认弹窗的「确认」");
    } else {
        tracing::warn!("豆包安全确认弹窗里没找到「确认」按钮");
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::tasks::ConfirmationMode;
    use std::sync::{Arc, Mutex};

    fn image_request() -> GenerateRequest {
        GenerateRequest {
            account: "张三".into(),
            task_type: TaskType::Image,
            model_name: "Seedream 5.0 Pro".into(),
            prompt: "一只蓝色小鸟".into(),
            ratio: "1:1".into(),
            duration: None,
            reference_files: vec![],
            confirmation_mode: ConfirmationMode::Auto,
            auto_use_membership: true,
            callback_url: None,
        }
    }

    fn video_request() -> GenerateRequest {
        GenerateRequest {
            task_type: TaskType::Video,
            model_name: "Seedance 2.5".into(),
            ratio: "16:9".into(),
            duration: Some(10),
            confirmation_mode: ConfirmationMode::Manual,
            auto_use_membership: false,
            ..image_request()
        }
    }

    const TASK_ID: &str = "01HQZ9ABCDEFGHJKMNPQRSTVWX";

    // ---- 纯函数 ----

    #[test]
    fn plan_image_is_conversation_video_is_skill_prompt_unchanged() {
        let p = SubmissionPlan::for_request(&image_request(), TASK_ID);
        assert_eq!(p.submission_mode, "conversation");
        assert_eq!(p.prompt, "一只蓝色小鸟");
        assert_eq!(
            p.tracking_key,
            "shenlangDoubaoTask_01hqz9abcdefghjkmnpqrstvwx"
        );

        let p = SubmissionPlan::for_request(&video_request(), TASK_ID);
        assert_eq!(p.submission_mode, "skill");
        assert_eq!(p.prompt, "一只蓝色小鸟");
        // Mini 11–15s 也走 skill,不拼"请使用 Seedance 2.0 Mini"前缀
        let mini = GenerateRequest {
            model_name: "Seedance 2.0 Mini".into(),
            duration: Some(12),
            ..video_request()
        };
        let p = SubmissionPlan::for_request(&mini, TASK_ID);
        assert_eq!(p.submission_mode, "skill");
        assert_eq!(p.prompt, "一只蓝色小鸟");
    }

    #[test]
    fn build_js_request_uses_camel_case_fields() {
        let r = video_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let v = build_js_request(
            &r,
            &plan,
            &[PathBuf::from("/tmp/a.png"), PathBuf::from("/tmp/b.jpg")],
            1,
            true,
        );
        let obj = v.as_object().unwrap();
        assert_eq!(obj["type"], "video");
        assert_eq!(obj["modelName"], "Seedance 2.5");
        assert_eq!(obj["prompt"], "一只蓝色小鸟");
        assert_eq!(obj["ratio"], "16:9");
        assert_eq!(obj["duration"], 10);
        assert_eq!(obj["referenceFiles"], json!(["/tmp/a.png", "/tmp/b.jpg"]));
        assert_eq!(obj["submissionMode"], "skill");
        assert_eq!(obj["parallelTracking"], true);
        assert_eq!(obj["trackingKey"], plan.tracking_key);
        assert_eq!(obj["confirmationMode"], "manual");
        assert_eq!(obj["autoUseMembership"], false);
        assert_eq!(obj["uploadedCount"], 1);
        assert_eq!(obj["userConfirmed"], true);
        assert_eq!(obj.len(), 13);
        for k in obj.keys() {
            assert!(!k.contains('_'), "字段应为 camelCase: {k}");
        }

        let r = image_request();
        let v = build_js_request(&r, &SubmissionPlan::for_request(&r, TASK_ID), &[], 0, false);
        assert!(v.get("duration").is_none());
        assert_eq!(v["submissionMode"], "conversation");
        assert_eq!(v["referenceFiles"], json!([]));
        assert_eq!(v["confirmationMode"], "auto");
    }

    #[test]
    fn step_expression_wraps_run_in_promise() {
        let e = step_expression("shenlangDoubaoTask_x", "prepare", &json!({ "a": 1 }));
        assert_eq!(
            e,
            r#"Promise.resolve(shenlangDoubaoTask_x.run("prepare", {"a":1}))"#
        );
    }

    #[test]
    fn generation_bootstrap_guards_and_renames_global() {
        let key = "shenlangDoubaoTask_abc";
        let b = generation_bootstrap(key);
        assert!(b.starts_with("if (typeof shenlangDoubaoTask_abc === 'undefined') {"));
        assert!(b.trim_end().ends_with('}'));
        assert!(!b.contains(GENERATION_GLOBAL));
        assert!(b.contains("'shenlangDoubaoTask_abc'"));
        assert!(b.contains("normalizeRatioControlForCurrentDoubao"));
        assert!(b.contains("`比例 ${ratio}`"));
        // 脚本只在字符串表里出现一次全局名,替换后恰好一处
        assert_eq!(GENERATION_SCRIPT.matches(GENERATION_GLOBAL).count(), 1);
        let r = reader_bootstrap();
        assert!(r.starts_with("if (typeof shenlangDoubaoResultReader === 'undefined') {"));
        assert!(r.contains("submissionIdentity"));
    }

    #[test]
    fn new_generation_toolbar_waits_for_video_controls_to_stabilize() {
        let expression =
            open_new_generation_toolbar_expression(TaskType::Video, Duration::from_secs(2));

        assert!(expression.contains("wait_generation_controls_ready"));
        assert!(!expression.contains("skill_input_exit_button"));
        assert!(expression.contains("Seedance"));
        assert!(expression.contains("\\d+\\s*s"));
    }

    #[test]
    fn new_generation_toolbar_accepts_current_doubao_model_prefixes() {
        let image = open_new_generation_toolbar_expression(TaskType::Image, Duration::from_secs(2));
        assert!(image.contains("(?:(?:图像|图片)生成)?模型\\s*Seedream"));

        let video = open_new_generation_toolbar_expression(TaskType::Video, Duration::from_secs(2));
        assert!(video.contains("(?:视频生成)?模型\\s*Seedance"));
    }

    #[test]
    fn new_generation_toolbar_reads_current_non_button_controls() {
        let expression =
            open_new_generation_toolbar_expression(TaskType::Image, Duration::from_secs(2));
        assert!(expression.contains("button,[role=\"button\"],div,span"));
    }

    #[test]
    fn new_generation_toolbar_ignores_controls_hidden_by_a_parent_container() {
        let expression =
            open_new_generation_toolbar_expression(TaskType::Image, Duration::from_secs(2));
        assert!(expression.contains("for (let node = el; node; node = node.parentElement)"));
        assert!(expression.contains("s.visibility === 'hidden'"));
    }

    #[test]
    fn exit_skill_ignores_a_button_hidden_by_its_parent_container() {
        let expression = exit_skill_expression(Duration::from_secs(2));
        assert!(expression.contains("for (let node = el; node; node = node.parentElement)"));
        assert!(expression.contains("s.visibility === 'hidden'"));
    }

    #[test]
    fn new_generation_toolbar_scopes_entries_and_controls_to_the_input_engine() {
        let expression =
            open_new_generation_toolbar_expression(TaskType::Video, Duration::from_secs(2));

        assert!(expression.contains("const root = document.querySelector('#input-engine-container"));
        assert!(expression.contains("root.querySelectorAll('button,[role=\"button\"],div,span')"));
        assert!(!expression.contains("[...document.querySelectorAll('button,[role=\"button\"]')"));
    }

    #[test]
    fn new_generation_toolbar_supports_current_chat_input_container() {
        let expression =
            open_new_generation_toolbar_expression(TaskType::Image, Duration::from_secs(2));

        assert!(expression.contains("[data-testid=\"chat_input\"]"));
    }

    #[test]
    fn new_generation_toolbar_opens_a_fresh_chat_when_the_input_is_empty() {
        let expression =
            open_new_generation_toolbar_expression(TaskType::Image, Duration::from_secs(2));

        assert!(
            expression.contains("const NEW_CHAT = '[data-testid=\"create_conversation_button\"]'")
        );
        assert!(expression.contains("press(newChat)"));
        assert!(expression.contains("new_chat_open_failed"));
    }

    #[test]
    fn new_generation_toolbar_reuses_an_already_visible_current_toolbar() {
        let expression =
            open_new_generation_toolbar_expression(TaskType::Image, Duration::from_secs(2));

        assert!(expression.contains(
            "if (controlsReady()) return { opened: true, selected: TARGET, reused: true };"
        ));
    }

    #[test]
    fn exit_skill_recognizes_current_generation_toolbar_before_opening_new_chat() {
        let expression = exit_skill_expression(Duration::from_secs(2));

        assert!(expression.contains("generationReady"));
        assert!(expression.contains("[data-testid=\"chat_input\"]"));
        assert!(expression.contains("模型\\s*(?:Seedream|Seedance)"));
    }

    #[test]
    fn new_generation_toolbar_accepts_ratio_inside_current_combined_control() {
        let image = open_new_generation_toolbar_expression(TaskType::Image, Duration::from_secs(2));
        assert!(image.contains("has(/比例\\s*(自动|\\d+:\\d+)/)"));

        let video = open_new_generation_toolbar_expression(TaskType::Video, Duration::from_secs(2));
        assert!(video.contains("has(/(?:自动|\\d+:\\d+)\\s*[·•.]?\\s*\\d+\\s*s$/i)"));
    }

    #[test]
    fn new_generation_toolbar_does_not_require_removed_skill_exit_control() {
        let expression =
            open_new_generation_toolbar_expression(TaskType::Image, Duration::from_secs(2));
        assert!(!expression.contains("return exit &&"));
    }

    #[test]
    fn generation_page_script_accepts_current_doubao_model_prefixes() {
        assert!(GENERATION_SCRIPT.contains("(?:(?:图像|图片)生成)?模型\\s*Seedream"));
        assert!(GENERATION_SCRIPT.contains("(?:视频生成)?模型\\s*Seedance"));
        assert!(GENERATION_SCRIPT.contains("/比例\\s*(自动|\\d+:\\d+)/"));
        assert!(!GENERATION_SCRIPT.contains("!/(比例|风格|模板)/"));
    }

    #[test]
    fn generation_compatibility_normalizes_current_combined_video_settings() {
        let script = generation_ratio_compatibility_script("shenlangDoubaoTask_test");

        assert!(script.contains("自动|\\d+:\\d+"));
        assert!(script.contains("data-shenlang-video-settings-summary"));
        assert!(script.contains("${ratio} · ${seconds}s"));
        assert!(script.contains("ensureCurrentVideoEntry"));
    }

    #[test]
    fn generation_compatibility_uses_the_leaf_for_nested_video_setting_labels() {
        let script = generation_ratio_compatibility_script("shenlangDoubaoTask_test");

        assert!(script.contains("const settingsCandidates ="));
        assert!(script.contains("!settingsCandidates.some(other => other !== entry && entry.control.contains(other.control))"));
    }

    #[test]
    fn generation_compatibility_has_a_scoped_video_prepare_fallback() {
        let script = generation_ratio_compatibility_script("shenlangDoubaoTask_test");

        assert!(script.contains("videoPrepareFallback"));
        assert!(script.contains("mode: 'video'"));
        assert!(script.contains("needsDurationInput: false"));
        assert!(script.contains("videoCheckSendFallback"));
        assert!(script.contains("editorText.includes(prompt)"));
        assert!(script.contains("if (step !== 'prepare' || request?.type !== 'video')"));
        assert!(script.contains("新版视频控件诊断"));
    }

    #[test]
    fn generation_compatibility_checks_current_image_skill_before_send() {
        let script = generation_ratio_compatibility_script("shenlangDoubaoTask_test");

        assert!(script.contains("rememberImageSettings"));
        assert!(script.contains("imageCheckSendFallback"));
        assert!(script.contains("data-shenlang-image-model"));
        assert!(script.contains("data-shenlang-image-ratio"));
        assert!(script.contains("editorText.includes(prompt)"));
        assert!(script.contains("if (step === 'checkSend' && request?.type === 'image')"));
    }

    #[test]
    fn generation_compatibility_supports_current_chat_send_button() {
        let script = generation_ratio_compatibility_script("shenlangDoubaoTask_test");

        assert!(script.contains("if (step === 'send' && request?.type === 'image')"));
        assert!(script.contains("chat_input_send_button"));
    }

    #[test]
    fn generation_compatibility_can_select_an_image_ratio_from_the_combined_toolbar_control() {
        let script = generation_ratio_compatibility_script("shenlangDoubaoTask_test");

        assert!(script.contains("imagePrepareFallback"));
        assert!(script.contains("图片比例候选"));
        assert!(script.contains("const ratioCandidate = await waitForMenuChoice(ratio)"));
        assert!(script.contains("[role=\"menuitemradio\"],button,[role=\"button\"],span"));
        assert!(script.contains("value === expected || value.startsWith(`${expected} `)"));
        assert!(script.contains("entry.closest('[role=\"menuitem\"],[role=\"option\"],[role=\"menuitemradio\"],button,[role=\"button\"]')"));
    }

    #[test]
    fn generation_compatibility_opens_the_current_image_settings_menu() {
        let script = generation_ratio_compatibility_script("shenlangDoubaoTask_test");

        assert!(script.contains("button[aria-haspopup=\"menu\"]"));
        assert!(script.contains("[data-input-engine-actionbar-control-key=\"model\"]"));
        assert!(script.contains("[data-creation-params-panel-id]"));
        assert!(script.contains("role=\"menuitem\""));
        assert!(script.contains("图片设置菜单未打开"));
    }

    #[test]
    fn generation_compatibility_reuses_visible_model_and_ratio_controls_without_a_parent_menu() {
        let script = generation_ratio_compatibility_script("shenlangDoubaoTask_test");
        let direct_controls = script
            .find("const currentSettings = imageSettingsControls();")
            .expect("应先读取新版豆包直接显示的模型和比例控件");
        let menu_fallback = script
            .find("const settings = await openImageSettings();")
            .expect("直接控件不可用时仍应保留菜单回退");

        assert!(direct_controls < menu_fallback);
        assert!(script.contains("if (currentSettings.model && currentSettings.ratio)"));
    }

    #[test]
    fn generation_compatibility_uses_current_image_skill_before_legacy_mode_switch() {
        let script = generation_ratio_compatibility_script("shenlangDoubaoTask_test");
        let image_branch = script
            .split("if (step === 'prepare' && request?.type === 'image')")
            .nth(1)
            .expect("应包含图片 prepare 兼容分支");
        let original_run = image_branch
            .find("return await originalRun(step, request)")
            .expect("应先让原页面脚本切换图片模式");
        let fallback = image_branch
            .find("const currentFallback = await imagePrepareFallback(request)")
            .expect("应读取当前新版图片控件");

        assert!(fallback < original_run);
        assert!(image_branch.contains("new MutationObserver(() => normalize(request))"));
    }

    #[test]
    fn generation_compatibility_scopes_video_entry_to_the_input_engine() {
        let script = generation_ratio_compatibility_script("shenlangDoubaoTask_test");

        assert!(script
            .contains("const inputRoot = () => document.querySelector('#input-engine-container"));
        assert!(script.contains("const root = inputRoot();"));
        assert!(script.contains("root.querySelectorAll('button,[role=\"button\"]')"));
        assert!(!script.contains("const buttons = () => [...document.querySelectorAll('button,[role=\"button\"]')].filter(visible);"));
    }

    #[test]
    fn generation_compatibility_supports_current_chat_input_container() {
        let script = generation_ratio_compatibility_script("shenlangDoubaoTask_test");

        assert!(script.contains("[data-testid=\"chat_input\"]"));
    }

    #[test]
    fn native_generation_mode_prefers_the_portal_target_over_conversation() {
        let expression = native_generation_mode_probe_expression(TaskType::Image, true, true);
        let target_menu = expression
            .find("const targetMenu = modeMenuOpen")
            .expect("应在全页面浮层中查找图像生成");
        let controls_ready = expression
            .find("const generationControlsReady = (has(/^(?:(?:图像|图片)生成)?模型")
            .expect("应先判断图像控件是否已经就绪");
        let input_engine_guard = expression
            .rfind("if (!root || !visible(root))")
            .expect("应保留输入区兼容判断");
        let fresh_chat = expression
            .find("if (ALLOW_FRESH_CHAT && visible(newChat))")
            .expect("结果页应允许通过新对话恢复输入区");
        let conversation_menu = expression
            .find("if (modeMenuOpen && visible(modeTrigger)")
            .expect("应保留对话菜单兼容逻辑");

        assert!(controls_ready < target_menu);
        assert!(expression.contains("[data-testid=\"chat_input_input\"]"));
        assert!(expression.contains("input?.closest('.guidance-input-surface')"));
        assert!(target_menu < input_engine_guard);
        assert!(fresh_chat < input_engine_guard);
        assert!(target_menu < conversation_menu);
    }

    #[test]
    fn native_generation_mode_opens_fresh_chat_before_selecting_generation_mode() {
        let expression = native_generation_mode_probe_expression(TaskType::Image, true, true);
        let conversation_button = expression
            .find("if (visible(mode))")
            .expect("已有输入区时应先使用模式按钮");
        let fresh_chat = expression
            .find("if (ALLOW_FRESH_CHAT && visible(newChat))")
            .expect("输入区不可用时仍应允许新建对话");

        assert!(
            conversation_button < fresh_chat,
            "已有输入区但没有目标技能时,应先尝试当前页面的模式按钮"
        );
    }

    #[test]
    fn native_generation_mode_opens_skill_more_before_conversation_mode() {
        let expression = native_generation_mode_probe_expression(TaskType::Image, true, true);
        let skill_more = expression
            .find("[data-testid=\"skill_bar_button_more\"]")
            .expect("新版豆包应从更多菜单进入图像生成");
        let conversation_button = expression
            .find("if (visible(mode))")
            .expect("应保留旧版对话模式兼容");

        assert!(skill_more < conversation_button);
        assert!(expression.contains("'skill_more'"));
    }

    #[test]
    fn native_image_settings_probe_uses_current_hidden_controls() {
        let expression = native_image_settings_probe_expression("Seedream 4.5", "16:9");

        assert!(expression.contains("button[aria-haspopup=\"menu\"]"));
        assert!(expression.contains("[data-input-engine-actionbar-control-key=\"model\"]"));
        assert!(expression.contains("[data-creation-params-panel-id]"));
        assert!(expression.contains("Seedream 4.5"));
        assert!(expression.contains("16:9"));
        assert!(expression.contains("value === expected || value.startsWith(`${expected} `)"));
    }

    #[test]
    fn native_generation_mode_requires_stable_controls_before_prepare() {
        let expression = native_generation_mode_probe_expression(TaskType::Image, true, true);

        assert!(expression.contains("data-shenlang-generation-controls-ready-since"));
        assert!(expression.contains("Date.now() - controlsReadySince < 750"));
        assert!(expression.contains("reason: 'generation_controls_stabilizing'"));
    }

    #[test]
    fn native_generation_mode_closes_an_already_selected_conversation_menu_by_trigger() {
        let expression = native_generation_mode_probe_expression(TaskType::Image, true, true);

        assert!(expression.contains("reason: 'conversation_menu_already_selected'"));
        assert!(expression.contains("const modeTrigger = document.querySelector(MODE)"));
        assert!(expression.contains("svg path[d*=\"M20.2246\"]"));
        assert!(expression.contains("modeIsConversation"));
        assert!(expression.contains("conversation_menu_open"));
        assert!(expression.contains("dismissPoint"));
        assert!(expression.contains("data-state=\"open\""));
        let source = include_str!("generation.rs");
        assert!(source.contains("dispatch_trusted_click(session, &point"));
        assert!(expression.contains("Windows 版比单纯发 Escape 更稳定"));
    }

    #[test]
    fn native_generation_mode_dismisses_conversation_menu_before_clicking_hidden_skill() {
        let expression = native_generation_mode_probe_expression(TaskType::Image, true, true);
        let menu_state = expression
            .find("const modeMenuOpen =")
            .expect("应先识别对话下拉菜单是否仍然打开");
        let target = expression
            .find("const targetMenu = modeMenuOpen")
            .expect("应查找图像生成入口");

        assert!(
            menu_state < target,
            "对话菜单还开着时,不得点击被菜单盖住的图像生成按钮"
        );
        assert!(expression.contains("const targetMenu = modeMenuOpen"));
    }

    #[test]
    fn native_image_preparation_uses_the_full_prepare_deadline() {
        let source = include_str!("generation.rs");
        let selection = source
            .split("async fn select_native_generation_mode(")
            .nth(1)
            .unwrap()
            .split("async fn prepare_native_image_settings(")
            .next()
            .unwrap();
        let settings = source
            .split("async fn prepare_native_image_settings(")
            .nth(1)
            .unwrap()
            .split("async fn prepare_and_fill(")
            .next()
            .unwrap();

        assert!(selection.contains("Instant::now() + timing.prepare_timeout"));
        assert!(settings.contains("Instant::now() + timing.prepare_timeout"));
    }

    #[test]
    fn native_image_settings_waits_through_a_toolbar_remount() {
        let source = include_str!("generation.rs");
        let production = source.split("mod tests {").next().unwrap();

        assert!(production.contains("let mut unsupported_since = None;"));
        assert!(production.contains("image_settings_unsupported"));
        assert!(production.contains("Duration::from_secs(5)"));
    }

    #[test]
    fn native_generation_mode_accepts_the_active_skill_before_reclicking_its_entry() {
        let expression = native_generation_mode_probe_expression(TaskType::Image, true, true);
        let active_skill = expression
            .find("const activeSkill =")
            .expect("应识别已经进入的图像技能");
        let target_menu = expression
            .find("const targetMenu = modeMenuOpen")
            .expect("应保留技能入口查找");

        assert!(active_skill < target_menu);
        assert!(expression.contains("if (activeSkill && !generationControlsReady)"));
        assert!(expression.contains("reason: 'generation_controls_mounting'"));
    }

    #[test]
    fn new_generation_toolbar_keeps_waiting_for_slow_image_skill_controls() {
        let expression =
            open_new_generation_toolbar_expression(TaskType::Image, Duration::from_secs(8));

        assert!(expression.contains("while (Date.now() - started < 8000)"));
    }

    #[test]
    fn identity_expression_embeds_request_prompt_and_fixes_big_ints() {
        let e = identity_expression(
            "https://www.doubao.com/samantha/chat/completion",
            "POST",
            r#"{"conversation_id":12345678901234567890,"text":"一只蓝色小鸟"}"#,
            "一只\"蓝色\"小鸟",
        );
        assert!(e.starts_with("Promise.resolve(shenlangDoubaoResultReader.submissionIdentity({"));
        assert!(e.ends_with(", { verifiedSend: false }))"));
        assert!(e.contains(r#""method":"POST""#));
        assert!(e.contains(r#""url":"https://www.doubao.com/samantha/chat/completion""#));
        // postData 是字符串,里面的大整数已包成字符串(转义后的引号)
        assert!(e.contains(r#"\"conversation_id\":\"12345678901234567890\""#));
        assert!(e.contains(r#", "一只\"蓝色\"小鸟", {"#));
    }

    #[test]
    fn is_generation_request_table() {
        assert!(is_generation_request(
            "https://www.doubao.com/samantha/chat/completion",
            "POST"
        ));
        assert!(is_generation_request(
            "https://www.doubao.com/samantha/chat/completion?aid=1",
            "post"
        ));
        assert!(is_generation_request(
            "https://api.snssdk.com/x/generate",
            "POST"
        ));
        assert!(is_generation_request(
            "https://www.doubao.com/im/message/send",
            "POST"
        ));
        assert!(!is_generation_request(
            "https://www.doubao.com/samantha/chat/history",
            "POST"
        ));
        assert!(!is_generation_request(
            "https://www.doubao.com/samantha/chat/completion",
            "GET"
        ));
        assert!(!is_generation_request(
            "http://www.doubao.com/samantha/chat/completion",
            "POST"
        ));
        assert!(!is_generation_request(
            "https://evil-doubao.com/chat/completion",
            "POST"
        ));
        assert!(!is_generation_request(
            "https://www.doubao.com/generate/config",
            "POST"
        ));
        assert!(!is_generation_request(
            "https://www.doubao.com/im/chain/single",
            "POST"
        ));
        assert!(!is_generation_request("not a url", "POST"));
    }

    #[test]
    fn parse_identity_shapes() {
        assert_eq!(parse_identity(&Value::Null), None);
        assert_eq!(parse_identity(&json!({})), None);
        assert_eq!(
            parse_identity(
                &json!({ "localMessageId": null, "messageId": null, "conversationId": "1" })
            ),
            None
        );
        let id = parse_identity(&json!({
            "localMessageId": "abc-123", "messageId": null, "conversationId": "12345678901234567"
        }))
        .unwrap();
        assert_eq!(id.local_message_id.as_deref(), Some("abc-123"));
        assert_eq!(id.message_id, None);
        assert_eq!(id.conversation_id.as_deref(), Some("12345678901234567"));
        let id = parse_identity(&json!({ "localMessageId": null, "messageId": "m9" })).unwrap();
        assert_eq!(id.local_message_id, None);
        assert_eq!(id.message_id.as_deref(), Some("m9"));
        assert_eq!(id.conversation_id, None);
        // "0" 当作没有
        let id = parse_identity(&json!({ "localMessageId": "x", "conversationId": "0" })).unwrap();
        assert_eq!(id.conversation_id, None);
    }

    #[test]
    fn error_constructors_carry_codes_and_details() {
        let e = preparing_failed("prepare", Some("未找到所选生成比例。"));
        assert_eq!(e.code, "generation_preparing_failed");
        assert_eq!(
            e.details.as_ref().unwrap()["js_exception"],
            "未找到所选生成比例。"
        );
        assert_eq!(e.details.unwrap()["reason"], "prepare");
        assert_eq!(
            preparing_timeout("prepare").code,
            "generation_preparing_timeout"
        );
        let e = submit_failed("send", Some("豆包发送按钮尚未就绪，未发送。"));
        assert_eq!(e.code, "generation_submit_failed");
        assert_eq!(
            e.details.unwrap()["js_exception"],
            "豆包发送按钮尚未就绪，未发送。"
        );
        let e = identity_missing(123, 2);
        assert_eq!(e.code, "generation_submit_identity_missing");
        assert_eq!(e.details.as_ref().unwrap()["sent"], true);
        assert_eq!(e.details.as_ref().unwrap()["send_started_at"], 123);
        assert_eq!(e.details.unwrap()["candidates"], 2);
        assert_eq!(cancelled("x").code, "generation_cancelled");
        for e in [
            preparing_failed("a", None),
            preparing_timeout("a"),
            submit_failed("a", None),
            identity_missing(0, 0),
            cancelled("a"),
        ] {
            let last = e.message.chars().last().unwrap();
            assert!(
                !matches!(last, '。' | '.' | '!' | '\u{FF01}' | ',' | '\u{FF0C}'),
                "{}",
                e.message
            );
            assert_eq!(e.http_status(), 500);
        }
    }

    #[test]
    fn exit_skill_and_clear_expressions_are_self_contained() {
        let e = exit_skill_expression(Duration::from_millis(3000));
        assert!(e.starts_with("(async () => {"));
        assert!(e.ends_with("})()"));
        assert!(e.contains("skill_input_exit_button"));
        assert!(e.contains("guidance-skill-bar"));
        assert!(e.contains("< 3000"));
        let c = clear_input_expression();
        assert!(c.contains("execCommand('selectAll')"));
        assert!(c.contains("execCommand('delete')"));
    }

    #[test]
    fn submit_timing_defaults_follow_spec() {
        let t = SubmitTiming::default();
        assert_eq!(t.prepare_timeout, Duration::from_secs(120));
        assert_eq!(t.step_timeout, Duration::from_secs(20));
        assert_eq!(t.identity_window, Duration::from_secs(15));
        assert_eq!(t.skill_exit_timeout, Duration::from_secs(3));
    }

    #[tokio::test]
    async fn submitted_task_minimizes_doubao_without_closing_its_cdp_session() {
        let (session, log) = spawn_mock_page(happy_handler()).await;

        minimize_window_after_submission(&session).await;

        let log = log.lock().unwrap().clone();
        assert!(log
            .iter()
            .any(|line| line.starts_with("Browser.getWindowForTarget ")));
        assert!(log.iter().any(|line| {
            line.starts_with("Browser.setWindowBounds ")
                && line.contains("\"windowId\":92")
                && line.contains("\"windowState\":\"minimized\"")
        }));
        assert!(!session.is_closed());
        session.close().await;
    }

    // ---- mock 豆包页面 ----

    /// 对一条 `Runtime.evaluate` 的回应:先推 `events`,再回 `value` 或 JS 异常。
    #[derive(Default)]
    struct MockReply {
        value: Value,
        exception: Option<String>,
        events: Vec<Value>,
    }

    fn ok(value: Value) -> MockReply {
        MockReply {
            value,
            ..Default::default()
        }
    }

    fn throw(msg: &str) -> MockReply {
        MockReply {
            exception: Some(msg.to_string()),
            ..Default::default()
        }
    }

    /// 绝大多数旧流程用例只关心 prepare/fill/send。工具栏步骤加入后，未声明
    /// 此步骤的 mock 以空值回应；这不代表真实页面打不开，因此按当前页面已就绪
    /// 处理。显式给出工具栏回应的用例仍完整覆盖失败和等待分支。
    fn default_generation_toolbar_reply(kind: &str, reply: MockReply) -> MockReply {
        if kind == "generation_toolbar"
            && reply.exception.is_none()
            && !reply
                .value
                .as_object()
                .is_some_and(|value| value.contains_key("opened"))
        {
            return ok(json!({ "opened": true, "selected": "mock" }));
        }
        reply
    }

    type Handler = Arc<Mutex<Box<dyn FnMut(&str) -> MockReply + Send>>>;
    type Log = Arc<Mutex<Vec<String>>>;

    /// 把表达式归类成短名,便于断言调用顺序。
    fn classify(expr: &str) -> String {
        if expr.contains("native_generation_mode_probe") {
            "generation_mode".into()
        } else if expr.contains("const expectedModel =")
            && expr.contains("data-input-engine-actionbar-control-key")
        {
            "image_settings".into()
        } else if expr.contains("open_new_generation_toolbar") {
            "generation_toolbar".into()
        } else if expr.starts_with("if (typeof shenlangDoubaoTask_") {
            "bootstrap".into()
        } else if expr.starts_with("if (typeof shenlangDoubaoResultReader") {
            "reader".into()
        } else if expr.contains("skill_input_exit_button") {
            "exit_skill".into()
        } else if expr.contains("wait_generation_skill_bar") {
            "wait_skill_bar".into()
        } else if expr.contains("attachment-delete-btn") {
            "attachments".into()
        } else if expr.contains("upload_file_button") {
            "open_upload".into()
        } else if expr.contains("safety_authorization_dialog") {
            "safety".into()
        } else if expr.contains("execCommand('selectAll')") {
            "clear".into()
        } else if expr.contains(".submissionIdentity(") {
            "identity".into()
        } else if let Some(rest) = expr.strip_prefix("Promise.resolve(shenlangDoubaoTask_") {
            let step = rest
                .split(".run(\"")
                .nth(1)
                .unwrap_or("")
                .split('"')
                .next()
                .unwrap_or("");
            step.to_string()
        } else {
            format!("other:{}", &expr[..expr.len().min(30)])
        }
    }

    /// mock 页面:`Runtime.evaluate` 按归类交给 handler;其它命令回 `{}` 并记录 `method params`。
    async fn spawn_mock_page(handler: Handler) -> (Arc<CdpSession>, Log) {
        use futures_util::{SinkExt, StreamExt};
        use tokio_tungstenite::tungstenite::Message;
        let log: Log = Arc::new(Mutex::new(Vec::new()));
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        let log2 = Arc::clone(&log);
        tokio::spawn(async move {
            let (sock, _) = listener.accept().await.unwrap();
            let mut ws = tokio_tungstenite::accept_async(sock).await.unwrap();
            while let Some(Ok(Message::Text(txt))) = ws.next().await {
                let req: Value = serde_json::from_str(&txt).unwrap();
                let method = req["method"].as_str().unwrap_or("").to_string();
                let result = if method == "Runtime.evaluate" {
                    let expr = req["params"]["expression"].as_str().unwrap_or("");
                    let kind = classify(expr);
                    // 图片页就绪探针是无副作用的内部判断，不计入生成步骤顺序。
                    if expr != image_page_ready_expression() {
                        log2.lock().unwrap().push(kind.clone());
                    }
                    let reply =
                        default_generation_toolbar_reply(&kind, (handler.lock().unwrap())(expr));
                    for ev in reply.events {
                        if ws.send(Message::Text(ev.to_string().into())).await.is_err() {
                            return;
                        }
                    }
                    match reply.exception {
                        Some(msg) => json!({ "exceptionDetails": {
                            "text": "Uncaught",
                            "exception": { "description": format!("Error: {msg}\n    at <anonymous>:1:1") }
                        }}),
                        None => json!({ "result": { "type": "object", "value": reply.value } }),
                    }
                } else {
                    log2.lock()
                        .unwrap()
                        .push(format!("{method} {}", req["params"]));
                    match method.as_str() {
                        "DOM.getDocument" => json!({ "root": { "nodeId": 1 } }),
                        "DOM.querySelector" => json!({ "nodeId": 7 }),
                        "Browser.getWindowForTarget" => json!({ "windowId": 92 }),
                        _ => json!({}),
                    }
                };
                let resp = json!({ "id": req["id"], "result": result });
                if ws
                    .send(Message::Text(resp.to_string().into()))
                    .await
                    .is_err()
                {
                    break;
                }
            }
        });
        let session = CdpSession::connect(&format!("ws://127.0.0.1:{port}/devtools/page/X"), "X")
            .await
            .unwrap();
        (session, log)
    }

    fn fast() -> SubmitTiming {
        SubmitTiming {
            prepare_timeout: Duration::from_millis(500),
            step_timeout: Duration::from_millis(500),
            skill_exit_timeout: Duration::from_millis(10),
            identity_window: Duration::from_millis(400),
            duration: duration::DurationTiming {
                step_timeout: Duration::from_millis(500),
                verify_timeout: Duration::from_millis(500),
                after_endpoint_key: Duration::from_millis(1),
                after_fine_tune_key: Duration::from_millis(1),
                drag_step: Duration::from_millis(1),
                after_drag: Duration::from_millis(1),
            },
            upload: upload::UploadTiming {
                step_timeout: Duration::from_millis(500),
                ready_timeout: Duration::from_millis(500),
            },
        }
    }

    fn request_event(url: &str, post_data: &str, wall_time_ms: i64) -> Value {
        json!({ "method": "Network.requestWillBeSent", "params": {
            "requestId": "req-1",
            "wallTime": wall_time_ms as f64 / 1000.0,
            "request": { "url": url, "method": "POST", "postData": post_data, "hasPostData": true }
        }})
    }

    fn identity_value() -> Value {
        json!({ "localMessageId": "lm-1", "messageId": null, "conversationId": "12345678901234567" })
    }

    /// 默认 handler:一切顺利,`send` 时推一条生成请求事件。
    fn happy_handler() -> Handler {
        Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "exit_skill" => ok(json!({ "found": false, "exited": false, "skillBar": true })),
                "prepare" => {
                    ok(json!({ "mode": "image", "model": "Seedream 5.0 Pro", "ratio": "1:1" }))
                }
                "fill" => ok(json!({ "filled": true })),
                "checkSend" => ok(json!({ "ready": true })),
                "send" => MockReply {
                    value: json!({ "sent": true }),
                    exception: None,
                    events: vec![
                        json!({ "method": "Network.loadingFinished", "params": { "requestId": "x" } }),
                        request_event(
                            "https://www.doubao.com/samantha/chat/history",
                            "{}",
                            now_ms(),
                        ),
                        request_event(
                            "https://www.doubao.com/samantha/chat/completion",
                            r#"{"messages":[{"content":"{\"text\":\"一只蓝色小鸟\"}","local_message_id":"lm-1"}],"conversation_id":12345678901234567}"#,
                            now_ms(),
                        ),
                    ],
                },
                "identity" => {
                    assert!(
                        expr.contains(r#"\"conversation_id\":\"12345678901234567\""#),
                        "大整数应已修复: {expr}"
                    );
                    ok(identity_value())
                }
                _ => ok(Value::Null),
            }
        })))
    }

    type StageLog = Arc<Mutex<Vec<(TaskStage, Option<String>)>>>;

    fn stage_log() -> (StageLog, impl Fn(TaskStage, Option<String>) + Send + Sync) {
        let stages = Arc::new(Mutex::new(Vec::new()));
        let s2 = Arc::clone(&stages);
        (stages, move |stage, msg| {
            s2.lock().unwrap().push((stage, msg))
        })
    }

    #[tokio::test]
    async fn prepare_uses_native_generation_mode_selection_on_current_page() {
        let h: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            match classify(expr).as_str() {
                "exit_skill" => ok(json!({
                    "found": false,
                    "exited": false,
                    "skillBar": false,
                    "conversationReady": true
                })),
                "wait_skill_bar" => ok(json!({
                    "skillBar": false,
                    "conversationReady": true,
                    "waitedMs": 0
                })),
                "generation_mode" => ok(json!({ "ready": true, "supported": true })),
                "image_settings" => ok(json!({ "ready": true, "supported": true })),
                "prepare" => ok(json!({
                    "mode": "image",
                    "model": "Seedream 5.0 Pro",
                    "ratio": "1:1"
                })),
                "fill" => ok(json!({ "filled": true })),
                "checkSend" => ok(json!({ "ready": true })),
                _ => ok(Value::Null),
            }
        })));
        let (session, log) = spawn_mock_page(h).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_stages, on_stage) = stage_log();

        let result = prepare_and_fill(
            &session,
            &r,
            &plan,
            &[],
            &fast(),
            &CancellationToken::new(),
            &on_stage,
        )
        .await;

        assert!(result.is_ok(), "{result:?}");
        let log = log.lock().unwrap().clone();
        assert!(!log.iter().any(|entry| entry == "wait_skill_bar"));
        assert!(!log.iter().any(|entry| entry == "generation_toolbar"));
        let head: Vec<&str> = log.iter().take(7).map(String::as_str).collect();
        assert_eq!(
            head,
            [
                "exit_skill",
                "generation_mode",
                "image_settings",
                "bootstrap",
                "attachments",
                "prepare",
                "fill"
            ]
        );
    }

    #[tokio::test]
    async fn native_generation_mode_waits_while_the_input_engine_remounts() {
        let probes = Arc::new(Mutex::new(0usize));
        let probes2 = Arc::clone(&probes);
        let handler: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            if classify(expr) != "generation_mode" {
                return ok(Value::Null);
            }
            let mut count = probes2.lock().unwrap();
            *count += 1;
            match *count {
                1 => ok(json!({
                    "ready": false,
                    "supported": true,
                    "step": "conversation_button",
                    "point": { "x": 120.0, "y": 80.0 }
                })),
                2 => ok(json!({
                    "ready": false,
                    "supported": false,
                    "reason": "input_engine_missing"
                })),
                _ => ok(json!({ "ready": true, "supported": true })),
            }
        })));
        let (session, _log) = spawn_mock_page(handler).await;
        let timing = SubmitTiming {
            step_timeout: Duration::from_secs(1),
            ..fast()
        };

        let result = select_native_generation_mode(
            &session,
            TaskType::Image,
            &timing,
            &CancellationToken::new(),
        )
        .await;

        assert!(result.is_ok(), "{result:?}");
        assert_eq!(*probes.lock().unwrap(), 3);
        session.close().await;
    }

    #[tokio::test]
    async fn native_generation_mode_waits_for_the_skill_menu_animation() {
        let probes = Arc::new(Mutex::new(0usize));
        let probes2 = Arc::clone(&probes);
        let handler: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            if classify(expr) != "generation_mode" {
                return ok(Value::Null);
            }
            let mut count = probes2.lock().unwrap();
            *count += 1;
            match *count {
                1 => ok(json!({
                    "ready": false,
                    "supported": true,
                    "waiting": true,
                    "reason": "target_animating"
                })),
                2 => ok(json!({
                    "ready": false,
                    "supported": true,
                    "step": "target",
                    "point": { "x": 120.0, "y": 80.0 }
                })),
                _ => ok(json!({ "ready": true, "supported": true })),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let timing = SubmitTiming {
            step_timeout: Duration::from_secs(1),
            ..fast()
        };

        let result = select_native_generation_mode(
            &session,
            TaskType::Image,
            &timing,
            &CancellationToken::new(),
        )
        .await;

        assert!(result.is_ok(), "{result:?}");
        assert_eq!(*probes.lock().unwrap(), 3);
        let click_events = log
            .lock()
            .unwrap()
            .iter()
            .filter(|entry| entry.starts_with("Input.dispatchMouseEvent"))
            .count();
        assert_eq!(click_events, 3, "菜单动画期间不得提前点击");
        session.close().await;
    }

    #[tokio::test]
    async fn native_generation_mode_waits_when_the_input_engine_is_initially_remounting() {
        let probes = Arc::new(Mutex::new(0usize));
        let probes2 = Arc::clone(&probes);
        let handler: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            if classify(expr) != "generation_mode" {
                return ok(Value::Null);
            }
            let mut count = probes2.lock().unwrap();
            *count += 1;
            match *count {
                1 | 2 => ok(json!({
                    "ready": false,
                    "supported": false,
                    "reason": "input_engine_missing"
                })),
                _ => ok(json!({ "ready": true, "supported": true })),
            }
        })));
        let (session, _log) = spawn_mock_page(handler).await;
        let timing = SubmitTiming {
            step_timeout: Duration::from_secs(1),
            ..fast()
        };

        let result = select_native_generation_mode(
            &session,
            TaskType::Image,
            &timing,
            &CancellationToken::new(),
        )
        .await;

        assert!(result.is_ok(), "{result:?}");
        assert_eq!(*probes.lock().unwrap(), 3);
        session.close().await;
    }

    #[tokio::test]
    async fn native_generation_mode_clicks_fresh_chat_only_once_before_selecting_target() {
        let probes = Arc::new(Mutex::new(0usize));
        let probes2 = Arc::clone(&probes);
        let handler: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            if classify(expr) != "generation_mode" {
                return ok(Value::Null);
            }
            let mut count = probes2.lock().unwrap();
            *count += 1;
            if *count == 1 || !expr.contains("const ALLOW_FRESH_CHAT = false;") {
                return ok(json!({
                    "ready": false,
                    "supported": true,
                    "step": "fresh_chat",
                    "point": { "x": 40.0, "y": 30.0 }
                }));
            }
            if *count == 2 {
                return ok(json!({
                    "ready": false,
                    "supported": true,
                    "step": "target",
                    "point": { "x": 120.0, "y": 80.0 }
                }));
            }
            ok(json!({ "ready": true, "supported": true }))
        })));
        let (session, _log) = spawn_mock_page(handler).await;
        let timing = SubmitTiming {
            step_timeout: Duration::from_millis(700),
            ..fast()
        };

        let result = select_native_generation_mode(
            &session,
            TaskType::Image,
            &timing,
            &CancellationToken::new(),
        )
        .await;

        assert!(result.is_ok(), "{result:?}");
        assert_eq!(*probes.lock().unwrap(), 3);
        session.close().await;
    }

    /// 豆包首次上传素材时,点"发送"后弹"安全确认"(`safety_authorization_dialog`):自动点确认,
    /// 确认之后豆包才发出生成请求,识别窗口从确认后重新计算。
    #[tokio::test]
    async fn safety_dialog_after_send_is_confirmed_and_identity_still_captured() {
        let confirmed = Arc::new(Mutex::new(false));
        let c2 = Arc::clone(&confirmed);
        let h: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            match classify(expr).as_str() {
                "exit_skill" => ok(json!({ "found": false, "exited": false, "skillBar": true })),
                "prepare" => {
                    ok(json!({ "mode": "image", "model": "Seedream 5.0 Pro", "ratio": "1:1" }))
                }
                "fill" => ok(json!({ "filled": true })),
                "checkSend" => ok(json!({ "ready": true })),
                "send" => ok(json!({ "sent": true })),
                "safety" => {
                    *c2.lock().unwrap() = true;
                    MockReply {
                        value: json!({ "found": true, "clicked": true, "closed": true }),
                        exception: None,
                        events: vec![request_event(
                            "https://www.doubao.com/samantha/chat/completion",
                            r#"{"messages":[{"content":"{\"text\":\"一只蓝色小鸟\"}","local_message_id":"lm-1"}],"conversation_id":12345678901234567}"#,
                            now_ms(),
                        )],
                    }
                }
                "identity" => ok(identity_value()),
                _ => ok(Value::Null),
            }
        })));
        let (session, log) = spawn_mock_page(h).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_stages, on_stage) = stage_log();
        let report = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap();
        assert!(*confirmed.lock().unwrap());
        assert_eq!(
            report.identity.unwrap().local_message_id.as_deref(),
            Some("lm-1")
        );
        let log = log.lock().unwrap().clone();
        let send = log.iter().position(|l| l.starts_with("send")).unwrap();
        let safety = log.iter().position(|l| l.starts_with("safety")).unwrap();
        assert!(safety > send, "{log:?}");
    }

    #[tokio::test]
    async fn submit_runs_steps_in_order_and_returns_identity() {
        let (session, log) = spawn_mock_page(happy_handler()).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let cancel = CancellationToken::new();
        let (stages, on_stage) = stage_log();
        let before = now_ms();
        let report = submit_with(&session, None, &r, &plan, &[], &cancel, &on_stage, &fast())
            .await
            .unwrap();
        let identity = report.identity.unwrap();
        assert_eq!(identity.local_message_id.as_deref(), Some("lm-1"));
        assert_eq!(identity.message_id, None);
        assert_eq!(
            identity.conversation_id.as_deref(),
            Some("12345678901234567")
        );
        assert!(report.send_started_at >= before);
        // 携带 identity 的生成请求交给 capture 预置(history 请求不算生成请求)
        assert_eq!(
            report.generation_requests,
            vec![ObservedRequest {
                target_id: session.target_id.clone(),
                request_id: "req-1".into(),
            }]
        );

        let log = log.lock().unwrap().clone();
        let evals: Vec<&str> = log
            .iter()
            .filter(|l| !l.starts_with("Network."))
            .map(String::as_str)
            .collect();
        assert_eq!(
            evals,
            [
                "exit_skill",
                "generation_mode",
                "image_settings",
                "bootstrap",
                "attachments",
                "prepare",
                "fill",
                "checkSend",
                "send",
                "safety",
                "Browser.getWindowForTarget {\"targetId\":\"X\"}",
                "Browser.setWindowBounds {\"bounds\":{\"windowState\":\"minimized\"},\"windowId\":92}",
                "reader",
                "identity"
            ]
        );
        assert!(
            !log.iter().any(|l| l == "Page.bringToFront {}"),
            "后台提交不得把豆包窗口拉到前台: {log:?}"
        );
        // send 之前已 Network.enable {maxPostDataSize}
        let enable_idx = log
            .iter()
            .position(|l| l.starts_with("Network.enable"))
            .expect("应发 Network.enable");
        let send_idx = log.iter().position(|l| l == "send").unwrap();
        assert!(enable_idx < send_idx);
        assert!(log[enable_idx].contains("\"maxPostDataSize\":16777216"));

        let stages = stages.lock().unwrap().clone();
        assert_eq!(stages.first().unwrap().0, TaskStage::Preparing);
        assert_eq!(stages.last().unwrap(), &(TaskStage::Sent, None));
        assert!(stages.iter().filter(|(s, _)| *s == TaskStage::Sent).count() == 1);
        session.close().await;
    }

    #[tokio::test]
    async fn submit_exits_residual_skill_before_prepare_and_fails_when_skill_bar_missing() {
        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "exit_skill" => {
                    ok(json!({ "found": true, "exited": false, "skillBar": false, "waitedMs": 10 }))
                }
                _ => ok(Value::Null),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        let d = err.details.unwrap();
        assert_eq!(d["reason"], "skill_exit_failed");
        assert_eq!(d["failed_selectors"], json!([SKILL_BAR_SELECTOR]));
        let log = log.lock().unwrap().clone();
        let sequence: Vec<&str> = log
            .iter()
            .filter(|line| !line.starts_with("Browser."))
            .map(String::as_str)
            .collect();
        assert_eq!(sequence, ["exit_skill"]);
        session.close().await;
    }

    #[tokio::test]
    async fn leftover_attachment_in_skill_draft_is_removed_and_prepare_retried_once() {
        let mut prepares = 0;
        let handler: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            match classify(expr).as_str() {
                "prepare" => {
                    prepares += 1;
                    if prepares == 1 {
                        throw("豆包输入框已有参考图片，请先移除后重试。")
                    } else {
                        ok(json!({ "mode": "image", "model": "Seedream 5.0 Pro", "ratio": "1:1" }))
                    }
                }
                "attachments" => ok(json!({ "before": 1, "after": 0 })),
                _ => (happy_handler().lock().unwrap())(expr),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap();
        let log = log.lock().unwrap().clone();
        let head: Vec<&str> = log
            .iter()
            .filter(|line| !line.starts_with("Browser."))
            .take(9)
            .map(String::as_str)
            .collect();
        assert_eq!(
            head,
            [
                "exit_skill",
                "generation_mode",
                "image_settings",
                "bootstrap",
                "attachments",
                "prepare",
                "attachments",
                "clear",
                "prepare"
            ]
        );
        session.close().await;

        // 重试仍失败 → 报第二次的异常,不再重试
        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "prepare" => throw("豆包输入框已有参考图片，请先移除后重试。"),
                _ => ok(json!({ "before": 0, "after": 0 })),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        assert_eq!(
            log.lock()
                .unwrap()
                .iter()
                .filter(|l| *l == "prepare")
                .count(),
            2
        );
        session.close().await;
    }

    #[tokio::test]
    async fn transient_duplicate_controls_retry_prepare_once() {
        let mut prepares = 0;
        let handler: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            match classify(expr).as_str() {
                "prepare" => {
                    prepares += 1;
                    if prepares == 1 {
                        throw("豆包控件未唯一匹配，已停止发送。")
                    } else {
                        ok(json!({
                            "mode": "video",
                            "model": "Seedance 2.5",
                            "ratio": "16:9",
                            "duration": 10
                        }))
                    }
                }
                _ => (happy_handler().lock().unwrap())(expr),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = video_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_stages, on_stage) = stage_log();

        submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap();

        assert_eq!(
            log.lock()
                .unwrap()
                .iter()
                .filter(|line| *line == "prepare")
                .count(),
            2
        );
        session.close().await;
    }

    #[tokio::test]
    async fn transient_duplicate_controls_keep_retrying_until_the_ui_settles() {
        let mut prepares = 0;
        let handler: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            match classify(expr).as_str() {
                "prepare" => {
                    prepares += 1;
                    if prepares < 3 {
                        throw("豆包控件未唯一匹配，已停止发送。")
                    } else {
                        ok(json!({
                            "mode": "video",
                            "model": "Seedance 2.5",
                            "ratio": "16:9",
                            "duration": 10
                        }))
                    }
                }
                _ => (happy_handler().lock().unwrap())(expr),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = video_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_stages, on_stage) = stage_log();

        submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap();

        assert_eq!(
            log.lock()
                .unwrap()
                .iter()
                .filter(|line| *line == "prepare")
                .count(),
            3
        );
        session.close().await;
    }

    #[tokio::test]
    async fn missing_image_toolbar_reopens_mode_and_retries_prepare_once() {
        let mut prepares = 0;
        let handler: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            match classify(expr).as_str() {
                "generation_mode" => ok(json!({ "ready": true, "supported": true })),
                "prepare" => {
                    prepares += 1;
                    if prepares == 1 {
                        throw(
                            r#"未找到所选生成比例。；新版图片比例诊断：{"model":"","settings":[],"expectedRatio":"16:9","reason":"图片比例控件未唯一匹配"}"#,
                        )
                    } else {
                        ok(json!({
                            "mode": "image",
                            "model": "Seedream 4.5",
                            "ratio": "1:1"
                        }))
                    }
                }
                _ => (happy_handler().lock().unwrap())(expr),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_stages, on_stage) = stage_log();

        submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap();

        {
            let log = log.lock().unwrap();
            assert_eq!(
                log.iter().filter(|line| *line == "generation_mode").count(),
                2
            );
            assert_eq!(log.iter().filter(|line| *line == "prepare").count(), 2);
        }
        session.close().await;
    }

    #[test]
    fn leftover_exception_covers_image_and_video_skill_messages() {
        assert!(is_leftover_exception(
            "Error: 豆包输入框已有参考图片，请先移除后重试。\n    at <anonymous>:1:1"
        ));
        assert!(is_leftover_exception(
            "Error: 视频输入区已有内容，任务未发送"
        ));
        assert!(!is_leftover_exception("Error: 新图片对话未准备好"));
        assert!(!is_leftover_exception("Error: 视频模型不可用"));
    }

    #[tokio::test]
    async fn leftover_in_video_skill_draft_is_cleaned_and_prepare_retried_once() {
        // 真机(Task 3.2 复审):视频草稿里残留参考图/参考视频卡片时,prepare 抛
        // `视频输入区已有内容，任务未发送`;要在技能里删卡片 + 清输入框后重试。
        let mut prepares = 0;
        let handler: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            match classify(expr).as_str() {
                "prepare" => {
                    prepares += 1;
                    if prepares == 1 {
                        throw("视频输入区已有内容，任务未发送")
                    } else {
                        ok(
                            json!({ "mode": "video", "model": "Seedance 2.5", "ratio": "16:9",
                                   "duration": 10, "needsDurationInput": false }),
                        )
                    }
                }
                "attachments" => ok(json!({ "before": 1, "after": 0 })),
                _ => (happy_handler().lock().unwrap())(expr),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = video_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap();
        let log = log.lock().unwrap().clone();
        let head: Vec<&str> = log
            .iter()
            .filter(|line| !line.starts_with("Browser."))
            .take(9)
            .map(String::as_str)
            .collect();
        assert_eq!(
            head,
            [
                "exit_skill",
                "generation_mode",
                "bootstrap",
                "attachments",
                "prepare",
                "attachments",
                "clear",
                "prepare",
                "fill"
            ]
        );
        assert_eq!(log.iter().filter(|l| *l == "prepare").count(), 2);
        session.close().await;
    }

    #[tokio::test]
    async fn upload_failure_removes_uploaded_attachments() {
        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "uploadReady" => throw("参考图尚未全部上传完成，未发送，请检查豆包上传状态。"),
                "attachments" => ok(json!({ "before": 0, "after": 0 })),
                _ => (happy_handler().lock().unwrap())(expr),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[PathBuf::from("/tmp/ref-1.png")],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        assert_eq!(err.details.unwrap()["reason"], "uploadReady");
        let log = log.lock().unwrap().clone();
        assert_eq!(log.last().unwrap(), "attachments", "{log:?}");
        assert!(!log.iter().any(|l| l == "fill"));
        session.close().await;
    }

    #[tokio::test]
    async fn prepare_exception_maps_to_preparing_failed_with_js_exception() {
        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "prepare" => throw("未找到所选生成比例。"),
                _ => ok(json!({ "found": false })),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        assert_eq!(err.details.unwrap()["js_exception"], "未找到所选生成比例。");
        let log = log.lock().unwrap().clone();
        let sequence: Vec<&str> = log
            .iter()
            .filter(|line| !line.starts_with("Browser."))
            .map(String::as_str)
            .collect();
        assert_eq!(
            sequence,
            [
                "exit_skill",
                "generation_mode",
                "image_settings",
                "bootstrap",
                "attachments",
                "prepare"
            ]
        );
        session.close().await;
    }

    #[tokio::test]
    async fn fill_exception_is_preparing_failed_check_send_exception_is_submit_failed_and_clears_input(
    ) {
        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "fill" => throw("豆包输入框已变化，未覆盖已有内容。"),
                _ => ok(json!({ "found": false })),
            }
        })));
        let (session, _log) = spawn_mock_page(handler).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        assert_eq!(
            err.details.unwrap()["js_exception"],
            "豆包输入框已变化，未覆盖已有内容。"
        );
        session.close().await;

        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "checkSend" => throw("豆包发送按钮尚未就绪，未发送。"),
                _ => ok(json!({ "found": false })),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_submit_failed");
        assert_eq!(
            err.details.unwrap()["js_exception"],
            "豆包发送按钮尚未就绪，未发送。"
        );
        let log = log.lock().unwrap().clone();
        let sequence: Vec<&str> = log
            .iter()
            .filter(|line| !line.starts_with("Browser."))
            .map(String::as_str)
            .collect();
        assert_eq!(
            sequence,
            [
                "exit_skill",
                "generation_mode",
                "image_settings",
                "bootstrap",
                "attachments",
                "prepare",
                "fill",
                "checkSend",
                "clear"
            ]
        );
        session.close().await;
    }

    #[tokio::test]
    async fn send_exception_is_submit_failed_and_no_identity_is_identity_missing() {
        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "send" => throw("豆包模式、模型、比例或提示词已变化，未发送。"),
                _ => ok(json!({ "found": false })),
            }
        })));
        let (session, _log) = spawn_mock_page(handler).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_submit_failed");
        session.close().await;

        // send 成功但窗口内没有任何生成请求 → identity_missing(details.sent=true)
        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "send" => MockReply {
                    value: json!({}),
                    exception: None,
                    // 只有历史请求,不算生成请求
                    events: vec![request_event(
                        "https://www.doubao.com/im/chain/single",
                        "{}",
                        now_ms(),
                    )],
                },
                _ => ok(json!({ "found": false })),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let (stages, on_stage) = stage_log();
        let started = Instant::now();
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_submit_identity_missing");
        let d = err.details.unwrap();
        assert_eq!(d["sent"], true);
        assert_eq!(d["candidates"], 0);
        assert!(d["send_started_at"].as_i64().unwrap() > 0);
        assert!(started.elapsed() >= Duration::from_millis(400));
        assert!(started.elapsed() < Duration::from_secs(3));
        assert!(!log.lock().unwrap().iter().any(|l| l == "identity"));
        assert_eq!(
            stages.lock().unwrap().last().unwrap(),
            &(TaskStage::Sent, None)
        );
        session.close().await;
    }

    #[tokio::test]
    async fn identity_null_from_reader_keeps_waiting_then_identity_missing() {
        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "send" => MockReply {
                    value: json!({}),
                    exception: None,
                    events: vec![request_event(
                        "https://www.doubao.com/samantha/chat/completion",
                        r#"{"text":"别的内容"}"#,
                        now_ms(),
                    )],
                },
                "identity" => ok(Value::Null),
                _ => ok(json!({ "found": false })),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_submit_identity_missing");
        assert_eq!(err.details.unwrap()["candidates"], 1);
        let log = log.lock().unwrap().clone();
        assert_eq!(log.iter().filter(|l| *l == "reader").count(), 1);
        assert_eq!(log.iter().filter(|l| *l == "identity").count(), 1);
        session.close().await;
    }

    #[tokio::test]
    async fn stale_requests_before_send_are_ignored() {
        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "send" => MockReply {
                    value: json!({}),
                    exception: None,
                    events: vec![request_event(
                        "https://www.doubao.com/samantha/chat/completion",
                        "{}",
                        now_ms() - 10_000,
                    )],
                },
                _ => ok(json!({ "found": false })),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_submit_identity_missing");
        assert_eq!(err.details.unwrap()["candidates"], 0);
        assert!(!log.lock().unwrap().iter().any(|l| l == "identity"));
        session.close().await;
    }

    #[tokio::test]
    async fn identity_from_background_session_is_evaluated_there() {
        // chat 页 send 后不推事件;background 页在收到 Network.enable 后推一条生成请求事件
        let chat_handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "identity" => panic!("identity 不应在 chat 页评估"),
                _ => ok(json!({ "found": false })),
            }
        })));
        let (chat, chat_log) = spawn_mock_page(chat_handler).await;

        use futures_util::{SinkExt, StreamExt};
        use tokio_tungstenite::tungstenite::Message;
        let bg_log: Log = Arc::new(Mutex::new(Vec::new()));
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        let bg_log2 = Arc::clone(&bg_log);
        tokio::spawn(async move {
            let (sock, _) = listener.accept().await.unwrap();
            let mut ws = tokio_tungstenite::accept_async(sock).await.unwrap();
            while let Some(Ok(Message::Text(txt))) = ws.next().await {
                let req: Value = serde_json::from_str(&txt).unwrap();
                let method = req["method"].as_str().unwrap_or("").to_string();
                let result = if method == "Runtime.evaluate" {
                    let kind = classify(req["params"]["expression"].as_str().unwrap_or(""));
                    bg_log2.lock().unwrap().push(kind.clone());
                    let value = if kind == "identity" {
                        json!({ "localMessageId": null, "messageId": "m9", "conversationId": null })
                    } else {
                        Value::Null
                    };
                    json!({ "result": { "type": "object", "value": value } })
                } else {
                    bg_log2.lock().unwrap().push(method.clone());
                    if method == "Network.enable" {
                        // 事件在 send 之后才到(延迟推送),模拟后台页看到的生成请求
                        let ev = request_event(
                            "https://api.snssdk.com/x/generate",
                            r#"{"message_id":"m9","text":"一只蓝色小鸟"}"#,
                            now_ms() + 50,
                        );
                        let mut ws_ev = ws;
                        tokio::time::sleep(Duration::from_millis(100)).await;
                        ws_ev
                            .send(Message::Text(ev.to_string().into()))
                            .await
                            .unwrap();
                        ws = ws_ev;
                    }
                    json!({})
                };
                let resp = json!({ "id": req["id"], "result": result });
                if ws
                    .send(Message::Text(resp.to_string().into()))
                    .await
                    .is_err()
                {
                    break;
                }
            }
        });
        let bg = CdpSession::connect(&format!("ws://127.0.0.1:{port}/devtools/page/BG"), "BG")
            .await
            .unwrap();

        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        let report = submit_with(
            &chat,
            Some(&bg),
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap();
        let id = report.identity.unwrap();
        assert_eq!(id.local_message_id, None);
        assert_eq!(id.message_id.as_deref(), Some("m9"));
        let bg_log = bg_log.lock().unwrap().clone();
        assert!(bg_log.iter().any(|l| l == "Network.enable"));
        assert_eq!(bg_log.iter().filter(|l| *l == "reader").count(), 1);
        assert_eq!(bg_log.iter().filter(|l| *l == "identity").count(), 1);
        assert!(!chat_log
            .lock()
            .unwrap()
            .iter()
            .any(|l| l == "identity" || l == "reader"));
        chat.close().await;
        bg.close().await;
    }

    #[tokio::test]
    async fn video_needs_duration_input_sets_duration_and_direct_shape_proceeds() {
        let send_event = || {
            request_event(
                "https://www.doubao.com/samantha/chat/completion",
                r#"{"local_message_id":"lm-v","text":"一只蓝色小鸟"}"#,
                now_ms(),
            )
        };
        let handler: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            match classify(expr).as_str() {
                "prepare" => ok(json!({ "needsDurationInput": true })),
                "durationDragPlan" => ok(json!({ "matched": true, "duration": 10 })),
                "verifyVideoSettings" => ok(
                    json!({ "mode": "video", "model": "Seedance 2.5", "ratio": "16:9", "duration": 10 }),
                ),
                "send" => MockReply {
                    value: json!({}),
                    exception: None,
                    events: vec![send_event()],
                },
                "identity" => ok(json!({ "localMessageId": "lm-v" })),
                _ => ok(json!({ "found": false })),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = video_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (stages, on_stage) = stage_log();
        let report = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap();
        assert!(report.identity.is_some());
        let evals: Vec<String> = log
            .lock()
            .unwrap()
            .iter()
            .filter(|l| !l.starts_with("Network.") && !l.starts_with("other"))
            .cloned()
            .collect();
        let pos = |k: &str| evals.iter().position(|l| l == k).unwrap();
        assert!(pos("prepare") < pos("durationDragPlan"));
        assert!(pos("durationDragPlan") < pos("verifyVideoSettings"));
        assert!(pos("verifyVideoSettings") < pos("fill"));
        assert!(stages
            .lock()
            .unwrap()
            .iter()
            .any(|(_, m)| m.as_deref() == Some("设置视频时长 10 秒")));
        session.close().await;

        // 时长调不到 → duration_not_applied,不 fill
        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "prepare" => ok(json!({ "needsDurationInput": true })),
                "durationDragPlan" => ok(json!({ "matched": false, "duration": 6 })),
                _ => ok(json!({ "found": false })),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        assert_eq!(err.details.unwrap()["reason"], "duration_not_applied");
        assert!(!log.lock().unwrap().iter().any(|l| l == "fill"));
        session.close().await;

        // 摘要已匹配的直接形状 → 继续 fill;时长不符 → duration_mismatch
        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "prepare" => ok(
                    json!({ "mode": "video", "model": "Seedance 2.5", "ratio": "16:9", "duration": 10 }),
                ),
                "send" => MockReply {
                    value: json!({}),
                    exception: None,
                    events: vec![request_event(
                        "https://www.doubao.com/samantha/chat/completion",
                        r#"{"local_message_id":"lm-v","text":"一只蓝色小鸟"}"#,
                        now_ms(),
                    )],
                },
                "identity" => ok(json!({ "localMessageId": "lm-v" })),
                _ => ok(json!({ "found": false })),
            }
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let report = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap();
        assert_eq!(
            report.identity.unwrap().local_message_id.as_deref(),
            Some("lm-v")
        );
        assert!(log.lock().unwrap().iter().any(|l| l == "fill"));
        session.close().await;

        let handler: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            match classify(expr).as_str() {
                "prepare" => ok(
                    json!({ "mode": "video", "model": "Seedance 2.5", "ratio": "16:9", "duration": 12 }),
                ),
                _ => ok(json!({ "found": false })),
            }
        })));
        let (session, _log) = spawn_mock_page(handler).await;
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        let d = err.details.unwrap();
        assert_eq!(d["reason"], "duration_mismatch");
        assert_eq!(d["expected"], 10);
        assert_eq!(d["actual"], 12);
        session.close().await;
    }

    #[tokio::test]
    async fn reference_files_are_uploaded_before_fill_and_send_carries_uploaded_count() {
        let inner = happy_handler();
        let sent: Arc<Mutex<Vec<(String, String)>>> = Arc::new(Mutex::new(Vec::new()));
        let sent2 = Arc::clone(&sent);
        let handler: Handler = Arc::new(Mutex::new(Box::new(move |expr: &str| {
            let kind = classify(expr);
            if matches!(kind.as_str(), "fill" | "checkSend" | "send" | "uploadReady") {
                sent2.lock().unwrap().push((kind, expr.to_string()));
            }
            (inner.lock().unwrap())(expr)
        })));
        let (session, log) = spawn_mock_page(handler).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (stages, on_stage) = stage_log();
        let refs = [
            PathBuf::from("/tmp/ref-1.png"),
            PathBuf::from("/tmp/ref-2.png"),
        ];
        let report = submit_with(
            &session,
            None,
            &r,
            &plan,
            &refs,
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap();
        assert!(report.identity.is_some());
        let log = log.lock().unwrap().clone();
        let seq: Vec<&str> = log
            .iter()
            .filter(|l| {
                !l.starts_with("Network.")
                    && !l.starts_with("Page.")
                    && !l.starts_with("DOM.getDocument")
                    && !l.starts_with("DOM.querySelector")
            })
            .map(|l| l.split(' ').next().unwrap())
            .collect();
        assert_eq!(
            seq,
            [
                "exit_skill",
                "generation_mode",
                "image_settings",
                "bootstrap",
                "attachments",
                "prepare",
                "open_upload",
                "fileInput",
                "DOM.setFileInputFiles",
                "uploadReady",
                "open_upload",
                "fileInput",
                "DOM.setFileInputFiles",
                "uploadReady",
                "fill",
                "checkSend",
                "send",
                "safety",
                "Browser.getWindowForTarget",
                "Browser.setWindowBounds",
                "reader",
                "identity"
            ]
        );
        assert!(log
            .iter()
            .any(|l| l.starts_with("DOM.setFileInputFiles") && l.contains("/tmp/ref-2.png")));
        // fill / checkSend / send 都带 uploadedCount=2 与两张参考图
        for (kind, expr) in sent
            .lock()
            .unwrap()
            .iter()
            .filter(|(k, _)| k != "uploadReady")
        {
            assert!(expr.contains("\"uploadedCount\":2"), "{kind}: {expr}");
            assert!(expr.contains("/tmp/ref-1.png"), "{kind}");
        }
        let msgs: Vec<String> = stages
            .lock()
            .unwrap()
            .iter()
            .filter_map(|(_, m)| m.clone())
            .collect();
        assert!(msgs.contains(&"上传参考图 0/2".to_string()));
        assert!(msgs.contains(&"上传参考图 2/2".to_string()));
        session.close().await;
    }

    #[tokio::test]
    async fn cancel_before_start_and_during_identity_wait() {
        let (session, log) = spawn_mock_page(happy_handler()).await;
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        let cancel = CancellationToken::new();
        cancel.cancel();
        let err = submit_with(&session, None, &r, &plan, &[], &cancel, &on_stage, &fast())
            .await
            .unwrap_err();
        assert_eq!(err.code, "generation_cancelled");
        assert!(log.lock().unwrap().is_empty());
        session.close().await;

        // send 后没有事件,identity 窗口 5s;50ms 后取消 → 立刻返回 cancelled
        let handler: Handler = Arc::new(Mutex::new(Box::new(|_expr: &str| {
            ok(json!({ "found": false }))
        })));
        let (session, _log) = spawn_mock_page(handler).await;
        let cancel = CancellationToken::new();
        let timing = SubmitTiming {
            identity_window: Duration::from_secs(5),
            ..fast()
        };
        let c2 = cancel.clone();
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(80)).await;
            c2.cancel();
        });
        let started = Instant::now();
        let err = submit_with(&session, None, &r, &plan, &[], &cancel, &on_stage, &timing)
            .await
            .unwrap_err();
        assert_eq!(err.code, "generation_cancelled");
        assert_eq!(err.details.unwrap()["step"], "identity");
        assert!(
            started.elapsed() < Duration::from_secs(2),
            "{:?}",
            started.elapsed()
        );
        session.close().await;
    }

    #[tokio::test]
    async fn prepare_timeout_maps_to_preparing_timeout() {
        // mock:prepare 不回应 → send_timeout 超时
        use futures_util::{SinkExt, StreamExt};
        use tokio_tungstenite::tungstenite::Message;
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        tokio::spawn(async move {
            let (sock, _) = listener.accept().await.unwrap();
            let mut ws = tokio_tungstenite::accept_async(sock).await.unwrap();
            while let Some(Ok(Message::Text(txt))) = ws.next().await {
                let req: Value = serde_json::from_str(&txt).unwrap();
                let expr = req["params"]["expression"].as_str().unwrap_or("");
                let kind = classify(expr);
                if kind == "prepare" {
                    continue;
                }
                let value = if kind == "generation_toolbar" {
                    json!({ "opened": true })
                } else {
                    json!({ "found": false })
                };
                let resp = json!({ "id": req["id"], "result": { "result": { "type": "object", "value": value } } });
                if ws
                    .send(Message::Text(resp.to_string().into()))
                    .await
                    .is_err()
                {
                    break;
                }
            }
        });
        let session = CdpSession::connect(&format!("ws://127.0.0.1:{port}/devtools/page/X"), "X")
            .await
            .unwrap();
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        let timing = SubmitTiming {
            prepare_timeout: Duration::from_millis(150),
            ..fast()
        };
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &timing,
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_timeout");
        assert_eq!(err.details.unwrap()["step"], "prepare");
        session.close().await;
    }

    #[tokio::test]
    async fn closed_session_maps_to_cdp_disconnected() {
        use futures_util::StreamExt;
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        tokio::spawn(async move {
            let (sock, _) = listener.accept().await.unwrap();
            let mut ws = tokio_tungstenite::accept_async(sock).await.unwrap();
            let _ = ws.next().await;
            let _ = ws.close(None).await;
        });
        let session = CdpSession::connect(&format!("ws://127.0.0.1:{port}/devtools/page/X"), "X")
            .await
            .unwrap();
        let r = image_request();
        let plan = SubmissionPlan::for_request(&r, TASK_ID);
        let (_s, on_stage) = stage_log();
        let err = submit_with(
            &session,
            None,
            &r,
            &plan,
            &[],
            &CancellationToken::new(),
            &on_stage,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "doubao_cdp_disconnected");
    }

    /// 需要真豆包(端口 9707):`cargo test live_submit_dry_run -- --ignored --nocapture`。
    /// 跑到 `checkSend` 为止(退出残留技能 → 注入 → prepare → fill → checkSend),**不执行 send**,
    /// 不消耗配额;结束后 selectAll+delete 清空输入框。默认使用全站豆包图片通道的
    /// `Seedream 4.5 / 自动`;可用 `DOUBAO_LIVE_IMAGE_MODEL`、`DOUBAO_LIVE_IMAGE_RATIO`
    /// 显式验收已授权的其他模型和比例。
    #[tokio::test]
    #[ignore]
    async fn live_submit_dry_run() {
        let r = GenerateRequest {
            model_name: std::env::var("DOUBAO_LIVE_IMAGE_MODEL")
                .unwrap_or_else(|_| "Seedream 4.5".into()),
            ratio: std::env::var("DOUBAO_LIVE_IMAGE_RATIO").unwrap_or_else(|_| "自动".into()),
            prompt: "测试提示词(干跑,不发送)".into(),
            ..image_request()
        };
        live_dry_run(r, Vec::new()).await;
    }

    /// 当前豆包图片工具栏的方图比例回归检查：只跑到 `checkSend`，不会点击发送、不会消耗额度。
    /// 用于防止桌面端把 `1:1` 以非按钮控件或带附加文案的形式渲染后，Relay 无法选择比例。
    #[tokio::test]
    #[ignore]
    async fn live_image_square_ratio_dry_run() {
        live_dry_run(
            GenerateRequest {
                model_name: "Seedream 4.5".into(),
                ratio: "1:1".into(),
                prompt: "测试提示词(1:1干跑,不发送)".into(),
                ..image_request()
            },
            Vec::new(),
        )
        .await;
    }

    /// 只读取当前豆包“比例”下拉菜单的实际文本和元素类型，不填写提示词、不发送任务。
    #[tokio::test]
    #[ignore]
    async fn live_image_ratio_menu_probe() {
        use crate::cdp::{list_targets, select_pages};

        let targets = list_targets("127.0.0.1", 9707).await.unwrap();
        let (chat, _) = select_pages(&targets, 9707);
        let chat = chat.first().expect("需要 doubao-chat page");
        let session = CdpSession::connect(&chat.ws_url, &chat.id).await.unwrap();
        let cancel = CancellationToken::new();
        exit_residual_skill(&session, &SubmitTiming::default(), &cancel)
            .await
            .expect("退出残留技能失败");
        select_native_generation_mode(&session, TaskType::Image, &SubmitTiming::default(), &cancel)
            .await
            .expect("打开图片工具栏失败");
        let menu = session
            .evaluate(
                r#"(async () => {
  const visible = el => { const r = el?.getBoundingClientRect?.(); const s = el && getComputedStyle(el); return Boolean(r && r.width > 0 && r.height > 0 && s?.display !== 'none' && s?.visibility !== 'hidden'); };
  const label = el => String(el?.getAttribute?.('aria-label') || el?.innerText || el?.textContent || '').replace(/\s+/g, ' ').trim();
  const root = document.querySelector('#input-engine-container,[data-testid="input-engine-container"],[data-automation-id="input-engine-container"]');
  const ratioElements = [...root.querySelectorAll('*')].filter(visible).filter(el => /比例\s*(自动|\d+:\d+)/.test(label(el)) && label(el).length < 80);
  const describe = el => ({ tag: el.tagName.toLowerCase(), role: el.getAttribute('role'), testId: el.getAttribute('data-testid'), className: String(el.className || '').slice(0, 180), label: label(el), html: el.outerHTML.slice(0, 900) });
  const buttons = [...root.querySelectorAll('button')].filter(visible);
  const menuTrigger = root.querySelector('button[aria-haspopup="menu"]');
  const ratioButton = menuTrigger?.querySelector('button') || menuTrigger;
  if (!ratioButton) return { error: 'image_settings_button_missing', matches: ratioElements.map(describe), buttons: buttons.map(describe) };
  const rect = ratioButton.getBoundingClientRect();
  return { control: describe(ratioButton), rect: { x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 }, matches: ratioElements.map(describe), buttons: buttons.map(describe) };
})()"#,
            )
            .await
            .expect("读取比例菜单失败");
        println!("image ratio control probe: {menu:?}");
        let x = menu["rect"]["x"].as_f64().expect("比例按钮 x 坐标缺失");
        let y = menu["rect"]["y"].as_f64().expect("比例按钮 y 坐标缺失");
        for (kind, buttons) in [("mouseMoved", 0), ("mousePressed", 1), ("mouseReleased", 0)] {
            let mut params = json!({ "type": kind, "x": x, "y": y });
            if kind != "mouseMoved" {
                params["button"] = json!("left");
                params["buttons"] = json!(buttons);
                params["clickCount"] = json!(1);
            }
            session
                .send("Input.dispatchMouseEvent", params)
                .await
                .expect("可信点击图片设置失败");
        }
        tokio::time::sleep(Duration::from_millis(300)).await;
        let options = session.evaluate(r#"(() => { const v=e=>{const r=e?.getBoundingClientRect?.(),s=e&&getComputedStyle(e);return !!(r&&r.width>0&&r.height>0&&s?.display!== 'none'&&s?.visibility!== 'hidden')};const t=e=>String(e?.getAttribute?.('aria-label')||e?.innerText||e?.textContent||'').replace(/\s+/g,' ').trim();const d=e=>{const r=e.getBoundingClientRect(),p=e.closest('[role="menuitem"],[role="button"],button');return {text:t(e),tag:e.tagName,role:e.getAttribute('role'),html:e.outerHTML.slice(0,1200),rect:{x:r.x,y:r.y,width:r.width,height:r.height},parent:p&&{tag:p.tagName,role:p.getAttribute('role'),text:t(p),html:p.outerHTML.slice(0,1200),rect:(()=>{const q=p.getBoundingClientRect();return{x:q.x,y:q.y,width:q.width,height:q.height}})()}}};return { model: [...document.querySelectorAll('[data-input-engine-actionbar-control-key="model"]')].filter(v).map(d), ratio: [...document.querySelectorAll('[data-creation-params-panel-id]')].filter(v).map(d), menus: [...document.querySelectorAll('[role="menu"]')].filter(v).map(t) }})()"#).await.expect("读取图片设置菜单失败");
        println!("image ratio menu: {menu:?}, options: {options:?}");
        let model_rect = &options["model"][0]["rect"];
        session.send("Input.dispatchMouseEvent", json!({"type":"mouseMoved","x":model_rect["x"].as_f64().unwrap()+model_rect["width"].as_f64().unwrap()/2.0,"y":model_rect["y"].as_f64().unwrap()+model_rect["height"].as_f64().unwrap()/2.0,"pointerType":"mouse"})).await.expect("悬停模型菜单失败");
        tokio::time::sleep(Duration::from_millis(1000)).await;
        let model_options = session.evaluate(r#"(() => [...document.querySelectorAll('[role="menuitem"],[role="option"]')].filter(e=>{const r=e.getBoundingClientRect(),s=getComputedStyle(e);return r.width>0&&r.height>0&&s.display!=='none'&&s.visibility!=='hidden'}).map(e=>(e.innerText||e.textContent||'').replace(/\s+/g,' ').trim()).filter(Boolean))()"#).await.expect("读取模型子菜单失败");
        println!("image model submenu: {model_options:?}");
        session
            .send("Input.dispatchKeyEvent", json!({"type":"keyDown","key":"Escape","code":"Escape","windowsVirtualKeyCode":27,"nativeVirtualKeyCode":27}))
            .await
            .expect("收起比例菜单失败");
        session.close().await;
    }

    /// 真豆包干跑一条完整的准备链路,跑到 `checkSend` 为止(**不 send**),之后清空输入框并打印页面状态。
    async fn live_dry_run(r: GenerateRequest, refs: Vec<PathBuf>) {
        use crate::cdp::{list_targets, select_pages};
        let _ = tracing_subscriber::fmt()
            .with_env_filter("doubao_relay_lib=debug")
            .with_test_writer()
            .try_init();
        let targets = list_targets("127.0.0.1", 9707).await.unwrap();
        let (chat, _) = select_pages(&targets, 9707);
        let chat = chat.first().expect("需要 doubao-chat page");
        let session = CdpSession::connect(&chat.ws_url, &chat.id).await.unwrap();
        // 正式 Relay 会话建立时会启用这两个后台页面能力；干跑测试直连 CDP，
        // 也必须补上，否则新版豆包会把视频工具条保留为不可交互的隐藏副本。
        for (method, params) in [
            ("Page.setWebLifecycleState", json!({ "state": "active" })),
            (
                "Emulation.setFocusEmulationEnabled",
                json!({ "enabled": true }),
            ),
        ] {
            match session.send(method, params).await {
                Ok(_) => println!("live page capability enabled: {method}"),
                Err(error) => println!("live page capability unavailable: {method}: {error}"),
            }
        }
        let task_id = crate::tasks::new_task_id();
        let plan = SubmissionPlan::for_request(&r, &task_id);
        let (stages, on_stage) = stage_log();
        let started = Instant::now();
        let result = prepare_and_fill(
            &session,
            &r,
            &plan,
            &refs,
            &SubmitTiming::default(),
            &CancellationToken::new(),
            &on_stage,
        )
        .await;
        println!("prepare_and_fill: {result:?} in {:?}", started.elapsed());
        println!("stages: {:?}", stages.lock().unwrap());
        println!(
            "input controls after: {:?}",
            session
                .evaluate(generation_input_controls_snapshot_expression())
                .await
        );
        let verify = session
            .evaluate(&step_expression(
                &plan.tracking_key,
                "verifyVideoSettings",
                &build_js_request(&r, &plan, &refs, refs.len(), false),
            ))
            .await;
        println!("verifyVideoSettings(after): {verify:?}");
        let cleared = session.evaluate(&clear_input_expression()).await;
        println!("clear: {cleared:?}");
        // 还在技能里:删掉本次上传的卡片(参考图属于技能草稿,出了技能删不掉)。
        // `DOUBAO_LIVE_KEEP_REFS=1` 时保留,用来复现"草稿残留 → 下一次 prepare 清理重试"。
        if std::env::var("DOUBAO_LIVE_KEEP_REFS").as_deref() != Ok("1") {
            println!("removed: {}", upload::remove_attachments(&session).await);
        }
        let probe = session
            .evaluate(r#"({ href: location.href, exit: !!document.querySelector('[data-testid="skill_input_exit_button"]'), editor: (document.querySelector('[contenteditable="true"]') || {}).innerText, slider: [...document.querySelectorAll('[role=slider]')].map(s => s.getAttribute('aria-valuenow') + '/' + s.getAttribute('aria-valuemax')), marked: document.querySelectorAll('input[data-shenlang-reference-upload]').length })"#)
            .await;
        println!("after: {probe:?}");
        session.close().await;
        result.unwrap();
    }

    /// 仅观察普通聊天输入区的加号菜单：可信点击一次、打印弹层结构、再按 Escape 收起。
    /// 不选择菜单项、不填写提示词、不发送。豆包页面改版时可用它重新确认生成入口。
    #[tokio::test]
    #[ignore]
    async fn live_generation_launcher_menu_probe() {
        use crate::cdp::{list_targets, select_pages};

        let targets = list_targets("127.0.0.1", 9707).await.unwrap();
        let (chat, _) = select_pages(&targets, 9707);
        let chat = chat.first().expect("需要 doubao-chat page");
        let session = CdpSession::connect(&chat.ws_url, &chat.id).await.unwrap();
        session
            .send(
                "Input.dispatchKeyEvent",
                json!({
                    "type": "keyDown",
                    "key": "Escape",
                    "code": "Escape",
                    "windowsVirtualKeyCode": 27,
                    "nativeVirtualKeyCode": 27
                }),
            )
            .await
            .expect("收起已有菜单失败");
        tokio::time::sleep(Duration::from_millis(100)).await;
        let probe = session
            .evaluate(generation_launcher_probe_expression())
            .await
            .expect("读取生成入口失败");
        println!("generation launcher before: {probe:?}");
        let Some(rect) = probe.get("launcher").and_then(|value| value.get("rect")) else {
            panic!("未找到唯一的普通聊天加号，未执行点击");
        };
        let x = rect
            .get("x")
            .and_then(Value::as_f64)
            .expect("加号 x 坐标缺失");
        let y = rect
            .get("y")
            .and_then(Value::as_f64)
            .expect("加号 y 坐标缺失");
        let menu_before = session
            .evaluate(generation_launcher_menu_snapshot_expression())
            .await
            .expect("读取加号菜单初始状态失败");
        println!("generation launcher menu before: {menu_before:?}");
        for (kind, buttons) in [("mouseMoved", 0), ("mousePressed", 1), ("mouseReleased", 0)] {
            let mut params = json!({ "type": kind, "x": x, "y": y });
            if kind != "mouseMoved" {
                params["button"] = json!("left");
                params["buttons"] = json!(buttons);
                params["clickCount"] = json!(1);
            }
            session
                .send("Input.dispatchMouseEvent", params)
                .await
                .expect("可信点击加号失败");
        }
        tokio::time::sleep(Duration::from_millis(600)).await;
        let menu = session
            .evaluate(generation_launcher_menu_snapshot_expression())
            .await
            .expect("读取加号菜单失败");
        println!("generation launcher menu: {menu:?}");
        session
            .send(
                "Input.dispatchKeyEvent",
                json!({
                    "type": "keyDown",
                    "key": "Escape",
                    "code": "Escape",
                    "windowsVirtualKeyCode": 27,
                    "nativeVirtualKeyCode": 27
                }),
            )
            .await
            .expect("收起加号菜单失败");
        session.close().await;
    }

    /// 仅观察普通聊天输入区的“更多”菜单，不选择项目、不填写提示词、不发送。
    #[tokio::test]
    #[ignore]
    async fn live_skill_more_menu_probe() {
        use crate::cdp::{list_targets, select_pages};

        let targets = list_targets("127.0.0.1", 9707).await.unwrap();
        let (chat, _) = select_pages(&targets, 9707);
        let chat = chat.first().expect("需要豆包聊天页面");
        let session = CdpSession::connect(&chat.ws_url, &chat.id).await.unwrap();
        exit_residual_skill(
            &session,
            &SubmitTiming::default(),
            &CancellationToken::new(),
        )
        .await
        .unwrap();
        session
            .send(
                "Input.dispatchKeyEvent",
                json!({
                    "type": "keyDown", "key": "Escape", "code": "Escape",
                    "windowsVirtualKeyCode": 27, "nativeVirtualKeyCode": 27
                }),
            )
            .await
            .unwrap();
        tokio::time::sleep(Duration::from_millis(100)).await;
        let point = session
            .evaluate(
                r#"(() => {
  const el = document.querySelector('[data-testid="skill_bar_button_more"]');
  if (!el) return null;
  const r = el.getBoundingClientRect();
  return { x: r.left + r.width / 2, y: r.top + r.height / 2 };
})()"#,
            )
            .await
            .unwrap();
        let point: NativeClickPoint = serde_json::from_value(point).expect("未找到更多按钮");
        let clicked = dispatch_trusted_click(
            &session,
            &point,
            Duration::from_secs(3),
            &CancellationToken::new(),
        )
        .await;
        assert!(clicked.is_ok(), "点击更多按钮失败");
        for delay in [0, 50, 100, 150, 250, 400, 600] {
            if delay > 0 {
                tokio::time::sleep(Duration::from_millis(delay)).await;
            }
            println!(
                "skill entry positions +{delay}ms: {:?}",
                session
                    .evaluate(
                        r#"(() => [...document.querySelectorAll('button')]
  .filter(el => ['图像生成', '视频生成'].includes((el.innerText || el.textContent || '').trim()))
  .map(el => { const r = el.getBoundingClientRect(); return { text: (el.innerText || el.textContent || '').trim(), x: r.x, y: r.y, width: r.width, height: r.height }; }))()"#,
                    )
                    .await
            );
        }
        println!(
            "skill more menu: {:?}",
            session
                .evaluate(generation_launcher_menu_snapshot_expression())
                .await
        );
        println!(
            "skill entries: {:?}",
            session
                .evaluate(
                    r#"(() => [...document.querySelectorAll('button,[role="button"],div,span')]
  .filter(el => ['图像生成', '视频生成'].includes((el.innerText || el.textContent || '').trim()))
  .map(el => {
    const r = el.getBoundingClientRect();
    const p = el.parentElement;
    const pr = p?.getBoundingClientRect();
    return {
      tag: el.tagName.toLowerCase(), text: (el.innerText || el.textContent || '').trim(),
      className: String(el.className || '').slice(0, 160),
      rect: { x: r.x, y: r.y, width: r.width, height: r.height },
      parent: p ? { tag: p.tagName.toLowerCase(), className: String(p.className || '').slice(0, 160), rect: { x: pr.x, y: pr.y, width: pr.width, height: pr.height } } : null
    };
  }))()"#,
                )
                .await
        );
        session
            .send(
                "Input.dispatchKeyEvent",
                json!({
                    "type": "keyDown", "key": "Escape", "code": "Escape",
                    "windowsVirtualKeyCode": 27, "nativeVirtualKeyCode": 27
                }),
            )
            .await
            .unwrap();
        session.close().await;
    }

    /// 仅走“退出技能 → 注入 → 展开视频工具栏”，打印前后控件；不填提示词、不选参数、不发送。
    #[tokio::test]
    #[ignore]
    async fn live_video_toolbar_state_probe() {
        use crate::cdp::{list_targets, select_pages};

        let targets = list_targets("127.0.0.1", 9707).await.unwrap();
        let (chat, _) = select_pages(&targets, 9707);
        let chat = chat.first().expect("需要 doubao-chat page");
        let session = CdpSession::connect(&chat.ws_url, &chat.id).await.unwrap();
        let cancel = CancellationToken::new();
        let timing = SubmitTiming::default();
        exit_residual_skill(&session, &timing, &cancel)
            .await
            .expect("退出残留技能失败");
        println!(
            "video toolbar before: {:?}",
            session
                .evaluate(generation_input_controls_snapshot_expression())
                .await
        );
        let key = tracking_key_for(&crate::tasks::new_task_id());
        session
            .evaluate(&generation_bootstrap(&key))
            .await
            .expect("注入生成脚本失败");
        let toolbar = evaluate_step(
            &session,
            &open_new_generation_toolbar_expression(TaskType::Video, Duration::from_secs(8)),
            timing.step_timeout,
            &cancel,
        )
        .await;
        match toolbar {
            Ok(value) => println!("video toolbar result: {value:?}"),
            Err(StepError::Js(message)) => println!("video toolbar JS error: {message}"),
            Err(StepError::Timeout) => println!("video toolbar timed out"),
            Err(StepError::Cdp(error)) => println!("video toolbar CDP error: {error}"),
            Err(StepError::Cancelled) => println!("video toolbar cancelled"),
        }
        println!(
            "video toolbar after: {:?}",
            session
                .evaluate(generation_input_controls_snapshot_expression())
                .await
        );
        session.close().await;
    }

    /// `DOUBAO_LIVE_DURATION=7 cargo test live_video_duration_dry_run -- --ignored --nocapture`:
    /// 视频 Seedance 2.0 Mini / 16:9 / 目标秒数(默认 7,选一个与豆包记住的值不同的数才会走滑块)。
    #[tokio::test]
    #[ignore]
    async fn live_video_duration_dry_run() {
        let secs: u32 = std::env::var("DOUBAO_LIVE_DURATION")
            .ok()
            .and_then(|v| v.parse().ok())
            .unwrap_or(7);
        let model =
            std::env::var("DOUBAO_LIVE_MODEL").unwrap_or_else(|_| "Seedance 2.0 Mini".into());
        let ratio = std::env::var("DOUBAO_LIVE_VIDEO_RATIO").unwrap_or_else(|_| "16:9".into());
        let r = GenerateRequest {
            model_name: model,
            ratio,
            duration: Some(secs),
            prompt: "测试提示词(干跑,不发送)".into(),
            ..video_request()
        };
        live_dry_run(r, Vec::new()).await;
    }

    /// `DOUBAO_LIVE_REF=/abs/a.png cargo test live_reference_upload_dry_run -- --ignored --nocapture`:
    /// 图像 Seedream 5.0 Pro / 1:1(`DOUBAO_LIVE_TYPE=video` 则视频 Seedance 2.0 Fast / 16:9 / 5s)
    /// + 参考图(逗号分隔多张,视频技能也接受 mp4 参考视频),上传后跑到 `checkSend`,不发送;
    /// 结束时在技能里删掉卡片(`DOUBAO_LIVE_KEEP_REFS=1` 保留,下一次同技能 prepare 会清理后重试)。
    #[tokio::test]
    #[ignore]
    async fn live_reference_upload_dry_run() {
        let refs: Vec<PathBuf> = std::env::var("DOUBAO_LIVE_REF")
            .expect("需要 DOUBAO_LIVE_REF")
            .split(',')
            .map(PathBuf::from)
            .collect();
        let base = if std::env::var("DOUBAO_LIVE_TYPE").as_deref() == Ok("video") {
            GenerateRequest {
                model_name: "Seedance 2.0 Fast".into(),
                duration: Some(5),
                ..video_request()
            }
        } else {
            image_request()
        };
        let r = GenerateRequest {
            model_name: std::env::var("DOUBAO_LIVE_IMAGE_MODEL")
                .unwrap_or_else(|_| base.model_name.clone()),
            ratio: std::env::var("DOUBAO_LIVE_IMAGE_RATIO").unwrap_or_else(|_| base.ratio.clone()),
            prompt: std::env::var("DOUBAO_LIVE_PROMPT")
                .unwrap_or_else(|_| "测试提示词(干跑,不发送)".into()),
            ..base
        };
        live_dry_run(r, refs).await;
    }
}
