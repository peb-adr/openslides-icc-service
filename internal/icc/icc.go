package icc

import (
	"context"
	"errors"
	"io"
)

// ICC holds the state of the service.
type ICC struct{}

// New returns an initialized state of the service.
func New() *ICC {
	return &ICC{}
}

// Receive is a blocking function that sends data to w as soon as an icc event
// for the user happens.
func (icc *ICC) Receive(ctx context.Context, w io.Writer, uid int) error {
	return errors.New("TODO")
}

// Send reads and saves the icc event from the given reader.
func (icc *ICC) Send(r io.Reader, uid int) error {
	return errors.New("TODO")
}

// Applause saves an applause event from the given user.
func (icc *ICC) Applause(uid int) error {
	return errors.New("TODO")
}
