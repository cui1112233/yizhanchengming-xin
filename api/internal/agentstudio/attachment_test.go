package agentstudio

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/objectkey"
)

type objectsFake struct {
	putKey, getKey, deleted string
	data                    []byte
}

func (o *objectsFake) PutObjectFromFile(_ context.Context, _, key, path string) error {
	b, e := os.ReadFile(path)
	o.data = b
	o.putKey = key
	return e
}
func (o *objectsFake) GetObject(_ context.Context, _ string, key string) (io.ReadCloser, error) {
	o.getKey = key
	return io.NopCloser(bytes.NewReader(o.data)), nil
}
func (o *objectsFake) DeleteObject(_ context.Context, _, key string) error {
	o.deleted = key
	return nil
}
func TestAttachmentUploadReadDeleteIsProjectOwned(t *testing.T) {
	for _, prefix := range []string{"", "staging/"} {
		t.Run(prefix, func(t *testing.T) { attachmentUploadReadDelete(t, prefix) })
	}
}

func attachmentUploadReadDelete(t *testing.T, rawPrefix string) {
	t.Helper()
	s := newMemoryStore()
	svc := NewService(s, UnavailableExecutor{})
	o := &objectsFake{}
	svc.SetObjects(o, "private")
	prefix, err := objectkey.ParsePrefix(rawPrefix)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetKeyPrefix(prefix)
	a := Actor{UserID: 1}
	p, e := svc.CreateProject(context.Background(), a, CreateProjectInput{Title: "x"})
	if e != nil {
		t.Fatal(e)
	}
	v, e := svc.UploadAttachment(context.Background(), a, p.ID, "note.txt", "text/plain", strings.NewReader("hello"))
	if e != nil {
		t.Fatal(e)
	}
	if !strings.HasPrefix(v.ObjectKey, rawPrefix+"agent/project-") || v.ObjectKey != o.putKey || strings.Count(v.ObjectKey, "staging/") != strings.Count(rawPrefix, "staging/") {
		t.Fatalf("key=%s", v.ObjectKey)
	}
	// Reads and deletion must use the saved reference even after configuration changes.
	next, err := objectkey.ParsePrefix("next-run/")
	if err != nil {
		t.Fatal(err)
	}
	svc.SetKeyPrefix(next)
	_, r, e := svc.OpenAttachment(context.Background(), a, p.ID, v.ID)
	if e != nil {
		t.Fatal(e)
	}
	defer r.Close()
	b, _ := io.ReadAll(r)
	if string(b) != "hello" || o.getKey != v.ObjectKey {
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

type failingAttachmentStore struct{ Store }

func (f failingAttachmentStore) CreateAttachment(context.Context, Actor, Attachment) (Attachment, error) {
	return Attachment{}, errors.New("fake database failure")
}

func TestAttachmentKeyPrefixFailedCreateDeletesUploadedKey(t *testing.T) {
	for _, raw := range []string{"", "staging/"} {
		t.Run(raw, func(t *testing.T) {
			store := newMemoryStore()
			actor := Actor{UserID: 1}
			project, err := store.CreateProject(context.Background(), actor, CreateProjectInput{Title: "x"})
			if err != nil {
				t.Fatal(err)
			}
			svc := NewService(failingAttachmentStore{Store: store}, UnavailableExecutor{})
			objects := &objectsFake{}
			svc.SetObjects(objects, "private")
			prefix, err := objectkey.ParsePrefix(raw)
			if err != nil {
				t.Fatal(err)
			}
			svc.SetKeyPrefix(prefix)
			_, err = svc.UploadAttachment(context.Background(), actor, project.ID, "note.txt", "text/plain", strings.NewReader("hello"))
			if err == nil {
				t.Fatal("expected database failure")
			}
			if !strings.HasPrefix(objects.putKey, raw+"agent/project-") || objects.deleted != objects.putKey {
				t.Fatalf("upload=%q cleanup=%q", objects.putKey, objects.deleted)
			}
		})
	}
}
