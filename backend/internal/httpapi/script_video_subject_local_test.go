package httpapi

import "testing"

func TestLocalVideoRejectsUnimplementedSubjectModes(t *testing.T) {
	base := scriptVideoLocalInput{ModelKey: "local-doubao-executor-video", Prompt: "own prompt"}
	for _, change := range []func(*scriptVideoLocalInput){
		func(v *scriptVideoLocalInput) { v.SubjectStrategy = "seedance_subject" },
		func(v *scriptVideoLocalInput) { v.SubjectReferenceURIs = []string{"asset://asset-own"} },
		func(v *scriptVideoLocalInput) { v.SubjectReferenceIDs = []string{"asset-own"} },
		func(v *scriptVideoLocalInput) { v.GenerationMode = "invalid" },
		func(v *scriptVideoLocalInput) {
			v.GenerationMode = "text_to_video"
			v.ImageURLs = []string{"https://own.invalid/image.png"}
		},
		func(v *scriptVideoLocalInput) { v.GenerationMode = "image_to_video" },
	} {
		input := base
		change(&input)
		if err := validateLocalScriptVideoInput(input); err == nil {
			t.Fatalf("unsupported input accepted: %+v", input)
		}
	}
	if err := validateLocalScriptVideoInput(base); err != nil {
		t.Fatal("legacy behavior changed", err)
	}
	text := base
	text.GenerationMode = "text_to_video"
	if err := validateLocalScriptVideoInput(text); err != nil {
		t.Fatal(err)
	}
}
