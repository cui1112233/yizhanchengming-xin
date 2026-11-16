# G12真人素材上传：规格与实施步骤

本项承接27/28，补真实服务，不能以配置字段或模拟成功冒充外部上传。继续完整36组目标；本轮只做零外连测试，不消耗供应商额度。

## 原版规则

原Api.upload_character_real_person_asset声明api.py:10320（代码元数据，不是作者源码）。恢复前端65164-65201：需要图片、非空名称；上传成功把asset_id写入存档，保存group/project/provider等资料；失败不覆盖旧值；上传结果只进入草稿，点击应用才保存角色。空group火山按CreateAssetGroup→CreateAsset→GetAsset执行，返回组ID供角色下次复用。CreateAsset无Id或GetAsset Failed明确失败。得云c39/c40请求/v1/sd/assets，原配置和自家字段不相同，必须显式映射。原请求及自造响应记录在library-subject-20261116/native-upload-boundaries*.json。原最终空查询超时、实际供应商及视频消费仍未证明。

## 自有配置与协议

使用用户API接口配置realpersonChannel、volcanoProjectName/AccessKeyId/SecretAccessKey、deyunaiRegion与seedance2Api。不读原账号。支持已观察火山、得云两种协议，其他已列渠道明确暂未实现，不能静默换渠道。图片先用已有TOS ReferenceURLProvider，不能得到公网HTTPS地址时用用户配置的imageHostUrl上传；新增独立imageHostApiKey作为可选Bearer，不能把地址误当密钥。缺配置在任何外发前拒绝。所有网络请求使用现有安全HTTP工具，拒绝重定向、内网目标，测试明确注入本地替身。

火山空项目名按自家设置页约定取default；操作摘要同时包含文件名及供应商凭据的不可逆指纹，避免恢复时悄悄换账户。源对象ID由账号/请求/摘要确定，不因重复保存结果不明而生成额外对象。

火山签名：HMAC-SHA256/cn-beijing/ark/request，Version=2024-01-01，POST Action=CreateAssetGroup/CreateAsset/GetAsset；字段按已观察记录。得云POST /c39或c40/v1/sd/assets，Bearer；结果未激活需查询，未观察到的查询协议不猜测成功。客户端实现上下文取消、有界等待/响应体、失败解析和凭据脱敏。

## 对外接口与持久化

- GET `/api/comic-screenplays/asset-library/subject-upload-capability`：仅返回provider/label/available/reason，不返回凭据。
- POST `/api/comic-screenplays/asset-library/subject-uploads`：multipart requestKey/name/description/groupId + file，或entryId/expectedRevision读取当前账号已存图片。40MiB图片限制，禁止服务端任意路径和远程URL输入。
- GET同路径`/{requestKey}`：核对本账号操作。结果为`{operation:{requestKey,status,phase,message,result?}}`。status为running/paused/needs_review/failed/succeeded；result仅安全的assetId/assetUri/groupId/projectName/provider/status。同键同内容恢复，同键换内容409；未知创建结果暂停核对，不盲目重复建组或素材。
- 每次操作独立持久化到MySQL（新可选repository及迁移，确认编号未占用），不把密钥存任务。源图片按已有对象存储保存；Redis每请求租约保护，运行上限180秒，小于租约300秒。先持久记录create意图，再外发；成功保存组/素材ID后再下一步。崩溃或创建响应不明保留needs_review，禁止重试创建。已有素材ID仅查询，不重复上传。
- 请求取消停止后续请求和等待，远端已经创建的素材无法靠取消撤销；保存可核对的进度。没有后台重试或未经用户动作的收费请求。
- 上传本身不写角色。应用角色时可传subjectUploadKey；必须核对同账号成功操作，将安全结果存入私有Extra，公开和V2导出对应安全字段。省略保留，后续手工存档仍可改，不自动覆写手工值。换图/音频与CAS摘要纳入该引用；取消角色不自动清理远端素材。

## 页面交互

主体存档复制按钮旁接“上传火山引擎素材”或“上传得云AI素材”，读取自有配置显示。无配置或不支持时明确原因，不能点击后假成功。上传前检查图片与名称；进行中明确进度，提供停止等待，保持旧存档。成功只填草稿与上传引用，应用才保存。失败保留草稿，核对同请求避免重复；关闭/重新打开、迟到结果和切账号不能污染其他草稿。需要人工核对的创建结果明确说明，不提供会盲建的新重试。

待核对操作按账号及角色保留在页面内存，新建角色有独立草稿槽；关闭重开先GET同请求，源图片/名称/描述不匹配不填结果，可显式恢复原内存File后再核对。整页刷新不保留私密File，不声称跨刷新页面恢复；服务端MySQL记录仍在。

进程中断的running由GET取得同请求Redis租约后核对：活跃租约不动；无活跃租约且有PendingAction转needs_review，无PendingAction转paused。已有AssetID仅查询，不依赖重新签图床URL。图片准备阶段失败为paused，可修配置后同键继续；供应商明确Failed不假成功。

创建前校验失败，若已解析合法请求编号则尝试CAS0写入rejected-v1拒绝记录。只有这种持久记录可公开safeToStartNew:true，确保旧同键请求之后也无法创建；页面核对后可显式“修改后重新上传”。404或存储故障不能证明安全，继续保留原键、原文件，只允许重试本次请求。未知创建仍禁止新建。此恢复规则是避免重复上传及死锁的改进，未冒称原版已有。

## 实施与验收

1. 独立Go协议客户端：原同输入自造响应与Go请求/结果对照，签名固定向量、空ID/失败/超限/取消/等待边界。
2. 接真实配置、MySQL操作记录、Redis租约和路由，完整校验先于外发；恢复与未知结果防重测。
3. 接前端上传/停止/核对/草稿应用，真实React DOM验证；新增字段在配置前后端白名单同步。
4. 独立审查、真实MySQL/Redis加替身供应商测试；实际4001只验无配置/输入/鉴权等不外发路径。所有供应商成功必须标“受控替身”，不冒充真实服务。原桌面/付费真上传/最终视频消费继续保留未完成。
