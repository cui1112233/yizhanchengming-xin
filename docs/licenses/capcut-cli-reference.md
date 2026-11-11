# capcut-cli 参考说明

- 上游项目：`renezander030/capcut-cli`
- 上游地址：https://github.com/renezander030/capcut-cli
- 许可证：MIT，Copyright (c) 2026 Rene Zander
- 本项目用途：仅参考公开的命令契约、时间线安全边界，以及 `detect-scenes` / `detect-silence` 的可验证行为。
- 实现方式：一战晟铭继续使用自己的 Go 后端、项目数据结构、对象存储和权限体系；线上运行不依赖 Node，也不直接读写用户电脑上的剪映目录。
- 未复用内容：洛水私有代码、提示词、密钥、私有接口和专有资源均未复制。

MIT 许可要求在复制或实质性复用上游源码时保留上游版权与许可文本。当前实现为独立 Go 实现；若后续引入上游源码文件，发布包必须同时附带上游完整 `LICENSE`。
