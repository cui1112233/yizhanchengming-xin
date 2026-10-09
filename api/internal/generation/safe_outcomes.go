package generation

import "errors"

// Outcome is fixed public feedback. Keep internal causes separate for the
// existing correlated, sanitized logger and sentinel classification.
type Outcome struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RuntimeOutcomeError implements task9runtime.SafeExecutionError without
// making the runtime import the generation domain. Cause remains available to
// trusted diagnostics through errors.Unwrap, while persistence sees only the
// allowlisted code/message.
type RuntimeOutcomeError struct {
	cause     error
	outcome   Outcome
	retryable bool
}

func NewRuntimeOutcomeError(cause error) error {
	if cause == nil {
		return nil
	}
	outcome := OutcomeForError(cause)
	retryable := !errors.Is(cause, ErrInvalid) && !errors.Is(cause, ErrNotFound) && !errors.Is(cause, ErrConflict) && !errors.Is(cause, ErrProjectArchived)
	return &RuntimeOutcomeError{cause: cause, outcome: outcome, retryable: retryable}
}

func (e *RuntimeOutcomeError) Error() string       { return e.outcome.Message }
func (e *RuntimeOutcomeError) SafeCode() string    { return e.outcome.Code }
func (e *RuntimeOutcomeError) SafeMessage() string { return e.outcome.Message }
func (e *RuntimeOutcomeError) Retryable() bool     { return e.retryable }
func (e *RuntimeOutcomeError) Unwrap() error       { return e.cause }

var outcomeCatalogue = [...]struct {
	cause error
	Outcome
}{
	{ErrInvalid, Outcome{"GENERATION_INVALID", "生成参数无效，请检查后重试"}},
	{ErrNotFound, Outcome{"GENERATION_NOT_FOUND", "生成记录不存在，请刷新后重试"}},
	{ErrConflict, Outcome{"GENERATION_CONFLICT", "当前生成状态不允许此操作，请刷新后重试"}},
	{ErrProjectArchived, Outcome{"BATCH_PROJECT_ARCHIVED", "项目已归档，请先恢复后再生成"}},
	{ErrAudioMeasurementRequired, Outcome{"AUDIO_MEASUREMENT_REQUIRED", "请先生成或检测音频"}},
	{ErrAudioProbeUnavailable, Outcome{"AUDIO_PROBE_UNAVAILABLE", "音频检测服务暂不可用，请稍后重试"}},
	{ErrTimelineValidation, Outcome{"GENERATION_TIMELINE_INVALID", "导演分镜时长校验失败，请重试导演阶段"}},
	{ErrUnavailable, Outcome{"GENERATION_UNAVAILABLE", "生成服务暂不可用，请稍后重试"}},
	{nil, Outcome{"GENERATION_FAILED", "生成阶段执行失败，请稍后重试"}},
}

// OutcomeForError uses sentinel identity, never provider text or wrapped suffixes.
func OutcomeForError(err error) Outcome {
	if err == nil {
		return Outcome{}
	}
	for _, entry := range outcomeCatalogue {
		if entry.cause == nil || errors.Is(err, entry.cause) {
			return entry.Outcome
		}
	}
	return Outcome{}
}

// OutcomeForPersistedMessage recognizes only an exact approved message. Codes
// are derived response metadata; historical text is never interpreted as a cause.
func OutcomeForPersistedMessage(message string) Outcome {
	if message == "" {
		return Outcome{}
	}
	for _, entry := range outcomeCatalogue {
		if entry.Message == message || entry.cause == nil {
			return entry.Outcome
		}
	}
	return Outcome{}
}
