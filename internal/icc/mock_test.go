package icc

import (
	"context"
	"errors"
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

type backendStub struct {
	messages         chan []byte
	receivedMessages [][]byte
}

func newBackendStrub() *backendStub {
	b := backendStub{}
	b.messages = make(chan []byte, 10)
	return &b
}

func (b *backendStub) reset() {
	b.receivedMessages = b.receivedMessages[0:0]
	for {
		select {
		case <-b.messages:
		default:
			return
		}
	}
}

func (b *backendStub) SendICC(bs []byte) error {
	b.messages <- bs
	b.receivedMessages = append(b.receivedMessages, bs)
	return nil
}

func (b *backendStub) ReceiveICC(ctx context.Context) (message []byte, err error) {
	select {
	case m := <-b.messages:
		return m, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (b *backendStub) SendApplause(userID int, time int64) error {
	return errors.New("TODO")
}

func (b *backendStub) ReceiveApplause(since int64) (int, error) {
	return 0, errors.New("TODO")
}
