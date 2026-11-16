# ZIP画风、模型与三视图传递

## 画风选择界面绑定

原AIChat O:118144的Nn复制所选画风对象；Ln取当前选择或作品selectedStyle；Un按desc、description、prompt顺序取画风文字。O:118162附近从load_assets("styles")载入候选。图片预览优先images[0].url，其次本地path。不能把名称当作完整画风信息。

qn(O:118178–118195)先更新Wn并关闭选择窗口，再调用save_screenplay_ai_config({path,style:完整对象})。真实原生保存和新进程重开确认id/name/desc/images保留。Kn(O:118196–118207)把选中model保存到同一作品配置。模型配置中的characterThreeView另由模型配置保存；不要自行新增一套互不相通的选择状态。

主工作流原请求O:119002附近传art_style、style_desc、selected_style、image_gen_model、image_edit_model、character_three_view、aspect_ratio和pipeline。分镜生成对style文字的替换已在11验证。选择画风对象、分镜文字、资产图片参考是不同层次，不能只改预览图或按钮颜色。

## 三视图实际出图入口

执行ZIP ScreenplayAgentScheduler._run_character_agent/_run_scene_agent/_run_prop_agent，保留原prompt组装，在最终_generate_image_by_model处替换为不生成图片的失败记录器。9组有效记录在native-api-isolated/asset-image-prompt-v2/v3.json：

- 角色单人图：要求纯白背景、全身正面，使用作品比例，本组9:16。
- 角色三视图：切换为正面/侧面/背面、面部特写、配色板与细节设定参考图提示词，实际参数强制16:9。本组原作品比例仍为9:16，故明确不是仅改文案。
- 场景和道具：不受角色三视图开关影响，仍传作品比例9:16及各自资产提示词。
- 三类均把OWN_IMAGE_MODEL传给出图入口，模型来源是image_gen_model；这是边界参数验证，没有验证实际供应商支持所有模型/比例。

## 参考图与文字

所选画风只有name/desc但images=[]时，本组最终出图prompt未包含OWN_STYLE_DESC，style_images/style_image_paths为空。提供隔离目录中的测试参考图后，三类出图入口均收到1份style_images和1个style_image_paths。说明该路径确实通过参考图传画风，不能因文字未出现就断言画风完全没传，也不能仅有描述就宣称图像参考已经生效。

试验使用自造1像素PNG，仅用于检查参考参数，没有生成用户图片。未调用真实图片接口，也未验证供应商收到参考图后的效果。CM需按自己模型能力决定参考图参数；不支持的渠道应明确提示，不能静默丢弃画风。

## 失败和例外

原qn在保存返回success=false时仍提示“已切换风格”；缺API也会提示成功。保存抛错时提示失败，但界面内存选择已经改变。CM应明确区分临时选择与已保存选择，失败允许重试并避免重开后悄悄回退。

正确的异步图片失败记录器返回success=false后，原三个生成Agent仍返回success=true、data=[]、“生成了0个…图片”。必须按实际有效图片数量和失败项判断结果，不得把这种布尔值当作全部生成成功。首次v1误用同步记录器引发await错误，保留为测试工具错误；只有修正后的v2/v3用于此项结论。

## 站位和其他多视图

这份ZIP的本AIChat入口已确认角色三视图；此前截图的独立“站位”页和场景四视图不能从该开关推导。本批只检索到分镜衔接文字中的站位描述，尚未建立独立站位页/按钮/接口的ZIP调用证据。保留为未核对项，不把我们已有positioning.go实现反过来当原版证据。

真实原版窗口、图标尺寸、弹窗状态及像素对照仍待完成，本文件记录绑定和实际参数，不称UI一比一验收。
