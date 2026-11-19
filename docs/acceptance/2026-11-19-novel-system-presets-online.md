# 小说系统预设线上核验

用户明确要求上线。核验发现生产现行版本已包含本轮改动，因此本次没有重复上传、替换或重启线上服务。

- 公网：`https://yzcm.site`，health返回`ok:true`。
- 公网版本：`a42de82e1d0a688ea6b88abc0c607285aa5e79ec`；Git祖先检查确认包含小说预设实现提交`a0a43bb8`。
- 前端入口：`/assets/user-BVnCvltX.js`；公网运行版本与服务器runtime-release.json一致。
- 服务器：`qiantie.service`处于active，执行路径`/opt/www/qiantie/qiantie`。
- 实际生产目录：`/opt/www/qiantie-data/system/presets.json`，小说创作24条、已发布24条。仅只读核查，没有复制本地目录覆盖线上配置。
- 已登录浏览器打开`https://yzcm.site/admin/presets`，选择小说创作，实际显示六个分区；展开审稿与修订，章节审稿、原生修订基础词、局部修改/整章重写分流词均显示V1已发布。
- 本轮不消耗模型额度，不变更线上书籍、队列或用户配置。线上核验是入口、目录、版本及健康检查；发布A/B和MySQL快照验证见`docs/specs/2026-11-19-novel-system-presets.md`，不宣称线上真实模型端到端试生成。
