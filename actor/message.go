package actor

import (
	"fmt"
)

type Message interface {
	message()
}

type EnvelopeOption = TalkOption

/* implementations */

type Envelope struct {
	sender Ref
	msg    Message
	isTell bool
}

func NewEnvelope(msg Message, opts ...EnvelopeOption) *Envelope {
	e := &Envelope{
		msg: msg,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Envelope) String() string {
	return fmt.Sprintf("sender=%s msg=%T", e.sender, e.msg)
}

func (e *Envelope) Sender() Ref {
	return e.sender
}

func (e *Envelope) Msg() Message {
	return e.msg
}

func WithSender(sender Ref) EnvelopeOption {
	return func(e *Envelope) {
		e.sender = sender
	}
}

func Tell(e *Envelope) {
	e.isTell = true
}

/* pre defined messages */

type Stop struct {
	Message
}

type Error struct {
	Message
	Error error
}
