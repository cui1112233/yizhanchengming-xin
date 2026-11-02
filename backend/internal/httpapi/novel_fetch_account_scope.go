package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"qiantie/backend/internal/auth"
)

const HeaderNovelFetchAccount = "X-Qiantie-Novel-Fetch-Account"

var errNovelFetchAccountForbidden = errors.New("无权查看该账号的小说任务")

func resolveNovelFetchAccountScope(ctx context.Context, actor BridgeIdentity, requested string, accounts auth.Store, profiles *MemberProfileStore) (BridgeIdentity, error) {
	targetUsername := strings.TrimSpace(requested)
	if targetUsername == "" || targetUsername == actor.Username {
		return actor, nil
	}
	if accounts == nil || profiles == nil {
		return BridgeIdentity{}, errNovelFetchAccountForbidden
	}
	actorProfile := profiles.member(actor.Username, accounts)
	targetAccount, err := accounts.FindAccount(ctx, targetUsername)
	if actorProfile == nil || err != nil || !targetAccount.Active {
		return BridgeIdentity{}, errNovelFetchAccountForbidden
	}
	targetProfile := profiles.member(targetUsername, accounts)
	if targetProfile == nil {
		return BridgeIdentity{}, errNovelFetchAccountForbidden
	}
	switch toString(actorProfile["role"]) {
	case "dev":
		return BridgeIdentity{Username: targetUsername, IsOwner: targetAccount.IsOwner}, nil
	case "manager":
		if toString(targetProfile["role"]) == "member" && strings.TrimSpace(toString(targetProfile["boundTo"])) == actor.Username {
			return BridgeIdentity{Username: targetUsername, IsOwner: targetAccount.IsOwner}, nil
		}
	}
	return BridgeIdentity{}, errNovelFetchAccountForbidden
}

func novelFetchAccountScopeResolver(accounts auth.Store, profiles *MemberProfileStore) func(context.Context, BridgeIdentity, string) (BridgeIdentity, error) {
	return func(ctx context.Context, actor BridgeIdentity, requested string) (BridgeIdentity, error) {
		return resolveNovelFetchAccountScope(ctx, actor, requested, accounts, profiles)
	}
}

func novelFetchAccountScopeMiddleware(accounts auth.Store, profiles *MemberProfileStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		actor, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		identity, err := resolveNovelFetchAccountScope(req.Context(), actor, req.Header.Get(HeaderNovelFetchAccount), accounts, profiles)
		if err != nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), bridgeIdentityKey{}, identity)))
	})
}
