package novelfetchworkshop

import (
	"errors"
	"fmt"
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
		parts = append(parts, fmt.Sprintf("business %d", e.BusinessCode))
	}
	if message := strings.TrimSpace(e.Message); message != "" {
		parts = append(parts, message)
	}
	return strings.Join(parts, ": ")
}

func (e *UpstreamError) Unwrap() error { return e.Cause }

func ClassifyUpstreamFailure(err error) (UpstreamFailureClass, bool) {
	var target *UpstreamError
	if !errors.As(err, &target) {
		return "", false
	}
	return target.Class, true
}
