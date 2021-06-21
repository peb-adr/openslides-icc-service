// Package icctest contains test helpers for all other packages of this module.
package icctest

import (
	"context"
	"net/http"
)

type authError struct{}

func (authError) Error() string {
	return `{"error":"auth","msg":"auth error"}`
}

func (authError) Type() string {
	return "auth"
}

// AutherStub impplements the icchelper.Auther interface
type AutherStub struct {
	UserID  int
	AuthErr bool
}

// Authenticate does nothing. The Stub uses the userID that it was initialized
// with.
func (a *AutherStub) Authenticate(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	if a.AuthErr {
		return nil, authError{}
	}
	return r.Context(), nil
}

// FromContext returns the user id the stub was initializes with.
func (a *AutherStub) FromContext(context.Context) int {
	return a.UserID
}
