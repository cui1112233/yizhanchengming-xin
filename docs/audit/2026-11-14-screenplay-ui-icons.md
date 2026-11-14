# 漫剧UI和图标：原件核对与独立实现

本轮范围是洛水新三阶段工作区、模型配置与站位弹窗；保持一战晟铭品牌和既定玻璃主题。用户明确批准UI、图标、尺寸、间距、卡片、弹窗及交互状态对照。

## 原件与证据

原件 `F:/软件/洛水2026/洛水2026/dist/assets/index-DszfAkKD.js`，SHA256 `2e5c09685ff5bc10160480f963936b56bac38a57a9e9b6bd199492f9437df1e8`。Craft源导入证据 `ev_006921f3a8626ef3f0e1324799488343d85a9168886e1105e05fc62cd3411b8b`，解析产物位于案卷 `_local-archive-20261028/luoshui-screenplay-contract-20261114/icon-ui-20261114/`。

该目录icons.json有66个完整SVG树，usages.json有347处带条件的使用位置；provider-icons.json保留22个品牌分支。每项有原文件行列/字符偏移、size、class、外层disabled/loading和显隐条件。validation.json未发现未解析节点。347处不代表同时可见，包含原模板旧条件分支。

## 已确认的视觉规则

|位置|原图标与尺寸|原件定位|
|---|---|---|
|返回|ArrowBackOutline 18|layout-41.js:7–13，usage有精确原偏移|
|模型配置|SettingsOutline13、ChevronDownOutline12|layout-41.js:40–53|
|画风|ColorPaletteOutline13、ChevronDownOutline12|layout-41.js:54–64|
|Agent|纯文字，无Bot|layout-41.js:64–74|
|剧本排序/编辑|ReorderThreeOutline16、CreateOutline15|layout-41.js:780–844|
|资产五类|People/Location/Cube/PeopleCircle/MusicalNotesOutline16|layout-41.js:2630–2633与layout-01.js|
|站位hover动作|Create/Image或Refresh/CloudUpload/TrashOutline12|原UTF16起点11477176、11477380、11477570、11478288|
|底部mark|SparklesOutline16；右箭头为文字→|layout-41.js:3922–3931|

普通SVG保留原512viewBox、fill和32/48描边；不能全套替换成24viewBox/strokeWidth2的近义图。模型入口按原模型名映射品牌SVG，未知模型用原灰底白头像，渐变id须按实例隔离。未显式size且位于原small NButton icon-slot时继承18px（原主题证据29:334562、29:343362、29:591657），不能一律14px。

React实现只渲染白名单静态SVG树，不执行发布包JavaScript；图标为装饰，按钮保留清晰名称、键盘焦点和原有真实handler。功能未知不添空入口。

## 有意保留的差异与未完成范围

原容器≤1520隐藏Agent文字会形成空胶囊；为符合新手可用的要求，本项目保留纯文字Agent，不加Bot。品牌和玻璃主题按用户既定选择保留。

分镜原版“查看规划”“资产补充”涉及独立已存规划、逐集推理并确认写入；当前链路尚不完整，不用跳转资产页冒充。前缀、规划密度等无对应持久链路的控件不添假开关，记录为缺口。

原版窗口截图本轮仍捕获超时。源SVG/尺寸/条件可核对，不能因此宣布逐像素或全部Agent功能已一致；本轮不使用任何新增模型额度。

## 本轮实现和实际页面验收

功能提交74164055，短窗口修复6d4de5fe；最终本地4004 PID55608，candidate-132321，embed入口/assets/user-BGoeDwqT.js，runtime-build-info读回6d4de5fe。保留4001其他工作，未推送和发布公网。

实际接入60处字面+4处动态调用、35种可解析图标（含2个inline），不是66种全部使用；模型品牌有页头18px/菜单24px两处渲染，22分支用例通过。implementation-manifest.json逐项映射原使用位置。

页面66/66回归、图标5/5通过，前端构建成功。独立复核无新P1/P2。最新小视口回归在1280×720检查四弹窗footer可见、正文滚动、关闭不提交。

已登录本地页面实际验证图标DOM：返回18、齿轮13、调色盘13、三横线16、方框笔15、五资产tab16；模型列表真实Gemini图标24，未选择模型或执行生成。键盘切换三阶段、站位开关和取消正常；分镜单双列、资产紧凑/标准切换有效，缺模型/空结果仍明确禁用。测试资产和站位是手工验收数据，不是生成结果。

原先720高窗口站位footer出屏，现适配为top24、正文内部滚动。实际页面测得footerBottom661≤720、body540/scroll631；实图standing-final.png证实保存/取消可见。此为本项目React容器适配，不冒充原版已观察像素。

证据目录icon-ui-20261114/live-acceptance.json及before/after DOM记录。实图assets-final.png、standing-final.png、story-live.png、model-live.png、storyboard-live.png。预览留在IAB36资产页，旧的代理预览35已关闭，无尺寸覆盖。本轮新增文字/图片/视频模型调用均0。
