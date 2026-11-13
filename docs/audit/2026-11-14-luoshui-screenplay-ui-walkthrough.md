# 洛水漫剧UI：实际操作证据与对齐规则

这份记录回答“有没有真正调用Craft Anything，以及为什么旧界面不能称为一比一”。目标为本机已运行的洛水v2.9.0，PID33528、窗口10501110、样例1212。没有在原版点击生成、删除或提交业务任务。

## 已实际调用的工具

使用Craft Anything `runtime-ui-read`与`runtime-ui-walkthrough-record`，并在带原件坐标映射的JS/CSS派生证据上做静态核对。不是只阅读技能说明。原件bundle `index-DszfAkKD.js` SHA256 `2e5c09685ff5bc10160480f963936b56bac38a57a9e9b6bd199492f9437df1e8`。

|实际操作|观察到的页面状态|原始记录及证据|
|---|---|---|
|读取当前原版窗口|158控件、1450×900窗口；两行页头、三步、分集正文和底部下一步|`original-ui-focused.json:1`；`ev_ec6701ac9d109b84bd924855aef5f809f3175cd65438e04ece311998a8a881bc`|
|按“下一步”，工具返回dispatched|进入资产；人物4、场景2、道具0、站位0、AI音色0；资产卡和工具条|`original-ui-next.json:1`；`ev_64984ea40759face4bd75a070bc7ac087a78e29fa29e1e4e0a23d797c296f483`|
|按“站位0”，返回dispatched|站位空态；新增站位、批量出图、开始分析、搜索、数量；不是普通人物改type|`original-ui-standing.json:1`；`ev_5e556b8e8144b13ec1bf8e7ee20f0bea717d68867abbe09a0420f91bba28d3c5`|
|再按“下一步”，返回dispatched|“部分资产尚未生成”清单；“返回生成”“忽略，下一步”|`original-ui-assets-next.json:1`；`ev_368d847b9befd42f07dba6735632011c82a1740471168c0756dc61d01f2dbd65`|
|按“忽略，下一步”，返回dispatched|分镜页按集展示，当前1集、0分镜；查看剧本、规划、导入等|`original-ui-storyboard.json:1`；`ev_4cff12433fa9a14da304c1f13bca3078965deb50e9f846ca1478e76321ba6cc8`|
|按Agent，再按收起面板，均dispatched|右侧Agent按需显隐；原版含对话/附件/Skill/子会话等入口|`original-ui-agent.json:1`与`original-ui-agent-close.json:1`；`ev_3eea431d2741a730361aa2dcd1b9967a732832471e62855c2fef00081068124e`、`ev_edca4e02937056a82e4fe61eebfafb667a3de1a9dcedf5fa27d75771b501a3e6`|

这些记录统一位于 `_local-archive-20261028/luoshui-screenplay-contract-20261114/`。工具的“dispatched”与随后`nodes_after`状态一起作为切页证据，不能单凭命令返回成功推定页面变化。

## 静态结构与实现规格

完整规格见案卷`ui-layout-contract.md`和`standing-contract.md`，每条含派生行号及原bundle行列/UTF16范围。

- 新组件qga/ScreenplayAIChat渲染：原bundle `56:2560150`、`[11936931,12295925)`；派生`ui-layout/layout-41.js`。
- 两行页头、配置胶囊左/三步居中：该派生文件3–92行，`layout-css.css:16–50`。
- 剧本max1196px居中、分集卡和内联编辑：`layout-41.js:563–848`；“共N集”确实为原版`text-4xl`，不是自行推断的小字号。
- 模型配置是1176px/max96vw模态，不是常驻右栏：`layout-41.js:4870–5055`。
- 底部确认胶囊235px、bottom16px：`layout-css.css:5–15`；分镜最后一阶段没有第四步。
- 资产五子类与工具/图片卡：`layout-41.js:2556–2995`；分镜按集制作区：`1947–2555`。
- 站位是独立qaa组件：原bundle `56:2066546`、`[11443327,11495894)`；`standing-component.js:1160`之后默认搜索/姓名开关/数量/图片卡，详细编辑和分析/批量是独立弹窗。

## 为什么旧版不能叫一比一

原V88采用七个侧导航加七个重复页签，并常驻右侧大设置表；站位实体和接口缺失。它们与上述原版结构存在明确差别。只增加选项或修改颜色不能解决这个问题。

本轮按三步重排，保留本项目品牌和既定玻璃主题；站位独立保存、名称绑定、真实参考图、分析、生图、上传和标注改图接入本项目服务。原版隐藏站位分析正文/算法未取得，因此Go分析规则明确为独立实现，不冒充逐字原版。

## 仍然不能做出的声明

原版截图捕获多次超时；AX矩形和CSS不能替代同尺寸图片逐像素对比。本轮未完成原版正常/取消的模型HTTP对照。原版Agent的全部Skill、多会话和子代理入口，也不能仅凭复用本项目助手就称为完全相同。不能据此宣布“全软件所有功能、所有状态已100%一比一”。

本地界面验收与当前构建以`docs/acceptance/2026-11-14-screenplay-creative-contract.md`后续更新为准。测试数据“窗外的故事”为手工录入的验收剧本，不是假装AI已生成的结果。
