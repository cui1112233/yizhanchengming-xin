package agentstudio

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

type objectsFake struct {
	putKey, deleted string
	data            []byte
}

func (o *objectsFake) PutObjectFromFile(_ context.Context, _, key, path string) error {
	b, e := os.ReadFile(path)
	o.data = b
	o.putKey = key
	return e
}
func (o *objectsFake) GetObject(context.Context, string, string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(o.data)), nil
}
func (o *objectsFake) DeleteObject(_ context.Context, _, key string) error {
	o.deleted = key
	return nil
}
func TestAttachmentUploadReadDeleteIsProjectOwned(t *testing.T) {
	s := newMemoryStore()
	svc := NewService(s, UnavailableExecutor{})
	o := &objectsFake{}
	svc.SetObjects(o, "private")
	a := Actor{UserID: 1}
	p, e := svc.CreateProject(context.Background(), a, CreateProjectInput{Title: "x"})
	if e != nil {
		t.Fatal(e)
	}
	v, e := svc.UploadAttachment(context.Background(), a, p.ID, "note.txt", "text/plain", strings.NewReader("hello"))
	if e != nil {
		t.Fatal(e)
	}
	if !strings.HasPrefix(v.ObjectKey, "agent/project-") {
		t.Fatalf("key=%s", v.ObjectKey)
	}
	_, r, e := svc.OpenAttachment(context.Background(), a, p.ID, v.ID)
	if e != nil {
		t.Fatal(e)
	}
	b, _ := io.ReadAll(r)
	if string(b) != "hello" {
		t.Fatal(string(b))
	}
	if _, _, e = svc.OpenAttachment(context.Background(), Actor{UserID: 2}, p.ID, v.ID); e != ErrForbidden {
		t.Fatalf("foreign=%v", e)
	}
	if e = svc.DeleteAttachment(context.Background(), a, p.ID, v.ID); e != nil {
		t.Fatal(e)
	}
	if o.deleted != v.ObjectKey {
		t.Fatalf("deleted=%q", o.deleted)
	}
}
