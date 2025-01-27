package actor

import (
	"time"

	"github.com/thlcodes/go-actress/log"
)

const (
	// no passivation by default
	DefaultPassivationTimeout = 0 * time.Second

	// the default mailbox size
	DefaultMailboxSize = 1000
)

// Actor interface do be fulfilled by actors implementations
type Actor interface {
	Handle(ctx Context, msg Message) (reply Message, err error)
}

type actor struct {
	log          log.Logger
	impl         Actor
	stopper      chan struct{}
	mailbox      chan *Envelope
	dropWhenFull bool
}

/* Actor impl */

func newActor(impl Actor, mailboxSize uint, log log.Logger) *actor {
	return &actor{
		log:     log,
		impl:    impl,
		mailbox: make(chan *Envelope, mailboxSize),
		stopper: make(chan struct{}, 1), // make buffered so that stopping never blocks
	}
}

// start the actor with the given context
func (a *actor) start(ctx Context) {
	a.log.Trace("start()")
	go a.loop(ctx)
}

// stop the actor
func (a *actor) stop(graceful bool) {
	a.log.Trace("stop(graceful=%t)", graceful)
	if graceful {
		a.channel() <- NewEnvelope(&Stop{})
	} else {
		close(a.stopper)
	}
}

// channel returns the mailbox channel
func (a *actor) channel() chan<- *Envelope {
	return a.mailbox
}

func (a *actor) loop(ctx Context) {
	a.log.Trace("loop()")
	closeMailbox := true
loop:
	for {
		select {
		case envelope, ok := <-a.mailbox:
			// received a message envelope from the mailbox
			if !ok {
				a.log.Debug("> mailbox is closed")
				// stop when mailbox is closed
				closeMailbox = false
				break loop
			}
			a.log.Debug("> received envelope {%s}", envelope)
			// handel message with copy of current context extended with sender
			a.handle(ctx.WithSender(envelope.sender), envelope)
			// stop actor when message was the stop signal
			if _, ok := envelope.msg.(*Stop); ok {
				a.log.Debug("> got a stop message")
				break loop
			}
		case <-ctx.Done():
			a.log.Debug("> context is done")
			// supervised stop through context
			// this is handled as graceful stop but
			// all messages left in mailbox will not be
			// processed
			a.handle(ctx, NewEnvelope(&Stop{}))
			break loop
		case <-a.stopper:
			a.log.Debug("> received stop message")
			// ungraceful stop
			break loop
		}
	}
	if closeMailbox {
		a.log.Debug("> closing mailbox due to closeMailbox=true")
		close(a.mailbox)
	}
}

// handle message, send reply/error to sender if
// there is one in the contex
func (a *actor) handle(ctx Context, envelope *Envelope) {
	msg := envelope.Msg()
	a.log.Trace("handle(ctx,msg=%T)", msg)
	var err error
	var reply Message
	reply, err = a.impl.Handle(ctx, msg)
	if ctx.Sender() == nil || envelope.isTell {
		return
	}
	if err != nil {
		a.log.Debug("> sending error %s to sender %s", err, ctx.Sender())
		_ = ctx.Tell(ctx.Sender(), &Error{Error: err})
	} else {
		a.log.Debug("> sending reply %T to sender %s", reply, ctx.Sender())
		_ = ctx.Tell(ctx.Sender(), reply)
	}

}
