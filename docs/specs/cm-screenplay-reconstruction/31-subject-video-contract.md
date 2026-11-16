# 主体存档进入视频：两条原版链的独立实现

用户已持续授权完整复刻，本项承接上传29/30。配置使用自己的账号、模型和地址；不创建假模型选项，不消费真实视频额度。原证据在主案`subject-video-20261116/ANALYSIS.md`及sha256-manifest.json。

## 原版输入/输出

- 旧Sora：恢复前端299617-299630的E只处理镜头绑定角色和场景，不处理道具；301145/301441仅text_to_video且autoReplaceSora2Info开时传replaced_prompt。原Api.add_video_task:50036到原最终submit_task已隔离执行，非空replaced_prompt成为JSON.prompt，关闭保持原文。
- 现代Seedance：原_process_task:1502→原数据库读绑定实体559→bindings620→submit1022，经原自造数据库到假HTTP截获：use_subject_mode=true时content中追加`image_url.url=asset://ID, role=reference_image`，文字仍保留@姓名；关闭无主体引用。只选reference_mode并不足以开启。缺ID且无绑定图片在HTTP前失败。
- 这是原编译方法声明位置及受控执行，不是已取得完整作者源码/已点击原UI/供应商生成成功。其它通道不外推。

## 自家保存与页面

CreativeConfig新增`auto_replace_sora2_info`、`use_subject_mode`两个布尔默认false；同时开拒绝。另`video_generation_mode`为空时保留旧自动图片行为，可明确选text_to_video或image_to_video。已有配置不改。模型配置中分别展示“自动替换主体存档”“主体模式”和生成模式，说明适用协议；保存与重开真实持久。模型来自自有配置，不按原版模型名硬植选项。

自有视频模型可显式声明`provider=seedance_ark`（Ark content任务协议）或`provider=sora_json`（已观察JSON Sora create/query协议），采用条目真实modelId/endpoint/apiKey。不覆盖旧YFAI/H3/个人API/本地执行器协议，未知或不支持当前模式在外发前明确拒绝。新协议的轮询带上正确的配置条目ID，不退回固定模型；停用或删除配置后清晰报错。

## 绑定、边界和请求

1. 只读当前账号、当前作品的资产。优先明确绑定ID；现行分镜仅有名字时复用已修复的同类型完整名称规则，重名/缺失不猜测，主体模式报明对应资产。
2. 存档从Asset.Extra.sora2_info读取，不因旧volcengineAssetId字段存在就假称已绑定。必要JSON/URI/ID格式以原helper边界重放为准；不把任意JSON原文拼到asset://。项目安全上传快照只用于明确适配，手工清空不能从旧值复活。
3. Sora文字替换仅明确text_to_video时启用，镜头绑定角色/场景参与；按原@token边界，未知提及保留，道具不替换。不得在图生模式偷偷改文案。开启该策略却未选text_to_video应明确提示。
4. Seedance主体模式不替换文本；每个绑定实体优先有效主体URI，无ID时可用该实体当前已选图片公网引用；两者都缺则发送前失败。不能拿其它镜头图或部分重名角色充数。顺序稳定、重复URI去重，遵守供应商引用上限；关闭模式不改变旧图片链。
5. 显式text_to_video关闭旧自动图片来源，image_to_video需可用图片；空模式维持旧行为。主体引用是专用字段，不能把asset://混进所有旧adapter的普通图片数组，非适配协议不得忽略后继续收费。
6. 在任何镜头提交之前预检整集最终输入：缺主体、非法模式、配置/协议不匹配/引用上限均停止，保留旧任务与结果。错误不把“未发送”伪装“提交未知”；恢复不重复生成。
7. 指纹包含最终prompt、主体策略、模式、稳定ID/URI/引用顺序；签名URL用稳定媒体ID代替以免过期重签触发重建。换主体必须进入已有显式重建流程，重跑相同输入复用原任务。

## 技术接线与验收

分三块：项目配置/主体解析/runner；自有原生视频协议适配与轮询；页面两个开关及生成模式。各块先失败测试、再独立实现，统一接全路由检查和真实MySQL/受控HTTP生产执行。

必须比较同镜头开关A/B的最终HTTP正文，Seedance包含主体URI而保留文字，Sora仅文字替换；正常、空存档/坏JSON、重复、重名、取消、缺图、关闭/重开配置、提交未知及更换主体版本全覆盖。最终请求由HTTP替身截停，不推断真实供应商成功。实际4001只验配置/无外发路径和明确受控作业，不调用任何真实视频模型。全36组与原桌面/真实供应商仍未完成。
