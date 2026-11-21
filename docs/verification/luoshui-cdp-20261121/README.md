# 洛水原版 CDP 接入

2026-11-21（上海时间）。用户明确选择给洛水桌面程序接入 CDP。

- 原版目录：`F:\软件\洛水2026\洛水2026`，窗口显示 v2.9.4。
- 调试入口：`http://127.0.0.1:9226`；仅回环监听。应用页面：`http://127.0.0.1:21331/index.html`（重启可能更换页面端口和目标 ID）。
- 使用原版 Python312 执行未修改的 run.py，启动前设置已安装 pywebview 的 `settings['REMOTE_DEBUGGING_PORT']=9226`。
- 证据：edgechromium.py:81–84 将该设置追加到 WebView2 CreationProperties。run.py:40 会覆盖环境调试参数，因此没有使用只设置环境变量的无效方案。
- 原版 run.py SHA256 在启动前后相同，见 connection.json；没有修改原版安装文件或认证数据，没有注入页面脚本。
- REA `list_browser_targets` 和 `inspect_web_page` 已成功读取原程序真实页面、DOM 与可访问结构。运行时资源包含 `assets/index-DgTjggsH.js`，与已恢复版本的文件名一致（仅文件名不能代替摘要匹配）。
- 本轮没有点击生成、删除或作品提交，也没有记录凭据。捕获只覆盖附着后的短窗口，不能据此声称所有网络调用或功能已验收。

再次启动（先正常关闭旧洛水实例，确认9226无其他程序占用）：

```powershell
& 'F:\软件\洛水2026\洛水2026\Python312\python.exe' 'F:\脚本测试\一战晟铭-v88\scripts\launch-luoshui-cdp.py' --app-dir 'F:\软件\洛水2026\洛水2026' --port 9226
```

连接会随本次洛水进程退出而关闭；正常启动器启动不会自动开启此端口。
