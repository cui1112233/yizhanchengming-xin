// 首次启动的全屏隐私声明(docs/spec/12-security.md "隐私声明":MUST 明确同意一次才可用)。
// 同意状态存后端 `metadata.privacy_accepted`;同意前后端不启动 HTTP API、不自动连豆包。
import { signal } from '@preact/signals';
import { useLayoutEffect, useRef } from 'preact/hooks';
import * as api from '../api';
import { asRelayError } from '../api';
import { privacyAccepted } from '../state';
import './privacy-notice.css';

/** 声明正文(spec 12 原文)。 */
export const PROMISES = [
  '所有数据保存在本机,从不上传任何外部服务',
  '不收集使用指标、不记录 Prompt 用于分析',
  '不代理豆包登录,不收集账号密码',
  '代码开源,任何人可审计',
];
export const WARNINGS = [
  '本工具通过模拟客户端操作使用豆包账号。使用本工具可能违反豆包用户协议,账号被封等后果由用户承担。',
  '请仅使用自己拥有的豆包账号。',
];

const checked = signal(false);
const busy = signal<'accept' | 'quit' | null>(null);
const error = signal<string | null>(null);

/** "同意并继续"是否可点(纯函数,便于单测)。 */
export function canAccept(isChecked: boolean, isBusy: boolean): boolean {
  return isChecked && !isBusy;
}

/** 启动时读同意状态;读失败按未同意处理(宁可多弹一次声明)。 */
export async function loadPrivacy(): Promise<void> {
  try {
    privacyAccepted.value = (await api.privacyStatus()).accepted;
  } catch {
    privacyAccepted.value = false;
  }
}

export async function acceptPrivacy(): Promise<void> {
  if (!canAccept(checked.value, busy.value !== null)) return;
  busy.value = 'accept';
  error.value = null;
  try {
    const s = await api.privacyAccept();
    privacyAccepted.value = s.accepted;
  } catch (raw) {
    error.value = `保存同意状态失败:${asRelayError(raw).message}`;
  } finally {
    busy.value = null;
  }
}

export async function quitApp(): Promise<void> {
  if (busy.value !== null) return;
  busy.value = 'quit';
  error.value = null;
  try {
    await api.appQuit();
  } catch (raw) {
    error.value = `退出失败:${asRelayError(raw).message}`;
    busy.value = null;
  }
}

/** 测试用:复位模块内状态。 */
export function resetPrivacyNotice() {
  checked.value = false;
  busy.value = null;
  error.value = null;
}

export function PrivacyNotice() {
  const accepted = privacyAccepted.value;
  const boxRef = useRef<HTMLInputElement>(null);
  useLayoutEffect(() => {
    if (accepted === false) boxRef.current?.focus();
  }, [accepted]);
  if (accepted === true) return null;
  if (accepted === null) return <div class="privacy-backdrop" aria-busy="true" />;
  const isBusy = busy.value !== null;
  return (
    <div class="privacy-backdrop">
      <div class="privacy-card" role="dialog" aria-modal="true" aria-labelledby="privacy-title">
        <h1 id="privacy-title" class="privacy-title">
          使用前请阅读
        </h1>
        <p class="privacy-lead">一战晟铭执行器承诺:</p>
        <ul class="privacy-list privacy-promises">
          {PROMISES.map((t) => (
            <li key={t}>✓ {t}</li>
          ))}
        </ul>
        <p class="privacy-lead">警告:</p>
        <ul class="privacy-list privacy-warnings">
          {WARNINGS.map((t) => (
            <li key={t}>⚠ {t}</li>
          ))}
        </ul>
        <label class="privacy-check">
          <input
            ref={boxRef}
            type="checkbox"
            checked={checked.value}
            disabled={isBusy}
            onChange={(ev) => {
              checked.value = (ev.currentTarget as HTMLInputElement).checked;
            }}
          />
          我已阅读并理解上述说明与风险
        </label>
        {error.value && (
          <p class="privacy-error" role="alert">
            {error.value}
          </p>
        )}
        <div class="privacy-actions">
          <button class="btn" disabled={isBusy} onClick={() => void quitApp()}>
            {busy.value === 'quit' ? '正在退出…' : '不同意并退出'}
          </button>
          <button
            class="btn btn-primary"
            disabled={!canAccept(checked.value, isBusy)}
            onClick={() => void acceptPrivacy()}
          >
            {busy.value === 'accept' ? '正在保存…' : '同意并继续'}
          </button>
        </div>
      </div>
    </div>
  );
}
