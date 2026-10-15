// 改文工作台：后端路由（工厂函数）
// 把 Task 1-6 的模块（parse / classifier / tasks / rewrite / config / ai）串成 HTTP 接口，
// 由 app.js 挂载到 /api/novel-fetch-workshop，供前端改文工作台（Task 8）调用。
// 鉴权用 middleware/auth 的 apiAuth；tasks / configStore 为必填依赖（测试可注入 mock）。
const express = require('express');
const { apiAuth } = require('../middleware/auth');
const aiModule = require('../lib/novel-fetch-workshop/ai');
const rulesModule = require('../lib/novel-fetch-workshop/rules');
const { createKnowledgeStore } = require('../lib/novel-fetch-workshop/knowledge');
const { createOpeningStore } = require('../lib/novel-fetch-workshop/opening');
const { createMySQLWorkshopStore } = require('../lib/novel-fetch-workshop/mysql-store');
const { merge121BookInfoIntoMeta } = require('../lib/novel-fetch-workshop/121-bookinfo');

function mergeConfig(current, patch) {
  const merged = { ...(current || {}) };
  for (const [key, value] of Object.entries(patch || {})) {
    merged[key] = value && typeof value === 'object' && !Array.isArray(value)
      ? { ...(current?.[key] || {}), ...value }
      : value;
  }
  return merged;
}

function visibleWorkshopPlatforms(configStore) {
  const platforms = typeof configStore?.getPlatforms === 'function' ? configStore.getPlatforms() : [];
  if (!Array.isArray(platforms)) return [];
  return platforms.flatMap(platform => {
    if (!platform || platform.visible === false) return [];
    const id = String(platform.id ?? '').trim();
    const name = String(platform.name ?? '').trim();
    return id && name ? [{ id, name }] : [];
  });
}

// 并发限制执行器：把 items 按 limit 并发执行 worker，单任务异常不影响整体（结果按原序返回）
async function runWithConcurrency(items, limit, worker) {
  const results = new Array(items.length);
  let index = 0;
  const runner = async () => {
    while (index < items.length) {
      const current = index++;
      try {
        results[current] = { ok: true, value: await worker(items[current]) };
      } catch (error) {
        results[current] = { ok: false, value: null, error };
      }
    }
  };
  const count = Math.max(1, Math.floor(Number(limit) || 1));
  const runners = [];
  for (let i = 0; i < Math.min(count, items.length); i++) runners.push(runner());
  await Promise.all(runners);
  return results;
}

function createNovelFetchWorkshopRouter({
  auth = apiAuth,
  tasks: injectedTasks,
  configStore: injectedConfigStore,
  targetBaseUrl,
  bridgeSecret,
  resolveRuntimeModel,
  classifier = require('../lib/novel-fetch-workshop/classifier'),
  rewrite = require('../lib/novel-fetch-workshop/rewrite'),
  parse = require('../lib/novel-fetch-workshop/parse'),
  systemDir,
  knowledgeStore: injectedKnowledgeStore,
  openingStore: injectedOpeningStore,
  rules = rulesModule,
  sourceHook,
  processRouter
} = {}) {
  const router = express.Router();
  router.use(auth);
  if (processRouter) {
    router.use((req, res, next) => {
      if (!req.path.startsWith('/process/')) return next();
      return processRouter(req, res, next);
    });
  }

  async function resources(req) {
    const tasks = injectedTasks || createMySQLWorkshopStore({ targetBaseUrl, bridgeSecret, account: req.auth?.account });
    if (injectedTasks && injectedConfigStore) {
      return { tasks, config: await injectedConfigStore.getConfig(), configStore: injectedConfigStore };
    }
    const config = await tasks.getConfig();
    const configStore = injectedConfigStore || {
      getConfig: () => config,
      getAiConfig: () => ({
        ai: config.ai || {},
        ai_presets: config.ai_presets || [],
        ai_assignments: config.ai_assignments || {},
        text_model_id: config.text_model_id || config.textModelId || config.app_config?.text_model_id || config.app_config?.textModelId || ''
      }),
      resolveRuntimeModel: modelId => {
        if (typeof resolveRuntimeModel !== 'function') return null;
        return resolveRuntimeModel(req.username, 'text', modelId);
      },
      getPlatforms: tasks.getPlatforms,
      getStyles: tasks.getStyles
    };
    return { tasks, config, configStore };
  }

  const knowledge = injectedKnowledgeStore || (systemDir ? createKnowledgeStore({ systemDir }) : null);
  const opening = injectedOpeningStore || (systemDir ? createOpeningStore({ systemDir, styles: [] }) : null);

  async function fetchOriginal(tasks, username, bookId, maxTxt) {
    const hook = sourceHook || tasks?.sourceHook;
    let result;
    if (typeof hook === 'function') result = await hook(username, bookId, maxTxt);
    else if (hook && typeof hook.fetchOriginal === 'function') result = await hook.fetchOriginal(username, bookId, maxTxt);
    else result = await tasks.fetchOriginal(username, bookId, maxTxt);

    // 121 的 bookinfo 是低成本、确定性的分类来源。抓取成功后先把
    // category/genre/可明确推导的男女频写回任务，AI 只补剩余缺口。
    if (result && result.status !== 'failed' && typeof tasks?.getTask === 'function' && typeof tasks?.saveTasks === 'function') {
      const current = await tasks.getTask(username, bookId);
      if (current?.meta) {
        const bookinfo = result.bookinfo || current.meta.bookinfo || {};
        const enriched = merge121BookInfoIntoMeta(current.meta, bookinfo);
        await tasks.saveTasks(username, [{ ...enriched, bookId: enriched.bookId || bookId }]);
      }
    }
    return result;
  }

  function hasSourceFetcher(tasks) {
    const hook = sourceHook || tasks?.sourceHook;
    return typeof hook === 'function' || typeof hook?.fetchOriginal === 'function' || typeof tasks?.fetchOriginal === 'function';
  }

  function readKnowledge(kind) {
    if (!knowledge) throw new Error('知识库未启用（缺少 systemDir）');
    return knowledge.list(kind);
  }

  // POST /process：解析批量清单 → 保存任务 →（可选）抓 121 →（可选）AI 补分类 →（可选）AI 改文
  router.post('/process', async (req, res) => {
    try {
      const username = req.username;
      const { tasks, configStore } = await resources(req);
      const { platformId = '2', inputText, parseMode = 'smart', columnPresetId, columnOrder = '', maxTxt, aiCount } = req.body || {};
      const config = configStore.getConfig() || {};
      const workflow = config.workflow || {};

      // 1. 解析批量清单
      const parsedResult = parse.parseBooks({
        inputText,
        parseMode,
        columnPresetId,
        columnOrder,
        styles: configStore.getStyles()
      });
      const parsedTasks = parsedResult.tasks || [];

      // 2. 任务携带平台信息与抓取/改文参数
      const platforms = configStore.getPlatforms() || [];
      const platform = platforms.find(item => item && String(item.id) === String(platformId));
      const platformName = platform ? platform.name : platformId;
      const defaultMaxTxt = (config.fetch && config.fetch.default_max_txt) || 4000;
      const defaultAiCount = (config.rewrite && config.rewrite.default_ai_count) || 1;
      const resolvedMaxTxt = maxTxt != null ? maxTxt : defaultMaxTxt;
      const resolvedAiCount = aiCount != null ? aiCount : defaultAiCount;
      for (const task of parsedTasks) {
        task.platformId = platformId;
        task.platformName = platformName;
        task.maxTxt = resolvedMaxTxt;
        task.aiCount = resolvedAiCount;
      }
      let tasksToProcess = parsedTasks;
      let classifyErrors = [];

      // 3. 先保存任务，121 抓取器需要从持久化任务读取平台与参数。
      if (tasksToProcess.length) await tasks.saveTasks(username, tasksToProcess);

      // 4. 可选：先抓 121。fetchOriginal 会把 bookinfo.category/genre 及可明确
      //    推导出的男女频写回任务；用户手填的 gender 永远不会被覆盖。
      let fetched = 0;
      let fetchFailed = 0;
      const fetchedDone = new Set();
      const concurrency = (config.fetch && config.fetch.concurrency) || 4;
      if (workflow.auto_fetch_original && tasksToProcess.length) {
        const results = await runWithConcurrency(tasksToProcess, concurrency, async task => {
          const result = await fetchOriginal(tasks, username, task.bookId, task.maxTxt);
          return { task, status: (result && result.status) || 'failed' };
        });
        for (const item of results) {
          if (item.ok && item.value && item.value.status === 'done') {
            fetched++;
            fetchedDone.add(item.value.task.bookId);
          } else {
            fetchFailed++;
          }
        }

        // 抓取后的持久化 meta 才是 121 分类后的权威输入，回读后再交给 AI。
        const refreshed = [];
        for (const task of tasksToProcess) {
          const current = await tasks.getTask(username, task.bookId);
          refreshed.push(current?.meta ? { ...task, ...current.meta } : task);
        }
        tasksToProcess = refreshed;
      }

      // 5. AI 只处理 121 之后仍缺失的男女频/风格；风格仍按原固定风格表判断。
      if (workflow.auto_classify_missing && tasksToProcess.length) {
        const classifyResult = await classifier.classifyMissingRows({ configStore, tasks: tasksToProcess });
        classifyErrors = (classifyResult && classifyResult.errors) || [];
        tasksToProcess = (classifyResult && classifyResult.tasks) || tasksToProcess;
        await tasks.saveTasks(username, tasksToProcess);
      }

      // 6. 可选：对原文成功的任务逐个生成 AI 改文版本（单任务失败不中断整批）
      //    改文失败并入 fetchFailed 计数：响应契约固定 9 字段，无法新增字段，
      //    故把改文异常计入 fetchFailed（语义扩展为"处理失败计数"，与 generatedAiFiles 的成功计数配对），
      //    成功仍累计到 generatedAiFiles；任一任务抛异常（如 AI 版本文件磁盘写失败）只计数不中断后续任务。
      let generatedAiFiles = 0;
      if (workflow.auto_rewrite_after_fetch && tasksToProcess.length) {
        for (const task of tasksToProcess) {
          if (!fetchedDone.has(task.bookId)) continue;
          try {
            const current = await tasks.getTask(username, task.bookId);
            if (!current || !current.meta || current.meta.originalStatus !== 'done') continue;
            const result = await rewrite.generateAiVersions({
              configStore,
              tasks,
              username,
              task: current.meta,
              count: task.aiCount
            });
            generatedAiFiles += (result && result.generated || []).filter(item => item && item.status === 'done').length;
          } catch (error) {
            // 单任务改文异常（如 AI 版本文件磁盘写失败）只计数，不中断后续任务
            fetchFailed++;
          }
        }
      }

      const list = await tasks.listTasks(username);
      return res.json({
        parsed: parsedResult.parsed,
        emptyIdCount: parsedResult.emptyIdCount,
        uniqueTasks: parsedResult.uniqueTasks,
        duplicateCount: parsedResult.duplicateCount,
        classifyErrors,
        fetched,
        fetchFailed,
        generatedAiFiles,
        tasks: list || []
      });
    } catch (error) {
      return res.status(500).json({ error: error.message || 'Internal server error' });
    }
  });

  // GET /tasks：任务列表
  router.get('/tasks', async (req, res) => {
    try {
      const { tasks } = await resources(req);
      const list = await tasks.listTasks(req.username);
      return res.json({ tasks: list || [] });
    } catch (error) {
      return res.status(500).json({ error: error.message || '获取任务列表失败' });
    }
  });

  // DELETE /tasks：批量删除任务（body { ids: [] }），返回删除统计与最新任务列表
  router.delete('/tasks', async (req, res) => {
    try {
      const username = req.username;
      const { tasks } = await resources(req);
      const ids = (req.body && Array.isArray(req.body.ids)) ? req.body.ids : [];
      const result = await tasks.deleteTasks(username, ids);
      const list = await tasks.listTasks(username);
      return res.json({
        ok: true,
        requested: result.requested,
        deleted: result.deleted,
        results: (result && result.results) || [],
        tasks: list || []
      });
    } catch (error) {
      return res.status(500).json({ ok: false, error: error.message || '删除任务失败' });
    }
  });

  // POST /tasks/batch-retry：兼容本地工作台的批量补跑；桥接存储仍可提供自己的实现。
  router.post('/tasks/batch-retry', async (req, res) => {
    try {
      const username = req.username;
      const { tasks, configStore } = await resources(req);
      const ids = Array.isArray(req.body?.ids) ? req.body.ids : [];
      if (typeof tasks.batchRetry === 'function') {
        const result = await tasks.batchRetry(username, ids, {
          fetchOriginal: (user, id, maxTxt) => fetchOriginal(tasks, user, id, maxTxt),
          generateAi: async (user, id, count) => {
            const current = await tasks.getTask(user, id);
            if (!current?.meta || current.meta.originalStatus !== 'done') return { status: 'skipped' };
            return rewrite.generateAiVersions({ configStore, tasks, username: user, task: current.meta, count });
          }
        });
        return res.json({ ok: true, ...result, tasks: await tasks.listTasks(username) });
      }
      for (const id of ids) {
        const current = await tasks.getTask(username, id);
        if (current && hasSourceFetcher(tasks)) {
          await fetchOriginal(tasks, username, id, current.meta?.maxTxt || 4000);
        }
      }
      return res.json({ ok: true, tasks: await tasks.listTasks(username) });
    } catch (error) {
      return res.status(500).json({ ok: false, error: error.message || '批量重试失败' });
    }
  });

  // POST /tasks/:bookId/restore-original：恢复原始备份。
  router.post('/tasks/:bookId/restore-original', async (req, res) => {
    try {
      const username = req.username;
      const { tasks } = await resources(req);
      const bookId = String(req.params.bookId || '').trim();
      const current = await tasks.getTask(username, bookId);
      if (!current) return res.status(404).json({ ok: false, error: '任务不存在' });
      const result = await tasks.restoreOriginal(username, bookId);
      return res.json({ ok: true, task: await tasks.getTask(username, bookId), result });
    } catch (error) {
      return res.status(400).json({ ok: false, error: error.message || '恢复原文失败' });
    }
  });

  // GET /tasks/:bookId：任务详情（meta + 清洗后原文 + 原始备份存在性 + 日志）
  router.get('/tasks/:bookId', async (req, res) => {
    try {
      const username = req.username;
      const { tasks } = await resources(req);
      const bookId = String(req.params.bookId || '').trim();
      const task = await tasks.getTask(username, bookId);
      if (!task) return res.status(404).json({ ok: false, error: '任务不存在' });
      const original = await tasks.readOriginal(username, bookId);
      const logs = await tasks.readLogs(username, bookId);
      return res.json({
        meta: task.meta || {},
        original,
        hasOriginalRaw: task.hasOriginalRaw === true,
        logs: logs || []
      });
    } catch (error) {
      return res.status(500).json({ error: error.message || '获取任务详情失败' });
    }
  });

  // POST /tasks/:bookId/fetch：抓取单本原文
  router.post('/tasks/:bookId/fetch', async (req, res) => {
    try {
      const username = req.username;
      const { tasks } = await resources(req);
      const bookId = String(req.params.bookId || '').trim();
      const task = await tasks.getTask(username, bookId);
      if (!task) return res.status(400).json({ ok: false, error: '任务不存在' });
      const { maxTxt } = req.body || {};
      const resolvedMaxTxt = maxTxt != null ? maxTxt : ((task.meta && task.meta.maxTxt) || 4000);
      const result = await fetchOriginal(tasks, username, bookId, resolvedMaxTxt);
      const current = await tasks.getTask(username, bookId);
      if (!result || result.status === 'failed') {
        const errorText = (current && current.meta && current.meta.error) || '抓取失败';
        return res.status(400).json({ ok: false, error: errorText });
      }
      return res.json({ task: current });
    } catch (error) {
      return res.status(500).json({ ok: false, error: error.message || '抓取失败' });
    }
  });

  // POST /tasks/:bookId/generate-ai：生成 AI 改文版本（数量 1~20）
  router.post('/tasks/:bookId/generate-ai', async (req, res) => {
    try {
      const username = req.username;
      const { tasks, configStore } = await resources(req);
      const bookId = String(req.params.bookId || '').trim();
      const task = await tasks.getTask(username, bookId);
      if (!task) return res.status(404).json({ ok: false, error: '任务不存在' });
      let { count = 1 } = req.body || {};
      count = Math.floor(Number(count));
      if (!Number.isFinite(count)) count = 1;
      if (count < 1 || count > 20) return res.status(400).json({ ok: false, error: '生成数量需为 1~20 的整数' });
      await rewrite.generateAiVersions({ configStore, tasks, username, task: task.meta, count });
      const current = await tasks.getTask(username, bookId);
      return res.json({ task: current });
    } catch (error) {
      return res.status(500).json({ ok: false, error: error.message || '生成失败' });
    }
  });

  // Batch Factory uses this lightweight endpoint before opening the intake form.
  router.get('/platforms', async (req, res) => {
    try {
      const { configStore } = await resources(req);
      return res.json({ platforms: visibleWorkshopPlatforms(configStore) });
    } catch (error) {
      return res.status(500).json({ error: error.message || '读取平台表失败' });
    }
  });

  // GET /config：读取工作台配置（主配置 / 平台表 / 风格表 / AI 配置）
  router.get('/config', async (req, res) => {
    try {
      const { tasks, configStore } = await resources(req);
      return res.json({
        appConfig: configStore.getConfig(),
        platforms: configStore.getPlatforms(),
        styles: configStore.getStyles(),
        aiConfig: typeof configStore.getAiConfig === 'function' ? configStore.getAiConfig() : {}
      });
    } catch (error) {
      return res.status(500).json({ error: error.message || '读取配置失败' });
    }
  });

  // 规则、知识库和爆款开头接口由本地工作台提供；桥接模式未配置时明确返回不可用。
  router.post('/rules/preview', async (req, res) => {
    try {
      const { configStore } = await resources(req);
      const { text = '', scope = 'original', layoutConfig = {}, knowledge: suppliedKnowledge } = req.body || {};
      const knowledgeData = suppliedKnowledge || (knowledge ? {
        layout_rules: readKnowledge('layout_rules'),
        symbol_rules: readKnowledge('symbol_rules'),
        chapter_rules: readKnowledge('chapter_rules')
      } : {});
      const trace = rules.processDocumentTrace(text, scope, layoutConfig, knowledgeData);
      const result = rules.processDocumentText(text, scope, layoutConfig, knowledgeData);
      return res.json({ ok: true, text: result, result, trace, config: configStore.getConfig() });
    } catch (error) {
      return res.status(400).json({ ok: false, error: error.message || '规则排版预览失败' });
    }
  });

  router.get('/knowledge/summary', (req, res) => {
    try {
      if (!knowledge) return res.status(500).json({ ok: false, error: '知识库未启用（缺少 systemDir）' });
      return res.json({ ok: true, summary: knowledge.getSummary() });
    } catch (error) {
      return res.status(500).json({ ok: false, error: error.message || '读取知识库汇总失败' });
    }
  });

  router.get('/knowledge/:kind', (req, res) => {
    try {
      const data = readKnowledge(req.params.kind);
      return res.json({ ok: true, kind: req.params.kind, data, ...(data && typeof data === 'object' ? data : {}) });
    } catch (error) {
      return res.status(400).json({ ok: false, error: error.message || '读取知识库失败' });
    }
  });

  router.post('/knowledge/:kind', (req, res) => {
    try {
      if (!knowledge) return res.status(500).json({ ok: false, error: '知识库未启用（缺少 systemDir）' });
      return res.json(knowledge.save(req.params.kind, req.body?.item || req.body));
    } catch (error) {
      return res.status(400).json({ ok: false, error: error.message || '保存知识条目失败' });
    }
  });

  router.delete('/knowledge/:kind/:id', (req, res) => {
    try {
      if (!knowledge) return res.status(500).json({ ok: false, error: '知识库未启用（缺少 systemDir）' });
      return res.json(knowledge.remove(req.params.kind, req.params.id));
    } catch (error) {
      return res.status(400).json({ ok: false, error: error.message || '删除知识条目失败' });
    }
  });

  router.post('/knowledge/:kind/:id/optimize', async (req, res) => {
    try {
      if (!knowledge) return res.status(500).json({ ok: false, error: '知识库未启用（缺少 systemDir）' });
      const { configStore } = await resources(req);
      const settings = aiModule.resolveAiSettings(configStore, 'rewrite');
      return res.json(await knowledge.optimizeItem({ kind: req.params.kind, id: req.params.id }, aiModule, settings));
    } catch (error) {
      return res.status(400).json({ ok: false, error: error.message || '优化知识条目失败' });
    }
  });

  router.post('/opening/analyze', async (req, res) => {
    try {
      if (!opening) return res.status(500).json({ ok: false, error: '开头词库未启用（缺少 systemDir）' });
      const { configStore } = await resources(req);
      const settings = aiModule.resolveAiSettings(configStore, 'rewrite');
      return res.json(await opening.analyze(aiModule, settings, req.body?.text));
    } catch (error) {
      return res.status(400).json({ ok: false, error: error.message || '开头分析失败' });
    }
  });

  router.post('/opening/save', (req, res) => {
    try {
      if (!opening) return res.status(500).json({ ok: false, error: '开头词库未启用（缺少 systemDir）' });
      return res.json(opening.save(req.body?.item || req.body));
    } catch (error) {
      return res.status(400).json({ ok: false, error: error.message || '保存开头词失败' });
    }
  });

  router.post('/opening/normalize', (req, res) => {
    try {
      if (!opening) return res.status(500).json({ ok: false, error: '开头词库未启用（缺少 systemDir）' });
      return res.json(opening.normalize());
    } catch (error) {
      return res.status(400).json({ ok: false, error: error.message || '规范化开头词失败' });
    }
  });

  // POST /config：保存工作台配置（一期仅持久化 appConfig；platforms/styles 忽略）
  router.post('/config', async (req, res) => {
    try {
      const { appConfig } = req.body || {};
      const { tasks, config } = await resources(req);
      if (appConfig && typeof appConfig === 'object') {
        await tasks.saveConfig(mergeConfig(config, appConfig));
      }
      return res.json({ ok: true, config: appConfig && typeof appConfig === 'object' ? mergeConfig(config, appConfig) : config });
    } catch (error) {
      return res.status(500).json({ ok: false, error: error.message || '保存配置失败' });
    }
  });

  // POST /ai/test：AI 接口连通性测试
  router.post('/ai/test', async (req, res) => {
    try {
      const { purpose = 'rewrite' } = req.body || {};
      const { configStore } = await resources(req);
      const settings = aiModule.resolveAiSettings(configStore, purpose);
      const response = await aiModule.chatCompletion(settings, [
        { role: 'system', content: '你是接口连通性测试助手。' },
        { role: 'user', content: '请只回复：ok' }
      ], { temperature: 0 });
      const text = (response && response.text) || '';
      return res.json({ ok: true, content: text.slice(0, 200) });
    } catch (error) {
      return res.status(400).json({ ok: false, error: error.message || 'AI 接口测试失败' });
    }
  });

  return router;
}

module.exports = { createNovelFetchWorkshopRouter, visibleWorkshopPlatforms };
