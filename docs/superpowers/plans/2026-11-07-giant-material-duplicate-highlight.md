# 巨量获取重复书籍高亮 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在巨量素材解析列表中，把后续重复 Book ID 的行显示为橙黄色并标注“重复，将跳过”。

**Architecture:** 在现有队列工具里提供独立的重复位置计算函数，界面只消费计算结果并渲染样式。重复判断不改变提交数据和后端逻辑，书籍选择变化时由 React 重新计算。

**Tech Stack:** React 18、Ant Design 5、原生 Node.js test、CSS

## Global Constraints

- 第一次出现的 Book ID 正常显示，后续相同 Book ID 才标记。
- 重复项保留展示，不从列表删除。
- 标签文案固定为“重复，将跳过”。
- 不按书名去重。
- 本次只修改本地源码，不部署线上。

---

### Task 1: 重复位置计算

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/batchFactoryGiantMaterialQueue.js`
- Test: `frontend/src/user/pages/shuihuo/batchFactoryGiantMaterialQueue.test.js`

**Interfaces:**
- Consumes: `entries: Array<{ platformBookId?: string, bookId?: string }>`
- Produces: `duplicateGiantMaterialBookIndexes(entries): Set<number>`

- [ ] **Step 1: 写失败测试**

```js
test('marks only later rows with the same resolved Book ID as duplicates', () => {
  const duplicates = duplicateGiantMaterialBookIndexes([
    { platformBookId: '7703405385419328574' },
    { platformBookId: '7703405385419328574' },
    { platformBookId: 'other' },
    { platformBookId: '' }
  ]);
  assert.deepEqual([...duplicates], [1]);
});
```

- [ ] **Step 2: 运行测试并确认因缺少函数而失败**

Run: `node --test src/user/pages/shuihuo/batchFactoryGiantMaterialQueue.test.js`

- [ ] **Step 3: 实现最小纯函数**

```js
export function duplicateGiantMaterialBookIndexes(entries = []) {
  const seen = new Set();
  const duplicates = new Set();
  entries.forEach((entry, index) => {
    const bookId = text(entry?.platformBookId || entry?.bookId);
    if (!bookId) return;
    if (seen.has(bookId)) duplicates.add(index);
    else seen.add(bookId);
  });
  return duplicates;
}
```

- [ ] **Step 4: 运行测试并确认通过**

Run: `node --test src/user/pages/shuihuo/batchFactoryGiantMaterialQueue.test.js`

### Task 2: 重复行视觉提示

**Files:**
- Modify: `frontend/src/user/pages/shuihuo/BatchFactoryCreateModal.jsx`
- Modify: `frontend/src/user/pages/shuihuo-production.css`
- Test: `frontend/src/user/pages/shuihuo/BatchFactoryCreateModal.source.test.js`

**Interfaces:**
- Consumes: `duplicateGiantMaterialBookIndexes(entries)`
- Produces: `.batch-factory-giant-item.is-duplicate` 行样式和“重复，将跳过”标签

- [ ] **Step 1: 写失败测试**

```js
test('highlights later giant material rows that resolve to a duplicate Book ID', () => {
  assert.match(source, /duplicateGiantMaterialBookIndexes/);
  assert.match(source, /重复，将跳过/);
  assert.match(stylesheet, /batch-factory-giant-item\.is-duplicate/);
});
```

- [ ] **Step 2: 运行测试并确认缺少视觉提示而失败**

Run: `node --test src/user/pages/shuihuo/BatchFactoryCreateModal.source.test.js`

- [ ] **Step 3: 接入重复计算并渲染提示**

在 `giantItems` 的当前选中书籍列表上调用纯函数；重复行增加 `is-duplicate` 类，并渲染 `<Tag className="batch-factory-giant-duplicate-tag">重复，将跳过</Tag>`。

- [ ] **Step 4: 增加橙黄色样式**

为重复行增加橙黄色半透明背景、边框和标签颜色；保持错误行红色样式优先。

- [ ] **Step 5: 运行聚焦测试**

Run: `node --test src/user/pages/shuihuo/batchFactoryGiantMaterialQueue.test.js src/user/pages/shuihuo/BatchFactoryCreateModal.source.test.js`

- [ ] **Step 6: 运行前端构建**

Run: `npm run build`
