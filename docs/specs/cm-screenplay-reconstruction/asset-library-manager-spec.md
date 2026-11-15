# G12 普通资产库管理页：静态恢复规则

2026-11-15。本规格来自已恢复前端字节，未启动原软件、未操作账户或真实库、未运行模型。所有行号指向 `_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/01-readable-frontend/pretty/index-DszfAkKD.js`，SHA256 `8c5145ee0f26f594c35c6ca954e24d87dbbb44f5506029343f211a374992e492`。这些是静态推导，不是本轮原版截图或实际存储验收。

证据目录：同归档 `07-parity-execution/asset-library-original/manager/`。与当前作品的新增/换图/绑定规则参见 `asset-library-spec.md`；这里不把头像参考库、Canvas 库、JM 主体库等替代为普通资产库。

`source-excerpts/manifest.json` 固定六组完整组件/入口片段的原行号与SHA，`recover_manager.py` 可重建。实际 REA 派生静态分析 `rea-analysis.json`：`ev_048fdb7e98ef72806ecb1854e96900427b890736de99e52638ca74b93a67ae96`，1文件/71687字节/14108 AST节点/0解析失败；语义关系总体coverage=unknown，动态闭包关系partial，无截断，不证明运行效果。Craft独立工程`manager/reconstruction.cany`及可导入`asset-library-manager-craft-bundle.json`已保存；bundle SHA `22c4d87399efe6e3c1479c5f1f2631f75b62bd7fe47cd7d2ecb2045114e44d52`，source inventory `ev_3f2314364f53e26269acc723d7a63c9967cfd25e3924cc5548d3b3a382a34bdb`，derived inventory `ev_23a523add763d59f8ba53d1bd9bc2a82e4e521557d3d9ae9e2f5d8e7a80dec8c`。REA记录另存，不冒称已作为Craft自动解析记录导入。

## 页面入口与布局

主菜单“角色库” key=`characters`、图标标识 `D2`（611502–506），主内容 `characters` 分支渲染 `tee=CharacterLibrary`（612793+；组件名65427、封装66081）。它有两个 tab：`public`“角色库”与 `personal`“个人档案”，默认 personal（65439；65680–65714）。public 是**本地文件夹导入说明页**，并非静态证据中已实现的远端公共目录。

`MyAssets` 组件名266820；主内容 `assets` 分支612872–875；标题“我的资产”，副文案“管理您的音效、BGM、特效、场景和道具资源”（267145–146）。普通场景/道具库在此组件六个分类文件夹中。顶层分类卡是文件夹图标 `IX`，80×64 区域内图标40，数量大于0显示角标（267317–339）；分类内部顶部显示对应类型图标18。

```text
角色库
  [角色库 public] [个人档案 personal（默认）]
  public: 本地模式说明 + [批量导入角色库] + 文件夹/分类/自动命名说明
  personal:
    共N个档案 | 大小 | 分类 | 排序 | 搜索 | 同步真人素材 | 同步JM主体
              | 导入档案 | 清空角色 | 导出档案 | 新增档案
    图片卡片网格：图、名称、试听/编辑/删除、分类标签、最多2个音频标签
    分页（超过每页数量时）

我的资产
  标题与说明
  顶层：搜索框 + 六个分类文件夹与计数
  进入场景/道具：←返回 | 分类图标与名称 | 搜索 | 上传
    140px卡片（124×124缩略图），名称；悬浮播放图标/删除图标
    点击卡片：编辑窗（560px）
```

角色尺寸已由配套原CSS补齐（见下一节），图标也由原定义名映射，均非运行时截图或像素验收。

### 配套 CSS 与图标原件补证

匹配当前角色组件scope=`data-v-abff22ea`、编辑窗scope=`data-v-197a0341`的原文件为 `00-original/dist/assets/index-BbbJnhtu.css`，SHA256 `e0e7e16ac35b7f99b354c0b2f89295e459b4fb36aa7957454f8ac689ede82211`。另两份CSS中存在旧scope=`9b58020b`，不据它替代当前规则。`manager/visual-excerpts/css-rules.json` 保存35条相关原规则、原文件行号及字符起止偏移；minified文件用字符偏移定位，不能把派生JSON行号当原CSS行。

| 档位 | wrapper/card宽 | card高 | 图片区高 |
| --- | --- | --- | --- |
| standard | 140px | 196px | 154px |
| large | 196px | 274px | 216px |
| small | 112px | 157px | 123px |

`.character-grid` 为display:flex/flex-wrap:wrap/gap:12px；当前scope规则没有响应式媒体断点，按可用宽度自然换行，并非固定列数。卡片content-style padding:0（JS65797），信息区`px-2 py-1`（65411）对应左右.5rem、上下.25rem；名称`text-xs`=.75rem/line-height1rem。标签区最高24px。外层`p-6`=1.5rem，工具条gap-3=.75rem、mb-6=1.5rem。rem值不擅自换算成固定px，当前根字体/缩放未运行核实。

卡片纯黑背景#1e1f23、边#2d2f36；暗色#1e293b、边#334155；亮色#fff、边#e5e7eb。图片区纯黑#141518、暗#111827、亮#f3f4f6。原`.modal-pureblack .n-card`背景#111113且border:none；`.modal-dark .n-card`背景#1e293b；预览纯黑#0c0d0f、1px rgba(128,128,128,.12)边，暗#1f2937、亮#f3f4f6。

角色编辑窗900px由JS65286设置；flex两栏gap-6=1.5rem、左w-72=18rem、预览h-96=24rem、右flex-1（64972、65301–22）。没有本scope手机改列/弹窗宽度自动适配规则，Naive组件实际内部间距/运行视口仍需单独验收，不能以这些固定声明宣称小屏已完整适配。

`manager/visual-excerpts/icon-map.json` 保存原JS定义片段、原行号、原SHA及当前CM图标catalog哈希快照。如下name均从原`ea({name:...})`直接读取：

| 原标识 | OriginalIcon name | 原定义行 | 当前CM catalog快照 |
| --- | --- | --- | --- |
| D2 | PeopleOutline | 57332 | 已存在 |
| EV | AddOutline | 54690 | 已存在 |
| XJ | CreateOutline | 55675 | 已存在 |
| P8 | TrashOutline | 58519 | 已存在 |
| rJ | CloudDownloadOutline | 55473 | **缺少，需另外按原SVG静态补录** |
| vJ | CloudUploadOutline | 55507 | 已存在 |
| IX | FolderOutline | 56337 | **缺少，需另外按原SVG静态补录** |
| QY | ImageOutline | 56607 | 已存在 |
| dZ | CubeOutline | 55751 | 已存在 |
| W5 / x2 | PlayOutline / PauseOutline | 57465 / 57259 | 已存在 |
| e3 | SearchOutline | 57925 | 已存在 |
| CX | FolderOpenOutline | 56314 | 已存在 |
| UJ | CopyOutline | 55649 | 已存在 |
| j1 | MusicalNotesOutline | 57105 | 已存在 |

catalog是否存在是捕获时快照，后续实现可能新增，不作为永久缺失断言。本证据worker未改业务图标数据。Craft新增visual inventory `ev_2537e2ea2f97acfda33cd4c8194507e098eff87556ec02c60fc99d5640237a26`，独立可导入`manager/manager-visual-craft-bundle.json`，SHA256 `6beccc2945e5071e48264af267cd30b104725b120ca47ea824a757c98df69cd6`；之前manager bundle未覆盖。

## 角色列表与筛选

`T` 65508–20 在 mounted 时执行（65674），`load_characters()`→`{success,data}`；success才 `v=data||[]`，busy p finally恢复，失败返回不提示，异常仅日志。未提供 api 则空列表且解除busy。无分页参数，全部数据在前端筛选分页。

65439–65507：默认 size=standard，选项标准/放大/缩小；pageSize=small?200:100。分类取列表中 truthy category、Set去重、排序，前置“全部分类”。搜索 trim+lower，name/description/audioTags 任一匹配；排序用 id 中首个数字，找不到为0，默认 desc，asc可选。大小/分类/搜索改变均 q()重置页1，排序变更 handler 不调用q（65716–46）。计数 G 是筛选后条数；页数ceil(G/pageSize)，G>pageSize才显示（65901–10）。清空/删除后没有显式矫正当前页，不能声称自动返回有效页。

卡片（65790–65900）image=`thumbnail||image`，lazy加载；无图人物图标QY40。名称截断；显示category以及audioTags.slice(0,2)，标签区最高24px（65411–14）。整卡点击（65798–65809）先按imagePath调用read_image_file(path)，success显示返回data；失败/异常fallback item.image；没有两者不打开预览。大图弹窗宽auto、最大90vw，图最大80vh（65931–44）。

## 角色页按钮 → API → 结果

| 按钮/入口 | handler和原行 | 输入/输出及状态 |
| --- | --- | --- |
| 新增档案 EV | D65522–25；按钮65781–88 | g=null,h=true，打开AddCharacterModal；本身不写库 |
| 卡片编辑 XJ14 | 65839–53 | g={...item}，h=true；修改草稿，保存才save_character |
| 单项删除 P8 14 | 65854–76 | **无前端确认**；delete_character(id)；success过滤所有同id项并成功，false message错误，异常通用失败；无api静默无操作 |
| 试听 W5/x2 14 | 65822–38 | 有audio才new Audio(audio)，同id再点停止归零，切换先停旧音；结束清播放id；无audio warning。无桥调用 |
| 导出档案 vJ | Q65538–49；65776–80 | api缺失错误，原始列表v为空warning；export_characters()无筛选参数，故不能称“仅导出当前筛选”；success message或count；cancelled静默；false/异常错误 |
| 导入档案 rJ | j65550–60；65763–67 | import_characters()；success message/count并T重载；cancelled静默，false/异常错误；文件格式/合并冲突native未知 |
| 清空角色 P8 | J65640–51；65768–75 | v为空warning；window.confirm含原始总数及不可恢复；取消不调用；clear_all_characters()成功T重载；false/异常错误 |
| public批量导入 rJ | H65626–39；65691–65707 | api缺失错误；busy x；batch_import_character_library()；success message/count，T重载并切personal；cancelled静默；false/异常错误，finally解除busy |

public说明（65700–02）是“选择一个角色图片文件夹（如古言）→文件夹名称作为分类→自动命名古言1、古言2、古言3…”。这只是前端说明文字；实际扩展名过滤、排序、冲突/追加、命名起点与导入事务尚未恢复。无批量勾选删除/批量编辑证据，不要将清空等同于勾选批删。

## 新增/编辑角色表单

AddCharacterModal 64973–65404，宽900px、mask-closable=false，标题新建/编辑档案（65075、65281–88）。左侧w-72图片区，预览高h-96，可点击选图；编辑时显示“打开文件夹”且没有imagePath禁用。右侧名称、分类、上传图片、参考音频、主体存档、描述；底部取消/应用（65301–96）。主体上传/同步是有网络风险的独立能力，未执行。

初始字段（64992–65009、65260–78）：id=null,name="",type="character",category="",image="",description="",audio="",audioFileName="",sora2Info="",volcengineAssetId/Uri/GroupId/ProjectName/ImageUrl/Url="",volcengineAssetData=null,realPersonAssetProvider="volcengine"。imagePath由选图动态赋值。编辑时浅拷贝完整旧对象，未列出的旧字段保留；兼容字段补默认（65253）。因此不能只保存界面可见字段丢掉旧记录的其他引用。

每次打开先get_custom_categories()（65230–43），无api读localStorage.customCategories；失败仅日志。内建分类 65011–23：古言、现言、男生、女生、男主、女主、男配、女配、小孩、中老年、老人、其他职业、动物；再追加自定义。分类下拉clearable/filterable/tag（65329–39）；创建v65028–43仅非空且与内建/自定义不重复时push并save_custom_categories(array)，无api存localStorage；没有trim。删除自定义项p65044–74会立即保存分类列表，当前表单若选中它则清空category；**不删除角色，也不批量修改旧角色category**，取消表单不会撤销已经保存的分类变化。

选图 P65109–20→select_image()，要求success&&data，写image=data,imagePath=path,预览=data；名称trim为空时，取文件basename去最后扩展名再trim自动填名称。取消/false无变化，异常“选择图片失败”；未恢复native格式限制，不从说明推测。选音R65121–32→select_audio()，只检查返回对象truthy，不检查success，赋audio=data、audioFileName=fileName并停试听。故truthy取消对象可能清已有音频，这是原静态缺陷；不能声称取消绝对无变化。试听M/O65133–46；无audio warning。

打开文件夹I65147–64从imagePath求父路径，open_folder(parent)；空路径warning，false/异常报错。复制主体存档A65199–208走clipboard.writeText，空warning/失败错误。每次开窗还load_global_settings读取realPersonAssetProvider（65243–52），没有因此授权本轮读取真实设置。

应用z65210–26：busy f时忽略；要求name.trim非空及(image||imagePath)，但**实际发出的name不trim**。busy后等50ms，浅拷贝全部r，id为空生成`char_<Date.now()>`并写回r，emit("save",object)。父L65526–37→save_character(object)，success才T刷新、关窗、显示修改/添加成功，false message错误保留窗，异常通用错误。子emit不是await父异步保存，f很快解除，**不是可靠的整个请求防重复锁**；没有前端同名检查。取消_65076–78 emit update:show=false，关闭不save_character；重开重新建/复制表单。

真人素材按钮N65165–97调用upload_character_real_person_asset({image,imagePath,name,description,groupId})，成功回填主体字段但仍需普通应用才save_character；复制存档/上传与普通本地图档管理不得混称完全相同。具体服务端返回/付费/授权不在本轮验收。

## 普通场景/道具：MyAssets

类别定义266826–875：sound-effects/音效(q8,#60a5fa)，bgm/BGM(q1,#f472b6)，effects/特效(A3,#a78bfa)，scenes/场景(QY,#34d399)，props/道具(dZ,#fbbf24)，styles/风格(CJ,#f97316)。每个含count/items。mounted **267133–143**按六类顺序load_assets(category)，success写data||[]与count；单类异常仅日志，继续下一类。点击文件夹267320–23再次k(category)加载；返回x266917–20清a与搜索。

列表搜索y266898–902只按当前分类item.name.toLowerCase().includes(o.toLowerCase())，不trim、不搜prompt/description、不排序、不分页。顶层搜索框267306–17虽然绑定o，但分类卡遍历n.value，并没有搜索过滤；此是静态缺陷，不能称全库搜索已恢复。

普通卡片267192–276：宽140px、图124×124，img只读thumbnail_url，否则类别图标24；图片error隐藏。hover显示操作按钮：W5播放12没有onClick，因此点击会冒泡整卡编辑，不能宣称播放器实现；P8删除12有stop，直接S(item)。卡片点击浅拷贝p={...item}、v=true。右键菜单仅styles（267205–12、267278–90），不适用于场景/道具。

| 操作 | handler/输入 | 成功/空/取消/失败 |
| --- | --- | --- |
| 上传（EV图标） | w266921–50；按钮267178–90；空提示267299 | api不存在错误，未选类错误；select_asset_files(category)返回success&&data.length>0后逐文件顺序save_asset(category,{name:basename含扩展名,source_path:path}) |
| 多文件上传结果 | 同上 | 每条success将response.data unshift并更新count（最终显示顺序为成功返回顺序的逆序）；部分false仅日志，不回滚已有成功；总成功>0显示数量，否则文件保存失败；选择取消message=未选择文件静默，其他message warning，异常结束整个循环并错误 |
| 删除 | S266982–267007 | delete_asset(activeCategory,id)，无确认；success按首个匹配id splice/count，false message错误，异常通用错误；无api仍前端移除并提示成功，持久化未发生 |
| 编辑窗保存 | G267120–32；UI267486–528 | name.trim非空才save_asset(activeCategory,p整对象)；实际name不trim，**不检查返回success**，随后success toast/关窗/k重载；只有抛异常才错误；空name warning |
| 编辑窗取消/关闭 | 267488、267515 | v=false，不save_asset；p浅拷贝的name/prompt修改不应用列表；无撤销已有上传/删除 |

编辑窗宽560，标题“编辑场景/道具”，有thumbnail_url才显示预览；可改**name与prompt**，prompt标签“提示词/描述”且可选，**不是description字段**；file_path有值才显示只读（267492–514）。没有此窗替换图片/声音、迁移分类或重新生成的证据。

备用上传弹窗267340–376包含分类、Upload拖拽区multiple/directory-dnd/max10、取消/确认，handler `_`266951–81实际上重新调用select_asset_files(r)，而不是用Upload列表i；完成/失败后都关窗并清i。当前正常顶部按钮直接走w，所读范围中未见打开l=true的用户入口，不能称备用窗当前可达或将max10当作direct native文件选择限制。

原静态缺陷：267278–305是`styles ? 右键菜单 : 暂无资产块`，并不检查y.length；故场景/道具有卡片也渲染“暂无资产/上传资产”。不应将这个文案作为实际空库判断。

## 风格库边界

styles同页但不是场景/道具元数据模板。顶部“新建风格”清空{name,prompt,images}后开560px窗（267181–85、267377+）。P267009–43→select_multiple_images，追加到images；无api用浏览器多选data URL。R267044–85要求name.trim，save_style({name,prompt,images:[{name,data:url,path}]})，success追加返回项，false错误但仍关窗；无api仅内存项Date.now id/createTime。右键W edit/rename/delete只styles；rename M267086–92使用prompt输入名称，trim后改原对象并A save_style，A267103–15忽略false且捕获异常仅日志，所以rename仍可能提示成功。风格编辑z267116–19同样不能凭toast确认落盘。

## 外部同步、重复与未恢复项

个人档案上可见同步真人素材/同步JM主体（65747–62），分别F65570–85→list_volcengine_real_person_assets()，选Active的Id→V65589–65603 sync_volcengine_assets_to_characters(ids)；U65604–25→sync_jimeng_subjects_to_characters()。进度全局回调current/total/percent/message/currentName在65656–78装卸。这些是外部主体服务链，不是普通本地CRUD，本轮只记录边界，不连接或检验服务。勾选仅在该外部同步弹窗，不是角色普通列表的批量勾选。

角色save/import/batch import，场景道具save_asset均无前端重复名称检查；不能据此推断native重复名一定允许、一定覆盖或一定跳过。原生ID分配、图片压缩/thumbnail生成、音频编码、分类落盘、文件导入格式、文件复制与源删除、delete是否影响项目引用、跨重启持久性仍未知。G12前次api.pyd隔离导入被eager auth阻挡，本轮不重复native/Ghidra，不扩大授权。

CM自行选用MySQL、可靠锁、校验success、重名确认、文件托管、有效空状态与可用搜索等应标明是我们自己的适配/修正。源码静态恢复、CM实现测试、原版窗口/像素对照、真实存储验收必须分别记录，不能用其中一项冒充全部完成。
