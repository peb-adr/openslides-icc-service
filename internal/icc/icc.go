package icc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/OpenSlides/openslides-icc-service/internal/log"
	"github.com/ostcar/topic"
)

// Backend stores the icc messages.
type Backend interface {
	// SendNotify saves a valid icc message.
	SendNotify([]byte) error

	// ReceiveNotify is a blocking function that receives the messages.
	//
	// The first call returnes the first icc message, the next call the second
	// an so on. If there are no more messages to read, the function blocks
	// until there is or the context ist canceled.
	//
	// It is expected, that only one goroutine is calling this function. The
	// Backend keeps track what the last send message was.
	ReceiveNotify(ctx context.Context) (message []byte, err error)

	// SendApplause saves an applause fro the user. A user can save as many
	// applause as he wonts, but the backend only counds it once per user in a
	// given time.
	SendApplause(userID int) error

	// ReceiveApplause returned all applause since a given time in seconds. Each user is
	// only called once.
	ReceiveApplause(since int64) (int, error)
}

// ICC holds the state of the service.
type ICC struct {
	backend Backend
	cIDGen  cIDGen
	topic   *topic.Topic
}

// New returns an initialized state of the service.
//
// The New function is not blocking. The context is used to stop a goroutine
// that is started by this function.
func New(ctx context.Context, b Backend) *ICC {
	icc := ICC{
		backend: b,
		topic:   topic.New(),
	}

	go icc.listen(ctx)
	return &icc
}

// listen waits for notify messages from the backend and saves them into the
// topic.
func (icc *ICC) listen(ctx context.Context) {
	for {
		m, err := icc.backend.ReceiveNotify(ctx)
		if err != nil {
			var closing interface {
				Closing()
			}
			if errors.As(err, &closing) {
				return
			}

			log.Info("Notify: Can not receive data from backend: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		icc.topic.Publish(string(m))
	}
}

type flusher interface {
	Flush()
}

// Receive is a blocking function that sends data to w as soon as an icc event
// happens.
func (icc *ICC) Receive(ctx context.Context, w io.Writer, uid int) error {
	cid := icc.cIDGen.generate(uid)

	if _, err := fmt.Fprintf(w, `{"channel_id": "%s"}`+"\n", cid); err != nil {
		return fmt.Errorf("sending channel id: %w", err)
	}

	if f, ok := w.(flusher); ok {
		f.Flush()
	}

	encoder := json.NewEncoder(w)
	tid := icc.topic.LastID()
	var messages []string
	var err error
	for {
		tid, messages, err = icc.topic.Receive(ctx, tid)
		if err != nil {
			return fmt.Errorf("fetching message from topic: %w", err)
		}

		for _, message := range messages {
			var m iccMessage
			if err := json.Unmarshal([]byte(message), &m); err != nil {
				return fmt.Errorf("decoding message: %w", err)
			}

			if !m.forMe(uid, cid) {
				continue
			}

			out := struct {
				SenderUserID    int             `json:"sender_user_id"`
				SenderChannelID string          `json:"sender_channel_id"`
				Name            string          `json:"name"`
				Message         json.RawMessage `json:"message"`
			}{
				m.ChannelID.uid(),
				m.ChannelID.String(),
				m.Name,
				m.Message,
			}

			if err := encoder.Encode(out); err != nil {
				return fmt.Errorf("sending message: %w", err)
			}

		}

		if f, ok := w.(flusher); ok {
			f.Flush()
		}
	}
}

// Send reads and saves the icc event from the given reader.
func (icc *ICC) Send(r io.Reader, uid int) error {
	bs, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading message: %w", err)
	}

	if err := validateMessage(bs, uid); err != nil {
		return fmt.Errorf("validate message: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := json.Compact(buf, bs); err != nil {
		return fmt.Errorf("compacting message: %w", err)
	}

	if err := icc.backend.SendNotify(buf.Bytes()); err != nil {
		return fmt.Errorf("sending message to backend: %w", err)
	}

	return nil
}

// Applause saves an applause event from the given user.
func (icc *ICC) Applause(uid int) error {
	return errors.New("TODO")
}

func validateMessage(m []byte, userID int) error {
	var message iccMessage

	if err := json.Unmarshal(m, &message); err != nil {
		return newMessageError(ErrInvalid, "invalid json: %v", err)
	}

	if message.ChannelID.uid() != userID {
		return fmt.Errorf("invalid channel id")
	}

	if message.Name == "" {
		return fmt.Errorf("notify does not have required field `name`")
	}

	if message.Name == "applause" {
		return fmt.Errorf("notify name can not be applause")
	}
	return nil
}

type iccMessage struct {
	ChannelID  channelID       `json:"channel_id"`
	ToAll      bool            `json:"to_all"`
	ToUsers    []int           `json:"to_users"`
	ToChannels []string        `json:"to_channels"`
	Name       string          `json:"name"`
	Message    json.RawMessage `json:"message"`
}

func (m iccMessage) forMe(uid int, cID channelID) bool {
	if m.ToAll {
		return true
	}

	for _, toUID := range m.ToUsers {
		if toUID == uid {
			return true
		}
	}

	for _, toCID := range m.ToChannels {
		if toCID == cID.String() {
			return true
		}
	}
	return false
}
