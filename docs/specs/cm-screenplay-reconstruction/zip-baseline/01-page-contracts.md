# ZIP页面与控件地图：首轮静态基线

唯一来源为本目录00指定ZIP。以下O指`readable-supplied-frontend/pretty/index-CnO3GE3B.js`，SHA256 `fc78d66444269ecb4af9b6114c649e735c11c7770dbfa0cc1d262fd2f3168335`。行号属于恢复文件，不是原始压缩JS位置。

## 已定位的组件范围

| 组件 | O行范围 | 首轮调查范围 |
|---|---|---|
| ScreenplayCreation | 124517–126987 | 作品库、新建、搜索排序、宫格/列表、打开/复制/删除、创建后分章 |
| ScreenplayAIChat | 116208–124117 | AI推理分析、剧本/资产/分镜、模型配置、章节与Agent |
| ScreenplayDetail | 105537–115422 | 作品详情及制作相关面板，具体进入条件仍需逐项复核 |
| ScriptEditor | 93404–98694 | 编辑器及工作区控件，原生打开窗口关联待复核 |
| ScriptShotRow | 92905–93241 | 编辑器分镜行控件 |
| ScreenplayEngineModal | 98697–100721 | 引擎相关弹窗，不能凭名字当作AIChat的模型配置 |
| ScreenplayAssetManagerV2 | 100757–101642 | 资产管理控件 |
| SimpleStoryboardMode | 102180–104823 | 简化分镜模式 |
| SimpleStoryboardRow | 101661–102129 | 简化分镜行控件 |
| ShotMediaWorkspace | 104830–105287 | 媒体工作区 |
| AgentWorkflowModal | 115441–116014 | Agent流程弹窗 |
| ScreenplayScheduledTaskPanel | 124123–124433 | 定时任务面板 |

这是12个已定位的命名组件，不是12个已验收页面。匿名弹窗、组件内部条件面板及原生另开窗口仍须展开。

## 可回查清单

原件专属目录的`analysis/`下已有`pages.csv`、`controls.csv`、`native-calls.csv`、`navigation-edges.csv`、`control-call-candidates.csv`及同名JSON。控件与调用有稳定ID、原件摘要、行号和UTF16字符位置。

本轮语法扫描得到1034处事件绑定、242处直接pywebview调用、127个不同方法名、521条label/value候选、140处title属性候选。它们包括重复、隐藏状态和非按钮属性，不能解释成1034个独立按钮或140个已确认弹窗，更不能用作完成率。173条控件到函数/接口的名称候选尚需核对作用域。

11条组件引用边来自编译后的组件包装别名分析；只是组件包含候选，不等于点击导航已实测。动态API别名、动态属性、共享图标组件及外部窗口仍可能不在本次清单内。

## 已确认的界面版本差异

O:122721–122802的AIChat“模型配置”是宽560px弹窗，内容为生图模型、改图模型、音色设计模型、角色三视图开关、显示原文开关以及取消/保存。不能把另一安装版本的三栏配置截图说成此ZIP页面。

当前CM `frontend/src/user/pages/ComicScreenplayModelConfig.jsx`有多栏配置与额外模板、指令等内容。属于待重新核对的来源差异，不能在未检查用户授权扩展与其他入口前直接删除。用户要求的自己的模型、CM/小Y名称及统一主题仍保留。

## 当前未通过的门槛

原版实际窗口每种状态的同尺寸截图、图标轮廓/资源引用、所有按钮的真实跳转、CM4001同项截图和交互，均未在本轮完成。上述静态清单不是UI完整对照通过。

## 图标原件提取

新增`readable-supplied-icons/manifest.json`及130份SVG：从本ZIP可读JS的命名SVG组件和原始path/static markup提取，不重新画图标。共定位132个命名SVG组件，其中130个静态图标可直接导出并通过XML解析；BaseLoading和QrCode含动态参数，保留未解析，不能用自画图标补齐。

已登记524处所选剧本组件内的图标引用候选，记录组件、行号、原图标定义行号和SVG文件。比如SettingsOutline原定义O:56638，AIChat引用O:120183。`analysis/icon-usage.json`可查每处来源。图标文件可读不等于尺寸、颜色、间距和状态像素已经验收。
