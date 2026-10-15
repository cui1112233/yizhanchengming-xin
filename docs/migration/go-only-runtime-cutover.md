# Go-only 运行态切换记录

## 当前状态

生产应用使用 `deploy/v88-direct` 的 Go 二进制直部署。活动应用只有
`qiantie.service` 启动的 Go API 和主机 Nginx；MySQL、Redis、TOS、FFmpeg
作为外部基础设施存在。`v88-node`、浏览器 Worker、旧 `shuihuo-compat`
不再作为默认应用运行时。历史容器配置仅保留作审计和回滚参考，不能作为
新的发布入口。

## 验证要求

切换后必须检查：

- `systemctl restart qiantie.service` 成功；
- Go API `/health` 健康检查为 200；
- `/api/runtime-build-info` 返回本次精确 Git SHA；
- `/`、`/batch-factory`、`/novel-fetch`、`/script`、`/shuihuo-production`
  均返回 HTTP 200；
- 运行状态中没有应用 Node 进程或第二个 Go 应用实例。

## 回滚边界

只有发生 Go API 回归时，才临时恢复旧服务；恢复后必须记录原因并回写到
`v88`。目标站真实登录、测试书上传及读回仍需使用
`go-connector-live-smoke.md` 完成，完成后才允许物理删除旧镜像、容器和
回滚配置。
