# 水货生产素材手机操作

在已有素材页增加独立面板，不改主页。与当前Go路由对应：项目media导入；assets/image与assets/{id}/image导入参考图；assets/generate生成；images列表/primary选主图；segments/{id}/assets绑定。图片与视频来自相册/拍摄，音频来自微信聊天；JSON接口单次8MiB手机限制明确展示。付费操作显式确认，未知结果用现有会话锁。

资产详情合并project.books[0].assetRecords，保留facts/appearance/relationships等真实字段与revision；不要把description当成后端支持字段。只选本项目资产和分镜，生成使用唯一的真实模型编号。读取/查看不收费、不创建测试数据。

测试接口参数、完整字段保留、绑定范围、未知结果保护与面板交互，微信构建后原生只读验收。
