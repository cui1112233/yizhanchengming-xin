# 使用我们自己模型的映射规则

本文件固定配置含义与现有代码位置，不把原版厂商名称直接变成我们的可选模型。当前实际可用型号清单尚未验收：2026-11-17对4001文字/图片/视频目录的只读请求均401，不能解释为三个目录为空。未修改认证或索取密钥，也未启动供应商测试。

## 对应关系

| 原版字段/用途 | CM字段/入口 | 规则与当前证据 |
|---|---|---|
| 主流程model | 文字请求Model及作品ModelID | 选本账号文字目录中的真实ID；主流程请求使用当前选择。独立资产重解析与主流程的原版取值规则不同，提交前需明确作品配置已保存 |
| prompt_content/id/name | 后端prompt模块中的选择ID和冻结正文 | 前端传ID，后端解析真实正文和版本；不能把模型选择当提示词选择，也不能在前端硬编码系统词 |
| image_gen_model | CreativeConfig.image_model_id | 从listAvailableModels('image')读取本账号可用模型；图片请求显式modelId优先，未传时按实际操作读取项目配置，不回退成原版不存在的厂商模型 |
| image_edit_model | CreativeConfig.image_edit_model_id | 单独保存编辑模型；需要参考图的操作按自身适配器能力处理，不能把列表里任意图片模型都视为支持编辑 |
| selected_style/style_desc/art_style | CreativeStyleSnapshot及后端style数据 | 保存ID、名称、提示文字、图片和版本。按模型能力把文字和实际参考图送入请求，不能只显示选中状态 |
| character_three_view | CharacterImageMode/设计提示词选择 | 对齐ZIP单人/三视图的实际prompt和比例含义，保留自己合法扩展；三视图16:9需所选渠道支持，不支持应提示 |
| voice_model | 音色描述所需文字模型（原主请求传的是文字选择） | 不能直接当配音供应商型号。原版实际音色设计路径还需核查 |
| voiceDesignModel / dubbingModel | voice_design_model及真正配音服务/音色ID | 配置、描述生成、音色绑定、配音四层分开；当前代码中有minimax/qwen-tts选项，不证明对应服务已配置或可用 |
| 视频模型及模式 | 后续媒体请求的本账号video模型ID | 仍待原媒体链路规格；不依据同名模型推断参考图、时长或分辨率都兼容 |

## 已检查的我们代码

frontend/src/user/pages/ComicScreenplayCreationFields.jsx:83–104动态读取图片目录，分别保存生成/编辑模型；已保存但目录中未找到的ID会显示“已保存模型”，因此仍需后端验证当前是否可用，不能把这个标签当可执行证明。

frontend/src/shared/api/modelCatalog.js:56通过GET /api/models?kind=...读取；backend/internal/screenplay/creative_config.go保存ImageModelID、ImageEditModelID和Style；backend/internal/httpapi/screenplay_media_workbench.go:627附近在图片请求中读取配置，缺模型返回请选择图片模型。资产处理、普通生图、带参考图操作的选模含义不能混用。

本批运行六项现有定向测试：模型配置字段保存、设计提示词和资产传入、图片引用/失败历史、资产处理的编辑模型、配音音色/重复提交、生成和编辑权限，全部通过。只证明这些本地受控代码路径，不代表所有账号渠道、真实供应商或当前4001已通过。

## 上线前还必须核实

实际登录会话下读取模型目录；验证禁用/删除/无权限型号；验证图片参考能力、比例和真实通道参数；确认音色与配音的可用服务；同一输入比较原版和CM最终模型消息。本次未做配置写入、未改产品、未重启4001。

只读快照PID26264、git_sha与manifest均9f28ee75；这是当时已运行版本，不是本批部署。目录401证据在analysis/cm-model-mapping-readonly-v2-20261117.json，不应列成选模验收成功。
