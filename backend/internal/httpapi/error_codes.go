package httpapi

import "errors"

// 四类“缺配置”错误的稳定错误码。前端按码显示引导文案，后端不再各自拼接
// 一套说法；code 是新增字段，原有 error 文案保持不变以兼容旧调用方。
const (
	ErrorCodeImageModelNotConfigured    = "IMAGE_MODEL_NOT_CONFIGURED"
	ErrorCodeVideoProviderNotConfigured = "VIDEO_PROVIDER_NOT_CONFIGURED"
	ErrorCodeExecutorNotReady           = "EXECUTOR_NOT_READY"
	ErrorCodeGiantPluginNotConfigured   = "GIANT_PLUGIN_NOT_CONFIGURED"
)

// 与 frontend/src/shared/apiErrorCodes.json 完全一致的文案镜像，仅用于契约
// 测试与未来按码取文案；现有 error 响应仍走各自的既有中文。
const (
	MessageImageModelNotConfigured    = "图片模型未配置：请到「管理员 → 水货生产模型」添加并启用图片模型，并填写服务地址与服务端密钥。"
	MessageVideoProviderNotConfigured = "视频生成通道未配置：请到「批量工厂 → 设置 → 视频生成通道」选择通道并填写密钥后保存。"
	MessageExecutorNotReady           = "执行器未就绪：请到「设置 → 执行器状态」生成配对码并安装执行器；巨量素材也可改用“优先直接获取原文”。"
	MessageGiantPluginNotConfigured   = "巨量素材解析服务未配置：请在服务器设置 QIANTIE_GIANT_MATERIAL_PLUGIN_URL/TOKEN（或青语 N8 令牌）后重试。"
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
