const test = require('node:test');
const assert = require('node:assert/strict');
const { EventEmitter } = require('node:events');
const vm = require('node:vm');
const {
  DoubaoPageActions,
  DoubaoControlError,
  buildPreferredClickScript
} = require('../src/doubao-page-actions');

test('setPrompt uses the live prompt editor and fails closed when none exists', async () => {
  const calls = [];
  const webContents = {
    async executeJavaScript(script, userGesture) {
      calls.push([script, userGesture]);
      return { ok: true, kind: 'contenteditable' };
    }
  };
  const result = await new DoubaoPageActions().setPrompt(webContents, '雨夜城市');
  assert.equal(result.kind, 'contenteditable');
  assert.equal(calls[0][1], true);
  assert.equal(calls[0][0].includes('雨夜城市'), true);

  const missing = { async executeJavaScript() { return { ok: false, reason: 'not_found' }; } };
  await assert.rejects(() => new DoubaoPageActions().setPrompt(missing, 'x'), error => error instanceof DoubaoControlError && error.code === 'PROMPT_INPUT_NOT_FOUND');
});

test('exact option click fails when zero or multiple live controls match', async () => {
  const actions = new DoubaoPageActions();
  await assert.rejects(() => actions.clickExactControl({ async executeJavaScript() { return { count: 0 }; } }, '10秒'), error => error.code === 'CONTROL_NOT_FOUND');
  await assert.rejects(() => actions.clickExactControl({ async executeJavaScript() { return { count: 2 }; } }, '10秒'), error => error.code === 'CONTROL_AMBIGUOUS');
  const clicked = await actions.clickExactControl({ async executeJavaScript() { return { count: 1, clicked: true }; } }, '10秒');
  assert.equal(clicked, true);
});

test('preferred click treats the Doubao AI creation Video tab as an interactive control', () => {
  let clicked = false;
  const videoTab = {
    innerText: '视频',
    textContent: '视频',
    disabled: false,
    tabIndex: 0,
    getAttribute(name) {
      if (name === 'role') return 'tab';
      if (name === 'aria-disabled') return 'false';
      if (name === 'aria-label') return '';
      return null;
    },
    getBoundingClientRect() { return { left: 10, top: 10, right: 74, bottom: 42, width: 64, height: 32 }; },
    scrollIntoView() {},
    click() { clicked = true; }
  };
  const result = vm.runInNewContext(buildPreferredClickScript(['视频']), {
    document: {
      querySelectorAll(selector) {
        return selector.includes('[role="tab"]') ? [videoTab] : [];
      },
      elementFromPoint() { return videoTab; }
    },
    innerWidth: 1280,
    innerHeight: 720,
    getComputedStyle() { return { display: 'block', visibility: 'visible' }; }
  });

  assert.equal(result.count, 1);
  assert.equal(result.clicked, true);
  assert.equal(result.label, '视频');
  assert.equal(clicked, true);
});

test('preferred click ignores duplicated carousel Video controls that are not actually hit-testable', () => {
  let activeClicks = 0;
  let cloneClicks = 0;
  const active = {
    innerText: '视频',
    textContent: '视频',
    disabled: false,
    tabIndex: 0,
    getAttribute(name) {
      if (name === 'aria-disabled') return 'false';
      if (name === 'aria-label') return '';
      return null;
    },
    getBoundingClientRect() { return { left: 410, top: 300, right: 474, bottom: 332, width: 64, height: 32 }; },
    contains(node) { return node === active; },
    scrollIntoView() {},
    click() { activeClicks += 1; }
  };
  const clone = {
    innerText: '视频',
    textContent: '视频',
    disabled: false,
    tabIndex: -1,
    getAttribute(name) {
      if (name === 'aria-disabled') return 'false';
      if (name === 'aria-label') return '';
      return null;
    },
    getBoundingClientRect() { return { left: 800, top: 300, right: 864, bottom: 332, width: 64, height: 32 }; },
    contains(node) { return node === clone; },
    scrollIntoView() {},
    click() { cloneClicks += 1; }
  };
  const blocker = {};
  const result = vm.runInNewContext(buildPreferredClickScript(['视频']), {
    document: {
      querySelectorAll() { return [active, clone]; },
      elementFromPoint(x) { return x < 700 ? active : blocker; }
    },
    innerWidth: 1280,
    innerHeight: 720,
    getComputedStyle() { return { display: 'block', visibility: 'visible' }; }
  });

  assert.equal(result.count, 1);
  assert.equal(result.clicked, true);
  assert.equal(result.label, '视频');
  assert.equal(activeClicks, 1);
  assert.equal(cloneClicks, 0);
});

test('reference image upload uses CDP file input assignment without a file chooser', async () => {
  const debug = new EventEmitter();
  debug.isAttached = () => true;
  const commands = [];
  debug.sendCommand = async (method, params) => {
    commands.push([method, params]);
    if (method === 'Runtime.evaluate') return { result: { objectId: 'object-file-input' } };
    return {};
  };
  const actions = new DoubaoPageActions();
  await actions.setReferenceImages({ debugger: debug }, ['/tmp/a.png', '/tmp/b.jpg']);
  assert.ok(commands.some(([method]) => method === 'Runtime.evaluate'));
  assert.ok(commands.some(([method, params]) => method === 'DOM.setFileInputFiles'
    && params.objectId === 'object-file-input'
    && params.files.length === 2));
});

test('normal confirmation is limited to explicit generation confirmation labels', async () => {
  const actions = new DoubaoPageActions();
  const none = await actions.confirmNormal({ async executeJavaScript(script) {
    assert.equal(script.includes('确认生成'), true);
    assert.equal(script.includes('继续生成'), true);
    return { count: 0, clicked: false };
  } });
  assert.equal(none, false);
});
