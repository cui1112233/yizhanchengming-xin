package screenplay

import (
	"context"
	"errors"
	"testing"
)

func TestSubjectUploadOwnerAndCAS(t *testing.T) {
	r := NewMemorySubjectUploadRepository()
	ctx := context.Background()
	a, e := r.SaveSubjectUpload(ctx, SubjectUploadOperation{Owner: "alice", RequestKey: "request01", Digest: "x", Status: "paused"}, 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = r.GetSubjectUpload(ctx, "bob", a.RequestKey); !errors.Is(e, ErrNotFound) {
		t.Fatal(e)
	}
	if _, e = r.SaveSubjectUpload(ctx, a, 0); !errors.Is(e, ErrRevisionConflict) {
		t.Fatal(e)
	}
	a.Status = "running"
	if _, e = r.SaveSubjectUpload(ctx, a, a.Revision); e != nil {
		t.Fatal(e)
	}
}
