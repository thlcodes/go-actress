package actor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/thlcoes/go-actress/log"
)

type System interface {
	supervisor
	talker
	Stop()
	SetLogger(log.Logger)
}

var _ System = (*system)(nil)

type system struct {
	ctx       context.Context
	cancelCtx func()

	log log.Logger

	lock    sync.RWMutex
	currIdx uint64

	actors map[localRef]*actor
}

// NewSystem will create a new actor system
func NewSystem() System {
	ctx, cancel := context.WithCancel(context.Background())
	return &system{
		ctx:       ctx,
		log:       log.NewStdLogger().WithLevel(log.INFO).WithPrefix("System"),
		cancelCtx: cancel,
		currIdx:   0,
		actors:    map[localRef]*actor{},
	}
}

// Set logger
func (s *system) SetLogger(log log.Logger) {
	s.log = log
}

// Spawn will start given actor instance
func (s *system) Spawn(instance Actor) Ref {
	s.log.Trace("Spawn(instance=%T)", instance)
	s.currIdx++
	ref := newLocalRef(s.currIdx)
	actor := newActor(instance, DefaultMailboxSize, s.log.SubLogger(fmt.Sprintf("actor#%d", ref.id)))
	actor.start(newActorContext(s.ctx, s, &ref))
	s.lock.Lock()
	s.actors[ref] = actor
	s.lock.Unlock()
	s.log.Debug("Spawned new local actor with ref %#v", ref)
	return &ref
}

// Kill an actor, optinally graceful
func (s *system) Kill(ref Ref, graceful bool) error {
	s.log.Trace("Kill(ref=%#v,graceful=%t)", ref, graceful)
	lref, ok := ref.(*localRef)
	if !ok {
		return ErrUnsupportedRef(ref)
	}
	s.lock.Lock()
	actor := s.actors[*lref]
	delete(s.actors, *lref)
	s.lock.Unlock()
	actor.stop(graceful)
	return nil
}

func (s *system) Stop() {
	s.log.Trace("Stop()")
	// propagate cancel via context
	s.cancelCtx()
}

// Tell sends a message to an actor ref but not wait for a reply
func (s *system) Tell(whom Ref, what Message, opts ...TalkOptions) error {
	s.log.Trace("Tell(whom=%s,what=%T,opts=%T)", whom, what, opts)
	var ch chan<- *Envelope
	switch ref := whom.(type) {
	case *channelRef:
		ch = ref.ch
	case *localRef:
		s.lock.RLock()
		actor, ok := s.actors[*ref]
		s.lock.RUnlock()
		if !ok {
			return ErrActorNotFound(ref)
		}
		ch = actor.channel()
	default:
		return ErrUnsupportedRefForTalking(ref)
	}

	envelope := NewEnvelope(what, opts...)
	select {
	case ch <- envelope:
		// yeah!
	default:
		// mailbox/channel full
		return ErrMailboxFull(whom)
	}
	return nil
}

// TODO: as option
const AskTimeout = 3 * time.Second

// Ask will send a message to an actor ref, intercept the response/error and return
// it to the sender
func (s *system) Ask(whom Ref, what Message, opts ...TalkOptions) (reply Message, err error) {
	s.log.Trace("Ask(whom=%s,what=%T,opts=%T)", whom, what, opts)
	ch := make(chan *Envelope, 1)
	cref := newChannelRef(ch)
	s.Tell(whom, what, WithSender(&cref))
	var replyEnvelope *Envelope
	var ok bool
	select {
	case replyEnvelope, ok = <-cref.ch:
		if !ok {
			return nil, ErrChannelRefChannelClosed(&cref)
		}
	case <-time.After(AskTimeout):
		return nil, ErrTalkTimeout
	}
	return replyEnvelope.msg, nil
}
