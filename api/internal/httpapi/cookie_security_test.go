package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
)

func TestProductionAuthCookieAttributes(t *testing.T) {
	h := handler{deps: Dependencies{SecureCookies: true}}
	rec := httptest.NewRecorder()
	h.setAuthCookies(rec, authn.Credentials{AccessToken: "access-value", RefreshToken: "refresh-value"})

	cookies := rec.Result().Cookies()
	byName := make(map[string]*http.Cookie, len(cookies))
	for _, cookie := range cookies {
		byName[cookie.Name] = cookie
	}

	access := byName[AccessCookieName]
	if access == nil {
		t.Fatal("access cookie missing")
	}
	if !access.HttpOnly || !access.Secure || access.SameSite != http.SameSiteLaxMode || access.Path != "/" || access.MaxAge != 15*60 {
		t.Fatalf("unsafe access cookie attributes: %#v", access)
	}

	refresh := byName[RefreshCookieName]
	if refresh == nil {
		t.Fatal("refresh cookie missing")
	}
	if !refresh.HttpOnly || !refresh.Secure || refresh.SameSite != http.SameSiteLaxMode || refresh.Path != "/api/auth" || refresh.MaxAge != 30*24*60*60 {
		t.Fatalf("unsafe refresh cookie attributes: %#v", refresh)
	}
}
