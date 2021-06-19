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
	startMessage string
	expectedErr  error
	unblocking   bool

	sendChan chan string
}

func (r *receiverStub) Receive(ctx context.Context, w io.Writer, uid int) error {
	if _, err := w.Write([]byte(r.startMessage)); err != nil {
		return fmt.Errorf("writing first message: %w", err)
	}

	if r.unblocking {
		return r.expectedErr
	}

	r.sendChan = make(chan string, 1)

	for message := range r.sendChan {
		if _, err := w.Write([]byte(message)); err != nil {
			return fmt.Errorf("writing first message: %w", err)
		}
	}

	return r.expectedErr
}

func (r *receiverStub) send(message string) {
	r.sendChan <- message
}
