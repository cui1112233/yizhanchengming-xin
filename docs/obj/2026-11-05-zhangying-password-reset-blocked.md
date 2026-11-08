# zhangying 密码重置执行报告（阻断，未执行）

- 日期：2026-11-05 19:06 CST
- 目标：`yzcm.site` 主站 `accounts` 表中 `username='zhangying'` 的密码重置为授权的新密码
- 结果：**未执行，未改任何数据，未重置密码**

## 1. 结论

**改成功没有：没有。** 本次在第一步（建立 SSH 连接）就被拒绝，按硬要求第 6 条立即停止：

```text
ssh root@115.190.156.223
→ Permission denied (publickey,password)
→ exit code 255
```

没有可用的 SSH 通道，因此：

- 未创建 `/root/backup-20261105/account-zhangying-before-reset-<时间戳>.sql`
- 未产生备份 sha256
- 未执行 `UPDATE`，`affected rows` 不存在（0 次写操作）
- 未回读 `active/password_hash/updated_at`
- 未用新密码或错误密码调用线上登录接口做验证
- 未重启服务、未改代码、未 commit、未动其它账号或旧库

## 2. 为什么停止

硬要求第 6 条明确：SSH 仍被拒或 DB 写权限不足时立即停止，如实回报，不伪造结果，不改用其它非授权通道。

本次只做了一次只读连通性检查（`id -u; hostname`），未尝试密码登录、未尝试任何替代通道，也未用本地数据库冒充线上执行。之前排查阶段已确认本机可用密钥均被拒（`Permission denied (publickey,password)`），本次复核结果一致。

## 3. 待恢复条件

要继续执行，需要先满足任一条件：

1. 为本机提供可登录 `115.190.156.223` 的 SSH 凭据（root 或具备 sudo/写库权限的账号 + 对应私钥）；或
2. 由授权方提供线上 MySQL 的可写通道，并明确允许在 `/root/backup-20261105/` 落备份。

拿到通道后，按原要求顺序执行：

1. 导出 `zhangying` 整行到 `/root/backup-20261105/account-zhangying-before-reset-<时间戳>.sql`，计算 sha256；
2. 用仓库 `backend/internal/auth/password.go` 的 scrypt-v1 同一实现生成新哈希，只打印前缀 `scrypt$` 与长度 104；
3. `UPDATE accounts SET password_hash=?, updated_at=NOW(6) WHERE username='zhangying' LIMIT 1`，确认 affected rows=1；
4. 回读 `active=1`、前缀、长度、`updated_at`；
5. 用新密码和错误密码各调一次 `/api/login`，记录状态码与耗时，会话用完即弃。

## 4. 当前线上状态

未发生任何变更，`zhangying` 仍使用 2026-10-16 12:51:02 写入的原哈希；线上服务未重启。此前只读排查结论仍然成立（账号存在、启用、哈希格式可解析，失败点为原密码不匹配）。

## 5. 敏感信息

本次未读取、未生成、未输出任何完整哈希、盐、令牌或其它凭据；授权的新密码未被写入任何文件或日志。
