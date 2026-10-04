package videogen

import "testing"

func TestProviderForModelPreservesV88Bindings(t *testing.T) {
	cases := []struct{ model, fallback, want string }{
		{"minimax-h3-video", ProviderPersonalAPI, ProviderAutoDLH3},
		{"yd2-mini-video", ProviderAutoDLH3, ProviderPersonalAPI},
		{"yd2.0-mini", ProviderYFAISeedance, ProviderPersonalAPI},
		{"seedance-2-0-official", ProviderPersonalAPI, ProviderYFAISeedance},
		{"local-doubao-executor-video", ProviderPersonalAPI, ProviderDoubaoLocal},
		{"doubao-seedance", ProviderPersonalAPI, ProviderDoubaoLocal},
		{"custom-video", ProviderAutoDLH3, ProviderAutoDLH3},
	}
	for _, tc := range cases {
		if got := ProviderForModel(tc.model, tc.fallback); got != tc.want {
			t.Fatalf("ProviderForModel(%q,%q)=%q want %q", tc.model, tc.fallback, got, tc.want)
		}
	}
}

func TestNormalizeProviderPreservesV88Aliases(t *testing.T) {
	cases := map[string]string{
		"": ProviderPersonalAPI,
		"personal": ProviderPersonalAPI,
		"yd_video": ProviderPersonalAPI,
		"yadi": ProviderPersonalAPI,
		"doubao": ProviderDoubaoLocal,
		"local-doubao-executor-video": ProviderDoubaoLocal,
		"h3": ProviderAutoDLH3,
		"minimax-h3-video": ProviderAutoDLH3,
		"yfai": ProviderYFAISeedance,
		"seedance-2-0-official": ProviderYFAISeedance,
	}
	for input, want := range cases {
		if got := NormalizeProvider(input); got != want { t.Fatalf("NormalizeProvider(%q)=%q want %q", input, got, want) }
	}
}

func TestProviderDefaultsMatchV88(t *testing.T) {
	cfg, err := NormalizeConfig(Config{Provider: ProviderPersonalAPI, APIKey: "key"})
	if err != nil { t.Fatal(err) }
	if cfg.Model != "yd2.0-mini" || cfg.CreateURL != DefaultPersonalCreateURL || cfg.TasksURL != DefaultPersonalTasksURL {
		t.Fatalf("personal defaults=%#v", cfg)
	}
	if cfg.ResultURL != DefaultPersonalTasksURL+"/{id}/result" { t.Fatalf("result url=%q", cfg.ResultURL) }

	yfai, err := NormalizeConfig(Config{Provider: ProviderYFAISeedance, APIKey: "key"})
	if err != nil { t.Fatal(err) }
	if yfai.Model != "seedance-2-0-official" || yfai.CreateURL != "https://yf.token6688.com" { t.Fatalf("yfai defaults=%#v", yfai) }
}

func TestProviderConfigRequiresCredentialsExceptLocalExecutor(t *testing.T) {
	for _, provider := range []string{ProviderPersonalAPI, ProviderAutoDLH3, ProviderYFAISeedance} {
		if _, err := NormalizeConfig(Config{Provider: provider}); err == nil { t.Fatalf("provider %s must require API key", provider) }
	}
	if _, err := NormalizeConfig(Config{Provider: ProviderDoubaoLocal}); err != nil { t.Fatal(err) }
}
