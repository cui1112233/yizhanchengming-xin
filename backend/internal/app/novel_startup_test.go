package app

import (
	"context"
	"errors"
	"os"
	"qiantie/backend/internal/httpapi"
	"qiantie/backend/internal/novelcreation"
	"strings"
	"testing"
	"time"
)

func TestNovelWorkerStartupFollowsAllFallibleBuildPreparation(t *testing.T) {
	raw, err := os.ReadFile("app.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	build := source[strings.Index(source, "func Build("):strings.Index(source, "func shouldStartLocalArtifactPurger(")]
	start := strings.Index(build, "novelcreation.StartWorker(")
	if start < 0 {
		start = strings.Index(build, "novelStartup.Start(")
	}
	configured := strings.Index(build, "userConfigStore = httpapi.NewUserConfigStore(")
	lastFailure := strings.LastIndex(build, "return nil,")
	router := strings.Index(build, "httpapi.NewRouter(")
	if start < configured || start < lastFailure || start < router {
		t.Fatal("novel workers start before configuration/router and fallible startup preparation finish")
	}
}

type delayedConfigNovelRunner struct {
	config **httpapi.UserConfigStore
	called chan bool
}

func (r delayedConfigNovelRunner) Run(_ context.Context, in novelcreation.RunContext, _ novelcreation.EventSink) (novelcreation.ChapterBundle, error) {
	ready := *r.config != nil
	r.called <- ready
	if !ready {
		return novelcreation.ChapterBundle{}, errors.New("channel configuration not ready")
	}
	return novelcreation.ChapterBundle{ProjectID: in.Project.ID, RunID: in.Run.ID, Chapter: novelcreation.Chapter{ID: "chapter", ProjectID: in.Project.ID, Number: 1, Title: "ready", Content: "saved", Version: 1}, FinalRunStatus: novelcreation.RunSucceeded}, nil
}
func TestNovelWorkerStartupDelaysExistingRunUntilConfigurationAndBuildReady(t *testing.T) {
	for _, failPreparation := range []bool{false, true} {
		t.Run(map[bool]string{false: "ready", true: "build_failure"}[failPreparation], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			store := novelcreation.NewMemoryStore()
			var configStore *httpapi.UserConfigStore
			called := make(chan bool, 1)
			service := novelcreation.NewService(store, novelcreation.NewMemoryRuntime(), delayedConfigNovelRunner{&configStore, called})
			project, err := service.CreateProject(ctx, "alice", novelcreation.CreateProjectInput{Title: "existing", Genre: "test", ModelID: "channel:text"})
			if err != nil {
				t.Fatal(err)
			}
			session, err := service.CreateSession(ctx, "alice", novelcreation.CreateSessionInput{ProjectID: project.ID, Mode: "long_form_original"})
			if err != nil {
				t.Fatal(err)
			}
			run, err := service.SendMessage(ctx, "alice", novelcreation.SendMessageInput{ProjectID: project.ID, SessionID: session.ID, ModelID: "channel:text", Content: "write", IdempotencyKey: "persisted-before-restart"})
			if err != nil {
				t.Fatal(err)
			}
			startup := &novelCreationStartup{service: service}
			allowConfig := make(chan struct{})
			configured := make(chan struct{})
			allowReady := make(chan struct{})
			prepared := make(chan error, 1)
			directory := t.TempDir()
			go func() {
				<-allowConfig
				configStore = httpapi.NewUserConfigStore(directory)
				close(configured)
				<-allowReady
				if failPreparation {
					prepared <- errors.New("later build validation failed")
					return
				}
				startup.Start(ctx)
				prepared <- nil
			}()
			assertUnstarted := func() {
				t.Helper()
				select {
				case <-called:
					t.Fatal("existing task ran during delayed initialization")
				case <-time.After(20 * time.Millisecond):
				}
				got, e := store.GetRun(ctx, "alice", run.ID)
				if e != nil || got.Status != novelcreation.RunQueued {
					t.Fatalf("startup changed queued task: %s %v", got.Status, e)
				}
			}
			assertUnstarted()
			close(allowConfig)
			<-configured
			assertUnstarted()
			close(allowReady)
			err = <-prepared
			if failPreparation {
				if err == nil {
					t.Fatal("expected startup failure")
				}
				assertUnstarted()
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			select {
			case ready := <-called:
				if !ready {
					t.Fatal("worker read uninitialized config")
				}
			case <-time.After(time.Second):
				t.Fatal("existing task not scanned after ready")
			}
		})
	}
}
