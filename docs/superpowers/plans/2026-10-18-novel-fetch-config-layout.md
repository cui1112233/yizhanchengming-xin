# 小说获取配置页布局整理 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将三个长提示词字段从“工作流与原文抓取”移到独立的“提示词配置”分组，减少配置页无效留白并保持现有保存行为。

**Architecture:** 继续复用 `NovelFetchConfigPanel` 的单一 Ant Design Form，只增加一个展示分组，不拆分数据状态。样式沿用现有三栏工作区与主题变量，桌面端保持左右栏吸顶，窄屏继续转为单列。

**Tech Stack:** React、Ant Design、CSS、Node.js 内置测试、Go embed 本地发布链路

## Global Constraints

- 不删除配置项，不改变 `rewrite.prompt`、`rewrite.processing_rule_prompt`、`knowledge.usage_prompt` 字段路径。
- 不修改 Go 接口、MySQL 配置结构、默认值、保存方法或本地存储。
- 深色和浅色主题继续使用现有 CSS 变量。
- 不清理、重置或覆盖工作区中的其他修改。
- 不发布公网。

---

### Task 1: 配置分组结构和回归测试

**Files:**
- Modify: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`
- Modify: `frontend/src/user/pages/NovelFetchPage.jsx:214-315`
- Modify: `frontend/src/user/pages/novel-fetch-subpages.css:136-154`

**Interfaces:**
- Consumes: `NovelFetchConfigPanel({ form, effectiveValues, dirty, onDirty, versionPanel, section, onSectionChange })`
- Produces: 新分组键 `prompts`，继续使用原有三个 Form 字段路径和同一个 `form` 实例。

- [x] **Step 1: 写入失败测试**

在 `novel-fetch-subpages.layout.test.js` 中截取 `NovelFetchConfigPanel` 源码范围，并新增测试：

```js
test('配置页将长提示词放进独立分组并保持摘要吸顶', () => {
  const panelStart = source.indexOf('function NovelFetchConfigPanel');
  const panelEnd = source.indexOf('function TaskDetailDrawer');
  const panelSource = source.slice(panelStart, panelEnd);
  const workflowStart = panelSource.indexOf('<section hidden={section !== \'workflow\'}>');
  const promptsStart = panelSource.indexOf('<section hidden={section !== \'prompts\'}>');
  const supplementStart = panelSource.indexOf('<section hidden={section !== \'supplement\'}>');
  const workflowSource = panelSource.slice(workflowStart, promptsStart);
  const promptsSource = panelSource.slice(promptsStart, supplementStart);

  assert.match(panelSource, /key: 'prompts', label: '提示词配置'/);
  assert.doesNotMatch(workflowSource, /改文提示词|处理规则提示词|知识库使用提示词/);
  assert.match(promptsSource, /name=\{\['rewrite', 'prompt'\]\}/);
  assert.match(promptsSource, /name=\{\['rewrite', 'processing_rule_prompt'\]\}/);
  assert.match(promptsSource, /name=\{\['knowledge', 'usage_prompt'\]\}/);
  assert.match(css, /\.novel-fetch-config-summary\s*\{[^}]*position:\s*sticky/s);
});
```

- [x] **Step 2: 运行测试并确认先失败**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: 新测试因缺少 `prompts` 分组和对应 section 而失败，既有测试继续通过。

- [x] **Step 3: 最小实现独立提示词分组**

在 `NovelFetchConfigPanel` 的 `sectionItems` 中把以下项目插入 `workflow` 之后：

```js
{ key: 'prompts', label: '提示词配置', description: '改文、规则和知识库提示词' }
```

从 `workflow` section 删除三个现有 `Form.Item`，并在它与 `supplement` section 之间新增：

```jsx
<section hidden={section !== 'prompts'}>
  <Typography.Title level={5}>提示词配置</Typography.Title>
  <div className="novel-fetch-config-prompts">
    <Form.Item name={['rewrite', 'prompt']} label="改文提示词"><Input.TextArea rows={6} /></Form.Item>
    <Form.Item name={['rewrite', 'processing_rule_prompt']} label="处理规则提示词"><Input.TextArea rows={6} /></Form.Item>
    <Form.Item name={['knowledge', 'usage_prompt']} label="知识库使用提示词"><Input.TextArea rows={5} /></Form.Item>
  </div>
</section>
```

在 `novel-fetch-subpages.css` 中增加只负责编辑区间距的样式：

```css
.novel-fetch-config-prompts { display:flex; flex-direction:column; gap:2px; }
.novel-fetch-config-prompts .ant-form-item { margin-bottom:12px; }
```

- [x] **Step 4: 运行定向测试并确认通过**

Run: `node --test frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: 全部通过。

---

### Task 2: 完整构建和本地页面验收

**Files:**
- Verify: `frontend/src/user/pages/NovelFetchPage.jsx`
- Verify: `frontend/src/user/pages/novel-fetch-subpages.css`
- Verify: `frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

**Interfaces:**
- Consumes: Task 1 的 `prompts` 分组与现有 Go embed 构建脚本。
- Produces: 可在 `http://localhost:4001/novel-fetch?tab=config` 验收的本地 Go 嵌入版本。

- [x] **Step 1: 运行完整前端测试**

Run: `npm --prefix frontend test`

Expected: 所有测试通过，无新回归。

- [x] **Step 2: 构建 React 前端**

Run: `npm --prefix frontend run build`

Expected: 构建成功，未出现语法或资源错误。

- [x] **Step 3: 构建并启动本地 Go 嵌入版本**

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-go-local.ps1`

确认 `4001` 当前监听进程属于本项目后，再运行：

Run: `powershell -ExecutionPolicy Bypass -File scripts/start-local-go-review.ps1 -UseYcmMysql`

Run: `powershell -ExecutionPolicy Bypass -File scripts/check-local-go.ps1`

Expected: 健康检查和小说获取路由检查通过。

- [x] **Step 4: 在真实页面验收**

打开 `http://localhost:4001/novel-fetch?tab=config`，逐项确认：

- 左侧“提示词配置”可点击，三个提示词字段均存在。
- “工作流与原文抓取”不再显示三个长文本框。
- 在两个分组间来回切换后，未保存输入仍保留。
- 右侧“当前选择 / 已生效”滚动时保持可见。
- 深色和浅色主题下无文字、边框或输入框颜色异常。
- 窄屏下页面转为单列，无横向溢出或控件重叠。
- 保存按钮和未保存状态正常，浏览器控制台无新增错误。

- [x] **Step 5: 检查改动边界**

Run: `git diff --check -- frontend/src/user/pages/NovelFetchPage.jsx frontend/src/user/pages/novel-fetch-subpages.css frontend/src/user/pages/novel-fetch-subpages.layout.test.js`

Expected: 无空白错误；只报告本轮预期源码差异。由于这些文件包含前序未提交修改，本轮不创建混合提交，也不发布公网。
