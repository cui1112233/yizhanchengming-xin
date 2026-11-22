package novelfetchworkshop

import (
	"context"
	"fmt"
	"testing"
)

func TestClassifyUpstreamFailure(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want UpstreamFailureClass
	}{
		{"business 400", &UpstreamError{Class: UpstreamFailurePermanent, BusinessCode: 400}, UpstreamFailurePermanent},
		{"http 403", &UpstreamError{Class: UpstreamFailureConfiguration, HTTPStatus: 403}, UpstreamFailureConfiguration},
		{"http 429", &UpstreamError{Class: UpstreamFailureTransient, HTTPStatus: 429}, UpstreamFailureTransient},
		{"timeout", &UpstreamError{Class: UpstreamFailureTransient, Cause: context.DeadlineExceeded}, UpstreamFailureTransient},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ClassifyUpstreamFailure(fmt.Errorf("fetch: %w", tt.err))
			if !ok || got != tt.want {
				t.Fatalf("got %q %v, want %q", got, ok, tt.want)
			}
		})
	}
}
