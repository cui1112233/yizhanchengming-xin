package batchfactoryv11

import "testing"

func TestVideoModelProviderBindingUsesTheModelAsTheSourceOfTruth(t *testing.T) {
	cases := []struct {
		name     string
		modelID  string
		fallback string
		want     string
	}{
		{name: "H3 catalog model", modelID: "minimax-h3-video", fallback: VideoProviderPersonalAPI, want: VideoProviderAutoDLH3},
		{name: "YD catalog model", modelID: "yd2-mini-video", fallback: VideoProviderAutoDLH3, want: VideoProviderPersonalAPI},
		{name: "YFAI Seedance catalog model", modelID: "seedance-2-0-official", fallback: VideoProviderPersonalAPI, want: VideoProviderYFAISeedance},
		{name: "local executor catalog model", modelID: "local-doubao-executor-video", fallback: VideoProviderPersonalAPI, want: VideoProviderDoubaoLocal},
		{name: "unknown model keeps saved provider", modelID: "custom-video", fallback: VideoProviderAutoDLH3, want: VideoProviderAutoDLH3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := VideoProviderForModel(tc.modelID, tc.fallback); got != tc.want {
				t.Fatalf("VideoProviderForModel(%q, %q) = %q, want %q", tc.modelID, tc.fallback, got, tc.want)
			}
		})
	}
}

func TestVideoCatalogEnginesAddYFAIWithoutDroppingExistingGroups(t *testing.T) {
	want := []string{VideoProviderPersonalAPI, VideoProviderDoubaoLocal, VideoProviderAutoDLH3, VideoProviderYFAI, VideoProviderYFAISeedance}
	if len(VideoCatalogEngines) != len(want) {
		t.Fatalf("engine count = %d, want %d: %#v", len(VideoCatalogEngines), len(want), VideoCatalogEngines)
	}
	for index, engine := range want {
		if VideoCatalogEngines[index] != engine {
			t.Fatalf("engine[%d] = %q, want %q", index, VideoCatalogEngines[index], engine)
		}
	}
	if len(want) != 5 || VideoCatalogEngines[3] != VideoProviderYFAI {
		t.Fatal("YFAI must be added as a new group without renaming existing ones")
	}
}

func TestCatalogEngineForModelSplitsYFAIOfficialFromFanKe(t *testing.T) {
	cases := []struct{ modelID, fallback, want string }{
		{modelID: "seedance-2-0-official", fallback: VideoProviderYFAISeedance, want: VideoProviderYFAI},
		{modelID: "fanke-video-api", fallback: VideoProviderYFAI, want: VideoProviderYFAISeedance},
		{modelID: "minimax-h3-video", fallback: VideoProviderYFAI, want: VideoProviderAutoDLH3},
		{modelID: "yd2-mini-video", fallback: VideoProviderYFAI, want: VideoProviderPersonalAPI},
		{modelID: "local-doubao-executor-video", fallback: VideoProviderYFAI, want: VideoProviderDoubaoLocal},
		// 旧配置仍停留在梵客分组，不被新分组吞并。
		{modelID: "legacy-video", fallback: VideoProviderYFAISeedance, want: VideoProviderYFAISeedance},
		{modelID: "legacy-video", fallback: VideoProviderYFAI, want: VideoProviderYFAI},
	}
	for _, tc := range cases {
		if got := CatalogEngineForModel(tc.modelID, tc.fallback); got != tc.want {
			t.Fatalf("CatalogEngineForModel(%q, %q) = %q, want %q", tc.modelID, tc.fallback, got, tc.want)
		}
	}
	// YFAI 只是分组口径，执行层仍走既有 YFAI Seedance 适配器。
	if got := normalizeVideoProvider(VideoProviderYFAI); got != VideoProviderYFAISeedance {
		t.Fatalf("normalizeVideoProvider(yfai) = %q, want %q", got, VideoProviderYFAISeedance)
	}
}

func TestFanKeConcreteModelSelectionBindsToFanKeProvider(t *testing.T) {
	const selection = "fanke-video-api::ft-video-v1-77e8ee7a636f15dac27b2ce6d6fcd746"
	if got := VideoProviderForModel(selection, VideoProviderPersonalAPI); got != VideoProviderYFAISeedance {
		t.Fatalf("VideoProviderForModel(%q) = %q, want %q", selection, got, VideoProviderYFAISeedance)
	}
	if got := CatalogEngineForModel(selection, VideoProviderYFAI); got != VideoProviderYFAISeedance {
		t.Fatalf("CatalogEngineForModel(%q) = %q, want %q", selection, got, VideoProviderYFAISeedance)
	}
	if got := normalizeVideoProvider(selection); got != VideoProviderYFAISeedance {
		t.Fatalf("normalizeVideoProvider(%q) = %q, want %q", selection, got, VideoProviderYFAISeedance)
	}
	if got := FanKeModelFromSelection(selection); got != "ft-video-v1-77e8ee7a636f15dac27b2ce6d6fcd746" {
		t.Fatalf("FanKeModelFromSelection(%q) = %q", selection, got)
	}
	// 旧值（裸 fanke-video-api）仍能执行，并回退到目录默认型号。
	if got := FanKeModelFromSelection("fanke-video-api"); got != "" {
		t.Fatalf("bare FanKe selection must fall back to the catalog default, got %q", got)
	}
	if got := VideoProviderForModel("fanke-video-api", VideoProviderYFAI); got != VideoProviderYFAISeedance {
		t.Fatalf("legacy bare FanKe model = %q, want %q", got, VideoProviderYFAISeedance)
	}
	if !VideoModelMatchesProviderModel(selection, "seedance-2-0-official", VideoProviderYFAISeedance) {
		t.Fatal("a concrete FanKe selection must match the catalog default model")
	}
}

func TestVideoModelMatchesProviderModelAcceptsCatalogAliases(t *testing.T) {
	if !VideoModelMatchesProviderModel("yd2-mini-video", "yd2.0-mini", VideoProviderPersonalAPI) {
		t.Fatal("YD catalog id must match the personal API provider model")
	}
	if !VideoModelMatchesProviderModel("local-doubao-executor-video", "doubao-seedance", VideoProviderDoubaoLocal) {
		t.Fatal("local catalog id must match the local executor provider model")
	}
	if VideoModelMatchesProviderModel("minimax-h3-video", "yd2.0-mini", VideoProviderPersonalAPI) {
		t.Fatal("H3 model must not match the personal API provider model")
	}
	if !VideoModelMatchesProviderModel("seedance-2-0-official", "seedance-2-0-official", VideoProviderYFAISeedance) {
		t.Fatal("Seedance catalog and provider model should match")
	}
}
