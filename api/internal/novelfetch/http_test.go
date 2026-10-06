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

func request(handler http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	req:=httptest.NewRequest(method,path,bytes.NewReader(body))
	rec:=httptest.NewRecorder()
	handler.ServeHTTP(rec,req)
	return rec
}

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

	rec:=request(handler,http.MethodPost,"/batches",[]byte(`{"name":"http","groups":[{"source":"阳光","platformId":"4","books":[{"bookId":"123"}]}]}`))
	if rec.Code!=http.StatusCreated{t.Fatalf("create status=%d body=%s",rec.Code,rec.Body.String())}
	var created struct{Batch Batch `json:"batch"`}
	if err:=json.Unmarshal(rec.Body.Bytes(),&created);err!=nil{t.Fatal(err)}

	rec=request(handler,http.MethodPost,"/batches/"+created.Batch.ID+"/runs",[]byte(`{}`))
	if rec.Code!=http.StatusCreated{t.Fatalf("run status=%d body=%s",rec.Code,rec.Body.String())}
	var started struct{Run Run `json:"run"`}
	if err:=json.Unmarshal(rec.Body.Bytes(),&started);err!=nil{t.Fatal(err)}

	rec=request(handler,http.MethodPost,"/runs/"+started.Run.ID+"/execute",nil)
	if rec.Code!=http.StatusOK{t.Fatalf("execute=%d body=%s",rec.Code,rec.Body.String())}

	rec=request(handler,http.MethodGet,"/runs/"+started.Run.ID,nil)
	if rec.Code!=http.StatusOK{t.Fatalf("state=%d body=%s",rec.Code,rec.Body.String())}
	var state struct{Run Run `json:"run"`; Books []Book `json:"books"`}
	if err:=json.Unmarshal(rec.Body.Bytes(),&state);err!=nil{t.Fatal(err)}
	if state.Run.Status!=StatusSucceeded||len(state.Books)!=1{t.Fatalf("state=%+v books=%d",state.Run,len(state.Books))}

	rec=request(handler,http.MethodGet,"/runs/"+started.Run.ID+"/records",nil)
	if rec.Code!=http.StatusOK{t.Fatalf("records=%d body=%s",rec.Code,rec.Body.String())}
}

func TestKnowledgeCreateGeneratesDistinctIDsAndEditsDoNotOverwriteSibling(t *testing.T){
	store:=newMemoryStore()
	handler:=NewModuleHandler(NewService(Options{Store:store,Now:fixedNow}))

	first:=request(handler,http.MethodPost,"/knowledge",[]byte(`{"kind":"rewrite","title":"A","content":"one","enabled":true}`))
	second:=request(handler,http.MethodPost,"/knowledge",[]byte(`{"kind":"rewrite","title":"B","content":"two","enabled":true}`))
	if first.Code!=http.StatusCreated||second.Code!=http.StatusCreated{t.Fatalf("create codes=%d,%d",first.Code,second.Code)}
	var a,b struct{Item KnowledgeEntry `json:"item"`}
	if err:=json.Unmarshal(first.Body.Bytes(),&a);err!=nil{t.Fatal(err)}
	if err:=json.Unmarshal(second.Body.Bytes(),&b);err!=nil{t.Fatal(err)}
	if a.Item.ID==""||b.Item.ID==""||a.Item.ID==b.Item.ID{t.Fatalf("ids=%q,%q",a.Item.ID,b.Item.ID)}

	editA:=request(handler,http.MethodPut,"/knowledge/"+a.Item.ID,[]byte(`{"kind":"rewrite","title":"A","content":"one-edited","enabled":false}`))
	editB:=request(handler,http.MethodPut,"/knowledge/"+b.Item.ID,[]byte(`{"kind":"rewrite","title":"B","content":"two-edited","enabled":true}`))
	if editA.Code!=http.StatusOK||editB.Code!=http.StatusOK{t.Fatalf("edit codes=%d,%d",editA.Code,editB.Code)}

	list:=request(handler,http.MethodGet,"/knowledge?kind=rewrite",nil)
	var payload struct{Items []KnowledgeEntry `json:"items"`}
	if err:=json.Unmarshal(list.Body.Bytes(),&payload);err!=nil{t.Fatal(err)}
	if len(payload.Items)!=2{t.Fatalf("items=%+v",payload.Items)}
	byID:=map[string]KnowledgeEntry{};for _,item:=range payload.Items{byID[item.ID]=item}
	if byID[a.Item.ID].Content!="one-edited"||byID[a.Item.ID].Enabled{t.Fatalf("A=%+v",byID[a.Item.ID])}
	if byID[b.Item.ID].Content!="two-edited"||!byID[b.Item.ID].Enabled{t.Fatalf("B=%+v",byID[b.Item.ID])}

	bad:=request(handler,http.MethodPut,"/knowledge/new",[]byte(`{"kind":"rewrite","content":"bad","enabled":true}`))
	if bad.Code!=http.StatusBadRequest{t.Fatalf("/knowledge/new must not create: %d",bad.Code)}
}

func TestRetryHTTPOnlySchedulesAndRunReadbackShowsQueued(t *testing.T){
	store:=newMemoryStore()
	dispatcher:=&fakeDispatcher{}
	model:=&flakyModel{calls:map[string]int{},failVersion:"ai2"}
	service:=NewService(Options{Store:store,Fetcher:&fakeFetcher{result:FetchResult{OriginalRaw:"正文123456789012345"}},Model:model,Dispatcher:dispatcher,Now:fixedNow})
	batch,_,_:=service.CreateBatch(context.Background(),CreateBatchInput{Groups:[]GroupInput{{Source:"阳光",PlatformID:"4",Books:[]BookInput{{BookID:"1"}}}}})
	run,_:=service.StartRun(context.Background(),batch.ID,time.Time{})
	_,_ = service.ExecuteRun(context.Background(),run.ID)
	handler:=NewModuleHandler(service)

	rec:=request(handler,http.MethodPost,"/runs/"+run.ID+"/books/4_1/retry",nil)
	if rec.Code!=http.StatusAccepted{t.Fatalf("retry=%d body=%s",rec.Code,rec.Body.String())}
	if len(dispatcher.retries)!=1{t.Fatalf("retry dispatch=%+v",dispatcher.retries)}
	if model.calls["ai2"]!=1{t.Fatalf("HTTP retry executed model: %v",model.calls)}

	rec=request(handler,http.MethodGet,"/runs/"+run.ID,nil)
	var state struct{Run Run `json:"run"`; Books []Book `json:"books"`}
	if err:=json.Unmarshal(rec.Body.Bytes(),&state);err!=nil{t.Fatal(err)}
	if state.Run.Status!=StatusQueued||state.Books[0].Status!=StatusQueued{t.Fatalf("state=%s book=%s",state.Run.Status,state.Books[0].Status)}
}

func TestSecondRunReturnsConflictInsteadOfReusingSharedBookResults(t *testing.T){
	store:=newMemoryStore()
	service:=NewService(Options{Store:store,Dispatcher:&fakeDispatcher{},Now:fixedNow})
	batch,_,_:=service.CreateBatch(context.Background(),CreateBatchInput{Groups:[]GroupInput{{Source:"阳光",PlatformID:"4",Books:[]BookInput{{BookID:"1"}}}}})
	handler:=NewModuleHandler(service)
	first:=request(handler,http.MethodPost,"/batches/"+batch.ID+"/runs",[]byte(`{}`))
	second:=request(handler,http.MethodPost,"/batches/"+batch.ID+"/runs",[]byte(`{}`))
	if first.Code!=http.StatusCreated||second.Code!=http.StatusConflict{t.Fatalf("codes=%d,%d second=%s",first.Code,second.Code,second.Body.String())}
}

func TestModuleHandlerUnavailableBoundariesDoNotFakeSuccess(t *testing.T){
	store:=newMemoryStore()
	service:=NewService(Options{Store:store,Dispatcher:&fakeDispatcher{},Now:fixedNow})
	batch,_,_:=service.CreateBatch(context.Background(),CreateBatchInput{Groups:[]GroupInput{{Source:"阳光",PlatformID:"4",Books:[]BookInput{{BookID:"1"}}}}})
	run,err:=service.StartRun(context.Background(),batch.ID,time.Time{});if err!=nil{t.Fatal(err)}
	handler:=NewModuleHandler(service)

	rec:=request(handler,http.MethodPost,"/runs/"+run.ID+"/handoff",nil)
	if rec.Code!=http.StatusServiceUnavailable{t.Fatalf("handoff=%d body=%s",rec.Code,rec.Body.String())}

	rec=request(handler,http.MethodPost,"/runs/"+run.ID+"/books/4_1/submit-intents",[]byte(`{"version":"ai1","publishingAccountId":1}`))
	if rec.Code!=http.StatusServiceUnavailable{t.Fatalf("submit=%d body=%s",rec.Code,rec.Body.String())}
}
