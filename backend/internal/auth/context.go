package auth

import "context"

type contextKey string

const accessTokenClaimsContextKey contextKey = "accessTokenClaims"

func ContextWithAccessTokenClaims(ctx context.Context, claims AccessTokenClaims) context.Context {
	return context.WithValue(ctx, accessTokenClaimsContextKey, claims)
}

func AccessTokenClaimsFromContext(ctx context.Context) (AccessTokenClaims, bool) {
	claims, ok := ctx.Value(accessTokenClaimsContextKey).(AccessTokenClaims)
	return claims, ok
}
