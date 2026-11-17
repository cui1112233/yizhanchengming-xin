import test from 'node:test';
import assert from 'node:assert/strict';
import { navigateToHomeLogin } from './loginNavigation.js';

test('login requested from a protected page navigates home and opens the homepage login', () => {
  const calls = [];
  const browser = { queueMicrotask: callback => callback(), location: { pathname: '/settings' }, history: { replaceState: (...args) => calls.push(args[2]) }, dispatchEvent: event => calls.push(event.type), PopStateEvent: class { constructor(type) { this.type = type; } } };
  navigateToHomeLogin(() => calls.push('open'), browser);
  assert.deepEqual(calls, ['open', '/', 'popstate']);
});
test('login requested on the homepage opens without replacing its URL', () => {
  let opened = false;
  navigateToHomeLogin(() => { opened = true; }, { location: { pathname: '/' } });
  assert.equal(opened, true);
});
test('initial protected entry waits until parent route listeners can mount', () => {
  const actions = [], queued = [];
  const browser = { location: { pathname: '/settings' }, queueMicrotask: callback => queued.push(callback), history: { replaceState: () => actions.push('home-url') }, dispatchEvent: () => actions.push('route-render'), PopStateEvent: class {} };
  navigateToHomeLogin(() => actions.push('login'), browser);
  assert.deepEqual(actions, ['login', 'home-url']);
  queued[0]();
  assert.deepEqual(actions, ['login', 'home-url', 'route-render']);
});
