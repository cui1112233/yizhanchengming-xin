package videogen

import (
	"context"
	"errors"
	"testing"
)

type configSourceFake struct {
	provider string
	config Config
	err error
	owner string
}
func (f *configSourceFake) Resolve(_ context.Context, owner, provider string) (Config,error) {
	f.owner, f.provider = owner, provider
	return f.config, f.err
}

type localProviderFake struct{}
func (localProviderFake) Submit(context.Context, Request)(Task,error){ return Task{ID:"local",State:StateQueued},nil }
func (localProviderFake) Poll(context.Context, Task)(Task,error){ return Task{ID:"local",State:StateRunning},nil }

func TestFactoryResolvesProviderFromSelectedModel(t *testing.T) {
	source:=&configSourceFake{config:Config{Provider:ProviderPersonalAPI,APIKey:"k",Model:"yd2.0-mini"}}
	factory:=Factory{Configs:source}
	provider, resolved, err:=factory.ForModel(context.Background(),"owner","yd2-mini-video",ProviderAutoDLH3)
	if err!=nil{t.Fatal(err)}
	if _,ok:=provider.(PersonalProvider);!ok{t.Fatalf("provider=%T",provider)}
	if resolved!=ProviderPersonalAPI||source.provider!=ProviderPersonalAPI||source.owner!="owner"{t.Fatalf("resolved=%q source=%#v",resolved,source)}
}

func TestFactoryBuildsYFAIAndAutoDL(t *testing.T){
	for _,tc:=range []struct{model,provider string; want any}{
		{"seedance-2-0-official",ProviderYFAISeedance,YFAIProvider{}},
		{"minimax-h3-video",ProviderAutoDLH3,AutoDLProvider{}},
	}{
		source:=&configSourceFake{config:Config{Provider:tc.provider,APIKey:"k",Model:tc.model}}
		factory:=Factory{Configs:source}
		got,_,err:=factory.ForModel(context.Background(),"owner",tc.model,ProviderPersonalAPI);if err!=nil{t.Fatal(err)}
		switch tc.provider{case ProviderYFAISeedance:if _,ok:=got.(YFAIProvider);!ok{t.Fatalf("provider=%T",got)};case ProviderAutoDLH3:if _,ok:=got.(AutoDLProvider);!ok{t.Fatalf("provider=%T",got)}}
	}
}

func TestFactoryUsesInjectedLocalExecutorProvider(t *testing.T){
	source:=&configSourceFake{config:Config{Provider:ProviderDoubaoLocal,Model:"doubao-seedance"}}
	factory:=Factory{Configs:source,Local:localProviderFake{}}
	got,resolved,err:=factory.ForModel(context.Background(),"owner","local-doubao-executor-video",ProviderPersonalAPI)
	if err!=nil{t.Fatal(err)}
	if _,ok:=got.(localProviderFake);!ok{t.Fatalf("provider=%T",got)}
	if resolved!=ProviderDoubaoLocal{t.Fatalf("resolved=%q",resolved)}
}

func TestFactoryRejectsUnavailableLocalExecutor(t *testing.T){
	source:=&configSourceFake{config:Config{Provider:ProviderDoubaoLocal,Model:"doubao-seedance"}}
	factory:=Factory{Configs:source}
	if _,_,err:=factory.ForModel(context.Background(),"owner","local-doubao-executor-video",ProviderPersonalAPI);!errors.Is(err,ErrProviderUnavailable){t.Fatalf("err=%v",err)}
}
