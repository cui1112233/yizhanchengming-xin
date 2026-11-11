package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
)

const (
	HeaderUsername  = "X-Qiantie-Username"
	HeaderIsOwner   = "X-Qiantie-Is-Owner"
	HeaderIssuedAt  = "X-Qiantie-Issued-At"
	HeaderSignature = "X-Qiantie-Signature"
)

type BridgeIdentity struct {
	Username string
	IsOwner  bool
}

type bridgeIdentityKey struct{}

type BridgeUserResolver interface {
	ResolveBridgeUser(ctx context.Context, username string, isOwner bool) error
}

type BridgeAuth struct {
	Secret          string
	Now             func() time.Time
	MaxSkew         time.Duration
	Users           BridgeUserResolver
	AuthStore       auth.Store
	SessionResolver func(context.Context, string) (BridgeIdentity, error)
	IdentityScope   func(context.Context, BridgeIdentity, string) (BridgeIdentity, error)
}

func canonicalBridgePayload(username, issuedAt, isOwner, method, pathname string) string {
	return username + issuedAt + isOwner + method + pathname
}

func bridgeSignature(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func SignBridgeRequest(req *http.Request, username string, isOwner bool, issuedAt time.Time, secret string) {
	issued := strconv.FormatInt(issuedAt.Unix(), 10)
	owner := strconv.FormatBool(isOwner)
	req.Header.Set(HeaderUsername, username)
	req.Header.Set(HeaderIsOwner, owner)
	req.Header.Set(HeaderIssuedAt, issued)
	req.Header.Set(HeaderSignature, bridgeSignature(secret, canonicalBridgePayload(username, issued, owner, req.Method, req.URL.Path)))
}

func (a BridgeAuth) authenticate(req *http.Request) (BridgeIdentity, error) {
	if token := bearerToken(req); token != "" {
		if a.AuthStore != nil {
			now := time.Now
			if a.Now != nil {
				now = a.Now
			}
			if session, err := a.AuthStore.FindSession(req.Context(), token, now().UTC()); err == nil {
				account, accountErr := a.AuthStore.FindAccount(req.Context(), session.Username)
				if accountErr == nil && account.Active {
					return BridgeIdentity{Username: account.Username, IsOwner: account.IsOwner}, nil
				}
			}
		}
		if a.SessionResolver != nil {
			if identity, err := a.SessionResolver(req.Context(), token); err == nil {
				return identity, nil
			}
		}
	}
	if strings.TrimSpace(a.Secret) == "" {
		return BridgeIdentity{}, errors.New("bridge secret is not configured")
	}
	username := strings.TrimSpace(req.Header.Get(HeaderUsername))
	ownerRaw := req.Header.Get(HeaderIsOwner)
	issuedRaw := req.Header.Get(HeaderIssuedAt)
	signature := strings.TrimSpace(req.Header.Get(HeaderSignature))
	if username == "" || ownerRaw == "" || issuedRaw == "" || signature == "" {
		return BridgeIdentity{}, errors.New("missing signed bridge headers")
	}
	isOwner, err := strconv.ParseBool(ownerRaw)
	if err != nil {
		return BridgeIdentity{}, errors.New("invalid owner header")
	}
	issuedUnix, err := strconv.ParseInt(issuedRaw, 10, 64)
	if err != nil {
		return BridgeIdentity{}, errors.New("invalid issued-at header")
	}
	now := time.Now
	if a.Now != nil {
		now = a.Now
	}
	maxSkew := 5 * time.Minute
	if a.MaxSkew > 0 {
		maxSkew = a.MaxSkew
	}
	delta := now().Sub(time.Unix(issuedUnix, 0))
	if delta < 0 {
		delta = -delta
	}
	if delta > maxSkew {
		return BridgeIdentity{}, errors.New("expired bridge signature")
	}
	expected := bridgeSignature(a.Secret, canonicalBridgePayload(username, issuedRaw, ownerRaw, req.Method, req.URL.Path))
	provided, err := hex.DecodeString(signature)
	if err != nil {
		return BridgeIdentity{}, errors.New("malformed bridge signature")
	}
	expectedBytes, _ := hex.DecodeString(expected)
	if !hmac.Equal(provided, expectedBytes) {
		return BridgeIdentity{}, errors.New("invalid bridge signature")
	}
	if a.Users != nil {
		if err := a.Users.ResolveBridgeUser(req.Context(), username, isOwner); err != nil {
			return BridgeIdentity{}, fmt.Errorf("resolve bridge user: %w", err)
		}
	}
	return BridgeIdentity{Username: username, IsOwner: isOwner}, nil
}

func (a BridgeAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, err := a.authenticate(req)
		if err != nil {
			if bearerToken(req) != "" {
				w.Header().Set("X-Qiantie-Auth-Failure", "session")
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		if a.IdentityScope != nil {
			identity, err = a.IdentityScope(req.Context(), identity, req.Header.Get(HeaderNovelFetchAccount))
			if err != nil {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
				return
			}
		}
		next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), bridgeIdentityKey{}, identity)))
	})
}

func BridgeIdentityFromContext(ctx context.Context) (BridgeIdentity, bool) {
	identity, ok := ctx.Value(bridgeIdentityKey{}).(BridgeIdentity)
	return identity, ok
}

// SessionOnly returns a copy of the bridge auth that never accepts signed
// X-Qiantie-* headers. Only a real login session (Bearer token) can pass.
func (a BridgeAuth) SessionOnly() BridgeAuth {
	a.Secret = ""
	return a
}

// stripBridgeIdentityHeaders removes caller-supplied identity headers before
// forwarding a request to an internal hop that authenticates again.
func stripBridgeIdentityHeaders(req *http.Request) {
	if req == nil {
		return
	}
	for _, header := range []string{HeaderUsername, HeaderIsOwner, HeaderIssuedAt, HeaderSignature} {
		req.Header.Del(header)
	}
}
