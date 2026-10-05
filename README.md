# 一战晟铭新盘

这是新主线工程仓库。

- 后端：Go
- 用户端：React + Ant Design
- 管理端：React + Ant Design
- 数据库：MySQL + Goose
- Redis：任务 / 队列 / 锁
- 对象存储：TOS
- 部署：Go embed + 单二进制

## Phase 1 当前范围

当前第一阶段已经落地：

1. 小说 Intake API
2. TXT / Excel 上传
3. 121 文本抓取客户端与 Go provider
4. 元数据识别
5. 分类 / 类型 / 男女频 / 风格基础识别
6. BatchProject 创建
7. Run 创建
8. MySQL + Goose
9. React Novel Fetch 页面（输入书城 → 添加书城 → 标签 → 立即执行 / 自动化 → 新建批量）
10. 前后端生产镜像与 `docker compose`
11. GitHub Actions CI

## 本地启动

```bash
docker compose up --build
```

- 用户端 / API：`http://127.0.0.1:8080`
- 管理端：`http://127.0.0.1:8081`
- MySQL：`127.0.0.1:3306`
- Redis：`127.0.0.1:6379`
