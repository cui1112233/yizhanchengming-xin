import { describe, expect, it } from 'vitest';
import { asRelayError, isDoubaoUnavailable } from './api';

describe('asRelayError', () => {
  it('结构化 IpcError 原样返回(含 details)', () => {
    const e = asRelayError({ code: 'account_not_found', message: '账号不存在', details: { a: 1 } });
    expect(e).toEqual({ code: 'account_not_found', message: '账号不存在', details: { a: 1 } });
  });

  it('缺 details 时不带 details 字段', () => {
    const e = asRelayError({ code: 'doubao_not_ready', message: '豆包未就绪' });
    expect(e).toEqual({ code: 'doubao_not_ready', message: '豆包未就绪' });
    expect('details' in e).toBe(false);
  });

  it('Error / 字符串 / 其它值 → system_internal_error,message 保留原文', () => {
    expect(asRelayError(new Error('boom'))).toEqual({
      code: 'system_internal_error',
      message: 'boom',
    });
    expect(asRelayError('纯字符串')).toEqual({ code: 'system_internal_error', message: '纯字符串' });
    expect(asRelayError(undefined).code).toBe('system_internal_error');
    expect(asRelayError(null).message.length).toBeGreaterThan(0);
    expect(asRelayError({ foo: 'bar' }).code).toBe('system_internal_error');
  });

  it('code 不是字符串的对象也归为 system_internal_error', () => {
    expect(asRelayError({ code: 42, message: 'x' }).code).toBe('system_internal_error');
  });
});

describe('isDoubaoUnavailable', () => {
  it('doubao_* 错误视为豆包不可用(空状态已覆盖,不弹 toast)', () => {
    expect(isDoubaoUnavailable({ code: 'doubao_cdp_disconnected', message: '' })).toBe(true);
    expect(isDoubaoUnavailable({ code: 'doubao_not_ready', message: '' })).toBe(true);
    expect(isDoubaoUnavailable({ code: 'doubao_not_logged_in', message: '' })).toBe(true);
  });
  it('其它错误不是', () => {
    expect(isDoubaoUnavailable({ code: 'account_not_found', message: '' })).toBe(false);
    expect(isDoubaoUnavailable({ code: 'system_internal_error', message: '' })).toBe(false);
  });
});
