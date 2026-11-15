# G12 public 文件夹批量导入：已恢复规则和未知边界

本规格为静态恢复，未启动原软件、未调用native导入函数、未打开真实账户/数据库、未联网或调用模型。证据来自原归档恢复前端与**已存在的提取文本**，不把CM实现选择当成原版行为。

## 原件和入口

恢复前端：`_local-archive-20261028/luoshui-screenplay-reconstruction-20261115/01-readable-frontend/pretty/index-DszfAkKD.js`，SHA256 `8c5145ee0f26f594c35c6ca954e24d87dbbb44f5506029343f211a374992e492`；下述全部原行号指向这个文件。

主菜单角色库`characters`（611502–506），内容分支渲染`CharacterLibrary`（612793+；组件65427）。组件默认tab=`personal`（65439），两个tab分别public“角色库”（65686–89）、personal“个人档案”（65710–14）。本功能在public；它不是已观察到的在线公共人物目录，旁边说明“本地模式，请使用批量导入默认角色库”（65691）。

public布局65690–65707：左侧本地模式说明，右侧primary“批量导入角色库”，图标rJ；其下有使用说明块，再下是正在导入spinner或空说明。personal保留角色列表功能，详情见asset-library-manager-spec.md。

## 按钮、状态、桥接口

按钮65692–97 `onClick:H`，H完整函数65626–39：

1. 无`window.pywebview.api`：error“此功能需要在应用程序中运行”，返回；不会开始busy。
2. `x.value=true`，调用 `await window.pywebview.api.batch_import_character_library()`，**前端没有传参数**，没有传目录路径/扩展名/嵌套/排序设置。
3. success：提示`result.message || 成功导入 ${result.count} 个角色`，`await T()`重新加载角色库，随后`o.value="personal"`切到个人档案。前端未检查count是否正数，也没有对count做数字归一化。
4. 非success：`result.cancelled`为truthy时不提示；否则error `result.message||"导入失败"`。返回null也走通用失败。此分支不会主动切tab或T刷新。
5. 抛异常：日志“批量导入角色库失败”，error“导入失败: ”+`error.message||error`。
6. finally：`x.value=false`。函数没有`if(x.value)return`重入守卫；按钮也未绑定disabled/loading（65692–97）。busy仅改变下方内容，不证明重复点击被阻止。

busy内容65704：`x ? spinner(size:large)+正在导入角色库... : Empty(description:点击上方按钮开始导入角色)`。这里不是逐文件百分比、不是每文件错误面板。未观察到本功能独立取消按钮、进度回调、后台job、离页/关窗续看、失败恢复或requestKey。全局真人/JM同步progress回调65656–78是其他功能，不能套给此导入。

T完整65508–20：busy p置true，load_characters()，仅success时`v=data||[]`；错误只日志，finally恢复p。故H即使导入成功、后续T返回失败，也可能显示成功并切personal但保留旧列表；不能凭toast证明全部已落盘。

## 前端说明文字明确承诺的范围

65700：“选择一个角色图片文件夹（如：古言）”。65701：“文件夹名称作为分类（如：古言）”。65702：“角色自动命名为分类名加序号（如：古言1、古言2、古言3...）”。这些文字可以用于还原界面说明，但未执行native导入，**不能据示例证明序号从1开始、每次重置、连续编号或按任何固定顺序分配**。

与其他入口不要混用：个人档案“导入档案”j65550–60调用`import_characters()`；public调用`batch_import_character_library()`。普通场景/道具“上传”w266921–50调用`select_asset_files(category)`后逐文件`save_asset`，默认名称含扩展名；这些语义不能移植成public native的已恢复规则。

## CloudDownloadOutline 原SVG

public按钮使用rJ（65695），原定义55444–78：tJ为512×512 viewBox；nJ/aJ/oJ为三条path，fill=none、stroke=currentColor、round linecap/linejoin、stroke-width=32；rJ name=`CloudDownloadOutline`（55473–77）。按实际path树补录到`frontend/src/user/pages/comic-screenplay-icon-data.js`，没有用通用图标猜形。

压缩原包`00-original/dist/assets/index-DszfAkKD.js`组件定义span：line29、column1363856、start1472295、end1472381（字符索引，非native偏移）。SVG树SHA256 `548411eee2b38aac5c8709216b85a91d768a222becd2d35b532ff46e7989fe58`。证据`manager/visual-excerpts/cloud-download-icon-added.json`包含完整tree、原文件SHA和定位。`add_folder_icon.py --cloud-download`静态提取，可重复验证但不执行原代码。Icon.test七项通过，其中本项验证三条路径、描边和原几何SHA。

## 现有native提取文本检索结果

证据目录：`07-parity-execution/asset-library-original/manager/folder-import-evidence/`。`search-manifest.json`保存检索范围、118个已提取UTF8文本文件的路径/字节数/SHA与命中，0个超过上限或无法解码；方法名`batch_import_character_library`未命中。范围为00-original/modules内可读文本、01-readable-frontend/modules内可读文本、02-native、04-analysis、evidence、scripts及既有candidate目录；只检查py/pyi/txt/md/json/csv/log，不加载PE，不重新做反编译。已知前端caller另以四组行片段保存，不把它当native正文。

`02-native/status.json`记录原native为PE x86_64、Nuitka marker、not_fully_decompiled；此前Ghidra330秒与900秒provider_timeout，没有可用伪代码。本轮没有重复这些调用。以前asset-library原生api隔离导入被eager auth挡住，本轮也未重复。**“已提取文本未发现”仅覆盖上述manifest，不代表二进制不存在方法或原版没有实现。**

## 明确未知的native规则

允许图片扩展名与大小限制；文件选择器是否只选目录；同层/递归/子目录分类；文件枚举顺序、自然排序或导入完成顺序；序号起点、重试/再次导入编号；同名分类/同名人物/同一文件的去重与覆盖；图片是否复制、压缩或生成缩略图；新ID规则；损坏文件/不支持文件怎样报错；部分成功是否回滚；取消发生在选择阶段还是导入中；最终落盘与跨重启、原图删除后的引用寿命。这些均不得猜成原版事实。

## CM适配边界

浏览器目录选择、多文件上传、确定性排序、扩展名白名单、按顶层目录取分类、编号策略、每文件错误/成功列表、重名处理、防重复请求、服务端MySQL持久与图片托管、取消与重试等必须由CM明确选择并测试；本文件不替root指定这些策略。界面可以沿用原public说明与成功后切personal，同时必须区分“原前端静态已观察”“native尚未知”“CM独立实现选择”。不得用普通逐文件上传或整对象快照导入的实现测试宣称native文件夹导入全部一比一通过。
