# 视频管理系统：Go 连接器与旧 Worker 对照

这份表是 Node/Playwright Worker 退役前的唯一对照清单。正常运行链路统一走 Go；旧 Worker 只允许作为明确开启的应急回滚件存在。

| 旧 Worker 动作 | Go 对等能力 | 证据 | 状态 |
| --- | --- | --- | --- |
| `session/login` | `POST /v1/sessions/login` | `TargetHTTPService.Login`、连接器服务测试 | 已迁移 |
| `session/test` | `POST /v1/sessions/test` | `TargetHTTPService.CheckSession`、连接器服务测试 | 已迁移 |
| `session/refresh` | `POST /v1/sessions/refresh` | 复用 Go 登录路径，避免第二套刷新逻辑 | 已迁移 |
| `session` 删除 | `DELETE /v1/sessions` | 加密会话存储 `Remove`、连接器服务测试 | 已迁移 |
| `dashboard` | `CheckSession` | `TargetHTTPService.CheckSession` | 已迁移 |
| `config_list` | `GET /v1/profiles` | `RefreshProfiles` | 已迁移 |
| `organization_list` | `POST /v1/catalogs/refresh` | `RefreshCatalogs` | 已迁移 |
| `book_list` | `GET /v1/books/{id}` | `GetBook` | 已迁移 |
| `upload` | `POST /v1/submissions` | `Submit`、批量工厂发布器 | 已迁移 |
| `asset_presign` | `UploadVideo` | `VideoUploadService` | 已迁移 |
| 提交回读 | `GET /v1/submissions/{id}` | `GetSubmission` | 已迁移 |

## 退役门槛

1. Go 服务已部署并通过精确 Git SHA 校验。
2. 目标站真实会话完成一次 Go 登录、会话验证、内容上传、视频资源上传和提交回读。
3. 运行态停止 `browser-worker` 后，小说获取、剧本获取、批量工厂页面和上传接口仍通过真实验收。
4. 满足以上条件后，才删除 Worker 服务、Node 客户端和对应配置；历史回滚资料另行归档，不再作为默认入口。

## 当前结论

Go 主服务已满足第 1 项，页面/健康检查已满足第 3 项的基础部分；第 2 项的真实目标站上传回读证据仍待完成，因此暂不删除 Worker。
