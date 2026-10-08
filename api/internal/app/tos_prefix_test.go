package app

import (
	"os"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/objectkey"
)

func TestTOSKeyPrefixConfiguration(t *testing.T) {
	for _, name := range []string{"TOS_ENDPOINT", "TOS_REGION", "TOS_BUCKET", "TOS_ACCESS_KEY", "TOS_SECRET_KEY", "TOS_PUBLIC_BASE_URL", "QIANTIE_TEXT_API_BASE_URL", "QIANTIE_TEXT_API_KEY", "QIANTIE_TEXT_MODEL", "SHUIHUO_IMAGE_PROVIDER_BASE_URL", "SHUIHUO_IMAGE_PROVIDER_API_KEY", "SHUIHUO_IMAGE_PROVIDER_MODEL", "SHUIHUO_TTS_PROVIDER_BASE_URL", "SHUIHUO_TTS_PROVIDER_API_KEY", "SHUIHUO_TTS_PROVIDER_MODEL", "REDIS_ADDR", "QIANTIE_REDIS_ADDR", "TASK9_REDIS_ADDR"} {
		t.Setenv(name, "")
	}
	for _, tc := range []struct {
		raw, wantPrefix string
		invalid         bool
	}{
		{"", "", false}, {" \t", "", false}, {" staging/// ", "staging/", false},
		{"staging/local-run/", "staging/local-run/", false}, {"/staging", "", true},
		{"staging/../production", "", true}, {"staging//x", "", true}, {"staging/%2f", "", true},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			t.Setenv("TOS_KEY_PREFIX", tc.raw)
			calls := 0
			err := configureTOSKeyPrefix(os.Getenv("TOS_KEY_PREFIX"), func(prefix objectkey.Prefix) {
				calls++
				for _, relative := range []string{"video/7/9.mp4", "merge/7/2.mp4", "shuihuo/project-4/book-9/output.png", "agent/project-9/attachment"} {
					key, err := prefix.Apply(relative)
					if err != nil || key != tc.wantPrefix+relative {
						t.Fatalf("key=%q err=%v", key, err)
					}
				}
			})
			if tc.invalid {
				if err == nil || calls != 0 {
					t.Fatalf("invalid config reached wiring: calls=%d err=%v", calls, err)
				}
				if err.Error() != "app: TOS key prefix configuration is invalid" {
					t.Fatalf("unsafe diagnostic=%q", err.Error())
				}
			} else if err != nil || calls != 1 {
				t.Fatalf("valid wiring: calls=%d err=%v", calls, err)
			}
		})
	}
}
