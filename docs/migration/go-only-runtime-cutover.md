# Go-only 运行态切换记录

## 当前状态

生产活动拓扑使用：

```text
docker-compose.yml
+ docker-compose.direct.yml
+ docker-compose.go-only.yml
```

活动服务只有 Go API、Nginx 和 MySQL。`v88-node`、浏览器 Worker、旧
`shuihuo-compat` 不再作为运行时依赖；旧容器保留为停止状态，便于在真实
目标站上传读回验收完成前快速回滚。

## 验证要求

切换后必须检查：

- `docker compose ... config -q` 成功；
- Go API 健康检查为 healthy；
- `/`、`/batch-factory`、`/novel-fetch`、`/script`、`/shuihuo-production`
  均返回 HTTP 200；
- `docker ps` 中没有运行中的 `v88-node`、`browser-worker` 或
  `shuihuo-compat`。

## 回滚边界

只有发生 Go API 回归时，才临时恢复旧服务；恢复后必须记录原因并回写到
`v88`。目标站真实登录、测试书上传及读回仍需使用
`go-connector-live-smoke.md` 完成，完成后才允许物理删除旧镜像、容器和
回滚配置。
