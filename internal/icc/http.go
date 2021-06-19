package icc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/OpenSlides/openslides-icc-service/internal/log"
)

const (
	httpPath = "/system/icc"
)

type authenticater interface {
	Authenticate(http.ResponseWriter, *http.Request) (context.Context, error)
	FromContext(context.Context) int
}

// receiver is a type with the function Receive(). It is a blocking function
// that writes the icc-messages to the writer as soon as they occur.
type receiver interface {
	Receive(ctx context.Context, w io.Writer, uid int) error
}

// Notify registers the icc route.
func handleReceive(mux *http.ServeMux, icc receiver, auth authenticater) {
	mux.HandleFunc(
		httpPath,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/octet-stream")

			uid := auth.FromContext(r.Context())
			// TODO: Can anonymous receive icc messages?

			if err := icc.Receive(r.Context(), w, uid); err != nil {
				handleErrorNoStatus(w, err)
				return
			}
		},
	)
}

type sender interface {
	Send(io.Reader, int) error
}

// handleSend registers the icc/send route.
func handleSend(mux *http.ServeMux, icc sender, auth authenticater) {
	mux.HandleFunc(
		httpPath+"/send",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			uid := auth.FromContext(r.Context())
			if uid == 0 {
				http.Error(w, MessageError{ErrNotAllowed, "Anonymous user can not send icc messages"}.Error(), 401)
				return
			}

			if err := icc.Send(r.Body, uid); err != nil {
				handleError(w, err)
				return
			}
		},
	)
}

type applauser interface {
	Applause(uid int) error
}

// handleSendApplause registers the icc/applause route.
func handleSendApplause(mux *http.ServeMux, icc applauser, auth authenticater) {
	mux.HandleFunc(
		httpPath+"/applause",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			uid := auth.FromContext(r.Context())
			if uid == 0 {
				http.Error(w, MessageError{ErrNotAllowed, "Anonymous user can not send applause."}.Error(), 401)
				return
			}

			if err := icc.Applause(uid); err != nil {
				handleError(w, err)
				return
			}
		},
	)
}

func handleErrorNoStatus(w io.Writer, err error) {
	msg := err.Error()

	var errTyped interface {
		error
		Type() string
	}
	if !errors.As(err, &errTyped) {
		// Unknown error. Handle as 500er.
		msg = ErrInternal.Error()
		log.Info("Error: %v", err)
	}

	fmt.Fprint(w, msg)
}

func handleError(w http.ResponseWriter, err error) {
	var errTyped interface {
		error
		Type() string
	}
	status := 400
	if !errors.As(err, &errTyped) {
		// Unknown error. Handle as 500er.
		status = 500
	}

	w.WriteHeader(status)
	log.Debug("HTTP: Returning status %d", status)
	handleErrorNoStatus(w, err)
}
