# 分镜词选择器：41条目录与原版布局规则

2026-11-15静态恢复。本规格不代表原版窗口本轮实际操作、完整像素对照或付费生成验收。只改证据/规格，不改选择器业务代码。

## 原件、组件与入口

恢复前端`01-readable-frontend/pretty/index-DszfAkKD.js` SHA256 `8c5145ee0f26f594c35c6ca954e24d87dbbb44f5506029343f211a374992e492`，当前`ScreenplayAIChat` scope=`data-v-0f83330a`。配套原CSS为`00-original/dist/assets/index-BbbJnhtu.css`，SHA256 `e0e7e16ac35b7f99b354c0b2f89295e459b4fb36aa7957454f8ac689ede82211`。以上路径均相对于`_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/`。

证据`07-parity-execution/storyboard-picker-original/source-excerpts/`：四组带原行号的完整相关逻辑/DOM片段，19条CSS规则带原文件行与字符起止偏移，以及catalog-classification.json。派生片段不是完整原源码。另一通用prompt-picker（480825+）也有860宽，不能拿它的CSS替代当前sb-prompt-picker。

实际REA对派生完整函数检查`rea-analysis.json`：`ev_e449695b587e8935d1af90492e98377d6686cb60c146d7b9aff0b954d9e921dc`，1文件/4492字节/816AST、0解析失败；只覆盖派生函数，不证明运行/完整组件布局。独立Craft`reconstruction.cany`已记录finding/unknown；可导入`storyboard-picker-craft-bundle.json` SHA256 `b75fbd03b388e178ba6d0f479f9ed59e03e955d099b8cb741ba9c20d2a5583fa`；source inventory `ev_4c39bd52dd1c1a84b1670dece64f45b669830426fc36f8745875f52bcb59db6e`，derived inventory `ev_be79a8fbf4cb3de9d64c3d6a689c02ec0c057c0464f51c8d877ec566ffb5dec3`。REA原记录另外保留。

打开函数Ou442637–41清搜索pu，vu=true，调用Ru加载收藏/最近偏好；不重新请求全部prompt。弹窗DOM457469–576，mask-closable=true；遮罩和关闭按钮（457494–505）关闭vu。选择行Au442649–63更新du，异步保存最近/配置，**立即关闭弹窗**，没有额外确认按钮。

## 目录为何应显示原41条

448126原调用`get_all_prompts("剧本分镜",1,1000,"")`。448127–33将每行映射为label=name,value=id,description=description||"",createdAt=created_at||"",updatedAt=updated_at||created_at||""。原63条allowlist catalog中type=剧本分镜恰**41条**；不是只列3个固定引擎。没有选中项或当前选中ID已不在列表才默认cu[0]；这发生在排序派生前，不能假定默认一定是UI按日期最新的一项。

41个原ID：137,138,139,146,147,149,156,157,175,176,177,178,180,181,190,192,193,202,222,224,232,235,236,237,238,239,240,247,252,255,268,269,275,276,277,284,303,304,306,307,310。

CM当前`CreativePromptOptions`额外append两个旧兼容项（backend/internal/prompt/creative.go46），使storyboard总数43：

| CM旧ID | 身份 | 原41主目录归类 |
| --- | --- | --- |
| storyboard_xiaoyunque_v3 | 旧正文与原#268 strip后完全相同，SHA630a00ee06267fc723f69ca0ce526cc0f1b91211441fd022fec666b5564db7e8 | 属兼容引用，不能重复计第42条 |
| storyboard_xiaoluo_v7_15s | 旧manifest自报另一历史row225，SHA094f98b7c49e829d7cb3a9816c74c6434c41b6a1f5cc2e120c778383960f5b93；不等本轮原#224 | 非本轮原41，归compatibility，不能按名字强并正文 |

原41自身已经包含V7（222/224）、小云雀（202/255/268/269）。不能把额外2条当作原版缺失资产；也不能删旧ID导致历史作品引用失效。`catalog_group=original/compatibility`是CM明确分类适配，不是原数据库自带字段。原41中138/139在当前CM协议实现仍Unsupported，显示目录41与可运行39要区分，不能因不能解析而谎称原版只有39条，也不能让用户无提示执行unsupported。

## 布局尺寸

DOM457473–76：width=860px，max-width=94vw。内壳rounded-2xl/overflow-hidden/shadow-2xl，暗色#0c0d0f、亮色#fff。截图中总高620、y=114是本轮父任务提供的对照测量目标；**没有发现原CSS固定height620或top114**，不能把它们写成原版固定声明。

| 区域 | 原声明与行号 |
| --- | --- |
| header | JS437421 class flex items-center justify-between gap-3 px-5 pt-5 pb-3；全局utility左右1.25rem、上1.25rem、下.75rem、gap.75rem；控件实际高依Naive及根字体 |
| 标题与计数 | 457479“选择分镜词”+cu.length；字号text-sm、font-semibold，计数11px圆角badge；总数不随搜索减少 |
| 搜索 | 437421、457480–93：flex-1，max-width320px，小型clearable，placeholder搜索分镜词...，SearchOutline14 |
| 关闭 | 小型quaternary圆形button，CloseOutline16 |
| body | display:flex；min-height420px；max-height:min(62vh,560px)；顶部border1px；不是固定560px高度 |
| sidebar | width280px、flex-shrink0、column、gap16px；padding14px 14px 16px；overflow:auto、scrollbar-width:none；右border |
| main | flex1/min-width0；padding14px 16px 16px；overflow:auto；column gap10px；隐藏滚动条 |
| section/list | column gap8px；section-title flex center/gap6px、11px/weight600、letter-spacing.04em |
| row | width100%、display:flex、center、gap8px；border1px/radius12px；padding10px 10px 10px 12px；star26px高+上下20px+边2px→通常48px行高；不是另写height48px |
| row文本 | flex1/min-width0、center/gap8px/font13px；标题truncate单行；日期右侧见下 |
| star | 26×26px/radius8px/gray#9ca3af；hover背景#fbbf241f、色#f59e0b；is-on色#f59e0b；内部icon15px |
| recent | flex-wrap、gap6px；chip max-width100%/ellipsis/nowrap、border1px/pill、padding5px 10px、font12px |

body的视口表达式、字体、header控件和Naive弹窗定位共同影响总高及垂直位置。本轮未提供可独立核验的原浏览器computed-layout，截图620/114应在同视口视觉复核，不可凭静态CSS推导所有环境恒定等值。当前scope无专门移动断点，sidebar固定280可能在小视口挤压main，不能据max-width94vw宣称移动端完整适配。

## 列表、日期与选中态

cu/du/vu/pu/fu/mu/hu状态442544。搜索bu442549–54 trim+lower匹配label或description；favorites/recent/all全部过滤。标题总数cu.length；侧收藏计数wu.length、最近_u.length、主栏Su.length（457514/39/55），后三者随搜索变。

主栏“全部 · 按时间”（457549–55）：Su442572–75先按`Date.parse(String(updatedAt||createdAt||"").replace(" ","T"))||0`倒序，同时间按value的zh-CN numeric倒序。收藏按fu顺序映射现存记录，失效ID过滤；最近按mu顺序，排除收藏，过滤搜索，最多8条（442566–71）。

yu442555–61：updatedAt优先createdAt，trim；空值不显示日期；将第一个空格转T给new Date；可解析时按浏览器本地年月日时分显示**MM-DD HH:mm**；无效日期显示原串前16字符。日期span margin-left:auto/flex-shrink0/font11px/tabular-nums。DOM仅在yu非空时渲染（457564–67），不能用任意生成时间代替。

同组后端证据worker已只读恢复allowlist日期：`07-parity-execution/storyboard-catalog-metadata/source-dates.json`。原字段TEXT，形如`2026-11-10 15:08:11`，无时区；不得凭空补Z。示例#137=2026-03-04 18:21:53，#268=2026-11-09 11:39:38，#307=2026-11-10 15:08:11。此独立日期证据由protocol_fields工作项负责，不是本worker重新访问真实DB。

Cu442576–78：按Wu规范化ID比较du；选中亮色bg-indigo-50/border-indigo-200/text-indigo-700，暗/纯黑bg-indigo-500/15/border-indigo-400/30/text-indigo-300。未选中亮色bg-white/border-gray-100/text-gray-600，hover gray50/gray200；暗色slate800/70、border-slate700/80、text-gray300，hover slate800/slate600；纯黑light/border/text-gray300，hover muted/white15。行无单独disabled状态原证据；CM针对unsupported禁用属于明确协议约束。

## 收藏、最近和持久化边界

Ru442583–615并发共用gu加载：优先load_storyboard_prompt_picker_prefs→data.favoriteIds/recentIds，成功hu=true；失败或无专用接口fallback load_global_settings→storyboardPromptFavoriteIds/RecentIds。Mu442616–36仅hu=true时保存，优先save_storyboard_prompt_picker_prefs({favoriteIds,recentIds})，否则load+save_global_settings保留其他设置；异常仅日志，没有回滚/用户错误提示。

Nu442645–48需等待Iu加载成功才收藏：新收藏置顶，取消过滤同规范ID，保存。星标DOM有stop，点击不会选中/关窗（457528/70）。选择Au立即du=e并关窗，最近置顶去重最多8，再sd(true)存项目配置；异步保存失败仅日志，选择窗不等保存成功才关闭。刷新/重开偏好加载不是后台生成，也不会消耗模型。

左栏空收藏文案“点击右侧星标，收藏常用分镜词”；最近空文案“选择分镜词后会出现在这里”；主栏空“无匹配的分镜词”（457516–59）。没有原证据表明默认仅3条、搜索能改变目录总数或legacy项属于本轮41。
