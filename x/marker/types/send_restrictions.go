package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var bypassKey = "bypass-marker-restriction"

// WithBypass returns a new context that will cause the marker bank send restriction to be skipped.
func WithBypass(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(bypassKey, true)
}

// WithoutBypass returns a new context that will cause the marker bank send restriction to not be skipped.
func WithoutBypass(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(bypassKey, false)
}

// HasBypass checks the context to see if the marker bank send restriction should be skipped.
func HasBypass(ctx sdk.Context) bool {
	bypassValue := ctx.Value(bypassKey)
	if bypassValue == nil {
		return false
	}
	bypass, isBool := bypassValue.(bool)
	return isBool && bypass
}
