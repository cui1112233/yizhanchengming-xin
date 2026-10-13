const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');

const {
  BATCH_FACTORY_PRESET_REQUIREMENTS,
  defaultBody,
  isPublishedPresetAllowed,
  listPublishedForSlot,
  seedSystemPresets,
  SYSTEM_PRESETS
} = require('./system-preset-catalog');

test('Batch Factory extraction accepts only the combined people-scene-prop preset', () => {
  assert.equal(isPublishedPresetAllowed(
    { id: 'script-extract-assets', module: 'script', kind: 'base', extractionPreset: true, protocolLock: { format: 'extract', slot: 'script.asset-extraction' } },
    BATCH_FACTORY_PRESET_REQUIREMENTS['assets.extraction']
  ), true);
  assert.equal(isPublishedPresetAllowed(
    { id: 'script-extract', module: 'script', kind: 'base', extractionPreset: true, protocolLock: { format: 'extract', slot: 'script.extract' } },
    BATCH_FACTORY_PRESET_REQUIREMENTS['assets.extraction']
  ), false);
});

test('seeding upgrades only the shipped invalid script extraction example to the corrected JSON contract', () => {
  const extraction = SYSTEM_PRESETS.find(preset => preset.id === 'script-extract');
  const legacy = extraction.body.replace(/\r\n/g, '\n')
    .replace('    }\n  ],\n  "场景设定": [', '    },\n    \n  "场景设定": [')
    .replace('    }\n  ]\n}', '    },\n  ]\n}');
  const published = new Map([[extraction.id, { ...extraction, body: legacy, version: 3, status: 'published' }]]);
  const drafts = new Map();
  const store = {
    getPublished: id => published.get(id) || null,
    createDraft: (_, input) => {
      const draft = { ...input, version: (published.get(input.id)?.version || 0) + 1 };
      drafts.set(input.id, draft);
      return draft;
    },
    publish: (_, id) => {
      const publishedDraft = { ...drafts.get(id), status: 'published' };
      published.set(id, publishedDraft);
      return publishedDraft;
    }
  };

  seedSystemPresets(store, 'admin');

  assert.equal(published.get(extraction.id).version, 4);
  assert.equal(published.get(extraction.id).body, extraction.body.trim());
});

test('seeding preserves a script extraction template customized only with leading whitespace', () => {
  const extraction = SYSTEM_PRESETS.find(preset => preset.id === 'script-extract');
  const legacy = extraction.body.replace(/\r\n/g, '\n')
    .replace('    }\n  ],\n  "场景设定": [', '    },\n    \n  "场景设定": [')
    .replace('    }\n  ]\n}', '    },\n  ]\n}');
  const customized = `\n${legacy}`;
  const published = new Map([[extraction.id, { ...extraction, body: customized, version: 3, status: 'published' }]]);
  const drafts = new Map();
  const store = {
    getPublished: id => published.get(id) || null,
    createDraft: (_, input) => {
      const draft = { ...input, version: (published.get(input.id)?.version || 0) + 1 };
      drafts.set(input.id, draft);
      return draft;
    },
    publish: (_, id) => {
      const publishedDraft = { ...drafts.get(id), status: 'published' };
      published.set(id, publishedDraft);
      return publishedDraft;
    }
  };

  seedSystemPresets(store, 'admin');

  assert.equal(published.get(extraction.id).version, 3);
  assert.equal(published.get(extraction.id).body, customized);
});

test('Batch Factory exposes separate character and scene renderers while prop remains a combined extraction output', () => {
	assert.deepEqual(BATCH_FACTORY_PRESET_REQUIREMENTS['assets.character'], { module: 'batch-factory', kind: 'base', slot: 'batch.character-meta' });
	assert.deepEqual(BATCH_FACTORY_PRESET_REQUIREMENTS['assets.scene'], { module: 'batch-factory', kind: 'base', slot: 'batch.scene-meta' });
	assert.equal(BATCH_FACTORY_PRESET_REQUIREMENTS['assets.prop'], undefined);
});

test('Batch Factory provides one combined people-scene-prop extraction preset', () => {
  const combined = SYSTEM_PRESETS.find(preset => preset.id === 'script-extract-assets');
  assert.equal(combined?.protocolLock?.slot, 'script.asset-extraction');
  assert.equal(combined?.protocolLock?.format, 'extract');
  assert.match(combined?.body || '', /人物场景道具统一提取/);
});

test('Batch Factory storyboard meta default contains the full script segmented-opening and shotlist bodies', () => {
  const storyboard = SYSTEM_PRESETS.find(preset => preset.id === 'batch-video-meta');
  assert.match(storyboard?.body || '', /## 剧本生成预设：分段开头/);
  assert.match(storyboard?.body || '', /先全文理解，再开始分段/);
  assert.match(storyboard?.body || '', /## 剧本生成预设：分镜模式/);
  assert.match(storyboard?.body || '', /每条画面的完整导演信息/);
  assert.match(storyboard?.body || '', /Batch Factory V11 JSON 输出适配/);
  assert.doesNotMatch(storyboard?.body || '', /^你负责把导演结果组织成可执行的视频分镜描述/);
});

test('Batch Factory publishes the H3 director VIDEO renderer as a selectable video preset', () => {
	const h3 = SYSTEM_PRESETS.find(preset => preset.id === 'batch-video-h3-director');
	assert.equal(h3?.protocolLock?.slot, 'batch.video-meta');
	assert.equal(h3?.protocolLock?.key, 'h3-video-normal');
	for (const requiredRule of ['H3_DIRECTOR_R8_FULL_RULESET_V1', '通用小说剧情脑补与段内微镜头硬约束', '每个非空原文行固定对应一张外层分镜卡', '1到6个micro_shots', 'character_slot_ids', '180度轴线', '不负责最终精确总时长数学']) {
		assert.match(h3?.body || '', new RegExp(requiredRule));
	}
});

test('Batch Factory H3 director preset retains the original H3 R8 director protocol', () => {
  const h3 = SYSTEM_PRESETS.find(preset => preset.id === 'batch-video-h3-director');
  const body = h3?.body || '';
  assert.match(body, /通用小说剧情脑补与段内微镜头硬约束｜V78\.3\.0\.17 R8/);
  assert.match(body, /顶层固定为 \{"timeline_segments":\[\.\.\.\]\}/);
  assert.match(body, /scene_id、location、axis、light_direction、positions、facings、gazes、held_props、action_ends/);
});

test('H3 selectable presets embed the complete supplied V78.3.0.59 rule bodies instead of adapted summaries', () => {
  const smart = defaultBody('script-constraint-prefix-smart-unified');
  const assets = defaultBody('batch-assets-h3');
  const director = defaultBody('batch-video-h3-director');

  // Distinctive source-only clauses make this a provenance test, rather than
  // a test that accepts a hand-written paraphrase of the H3 protocols.
  assert.match(smart, /【V78\.3\.0\.59 最新版｜内容类型与统一风格质量协议】/);
  assert.match(smart, /\{\{\.StyleAdvice\}\}/);
  assert.match(assets, /【V59 AUTO正式人物名单判断｜只确定正式人物基础身份】/);
  assert.match(assets, /\{\{\.ForcedRoster\}\}/);
  assert.match(assets, /【V52通用小说人物关系图与分镜选角Skill｜关系图独立，只读人物卡】/);
  assert.match(assets, /【V55人物卡外形隔离系统规则｜人物外形接口最终系统级规则】/);
  assert.match(director, /【v78\.3\.0\.59 H3普通模式｜整段分镜质量冻结母版】/);
  assert.match(director, /\{\{\.H3_CANONICAL_PROMPT\}\}/);
});

test('Batch Factory publishes H3 as one complete asset scheme in the unified selector', () => {
	const h3 = SYSTEM_PRESETS.find(preset => preset.id === 'batch-assets-h3');
	assert.equal(h3?.module, 'script');
	assert.equal(h3?.protocolLock?.slot, 'script.asset-extraction');
	assert.equal(h3?.protocolLock?.key, 'h3-assets-full');
	assert.match(h3?.body || '', /一次模型请求、一个 JSON 返回/);
	assert.match(h3?.body || '', /V59 AUTO正式人物名单判断/);
	assert.match(h3?.body || '', /V55人物卡外形隔离系统规则/);
});

test('Batch Factory publishes the H3 four-step people-asset Skill without changing the legacy H3 mapping', () => {
  const legacy = SYSTEM_PRESETS.find(preset => preset.id === 'batch-assets-h3');
  const skill = SYSTEM_PRESETS.find(preset => preset.id === 'batch-assets-h3-skill');

  // 旧 H3 资产预设映射保持不变。
  assert.equal(legacy?.protocolLock?.slot, 'script.asset-extraction');
  assert.equal(legacy?.protocolLock?.key, 'h3-assets-full');

  // 人物资产 Skill 注册信息：四步资产获取，复用现有智能统一预设。
  assert.equal(skill?.module, 'batch-factory');
  assert.equal(skill?.kind, 'base');
  assert.equal(skill?.name, 'H3 人物资产 Skill');
  assert.match(skill?.description || '', /人物资产/);
  assert.equal(skill?.protocolLock?.slot, 'batch.assets-h3-skill');
  assert.equal(skill?.protocolLock?.format, 'skill');
  assert.equal(skill?.protocolLock?.operation, 'h3-assets-with-smart-unified');
  assert.equal(skill?.protocolLock?.version, 'v78.3.0.59');
  assert.equal(skill?.protocolLock?.smartUnifiedPresetId, 'script-constraint-prefix-smart-unified');

  const normalize = text => String(text).replace(/\r\n/g, '\n').trim().replace(/\n{3,}/g, '\n\n');
  const rawSkill = String(skill?.body || '').replace(/\r\n/g, '\n');
  const skillBody = normalize(rawSkill);
  const stepIds = [
    'character.roster.system__v78.3.0.59',
    'character.facts.system__v78.3.0.59',
    'character.relationships.system__v78.3.0.59',
    'character.appearance.system__v78.3.0.59'
  ];

  // 注册 steps 与正文都只有四步，且顺序固定。
  assert.deepEqual(skill?.protocolLock?.steps, stepIds);
  assert.deepEqual(
    rawSkill.match(/^## 原 H3：.+$/gm),
    stepIds.map(id => `## 原 H3：${id}`)
  );

  // 分镜规划属于现有导演分镜提示词：分镜结构、模板与导演输出格式
  // 一律不得出现在人物资产 Skill。
  assert.doesNotMatch(skillBody, /storyboard/i);
  assert.doesNotMatch(skillBody, /timeline_segments/);
  assert.doesNotMatch(skillBody, /micro_shots/);
  assert.doesNotMatch(skillBody, /h3\.normal\.template/);
  assert.doesNotMatch(skillBody, /H3_CANONICAL_PROMPT/);
  assert.doesNotMatch(skillBody, /批量工厂导演输出格式/);

  // 资产阶段仍声明 characters/scenes/props 顶层字段以兼容资产库。
  assert.match(skillBody, /characters、scenes、props/);

  // 智能统一 style.system 正文未被复制进 Skill（只允许引用 preset id
  // 与“不得复制 style.system 正文”的声明）。
  const smartBody = normalize(defaultBody('script-constraint-prefix-smart-unified'));
  assert.doesNotMatch(skillBody, /【V78\.3\.0\.59 最新版｜内容类型与统一风格质量协议】/);
  assert.doesNotMatch(skillBody, /\{\{\.StyleAdvice\}\}/);
  for (const clause of [
    '# 智能统一｜原 H3 style.system',
    'Sensor Format 与采集解析度倾向',
    'remark_prompt 字段如存在必须为空字符串'
  ]) {
    assert.ok(smartBody.includes(clause), `smart-unified fixture clause missing: ${clause}`);
    assert.ok(!skillBody.includes(clause), `smart-unified body copied into Skill: ${clause}`);
  }

  // 强校验：四步正文分别与四份用户原始 txt 逐字相等（仅归一化换行与
  // 多余空行，不得删改内容）。原始 txt 在隔离工作区之外，使用绝对路径
  // 读取；跨盘符路径用 String.raw 表达。
  const sourceDir = String.raw`e:\Adboe\xwechat_files\wxid_6ysdu12uz18b22_9198\msg\file\2026-09`;
  stepIds.forEach((id, index) => {
    const header = `## 原 H3：${id}`;
    const start = rawSkill.indexOf(header);
    assert.ok(start >= 0, `missing step header: ${id}`);
    const bodyStart = start + header.length;
    const laterStarts = stepIds.slice(index + 1)
      .map(next => rawSkill.indexOf(`## 原 H3：${next}`, bodyStart))
      .filter(position => position >= 0);
    const end = laterStarts.length ? Math.min(...laterStarts) : rawSkill.length;
    const section = normalize(rawSkill.slice(bodyStart, end));
    const authority = normalize(fs.readFileSync(path.join(sourceDir, `${id}.txt`), 'utf8'));
    assert.equal(section, authority, `step body differs from authoritative txt: ${id}`);
  });
});

test('seeding upgrades only the shipped abbreviated H3 director rule to the full R8 rule', () => {
  const h3 = SYSTEM_PRESETS.find(preset => preset.id === 'batch-video-h3-director');
  const legacy = [
    '## H3 导演分镜 VIDEO 输出规则', '',
    '以人物、场景、道具、原文事实和当前镜头连续性为唯一依据。先在内部保持 Scene Memory：场景、空间轴线、人物站位、朝向、视线、手持道具、光线与动作结束状态必须能衔接下一镜。', '',
    '每个 VIDEO 必须输出可拍、可见的镜头动作；明确景别、机位、运镜与节奏。人物必须引用已确认的人物资产，禁止临时改写身份、外观或关系。不要用空镜、重复动作或静止画面填充时长。', '',
    '最终 VIDEO 由批量工厂 H3 编译器生成：人物定义、视听分层、H3 时间轴和无字幕画面规则均由后端根据结构化分镜输出，浏览器不得拼接或伪造。'
  ].join('\n');
  const published = new Map([[h3.id, { ...h3, body: legacy, version: 1, status: 'published' }]]);
  const drafts = new Map();
  const store = {
    getPublished: id => published.get(id) || { id, body: `custom ${id}`, version: 1, status: 'published', protocolLock: {} },
    createDraft: (_, input) => { const draft = { ...input, version: (published.get(input.id)?.version || 0) + 1 }; drafts.set(input.id, draft); return draft; },
    publish: (_, id) => { const draft = drafts.get(id); published.set(id, { ...draft, status: 'published' }); return published.get(id); }
  };
  seedSystemPresets(store, 'admin');
  assert.match(published.get(h3.id).body, /H3_DIRECTOR_R8_FULL_RULESET_V1/);
});

test('seeding upgrades only the shipped two-call H3 asset rule to the one-call full rule', () => {
  const h3 = SYSTEM_PRESETS.find(preset => preset.id === 'batch-assets-h3');
  const legacy = [
    '# H3 人物场景道具提取', '',
    '仅当用户在“人物场景道具提示词”选择本方案时执行。本方案独立于普通提取方案；不会追加、修改或覆盖任何普通人物、场景或道具提示词。', '',
    '## 调用一：H3 人物场景道具事实提取', '',
    '只输出合法 JSON：{"characters":[{"name":"人物名","prompt":"可追溯的人物事实与外形依据"}],"scenes":[],"props":[]}。', '',
    '## 调用二：H3 全人物外形编译', '',
    '只输出 JSON：{"characters":[{"character_id":"C001","name":"输入姓名","prompt":"该人物完整外形正文"}]}。'
  ].join('\n');
  const published = new Map([[h3.id, { ...h3, description: 'H3 两次调用：资产事实提取后一次编译全部人物详细外形', body: legacy, version: 1, status: 'published' }]]);
  const drafts = new Map();
  const store = {
    getPublished: id => published.get(id) || null,
    createDraft: (_, input) => { const draft = { ...input, version: (published.get(input.id)?.version || 0) + 1 }; drafts.set(input.id, draft); return draft; },
    publish: (_, id) => { published.set(id, { ...drafts.get(id), status: 'published' }); return published.get(id); }
  };
  seedSystemPresets(store, 'admin');
  const upgraded = published.get(h3.id);
  assert.equal(upgraded.version, 2);
  assert.equal(upgraded.description, h3.description);
  assert.match(upgraded.body, /一次模型请求、一个 JSON 返回/);
  assert.match(upgraded.body, /V55人物卡外形隔离系统规则/);
});

test('Batch Factory extraction lists every published combined asset preset in its dedicated slot', () => {
  const records = [
    { id: 'script-extract-assets', name: '默认统一提取', module: 'script', kind: 'base', status: 'published', version: 2, protocolLock: { slot: 'script.asset-extraction', format: 'extract' } },
    { id: 'script-extract-assets-female', name: '女频统一提取', module: 'script', kind: 'base', status: 'published', version: 1, protocolLock: { slot: 'script.asset-extraction', format: 'extract' } },
    { id: 'script-extract', name: '人物场景提取', module: 'script', kind: 'base', status: 'published', version: 3, protocolLock: { slot: 'script.extract', format: 'extract' } }
  ];
  const listed = listPublishedForSlot({ listAll: () => records }, 'script.asset-extraction');
  assert.deepEqual(listed.map(item => item.id), ['script-extract-assets', 'script-extract-assets-female']);
  assert.equal(listed[0].kind, 'base');
  assert.equal(listed[0].constraintCategory, '');
});

test('system preset catalog keeps constraint metadata for the configuration selector', () => {
  const listed = listPublishedForSlot({ listAll: () => [{
    id: 'script-constraint-quality-4k', name: '4K 画质约束', description: '画质规则', module: 'script', kind: 'addon', status: 'published', version: 1,
    protocolLock: { slot: 'script.constraint.quality', format: 'constraint', category: 'quality' }
  }] }, 'script.constraint.quality');
  assert.deepEqual(listed[0], {
    id: 'script-constraint-quality-4k',
    name: '4K 画质约束',
    description: '画质规则',
    version: 1,
    slot: 'script.constraint.quality',
    kind: 'addon',
    constraintCategory: 'quality'
  });
});

test('Batch Factory constraint selections accept only published script constraint slots', () => {
  const constraint = { module: 'script', kind: 'addon', protocolLock: { format: 'constraint', slot: 'script.constraint.quality' } };
  assert.equal(isPublishedPresetAllowed(constraint, BATCH_FACTORY_PRESET_REQUIREMENTS.constraints), true);
  assert.equal(isPublishedPresetAllowed({ ...constraint, kind: 'base' }, BATCH_FACTORY_PRESET_REQUIREMENTS.constraints), false);
});

test('publishes H3 shot narrative and visual stability as independently selectable Chinese constraints', () => {
  const shotNarrative = SYSTEM_PRESETS.find(preset => preset.id === 'script-constraint-quality-h3-shot-narrative');
  const visualPolicy = SYSTEM_PRESETS.find(preset => preset.id === 'script-constraint-restriction-h3-visual-policy');
  assert.equal(shotNarrative?.protocolLock?.category, 'quality');
  assert.match(shotNarrative?.body || '', /严格依次呈现列出的场景与镜头/);
  assert.match(shotNarrative?.body || '', /人物情绪和关系主要通过视线、手势、肢体动作/);
  assert.doesNotMatch(shotNarrative?.body || '', /AUDIOVISUAL PRESENTATION/);
  assert.equal(visualPolicy?.protocolLock?.category, 'restriction');
  assert.match(visualPolicy?.body || '', /人物身份、年龄、脸型、五官、发型、服装和饰品/);
});

test('smart unified preset owns the editable H3 style.system protocol', () => {
  const body = defaultBody('script-constraint-prefix-smart-unified');
  assert.match(body, /H3 `style\.system`/);
  assert.match(body, /"final_genre"/);
  assert.match(body, /"trailer_style"/);
  assert.match(body, /"story_era"/);
});

test('seeding upgrades the untouched smart-unified placeholder and its description to H3 style.system', () => {
  const smartUnified = SYSTEM_PRESETS.find(preset => preset.id === 'script-constraint-prefix-smart-unified');
  const published = new Map([[smartUnified.id, {
    ...smartUnified,
    description: '根据当前项目分析结果统一全片视觉风格的画面前缀词',
    body: [
      '【智能统一画面前缀】', '',
      '使用当前项目分析得到的统一视觉风格，作为全片所有镜头的画面前缀。', '',
      '统一控制：影像媒介、色彩体系、光线质感、明暗关系、对比度、饱和度、颗粒感、镜头质感、氛围和整体美术风格。', '',
      '不得改变或新增剧情、人物身份、人物造型、场景内容、道具、动作、台词和事件顺序。', '',
      '不要在最终分镜中单独输出“统一风格”标题，只需将该内容作为“画面前缀词”作用于每个镜头。', '',
      '如果没有有效的风格分析结果，则根据原文自动判断一种统一视觉风格，并保持全片一致。'
    ].join('\n'),
    version: 1,
    status: 'published'
  }]]);
  const drafts = new Map();
  const store = {
    getPublished: id => published.get(id) || null,
    createDraft: (_, input) => {
      const draft = { ...input, version: (published.get(input.id)?.version || 0) + 1 };
      drafts.set(input.id, draft);
      return draft;
    },
    publish: (_, id) => {
      const draft = drafts.get(id);
      published.set(id, { ...draft, status: 'published' });
      return published.get(id);
    }
  };
  seedSystemPresets(store, 'admin');
  const upgraded = published.get(smartUnified.id);
  assert.equal(upgraded.description, 'H3 style.system：按视频原文分析内容类型、时代与统一视觉风格');
  assert.match(upgraded.body, /H3 `style\.system`/);
  assert.match(upgraded.body, /"final_genre"/);
});

test('seeding upgrades the shipped Chinese-label smart-unified V3 body to the H3 style.system contract', () => {
  const smartUnified = SYSTEM_PRESETS.find(preset => preset.id === 'script-constraint-prefix-smart-unified');
  const published = new Map([[smartUnified.id, {
    ...smartUnified,
    version: 3,
    status: 'published',
    body: [
      '你是全片统一视觉分析师。',
      '最终将十一项视觉判断融合成一条紧凑、专业、可直接作为后续所有分镜画面提示词共同摄影前缀使用的：',
      '`统一风格附加视觉信息`',
      '【最终输出】',
      '内容类型：XXXX短剧',
      '统一风格附加视觉信息：XXXX'
    ].join('\n')
  }]]);
  const drafts = new Map();
  const store = {
    getPublished: id => published.get(id) || null,
    createDraft: (_, input) => {
      const draft = { ...input, version: (published.get(input.id)?.version || 0) + 1 };
      drafts.set(input.id, draft);
      return draft;
    },
    publish: (_, id) => {
      const draft = drafts.get(id);
      published.set(id, { ...draft, status: 'published' });
      return published.get(id);
    }
  };

  seedSystemPresets(store, 'admin');

  const upgraded = published.get(smartUnified.id);
  assert.equal(upgraded.version, 4);
  assert.match(upgraded.body, /H3 `style\.system`/);
  assert.match(upgraded.body, /"final_genre"/);
  assert.doesNotMatch(upgraded.body, /统一风格附加视觉信息：XXXX/);
});

test('seeding upgrades the untouched legacy Batch Factory video body to the complete script-preset composition', () => {
  const published = new Map([['batch-video-meta', {
    id: 'batch-video-meta', module: 'batch-factory', name: '视频提示词元提示词', kind: 'base',
    description: '批量工厂视频单元、镜头与整数时间轴规则', compatibleBaseIds: [], body: '你负责把导演结果组织成可执行的视频分镜描述。\n\n规则：旧版默认正文',
    protocolLock: { slot: 'batch.video-meta' }, version: 1, status: 'published'
  }]]);
  const drafts = new Map();
  const store = {
    getPublished: id => published.get(id) || null,
    createDraft: (_, input) => { const draft = { ...input, version: (published.get(input.id)?.version || 0) + 1 }; drafts.set(input.id, draft); return { id: draft.id, version: draft.version }; },
    publish: (_, id) => { published.set(id, { ...drafts.get(id), status: 'published' }); return published.get(id); }
  };
  seedSystemPresets(store, 'choushiyiguai');
  const upgraded = published.get('batch-video-meta');
  const body = upgraded.body;
  assert.equal(upgraded.name, '分镜元提示词（自动组合）');
  assert.equal(upgraded.description, '自动组合剧本生成的分段开头与分镜模式；批量工厂只保留 JSON 分镜输出协议');
  assert.match(body, /剧本生成预设：分段开头/);
  assert.match(body, /剧本生成预设：分镜模式/);
  assert.match(body, /Batch Factory V11 JSON 输出适配/);
  assert.doesNotMatch(body, /旧版默认正文/);
});

test('Batch Factory exposes the combined SD asset and video presets in their runtime slots', () => {
  const assetPreset = SYSTEM_PRESETS.find(item => item.id === 'batch-assets-sd-character-scene');
  const videoPreset = SYSTEM_PRESETS.find(item => item.id === 'batch-video-sd');
  assert.equal(assetPreset?.module, 'script');
  assert.equal(assetPreset?.protocolLock?.slot, 'script.asset-extraction');
  assert.equal(assetPreset?.protocolLock?.format, 'extract');
  assert.equal(videoPreset?.module, 'batch-factory');
  assert.equal(videoPreset?.protocolLock?.slot, 'batch.video-meta');
  assert.equal(videoPreset?.protocolLock?.operation, 'video-meta');
});

test('Batch Factory publishes the SD fixed-shot preset with the locked-lens video-meta contract', () => {
  const preset = SYSTEM_PRESETS.find(item => item.id === 'batch-video-sd-fixed-shot');
  assert.equal(preset?.module, 'batch-factory');
  assert.equal(preset?.name, 'SD 固定镜头提示词');
  assert.equal(preset?.description, '系统锁定镜头数量与时长，AI 只填画面内容');
  assert.equal(preset?.source, '批量工厂-SD固定镜头提示词.md');
  assert.equal(preset?.protocolLock?.format, 'json');
  assert.equal(preset?.protocolLock?.operation, 'video-meta');
  assert.equal(preset?.protocolLock?.slot, 'batch.video-meta');
});

test('SD fixed-shot preset body contains the locked-lens storyboard rules without constraint toggles', () => {
  const body = defaultBody('batch-video-sd-fixed-shot');
  // 原文理解规则：必须逐行理解原文
  assert.match(body, /逐行理解原文/);
  // 分镜输出规则：固定镜头、不得增加/删除/合并/修改时长
  assert.match(body, /固定镜头/);
  assert.match(body, /不得增加镜头/);
  assert.match(body, /不得删除镜头/);
  assert.match(body, /不得合并镜头/);
  assert.match(body, /不得修改时长/);
  // 不包含约束设置开关逻辑（不写“如果开启…就…”）
  assert.doesNotMatch(body, /如果开启/);
  assert.doesNotMatch(body, /如果打开/);
  // 不包含智能统一/画质/画面限制/负面提示词的“正文”（即这些其他预设的实际规则内容），
  // 只允许按名称提及由系统统一组装。
  // 智能统一正文特征：H3 style.system 协议字段
  assert.doesNotMatch(body, /H3 style\.system/);
  assert.doesNotMatch(body, /final_genre/);
  assert.doesNotMatch(body, /style\.system/);
  // 画质约束正文特征：4K 画质条款与 H3 镜头叙事约束
  assert.doesNotMatch(body, /4K 超高清/);
  assert.doesNotMatch(body, /严格依次呈现列出的场景与镜头/);
  // 画面限制正文特征：跨镜人物稳定限制与无关文字限制
  assert.doesNotMatch(body, /人物身份、年龄、脸型/);
  assert.doesNotMatch(body, /禁止出现任何无关文字/);
  // 负面提示词正文特征：禁止字幕条款
  assert.doesNotMatch(body, /禁止字幕、台词字幕/);
});

test('SD fixed-shot preset keeps all prior batch presets unchanged', () => {
  const expectedIds = [
    'batch-hook-adaptation',
    'batch-original-director',
    'batch-viral-director',
    'batch-character-meta',
    'batch-character-h3',
    'batch-assets-h3-skill',
    'batch-scene-meta',
    'batch-scene-h3',
    'batch-prop-meta',
    'batch-video-meta',
    'batch-video-h3-director',
    'batch-visual-meta'
  ];
  for (const id of expectedIds) {
    const preset = SYSTEM_PRESETS.find(item => item.id === id);
    assert.ok(preset, `missing prior batch preset: ${id}`);
    assert.equal(preset.module, 'batch-factory');
  }
  // 新预设放在 batch-visual-meta 之后
  const visualMeta = SYSTEM_PRESETS.findIndex(item => item.id === 'batch-visual-meta');
  const sdFixed = SYSTEM_PRESETS.findIndex(item => item.id === 'batch-video-sd-fixed-shot');
  assert.ok(visualMeta >= 0, 'batch-visual-meta missing');
  assert.ok(sdFixed >= 0, 'batch-video-sd-fixed-shot missing');
  assert.equal(sdFixed, visualMeta + 1, 'SD fixed-shot preset should follow batch-visual-meta in SYSTEM_PRESETS');
});

test('Batch Factory publishes the opening-variant meta prompt preset in its dedicated slot', () => {
  const preset = SYSTEM_PRESETS.find(item => item.id === 'batch-opening-meta');
  assert.ok(preset, 'missing batch preset: batch-opening-meta');
  assert.equal(preset.module, 'batch-factory');
  assert.equal(preset.name, '换开头元提示词');
  assert.equal(preset.source, '批量工厂-换开头元提示词.md');
  assert.equal(preset.protocolLock?.format, 'opening-meta');
  assert.equal(preset.protocolLock?.slot, 'batch.opening-meta');
  assert.match(preset.body || '', /===VARIANT 1===/);
  assert.match(preset.body || '', /时长：X秒/);
});
