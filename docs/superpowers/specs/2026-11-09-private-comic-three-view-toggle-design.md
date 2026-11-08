# 私人定制人物三视图开关设计

## 目标

让私人定制漫画的人物资产编辑框与批量工厂保持相同能力：用户可以在 AI 生成人物资产图前选择是否使用三视图。

## 现状与根因

- 批量工厂维护 `assetImageThreeView` 状态，在人物资产编辑框中显示“人物三视图”开关，并把选择作为 `useThreeView` 传给后端。
- 私人定制虽然复用了 `PrivateStudioAssetEditor` 外壳，但没有接入开关。
- 私人定制的 `runPrivateComicAssetImages` 当前把 `useThreeView` 写死为 `true`，因此人物资产总是按三视图请求，用户无法选择。

## 方案选择

采用最小且与批量工厂一致的方案：

- 私人定制增加独立的 `assetImageThreeView` 页面状态，默认 `true`。
- 只有当前编辑资产为 `character` 时显示“人物三视图”开关。
- 点击单个人物资产的“AI生成”时，把当前开关值传给父页面，再写入现有 `generateBookAssetImages` 请求的 `useThreeView`。
- 场景和道具不显示开关；它们生成时传入 `false`，避免人物专用模板误用。
- “生成缺失资产图”继续保持人物默认三视图；这是批量操作，没有单个资产编辑框可供选择。

不采用的方案：

- 不把开关放进统一外壳。统一外壳只负责布局，不应该混入不同业务的生成参数。
- 不新增后端接口或数据库字段。现有接口已经支持 `useThreeView`，开关只控制本次生成，不需要永久存储。

## 数据链路

1. `PrivateComicWorkspace` 渲染人物三视图开关并维护布尔值。
2. 单项生成调用 `onGenerateAssetImage(asset, useThreeView)`。
3. `ScriptPage` 的回调把值传给 `runPrivateComicAssetImages(assetId, useThreeView)`。
4. `runPrivateComicAssetImages` 请求 `generateBookAssetImages` 时写入 `useThreeView: Boolean(useThreeView)`。
5. 后端沿用现有逻辑：人物按开关选择三视图模板，场景和道具不套用人物模板。

## 测试与验收

- 源码测试先证明私人定制缺少开关和参数传递，再补实现。
- 测试确认开关只在 `character` 出现，默认开启，生成回调携带布尔值。
- 测试确认 `ScriptPage` 不再写死 `useThreeView: true`。
- 前端相关测试和构建通过。
- 重建本地嵌入式 Go 二进制并重启 4001。
- 在内置浏览器打开人物资产，看到默认开启的“人物三视图”；场景、道具资产不显示。
- 不更新线上。
