package videomanagementconnector

import (
	"context"
	"errors"
)

// Probe the readback endpoint, never replay a POST to discover whether the
// remote API is healthy. No matching record is a valid response to this probe.
func (s *TargetHTTPService) ProbeSubmissionContract(ctx context.Context, session Session) error {
	if err := s.CheckSession(ctx, session); err != nil {
		return err
	}
	_, err := s.GetSubmission(ctx, session, "0")
	if err != nil && (errors.Is(err, ErrSiteContractChanged) || errors.Is(err, ErrSessionExpired) || errors.Is(err, ErrSessionMissing)) {
		return err
	}
	// Only a structurally valid response may release the circuit. The normal
	// no-record condition is typed separately below; transport failures stay shut.
	if errors.Is(err, ErrSubmissionNotFound) {
		return nil
	}
	return err
}
