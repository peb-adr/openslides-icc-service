package icc

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type authError struct{}

func (authError) Error() string {
	return `{"error":"auth","msg":"auth error"}`
}

func (authError) Type() string {
	return "auth"
}

type autherStub struct {
	userID  int
	authErr bool
}

func (a *autherStub) Authenticate(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	if a.authErr {
		return nil, authError{}
	}
	return r.Context(), nil
}

func (a *autherStub) FromContext(context.Context) int {
	return a.userID
}

type receiverStub struct {
	expectedMessage string
	expectedErr     error
	called          bool
}

func (r *receiverStub) Receive(ctx context.Context, w io.Writer, uid int) error {
	r.called = true

	if _, err := w.Write([]byte(r.expectedMessage)); err != nil {
		return fmt.Errorf("writing first message: %w", err)
	}

	return r.expectedErr
}

type senderStub struct {
	expectedErr  error
	called       bool
	calledUserID int
}

func (s *senderStub) Send(r io.Reader, uid int) error {
	s.called = true
	s.calledUserID = uid
	return s.expectedErr
}

type applauserStrub struct {
	expectedErr  error
	called       bool
	calledUserID int
}

func (s *applauserStrub) Applause(uid int) error {
	s.called = true
	s.calledUserID = uid
	return s.expectedErr
}
