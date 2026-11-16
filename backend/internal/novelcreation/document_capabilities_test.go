package novelcreation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDocumentPublicCapabilitiesFollowUploadLease(t *testing.T) {
	ctx := context.Background()
	s, st, p := materialService(t)
	s.ConfigureDocuments(&fakeDocumentObjects{objects: map[string][]byte{}}, nil)
	fail := true
	s.store = failDocumentSave{st, &fail}
	j, err := s.UploadDocument(ctx, "alice", p.ID, "a.txt", strings.NewReader("正文"))
	if err == nil {
		t.Fatal("expected finalization failure")
	}
	rec, _, _ := st.GetContentRecord(ctx, "alice", p.ID, "document", j.ID)
	for _, expired := range []bool{false, true} {
		now := rec.LeaseUntil.Add(-time.Nanosecond)
		if expired {
			now = rec.LeaseUntil
		}
		s.store = documentClockStore{st, now}
		jobs, e := s.ListDocuments(ctx, "alice", p.ID)
		if e != nil || len(jobs) != 1 {
			t.Fatalf("list: %v %v", jobs, e)
		}
		single, e := s.GetDocument(ctx, "alice", p.ID, j.ID)
		if e != nil {
			t.Fatal(e)
		}
		for _, job := range []DocumentJob{jobs[0], single} {
			b, _ := json.Marshal(job)
			var v map[string]any
			json.Unmarshal(b, &v)
			if v["canRetry"] != expired || v["canDelete"] != expired {
				t.Fatalf("expired=%v public=%s", expired, b)
			}
			if strings.Contains(string(b), j.Object.Key) && j.Object.Key != "" {
				t.Fatal("private object key exposed")
			}
		}
	}
}
