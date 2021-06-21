package icc

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/OpenSlides/openslides-icc-service/internal/iccerror"
	"github.com/OpenSlides/openslides-icc-service/internal/icchttp"
)

// Receiver is a type with the function Receive(). It is a blocking function
// that writes the icc-messages to the writer as soon as they occur.
type Receiver interface {
	Receive(ctx context.Context, w io.Writer, uid int) error
}

// HandleReceive registers the icc route.
func HandleReceive(mux *http.ServeMux, icc Receiver, auth icchttp.Authenticater) {
	mux.HandleFunc(
		icchttp.Path,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/octet-stream")

			uid := auth.FromContext(r.Context())
			// TODO: Can anonymous receive icc messages?

			if err := icc.Receive(r.Context(), w, uid); err != nil {
				icchttp.ErrorNoStatus(w, fmt.Errorf("receiving icc messages: %w", err))
				return
			}
		},
	)
}

// Sender saves a icc message.
type Sender interface {
	Send(io.Reader, int) error
}

// HandleSend registers the icc/send route.
func HandleSend(mux *http.ServeMux, icc Sender, auth icchttp.Authenticater) {
	mux.HandleFunc(
		icchttp.Path+"/send",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			uid := auth.FromContext(r.Context())
			if uid == 0 {
				w.WriteHeader(401)
				icchttp.ErrorNoStatus(w, iccerror.NewMessageError(iccerror.ErrNotAllowed, "Anonymous user can not send icc messages."))
				return
			}

			if err := icc.Send(r.Body, uid); err != nil {
				icchttp.Error(w, fmt.Errorf("saving icc message: %w", err))
				return
			}
		},
	)
}
