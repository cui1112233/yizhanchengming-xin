# 魔剪 4040 独立工作台

这是魔剪合并进“一战景铭”前的独立开发区。它只复用 4001 的登录身份，工程、素材、时间线、任务和导出数据全部保持隔离。

## 本地运行

开发构建使用 Node 22 LTS 和 pnpm；运行已构建的 Go 单文件时不需要 Node。

```powershell
.\scripts\start-dev.ps1
```

默认地址是 `http://127.0.0.1:4040`，健康检查是 `http://127.0.0.1:4040/health`。

## 隔离规则

- MySQL 数据库：`qiantie_magic_cut_dev`
- Redis 前缀：`magic_cut:dev:`
- TOS 前缀：`magic-cut/dev/<owner>/`
- 系统提示词：只能放在 `api/internal/prompt/`，前端不存提示词和密钥
