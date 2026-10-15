# 编码风格约定

## 核心原则

把“为什么”看得比“是什么”更重要：代码靠命名自明，决策和踩坑靠注释与 commit 沉淀。写代码、写注释、写测试和提交 commit 都遵循本文。

## 1. 注释只写 WHY，不写 WHAT

代码做了什么应由命名表达。注释只解释为什么这样做，尤其是非显而易见的约束、上游怪癖和已经踩过的坑，并指名道姓写清楚原因。

```go
// MediaURLExpiry: 设为 1 天——足够覆盖一次正常用户会话，又短到让泄露
// 出去的 URL 很快失效。

// 跳过 ID 为空的 tool_use 块——这是有 bug 的 stream parser（例如
// gpt-5.5 + eino-ext）留下的残骸；否则会以 id="" 重放成 assistant
// tool_call，在下一个严格模式的 backbone（doubao 等，要求 tool_call_id 非空）
// 那里校验失败。
```

不要写 `// 遍历 nodes` 这类复述代码的注释。应说明“这里为什么不用 X 方案”以及“这个值为什么是 20h 而不是 24h”。注释是给未来维护者看的事故档案。技术名词（类型名、`media://`、`tool_call_id` 等）保留英文。

## 2. Go 后端

- **窄接口，贴近消费点**：只声明调用方需要的子集，方便测试 stub。

  ```go
  // GenerationService is the subset of *generation.Service that DynamicTool
  // needs. Interface so tests stub easily.
  type GenerationService interface {
      Generate(ctx context.Context, input Input) (Output, error)
  }
  ```

- **按功能组织包**：每个 `internal/<feature>` 自管 schema、repo、service 和 worker。`cmd/qiantie/main.go` 是唯一的依赖装配真源。
- **错误必须带上下文**：使用 `fmt.Errorf` 和 `%w` 包裹，不要裸 `return err`。

  ```go
  return nil, fmt.Errorf("load model: %w", err)
  return "", fmt.Errorf("%s: submit: %w", t.ToolName, err)
  ```

- **前置 guard 早返回**：让主逻辑保持浅层结构，不把正常流程深度嵌套在错误判断里。
- **仓储分层**：repo 用 interface 加 `MySQLXxxRepository` 实现；service 位于 handler 与 repo 之间。

## 3. 命名

命名要描述性强、少缩写、动词开头。长一点没关系，读名字就应知道它负责什么：

```text
ResumeOrphanedTasks
buildCanvasSummaryMessage
signContentBlockURLs
canonicalizeArtifacts
compileToSkill
useIsSuperAdmin
```

## 4. 务实，反过度设计

不要提前制造没有消费方的抽象。把当前简化方案和未来扩展点明确写下来，用注释说明“够用就好”的边界：

```text
Phase 0 executes tool calls serially; no fan-out concurrency yet.
Long-lived loop tied to context.Background — fine for now, can be wired
to a shutdown signal once we get one.
```

先卡清楚 MVP 范围。例如画布序列化器先只支持 `gen-image` / `gen-video`，不要一次性铺满所有未来类型。

## 5. 防御性但克制

只在真实发生过问题的边界加防御，例如 stream parser 异常、orphan `tool_result` 和签名 URL 过期。内部代码按约定工作，不要到处堆无意义的 nil 检查或 try-catch。

## 6. 前端 TypeScript / React

- 使用函数式组件和 hooks。
- 使用细粒度 Zustand selector，避免无关的重复渲染，例如 `useCanvasStore((s) => s.nodes)`；一次订阅一个字段。
- 使用 `dangerouslySetInnerHTML` 时，必须先配合 `DOMPurify.sanitize`，并限制标签和属性白名单。
- 类型要显式窄化，例如 `as string | undefined`，不要滥用 `any`。

## 7. 测试

- 测试名使用完整句子，格式为“主体_条件_预期”，例如 `TestRun_WithMentionImage_BuildsMultimodalUserMessage`。
- DB、LLM、上游接口等外部依赖必须通过 stub 接口隔离。
- 注释解释“这个测试在防什么回归”，不要复述断言本身。

## 8. 注释与文案语言

- 代码注释使用中文，技术名词保留英文，例如 `media://`、`canonicalize`、`LRU`、`tool_event`、类型名、函数名和字段名。
- UI 文案、错误提示、commit message 和设计文档使用中文，技术名词同样保留英文。
- 代码字符串字面量（日志、error message、struct tag、enum 值等）遵循原有约定；不要为了修改注释顺手改变它们。

## 9. Commit message

格式：

```text
<scope>: <动作>——<原因>
```

固定 scope 前缀按项目模块选择，例如 `agent`、`web`、`media`、`canvas`、`deploy`、`admin`、`api`。破折号 `——` 后必须说明 why，而不只是重复改了什么：

```text
agent: 撤掉 buildHistory 的 media:// 重写——它改了 LLM 输出行为
media: MediaSigner 加 20h LRU——所有签名 URL 在缓存期内稳定，浏览器 cache 不再每刷必失效
deploy: --binary-only 不再覆盖 static/ 占位——保留上次 build 的 SPA
```

中文为主，技术名词保留英文。
