# 小说创作手机实现计划

1. 写失败测试：独立novel入口、旧fetch任务恢复不受影响、六模式不同请求体、长短篇计数字段、续写章节前置、请求超时不可重复提交。
2. services/novelCreation.ts只负责真实接口；novelCreationInputs.ts定义可验证输入/状态/导入拆章，novelRecovery.ts按真实账号保存草稿与不可变请求。
3. 分包novel/index为作品/会话库；NovelStudio为原创讨论；NovelCreate为仿写/番外/导入；NovelProject为生产会话/运行/章节。互相以真实返回的project/session/run标识切换。
4. 共享小尺寸工作区样式，底部面板与键盘留白；不新建图片素材、不改变功能主页风格。注册分包与novel目录项，保留fetch旧路由和素材。
5. 验证重复点击、创建结果未知恢复、只读轮询、版本冲突和会话隔离；构建后处理WXSS，再跑全量产物测试；微信检查真实空状态和配置。
6. 将限制、截图、剩余入口核对写入功能清单，按明确路径提交本地v88，不上传或公开部署。
