package novelfetch

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestModuleHandlerLifecycleWithInjectedBoundaries(t *testing.T){
	store:=newMemoryStore()
	service:=NewService(Options{
		Store:store,
		Fetcher:&fakeFetcher{result:FetchResult{OriginalRaw:"正文123456789012345"}},
		Model:&flakyModel{calls:map[string]int{}},
		Dispatcher:&fakeDispatcher{},
		Publisher:&fakePublisher{},
		BatchFactory:&fakeHandoff{},
		Now:fixedNow,
	})
	handler:=NewModuleHandler(service)

	body:=[]byte(`{"name":"http","groups":[{"source":"阳光","platformId":"4","books":[{"bookId":"123"}]}]}`)
	req:=httptest.NewRequest(http.MethodPost,"/batches",bytes.NewReader(body))
	rec:=httptest.NewRecorder();handler.ServeHTTP(rec,req)
	if rec.Code!=http.StatusCreated{t.Fatalf("create status=%d body=%s",rec.Code,rec.Body.String())}
	var created struct{Batch Batch `json:"batch"`}
	if err:=json.Unmarshal(rec.Body.Bytes(),&created);err!=nil{t.Fatal(err)}

	req=httptest.NewRequest(http.MethodPost,"/batches/"+created.Batch.ID+"/runs",bytes.NewReader([]byte(`{}`)))
	rec=httptest.NewRecorder();handler.ServeHTTP(rec,req)
	if rec.Code!=http.StatusCreated{t.Fatalf("run status=%d body=%s",rec.Code,rec.Body.String())}
	var started struct{Run Run `json:"run"`}
	if err:=json.Unmarshal(rec.Body.Bytes(),&started);err!=nil{t.Fatal(err)}

	req=httptest.NewRequest(http.MethodPost,"/runs/"+started.Run.ID+"/execute",nil)
	rec=httptest.NewRecorder();handler.ServeHTTP(rec,req)
	if rec.Code!=http.StatusOK{t.Fatalf("execute=%d body=%s",rec.Code,rec.Body.String())}

	req=httptest.NewRequest(http.MethodGet,"/runs/"+started.Run.ID+"/records",nil)
	rec=httptest.NewRecorder();handler.ServeHTTP(rec,req)
	if rec.Code!=http.StatusOK{t.Fatalf("records=%d body=%s",rec.Code,rec.Body.String())}
	var audit struct{Records []Record `json:"records"`}
	if err:=json.Unmarshal(rec.Body.Bytes(),&audit);err!=nil{t.Fatal(err)}
	if len(audit.Records)==0{t.Fatal("expected processing records")}
}

func TestModuleHandlerUnavailableBoundariesDoNotFakeSuccess(t *testing.T){
	store:=newMemoryStore()
	service:=NewService(Options{Store:store,Dispatcher:&fakeDispatcher{},Now:fixedNow})
	batch,_,_:=service.CreateBatch(context.Background(),CreateBatchInput{Groups:[]GroupInput{{Source:"阳光",PlatformID:"4",Books:[]BookInput{{BookID:"1"}}}}})
	run,err:=service.StartRun(context.Background(),batch.ID,time.Time{});if err!=nil{t.Fatal(err)}
	handler:=NewModuleHandler(service)

	req:=httptest.NewRequest(http.MethodPost,"/runs/"+run.ID+"/handoff",nil)
	rec:=httptest.NewRecorder();handler.ServeHTTP(rec,req)
	if rec.Code!=http.StatusServiceUnavailable{t.Fatalf("handoff=%d body=%s",rec.Code,rec.Body.String())}

	req=httptest.NewRequest(http.MethodPost,"/runs/"+run.ID+"/books/4_1/submit-intents",bytes.NewReader([]byte(`{"version":"ai1","publishingAccountId":1}`)))
	rec=httptest.NewRecorder();handler.ServeHTTP(rec,req)
	if rec.Code!=http.StatusServiceUnavailable{t.Fatalf("submit=%d body=%s",rec.Code,rec.Body.String())}
}
