package objectkey

import "testing"

func TestObjectKeyPrefixNormalization(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"", "video/7/9.mp4"}, {" \t\n", "video/7/9.mp4"},
		{"staging", "staging/video/7/9.mp4"}, {"staging/", "staging/video/7/9.mp4"},
		{" staging/// ", "staging/video/7/9.mp4"}, {"staging/local-run/", "staging/local-run/video/7/9.mp4"},
		{"Run_1.2-3/", "Run_1.2-3/video/7/9.mp4"},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			p, err := ParsePrefix(tc.raw)
			if err != nil {
				t.Fatal(err)
			}
			got, err := p.Apply("video/7/9.mp4")
			if err != nil || got != tc.want {
				t.Fatalf("key=%q err=%v want=%q", got, err, tc.want)
			}
		})
	}
}

func TestObjectKeyPrefixRejectsUnsafeConfiguration(t *testing.T) {
	for _, raw := range []string{"/staging", "/", "../staging", "staging/../x", "staging/./x", "staging//x", `staging\x`, "staging/%2f", "staging/x\nq", "https://example.com", "staging?x", "staging#x", "staging x", "staging/中文"} {
		t.Run(raw, func(t *testing.T) {
			if _, err := ParsePrefix(raw); err == nil {
				t.Fatal("unsafe prefix accepted")
			}
		})
	}
}

func TestObjectKeyPrefixApplyContainment(t *testing.T) {
	p, err := ParsePrefix("staging/")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"", "/video/7/9.mp4", "video/7/", "video//9.mp4", "video/./x", "video/../x", `video\x`, "video/%2F/x", "video/x\nq", "video/x?q", "video/x#q", "staging/video/7/9.mp4", "staging/staging/video/7/9.mp4", "https://example.com/x"} {
		t.Run(key, func(t *testing.T) {
			if _, err := p.Apply(key); err == nil {
				t.Fatal("unsafe or already prefixed key accepted")
			}
		})
	}
	for _, tc := range []struct{ key, want string }{
		{"merge/7/2.mp4", "staging/merge/7/2.mp4"},
		{"shuihuo/project-4/book-9/random.png", "staging/shuihuo/project-4/book-9/random.png"},
		{"agent/project-9/random", "staging/agent/project-9/random"},
		{"staging-other/x", "staging/staging-other/x"},
	} {
		got, err := p.Apply(tc.key)
		if err != nil || got != tc.want {
			t.Fatalf("key=%q err=%v want=%q", got, err, tc.want)
		}
	}
}

func TestObjectKeyPrefixEmptyIsIdentity(t *testing.T) {
	var p Prefix
	for _, key := range []string{"", " /legacy//key ", "../legacy", "staging/video/x"} {
		got, err := p.Apply(key)
		if err != nil || got != key {
			t.Fatalf("key=%q err=%v want=%q", got, err, key)
		}
	}
}
