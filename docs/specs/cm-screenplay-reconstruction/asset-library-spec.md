# G12 普通跨项目资产库：原版规则与恢复边界

本文件只记录原版恢复证据，不代表 CM 已实现、截图验收或付费模型验收。2026-11-15，工作基线 `810ec38af086567d252b3b4c8d6b0044698d56e1`。本轮没有启动原软件或操作其窗口，没有打开真实账户数据库，没有模型调用。

## 证据来源和覆盖

原件：`_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/00-original`；恢复前端：同归档 `01-readable-frontend/pretty/index-DszfAkKD.js`，SHA256 `8c5145ee0f26f594c35c6ca954e24d87dbbb44f5506029343f211a374992e492`。下文行号全部指向该恢复前端，不能当成原生 Python 源行号。

本轮证据：`07-parity-execution/asset-library-original/source-excerpts/manifest.json` 保存 22 组原行范围、原件 SHA、各片段 SHA；`extract_evidence.py` 可重建。片段不是完整原源码。`rea-focused/` 是完整相关函数构成的派生静态分析输入，外部闭包与组件初始化不在其覆盖范围内。

`call-chain.md` 保存带原行号的事件/确认链图。实际 REA 静态检查 `rea-analysis.json`：evidence `ev_85fabe310d314adcc434db98b260a1fa506a9a49a2338b310081a9dfdd5cd64d`，1文件/20380字节/4185 AST节点、0解析失败。独立 Craft 工程 `asset-library-original/reconstruction.cany` 保存 finding/unknown；可导入 `asset-library-craft-bundle.json`（SHA256 `ef693c139eda74b3d8f2b2e62f75952cf5f79b05ae73090ed24f3a8da6b87a80`），其中 source inventory `ev_a428b918daeb3b07368174eb598e6e08e9bf64105a37ea2c5f0d193ab82e96c4`、派生函数 inventory `ev_9330e00a575055467e8e9988f01e6e52e133832d6ab1b7a3a1a43a70234cf1c3`。REA 原记录文件另存，不冒充 Craft bundle 内自动分析结果。

这里的“观察到”均指前端字节/调用链静态恢复；未提供本轮原版 UI 截图。组件和按钮定位可以复核源码，但不声称当前可见窗口已到达该状态。

## 库的范围不能混用

普通全局库采用 `load_characters()`、`save_character(object)` 与 `load_assets("scenes"|"props")`、`save_asset(category,object)`，跨项目调用没有项目路径参数。角色页 65423+、通用资产页 266800+。它们不同于 518223+ 的 `CanvasAssetLibraryPanel`、头像参考资产库、JM 主体库、音色库。后几种不属于 G12 普通库。

当前作品另有两套同时存在的界面：AI 聊天当前资产 448k–459k，以及 `ScreenplayAssetManagerV2`（399385+）。两套规则不同，不可合并为“导入完整资产快照”。

## A. AI 聊天当前资产：新增时从库选择

打开新增按钮（454915+）按当前类型设置 `EW`，默认 `VW="ai"`，清空表单、选中项、搜索、分类；关闭/取消（457861、457869、458002）仅关 `FW`。源选项为 AI 生成、库、本地上传（448899–915）；切换库且缓存 `ob` 为空才 `fb()`（457886–897）。`fb` 448920–935 **不论当前类型均调用 `load_characters()`**，有 `lb` 重入保护；success 时替换数组，失败只日志，finally 解除忙状态。不可擅自把这个旧入口解释为按三类加载。

筛选（448874–895）：分类是所有 truthy `category` 去重排序，加“全部”；精确匹配分类，再用搜索 trim+lower 匹配 name 或 description，保持返回顺序。列表只展示 `cb.slice(0,60)`，无此处翻页（457933–970）；预览 `thumbnail||image`；点击只选一个并执行 `UW.name=UW.name||item.name||""`，不覆盖已输入名称。

确认 `xb` 449036–82，library 且有选中项时调用 `wb`：

| 当前字段 | 取值 |
| --- | --- |
| name | 手填 trim 名称，否则 library.name，否则“未命名” |
| identity / 场景道具 description | 手填 identity.trim，否则 library.description，否则空 |
| prompt | **仅手填** UW.prompt.trim，不复制 library.prompt |
| voice | 角色仅手填 UW.voice.trim；其他类型为空 |
| image_path | library.imagePath 或空；**没有文件复制调用** |
| characterDesignSelection | 角色表单值；不是库整对象 |

`wb` 449083–117 直接 push 到当前 `Se.characters/scenes/props`；无重名检查，无保留库 ID，无继承声音 ID/主体 ID。仅图片路径非空时才向 `X` 生成图片记录，角色附带 identity/prompt/voice；场景道具只 name/image_path/type。成功分支后 `FW=false; await Fb()`。**library 未选中也会关窗、刷新，但没有新增或明确提示。**

上传分支不同：`gb` 448958–972 用 `copy_file_to_project({project_path,source_path,sub_dir:"ai_generated_<type>s"})`，失败/异常可回退原路径。不要把它套到库选择。

`Fb` 449380–385 汇总当前持久化；`jb` 449352–361 调 `screenplay_save_agent_assets(projectPath,deepClone(X))`；`Dm` 445779–793 调 `screenplay_save_asset_parse_results(projectPath,{characters,scenes,props})`。这不是写全局普通库。

独立的批量手工新增 `kb` 449001–35 会展开别名（448982–84 用换行/中英文逗号、顿号、分号切分），候选内按 trim 后名称精确去重，再通过 `Tm(...,"append")`/`qm` 445722–77 对当前同类 trim 名称精确去重：大小写敏感，旧条目保留，不更新，第一个新候选胜出。**这个规则不能倒推覆盖单个库新增 `wb`。**

## B. AI 聊天当前资产：只更换图片

角色 `Cy` 449923+ 打开库更换，列表同上。确认 `Wy` 449683–711 只用选中 `imagePath`；场景 `ww/_w/Cw` 451523–587，库 `load_assets("scenes")`；道具 `Jw/Zw/Xw` 451739–803，库 `load_assets("props")`。后两者按 name/description trim+lower 搜索，无分类，最多前 60 张，图源 `thumbnail_url||file_path`（场景 UI 458939–459010）。只在列表为空时加载，缓存不保证每次打开刷新。

场景/道具确认只取 `file_path`，角色只取 `imagePath`；没有复制库文件，没有导入名称描述提示词或源 ID。`ag` 447706–725 按当前类型+当前名称寻找当前对象，`tg` 447700–701 更新 `image_path/bound_image_path`，已有 imagePath 与嵌套 appearance/scene_ref/prop_ref 路径同步；同名只更新 `.find` 的第一项。`X` 不存在同名项时才插入最小图片记录。然后关窗、成功提示、`Fb()`。没选图/路径为空给 warning，不关闭。取消仅关窗，不应用选择。上传则复制到当前项目；返回 success=false 不等于异常回退，不能保证每种错误都复制成功。

## C. ScreenplayAssetManagerV2：编辑表单中绑定普通库

这是真正另一套完整图绑定流程，入口按钮在编辑窗“绑定资产”（403032–46），点击 `Na("library")`，按当前 c.value 显示角色库/场景库/道具库。相邻 `jmLibrary`、`aiAsset`、voice/aiVoice 是其他源，不可算普通库。

`mt` 400373–398 从当前对象复制编辑草稿 `pt`，包括 name、description、image_prompt、voice_description、bound_image_path、bound_character_id/bound_scene_id/bound_prop_id、sora2_info 与角色声音/服务引用字段。`Na` 401789–927 清搜索、开 picker，每次重新加载当前类：characters→load_characters()；scenes/props→load_assets(category)。success=false 变空数组；catch 仅日志；无 api 的提前 return 会使 busy 保持 true，是原前端缺陷，不应当成可靠成功。

搜索 `qa` 401966–974：trim+lower 匹配 name/description；无分类、无截断、无此处排序。卡片 403198–228 渲染全部 qa，key=id||name，click/Enter/Space 均 `ja(item)`；预览 `thumbnail||image||thumbnail_url||file_path`，图片加载失败隐藏图片。取消：遮罩/空白/右上关闭 403165–188 均只 `ka=false`，未选择则不改草稿。

选择 `ja` 401984–402000 **立即修改编辑草稿，不立即保存当前对象**：

1. 图片路径 `file_path||imagePath||thumbnail||image||image_data||""`。若以 data:image 开头、api 存在、当前编辑项存在，调用 `save_screenplay_item_image(projectPath,singularType,current.id,data,"bound_<library.id>.png")`，success 后使用返回 path；失败返回保留 data 值，异常此处无 catch。
2. 写 pt.bound_image_path。按当前类写 pt.bound_character_id / bound_scene_id / bound_prop_id = item.generatedId||item.id；普通文件路径直接引用，未复制。
3. source 是 library 或 jmLibrary 时写 pt.sora2_info = item.sora2Info||item.sora2_info||""。**不复制 name/description/image_prompt/voice。**
4. picker 关闭并提示“已绑定”；可以重新开 picker 选择，最后一次选择覆盖草稿引用。没有多选、同名自动合并或依库名称创建新对象。

编辑“取消” `ht` 400399–403 清 ct/dt/预览，丢弃未保存表单绑定；若选择 data 图时曾保存图片，取消不见删除文件清理。编辑“保存” `gt` 400404–426：组成基本 name/description/image_prompt + 当前类型引用，调 update_screenplay_character/scene/prop(projectPath,current.id,fields)，Object.assign 当前对象，再 en、refresh、关编辑窗。这里 **没有检查 update 返回 success**，只有抛出异常才进入 error。因此 CM 不能把原版 toast 当存储成功证据。

`en` 400914–950 调 `screenplay_upsert_agent_asset(projectPath,singularType,payload,oldNameIfRenamed)`，基本字符串 trim，角色携带声音与主体/库引用，场景道具带对应 bound_*_id。同步失败只 warning；它是项目内 agent 资产同步，不是全局库更新。清除绑定 `Ba("image")` 402001–3 清图片+三类引用ID+sora2_info，不删除库对象。

## D. 保存当前预设到普通全局库

已确认普通调用链存在于预设/AI资产组件，而非上述 AI 聊天图片右键保存：按钮 178925–943 → wo 177446–65 → `save_ai_asset_to_global_library(asset, singularType)`。

按钮只在预设缓存 thumb 存在时显示；点击重新取当前类型对应 appearance/scene_ref/prop_ref.thumbnail，缺少则 warning“该预设没有图片，无法保存到库”。实际传入对象只有：`{id:e.id,name:e.name,description:e.description||"",thumbnail:r,image:r,imagePath:e.bound_image_path||""}`；type=e.type||"character"。不存在此调用处的完整 prompt/voice 复制。success 后提示并强制重载对应全局库 `_a(true)/Ca(true)/Pa(true)`；返回 false 优先 message，异常通用失败；无 API 不写。此 caller 不显示确认窗或重复名检查。

AI 聊天右键“保存” 449656–658（角色）、451485–510（场景）、451701–726（道具）→ Og 448375–387 → `save_asset_image_to_local(imagePath,name)`，**仅本地保存图片**。cancelled 或 message=“未选择文件夹”静默取消；不可把此按钮改造成全局存库却宣称完全还原。

`ey` 449485–523 与 V2 `Nn` 401290–347 是真人素材上传链的角色库更新 helper：load_characters 按 trim 名称精确找首项，复用旧 id，保留旧对象再覆盖新字段，携带项目来源、图片、声音/服务字段，save_character 后检查 success。这是独立调用链，不能据它推断 `save_ai_asset_to_global_library` 原生冲突规则。没有执行真人上传。

## E. 普通全局库页面规则

角色页 65439–65520：分类从当前列表 category 去重排序；搜索 trim+lower 匹配 name/description/audioTags；按 id 首个数字排序，默认 desc；显示 small 每页200，其他100。加载 load_characters；save_character 成功才重载、关编辑、成功提示（65526–37）；batch_import_character_library 成功按 count 提示/刷新，cancelled 抑制失败（65626–39）；原生导入去重规则未恢复。

通用资产页 266898–267143：按选中 category，name.toLowerCase().includes(search.toLowerCase())，这里搜索不 trim、不搜描述；load_assets(category)。上传 select_asset_files(category)，逐文件 save_asset(category,{name:路径末段含扩展名,source_path:path})，成功 unshift；部分失败仅计成功数；取消 message“未选择文件”不报错。delete_asset(category,id) 成功删除前端项；确认/系统选择器效果需要 UI 或原生证据补验。编辑 G 267120–32 要 name.trim 非空，但 save_asset 返回 false 被忽略，只有异常报错。原生文件复制、删除联动和重复名规则仍未知。

## 未恢复与 CM 独立实现选择

原生 `api.pyd` 隔离导入结果见 `native-library-api-boundary.json`：eager auth 依赖尝试读取账户配置、外部写入、创建 socket 等，审计全部拒绝，最终导入失败。没有账户数据库读取、SQLite 连接或网络连接成功；没有执行 Api 构造或业务保存。不能以这个失败导入声称已经恢复原生字段归一化/冲突/文件持久化。

以上安全结论以 Python 审计日志和脚本阻止点为边界；Python audit 不覆盖任意 native 直接系统调用，不是系统级隔离。`__new__` 只能绕实例构造，不能绕模块级 eager auth；本轮没有放宽真实配置/网络/SQLite 权限，也没有继续猜造认证依赖。

明确未知：save_ai_asset_to_global_library 内部字段转换，角色/场景/道具 ID 分配或复用，重复名覆盖/追加规则，存储文件是否复制、删除库是否破坏项目引用，原生 load/save 返回完整字段和最终落盘行为，资源跨重启可用性、项目关闭后的引用寿命。也没有原版当前窗口截图验收。

CM 若采用 MySQL 持久普通库、稳定 ID、独立库文件存储、显式复制、重名确认、严格 success 校验、错误可重试，这些属于**我们自己的可靠性/存储适配**，不能标为原版已观察。G12 必须保留上述 A/B/C/D 不同语义；“全字段快照导入当前资产”不能替代 B/C 的仅图片/引用绑定。付费模型与第三方主体上传不在本轮验收范围。
