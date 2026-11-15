# G12主体存档：实现、原版规则和实际运行结果

代码提交`23bbdd92e0677d2156c0c0906f56bf2328b1101a`；本地4001已运行该版本，PID18276，入口`/assets/user-BoBJNXvo.js`，漫剧分块`ComicScreenplayPage-DKHfpruG.js`。HTTP前端资源逐字匹配构建，旧运行程序保留在本案`library-subject-20261116/rollback-73368/`。没有发布公网。

## 已接通的用户动作

普通角色编辑的参考音频下方有主体存档文字框及复制图标；读取旧档可恢复，取消不保存，应用后可重开。新建/编辑/换图/换音频全部保留当前值。省略该字段的旧请求不擦除，空串明确清空。文字不假设一定是火山ID，支持JSON文本，最多16KiB。

普通库选入新项目角色时存档随角色资料复制；仅换图片不会覆盖原角色文字和存档，库中后续修改不反写旧作品。项目角色保存回普通库时也保留该值。V2导出覆盖原档中的旧camel/snake值，未知资料继续保留。

## 原版证据和对照

主案目录`07-parity-execution/library-subject-20261116/`：

- `original-editor-replay.json`：原恢复函数A/z原文受控重放7案，复制成败、Unicode/JSON文本、缺名称/图片、忙时重复操作；不是原桌面点击。
- `frontend-sora-consumers.txt`：原普通库绑定178096-178106和其他不同拖放分支摘录，不能混为同一写入形态。
- `native-subject-interface.json`、`native-synthetic-roundtrip.json`：原编译上传入口api.py声明10320，配置分派及自造响应整理。原函数和请求均在隔离进程，连接被阻止；不存在真实供应商成功结论。
- `native-upload-boundaries*.json`：原空组会创建新组，CreateAsset缺ID失败，GetAsset明确Failed失败，得云c39/c40和密钥优先级；空查询原轮询在6请求处由测试截断，最终超时仍未知。

## 验证与修正

后端普通库聚焦、screenplay及characterarchive全包通过。真实MySQL独立账号28项通过（无迁移、无配置修改，媒体为内存替身），见`backend/mysql-result.json`。

前端接口6项及真实React DOM通过，包含普通库新增复制、取消、旧项目仅换图、原键重试和晚回复隔离。独立复查发现并修复两个实际可见问题：普通input吞换行、复制提示被Modal遮罩挡住；最终DOM先红后绿，提示在当前窗口可见，关闭/重开后迟到成败不污染。另修复SaveSource Extra浅拷贝可能污染来源项目。红绿日志及复查边界见`frontend/RESULTS.md`和`backend/README.md`。

未编辑直接保存/复制保留原CRLF；textarea显示为LF，主动编辑后保留行结构并按浏览器LF保存。此差异在规格27明确，不假称主动编辑后字节完全不变。

`actual-4001-result.json`：16项实际签名HTTP检查通过，独立测试账号，含旧值省略保持/明确清空、新存档保存、版本冲突、超限、项目角色Extra实际创建与重读、库修改不影响项目、真实导出及清空验收库。真实用户资料写入0、模型调用0。

`native-live-subject-archive-result.json`：该4001导出的归档由原编译解码器读取，当前camel/snake值与未知嵌套资料一致，旧值未复活。

## 未完成范围

真人素材上传服务及最终视频消费仍未实现，不能以存档保存完成关闭整个主体功能。当前自家真人素材渠道只是配置，下一步使用自有配置实现真实创建/激活查询与错误链，不植入原账户凭据。YDY在此原安装缺模块，不假称可用。

原桌面当前可读界面树，但截图返回`window capture timed out`，重新获取窗口后角色库导航返回`coordinate input geometry is unavailable`且页面未变化；没有重复点生成或上传，没有原桌面视觉通过证据。全部36组仍未完成。
