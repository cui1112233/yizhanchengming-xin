// 洛水「全局设置 → API 接口」字段清单。
// 已剔除洛水自家渠道（小鱼漫漫、洛水中转通用、洛水渠道 Grok/Seedance/生图包）。
// 骨架阶段只承载分组、顺序与文案；密钥读写下一步接服务端存储。
// 字段 key 一旦接入存储不要再改名，避免已保存的配置对不上。

export const API_INTEGRATION_GROUPS = [
  {
    id: 'text',
    title: '文本模型',
    tone: 'violet',
    description: '文本生成使用的官方密钥；OpenRouter 需要网络支持，国内官方密钥可直连。',
    fields: [
      { key: '__sub_openrouter', subheading: 'OpenRouter' },
      {
        key: 'openrouterApiKey',
        label: '',
        control: 'password',
        placeholder: '输入 OpenRouter API Key',
        help: false
      },
      { key: '__sub_official', subheading: '国内官方密钥' },
      {
        key: 'kimiOfficialApiKey',
        label: '月之暗面 kimi',
        control: 'password',
        placeholder: '输入月之暗面官方 API Key'
      },
      {
        key: 'deepseekOfficialApiKey',
        label: '深度探索 deepseek',
        control: 'password',
        placeholder: '输入 DeepSeek 官方 API Key'
      },
      {
        key: 'doubaoOfficialApiKey',
        label: '火山豆包',
        control: 'password',
        placeholder: '输入火山豆包官方 API Key'
      },
      {
        key: 'glmOfficialApiKey',
        label: '智谱 glm',
        control: 'password',
        placeholder: '输入智谱官方 API Key'
      },
      {
        key: 'qwenOfficialApiKey',
        label: '阿里 qwen',
        control: 'password',
        placeholder: '输入阿里通义官方 API Key'
      },
      {
        key: 'mimoOfficialApiKey',
        label: '小米 MiMo',
        control: 'password',
        placeholder: '输入小米 MiMo 官方 API Key'
      }
    ]
  },
  {
    id: 'video',
    title: '视频模型',
    tone: 'green',
    description: '视频生成与本地视频服务：Seedance、Vidu、RunningHub、阿里万相、本地 Grok、小云雀、MiniMax-H3。',
    fields: [
      {
        key: 'seedance2Api',
        label: 'Seedance2.0 官网API',
        control: 'password',
        placeholder: '输入Seedance2.0 官网 API key值',
        applyUrl: 'https://console.volcengine.com/ark/region:ark+cn-beijing/apikey?apikey=%7B%7D',
        help: true
      },
      {
        key: 'viduApi',
        label: 'Vidu API 官网Key',
        control: 'password',
        placeholder: '输入Vidu API 官网key值',
        help: true
      },
      {
        key: 'runninghubApi',
        label: 'RunningHub Key',
        control: 'password',
        placeholder: '输入 RunningHub API Key',
        applyUrl: 'https://www.runninghub.cn/call-api/api-detail/2032764885651525640',
        help: true
      },
      {
        key: 'runninghubConcurrency',
        label: 'RunningHub 并发数',
        control: 'number',
        placeholder: '默认 10',
        defaultValue: 10,
        min: 1,
        max: 100,
        rangeHint: '默认 10，范围 1-100。'
      },
      {
        key: 'wanxApiKey',
        label: '阿里万相模型 Key',
        control: 'password',
        placeholder: '输入阿里万相模型 API密钥（如需要）',
        applyUrl: 'https://bailian.console.aliyun.com/cn-beijing/?spm=5176.29597918.nav-v2-dropdown-menu-0.d_main_1_0_0.58107b08oinsbv&tab=model&scm=20140722.M_10929481._.V_1#/model-market/all'
      },
      {
        key: 'localGrokVideoBaseUrl',
        label: '本地Grok视频服务地址',
        control: 'text',
        placeholder: '默认 http://localhost:5400'
      },
      {
        key: 'localGrokVideoApiKey',
        label: '本地Grok视频服务 API Key',
        control: 'password',
        placeholder: '输入本地Grok服务 API key，未设置则不发送 Authorization'
      },
      {
        key: 'xiaoyunqueAccessKey',
        label: '小云雀 Access Key',
        control: 'password',
        placeholder: '输入小云雀 Access Key'
      },
      {
        key: 'xiaoyunqueConcurrency',
        label: '小云雀并发数量',
        control: 'number',
        placeholder: '默认 1',
        defaultValue: 1,
        min: 1,
        max: 20,
        rangeHint: '默认 1，范围 1-20。'
      },
      {
        key: 'xiaoyunqueIntlAccessKey',
        label: '小云雀-国际版 Access Key',
        control: 'password',
        placeholder: '输入 Pippit Access Key（pippit.ai）'
      },
      {
        key: 'xiaoyunqueIntlConcurrency',
        label: '小云雀-国际版并发数量',
        control: 'number',
        placeholder: '默认 1',
        defaultValue: 1,
        min: 1,
        max: 20,
        rangeHint: '默认 1，范围 1-20。国际版直连 https://www.pippit.ai'
      },
      {
        key: 'happyHorseKey',
        label: 'HappyHorse Key',
        control: 'password',
        placeholder: '输入 HappyHorse API Key（阿里云）',
        applyUrl: 'https://bailian.console.aliyun.com/cn-beijing/?spm=5176.29597918.nav-v2-dropdown-menu-0.d_main_1_0_0.58107b08oinsbv&tab=model&scm=20140722.M_10929481._.V_1#/model-market/all'
      },
      {
        key: 'minimaxH3Key',
        label: 'MiniMax-H3 Key',
        control: 'password',
        placeholder: '输入 MiniMax-H3 API Key',
        applyUrl: 'https://platform.minimaxi.com/user-center/basic-information/interface-key'
      }
    ]
  },
  {
    id: 'image',
    title: '图片模型',
    tone: 'purple',
    description: '本地出图服务与云端出图密钥；本地服务的 Key 未设置时不发送 Authorization。',
    fields: [
      {
        key: 'localGrokImageBaseUrl',
        label: '本地Grok图片服务地址',
        control: 'text',
        placeholder: '默认 http://localhost:8000'
      },
      {
        key: 'localGrokImageApiKey',
        label: '本地Grok图片服务 API Key',
        control: 'password',
        placeholder: '输入本地Grok图片服务 API key，未设置则不发送 Authorization'
      },
      {
        key: 'localBananaBaseUrl',
        label: '本地大香蕉服务地址',
        control: 'text',
        placeholder: '默认 http://127.0.0.1:38000/'
      },
      {
        key: 'localBananaApiKey',
        label: '本地大香蕉服务 API Key',
        control: 'password',
        placeholder: '输入本地大香蕉服务 API key，未设置则不发送 Authorization'
      },
      {
        key: 'qwenImageApi',
        label: 'Qwen-Image API',
        control: 'password',
        placeholder: '输入Qwen-Image API密钥',
        help: true
      },
      {
        key: 'zImageApi',
        label: '阿里云 Z-Image API',
        control: 'password',
        placeholder: '输入阿里云 Z-Image API密钥',
        help: true
      },
      {
        key: 'volcanoSeedreamApi',
        label: '火山 Seedream API',
        control: 'password',
        placeholder: '输入火山方舟 Seedream API Key',
        applyUrl: 'https://console.volcengine.com/ark/region:ark+cn-beijing/apikey?apikey=%7B%7D',
        help: true
      },
      {
        key: 'nano2BaseUrl',
        label: 'Nano2(自定义渠道) 基础链接',
        control: 'text',
        placeholder: '输入自定义渠道基础链接，如 https://api.example.com',
        help: true
      },
      {
        key: 'nano2ApiKey',
        label: 'Nano2(自定义渠道) API Key',
        control: 'password',
        placeholder: '输入自定义渠道API Key',
        help: true
      },
      {
        key: 'grsaiImage2Key',
        label: 'grsai-image2-key',
        control: 'password',
        placeholder: '输入 GRSAI 出图 API Key（gpt-image-2 / 2.5 / sunburst 共用）',
        applyUrl: 'https://grsai.ai/zh/dashboard/api-keys',
        help: true
      },
      {
        key: 'geeknowImageApi',
        label: 'geeknow生图api',
        control: 'password',
        placeholder: '输入 Geeknow 生图 API Key（gpt-image-2 / pro / vip / 2.5 共用）',
        applyUrl: 'https://www.geeknow.top',
        help: true
      }
    ]
  },
  {
    id: 'jimeng',
    title: '即梦浏览器模式',
    tone: 'pink',
    description: '开启浏览器模式后使用官网浏览器流程执行任务，不走即梦账号验证码登录。',
    channelNote: '本组是浏览器模式开关、并发与超时配置，没有可切换的多个渠道，所以保留字段形态，不套渠道卡。',
    warning: '浏览器模式为固定开启，任务由官网页面流程执行；请遵守官网使用条款，任务失败按官网提示处理。',
    fields: [
      {
        key: 'jimengBrowserEnabled',
        label: '开启浏览器',
        control: 'switch',
        locked: true
      },
      {
        key: 'jimengConcurrency',
        label: 'JM账号并发数',
        control: 'number',
        placeholder: '默认 10',
        defaultValue: 10,
        min: 1,
        max: 10,
        rangeHint: '默认 10，范围 1-10。'
      },
      {
        key: 'jimengTimeoutHours',
        label: 'JM任务超时（小时）',
        control: 'number',
        placeholder: '默认 10',
        defaultValue: 10,
        min: 1,
        max: 72,
        rangeHint: '默认 10 小时，范围 1-72。'
      }
    ]
  },
  {
    id: 'realperson',
    title: '真人素材库',
    tone: 'orange',
    description: '用于官网火山引擎、真人库、Geeknow、古门、得云AI、ydy-素材或 Xiaoxiao 素材组上传。',
    channelNote: '本组是素材来源与上传参数：来源由「真人素材渠道」下拉决定，凭证按来源共用，因此保留字段形态，不套渠道卡。',
    note: '同一作品共用一个素材组；上传后返回 Id，视频主体参考以 asset://Id 提交。',
    fields: [
      {
        key: 'realpersonChannel',
        label: '真人库渠道',
        control: 'select',
        placeholder: '选择真人库渠道',
        options: [
          { value: 'volcengine', label: '火山引擎官网' },
          { value: 'deyunai', label: '得云AI' },
          { value: 'gumen', label: '古门' },
          { value: 'xiaoxiao', label: 'Xiaoxiao' },
          { value: 'ydy', label: 'ydy-素材' }
        ]
      },
      { key: '__sub_deyunai', subheading: '得云AI素材' },
      {
        key: 'deyunaiRegion',
        label: '得云AI素材区域',
        control: 'select',
        placeholder: '选择素材区域',
        options: [
          { value: 'c39', label: '国内 c39' },
          { value: 'c40', label: '海外 c40' }
        ]
      },
      { key: '__sub_ydy', subheading: 'ydy-素材' },
      {
        key: 'ydyAccessKey',
        label: 'AccessKey',
        control: 'password',
        placeholder: '控制台「我的凭证-访问密钥」中的 Access Key'
      },
      {
        key: 'ydySecretKey',
        label: 'SecretKey',
        control: 'password',
        placeholder: '控制台「我的凭证-访问密钥」中的 Secret Key'
      },
      {
        key: 'ydyPoolId',
        label: 'Pool-Id（资源池）',
        control: 'text',
        placeholder: 'CIDC-CORE-00'
      },
      { key: '__sub_gumen', subheading: '古门' },
      {
        key: 'gumenApiToken',
        label: '古门 API Token',
        control: 'password',
        placeholder: '请联系古门客服开通后填写 token'
      },
      {
        key: 'gumenGroupId',
        label: '古门资源分组 ID（可选）',
        control: 'text',
        placeholder: '留空则使用默认分组'
      },
      { key: '__sub_xiaoxiao', subheading: 'Xiaoxiao' },
      {
        key: 'xiaoxiaoModel',
        label: 'Xiaoxiao 素材模型',
        control: 'select',
        placeholder: '选择素材模型',
        options: [
          { value: 'doubao-seedance-2-0-260128', label: 'Seedance 2.0（doubao-seedance-2-0-260128）' }
        ]
      },
      { key: '__sub_volcengine', subheading: '火山引擎官网' },
      {
        key: 'volcanoProjectName',
        label: 'ProjectName',
        control: 'text',
        placeholder: '默认 default，可填写 zhouen 等项目名'
      },
      {
        key: 'volcanoAccessKeyId',
        label: 'Access Key ID',
        control: 'password',
        placeholder: '请输入火山引擎 Access Key ID'
      },
      {
        key: 'volcanoSecretAccessKey',
        label: 'Secret Access Key',
        control: 'password',
        placeholder: '请输入火山引擎 Secret Access Key'
      }
    ]
  },
  {
    id: 'vidu',
    title: 'Vidu 账号',
    tone: 'cyan',
    description: '开启浏览器模式后直接使用官网浏览器流程，不走 Vidu 账号验证码登录。',
    channelNote: '本组是 Vidu 账号位与浏览器模式配置：只有一个账号位，没有多渠道可选，因此保留字段形态。',
    kind: 'vidu-accounts',
    fields: [
      { key: 'viduBrowserEnabled', label: '开启浏览器', control: 'switch' }
    ]
  },
  {
    id: 'subtitle',
    title: '字幕消除',
    tone: 'amber',
    description: '用于视频智能全消 / 字幕擦除；火山 Key 同时用于分镜超分，550wai 使用独立的用户编号和 API Key。',
    channelNote: '本组是字幕消除服务商与其凭证：服务商由下拉决定，字段按服务商共用，因此保留字段形态，不套渠道卡。',
    fields: [
      {
        key: 'subtitleProvider',
        label: '服务商',
        control: 'select',
        placeholder: '选择字幕消除服务',
        options: [
          { value: 'volcano', label: '火山（超分 / 去字幕共用）' },
          { value: '550wai', label: '550wai（极速去字幕）' },
          { value: 'kaipai', label: '开拍' },
          { value: 'niuma', label: '牛马（暂未接入）' }
        ]
      },
      {
        key: 'volcanoEraseVersion',
        label: '火山擦除版本',
        control: 'select',
        placeholder: '选择火山擦除版本',
        options: [
          { value: 'standard', label: '标准版 erase-video-subtitle' },
          { value: 'pro-v5', label: '精细化擦除 v5（默认）' }
        ]
      },
      {
        key: 'subtitleConcurrency',
        label: '任务并发数',
        control: 'number',
        placeholder: '默认 20',
        defaultValue: 20,
        min: 1,
        max: 50,
        rangeHint: '默认 20，范围 1-50。'
      },
      { key: '__sub_kaipai', subheading: '开拍' },
      {
        key: 'kaipaiAccessKey',
        label: '开拍 Access Key（MT_AK）',
        control: 'password',
        placeholder: '输入开拍 Access Key'
      },
      {
        key: 'kaipaiSecretKey',
        label: '开拍 Secret Key（MT_SK）',
        control: 'password',
        placeholder: '输入开拍 Secret Key'
      },
      { key: '__sub_550wai', subheading: '550wai' },
      {
        key: 'w550UserNo',
        label: '550wai 用户编号（userNo）',
        control: 'text',
        placeholder: '输入 550wai 用户编号'
      },
      {
        key: 'w550ApiKey',
        label: '550wai API Key',
        control: 'password',
        placeholder: 'sk-550w-... 或 sk-550w-sub-v1_...'
      },
      { key: '__sub_volcano', subheading: '火山 AI MediaKit' },
      {
        key: 'volcanoMediaKitApiKey',
        label: '火山视频超分/去字幕 API Key',
        control: 'password',
        placeholder: '输入火山 AI MediaKit API Key（超分与去字幕共用）',
        applyUrl: 'https://console.volcengine.com/imp/ai-mediakit/settings'
      }
    ]
  },
  {
    id: 'storage',
    title: '海螺 / NIJI7 / 图床',
    tone: 'blue',
    description: '海螺服务地址、NIJI7 出图与图床接口；图床返回的公网 URL 会用于第三方视频接口的参考图。',
    channelNote: '本组是海螺 / NIJI7 / 图床三类固定集成项，各自地址与密钥独立填写，没有“启用哪几个渠道”的概念，因此保留字段形态。',
    fields: [
      {
        key: 'conchApiUrl',
        label: '海螺API',
        control: 'password',
        placeholder: '输入海螺API地址'
      },
      {
        key: 'niji7KongyangKey',
        label: 'NIJI7 空氧key',
        control: 'password',
        placeholder: '输入NIJI7空氧key',
        help: true
      },
      {
        key: 'niji7Concurrency',
        label: 'NIJI7 并发数',
        control: 'number',
        min: 1,
        max: 500,
        rangeHint: '范围 1-500。'
      },
      {
        key: 'niji7Style',
        label: 'NIJI7 全局风格',
        control: 'select',
        placeholder: '选择风格',
        options: []
      },
      {
        key: 'imageHostUrl',
        label: '图床接口',
        control: 'password',
        placeholder: '输入图床接口地址',
        help: true
      },
      { key: 'imageHostApiKey', label: '图床 API Key（可选）', control: 'password', placeholder: '图床要求 Bearer 时填写' }
    ]
  },
  {
    id: 'tts',
    title: '配音设置',
    tone: 'teal',
    description: '配音服务地址与各语音供应商密钥；保存后新的配音任务立即使用新配置。',
    channelNote: '本组是配音服务与语音供应商密钥，属于固定集成项，不支持多渠道切换，因此保留字段形态。',
    fields: [
      {
        key: 'ttsServiceUrl',
        label: 'TTS服务地址',
        control: 'text',
        placeholder: '输入TTS服务地址'
      },
      {
        key: 'ttsApiKey',
        label: 'TTS API密钥',
        control: 'password',
        placeholder: '输入TTS API密钥（如需要）'
      },
      {
        key: 'qwenTtsKey',
        label: 'Qwen-tts密钥',
        control: 'password',
        placeholder: '输入Qwen-tts API密钥（如需要）',
        applyUrl: 'https://bailian.console.aliyun.com/cn-beijing/?spm=5176.29597918.nav-v2-dropdown-menu-0.d_main_1_0_0.58107b08oinsbv&tab=model&scm=20140722.M_10929481._.V_1#/model-market/all'
      },
      {
        key: 'cosyVoiceKey',
        label: 'CosyVoice 3.5密钥',
        control: 'password',
        placeholder: '输入CosyVoice 3.5 API密钥（如需要）',
        applyUrl: 'https://bailian.console.aliyun.com/cn-beijing/?spm=5176.29597918.nav-v2-dropdown-menu-0.d_main_1_0_0.58107b08oinsbv&tab=model&scm=20140722.M_10929481._.V_1#/model-market/all'
      },
      {
        key: 'minimaxApiKey',
        label: 'Minimax API密钥',
        control: 'password',
        placeholder: '输入Minimax API密钥',
        applyUrl: 'https://platform.minimaxi.com/'
      },
      { key: '__sub_tts', subheading: '配音设置' },
      {
        key: 'ttsSentenceLength',
        label: '单句配音长度',
        control: 'number',
        placeholder: '默认50',
        defaultValue: 50,
        min: 1,
        max: 1000,
        rangeHint: '默认 50，范围 1-1000。'
      }
    ]
  }
];

// 渠道化配置（先在「文本模型」做样例）：
// 预置渠道已经把 baseUrl 与模型名填好，用户只需要填 Key。
// 三组里预置渠道与自定义渠道常驻展示：填了 Key 就生效并进入模型下拉，
// 清空 Key 即从下拉移除。「自定义」弹窗只负责新增/删除自定义渠道，不再保存勾选结果。
export const API_CHANNEL_PRESETS = {
  text: {
    presets: [
      {
        id: 'deepseek',
        label: '深度探索 DeepSeek',
        fieldKey: 'deepseekOfficialApiKey',
        baseUrl: 'https://api.deepseek.com/v1',
        model: 'deepseek-chat',
        common: true
      },
      {
        id: 'doubao',
        label: '火山豆包',
        fieldKey: 'doubaoOfficialApiKey',
        baseUrl: 'https://ark.cn-beijing.volces.com/api/v3',
        model: 'doubao-seed-1-6',
        common: true
      },
      {
        id: 'qwen',
        label: '阿里通义千问',
        fieldKey: 'qwenOfficialApiKey',
        baseUrl: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
        model: 'qwen-max',
        common: true
      },
      {
        id: 'openrouter',
        label: 'OpenRouter',
        fieldKey: 'openrouterApiKey',
        baseUrl: 'https://openrouter.ai/api/v1',
        model: '自动路由',
        common: false
      },
      {
        id: 'kimi',
        label: '月之暗面 Kimi',
        fieldKey: 'kimiOfficialApiKey',
        baseUrl: 'https://api.moonshot.cn/v1',
        model: 'kimi-k2',
        common: false
      },
      {
        id: 'glm',
        label: '智谱 GLM',
        fieldKey: 'glmOfficialApiKey',
        baseUrl: 'https://open.bigmodel.cn/api/paas/v4',
        model: 'glm-4.5',
        common: false
      },
      {
        id: 'mimo',
        label: '小米 MiMo',
        fieldKey: 'mimoOfficialApiKey',
        baseUrl: 'https://api.xiaomimimo.com/v1',
        model: 'MiMo 默认',
        common: false,
        pending: true
      }
    ]
  },
  video: {
    presets: [
      {
        id: 'seedance2',
        label: 'Seedance2.0',
        fieldKey: 'seedance2Api',
        addressField: 'seedance2BaseUrl',
        baseUrl: 'https://ark.cn-beijing.volces.com/api/v3',
        model: 'doubao-seedance-2-0',
        common: true
      },
      {
        id: 'minimaxH3',
        label: 'MiniMax-H3',
        fieldKey: 'minimaxH3Key',
        baseUrl: 'https://api.minimaxi.com/v1',
        model: 'MiniMax-H3',
        common: true,
        pending: true
      },
      {
        id: 'runninghub',
        label: 'RunningHub',
        fieldKey: 'runninghubApi',
        baseUrl: 'https://www.runninghub.cn',
        model: '工作流任务',
        common: true,
        extraFields: ['runninghubConcurrency']
      },
      {
        id: 'viduVideo',
        label: 'Vidu',
        fieldKey: 'viduApi',
        baseUrl: 'https://api.vidu.cn',
        model: 'Vidu 默认',
        common: false,
        pending: true
      },
      {
        id: 'wanx',
        label: '阿里万相',
        fieldKey: 'wanxApiKey',
        baseUrl: 'https://dashscope.aliyuncs.com/api/v1',
        model: '万相视频',
        common: false,
        pending: true
      },
      {
        id: 'xiaoyunque',
        label: '小云雀',
        fieldKey: 'xiaoyunqueAccessKey',
        baseUrl: '',
        model: '小云雀默认',
        common: false,
        pending: true,
        extraFields: ['xiaoyunqueConcurrency']
      },
      {
        id: 'xiaoyunqueIntl',
        label: '小云雀-国际版',
        fieldKey: 'xiaoyunqueIntlAccessKey',
        baseUrl: 'https://www.pippit.ai',
        model: 'Pippit 默认',
        common: false,
        pending: true,
        extraFields: ['xiaoyunqueIntlConcurrency']
      },
      {
        id: 'localGrokVideo',
        label: '本地 Grok 视频',
        fieldKey: 'localGrokVideoApiKey',
        addressField: 'localGrokVideoBaseUrl',
        baseUrl: 'http://localhost:5400',
        model: '本地服务',
        common: false
      },
      {
        id: 'happyhorse',
        label: 'HappyHorse',
        fieldKey: 'happyHorseKey',
        baseUrl: 'https://dashscope.aliyuncs.com/api/v1',
        model: 'HappyHorse 默认',
        common: false,
        pending: true
      }
    ]
  },
  image: {
    presets: [
      {
        id: 'volcanoSeedream',
        label: '火山 Seedream',
        fieldKey: 'volcanoSeedreamApi',
        baseUrl: 'https://ark.cn-beijing.volces.com/api/v3',
        model: 'doubao-seedream-3-0-t2i',
        common: true
      },
      {
        id: 'grsai',
        label: 'grsai',
        fieldKey: 'grsaiImage2Key',
        baseUrl: 'https://grsai.ai',
        model: 'gpt-image-2 / 2.5 / sunburst',
        common: true,
        pending: true
      },
      {
        id: 'geeknow',
        label: 'geeknow',
        fieldKey: 'geeknowImageApi',
        baseUrl: 'https://www.geeknow.top',
        model: 'gpt-image-2 / pro / vip / 2.5',
        common: true,
        pending: true
      },
      {
        id: 'qwenImage',
        label: 'Qwen-Image',
        fieldKey: 'qwenImageApi',
        baseUrl: 'https://dashscope.aliyuncs.com/api/v1',
        model: 'qwen-image',
        common: false,
        pending: true
      },
      {
        id: 'zImage',
        label: '阿里云 Z-Image',
        fieldKey: 'zImageApi',
        baseUrl: 'https://dashscope.aliyuncs.com/api/v1',
        model: 'z-image',
        common: false,
        pending: true
      },
      {
        id: 'nano2',
        label: 'Nano2（自定义渠道）',
        fieldKey: 'nano2ApiKey',
        addressField: 'nano2BaseUrl',
        baseUrl: '',
        model: '自定义渠道',
        common: false
      },
      {
        id: 'localGrokImage',
        label: '本地 Grok 图片',
        fieldKey: 'localGrokImageApiKey',
        addressField: 'localGrokImageBaseUrl',
        baseUrl: 'http://localhost:8000',
        model: '本地服务',
        common: false
      },
      {
        id: 'localBanana',
        label: '本地大香蕉',
        fieldKey: 'localBananaApiKey',
        addressField: 'localBananaBaseUrl',
        baseUrl: 'http://127.0.0.1:38000/',
        model: '本地服务',
        common: false
      }
    ]
  }
};

export function apiIntegrationFieldKeys() {
  return API_INTEGRATION_GROUPS.flatMap(group => group.fields.filter(field => !field.subheading).map(field => field.key));
}
