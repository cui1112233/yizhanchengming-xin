# 批量工厂漫画参考素材胶囊设计

## 目标

把漫画提示词中的参考图从纯文字标签改成可识别、可预览的参考素材胶囊，外观参考用户提供的素材输入框截图。

## 展示

- 标题显示“参考素材”和当前数量。
- 素材只占一排，超出宽度后横向滚动。
- 每项显示 40×40 缩略图、素材名称和来源。
- 点击缩略图使用 Ant Design 图片预览放大。
- 图片加载失败时保留素材项，显示占位状态与名称。
- 普通界面不显示内部 URL；开发者仍可通过提示查看来源信息。

## 来源和删除规则

- 画风库、场景图库、角色资产等系统自动带入的参考图：标记来源，只能查看，不能在提示词弹窗中删除；应回到对应配置或资产处取消。
- 用户手动上传的当前页参考图：标记为“手动上传”，显示删除按钮；删除前确认，删除后保存当前页参考图列表。
- 删除一张手动图不影响系统自动参考图，也不影响其他漫画页。

## 数据契约

后端在漫画提示词预览结果中同时保留 `referenceImageUrls`，并新增结构化 `referenceImages`：

```json
{
  "id": "batch-comic-reference-scene-urban",
  "name": "都市",
  "url": "/api/local-executor/v1/batch-comic-reference-scenes/batch-comic-reference-scene-urban",
  "previewUrl": "/api/presets/comic-reference-scene/batch-comic-reference-scene-urban/preview",
  "source": "scene-library",
  "removable": false
}
```

手动上传图的 `source` 为 `manual-upload`，`removable` 为 `true`。`url` 是发送给图片模型的地址，`previewUrl` 是登录用户可读取的缩略图地址。前端按 `removable` 决定是否显示删除按钮，不自行猜测权限。

## 验收

- “都市”等内置图显示真实缩略图和中文名称，点击能放大。
- 多张参考图横向滚动，不撑高弹窗。
- 系统图没有删除按钮；手动图可单张删除。
- 界面不直接展示 URL，图片失败时仍看得到素材名称。
