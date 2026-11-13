# 玻璃主题叠层验收

## 后续视觉修订：隐约透光

用户指出上一版过于实体，要求磨砂后仍隐约透出背景。已按此修订，代码8173d6a6：深色卡片/浮层/弹窗底色不透明度改为34/60/70%，浅色50/72/80%，背景模糊分别18/28/36px；深色遮罩42%降至24%，浅色24%降至12%。以下原始验收数值保留为上一版记录，以本段为现行参数。播放器与时间线的稳定表面不变。

修订前端构建通过（1m18s），Go内嵌候选构建通过。本地4001运行qiantie-soft-frosted-local.exe，PID51036，health通过，运行清单8173d6a6，前端入口/assets/user-Ctq2nHfH.js。实际批量工厂统一配置弹窗读取rgba(29,38,52,.7)、blur(36px)，遮罩rgba(6,10,18,.24)，正文filter:none；稳定截图显示背景光影透出，未保存业务配置。此轮未再次切换浅色。

证据：`_local-archive-20261028/frosted-soft-build.txt`、frosted-soft-embed.txt、frosted-soft-start.txt、frosted-soft-dark.jpg。

2026-11-13，代码提交 fdc7e5bd。仅本地修改、构建与运行，没有推送或部署。

## 改动

新增统一样式 frosted-surfaces.css，最后加载。普通工作卡片、浮层、弹窗分别使用16/24/30px背景模糊，越靠上的面板底色越扎实。深色底色不透明度56/88/94%，浅色70/92/96%。没有使用元素 filter 模糊文字或媒体；弹窗内卡片避免重复叠加背景模糊。不支持背景模糊时，默认使用实色底（代码检查，未在不支持的浏览器上实测）。

覆盖 Antd 卡片、下拉菜单、弹窗、抽屉、日期菜单、提示、通知，以及 CM 卡片/编辑器/菜单/项目栏、宠物对话与 AI-CUT 稳定工作表面。其他配色主题保持原样。现有未纳入本提交的 cinematic-glass.css/global.css 改动保留，没有接管。

## 验证

- 现有主题相关检查20项：19通过，1项既有失败。theme-css-contract.test.js 的“仅允许 light 声明”断言与 HEAD global.css 中已有 dark 声明冲突；用 git show HEAD 独立确认原代码同样包含 dark/light，本任务未改 global.css 或放宽测试。
- 前端构建通过（1m46s，保留原有大文件提醒）；Go 内嵌候选构建通过。
- 本地4001运行 backend/.release/qiantie-frosted-local.exe，PID29340；health通过，runtime_manifest_sha=fdc7e5bd，frontend_entry=/assets/user-CuyTTiTL.js。
- 已登录实际页面：批量工厂“统一配置”弹窗中打开“图片生成通道”下拉，读到rgba(34,44,59,.88)、blur(24px)；弹窗rgba(29,38,52,.94)、blur(30px)，遮罩blur(4px)。未保存任何生产配置。
- 浅色弹窗rgba(250,252,255,.96)、blur(30px)；说明浮层rgba(247,250,255,.92)、blur(24px)，文字rgb(25,43,66)。实际截图检查内容可读。
- CM工作卡片rgba(30,39,52,.56)、blur(16px)、filter:none；宠物对话rgba(34,44,59,.88)、blur(24px)。没有调用生成服务。
- AI-CUT播放器与时间线底色rgba(16,23,34,.9)；视频仍为blur(0px)，正文容器filter:none。
- 验收后恢复深色。用户当时已有壁纸设置为模糊19px、亮度52%，整个检查保持这些值，未重置用户偏好；默认42px/50%的代码仍由现有背景测试验证。

## 证据

`_local-archive-20261028/frosted-surfaces-build.txt`、frosted-embed.txt、frosted-runtime-start.txt。

截图：frosted-before.jpg、frosted-dark.jpg（下拉叠层）、frosted-light.jpg（浅色说明浮层）、frosted-dark-final.jpg，均在同一证据目录。独立验收标签32保留；用户原标签31未重载。

没有逐一打开所有业务页面；通过统一控件规则覆盖共享表面，并在批量工厂、CM、AI-CUT三类实际工作区检查。
