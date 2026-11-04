# S1 只读 dry-run：漫画历史数据清单

- 日期：2026-11-04
- 范围：本地 `qiantie_go_local` 真实库，**只读查询**；本轮不写库、不 purge、不删数据。
- 结论：本轮 S1（议题 B + F）只新增读取字段与能力探测，未修改任何历史行。

## 1. 汇总

| 指标 | 数值 | 说明 |
| --- | --- | --- |
| `comic_pages` 总行数 | 1430 | 全表计数 |
| 含 `【最终画风锁定】` 的页 | 1428 | 历史上被重复/追加编译的页 |
| 锁段出现 ≥2 次的页 | 0 | 实测与审计所述“≥3 次”不一致，按实测记录 |
| `failed` 且有图的行 | 2 | 审计口径为 20，按实测记录，待人工确认 |
| `is_primary=1` 的图片行 | 1402 | “有图”统一按主图口径 |
| “有图但无主图”的页 | 0 | 当前库无此存量；代码已按只打标处理，不补主图 |

## 2. 按批次分布（含锁段的页）

| batch_id | 页数 |
| --- | --- |
| batch_b3ad67894483b23df9e5fb19cb526685 | 655 |
| batch_d41981a647691369b1f455e2a0b6e44f | 368 |
| batch_d7bffda3c94ec2813ef7cbb556d8c032 | 325 |
| batch_2f462bb6608219f96838d30435620386 | 43 |
| batch_1daf793822f2eacb0a5464b78cde0f39 | 34 |
| batch_d14752321b6417f69e238cce0196cc64 | 3 |

## 3. 逐页清单

逐页明细已落盘：[2026-11-04-s1-comic-dryrun-pages.csv](file:///f:/脚本测试/一战晟铭-v88/docs/verification/2026-11-04-s1-comic-dryrun-pages.csv)

CSV 列：`owner_username,batch_id,book_id,id,video_id,status,prompt_chars,lock_count`（1428 行，含表头）。
`prompt_chars` 是当前 `image_prompt` 字符长度，`lock_count` 是 `【最终画风锁定】` 出现次数。

## 4. 未做项与差异说明

- 本文件不包含“剥离后逐页 before/after 字节差”。该差异需要 S5/E 的 `comic-repair --dry-run` 工具用
  `StripComicImagePrompt` 逐页计算，本轮只做只读盘点，避免引入未验证的写路径。
- 237 条 patch-only 回填、1428 页去重、10 本重复合集清理均未执行；按用户拍板顺序在后续阶段进行。
