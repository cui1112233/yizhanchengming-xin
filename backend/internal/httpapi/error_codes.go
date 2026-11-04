package httpapi

import "errors"

// 四类“缺配置”错误的稳定错误码。前端按码显示引导文案，后端不再各自拼接
// 一套说法；code 是新增字段，原有 error 文案保持不变以兼容旧调用方。
const (
	ErrorCodeImageModelNotConfigured    = "IMAGE_MODEL_NOT_CONFIGURED"
	ErrorCodeVideoProviderNotConfigured = "VIDEO_PROVIDER_NOT_CONFIGURED"
	ErrorCodeExecutorNotReady           = "EXECUTOR_NOT_READY"
	ErrorCodeGiantPluginNotConfigured   = "GIANT_PLUGIN_NOT_CONFIGURED"

	// 视频管理系统（121 后台）凭证状态。状态接口与同步接口共用同一套码，
	// 前端据此区分“没登录 / 凭证不全 / 已过期 / 连接不可用”。
	ErrorCodeTargetNotLoggedIn          = "TARGET_NOT_LOGGED_IN"
	ErrorCodeTargetCredentialIncomplete = "TARGET_CREDENTIAL_INCOMPLETE"
	ErrorCodeTargetSessionExpired       = "TARGET_SESSION_EXPIRED"
	ErrorCodeTargetConnectorUnavailable = "TARGET_CONNECTOR_UNAVAILABLE"

	// N8 漫画发布依赖的作品/音频目录读取失败。这类错误是上游不可用，
	// 统一返回 503 而不是 500，便于前端区分“配置缺失”与“服务异常”。
	ErrorCodeN8DirectoryRejected = "N8_DIRECTORY_REJECTED"

	// 能力由 Slice 门控时，未注册路由仍返回 404，但响应体改为本码，便于
	// 前端和日志区分“能力未开启”与“路径写错”。
	ErrorCodeCapabilityNotEnabled = "CAPABILITY_NOT_ENABLED"
)

// 与 frontend/src/shared/apiErrorCodes.json 完全一致的文案镜像，仅用于契约
// 测试与未来按码取文案；现有 error 响应仍走各自的既有中文。
const (
	MessageImageModelNotConfigured    = "图片模型未配置：请到「管理员 → 水货生产模型」添加并启用图片模型，并填写服务地址与服务端密钥。"
	MessageVideoProviderNotConfigured = "视频生成通道未配置：请到「批量工厂 → 设置 → 视频生成通道」选择通道并填写密钥后保存。"
	MessageExecutorNotReady           = "执行器未就绪：请到「设置 → 执行器状态」生成配对码并安装执行器；巨量素材也可改用“优先直接获取原文”。"
	MessageGiantPluginNotConfigured   = "巨量素材解析服务未配置：请在服务器设置 QIANTIE_GIANT_MATERIAL_PLUGIN_URL/TOKEN（或青语 N8 令牌）后重试。"

	MessageTargetNotLoggedIn          = "尚未登录视频管理系统：请点击「登录并验证」，输入视频管理系统账号密码。"
	MessageTargetCredentialIncomplete = "凭证不全：只保存了视频管理系统账号，没有有效登录会话，请重新登录并验证。"
	MessageTargetSessionExpired       = "视频管理系统登录已过期：请重新登录并验证后再同步。"
	MessageTargetConnectorUnavailable = "无法连接视频管理系统：请检查服务器到 two.121w.com 的网络或 Go 连接器配置。"
	MessageTargetSessionVerified      = "已登录并验证"
)

// codeCarrier 让任何包（例如 batchfactoryv11）都能在不依赖 httpapi 的前提下
// 为错误附加稳定错误码。
type codeCarrier interface {
	ErrorCode() string
}

type codedError struct {
	code string
	err  error
}

func (e *codedError) Error() string     { return e.err.Error() }
func (e *codedError) Unwrap() error     { return e.err }
func (e *codedError) ErrorCode() string { return e.code }

// withErrorCode 保留原错误文本与 errors.Is 语义，只额外附加错误码。
func withErrorCode(err error, code string) error {
	if err == nil || code == "" {
		return err
	}
	return &codedError{code: code, err: err}
}

// errorCodeOf 读取错误链上的稳定错误码；没有则返回空串。
func errorCodeOf(err error) string {
	var carrier codeCarrier
	if errors.As(err, &carrier) {
		return carrier.ErrorCode()
	}
	return ""
}

// errorCodeOr 优先返回错误自带的错误码，否则使用兜底码。
func errorCodeOr(err error, fallback string) string {
	if code := errorCodeOf(err); code != "" {
		return code
	}
	return fallback
}
