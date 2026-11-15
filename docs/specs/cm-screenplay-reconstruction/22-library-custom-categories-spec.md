# G12-CAT：普通角色库自定义分类

用户已授权持续逐项复刻、逐项对比，本项直接实施，不再次审批。范围只覆盖角色编辑窗的分类，不替代整库导入导出或所有G12验收。

## 原版规则及定位

原恢复AddCharacterModal片段：`07-parity-execution/asset-library-original/manager/source-excerpts/character-editor.txt:40–97`，完整恢复JS64973以后。13个内置分类固定；额外分类独立load_custom_categories/save_custom_categories。分类选择可搜索、清空、输入创建；新增立即保存。仅自定义选项右侧有“删除”，点击不触发选中；删除立即保存，若是当前角色草稿选择则清空。取消角色编辑仅丢弃角色草稿，不撤销已经保存的分类。

静态证据不等于原版真实存储及视觉验收。原版保存失败仅记日志的行为不照搬为无提示：CM明确提示失败，未确认的变化不显示成功。

## 输入/输出契约

- GET `/api/comic-screenplays/asset-library/categories`：当前登录账号的`builtin_categories`、`custom_categories`、`revision`。
- POST同路径：`{action:"add"|"delete",name:string,expectedRevision:int}`，返回最新上述状态。
- 名称按精确文字去重；只用trim判空，不偷偷修改用户文字。长度不超过现有素材分类字段191字，禁止控制字符。内置分类不允许删除，新增同名内置项不重复创建。
- 账号隔离、MySQL持久化，独立于任何角色和项目。删除自定义选项不批量修改已有角色档案，也不删除图片；历史角色仍保留分类文本，直到用户明确应用修改。
- 重复新增/重复删除返回已确认状态。不同变化的旧版本明确冲突，前端先重新读取，不整表覆盖。拒绝空名称、未知操作、非法版本；读取/写入错误显示可理解提示，不假称成功。

后端公共类型：`AssetLibraryCategoryState{CustomCategories []string, Revision int64}`；独立`AssetLibraryCategoryRepository`提供LoadAssetLibraryCategories(ctx,owner)和ChangeAssetLibraryCategory(ctx,owner,action,name,expected)。由MemoryAssetLibraryRepository及MySQLStore实现，不扩大原AssetLibraryRepository以破坏已有测试替身。服务从现有Repository断言能力，缺失时明确不可用。

## 界面与取消

角色编辑分类由输入建议改为可搜索、清空、创建的单选；自定义选项有右侧“删除”，内置项没有。保留当前历史分类，即使不在下拉列表也不自动抹掉。创建/删除独立保存，明确说明；关闭角色编辑后重开应读到已保存的分类。分类持久化不等于角色名称/分类已经应用到档案。

忙碌时防止重复操作；失败保留旧列表和用户输入，允许核对后重试；禁止旧异步读取覆盖最新操作。取消编辑或切页时不写角色，不让迟到回复写入另一份编辑草稿。

## 实施与验收顺序

1. 后端先写失败测试，再补账号隔离、内置保护、版本和幂等、MySQL事务及最小迁移。
2. 前端先用真实React DOM写失败测试，再补API与独立分类组件，接到角色编辑器；现有换图/音频/目录导入回归。
3. 独立审查；实际HTTP/MySQL测试新增→取消角色→重开→删除→重读，以及失败、空值、重复和跨账号。
4. 构建Go embed，只在确认当前任务空闲后本地4001运行验收；防止与并行任务互相覆盖。
5. 单项台账分别记录原静态对比、受控DOM、真实持久化、运行包与原版实际操作。未完成的层级保留待验，完整36组目标不关闭。
