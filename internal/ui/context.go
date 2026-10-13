package ui

import (
	"context"
	"os"
)

type contextKey struct{}

var uiContextKey = contextKey{}

func WithContext(parentContext context.Context, uiInstance *UI) context.Context {
	return context.WithValue(parentContext, uiContextKey, uiInstance)
}

func FromContext(parentContext context.Context) *UI {
	if uiInstance, isFound := parentContext.Value(uiContextKey).(*UI); isFound && uiInstance != nil {
		return uiInstance
	}
	return New(Config{
		Out:     os.Stdout,
		Err:     os.Stderr,
		IsTTY:   IsTerminal(os.Stderr) && !IsCI(),
		NoColor: HasNoColor(),
	})
}
