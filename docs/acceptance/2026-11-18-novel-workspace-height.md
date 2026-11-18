# 小说创作页底部空白修复

代码提交297b3225，文件frontend/src/user/pages/NovelCreationPage.css:4。原规则100dvh减105px仍在扣除已经取消的顶部区域，720px窗口实际页面仅615px。现改为跟随父容器height:100%，min-height:0，矮窗口也不再强撑520px。

前端生产构建、Go嵌入构建通过。更新前33项生产状态读取均空闲，小说派发无ready/running/claimed；仅停止已经核对路径的本地4001旧进程，启动明确候选包。未调用模型、未推送、未发布公网。共享脏工作区其他改动保留。

实际浏览器：默认1280x720页面底部720、输入区底部719，原105px空白归零；1280x540与1440x900实测底部空白均0，输入区均在可见范围。已重置临时窗口尺寸。记录在_local-archive-20261028/novel-height-20261118/height-checks.json，真实截图4001-height-fixed.jpg，运行身份build-info.json，构建frontend-build.log。
