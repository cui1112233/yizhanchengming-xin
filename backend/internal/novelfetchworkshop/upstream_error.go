package novelfetchworkshop

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
)

type UpstreamFailureClass string

const (
	UpstreamFailurePermanent     UpstreamFailureClass = "permanent"
	UpstreamFailureConfiguration UpstreamFailureClass = "configuration"
	UpstreamFailureTransient     UpstreamFailureClass = "transient"
)

type UpstreamError struct {
	Class        UpstreamFailureClass
	Operation    string
	HTTPStatus   int
	BusinessCode int
	Message      string
	Cause        error
}

func (e *UpstreamError) Error() string {
	parts := []string{"121 upstream", strings.TrimSpace(e.Operation)}
	if e.HTTPStatus > 0 {
		parts = append(parts, fmt.Sprintf("HTTP status %d", e.HTTPStatus))
	}
	if e.BusinessCode > 0 {
		parts = append(parts, fmt.Sprintf("business error %d", e.BusinessCode))
	}
	if message := strings.TrimSpace(e.Message); message != "" {
		parts = append(parts, message)
	}
	return strings.Join(parts, ": ")
}

func (e *UpstreamError) Unwrap() error { return e.Cause }

// safeUpstreamIOCause keeps errors.Is checks without retaining the transport's
// error text, which can contain credentials from URLs or response headers.
func safeUpstreamIOCause(ctx context.Context, err error) error {
	switch contextErr := ctx.Err(); {
	case errors.Is(err, context.DeadlineExceeded) || contextErr == context.DeadlineExceeded:
		return context.DeadlineExceeded
	case errors.Is(err, context.Canceled) || contextErr == context.Canceled:
		return context.Canceled
	case errors.Is(err, io.ErrUnexpectedEOF):
		return io.ErrUnexpectedEOF
	default:
		return nil
	}
}

func ClassifyUpstreamFailure(err error) (UpstreamFailureClass, bool) {
	var target *UpstreamError
	if !errors.As(err, &target) {
		return "", false
	}
	return target.Class, true
}
