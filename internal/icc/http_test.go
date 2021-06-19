package icc

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleReceive(t *testing.T) {
	url := "/system/icc"

	t.Run("Receiver is called", func(t *testing.T) {
		receiver := receiverStub{
			expectedMessage: "my answer",
		}
		auther := autherStub{}
		mux := http.NewServeMux()
		handleReceive(mux, &receiver, &auther)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, httptest.NewRequest("GET", url, nil))

		if resp.Result().StatusCode != 200 {
			t.Fatalf("handler returned status %s: %s", resp.Result().Status, resp.Body.String())
		}

		if !receiver.called {
			t.Errorf("receiver was not called")
		}

		if resp.Body.String() != "my answer" {
			t.Errorf("resp body is `%s`, expected `my answer`", resp.Body.String())
		}
	})

	t.Run("Receiver has an internal error", func(t *testing.T) {
		myError := errors.New("Test error")
		receiver := receiverStub{
			expectedErr: myError,
		}
		auther := autherStub{}
		mux := http.NewServeMux()
		handleReceive(mux, &receiver, &auther)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, httptest.NewRequest("GET", url, nil))

		if resp.Result().StatusCode != 200 {
			t.Fatalf("handler returned status %s: %s", resp.Result().Status, resp.Body.String())
		}

		if strings.Contains(resp.Body.String(), myError.Error()) {
			t.Errorf("handler returned the error message: %s", resp.Body.String())
		}
	})

	t.Run("Receiver has an error for the client", func(t *testing.T) {
		myError := ErrInvalid
		receiver := receiverStub{
			expectedErr: myError,
		}
		auther := autherStub{}
		mux := http.NewServeMux()
		handleReceive(mux, &receiver, &auther)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, httptest.NewRequest("GET", url, nil))

		if resp.Result().StatusCode != 200 {
			t.Fatalf("handler returned status %s: %s", resp.Result().Status, resp.Body.String())
		}

		if !strings.Contains(resp.Body.String(), myError.Error()) {
			t.Errorf("handler did not return the error message: %s", resp.Body.String())
		}
	})
}

func TestHandleSend(t *testing.T) {
	url := "/system/icc/send"

	t.Run("Anonymous", func(t *testing.T) {
		auther := autherStub{}
		sender := senderStub{}
		mux := http.NewServeMux()
		handleSend(mux, &sender, &auther)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, httptest.NewRequest("GET", url, nil))

		if resp.Result().StatusCode != 401 {
			t.Fatalf("handler returned status %s: %s", resp.Result().Status, resp.Body.String())
		}

		if !strings.Contains(resp.Body.String(), ErrNotAllowed.Type()) {
			t.Errorf("handler returned message `%s`, expected to contain `%s`", resp.Body.String(), ErrNotAllowed.Type())
		}

		if sender.called {
			t.Errorf("handler did call the sender")
		}
	})

	t.Run("User", func(t *testing.T) {
		auther := autherStub{
			userID: 1,
		}
		sender := senderStub{}
		mux := http.NewServeMux()
		handleSend(mux, &sender, &auther)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, httptest.NewRequest("GET", url, nil))

		if resp.Result().StatusCode != 200 {
			t.Fatalf("handler returned status %s: %s", resp.Result().Status, resp.Body.String())
		}

		if !sender.called {
			t.Errorf("handler did not call the sender")
		}

		if sender.calledUserID != 1 {
			t.Errorf("sender was called with userID %d, expected 1", sender.calledUserID)
		}
	})

	t.Run("Internal error", func(t *testing.T) {
		myError := errors.New("Test error")
		sender := senderStub{
			expectedErr: myError,
		}
		auther := autherStub{
			userID: 1,
		}
		mux := http.NewServeMux()
		handleSend(mux, &sender, &auther)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, httptest.NewRequest("GET", url, nil))

		if resp.Result().StatusCode != 500 {
			t.Fatalf("handler returned status %s: %s", resp.Result().Status, resp.Body.String())
		}

		if strings.Contains(resp.Body.String(), myError.Error()) {
			t.Errorf("handler returned the error message: %s", resp.Body.String())
		}
	})
}

func TestHandleSendApplause(t *testing.T) {
	url := "/system/icc/applause"

	t.Run("Anonymous", func(t *testing.T) {
		auther := autherStub{}
		applauser := applauserStrub{}
		mux := http.NewServeMux()
		handleSendApplause(mux, &applauser, &auther)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, httptest.NewRequest("GET", url, nil))

		if resp.Result().StatusCode != 401 {
			t.Fatalf("handler returned status %s: %s", resp.Result().Status, resp.Body.String())
		}

		if !strings.Contains(resp.Body.String(), ErrNotAllowed.Type()) {
			t.Errorf("handler returned message `%s`, expected to contain `%s`", resp.Body.String(), ErrNotAllowed.Type())
		}

		if applauser.called {
			t.Errorf("handler did call the applauser")
		}
	})

	t.Run("User", func(t *testing.T) {
		auther := autherStub{
			userID: 1,
		}
		applauser := applauserStrub{}
		mux := http.NewServeMux()
		handleSendApplause(mux, &applauser, &auther)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, httptest.NewRequest("GET", url, nil))

		if resp.Result().StatusCode != 200 {
			t.Fatalf("handler returned status %s: %s", resp.Result().Status, resp.Body.String())
		}

		if !applauser.called {
			t.Errorf("handler did not call the applauser")
		}

		if applauser.calledUserID != 1 {
			t.Errorf("applauser was called with userID %d, expected 1", applauser.calledUserID)
		}
	})

	t.Run("Internal error", func(t *testing.T) {
		myError := errors.New("Test error")
		applauser := applauserStrub{
			expectedErr: myError,
		}
		auther := autherStub{
			userID: 1,
		}
		mux := http.NewServeMux()
		handleSendApplause(mux, &applauser, &auther)
		resp := httptest.NewRecorder()

		mux.ServeHTTP(resp, httptest.NewRequest("GET", url, nil))

		if resp.Result().StatusCode != 500 {
			t.Fatalf("handler returned status %s: %s", resp.Result().Status, resp.Body.String())
		}

		if strings.Contains(resp.Body.String(), myError.Error()) {
			t.Errorf("handler returned the error message: %s", resp.Body.String())
		}
	})
}
